package main

import (
	"io"
	"os/exec"
)

func RunCommand(name string, arg ...string) (io.ReadCloser, error) {
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

// ffmpeg -list_devices true -f dshow -i dummy
// go run . -rtbufsize 100M -f dshow -i video="PUT_DEVICE_NAME" -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264 -
