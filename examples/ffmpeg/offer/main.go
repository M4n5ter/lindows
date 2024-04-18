// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// pion-to-pion is an example of two pion instances communicating directly!
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/m4n5ter/lindows/pkg/yalog"
	"github.com/pion/webrtc/v4"
)

func signalCandidate(addr string, c *webrtc.ICECandidate) error {
	payload := []byte(c.ToJSON().Candidate)
	resp, err := http.Post(fmt.Sprintf("http://%s/candidate", addr), "application/json; charset=utf-8", bytes.NewReader(payload)) //nolint:noctx
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

func main() {
	offerAddr := flag.String("offer-address", ":50000", "Address that the Offer HTTP server is hosted on.")
	answerAddr := flag.String("answer-address", "127.0.0.1:60000", "Address that the Answer HTTP server is hosted on.")
	flag.Parse()

	var candidatesMux sync.Mutex
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		yalog.Error(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			yalog.Infof("cannot close peerConnection: %v\n", cErr)
		}
	}()

	// When an ICE candidate is available send to the other Pion instance
	// the other Pion instance will add this candidate by calling AddICECandidate
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		desc := peerConnection.RemoteDescription()
		if desc == nil {
			pendingCandidates = append(pendingCandidates, c)
		} else if onICECandidateErr := signalCandidate(*answerAddr, c); onICECandidateErr != nil {
			yalog.Error(onICECandidateErr)
		}
	})

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

	// A HTTP handler that processes a SessionDescription given to us from the other Pion process
	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		sdp := webrtc.SessionDescription{}
		if sdpErr := json.NewDecoder(r.Body).Decode(&sdp); sdpErr != nil {
			yalog.Error(sdpErr)
		}

		if sdpErr := peerConnection.SetRemoteDescription(sdp); sdpErr != nil {
			yalog.Error(sdpErr)
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		for _, c := range pendingCandidates {
			if onICECandidateErr := signalCandidate(*answerAddr, c); onICECandidateErr != nil {
				yalog.Error(onICECandidateErr)
			}
		}
	})

	go func() { yalog.Error(http.ListenAndServe(*offerAddr, nil)) }()

	// Create a datachannel with label 'data'
	dataChannel, err := peerConnection.CreateDataChannel("hello", nil)
	if err != nil {
		yalog.Error(err)
	}

	// // Register channel opening handling
	// dataChannel.OnOpen(func() {
	// 	yalog.Infof("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", dataChannel.Label(), dataChannel.ID())

	// 	for range time.NewTicker(5 * time.Second).C {
	// 		message := "offer ping"
	// 		yalog.Infof("Sending '%s'\n", message)

	// 		// Send the message as text
	// 		sendTextErr := dataChannel.SendText(message)
	// 		if sendTextErr != nil {
	// 			yalog.Error(sendTextErr)
	// 		}
	// 	}
	// })

	// Register text message handling
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		yalog.Infof("Message from DataChannel '%s': '%s'\n", dataChannel.Label(), string(msg.Data))
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		yalog.Infof("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			yalog.Info("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}

		if s == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			yalog.Info("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		yalog.Infof("Connection State has changed %s \n", connectionState.String())
		if connectionState.String() == "connected" {
			iceConnectedCtxCancel()
			yalog.Info("ICE 连接建立")
		}
	})

	file, err := os.Create("output.rtp")
	if err != nil {
		yalog.Error(err)
	}
	defer file.Close()

	// 接收Track
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		yalog.Infof("Track has started, of type %d: %s n", track.PayloadType(), track.Codec().MimeType)

		yalog.Info("等待连接建立")

		// Wait for connection established
		<-iceConnectedCtx.Done()

		yalog.Info("Track 开始接收")

		// 持续读取轨道的RTP包，并对其进行处理
		for {
			rtpPacket, _, readErr := track.ReadRTP()
			if readErr != nil {
				if readErr == io.EOF {
					break // RTP流结束
				}
				yalog.Error(readErr)
			}

			yalog.Info("接收到RTP包", yalog.Any("rtpPacket", rtpPacket))

			// 这里可以进行RTP包的处理，例如保存数据或解码成视频帧等
			// 将RTP包数据写入到文件中
			b, err := rtpPacket.Marshal()
			if err != nil {
				yalog.Error(err)
			}
			if _, writeErr := file.Write(b); writeErr != nil {
				yalog.Error(writeErr)
			}

		}
	})

	// Create an offer to send to the other process
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		yalog.Error(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	// Note: this will start the gathering of ICE candidates
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		yalog.Error(err)
	}

	// Send our offer to the HTTP server listening in the other process
	payload, err := json.Marshal(offer)
	if err != nil {
		yalog.Error(err)
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/sdp", *answerAddr), "application/json; charset=utf-8", bytes.NewReader(payload))
	if err != nil {
		yalog.Error(err)
	} else if err := resp.Body.Close(); err != nil {
		yalog.Error(err)
	}
	select {}
}
