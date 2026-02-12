//go:build darwin

package mediadevices

import "fmt"

// buildVideoCaptureArgs builds FFmpeg arguments for capturing video via AVFoundation on macOS.
func buildVideoCaptureArgs(p VideoCaptureParams) []string {
	args := []string{"-y"}

	// Input format
	args = append(args, "-f", "avfoundation")

	// Input options
	if p.Width > 0 && p.Height > 0 {
		args = append(args, "-video_size", fmt.Sprintf("%dx%d", p.Width, p.Height))
	}
	if p.FrameRate > 0 {
		args = append(args, "-framerate", fmt.Sprintf("%g", p.FrameRate))
	}

	// Input device: "INDEX:none" (video only, no audio)
	args = append(args, "-i", fmt.Sprintf("%s:none", p.DeviceID))

	// Output: raw YUV420p to stdout
	args = append(args, videoOutputArgs(p)...)

	return args
}

// buildAudioCaptureArgs builds FFmpeg arguments for capturing audio via AVFoundation on macOS.
func buildAudioCaptureArgs(p AudioCaptureParams) []string {
	args := []string{"-y"}

	// Input format
	args = append(args, "-f", "avfoundation")

	// Input options
	if p.SampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", p.SampleRate))
	}
	if p.Channels > 0 {
		args = append(args, "-ac", fmt.Sprintf("%d", p.Channels))
	}

	// Input device: "none:INDEX" (no video, audio only)
	args = append(args, "-i", fmt.Sprintf("none:%s", p.DeviceID))

	// Output: raw PCM S16LE to stdout
	args = append(args, audioOutputArgs(p)...)

	return args
}
