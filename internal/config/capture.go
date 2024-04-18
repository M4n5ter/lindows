package config

import (
	"strings"

	"github.com/m4n5ter/lindows/internal/types/codec"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type HwEnc int

const (
	HwEncNone HwEnc = iota
	HwEncVAAPI
	HwEncNVENC
)

type Capture struct {
	// Video
	Display      string
	VideoCodec   codec.RTPCodec
	VideoHwEnc   HwEnc
	VideoBitrate uint
	VideoMaxFPS  int16

	// Audio
	AudioDevice  string
	AudioCodec   codec.RTPCodec
	AudioBitrate uint
}

func (Capture) Init(cmd *cobra.Command) error {
	// Video
	cmd.PersistentFlags().String("display", "desktop", "要捕获的显示器")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video_codec", "vp8", "视频编解码器")
	if err := viper.BindPFlag("video_codec", cmd.PersistentFlags().Lookup("video_codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("hwenc", "", "硬件编码器")
	if err := viper.BindPFlag("hwenc", cmd.PersistentFlags().Lookup("hwenc")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("video_bitrate", 3072, "视频比特率, 单位 kbps")
	if err := viper.BindPFlag("video_bitrate", cmd.PersistentFlags().Lookup("video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("max_fps", 25, "通过WEBRTC传递的最大fps, 0 表示不限制")
	if err := viper.BindPFlag("max_fps", cmd.PersistentFlags().Lookup("max_fps")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video", "", "用于流的视频编解码器参数")
	if err := viper.BindPFlag("video", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	// Audio
	cmd.PersistentFlags().String("device", "audio_output.monitor", "要捕获的音频设备")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio_codec", "opus", "音频编解码器")
	if err := viper.BindPFlag("audio_codec", cmd.PersistentFlags().Lookup("audio_codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("audio_bitrate", 128, "音频比特率, 单位 kbps")
	if err := viper.BindPFlag("audio_bitrate", cmd.PersistentFlags().Lookup("audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("audio", "", "用于流的音频编解码器参数")
	err := viper.BindPFlag("audio", cmd.PersistentFlags().Lookup("audio"))
	return err
}

func (s *Capture) Set() {
	var ok bool
	// Video
	s.Display = viper.GetString("display")

	videoCodec := viper.GetString("video_codec")
	s.VideoCodec, ok = codec.ParseStr(videoCodec)
	if !ok || s.VideoCodec.Type != webrtc.RTPCodecTypeVideo {
		yalog.Error("无效的视频编解码器，改为 Vp8", "codec", videoCodec)
		s.VideoCodec = codec.VP8()
	}

	videoHWEnc := strings.ToLower(viper.GetString("hwenc"))
	switch videoHWEnc {
	case "":
		fallthrough
	case "none":
		s.VideoHwEnc = HwEncNone
	case "vaapi":
		s.VideoHwEnc = HwEncVAAPI
	case "nvenc":
		s.VideoHwEnc = HwEncNVENC
	default:
		yalog.Error("无效的硬件编码器，将使用 CPU", "hwenc", videoHWEnc)
	}

	s.VideoBitrate = uint(viper.GetInt("video_bitrate"))
	s.VideoMaxFPS = int16(viper.GetInt("max_fps"))

	// Audio
	s.AudioDevice = viper.GetString("device")

	audioCodec := viper.GetString("audio_codec")
	s.AudioCodec, ok = codec.ParseStr(audioCodec)
	if !ok || s.AudioCodec.Type != webrtc.RTPCodecTypeAudio {
		yalog.Error("无效的音频编解码器，改为 Opus", "codec", audioCodec)
		s.AudioCodec = codec.Opus()
	}

	s.AudioBitrate = uint(viper.GetInt("audio_bitrate"))
}
