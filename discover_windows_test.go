//go:build windows

package mediadevices

import "testing"

func TestParseDshowOutput_ExplicitFormat(t *testing.T) {
	// Simulates ffmpeg -list_devices true -f dshow -i dummy stderr output.
	output := `ffmpeg version 6.0 Copyright (c) 2000-2023 the FFmpeg developers
[dshow @ 000001] DirectShow video devices (some dshow devices may be not listed due to permissions issues.)
[dshow @ 000001] "Integrated Camera" (video)
[dshow @ 000001] "OBS Virtual Camera" (video)
[dshow @ 000001] DirectShow audio devices
[dshow @ 000001] "Microphone (Realtek Audio)" (audio)
[dshow @ 000001] "Stereo Mix (Realtek Audio)" (audio)
dummy: Immediate exit requested
`
	devices := parseDshowOutput(output)

	if len(devices) != 4 {
		t.Fatalf("got %d devices, want 4", len(devices))
	}

	// Video devices.
	if devices[0].Label != "Integrated Camera" || devices[0].Kind != MediaDeviceKindVideoInput {
		t.Errorf("devices[0] = %+v, want Integrated Camera video", devices[0])
	}
	if devices[1].Label != "OBS Virtual Camera" || devices[1].Kind != MediaDeviceKindVideoInput {
		t.Errorf("devices[1] = %+v, want OBS Virtual Camera video", devices[1])
	}

	// Audio devices.
	if devices[2].Label != "Microphone (Realtek Audio)" || devices[2].Kind != MediaDeviceKindAudioInput {
		t.Errorf("devices[2] = %+v, want Microphone audio", devices[2])
	}
	if devices[3].Label != "Stereo Mix (Realtek Audio)" || devices[3].Kind != MediaDeviceKindAudioInput {
		t.Errorf("devices[3] = %+v, want Stereo Mix audio", devices[3])
	}

	// DeviceIDs should equal Labels for dshow.
	for _, d := range devices {
		if d.DeviceID != d.Label {
			t.Errorf("device %q: DeviceID = %q, want same as Label", d.Label, d.DeviceID)
		}
	}
}

func TestParseDshowOutput_AltFormat(t *testing.T) {
	// Some ffmpeg versions use a different format without (video)/(audio) suffix.
	output := `[dshow @ 0x1234] DirectShow video devices (some dshow devices may be not listed due to permissions issues.)
[dshow @ 0x1234]  "USB Camera"
[dshow @ 0x1234]     Alternative name "@device_pnp_..."
[dshow @ 0x1234] DirectShow audio devices (some dshow devices may be not listed due to permissions issues.)
[dshow @ 0x1234]  "Built-in Mic"
[dshow @ 0x1234]     Alternative name "@device_cm_..."
`
	devices := parseDshowOutput(output)

	if len(devices) != 2 {
		t.Fatalf("got %d devices, want 2", len(devices))
	}
	if devices[0].Label != "USB Camera" || devices[0].Kind != MediaDeviceKindVideoInput {
		t.Errorf("devices[0] = %+v", devices[0])
	}
	if devices[1].Label != "Built-in Mic" || devices[1].Kind != MediaDeviceKindAudioInput {
		t.Errorf("devices[1] = %+v", devices[1])
	}
}

func TestParseDshowOutput_Empty(t *testing.T) {
	devices := parseDshowOutput("")
	if len(devices) != 0 {
		t.Errorf("got %d devices from empty output, want 0", len(devices))
	}
}
