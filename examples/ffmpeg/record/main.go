package main

import "github.com/m4n5ter/lindows/pkg/ffmpeg"

func main() {
	ffmpeg.RecordScreen("5", "output.webm")
}
