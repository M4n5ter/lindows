package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"

	"github.com/m4n5ter/lindows/pkg/ffmpeg"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/rtp"
)

func main() {
	// 打开一个 udp 端口，用于接收 ffmpeg 的 rtp 输出
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	},
	)
	if err != nil {
		yalog.Fatalf("failed to listen udp: %v", err)
	}

	bufferSize := 300 << 10 // 300 KB
	err = listener.SetReadBuffer(bufferSize)
	if err != nil {
		yalog.Fatalf("failed to set read buffer: %v", err)
	}

	defer listener.Close()

	rtpPort := listener.LocalAddr().(*net.UDPAddr).Port
	yalog.Infof("listening on udp port: %d", rtpPort)

	// 启动 ffmpeg
	go startFFmpeg(rtpPort, "desktop")

	inboundRTPPacket := make([]byte, 1600) // UDP MTU

	for {
		n, _, err := listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			yalog.Fatalf("failed to read from udp: %v", err)
		}

		// 将 rtp 包转换为 rtp.Packet
		packet := &rtp.Packet{}
		err = packet.Unmarshal(inboundRTPPacket[:n])
		if err != nil {
			yalog.Fatalf("failed to unmarshal rtp packet: %v", err)
		}

		yalog.Info("Parse rtp packet successful", "packet", packet)

	}
}

func startFFmpeg(rtpPort int, input string) {
	// 创建 ffmpeg 命令
	ffmpegPath, delFunc, err := ffmpeg.TempFFmpeg()
	if err != nil {
		yalog.Fatalf("failed to create temp ffmpeg: %v", err)
	}
	defer delFunc()

	rtpURL := fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", rtpPort)

	cmd := exec.Command(ffmpegPath,
		"-re",
		// "-f", "lavfi", "-i", "testsrc=size=640x480:rate=30",
		"-f", "gdigrab", "-i", input,
		"-vcodec", "libvpx", "-cpu-used", "5",
		"-deadline", "1", "-g", "10",
		"-error-resilient", "1", "-auto-alt-ref", "1",
		"-f", "rtp", rtpURL)

	// 启动 ffmpeg 命令
	err = cmd.Start()
	if err != nil {
		yalog.Fatalf("failed to start ffmpeg: %v", err)
	}

	yalog.Info("ffmpeg started",
		"pid", cmd.Process.Pid,
		"rtp_url", rtpURL)

	// 捕捉退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		sig := <-c // 接收信号

		// 退出时发送SIGTERM给子进程，命令其退出
		if err := cmd.Process.Signal(sig); err != nil {
			yalog.Fatalf("failed to send signal to proccess: %v", err)
		}
	}()
	cmd.Wait()
}
