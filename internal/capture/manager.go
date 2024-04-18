package capture

import (
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/internal/desktop"
	"github.com/m4n5ter/lindows/pkg/yalog"
)

type Manager struct {
	logger  *yalog.Logger
	desktop desktop.Manager
	audio   *StreamManager
	video   *StreamManager
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
	manager.logger.Info("Capture manager started")
}

func (manager *Manager) Audio() StreamManager {
	return *manager.audio
}

func (manager *Manager) Video() StreamManager {
	return *manager.video
}
