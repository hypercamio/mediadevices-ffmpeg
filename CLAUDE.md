# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go library for media device capture (audio/video) using FFmpeg 8.x as the backend. Cross-platform support for Windows (dshow), Linux (v4l2/ALSA), and macOS (avfoundation).

- **Module**: `github.com/hypercamio/mediadevices-ffmpeg`
- **Package**: `mediadevices`
- **Go version**: 1.25.0

## Build Commands

```bash
go build ./...              # Build all packages
go test ./...               # Run all tests
go test -run TestName .     # Run a single test
go vet ./...                # Static analysis
```

## Architecture

### API Design

The public API follows the [MDN MediaDevices Web API](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices) standard:

```
EnumerateDevices() → []MediaDeviceInfo
GetUserMedia(constraints) → *MediaStream
  ├── MediaStream.GetVideoTracks() → []*MediaStreamTrack
  ├── MediaStream.GetAudioTracks() → []*MediaStreamTrack
  └── each track: Read() / ReadAudio()
```

### Core Types

| Type | File | Purpose |
|------|------|---------|
| `MediaStream` | `stream_track.go` | Container for media tracks (MDN MediaStream) |
| `MediaStreamTrack` | `stream_track.go` | Single video/audio track (MDN MediaStreamTrack) |
| `VideoReader` | `video.go` | Reads raw YUV420p frames from FFmpeg |
| `AudioReader` | `audio.go` | Reads PCM S16LE audio chunks from FFmpeg |
| `ffmpegProcess` | `process.go` | Manages FFmpeg subprocess lifecycle |

### Platform-Specific Code

Device discovery and FFmpeg argument building use build tags:

| File | Platform | Backend |
|------|----------|---------|
| `discover_windows.go` | windows | DirectShow (`-f dshow`) |
| `discover_linux.go` | linux | V4L2 + ALSA |
| `discover_darwin.go` | darwin | AVFoundation |
| `command_windows.go` | windows | `video="name"`, `audio="name"` |
| `command_linux.go` | linux | `/dev/video*`, `hw:X` |
| `command_darwin.go` | darwin | Device indices |

### Data Flow

```
User code
    ↓
GetUserMedia(constraints)
    ↓
EnumerateDevices() → device selection
    ↓
newVideoTrack/newAudioTrack → VideoReader/AudioReader
    ↓
FFmpeg subprocess (stdin ← device, stdout → raw data)
    ↓
Read() → image.YCbCr (YUV420p) | AudioChunk (PCM S16LE)
```

### Output Formats

| Type | Format | Go Type |
|------|--------|---------|
| Video | YUV420p | `*image.YCbCr` |
| Audio | PCM S16LE | `*AudioChunk` (interleaved `[]int16`) |

## Dependencies

- **FFmpeg 8.x** — system-level, required at runtime (PATH or `SetConfig`)
- **Go deps**: `github.com/pion/rtp`, `github.com/google/uuid`, `github.com/denisbrodbeck/machineid`

## H264/RTP Streaming

For network streaming, use `h264_reader.go`:

- `H264VideoReader` — captures and parses H264 NAL units from MPEG-TS
- `RTPReader` — wraps NAL units into RTP packets
- `UDPWriter` — sends RTP packets over UDP
