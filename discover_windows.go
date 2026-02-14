//go:build windows

package mediadevices

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
)

// dshowDeviceRe matches lines like: [dshow @ 0x...] "Device Name" (video)
var dshowDeviceRe = regexp.MustCompile(`\[dshow\s+@\s+\S+\]\s+"([^"]+)"\s+\((video|audio)\)`)

// dshowAltRe matches alternative format lines like: [dshow @ 0x...]  "Device Name"
// that appear after a section header indicating video or audio.
var dshowAltRe = regexp.MustCompile(`\[dshow\s+@\s+\S+\]\s+"([^"]+)"`)

// dshowSectionRe matches section headers like: [dshow @ 0x...] DirectShow video devices
var dshowSectionRe = regexp.MustCompile(`\[dshow\s+@\s+\S+\]\s+DirectShow\s+(video|audio)\s+devices`)

func discoverDevices(ffmpegPath string) ([]MediaDeviceInfo, error) {
	cmd := exec.Command(ffmpegPath, "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	// FFmpeg writes device list to stderr and exits with error code; that's expected.
	output, _ := cmd.CombinedOutput()
	return parseDshowOutput(string(output)), nil
}

// getMachineID returns the unique machine ID for this device.
func getMachineID() string {
	id, err := machineid.ID()
	if err != nil {
		// Fallback to a constant if machine ID cannot be obtained
		return "unknown"
	}
	return id
}

// machineID is cached at package init
var cachedMachineID = getMachineID()

// generateDeviceUUID generates a deterministic UUID from machine ID, device name and kind.
// This ensures the same device on the same machine always gets the same UUID,
// while devices on different machines get different UUIDs even with identical names.
func generateDeviceUUID(name string, kind MediaDeviceKind) uuid.UUID {
	// Include machine ID, device name, and kind in the hash
	input := fmt.Sprintf("%s:%s:%s", cachedMachineID, name, kind)
	hash := sha256.Sum256([]byte(input))
	// Use first 16 bytes of SHA256 hash to create UUID v5 style
	return uuid.UUID{
		hash[0], hash[1], hash[2], hash[3],
		hash[4], hash[5], hash[6], hash[7],
		hash[8], hash[9], hash[10], hash[11],
		hash[12], hash[13], hash[14], hash[15],
	}
}

func parseDshowOutput(output string) []MediaDeviceInfo {
	var devices []MediaDeviceInfo
	lines := strings.Split(output, "\n")

	// Track seen name+kind combinations to handle potential duplicates
	seenDeviceKeys := make(map[string]int)

	// First try the explicit format: "Name" (video) / "Name" (audio)
	for _, line := range lines {
		m := dshowDeviceRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		kind := MediaDeviceKindVideoInput
		if m[2] == "audio" {
			kind = MediaDeviceKindAudioInput
		}
		// Generate unique key for this name+kind combination
		deviceKey := fmt.Sprintf("%s:%s", name, kind)
		seenDeviceKeys[deviceKey]++
		// If duplicate, append index to ensure unique UUID
		uniqueKey := deviceKey
		if seenDeviceKeys[deviceKey] > 1 {
			uniqueKey = fmt.Sprintf("%s:%d", deviceKey, seenDeviceKeys[deviceKey])
		}
		deviceID := generateDeviceUUID(uniqueKey, kind).String()
		devices = append(devices, MediaDeviceInfo{
			DeviceID:   deviceID,
			DeviceName: name, // Original device name for FFmpeg
			GroupID:    name, // dshow doesn't provide groupId, use name for grouping
			Kind:       kind,
			Label:      name,
			IsDefault:  false, // dshow doesn't indicate default
		})
	}

	if len(devices) > 0 {
		return devices
	}

	// Fallback: parse section headers + quoted device names
	currentKind := MediaDeviceKindVideoInput
	for _, line := range lines {
		if sm := dshowSectionRe.FindStringSubmatch(line); sm != nil {
			if sm[1] == "audio" {
				currentKind = MediaDeviceKindAudioInput
			} else {
				currentKind = MediaDeviceKindVideoInput
			}
			continue
		}
		if am := dshowAltRe.FindStringSubmatch(line); am != nil {
			name := am[1]
			// Skip alternative name lines (they contain "Alternative name")
			if strings.Contains(line, "Alternative name") {
				continue
			}
			// Generate unique key with kind and seen count
			deviceKey := fmt.Sprintf("%s:%s", name, currentKind)
			seenDeviceKeys[deviceKey]++
			uniqueKey := deviceKey
			if seenDeviceKeys[deviceKey] > 1 {
				uniqueKey = fmt.Sprintf("%s:%d", deviceKey, seenDeviceKeys[deviceKey])
			}
			deviceID := generateDeviceUUID(uniqueKey, currentKind).String()
			devices = append(devices, MediaDeviceInfo{
				DeviceID:   deviceID,
				DeviceName: name, // Original device name for FFmpeg
				GroupID:    name,
				Kind:       currentKind,
				Label:      name,
				IsDefault:  false,
			})
		}
	}

	return devices
}
