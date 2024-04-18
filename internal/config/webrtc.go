package config

import (
	"github.com/pion/webrtc/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WebRTC struct {
	ICEServers []webrtc.ICEServer
}

func (WebRTC) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringSlice("ice_servers", []string{"stun:stun.syncthing.net:3478"}, "ICE服务器")
	err := viper.BindPFlag("ice_servers", cmd.PersistentFlags().Lookup("ice_servers"))

	return err
}

func (s *WebRTC) Set() {
	iceServers := viper.GetStringSlice("ice_servers")
	for _, server := range iceServers {
		s.ICEServers = append(s.ICEServers, webrtc.ICEServer{
			URLs: []string{server},
		})
	}
}
