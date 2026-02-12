//go:build linux

package mediadevices

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cardRe matches lines from /proc/asound/cards like: " 0 [PCH            ]: HDA-Intel - HDA Intel PCH"
var cardRe = regexp.MustCompile(`^\s*(\d+)\s+\[`)

func discoverDevices(ffmpegPath string) ([]MediaDeviceInfo, error) {
	var devices []MediaDeviceInfo

	videoDevs, err := discoverV4L2Devices()
	if err == nil {
		devices = append(devices, videoDevs...)
	}

	audioDevs, err := discoverALSADevices()
	if err == nil {
		devices = append(devices, audioDevs...)
	}

	return devices, nil
}

func discoverV4L2Devices() ([]MediaDeviceInfo, error) {
	matches, err := filepath.Glob("/dev/video*")
	if err != nil {
		return nil, err
	}

	var devices []MediaDeviceInfo
	for _, path := range matches {
		// Only include devices we can open.
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		f.Close()

		name := filepath.Base(path)
		devices = append(devices, MediaDeviceInfo{
			DeviceID:  path,
			GroupID:   path, // v4l2 doesn't provide groupId
			Kind:      MediaDeviceKindVideoInput,
			Label:     name,
			IsDefault: path == "/dev/video0",
		})
	}
	return devices, nil
}

func discoverALSADevices() ([]MediaDeviceInfo, error) {
	f, err := os.Open("/proc/asound/cards")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var devices []MediaDeviceInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		m := cardRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		cardNum := m[1]
		// Extract the descriptive name from the line (after the ": " part).
		name := strings.TrimSpace(line)
		if idx := strings.Index(name, " - "); idx >= 0 {
			name = strings.TrimSpace(name[idx+3:])
		}

		devices = append(devices, MediaDeviceInfo{
			DeviceID:  fmt.Sprintf("hw:%s", cardNum),
			GroupID:   fmt.Sprintf("hw:%s", cardNum), // ALSA doesn't provide separate groupId
			Kind:      MediaDeviceKindAudioInput,
			Label:     name,
			IsDefault: cardNum == "0",
		})
	}
	return devices, scanner.Err()
}
