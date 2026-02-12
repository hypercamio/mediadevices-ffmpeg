package ffmpeg

// DeviceKind indicates whether a device captures video or audio.
type DeviceKind int

const (
	// VideoDevice represents a video capture device (camera, screen capture, etc.)
	VideoDevice DeviceKind = iota
	// AudioDevice represents an audio capture device (microphone, line-in, etc.)
	AudioDevice
)

// String returns a human-readable name for the device kind.
func (k DeviceKind) String() string {
	switch k {
	case VideoDevice:
		return "video"
	case AudioDevice:
		return "audio"
	default:
		return "unknown"
	}
}

// Device represents a discovered media capture device.
type Device struct {
	// Name is the human-readable device name.
	Name string

	// ID is the platform-specific device identifier used in FFmpeg commands.
	// On Windows (dshow): the device name string.
	// On Linux: device path (e.g., "/dev/video0") or ALSA ID (e.g., "hw:0,0").
	// On macOS (avfoundation): the device index string (e.g., "0", "1").
	ID string

	// Kind indicates whether this is a video or audio device.
	Kind DeviceKind

	// IsDefault indicates if this is the system default device.
	IsDefault bool
}
