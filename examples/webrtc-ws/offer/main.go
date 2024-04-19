package main

import (
	"encoding/json"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
)

type Manager struct {
	wsConn            *websocket.Conn
	pConn             *webrtc.PeerConnection
	pendingCandidates []*webrtc.ICECandidate
}

type wsMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

var (
	err           error
	candidatesMux sync.Mutex
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)

	m := Manager{
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	yalog.Infof("Connecting to %s", u.String())

	m.wsConn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		yalog.Fatal("Error connecting to WebSocket:", err)
	}
	defer func() {
		if cErr := m.wsConn.Close(); cErr != nil {
			yalog.Infof("cannot close wsConn: %v\n", cErr)
		}
	}()
	yalog.Info("Connected to WebSocket server")

	m.pConn, err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.syncthing.net:3478"},
			},
		},
	})
	if err != nil {
		yalog.Fatal("Error creating peer connection:", err)
	}
	defer func() {
		if cErr := m.pConn.Close(); cErr != nil {
			yalog.Errorf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	m.pConn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		desc := m.pConn.RemoteDescription()
		if desc == nil {
			m.pendingCandidates = append(m.pendingCandidates, c)
		} else {
			candidateData, err := json.Marshal(c)
			if err != nil {
				yalog.Errorf("marshal answer error: %v\n", err)
			}
			if err := m.wsConn.WriteJSON(&wsMessage{
				Event: "candidate",
				Data:  string(candidateData),
			}); err != nil {
				yalog.Errorf("write message error: %v\n", err)
			}
		}
	})

	_, err := m.pConn.CreateDataChannel("data", nil)
	if err != nil {
		yalog.Error(err)
	}

	m.pConn.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			yalog.Infof("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	m.pConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		yalog.Infof("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			yalog.Info("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}

		if s == webrtc.PeerConnectionStateClosed {
			yalog.Info("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})

	m.pConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		yalog.Infof("Connection State has changed %s \n", connectionState.String())
		if connectionState.String() == "connected" {
			yalog.Info("ICE 连接建立")
		}
	})

	offer, err := m.pConn.CreateOffer(nil)
	if err != nil {
		yalog.Fatal("Error creating offer:", err)
	}
	if err := m.pConn.SetLocalDescription(offer); err != nil {
		yalog.Fatal("Error setting local description:", err)
	}
	offerData, err := json.Marshal(offer)
	if err != nil {
		yalog.Fatal("Error marshaling offer:", err)
	}
	if err := m.wsConn.WriteJSON(&wsMessage{
		Event: "offer",
		Data:  string(offerData),
	}); err != nil {
		yalog.Fatal("Error sending offer:", err)
	}

	go func() {
		for {
			msg := &wsMessage{}
			err = m.wsConn.ReadJSON(msg)
			if err != nil {
				yalog.Errorf("read message error: %v\n", err)
				m.wsConn.Close()
				return
			}
			if msg.Event == "answer" {
				answer := webrtc.SessionDescription{}
				if err := json.Unmarshal([]byte(msg.Data), &answer); err != nil {
					yalog.Errorf("unmarshal offer error: %v\n", err)
				}

				if err := m.pConn.SetRemoteDescription(answer); err != nil {
					yalog.Errorf("set remote description error: %v\n", err)
				}

				candidatesMux.Lock()

				for _, c := range m.pendingCandidates {
					candidateData, err := json.Marshal(c)
					if err != nil {
						yalog.Errorf("marshal answer error: %v\n", err)
					}
					if err := m.wsConn.WriteJSON(&wsMessage{
						Event: "candidate",
						Data:  string(candidateData),
					}); err != nil {
						yalog.Errorf("write message error: %v\n", err)
					}
				}
				candidatesMux.Unlock()
			} else if msg.Event == "candidate" {
				var candidate webrtc.ICECandidate
				err := json.Unmarshal([]byte(msg.Data), &candidate)
				if err != nil {
					yalog.Errorf("parse ice candidate error: %v\n", err)
				}
				err = m.pConn.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate.ToJSON().Candidate})
				if err != nil {
					yalog.Errorf("add ice candidate error: %v\n", err)
				}
			} else {
				yalog.Errorf("Unknown event: %v %v\n", msg.Event, msg.Data)
			}
		}
	}()

	<-interrupt
}
