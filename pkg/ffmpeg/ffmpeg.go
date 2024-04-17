package ffmpeg

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
)

//go:embed ffmpeg70.exe
var ffmpeg []byte

// TempFFmpeg creates a temporary file with the ffmpeg binary and returns the path to the file and a function to delete it.
func TempFFmpeg() (name string, delFFmpeg func(), err error) {
	file, err := os.CreateTemp("", "ffmpeg*.exe")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(ffmpeg); err != nil {
		return "", nil, fmt.Errorf("failed to write to temp file: %v", err)
	}

	delFFmpeg = func() {
		os.Remove(file.Name())
	}

	return file.Name(), delFFmpeg, nil
}

// TODO: 不完整的实现，需要完善
func RecordScreen(target string) (*io.ReadCloser, error) {
	ffmpegPath, delFFmpeg, err := TempFFmpeg()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp ffmpeg: %v", err)
	}
	defer delFFmpeg()

	// Create the command to record the screen
	cmd := exec.Command(ffmpegPath, "-f", "gdigrab", "-framerate", "30", "-i", target, "-f", "h264", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg command: %v", err)
	}

	// Wait for the command to finish
	// err = cmd.Wait()
	// if err != nil {
	// 	return nil, fmt.Errorf("ffmpeg command failed: %v", err)
	// }

	return &stdout, nil
}
