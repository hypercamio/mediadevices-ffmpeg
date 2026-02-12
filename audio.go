package mediadevices

import (
	"fmt"
	"io"
	"time"
)

// AudioConfig configures audio capture from a device.
type AudioConfig struct {
	// Device is the capture device to use.
	Device Device

	// SampleRate is the desired sample rate in Hz (e.g., 48000, 44100). 0 = device default.
	SampleRate int

	// Channels is the desired number of channels (1 = mono, 2 = stereo). 0 = device default.
	Channels int

	// Latency is the desired chunk duration. Smaller values mean lower latency
	// but more overhead per chunk. Defaults to 20ms if zero.
	Latency time.Duration
}

// AudioReader reads raw audio chunks from an FFmpeg subprocess.
// Each call to Read() returns one chunk of interleaved PCM S16LE samples.
type AudioReader struct {
	proc              *ffmpegProcess
	buf               []byte
	channels          int
	sampleRate        int
	samplesPerChannel int
}

// NewAudioReader starts an FFmpeg subprocess to capture audio from the given device.
// The caller must call Close() when done to stop the subprocess.
func NewAudioReader(cfg AudioConfig) (*AudioReader, error) {
	if cfg.Device.Kind != AudioDevice {
		return nil, fmt.Errorf("ffmpeg: device %q is not an audio device", cfg.Device.Name)
	}

	sampleRate := cfg.SampleRate
	if sampleRate <= 0 {
		sampleRate = 48000
	}
	channels := cfg.Channels
	if channels <= 0 {
		channels = 2
	}
	latency := cfg.Latency
	if latency <= 0 {
		latency = 20 * time.Millisecond
	}

	params := AudioCaptureParams{
		DeviceID:   cfg.Device.ID,
		SampleRate: sampleRate,
		Channels:   channels,
	}

	args := buildAudioCaptureArgs(params)
	gcfg := GetConfig()

	proc, err := startProcess(gcfg.FFmpegPath, args)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg: start audio capture: %w", err)
	}

	// Calculate chunk size based on latency.
	// samplesPerChannel = sampleRate * latencySeconds
	samplesPerChannel := int(float64(sampleRate) * latency.Seconds())
	chunkBytes := samplesPerChannel * channels * 2 // 2 bytes per S16LE sample

	return &AudioReader{
		proc:              proc,
		buf:               make([]byte, chunkBytes),
		channels:          channels,
		sampleRate:        sampleRate,
		samplesPerChannel: samplesPerChannel,
	}, nil
}

// Read reads one audio chunk from the capture.
// Returns an *AudioChunk with interleaved S16LE samples.
// Returns io.EOF when the stream ends.
func (r *AudioReader) Read() (*AudioChunk, error) {
	_, err := io.ReadFull(r.proc, r.buf)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("ffmpeg: read audio chunk: %w\nstderr: %s", err, r.proc.LastStderr())
	}

	chunk, err := parseS16LEChunk(r.buf, r.channels, r.sampleRate)
	if err != nil {
		return nil, err
	}
	return chunk, nil
}

// Close stops the FFmpeg subprocess and releases resources.
func (r *AudioReader) Close() error {
	if r.proc != nil {
		return r.proc.Stop()
	}
	return nil
}
