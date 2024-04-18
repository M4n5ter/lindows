package main

import (
	"flag"
	"time"

	peerConn "github.com/m4n5ter/lindows/pkg/webrtc"
	"github.com/m4n5ter/lindows/pkg/yalog"

	"github.com/pion/webrtc/v3"
)

func main() {
	offerAddr := flag.String("offer-address", "127.0.0.1:50000", "Address that the Offer HTTP server is hosted on.")
	answerAddr := flag.String("answer-address", ":60000", "Address that the Answer HTTP server is hosted on.")

	pc, _ := peerConn.NewPeerConnection(webrtc.Configuration{}, *answerAddr, *offerAddr)
	time.Sleep(5 * time.Second)
	yalog.Info(pc.Conn.ConnectionState().String())
	select {}
}


