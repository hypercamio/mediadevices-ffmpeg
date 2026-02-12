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

func discoverDevices(ffmpegPath string) ([]MediaDeviceInfo, error) {
	cmd := exec.Command(ffmpegPath, "-f", "avfoundation", "-list_devices", "true", "-i", "")
	// FFmpeg writes device list to stderr and exits with error code; that's expected.
	output, _ := cmd.CombinedOutput()
	return parseAVFoundationOutput(string(output)), nil
}

func parseAVFoundationOutput(output string) []MediaDeviceInfo {
	var devices []MediaDeviceInfo
	lines := strings.Split(output, "\n")
	currentKind := MediaDeviceKindVideoInput

	for _, line := range lines {
		if sm := avfSectionRe.FindStringSubmatch(line); sm != nil {
			if sm[1] == "audio" {
				currentKind = MediaDeviceKindAudioInput
			} else {
				currentKind = MediaDeviceKindVideoInput
			}
			continue
		}

		if dm := avfDeviceRe.FindStringSubmatch(line); dm != nil {
			idx := dm[1]
			name := strings.TrimSpace(dm[2])
			devices = append(devices, MediaDeviceInfo{
				DeviceID:  idx,
				GroupID:   idx, // avfoundation doesn't provide groupId, use deviceId
				Kind:      currentKind,
				Label:     name,
				IsDefault: idx == "0",
			})
		}
	}

	return devices
}
