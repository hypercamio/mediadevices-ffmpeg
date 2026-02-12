package mediadevices

import (
	"fmt"
	"image"
	"io"
)

// VideoReader reads raw video frames from an FFmpeg subprocess.
// Each call to Read() returns one YUV420p frame as an *image.YCbCr.
type VideoReader struct {
	proc      *ffmpegProcess
	buf       []byte
	width     int
	height    int
	frameSize int
}

// newVideoReaderInternal starts an FFmpeg subprocess to capture video from the given device.
// This is an internal function used by MediaStreamTrack.
func newVideoReaderInternal(deviceID string, width, height int, frameRate float64) (*VideoReader, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("ffmpeg: video width and height must be positive (got %dx%d)", width, height)
	}

	params := VideoCaptureParams{
		DeviceID:  deviceID,
		Width:     width,
		Height:    height,
		FrameRate: frameRate,
	}

	args := buildVideoCaptureArgs(params)
	gcfg := GetConfig()

	proc, err := startProcess(gcfg.FFmpegPath, args)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg: start video capture: %w", err)
	}

	frameSize := width * height * 3 / 2 // YUV420p

	return &VideoReader{
		proc:      proc,
		buf:       make([]byte, frameSize),
		width:     width,
		height:    height,
		frameSize: frameSize,
	}, nil
}

// Read reads one video frame from the capture.
// Returns an *image.YCbCr with YUV420p data.
// Returns io.EOF when the stream ends.
func (r *VideoReader) Read() (image.Image, error) {
	_, err := io.ReadFull(r.proc, r.buf)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("ffmpeg: read video frame: %w\nstderr: %s", err, r.proc.LastStderr())
	}

	img, err := parseYUV420pFrame(r.buf, r.width, r.height)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Close stops the FFmpeg subprocess and releases resources.
func (r *VideoReader) Close() error {
	if r.proc != nil {
		return r.proc.Stop()
	}
	return nil
}

// Width returns the video width in pixels.
func (r *VideoReader) Width() int {
	return r.width
}

// Height returns the video height in pixels.
func (r *VideoReader) Height() int {
	return r.height
}
