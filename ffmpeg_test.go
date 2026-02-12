package mediadevices

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := GetConfig()
	if cfg.FFmpegPath != "ffmpeg" {
		t.Errorf("default FFmpegPath = %q, want %q", cfg.FFmpegPath, "ffmpeg")
	}
	if cfg.Verbose {
		t.Error("default Verbose should be false")
	}
}

func TestSetGetConfig(t *testing.T) {
	// Save original and restore after test.
	orig := GetConfig()
	defer SetConfig(orig)

	SetConfig(Config{
		FFmpegPath: "/usr/local/bin/ffmpeg",
		Verbose:    true,
	})

	cfg := GetConfig()
	if cfg.FFmpegPath != "/usr/local/bin/ffmpeg" {
		t.Errorf("FFmpegPath = %q, want %q", cfg.FFmpegPath, "/usr/local/bin/ffmpeg")
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
}

func TestSetConfig_EmptyPathDefaults(t *testing.T) {
	orig := GetConfig()
	defer SetConfig(orig)

	SetConfig(Config{FFmpegPath: ""})

	cfg := GetConfig()
	if cfg.FFmpegPath != "ffmpeg" {
		t.Errorf("empty FFmpegPath should default to %q, got %q", "ffmpeg", cfg.FFmpegPath)
	}
}

func TestDeviceKindString(t *testing.T) {
	tests := []struct {
		kind DeviceKind
		want string
	}{
		{VideoDevice, "video"},
		{AudioDevice, "audio"},
		{DeviceKind(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("DeviceKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}
