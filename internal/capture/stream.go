package capture

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/m4n5ter/lindows/internal/types/codec"
	"github.com/m4n5ter/lindows/pkg/ffmpeg"
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

func (manager *StreamManager) SetRTPChannel(audioVideoID string) (rtpChannel chan rtp.Packet, udpClose func(), delFunc func()) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		manager.logger.Fatalf("failed to listen udp: %v", err)
	}

	bufferSize := 300 << 10
	err = listener.SetReadBuffer(bufferSize)
	if err != nil {
		manager.logger.Fatalf("failed to set read buffer: %v", err)
	}

	rtpPort := listener.LocalAddr().(*net.UDPAddr).Port
	manager.logger.Infof("listening on udp port: %d", rtpPort)

	delFunc = manager.StartFFmpeg(rtpPort, audioVideoID)

	rtpChannel = make(chan rtp.Packet, 1000)
	buffer := make([]byte, 1600)
	go func() {
		for {
			n, _, err := listener.ReadFrom(buffer)
			if err != nil {
				manager.logger.Fatalf("failed to read from udp: %v", err)
			}

			packet := rtp.Packet{}
			err = packet.Unmarshal(buffer[:n])
			if err != nil {
				manager.logger.Fatalf("failed to unmarshal rtp packet: %v", err)
			}
			rtpChannel <- packet
		}
	}()

	return rtpChannel, func() {
		listener.Close()
	}, delFunc
}

func (manager *StreamManager) StartFFmpeg(rtpPort int, input string) func() {
	ffmpegPath, delFunc, err := ffmpeg.TempFFmpeg()
	if err != nil {
		manager.logger.Fatalf("failed to create temp ffmpeg: %v", err)
	}

	rtpURL := fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", rtpPort)

	cmd := exec.Command(ffmpegPath,
		"-re",
		// "-f", "lavfi", "-i", "testsrc=size=640x480:rate=30",
		"-f", "gdigrab", "-i", input,
		"-vcodec", "libvpx", "-cpu-used", "5",
		"-g", "10",
		"-error-resilient", "1", "-auto-alt-ref", "1",
		"-f", "rtp", rtpURL)

	err = cmd.Start()
	if err != nil {
		manager.logger.Fatalf("failed to start ffmpeg: %v", err)
	}

	manager.logger.Info("ffmpeg started",
		"pid", cmd.Process.Pid,
		"rtp_url", rtpURL)

	return delFunc
}
