# mediadevices-ffmpeg

A self-contained Go library for media device capture (audio/video) using FFmpeg as the backend. Cross-platform support for Windows (DirectShow), Linux (V4L2/ALSA), and macOS (AVFoundation).

Designed to follow the [MDN MediaDevices Web API](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices) standard.

## Requirements

- **Go** 1.25.0+
- **FFmpeg 8.x** installed and available in `PATH` (or configured via `SetConfig`)
- No external Go dependencies — pure stdlib

## Installation

```bash
go get github.com/hypercamio/mediadevices-ffmpeg
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	mediadevices "github.com/hypercamio/mediadevices-ffmpeg"
)

func main() {
	// Discover available devices
	devices, err := mediadevices.EnumerateDevices()
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range devices {
		fmt.Printf("%s: %s (id=%s)\n", d.Kind, d.Label, d.DeviceID)
	}

	// Request camera and microphone access
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaTrackConstraints{
		Video: &mediadevices.VideoTrackConstraints{
			Width:    mediadevices.IntPtr(1280),
			Height:   mediadevices.IntPtr(720),
			FrameRate: mediadevices.Float64Ptr(30.0),
		},
		Audio: &mediadevices.AudioTrackConstraints{
			SampleRate: mediadevices.IntPtr(48000),
			Channels:   mediadevices.IntPtr(2),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	fmt.Println("Stream ID:", stream.ID())
	fmt.Println("Active:", stream.Active())

	// Get video tracks
	videoTracks := stream.GetVideoTracks()
	if len(videoTracks) > 0 {
		track := videoTracks[0]
		fmt.Println("Track ID:", track.ID())
		fmt.Println("Track Kind:", track.Kind())

		// Read frames
		for {
			img, err := track.Read()
			if err != nil {
				break
			}
			fmt.Println("frame:", img.Bounds())
		}
	}
}
```

## API Reference

### Device Discovery

```go
// Enumerate all available media devices
devices, err := mediadevices.EnumerateDevices() ([]MediaDeviceInfo, error)

// Get only video input devices
videoDevs, err := mediadevices.VideoInputDevices() ([]MediaDeviceInfo, error)

// Get only audio input devices
audioDevs, err := mediadevices.AudioInputDevices() ([]MediaDeviceInfo, error)

// Get supported constraints
constraints := mediadevices.GetSupportedConstraints()
```

`MediaDeviceInfo` struct:

```go
type MediaDeviceInfo struct {
	DeviceID  string           // Unique device identifier
	GroupID   string           // Group ID for related devices
	Kind      MediaDeviceKind  // "videoinput", "audioinput", "audiooutput"
	Label     string           // Human-readable name (may be empty due to privacy)
	IsDefault bool             // True if system default
}
```

`MediaDeviceKind` constants:

```go
mediadevices.MediaDeviceKindVideoInput   // "videoinput"
mediadevices.MediaDeviceKindAudioInput   // "audioinput"
mediadevices.MediaDeviceKindAudioOutput  // "audiooutput"
```

### Media Capture

```go
// Request access to camera and/or microphone
stream, err := mediadevices.GetUserMedia(MediaTrackConstraints) (*MediaStream, error)
```

`MediaTrackConstraints`:

```go
type MediaTrackConstraints struct {
	Video *VideoTrackConstraints
	Audio *AudioTrackConstraints
}

type VideoTrackConstraints struct {
	Width       *int
	Height      *int
	FrameRate   *float64
	AspectRatio *float64
	DeviceID    *string
}

type AudioTrackConstraints struct {
	SampleRate       *int
	Channels         *int
	EchoCancellation *bool
	AutoGainControl  *bool
	NoiseSuppression *bool
	DeviceID         *string
}
```

### MediaStream

```go
stream.ID()                         // Get stream ID
stream.Active()                     // Check if stream is active
stream.GetTracks()                  // Get all tracks
stream.GetVideoTracks()             // Get video tracks only
stream.GetAudioTracks()            // Get audio tracks only
stream.GetTrackByID(id)            // Get track by ID
stream.AddTrack(track)             // Add a track
stream.RemoveTrack(track)          // Remove a track
stream.Clone()                     // Clone the stream
stream.Close()                     // Close and release resources
```

### MediaStreamTrack

```go
track.ID()                          // Get track ID
track.Kind()                        // Get track kind
track.Label()                       // Get track label
track.Enabled()                     // Check if enabled
track.SetEnabled(bool)             // Enable/disable track
track.ReadyState()                 // Get "live" or "ended"
track.Stop()                        // Stop the track
track.GetSettings()                // Get current settings
track.Close()                      // Stop the track (io.Closer)
```

Reading data:

```go
// Video track
img, err := track.Read()   // returns image.Image (*image.YCbCr, YUV420p)

// Audio track
chunk, err := track.ReadAudio() // returns *AudioChunk
```

### MediaTrackSettings

```go
type MediaTrackSettings struct {
	Width            int
	Height           int
	FrameRate        float64
	AspectRatio      float64
	SampleRate       int
	SampleSize       int
	EchoCancellation bool
	AutoGainControl  bool
	NoiseSuppression bool
}
```

### Helper Functions

```go
// Create pointer values for constraints
width := mediadevices.IntPtr(1280)
rate := mediadevices.Float64Ptr(30.0)
enabled := mediadevices.BoolPtr(true)
```

### Configuration

```go
// Set custom FFmpeg path or enable verbose logging
mediadevices.SetConfig(mediadevices.Config{
    FFmpegPath: "/usr/local/bin/ffmpeg",
    Verbose:    true,
})

cfg := mediadevices.GetConfig()
```

| Field | Default | Description |
|-------|---------|-------------|
| `FFmpegPath` | `"ffmpeg"` | Path to FFmpeg binary |
| `Verbose` | `false` | Enable debug logging to stderr |

## Data Formats

| Type | Format | Go Type |
|------|--------|---------|
| Video | YUV420p | `*image.YCbCr` |
| Audio | PCM S16LE | `*AudioChunk` (interleaved `[]int16`) |

`AudioChunk` struct:

```go
type AudioChunk struct {
	Data              []int16 // Interleaved PCM S16LE samples
	Channels          int
	SampleRate        int
	SamplesPerChannel int
}
```

## Platform Support

| Platform | Video | Audio | Backend |
|----------|-------|-------|---------|
| Windows | DirectShow (dshow) | DirectShow (dshow) | `ffmpeg -f dshow` |
| Linux | V4L2 (`/dev/video*`) | ALSA (`hw:X`) | `ffmpeg -f v4l2` / `ffmpeg -f alsa` |
| macOS | AVFoundation | AVFoundation | `ffmpeg -f avfoundation` |

## Examples

See the [`examples/`](examples/) directory:

- **[facedetection](examples/facedetection/)** — Real-time face detection using [pigo](https://github.com/esimov/pigo), demonstrating video capture and YUV frame processing.

## License

See [LICENSE](LICENSE) for details.
