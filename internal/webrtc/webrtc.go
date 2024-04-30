package webrtc

import (
	"errors"
	"io"

	"github.com/m4n5ter/lindows/internal/capture"
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
)

type Manager struct {
	logger     *yalog.Logger
	videoTrack *webrtc.TrackLocalStaticRTP
	audioTrack *webrtc.TrackLocalStaticRTP
	capture    capture.Manager
	config     *config.WebRTC
}

func New(cfg *config.WebRTC) *Manager {
	return &Manager{
		logger: yalog.Default().With("module", "webrtc"),
	}
}

func (manager *Manager) Start() {
	var err error

	// Video
	videoCodec := manager.capture.Video().Codec()
	manager.videoTrack, err = webrtc.NewTrackLocalStaticRTP(videoCodec.Capability, "video", "stream")
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

	// Audio
	audioCodec := manager.capture.Audio().Codec()
	manager.audioTrack, err = webrtc.NewTrackLocalStaticRTP(audioCodec.Capability, "audio", "stream")
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

	manager.logger.Info("WebRTC manager started",
		"ice_servers", manager.config.ICEServers,
	)
}
