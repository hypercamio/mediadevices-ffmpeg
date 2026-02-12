//go:build linux

package mediadevices

import "fmt"

// buildVideoCaptureArgs builds FFmpeg arguments for capturing video via V4L2 on Linux.
func buildVideoCaptureArgs(p VideoCaptureParams) []string {
	args := []string{"-y"}

	// Input format
	args = append(args, "-f", "v4l2")

	// Input options
	if p.Width > 0 && p.Height > 0 {
		args = append(args, "-video_size", fmt.Sprintf("%dx%d", p.Width, p.Height))
	}
	if p.FrameRate > 0 {
		args = append(args, "-framerate", fmt.Sprintf("%g", p.FrameRate))
	}

	// Input device: /dev/video0
	args = append(args, "-i", p.DeviceID)

	// Output: raw YUV420p to stdout
	args = append(args, videoOutputArgs(p)...)

	return args
}

// buildAudioCaptureArgs builds FFmpeg arguments for capturing audio via ALSA on Linux.
func buildAudioCaptureArgs(p AudioCaptureParams) []string {
	args := []string{"-y"}

	// Input format
	args = append(args, "-f", "alsa")

	// Input options
	if p.SampleRate > 0 {
		args = append(args, "-sample_rate", fmt.Sprintf("%d", p.SampleRate))
	}
	if p.Channels > 0 {
		args = append(args, "-channels", fmt.Sprintf("%d", p.Channels))
	}

	// Input device: hw:0,0
	args = append(args, "-i", p.DeviceID)

	// Output: raw PCM S16LE to stdout
	args = append(args, audioOutputArgs(p)...)

	return args
}
