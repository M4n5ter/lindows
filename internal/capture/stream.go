package capture

import (
	"io"
	"os/exec"
	"time"

	"github.com/m4n5ter/lindows/internal/types/codec"
	"github.com/m4n5ter/lindows/pkg/ffmpeg"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

type StreamManager struct {
	logger        *yalog.Logger
	codec         codec.RTPCodec
	streamChannel chan *media.Sample
}

const (
	h264FrameDuration = time.Millisecond * 33
)

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

func (manager *StreamManager) GetStreamChannel() chan *media.Sample {
	return manager.streamChannel
}

func (manager *StreamManager) SetStreamChannel(format string) error {
	// ffmpeg -rtbufsize 100M -f dshow -i video="screen-capture-recorder" -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb  -preset ultrafast   -tune fastdecode    -b:v 1M -max_delay 0 -bf 0 -f h264 -
	var cmd *exec.Cmd
	manager.streamChannel = make(chan *media.Sample, 1000)
	path, clean, err := ffmpeg.TempFFmpeg()
	if err != nil {
		return err
	}
	switch format {
	case "video":
		cmd = exec.Command(path, "-rtbufsize", "100M", "-f", "dshow", "-i", "video=screen-capture-recorder", "-pix_fmt", "yuv420p", "-c:v", "libx264", "-bsf:v", "h264_mp4toannexb", "-preset", "ultrafast", "-tune", "fastdecode", "-b:v", "1M", "-max_delay", "0", "-bf", "0", "-f", "h264", "-")
		err = cmd.Start()
		if err != nil {
			return err
		}

		dataPipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		go func() {
			h264, err := h264reader.NewReader(dataPipe)
			if err != nil {
				manager.logger.Fatal(err)
			}

			spsAndPpsCache := []byte{}
			ticker := time.NewTicker(h264FrameDuration)
			for ; true; <-ticker.C {
				nal, err := h264.NextNAL()
				if err == io.EOF {
					manager.logger.Infof("%s EOF", format)
					break
				}
				if err != nil {
					manager.logger.Error(err)
				}

				nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)
				if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
					spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
					continue
				} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
					nal.Data = append(spsAndPpsCache, nal.Data...)
					spsAndPpsCache = []byte{}
				}

				manager.streamChannel <- &media.Sample{Data: nal.Data, Duration: time.Second}
			}
		}()

	case "audio":
		// TODO audio
		cmd = exec.Command(path)
		err = cmd.Start()
		if err != nil {
			return err
		}

		_, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
	}

	go func() {
		time.Sleep(10 * time.Second)
		clean()
	}()
	return nil
}

// func (manager *StreamManager) SetRTPChannel(container, target, options string) (cleanFunc func()) {
// 	listener, err := net.ListenUDP("udp", &net.UDPAddr{
// 		IP:   net.ParseIP("127.0.0.1"),
// 		Port: 0,
// 	})
// 	if err != nil {
// 		manager.logger.Fatalf("failed to listen udp: %v", err)
// 	}

// 	// ffmpeg输出数据: bitrate=2513.9kbits/s 需要2513.9/8=314.2375KB/s
// 	bufferSize := 375 << 10
// 	err = listener.SetReadBuffer(bufferSize)
// 	if err != nil {
// 		manager.logger.Fatalf("failed to set read buffer: %v", err)
// 	}

// 	rtpPort := listener.LocalAddr().(*net.UDPAddr).Port
// 	manager.logger.Infof("listening on udp port: %d", rtpPort)

// 	delFunc := manager.StartFFmpeg(container, target, options, rtpPort)

// 	rtpChannel := make(chan rtp.Packet, 1000)
// 	buffer := make([]byte, 1500)
// 	go func() {
// 		for {
// 			n, _, err := listener.ReadFrom(buffer)
// 			if err != nil {
// 				if strings.Contains(err.Error(), "use of closed network connection") {
// 					break
// 				}
// 				manager.logger.Fatalf("failed to read from udp: %v", err)
// 			}

// 			packet := rtp.Packet{}
// 			err = packet.Unmarshal(buffer[:n])
// 			if err != nil {
// 				manager.logger.Fatalf("failed to unmarshal rtp packet: %v", err)
// 			}
// 			rtpChannel <- packet
// 		}
// 	}()

// 	return func() {
// 		delFunc()
// 		listener.Close()
// 	}
// }

// func (manager *StreamManager) StartFFmpeg(container, input, options string, rtpPort int) func() {
// 	ffmpegPath, delFunc, err := ffmpeg.TempFFmpeg()
// 	if err != nil {
// 		manager.logger.Fatalf("failed to create temp ffmpeg: %v", err)
// 	}

// 	rtpURL := fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", rtpPort)

// 	cmd := exec.Command(ffmpegPath,
// 		"-re",
// 		// "-f", "lavfi", "-i", "testsrc=size=640x480:rate=30",
// 		"-f", container, "-i", input,
// 		options, "-cpu-used", "5",
// 		"-g", "10",
// 		"-error-resilient", "1", "-auto-alt-ref", "1",
// 		"-f", "rtp", rtpURL)

// 	err = cmd.Start()
// 	if err != nil {
// 		manager.logger.Fatalf("failed to start ffmpeg: %v", err)
// 	}

// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

// 	go func() {
// 		sig := <-c

// 		if err := cmd.Process.Signal(sig); err != nil {
// 			yalog.Fatalf("failed to send signal to proccess: %v", err)
// 		}
// 	}()

// 	manager.logger.Info("ffmpeg started",
// 		"pid", cmd.Process.Pid,
// 		"rtp_url", rtpURL)

// 	return delFunc
// }
