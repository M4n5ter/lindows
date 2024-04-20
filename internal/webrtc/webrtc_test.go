package webrtc_test

import (
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/m4n5ter/lindows/internal/config"
	"github.com/m4n5ter/lindows/internal/webrtc"
	w "github.com/pion/webrtc/v4"
)

func TestWebRTC(t *testing.T) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)

	m := webrtc.New(&config.WebRTC{
		ICEServers: []w.ICEServer{
			{
				URLs: []string{"stun:stun.syncthing.net:3478"},
			},
		},
	})
	m.EstablishConn(":8080")

	<-interrupt
}
