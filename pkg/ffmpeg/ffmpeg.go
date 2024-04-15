package ffmpeg

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/m4n5ter/lindows/pkg/yalog"
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
func RecordScreen(duration, outputFilePath string) error {
	ffmpegPath, delFFmpeg, err := TempFFmpeg()
	if err != nil {
		return fmt.Errorf("failed to create temp ffmpeg: %v", err)
	}
	defer delFFmpeg()

	// Create the command to record the screen
	cmd := exec.Command(ffmpegPath, "-f", "gdigrab", "-framerate", "30", "-t", duration, "-i", "desktop", "-f", "webm", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Set up the output file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Create a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := io.Copy(outputFile, stdout)
		if err != nil {
			yalog.Error("failed to copy ffmpeg stdout to file", "error", err)
		}
	}()

	// Start the command
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg command: %v", err)
	}

	// Wait for the goroutine to finish
	wg.Wait()

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}

	return nil
}
