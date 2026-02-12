package ffmpeg

import "fmt"

// VideoCaptureParams holds parameters for building video capture FFmpeg arguments.
type VideoCaptureParams struct {
	DeviceID    string
	Width       int
	Height      int
	FrameRate   float64
	PixelFormat string // output pixel format, defaults to "yuv420p"
}

// AudioCaptureParams holds parameters for building audio capture FFmpeg arguments.
type AudioCaptureParams struct {
	DeviceID   string
	SampleRate int
	Channels   int
}

// videoOutputArgs returns the common output arguments for raw video capture.
func videoOutputArgs(p VideoCaptureParams) []string {
	pixFmt := p.PixelFormat
	if pixFmt == "" {
		pixFmt = "yuv420p"
	}
	args := []string{
		"-f", "rawvideo",
		"-pix_fmt", pixFmt,
	}
	if p.Width > 0 && p.Height > 0 {
		args = append(args, "-video_size", fmt.Sprintf("%dx%d", p.Width, p.Height))
	}
	args = append(args, "pipe:1")
	return args
}

// audioOutputArgs returns the common output arguments for raw audio capture.
func audioOutputArgs(p AudioCaptureParams) []string {
	args := []string{
		"-f", "s16le",
		"-acodec", "pcm_s16le",
	}
	if p.SampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", p.SampleRate))
	}
	if p.Channels > 0 {
		args = append(args, "-ac", fmt.Sprintf("%d", p.Channels))
	}
	args = append(args, "pipe:1")
	return args
}
