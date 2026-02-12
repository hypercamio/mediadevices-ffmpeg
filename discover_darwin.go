//go:build darwin

package mediadevices

import (
	"os/exec"
	"regexp"
	"strings"
)

// avfDeviceRe matches lines like: [AVFoundation ...] [0] FaceTime HD Camera
var avfDeviceRe = regexp.MustCompile(`\[AVFoundation[^\]]*\]\s+\[(\d+)\]\s+(.+)`)

// avfSectionRe matches section headers like: [AVFoundation ...] AVFoundation video devices:
var avfSectionRe = regexp.MustCompile(`\[AVFoundation[^\]]*\]\s+AVFoundation\s+(video|audio)\s+devices:`)

func discoverDevices(ffmpegPath string) ([]Device, error) {
	cmd := exec.Command(ffmpegPath, "-f", "avfoundation", "-list_devices", "true", "-i", "")
	// FFmpeg writes device list to stderr and exits with error code; that's expected.
	output, _ := cmd.CombinedOutput()
	return parseAVFoundationOutput(string(output)), nil
}

func parseAVFoundationOutput(output string) []Device {
	var devices []Device
	lines := strings.Split(output, "\n")
	currentKind := VideoDevice

	for _, line := range lines {
		if sm := avfSectionRe.FindStringSubmatch(line); sm != nil {
			if sm[1] == "audio" {
				currentKind = AudioDevice
			} else {
				currentKind = VideoDevice
			}
			continue
		}

		if dm := avfDeviceRe.FindStringSubmatch(line); dm != nil {
			idx := dm[1]
			name := strings.TrimSpace(dm[2])
			devices = append(devices, Device{
				Name:      name,
				ID:        idx,
				Kind:      currentKind,
				IsDefault: idx == "0",
			})
		}
	}

	return devices
}
