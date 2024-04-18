package desktop

import (
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/pkg/yalog"
)

type Manager struct {
	logger                  *yalog.Logger
	shutdown                chan struct{}
	config                  *config.Desktop
	screenSizeChangeChannel chan bool
}

func New(cfg *config.Desktop) *Manager {
	return &Manager{
		logger:                  yalog.Default().With("module", "desktop"),
		shutdown:                make(chan struct{}),
		config:                  cfg,
		screenSizeChangeChannel: make(chan bool),
	}
}

func (manager *Manager) Start() {
}
