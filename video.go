package ffmpeg

import (
	"fmt"
	"image"
	"io"
)

// VideoConfig configures video capture from a device.
type VideoConfig struct {
	// Device is the capture device to use.
	Device Device

	// Width is the desired frame width in pixels. 0 = device default.
	Width int

	// Height is the desired frame height in pixels. 0 = device default.
	Height int

	// FrameRate is the desired capture frame rate. 0 = device default.
	FrameRate float64
}

// VideoReader reads raw video frames from an FFmpeg subprocess.
// Each call to Read() returns one YUV420p frame as an *image.YCbCr.
type VideoReader struct {
	proc      *ffmpegProcess
	buf       []byte
	width     int
	height    int
	frameSize int
}

// NewVideoReader starts an FFmpeg subprocess to capture video from the given device.
// The caller must call Close() when done to stop the subprocess.
func NewVideoReader(cfg VideoConfig) (*VideoReader, error) {
	if cfg.Device.Kind != VideoDevice {
		return nil, fmt.Errorf("ffmpeg: device %q is not a video device", cfg.Device.Name)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, fmt.Errorf("ffmpeg: video width and height must be positive (got %dx%d)", cfg.Width, cfg.Height)
	}

	params := VideoCaptureParams{
		DeviceID:  cfg.Device.ID,
		Width:     cfg.Width,
		Height:    cfg.Height,
		FrameRate: cfg.FrameRate,
	}

	args := buildVideoCaptureArgs(params)
	gcfg := GetConfig()

	proc, err := startProcess(gcfg.FFmpegPath, args)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg: start video capture: %w", err)
	}

	frameSize := cfg.Width * cfg.Height * 3 / 2 // YUV420p

	return &VideoReader{
		proc:      proc,
		buf:       make([]byte, frameSize),
		width:     cfg.Width,
		height:    cfg.Height,
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
