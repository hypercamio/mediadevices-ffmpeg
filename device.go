package mediadevices

// DeviceKind indicates whether a device captures video or audio.
type DeviceKind int

const (
	// VideoDevice represents a video capture device (camera, screen capture, etc.)
	VideoDevice DeviceKind = iota
	// AudioDevice represents an audio capture device (microphone, line-in, etc.)
	AudioDevice
)
