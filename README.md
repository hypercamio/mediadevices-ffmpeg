# mediadevices-ffmpeg-go

A self-contained Go library for media device capture (audio/video) using FFmpeg as the backend. Cross-platform support for Windows (DirectShow), Linux (V4L2/ALSA), and macOS (AVFoundation).

Compatible with the [pion/mediadevices](https://github.com/pion/mediadevices/) interface style.

## Requirements

- **Go** 1.25.0+
- **FFmpeg 8.x** installed and available in `PATH` (or configured via `SetConfig`)
- No external Go dependencies — pure stdlib

## Installation

```bash
go get github.com/hypercamio/mediadevices-ffmpeg-go
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	mediadevices "github.com/hypercamio/mediadevices-ffmpeg-go"
)

func main() {
	// Discover available devices
	devices, err := mediadevices.Devices()
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range devices {
		fmt.Printf("%s: %s (%s)\n", d.Kind, d.Name, d.ID)
	}

	// Capture video frames
	videoDevs, _ := mediadevices.VideoDevices()
	if len(videoDevs) == 0 {
		log.Fatal("no video device found")
	}

	reader, err := mediadevices.NewVideoReader(mediadevices.VideoConfig{
		Device:    videoDevs[0],
		Width:     640,
		Height:    480,
		FrameRate: 30,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	img, err := reader.Read() // returns *image.YCbCr
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("frame size:", img.Bounds().Dx(), "x", img.Bounds().Dy())
}
```

## API

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

### Device Discovery

```go
func DiscoverDevices() ([]Device, error) // Query system for all devices
func Initialize() ([]Device, error)      // Discover once, cache result
func Devices() ([]Device, error)          // Return cached devices
func VideoDevices() ([]Device, error)     // Filter to video devices only
func AudioDevices() ([]Device, error)     // Filter to audio devices only
```

`Device` struct:

```go
type Device struct {
    Name      string     // Human-readable name
    ID        string     // Platform-specific identifier
    Kind      DeviceKind // VideoDevice or AudioDevice
    IsDefault bool       // True if system default
}
```

### Video Capture

```go
reader, err := mediadevices.NewVideoReader(mediadevices.VideoConfig{
    Device:    device,
    Width:     1280,  // 0 = device default
    Height:    720,   // 0 = device default
    FrameRate: 30,    // 0 = device default
})
defer reader.Close()

img, err := reader.Read() // returns image.Image (*image.YCbCr, YUV420p)
```

### Audio Capture

```go
reader, err := mediadevices.NewAudioReader(mediadevices.AudioConfig{
    Device:     device,
    SampleRate: 48000,               // 0 = 48000 Hz
    Channels:   2,                   // 0 = 2 (stereo)
    Latency:    20 * time.Millisecond, // 0 = 20ms
})
defer reader.Close()

chunk, err := reader.Read() // returns *AudioChunk
```

`AudioChunk` struct:

```go
type AudioChunk struct {
    Data              []int16 // Interleaved PCM S16LE samples
    Channels          int
    SampleRate        int
    SamplesPerChannel int
}
```

## Data Formats

| Type | Format | Go Type |
|------|--------|---------|
| Video | YUV420p | `*image.YCbCr` |
| Audio | PCM S16LE | `*AudioChunk` (interleaved `[]int16`) |

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
