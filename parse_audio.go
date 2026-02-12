package ffmpeg

import (
	"encoding/binary"
	"fmt"
)

// AudioChunk holds a chunk of interleaved PCM audio samples.
type AudioChunk struct {
	// Data contains interleaved int16 samples: [L0, R0, L1, R1, ...] for stereo.
	Data []int16

	// Channels is the number of audio channels (1 = mono, 2 = stereo).
	Channels int

	// SampleRate is the sampling rate in Hz (e.g., 48000, 44100).
	SampleRate int

	// SamplesPerChannel is the number of samples per channel in this chunk.
	SamplesPerChannel int
}

// parseS16LEChunk converts raw PCM S16LE interleaved bytes into an *AudioChunk.
// The input length must be a multiple of (channels * 2) bytes.
func parseS16LEChunk(data []byte, channels, sampleRate int) (*AudioChunk, error) {
	bytesPerSample := 2 // S16LE = 2 bytes per sample
	frameSize := channels * bytesPerSample

	if len(data)%frameSize != 0 {
		return nil, fmt.Errorf("S16LE chunk: %d bytes not aligned to frame size %d (channels=%d)", len(data), frameSize, channels)
	}

	totalSamples := len(data) / bytesPerSample
	samplesPerChannel := totalSamples / channels

	samples := make([]int16, totalSamples)
	for i := 0; i < totalSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}

	return &AudioChunk{
		Data:              samples,
		Channels:          channels,
		SampleRate:        sampleRate,
		SamplesPerChannel: samplesPerChannel,
	}, nil
}
