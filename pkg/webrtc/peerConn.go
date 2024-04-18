package webrtc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v3"
)

type PeerConnection struct {
	yalog.Logger
	Conn              *webrtc.PeerConnection
	pendingCandidates []*webrtc.ICECandidate
	Close             func() error
}

var candidatesMux sync.Mutex

func NewPeerConnection(config webrtc.Configuration, answerAddr, offerAddr string) (*PeerConnection, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.syncthing.net:3478"},
			},
		},
	})
	pc := &PeerConnection{
		Conn:              peerConnection,
		Close:             peerConnection.Close,
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	if err != nil {
		pc.Logger.Errorf("cannot create peer connection: %v\n", err.Error())
	}

	helloDataChannel, err := pc.Conn.CreateDataChannel("hello", nil)
	if err != nil {
		pc.Logger.Error(err)
	}
	helloDataChannel.OnOpen(func() {
		pc.Logger.Info("Data channel '%s'-'%d' open.Say hello", helloDataChannel.Label(), helloDataChannel.ID())
		sendTextErr := helloDataChannel.SendText("Hello, World!")
		if sendTextErr != nil {
			panic(sendTextErr)
		}
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		desc := peerConnection.RemoteDescription()
		if desc == nil {
			pc.pendingCandidates = append(pc.pendingCandidates, c)
		} else if onICECandidateErr := signalCandidate(offerAddr, c); onICECandidateErr != nil {
			panic(onICECandidateErr)
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		pc.Logger.Debugf("Connection State has changed %s \n", connectionState.String())
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		pc.Logger.Infof("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed {
			pc.Logger.Error("Peer Connection has gone to failed exiting")
		}

		if s == webrtc.PeerConnectionStateClosed {
			pc.Logger.Error("Peer Connection has gone to closed exiting")
		}
	})

	http.HandleFunc("/candidate", func(w http.ResponseWriter, r *http.Request) {
		candidate, candidateErr := io.ReadAll(r.Body)
		if candidateErr != nil {
			pc.Logger.Error(candidateErr.Error())
		}
		if candidateErr := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: string(candidate)}); candidateErr != nil {
			pc.Logger.Error(candidateErr)
		}
	})

	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		sdp := webrtc.SessionDescription{}
		if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
			pc.Logger.Error(err.Error())
		}

		pc.Logger.Debugf("Received SDP: %v\n", sdp)

		if err := peerConnection.SetRemoteDescription(sdp); err != nil {
			pc.Logger.Error(err)
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			pc.Logger.Error(err)
		}

		payload, err := json.Marshal(answer)
		if err != nil {
			pc.Logger.Error(err)
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/sdp", offerAddr), "application/json; charset=utf-8", bytes.NewReader(payload))
		if err != nil {
			pc.Logger.Error(err)
		} else if closeErr := resp.Body.Close(); closeErr != nil {
			pc.Logger.Error(closeErr)
		}

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			pc.Logger.Error(err)
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()
		for _, c := range pc.pendingCandidates {
			if c == nil {
				pc.Logger.Info("c is nil")
			}
			onICECandidateErr := signalCandidate(offerAddr, c)
			if onICECandidateErr != nil {
				pc.Logger.Error(onICECandidateErr)
			}
		}
	})
	go func() { pc.Logger.Error((http.ListenAndServe(answerAddr, nil).Error())) }()
	return pc, nil
}

func (pc *PeerConnection) CreateDataChannel(dataChan chan []byte, label string, options *webrtc.DataChannelInit) chan []byte {
	var reply chan []byte
	dataChannel, err := pc.Conn.CreateDataChannel(label, options)
	if err != nil {
		pc.Logger.Error(err)
	}
	dataChannel.OnOpen(func() {
		pc.Logger.Infof("Data channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID())
		for _, b := range <-dataChan {
			err := dataChannel.Send([]byte{b})
			if err != nil {
				pc.Logger.Error(err)
			}
		}
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		reply <- msg.Data
	})
	return reply
}

func signalCandidate(addr string, c *webrtc.ICECandidate) error {
	payload := []byte(c.ToJSON().Candidate)
	resp, err := http.Post(fmt.Sprintf("http://%s/candidate", addr),
		"application/json; charset=utf-8", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
