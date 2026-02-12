//go:build linux

package ffmpeg

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

func discoverDevices(ffmpegPath string) ([]Device, error) {
	var devices []Device

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

func discoverV4L2Devices() ([]Device, error) {
	matches, err := filepath.Glob("/dev/video*")
	if err != nil {
		return nil, err
	}

	var devices []Device
	for _, path := range matches {
		// Only include devices we can open.
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		f.Close()

		name := filepath.Base(path)
		devices = append(devices, Device{
			Name:      name,
			ID:        path,
			Kind:      VideoDevice,
			IsDefault: path == "/dev/video0",
		})
	}
	return devices, nil
}

func discoverALSADevices() ([]Device, error) {
	f, err := os.Open("/proc/asound/cards")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var devices []Device
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

		devices = append(devices, Device{
			Name:      name,
			ID:        fmt.Sprintf("hw:%s", cardNum),
			Kind:      AudioDevice,
			IsDefault: cardNum == "0",
		})
	}
	return devices, scanner.Err()
}
