package mediadevices

import (
	"fmt"
	"io"
	"time"
)

// AudioReader reads raw audio chunks from an FFmpeg subprocess.
// Each call to Read() returns one chunk of interleaved PCM S16LE samples.
type AudioReader struct {
	proc              *ffmpegProcess
	buf               []byte
	channels          int
	sampleRate        int
	samplesPerChannel int
}

// newAudioReaderInternal starts an FFmpeg subprocess to capture audio from the given device.
// This is an internal function used by MediaStreamTrack.
func newAudioReaderInternal(deviceID string, sampleRate, channels int) (*AudioReader, error) {
	if sampleRate <= 0 {
		sampleRate = 48000
	}
	if channels <= 0 {
		channels = 2
	}
	latency := 20 * time.Millisecond

	params := AudioCaptureParams{
		DeviceID:   deviceID,
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

// SampleRate returns the audio sample rate in Hz.
func (r *AudioReader) SampleRate() int {
	return r.sampleRate
}

// Channels returns the number of audio channels.
func (r *AudioReader) Channels() int {
	return r.channels
}
