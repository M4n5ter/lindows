package capture

import (
	"github.com/m4n5ter/lindows/internal/types/codec"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/rtp"
)

type StreamManager struct {
	logger     *yalog.Logger
	codec      codec.RTPCodec
	rtpChannel chan rtp.Packet
}

func newStreamManager(codec codec.RTPCodec, audioVideoID string) *StreamManager {
	logger := yalog.Default().With(
		"module", "capture",
		"submodule", "stream",
		"audio_video_id", audioVideoID,
	)

	return &StreamManager{
		logger: logger,
		codec:  codec,
	}
}

func (manager *StreamManager) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager StreamManager) GetRTPChannel() chan rtp.Packet {
	return manager.rtpChannel
}
