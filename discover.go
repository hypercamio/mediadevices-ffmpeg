package mediadevices

// DiscoverDevices runs FFmpeg to enumerate available capture devices on the system.
// It uses the platform-appropriate method (dshow on Windows, v4l2/ALSA on Linux,
// avfoundation on macOS). The ffmpegPath from the global config is used.
//
// Returns an empty slice (not an error) if FFmpeg is not found or no devices are detected.
func DiscoverDevices() ([]Device, error) {
	cfg := GetConfig()
	return discoverDevices(cfg.FFmpegPath)
}
