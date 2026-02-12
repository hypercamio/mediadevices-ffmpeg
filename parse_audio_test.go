package mediadevices

import (
	"encoding/binary"
	"testing"
)

func TestParseS16LEChunk_Stereo(t *testing.T) {
	channels := 2
	sampleRate := 48000
	samplesPerChannel := 4

	// Build raw PCM data: 4 frames of stereo = 8 samples = 16 bytes.
	totalSamples := samplesPerChannel * channels
	data := make([]byte, totalSamples*2)

	expected := []int16{100, -100, 200, -200, 300, -300, 400, -400}
	for i, v := range expected {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(v))
	}

	chunk, err := parseS16LEChunk(data, channels, sampleRate)
	if err != nil {
		t.Fatalf("parseS16LEChunk: %v", err)
	}

	if chunk.Channels != channels {
		t.Errorf("channels = %d, want %d", chunk.Channels, channels)
	}
	if chunk.SampleRate != sampleRate {
		t.Errorf("sampleRate = %d, want %d", chunk.SampleRate, sampleRate)
	}
	if chunk.SamplesPerChannel != samplesPerChannel {
		t.Errorf("samplesPerChannel = %d, want %d", chunk.SamplesPerChannel, samplesPerChannel)
	}
	if len(chunk.Data) != totalSamples {
		t.Fatalf("len(Data) = %d, want %d", len(chunk.Data), totalSamples)
	}

	for i, v := range expected {
		if chunk.Data[i] != v {
			t.Errorf("Data[%d] = %d, want %d", i, chunk.Data[i], v)
		}
	}
}

func TestParseS16LEChunk_Mono(t *testing.T) {
	channels := 1
	sampleRate := 44100
	samples := []int16{0, 32767, -32768, 1000}

	data := make([]byte, len(samples)*2)
	for i, v := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(v))
	}

	chunk, err := parseS16LEChunk(data, channels, sampleRate)
	if err != nil {
		t.Fatalf("parseS16LEChunk: %v", err)
	}

	if chunk.Channels != 1 {
		t.Errorf("channels = %d, want 1", chunk.Channels)
	}
	if chunk.SamplesPerChannel != len(samples) {
		t.Errorf("samplesPerChannel = %d, want %d", chunk.SamplesPerChannel, len(samples))
	}
	for i, v := range samples {
		if chunk.Data[i] != v {
			t.Errorf("Data[%d] = %d, want %d", i, chunk.Data[i], v)
		}
	}
}

func TestParseS16LEChunk_BadAlignment(t *testing.T) {
	// 3 bytes is not aligned to stereo frame size (4 bytes).
	_, err := parseS16LEChunk([]byte{1, 2, 3}, 2, 48000)
	if err == nil {
		t.Fatal("expected error for misaligned data")
	}
}

func TestParseS16LEChunk_Empty(t *testing.T) {
	chunk, err := parseS16LEChunk([]byte{}, 2, 48000)
	if err != nil {
		t.Fatalf("parseS16LEChunk with empty data: %v", err)
	}
	if len(chunk.Data) != 0 {
		t.Errorf("expected empty data, got %d samples", len(chunk.Data))
	}
	if chunk.SamplesPerChannel != 0 {
		t.Errorf("expected 0 samplesPerChannel, got %d", chunk.SamplesPerChannel)
	}
}
