// Package ffmpeg provides media device capture (audio/video) using FFmpeg as the backend.
// It discovers available capture devices on the system and provides readers for
// raw video frames (image.YCbCr) and audio samples (PCM S16LE).
//
// Usage:
//
//	cfg := ffmpeg.GetConfig()
//	cfg.FFmpegPath = "/usr/local/bin/ffmpeg"
//	ffmpeg.SetConfig(cfg)
//
//	devices, err := ffmpeg.DiscoverDevices()
//	// pick a video device, then:
//	reader, err := ffmpeg.NewVideoReader(ffmpeg.VideoConfig{
//	    Device:    devices[0],
//	    Width:     1280,
//	    Height:    720,
//	    FrameRate: 30,
//	})
//	defer reader.Close()
//	img, err := reader.Read()
package mediadevices

import "sync"

// Config holds global configuration for FFmpeg operations.
type Config struct {
	// FFmpegPath is the path to the ffmpeg binary. Defaults to "ffmpeg" (resolved via PATH).
	FFmpegPath string

	// Verbose enables debug logging of FFmpeg stderr output.
	Verbose bool
}

var (
	globalConfig = Config{
		FFmpegPath: "ffmpeg",
	}
	configMu sync.RWMutex
)

// SetConfig updates the global FFmpeg configuration.
func SetConfig(cfg Config) {
	configMu.Lock()
	defer configMu.Unlock()
	if cfg.FFmpegPath == "" {
		cfg.FFmpegPath = "ffmpeg"
	}
	globalConfig = cfg
}

// GetConfig returns a copy of the current global FFmpeg configuration.
func GetConfig() Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}
