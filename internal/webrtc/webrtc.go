package webrtc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/m4n5ter/lindows/internal/capture"
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
)

type Manager struct {
	logger            *yalog.Logger
	videoTrack        *webrtc.TrackLocalStaticRTP
	audioTrack        *webrtc.TrackLocalStaticRTP
	capture           capture.Manager
	config            *config.WebRTC
	conn              *webrtc.PeerConnection
	pendingCandidates []*webrtc.ICECandidate
}

var candidatesMux sync.Mutex

func New(cfg *config.WebRTC) *Manager {
	return &Manager{
		logger: yalog.Default().With("module", "webrtc"),
	}
}

func (manager *Manager) Start() {
	var err error

	// Video
	manager.videoTrack, err = webrtc.NewTrackLocalStaticRTP(manager.videoTrack.Codec(), "video", "stream")
	if err != nil {
		manager.logger.Fatal("Failed to create video track", "error", err)
	}

	go func() {
		for {
			packet, ok := <-manager.capture.Video().GetRTPChannel()
			if !ok {
				manager.logger.Debug("Video sample channel closed")
				continue
			}

			if err := manager.videoTrack.WriteRTP(&packet); err != nil && errors.Is(err, io.ErrClosedPipe) {
				manager.logger.Error("Video track closed", "error", err)
			}

		}
	}()

	videoSender, videoErr := manager.conn.AddTrack(manager.videoTrack)
	if videoErr != nil {
		manager.logger.Error("Failed to add video track", "error", videoErr)
	}
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := videoSender.Read(rtcpBuf); rtcpErr != nil {
				manager.logger.Error("Failed to read videp RTCP", "error", rtcpErr)
			}
		}
	}()

	// Audio
	manager.audioTrack, err = webrtc.NewTrackLocalStaticRTP(manager.audioTrack.Codec(), "audio", "stream")
	if err != nil {
		manager.logger.Fatal("Failed to create audio track", "error", err)
	}

	go func() {
		for {
			packet, ok := <-manager.capture.Audio().GetRTPChannel()
			if !ok {
				manager.logger.Debug("Audio sample channel closed")
				continue
			}

			if err := manager.audioTrack.WriteRTP(&packet); err != nil && errors.Is(err, io.ErrClosedPipe) {
				manager.logger.Error("Audio track closed", "error", err)
			}
		}
	}()

	audioSender, audioErr := manager.conn.AddTrack(manager.audioTrack)
	if audioErr != nil {
		manager.logger.Error("Failed to add audio track", "error", audioErr)
	}
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := audioSender.Read(rtcpBuf); rtcpErr != nil {
				manager.logger.Error("Failed to read RTCP", "error", rtcpErr)
			}
		}
	}()

	manager.logger.Info("WebRTC manager started",
		"ice_servers", manager.config.ICEServers,
	)
}

func (manager *Manager) EstablishPeer(answerAddr, offerAddr string) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: manager.config.ICEServers,
	})

	manager.conn = peerConnection

	if err != nil {
		manager.logger.Errorf("cannot create peer connection: %v\n", err.Error())
	}

	helloDataChannel, err := manager.conn.CreateDataChannel("hello", nil)
	if err != nil {
		manager.logger.Error(err)
	}
	helloDataChannel.OnOpen(func() {
		manager.logger.Info("Data channel '%s'-'%d' open.Say hello", helloDataChannel.Label(), helloDataChannel.ID())
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
			manager.pendingCandidates = append(manager.pendingCandidates, c)
		} else if onICECandidateErr := signalCandidate(offerAddr, c); onICECandidateErr != nil {
			panic(onICECandidateErr)
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		manager.logger.Debugf("Connection State has changed %s \n", connectionState.String())
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		manager.logger.Infof("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed {
			manager.logger.Error("Peer Connection has gone to failed exiting")
		}

		if s == webrtc.PeerConnectionStateClosed {
			manager.logger.Error("Peer Connection has gone to closed exiting")
		}
	})

	http.HandleFunc("/candidate", func(w http.ResponseWriter, r *http.Request) {
		candidate, candidateErr := io.ReadAll(r.Body)
		if candidateErr != nil {
			manager.logger.Error(candidateErr.Error())
		}
		if candidateErr := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: string(candidate)}); candidateErr != nil {
			manager.logger.Error(candidateErr)
		}
	})

	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		sdp := webrtc.SessionDescription{}
		if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
			manager.logger.Error(err.Error())
		}

		manager.logger.Debugf("Received SDP: %v\n", sdp)

		if err := peerConnection.SetRemoteDescription(sdp); err != nil {
			manager.logger.Error(err)
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			manager.logger.Error(err)
		}

		payload, err := json.Marshal(answer)
		if err != nil {
			manager.logger.Error(err)
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/sdp", offerAddr), "application/json; charset=utf-8", bytes.NewReader(payload))
		if err != nil {
			manager.logger.Error(err)
		} else if closeErr := resp.Body.Close(); closeErr != nil {
			manager.logger.Error(closeErr)
		}

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			manager.logger.Error(err)
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()
		for _, c := range manager.pendingCandidates {
			if c == nil {
				manager.logger.Info("c is nil")
			}
			onICECandidateErr := signalCandidate(offerAddr, c)
			if onICECandidateErr != nil {
				manager.logger.Error(onICECandidateErr)
			}
		}
	})
	go func() { manager.logger.Error((http.ListenAndServe(answerAddr, nil).Error())) }()
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
