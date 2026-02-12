package ffmpeg

import (
	"log"
	"sync"
)

var (
	initOnce       sync.Once
	cachedDevices  []Device
	cachedDevErr   error
)

// Initialize discovers available media devices using FFmpeg. It is safe to call
// multiple times; discovery only runs once. Subsequent calls return the cached result.
//
// This is called automatically on the first call to Devices(). You may call it
// explicitly to control when discovery happens (e.g., at startup).
func Initialize() ([]Device, error) {
	initOnce.Do(func() {
		cfg := GetConfig()
		cachedDevices, cachedDevErr = discoverDevices(cfg.FFmpegPath)
		if cachedDevErr != nil && cfg.Verbose {
			log.Printf("ffmpeg: device discovery failed: %v", cachedDevErr)
		}
		if cfg.Verbose && cachedDevErr == nil {
			log.Printf("ffmpeg: discovered %d devices", len(cachedDevices))
			for _, d := range cachedDevices {
				log.Printf("ffmpeg:   [%s] %s (id=%s, default=%v)", d.Kind, d.Name, d.ID, d.IsDefault)
			}
		}
	})
	return cachedDevices, cachedDevErr
}

// Devices returns the list of discovered media devices.
// Calls Initialize() on first use.
func Devices() ([]Device, error) {
	return Initialize()
}

// VideoDevices returns only the discovered video capture devices.
func VideoDevices() ([]Device, error) {
	all, err := Devices()
	if err != nil {
		return nil, err
	}
	var result []Device
	for _, d := range all {
		if d.Kind == VideoDevice {
			result = append(result, d)
		}
	}
	return result, nil
}

// AudioDevices returns only the discovered audio capture devices.
func AudioDevices() ([]Device, error) {
	all, err := Devices()
	if err != nil {
		return nil, err
	}
	var result []Device
	for _, d := range all {
		if d.Kind == AudioDevice {
			result = append(result, d)
		}
	}
	return result, nil
}
