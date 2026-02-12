//go:build windows

package ffmpeg

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
	if devices[0].Name != "Integrated Camera" || devices[0].Kind != VideoDevice {
		t.Errorf("devices[0] = %+v, want Integrated Camera video", devices[0])
	}
	if devices[1].Name != "OBS Virtual Camera" || devices[1].Kind != VideoDevice {
		t.Errorf("devices[1] = %+v, want OBS Virtual Camera video", devices[1])
	}

	// Audio devices.
	if devices[2].Name != "Microphone (Realtek Audio)" || devices[2].Kind != AudioDevice {
		t.Errorf("devices[2] = %+v, want Microphone audio", devices[2])
	}
	if devices[3].Name != "Stereo Mix (Realtek Audio)" || devices[3].Kind != AudioDevice {
		t.Errorf("devices[3] = %+v, want Stereo Mix audio", devices[3])
	}

	// IDs should equal names for dshow.
	for _, d := range devices {
		if d.ID != d.Name {
			t.Errorf("device %q: ID = %q, want same as Name", d.Name, d.ID)
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
	if devices[0].Name != "USB Camera" || devices[0].Kind != VideoDevice {
		t.Errorf("devices[0] = %+v", devices[0])
	}
	if devices[1].Name != "Built-in Mic" || devices[1].Kind != AudioDevice {
		t.Errorf("devices[1] = %+v", devices[1])
	}
}

func TestParseDshowOutput_Empty(t *testing.T) {
	devices := parseDshowOutput("")
	if len(devices) != 0 {
		t.Errorf("got %d devices from empty output, want 0", len(devices))
	}
}
