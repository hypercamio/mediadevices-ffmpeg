//go:build windows

package mediadevices

import (
	"os/exec"
	"regexp"
	"strings"
)

// dshowDeviceRe matches lines like: [dshow @ 0x...] "Device Name" (video)
var dshowDeviceRe = regexp.MustCompile(`\[dshow\s+@\s+\S+\]\s+"([^"]+)"\s+\((video|audio)\)`)

// dshowAltRe matches alternative format lines like: [dshow @ 0x...]  "Device Name"
// that appear after a section header indicating video or audio.
var dshowAltRe = regexp.MustCompile(`\[dshow\s+@\s+\S+\]\s+"([^"]+)"`)

// dshowSectionRe matches section headers like: [dshow @ 0x...] DirectShow video devices
var dshowSectionRe = regexp.MustCompile(`\[dshow\s+@\s+\S+\]\s+DirectShow\s+(video|audio)\s+devices`)

func discoverDevices(ffmpegPath string) ([]Device, error) {
	cmd := exec.Command(ffmpegPath, "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	// FFmpeg writes device list to stderr and exits with error code; that's expected.
	output, _ := cmd.CombinedOutput()
	return parseDshowOutput(string(output)), nil
}

func parseDshowOutput(output string) []Device {
	var devices []Device
	lines := strings.Split(output, "\n")

	// First try the explicit format: "Name" (video) / "Name" (audio)
	for _, line := range lines {
		m := dshowDeviceRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		kind := VideoDevice
		if m[2] == "audio" {
			kind = AudioDevice
		}
		devices = append(devices, Device{
			Name: name,
			ID:   name,
			Kind: kind,
		})
	}

	if len(devices) > 0 {
		return devices
	}

	// Fallback: parse section headers + quoted device names
	currentKind := VideoDevice
	for _, line := range lines {
		if sm := dshowSectionRe.FindStringSubmatch(line); sm != nil {
			if sm[1] == "audio" {
				currentKind = AudioDevice
			} else {
				currentKind = VideoDevice
			}
			continue
		}
		if am := dshowAltRe.FindStringSubmatch(line); am != nil {
			name := am[1]
			// Skip alternative name lines (they contain "Alternative name")
			if strings.Contains(line, "Alternative name") {
				continue
			}
			devices = append(devices, Device{
				Name: name,
				ID:   name,
				Kind: currentKind,
			})
		}
	}

	return devices
}
