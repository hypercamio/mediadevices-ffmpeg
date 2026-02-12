# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A self-contained Go library for media device capture (audio/video) using FFmpeg 8.x as the backend. Cross-platform support for Windows (dshow), Linux (v4l2/ALSA), and macOS (avfoundation).

- **Module**: `github.com/hypercamio/mediadevices-ffmpeg`
- **Package**: `mediadevices`
- **Go version**: 1.25.0

## Build Commands

```bash
go build ./...       # Build all packages
go test ./...        # Run all tests
go test -run TestFoo .  # Run a single test
go vet ./...         # Static analysis
```

## Key Dependencies

- FFmpeg 8.x — system-level dependency required at runtime (resolved via PATH by default)
- No external Go dependencies — pure stdlib
