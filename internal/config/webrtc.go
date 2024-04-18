package config

import "github.com/pion/webrtc/v4"

type WebRTC struct {
	ICEServers []webrtc.ICEServer
}
