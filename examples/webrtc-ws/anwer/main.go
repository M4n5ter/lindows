package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	err           error
	candidatesMux sync.Mutex
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)

	m := Manager{
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	m.pendingCandidates = make([]*webrtc.ICECandidate, 0)

	m.pConn, err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.syncthing.net:3478"},
			},
		},
	})
	if err != nil {
		yalog.Error(err)
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

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.wsConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			yalog.Fatal(err)
		}
		go func() {
			process(m.wsConn, m.pConn, m.pendingCandidates)
		}()
	})

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	dataChannel, err := m.pConn.CreateDataChannel("hello", nil)
	if err != nil {
		yalog.Error(err)
	}

	dataChannel.OnOpen(func() {
		yalog.Info("Data channel 'hello'-'open' event")

		// Wait for connection established
		<-iceConnectedCtx.Done()

		for range time.NewTicker(1 * time.Second).C {
			err = dataChannel.SendText("Hello world")
			if err != nil {
				yalog.Error("SendText error: ", err)
			}
			yalog.Info("Sent 'Hello world' to data channel")
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	m.pConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		yalog.Infof("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	m.pConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		yalog.Infof("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			yalog.Error("Peer Connection has gone to failed exiting")
		}

		if s == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			yalog.Error("Peer Connection has gone to closed exiting")
			// remoteErr := m.pConn.SetRemoteDescription(webrtc.SessionDescription{})
			// localErr := m.pConn.SetLocalDescription(webrtc.SessionDescription{})
			// if remoteErr != nil || localErr != nil {
			// 	yalog.Error("SetRemoteDescription error:", err)
			// }
		}
	})

	yalog.Info("http服务器在 :8080 端口上启动...")

	go func() {
		yalog.Fatal("ListenAndServe: ", http.ListenAndServe(":8080", nil))
	}()

	<-interrupt

	yalog.Info("关闭连接...")
}

func process(ws *websocket.Conn, pConn *webrtc.PeerConnection, pendingCandidates []*webrtc.ICECandidate) {
	for {
		msg := &wsMessage{}
		err = ws.ReadJSON(msg)
		if err != nil {
			yalog.Errorf("read message error: %v\n", err)
			ws.Close()
			return
		}
		if msg.Event == "offer" {
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(msg.Data), &offer); err != nil {
				yalog.Errorf("unmarshal offer error: %v\n", err)
			}

			if err := pConn.SetRemoteDescription(offer); err != nil {
				yalog.Errorf("set remote description error: %v\n", err)
			}
			answer, err := pConn.CreateAnswer(nil)
			if err != nil {
				yalog.Errorf("create answer error: %v\n", err)
			}
			if err := pConn.SetLocalDescription(answer); err != nil {
				yalog.Errorf("set local description error: %v\n", err)
			}
			answerData, err := json.Marshal(answer)
			if err != nil {
				yalog.Errorf("marshal answer error: %v\n", err)
			}
			if err := ws.WriteJSON(&wsMessage{
				Event: "answer",
				Data:  string(answerData),
			}); err != nil {
				yalog.Errorf("write message error: %v\n", err)
			}
			candidatesMux.Lock()
			for _, c := range pendingCandidates {
				if c == nil {
					yalog.Info("c is nil")
				}
				candidateData, err := json.Marshal(c)
				if err != nil {
					yalog.Errorf("marshal answer error: %v\n", err)
				}
				if err := ws.WriteJSON(&wsMessage{
					Event: "candidate",
					Data:  string(candidateData),
				}); err != nil {
					yalog.Errorf("write message error: %v\n", err)
				}
			}
			candidatesMux.Unlock()
		} else if msg.Event == "candidate" {
			// if pConn.RemoteDescription() != nil {
			var candidate webrtc.ICECandidate
			err := json.Unmarshal([]byte(msg.Data), &candidate)
			if err != nil {
				yalog.Errorf("parse ice candidate error: %v\n", err)
			}
			err = pConn.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate.ToJSON().Candidate})
			if err != nil {
				yalog.Errorf("add ice candidate error: %v\n", err)
			}
			// }
		} else {
			yalog.Errorf("unknown event: %s %s \n", msg.Event, msg.Data)
		}
	}
}
