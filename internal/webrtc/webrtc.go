package webrtc

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m4n5ter/lindows/internal/capture"
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/internal/types/codec"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
)

type Manager struct {
	logger            *yalog.Logger
	videoTrack        *webrtc.TrackLocalStaticRTP
	audioTrack        *webrtc.TrackLocalStaticRTP
	capture           *capture.Manager
	config            *config.WebRTC
	pc                *webrtc.PeerConnection
	wc                *websocket.Conn
	pendingCandidates *pendingCandidates
}

type pendingCandidates struct {
	iCECandidates []*webrtc.ICECandidate
	sync.Mutex
}

type wsMessage struct {
	Event   string `json:"event"`
	PayLoad string `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func New(capture *capture.Manager, cfg *config.WebRTC) *Manager {
	return &Manager{
		logger:            yalog.Default().With("module", "webrtc"),
		pendingCandidates: &pendingCandidates{iCECandidates: make([]*webrtc.ICECandidate, 0)},
		config:            cfg,
		capture:           capture,
	}
}

func (manager *Manager) EstablishConn(addr string) {
	http.HandleFunc("/ws", manager.websocketHandler)
	go func() {
		manager.logger.Fatal("ListenAndServe: ", http.ListenAndServe(addr, nil))
	}()
}

func (manager *Manager) Start() (err error) {
	manager.capture.Start()
	// Video
	manager.videoTrack, err = webrtc.NewTrackLocalStaticRTP(codec.VP8().Capability, "video", "stream")
	if err != nil {
		return err
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

	videoSender, videoErr := manager.pc.AddTrack(manager.videoTrack)
	if videoErr != nil {
		return videoErr
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
	manager.audioTrack, err = webrtc.NewTrackLocalStaticRTP(codec.Opus().Capability, "audio", "stream")
	if err != nil {
		return err
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

	audioSender, audioErr := manager.pc.AddTrack(manager.audioTrack)
	if audioErr != nil {
		return audioErr
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
	return nil
}

func (manager Manager) ReceiveMsg(f func(dc *webrtc.DataChannel)) {
	manager.pc.OnDataChannel(f)
}

func (manager Manager) SendMsg(label string, option *webrtc.DataChannelInit) (chan<- string, <-chan string, func() error, error) {
	sendChan := make(chan string, 10)
	receiveChan := make(chan string, 10)
	dc, err := manager.pc.CreateDataChannel(label, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	dc.OnOpen(func() {
		for {
			msg := <-sendChan
			if err := dc.SendText(msg); err != nil {
				manager.logger.Error("Failed to send message", "error", err)
			}
		}
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			receiveChan <- string(msg.Data)
		}
		manager.logger.Debug("Received message is not string", "data", msg.Data)
	})
	return sendChan, receiveChan, dc.Close, nil
}

func (manager *Manager) Connected() <-chan struct{} {
	connected := make(chan struct{})
	manager.pc.OnICEConnectionStateChange(func(s webrtc.ICEConnectionState) {
		if s == webrtc.ICEConnectionStateConnected {
			close(connected)
		}
	})
	return connected
}

func (manager *Manager) websocketHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	manager.pc, err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: manager.config.ICEServers,
	})
	if err != nil {
		manager.logger.Fatal("Error creating peer connection:", err)
	}
	defer func() {
		if cErr := manager.pc.Close(); cErr != nil {
			manager.logger.Errorf("cannot close peerConnection: %v\n", cErr)
		}
	}()
	manager.wc, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		manager.logger.Fatal(err)
	}
	defer func() {
		if cErr := manager.wc.Close(); cErr != nil {
			manager.logger.Infof("cannot close wsConn: %v\n", cErr)
		}
	}()

	manager.logger.Info("Connected to WebSocket server", "remote_addr", manager.wc.RemoteAddr())

	go func() {
		for range time.NewTicker(5 * time.Second).C {
			if err := manager.wc.WriteJSON(&wsMessage{
				Event:   "ping",
				PayLoad: "ping",
			}); err != nil {
				manager.logger.Errorf("write message error: %v\n", err)
			}
		}
	}()

	manager.pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		manager.pendingCandidates.Lock()
		defer manager.pendingCandidates.Unlock()

		desc := manager.pc.RemoteDescription()
		if desc == nil {
			manager.pendingCandidates.iCECandidates = append(manager.pendingCandidates.iCECandidates, c)
		} else {
			if err := manager.wc.WriteJSON(&wsMessage{
				Event:   "candidate",
				PayLoad: c.ToJSON().Candidate,
			}); err != nil {
				manager.logger.Errorf("write message error: %v\n", err)
			}
		}
	})

	manager.pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		manager.logger.Infof("Peer Connection State has changed: %s\n", s.String())
		switch s {
		case webrtc.PeerConnectionStateFailed:
			if err := manager.pc.Close(); err != nil {
				manager.logger.Errorf("cannot close peerConnection: %v\n", err)
			}
		case webrtc.PeerConnectionStateClosed:
			manager.logger.Info("Peer Connection Closed")
		}
	})

	msg := &wsMessage{}
	for {
		err = manager.wc.ReadJSON(msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) ||
				manager.pc.ConnectionState() == webrtc.PeerConnectionStateClosed {
				manager.wc.Close()
				manager.logger.Infof("wsConn closed: %v\n", err)
				break
			} else {
				manager.logger.Errorf("read message error: %v\n", err)
			}
		}

		switch msg.Event {
		case "offer":
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg.PayLoad,
			}

			if err := manager.pc.SetRemoteDescription(offer); err != nil {
				manager.logger.Errorf("set remote description error: %v\n", err)
			}

			answer, err := manager.pc.CreateAnswer(nil)
			if err != nil {
				manager.logger.Errorf("create answer error: %v\n", err)
			}

			if err := manager.pc.SetLocalDescription(answer); err != nil {
				manager.logger.Errorf("set local description error: %v\n", err)
			}

			if err := manager.wc.WriteJSON(&wsMessage{
				Event:   "answer",
				PayLoad: answer.SDP,
			}); err != nil {
				manager.logger.Errorf("write message error: %v\n", err)
			}

			manager.pendingCandidates.Lock()

			for _, c := range manager.pendingCandidates.iCECandidates {
				if c == nil {
					manager.logger.Info("Candidates  Synchronization complete")
					continue
				}
				if err := manager.wc.WriteJSON(&wsMessage{
					Event:   "candidate",
					PayLoad: c.ToJSON().Candidate,
				}); err != nil {
					manager.logger.Errorf("write message error: %v\n", err)
				}
			}
			manager.pendingCandidates.iCECandidates = nil
			manager.pendingCandidates.Unlock()
		case "answer":
			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  msg.PayLoad,
			}
			if err := manager.pc.SetRemoteDescription(answer); err != nil {
				manager.logger.Errorf("set remote description error: %v\n", err)
			}
		case "candidate":
			var candidate webrtc.ICECandidateInit
			candidate.Candidate = msg.PayLoad
			err = manager.pc.AddICECandidate(candidate)
			if err != nil {
				manager.logger.Errorf("add ice candidate error: %v\n", err)
			}
		case "ping":
			if err := manager.wc.WriteJSON(&wsMessage{
				Event:   "pong",
				PayLoad: "",
			}); err != nil {
				manager.logger.Errorf("write message error: %v\n", err)
			}
		case "pong":
			manager.logger.Debug("pong received")
		default:
			manager.logger.Info("unknown event: %s %s \n", msg.Event, msg.PayLoad)
		}
	}
}
