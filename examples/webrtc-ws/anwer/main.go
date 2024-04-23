package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

const (
	h264FrameDuration = time.Millisecond * 33
)

type Manager struct {
	wsConn            *websocket.Conn
	pConn             *webrtc.PeerConnection
	pendingCandidates []*webrtc.ICECandidate
}

type wsMessage struct {
	Event   string `json:"event"`
	PayLoad string `json:"payload"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	err           error
	candidatesMux sync.Mutex
)

// go run . -rtbufsize 100M -f dshow -i video="screen-capture-recorder" -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb  -preset ultrafast   -tune fastdecode    -b:v 1M -max_delay 0 -bf 0 -f h264 -
func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m := Manager{}
		m.wsConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			yalog.Fatal(err)
		}
		// // TODO
		// go func() {
		// 	for range time.NewTicker(2 * time.Second).C {
		// 		if err := m.wsConn.WriteJSON(&wsMessage{
		// 			Event:   "ping",
		// 			PayLoad: "ping",
		// 		}); err != nil {
		// 			yalog.Errorf("write message error: %v\n", err)
		// 		}
		// 	}
		// }()

		iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

		m.pendingCandidates = make([]*webrtc.ICECandidate, 0)

		m.pConn, err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.syncthing.net:3478"},
				},
			},
		})
		if err != nil {
			yalog.Error(err)
		}

		m.pConn.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {
				return
			}

			candidatesMux.Lock()
			defer candidatesMux.Unlock()

			desc := m.pConn.RemoteDescription()
			if desc == nil {
				m.pendingCandidates = append(m.pendingCandidates, c)
			} else {
				if err := m.wsConn.WriteJSON(&wsMessage{
					Event:   "candidate",
					PayLoad: c.ToJSON().Candidate,
				}); err != nil {
					yalog.Errorf("write message error: %v\n", err)
				}
			}
		})

		dataChannel, err := m.pConn.CreateDataChannel("hello", nil)
		if err != nil {
			yalog.Error(err)
		}

		dataChannel.OnOpen(func() {
			yalog.Info("Data channel 'hello'-'open' event")

			// Wait for connection established
			<-iceConnectedCtx.Done()

			// for range time.NewTicker(1 * time.Second).C {
			// 	err = dataChannel.SendText("Hello world")
			// 	if err != nil {
			// 		yalog.Error("SendText error: ", err)
			// 	}
			// 	// yalog.Info("Sent 'Hello world' to data channel")
			// }
		})

		go func() {
			yalog.Info("等待连接...后添加轨道")
			<-iceConnectedCtx.Done()
			dataPipe, err := runCommand("ffmpeg", os.Args[1:]...)
			if err != nil {
				yalog.Error(err)
			}

			videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
			if videoTrackErr != nil {
				yalog.Error(videoTrackErr)
			}

			rtpSender, videoTrackErr := m.pConn.AddTrack(videoTrack)
			if videoTrackErr != nil {
				yalog.Error(videoTrackErr)
			}

			offer, err := m.pConn.CreateOffer(nil)
			if err != nil {
				yalog.Fatal("Error creating offer:", err)
			}
			err = m.pConn.SetLocalDescription(offer)
			if err != nil {
				yalog.Error("Error setting local description:", err)
			}

			if err := m.wsConn.WriteJSON(&wsMessage{
				Event:   "offer",
				PayLoad: offer.SDP,
			}); err != nil {
				yalog.Fatal("Error sending offer:", err)
			}

			go func() {
				rtcpBuf := make([]byte, 1500)
				for {
					if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
						return
					}
				}
			}()

			h264, h264Err := h264reader.NewReader(dataPipe)
			if h264Err != nil {
				panic(h264Err)
			}

			spsAndPpsCache := []byte{}
			ticker := time.NewTicker(h264FrameDuration)
			for ; true; <-ticker.C {
				nal, h264Err := h264.NextNAL()
				if h264Err == io.EOF {
					os.Exit(0)
				}
				if h264Err != nil {
					panic(h264Err)
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
					panic(h264Err)
				}
			}
		}()

		// Set the handler for ICE connection state
		// This will notify you when the peer has connected/disconnected
		m.pConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			yalog.Infof("Connection State has changed %s \n", connectionState.String())
			if connectionState == webrtc.ICEConnectionStateConnected {
				iceConnectedCtxCancel()
			}
		})

		// Set the handler for Peer connection state
		// This will notify you when the peer has connected/disconnected
		m.pConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
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
				if cErr := m.pConn.Close(); cErr != nil {
					yalog.Errorf("cannot close peerConnection: %v\n", cErr)
				}
				// remoteErr := m.pConn.SetRemoteDescription(webrtc.SessionDescription{})
				// localErr := m.pConn.SetLocalDescription(webrtc.SessionDescription{})
				// if remoteErr != nil || localErr != nil {
				// 	yalog.Error("SetRemoteDescription error:", err)
				// }
			}
		})

		go func() {
			process(m.wsConn, m.pConn, m.pendingCandidates)
		}()
	})

	yalog.Info("http服务器在 :8080 端口上启动...")

	go func() {
		yalog.Fatal("ListenAndServe: ", http.ListenAndServe(":8080", nil))
	}()

	<-interrupt

	yalog.Info("关闭连接...")
}

func process(ws *websocket.Conn, pConn *webrtc.PeerConnection, pendingCandidates []*webrtc.ICECandidate) {
	for {
		msg := &wsMessage{}
		err = ws.ReadJSON(msg)
		if err != nil {
			yalog.Errorf("read message error: %v\n", err)
			ws.Close()
			return
		}
		switch msg.Event {
		case "answer":
			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  msg.PayLoad,
			}
			if err := pConn.SetRemoteDescription(answer); err != nil {
				yalog.Errorf("set remote description error: %v\n", err)
			}

		case "offer":
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg.PayLoad,
			}
			if err := pConn.SetRemoteDescription(offer); err != nil {
				yalog.Errorf("set remote description error: %v\n", err)
			}
			answer, err := pConn.CreateAnswer(nil)
			if err != nil {
				yalog.Errorf("create answer error: %v\n", err)
			}
			if err := pConn.SetLocalDescription(answer); err != nil {
				yalog.Errorf("set local description error: %v\n", err)
			}
			if err := ws.WriteJSON(&wsMessage{
				Event:   "answer",
				PayLoad: answer.SDP,
			}); err != nil {
				yalog.Errorf("write message error: %v\n", err)
			}
			candidatesMux.Lock()
			for _, c := range pendingCandidates {
				if err := ws.WriteJSON(&wsMessage{
					Event:   "candidate",
					PayLoad: c.ToJSON().Candidate,
				}); err != nil {
					yalog.Errorf("write message error: %v\n", err)
				}
			}
			candidatesMux.Unlock()

		case "candidate":
			var candidateInit webrtc.ICECandidateInit
			candidateInit.Candidate = msg.PayLoad
			err = pConn.AddICECandidate(candidateInit)
			if err != nil {
				yalog.Errorf("add ice candidate error: %v\n", err)
			}
		case "pong":
			yalog.Debug("pong")
		default:
			if msg.Event != "" {
				yalog.Errorf("unknown event: %s %s \n", msg.Event, msg.PayLoad)
			}
		}
	}
}

func runCommand(name string, arg ...string) (io.ReadCloser, error) {
	cmd := exec.Command(name, arg...)

	dataPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return dataPipe, nil
}
