//go:build windows

package main

import (
	"os"
	"os/signal"

	"github.com/m4n5ter/lindows/internal/capture"
	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/internal/desktop"
	"github.com/m4n5ter/lindows/internal/webrtc"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/spf13/cobra"
)

var serve = &cobra.Command{
	Use:   "serve",
	Short: "Start Lindows service",
	Run:   service.ServeCommand,
}

var service = &Lindows{
	Capture: &config.Capture{},
	Desktop: &config.Desktop{},
	WebRTC:  &config.WebRTC{},
}

type Lindows struct {
	Capture *config.Capture
	Desktop *config.Desktop
	WebRTC  *config.WebRTC

	logger         *yalog.Logger
	captureManager *capture.Manager
	desktopManager *desktop.Manager
	webRTCManager  *webrtc.Manager
}

func (lindows *Lindows) ServeCommand(cmd *cobra.Command, args []string) {
	lindows.logger.Info("Starting Lindows service")
	lindows.Start()
	lindows.logger.Info("Lindows ready to serve")

	quit := make((chan os.Signal), 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	lindows.logger.Info("Shutting down Lindows service", "signal", sig)
	lindows.Stop()
	lindows.logger.Info("Lindows service stopped")
}

func (lindows *Lindows) Start() {
	desktopManager := desktop.New(lindows.Desktop)
	desktopManager.Start()

	captureManager := capture.New(*desktopManager, lindows.Capture)
	captureManager.Start()

	webRTCManager := webrtc.New(lindows.WebRTC)
	webRTCManager.Start()

	lindows.desktopManager = desktopManager
	lindows.captureManager = captureManager
	lindows.webRTCManager = webRTCManager
}

func (lindows *Lindows) Stop() {}

func main() {
	service.logger = yalog.Default().With("service", "lindows")

	configs := []config.Config{
		service.Capture,
		service.Desktop,
		service.WebRTC,
	}

	cobra.OnInitialize(func() {
		for _, cfg := range configs {
			cfg.Set()
		}
	})

	for _, cfg := range configs {
		if err := cfg.Init(serve); err != nil {
			service.logger.Fatal("Failed to initialize config", "error", err)
		}
	}

	root.AddCommand(serve)

	Execute()
}
