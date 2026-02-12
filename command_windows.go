//go:build windows

package ffmpeg

import "fmt"

// buildVideoCaptureArgs builds FFmpeg arguments for capturing video via DirectShow on Windows.
func buildVideoCaptureArgs(p VideoCaptureParams) []string {
	args := []string{"-y"}

	// Input format
	args = append(args, "-f", "dshow")

	// Input options
	if p.Width > 0 && p.Height > 0 {
		args = append(args, "-video_size", fmt.Sprintf("%dx%d", p.Width, p.Height))
	}
	if p.FrameRate > 0 {
		args = append(args, "-framerate", fmt.Sprintf("%g", p.FrameRate))
	}

	// Input device: video="Device Name"
	args = append(args, "-i", fmt.Sprintf("video=%s", p.DeviceID))

	// Output: raw YUV420p to stdout
	args = append(args, videoOutputArgs(p)...)

	return args
}

// buildAudioCaptureArgs builds FFmpeg arguments for capturing audio via DirectShow on Windows.
func buildAudioCaptureArgs(p AudioCaptureParams) []string {
	args := []string{"-y"}

	// Input format
	args = append(args, "-f", "dshow")

	// Input options
	if p.SampleRate > 0 {
		args = append(args, "-sample_rate", fmt.Sprintf("%d", p.SampleRate))
	}
	if p.Channels > 0 {
		args = append(args, "-channels", fmt.Sprintf("%d", p.Channels))
	}

	// Input device: audio="Device Name"
	args = append(args, "-i", fmt.Sprintf("audio=%s", p.DeviceID))

	// Output: raw PCM S16LE to stdout
	args = append(args, audioOutputArgs(p)...)

	return args
}
