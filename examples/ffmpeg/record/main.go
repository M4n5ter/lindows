package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

const (
	h264FrameDuration = time.Millisecond * 33
)

var candidatesMux sync.Mutex

func main() {
	offerAddr := "127.0.0.1:50000"
	answerAddr := ":60000"

	pendingCandidates := make([]*webrtc.ICECandidate, 0)
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.syncthing.net:3478"},
			},
		},
	})
	if err != nil {
		yalog.Error(err)
	}
	defer func() {
		if err := peerConnection.Close(); err != nil {
			yalog.Infof("cannot close peerConnection: %v\n", err)
		}
	}()

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	http.HandleFunc("/candidate", func(w http.ResponseWriter, r *http.Request) {
		candidate, candidateErr := io.ReadAll(r.Body)
		if candidateErr != nil {
			yalog.Error(candidateErr)
		}
		if candidateErr := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: string(candidate)}); candidateErr != nil {
			yalog.Error(candidateErr)
		}
	})

	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		sdp := webrtc.SessionDescription{}
		if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
			yalog.Error(err)
		}

		yalog.Infof("Received SDP: %v\n", sdp)

		if err := peerConnection.SetRemoteDescription(sdp); err != nil {
			yalog.Error(err)
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			yalog.Error(err)
		}

		payload, err := json.Marshal(answer)
		if err != nil {
			yalog.Error(err)
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/sdp", offerAddr), "application/json; charset=utf-8", bytes.NewReader(payload))
		if err != nil {
			yalog.Error(err)
		} else if closeErr := resp.Body.Close(); closeErr != nil {
			yalog.Error(closeErr)
		}

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			yalog.Error(err)
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()
		for _, c := range pendingCandidates {
			if c == nil {
				yalog.Info("c is nil")
			}
			onICECandidateErr := signalCandidate(offerAddr, c)
			if onICECandidateErr != nil {
				yalog.Error(onICECandidateErr)
			}
		}
	})

	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "1")
	if videoTrackErr != nil {
		yalog.Error(videoTrackErr)
	}

	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		yalog.Error(videoTrackErr)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called. NACK 重新传输请求
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		stdout, ffmpegErr := RunCommand("ffmpeg", os.Args[1:]...)
		if err != nil {
			yalog.Error(ffmpegErr)
		}
		if err != nil {
			yalog.Error(err)
		}

		h264, h264Err := h264reader.NewReader(stdout)
		if h264Err != nil {
			yalog.Error(h264Err)
		}

		yalog.Info("等建立连接")

		// Wait for connection established
		<-iceConnectedCtx.Done()

		yalog.Info("开始发送视频")

		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		//
		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		spsAndPpsCache := []byte{}
		ticker := time.NewTicker(h264FrameDuration)
		for ; true; <-ticker.C {
			nal, h264Err := h264.NextNAL()
			if h264Err == io.EOF {
				yalog.Infof("All video frames parsed and sent")
				return
			}
			if h264Err != nil {
				yalog.Error(h264Err)
			}

			nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

			if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
				spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
				continue
			} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
				nal.Data = append(spsAndPpsCache, nal.Data...)
				spsAndPpsCache = []byte{}
			}

			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); h264Err != nil {
				yalog.Error(h264Err)
			}
			yalog.Infof("Sent frame")
		}
	}()

	dataChannel, err := peerConnection.CreateDataChannel("hello", nil)
	if err != nil {
		yalog.Error(err)
	}
	dataChannel.OnOpen(func() {
		yalog.Info("Data channel 'hello'-'open' event")

		// Wait for connection established
		<-iceConnectedCtx.Done()

		ticker := time.NewTicker(5 * time.Second)
		for ; true; <-ticker.C {
			_ = dataChannel.SendText("Hello world")
		}
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		desc := peerConnection.RemoteDescription()
		if desc == nil {
			pendingCandidates = append(pendingCandidates, c)
		} else if onICECandidateErr := signalCandidate(offerAddr, c); onICECandidateErr != nil {
			panic(onICECandidateErr)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		yalog.Infof("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		yalog.Infof("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			yalog.Error("Peer Connection has gone to failed exiting")
		}

		if s == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			yalog.Error("Peer Connection has gone to closed exiting")
		}
	})

	yalog.Error((http.ListenAndServe(answerAddr, nil)))
	select {}
}

func signalCandidate(addr string, c *webrtc.ICECandidate) error {
	payload := []byte(c.ToJSON().Candidate)
	resp, err := http.Post(fmt.Sprintf("http://%s/candidate", addr),
		"application/json; charset=utf-8", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
