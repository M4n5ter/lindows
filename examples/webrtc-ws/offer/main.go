package main

import (
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
	Event   string `json:"event"`
	PayLoad string `json:"payLoad"`
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
			if err := m.wsConn.WriteJSON(&wsMessage{
				Event:   "candidate",
				PayLoad: c.ToJSON().Candidate,
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
		switch s {
		case webrtc.PeerConnectionStateFailed:
			if err := m.pConn.Close(); err != nil {
				yalog.Errorf("cannot close peerConnection: %v\n", err)
			}
		case webrtc.PeerConnectionStateClosed:
			yalog.Info("Peer Connection Closed")
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

	if err := m.wsConn.WriteJSON(&wsMessage{
		Event:   "offer",
		PayLoad: offer.SDP,
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

			switch msg.Event {
			case "answer":
				answer := webrtc.SessionDescription{
					Type: webrtc.SDPTypeAnswer,
					SDP:  msg.PayLoad,
				}

				if err := m.pConn.SetRemoteDescription(answer); err != nil {
					yalog.Errorf("set remote description error: %v\n", err)
				}

				candidatesMux.Lock()
				for _, c := range m.pendingCandidates {
					if err := m.wsConn.WriteJSON(&wsMessage{
						Event:   "candidate",
						PayLoad: c.ToJSON().Candidate,
					}); err != nil {
						yalog.Errorf("write message error: %v\n", err)
					}
				}
				candidatesMux.Unlock()

			case "candidate":
				var candidateInit webrtc.ICECandidateInit
				candidateInit.Candidate = msg.PayLoad
				err = m.pConn.AddICECandidate(candidateInit)
				if err != nil {
					yalog.Errorf("add ice candidate error: %v\n", err)
				}

			default:
				yalog.Errorf("Unknown event: %v %v\n", msg.Event, msg.PayLoad)
			}
		}
	}()

	<-interrupt
}
