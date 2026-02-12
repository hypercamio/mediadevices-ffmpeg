//go:build windows

package ffmpeg

import (
	"strings"
	"testing"
)

func TestBuildVideoCaptureArgs_Windows(t *testing.T) {
	args := buildVideoCaptureArgs(VideoCaptureParams{
		DeviceID:  "Integrated Camera",
		Width:     1280,
		Height:    720,
		FrameRate: 30,
	})

	joined := strings.Join(args, " ")

	// Must use dshow format.
	if !contains(args, "-f", "dshow") {
		t.Errorf("missing -f dshow in args: %s", joined)
	}
	// Must reference the device.
	if !containsPrefix(args, "video=Integrated Camera") {
		t.Errorf("missing video=device in args: %s", joined)
	}
	// Must output rawvideo.
	if !contains(args, "-f", "rawvideo") {
		t.Errorf("missing -f rawvideo in args: %s", joined)
	}
	// Must output to pipe.
	if !containsValue(args, "pipe:1") {
		t.Errorf("missing pipe:1 in args: %s", joined)
	}
}

func TestBuildAudioCaptureArgs_Windows(t *testing.T) {
	args := buildAudioCaptureArgs(AudioCaptureParams{
		DeviceID:   "Microphone (Realtek Audio)",
		SampleRate: 48000,
		Channels:   2,
	})

	joined := strings.Join(args, " ")

	if !contains(args, "-f", "dshow") {
		t.Errorf("missing -f dshow in args: %s", joined)
	}
	if !containsPrefix(args, "audio=Microphone") {
		t.Errorf("missing audio=device in args: %s", joined)
	}
	if !contains(args, "-f", "s16le") {
		t.Errorf("missing -f s16le in args: %s", joined)
	}
	if !containsValue(args, "pipe:1") {
		t.Errorf("missing pipe:1 in args: %s", joined)
	}
}

// contains checks if args has a consecutive pair [flag, value].
func contains(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

// containsValue checks if any arg equals value.
func containsValue(args []string, value string) bool {
	for _, a := range args {
		if a == value {
			return true
		}
	}
	return false
}

// containsPrefix checks if any arg starts with prefix.
func containsPrefix(args []string, prefix string) bool {
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			return true
		}
	}
	return false
}
