package capture

import (
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/internal/desktop"
	"github.com/m4n5ter/lindows/pkg/yalog"
)

type Manager struct {
	logger     *yalog.Logger
	desktop    desktop.Manager
	audio      *StreamManager
	video      *StreamManager
	audioClean func()
	videoClean func()
}

func New(desktop desktop.Manager, cfg *config.Capture) *Manager {
	return &Manager{
		logger:  yalog.Default().With("module", "capture"),
		desktop: desktop,
		audio:   newStreamManager(cfg.AudioCodec, "audio"),
		video:   newStreamManager(cfg.VideoCodec, "video"),
	}
}

func (manager *Manager) Start() {
	manager.videoClean = manager.video.SetRTPChannel("gdigrab", "desktop", "-vcodec libvpx")
	manager.audioClean = manager.audio.SetRTPChannel("dshow", "audio", "-f dshow -i audio=virtual-audio-capturer")
	manager.logger.Info("Capture manager started")
}

func (manager *Manager) Audio() StreamManager {
	return *manager.audio
}

func (manager *Manager) Video() StreamManager {
	return *manager.video
}

func (manager *Manager) Close() {
	manager.audioClean()
	manager.videoClean()
}
