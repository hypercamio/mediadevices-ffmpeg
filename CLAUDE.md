# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

An FFmpeg-based Go implementation of media device handling, compatible with the [pion/mediadevices](https://github.com/pion/mediadevices/) interface. The goal is to provide media capture (audio/video) using FFmpeg as the backend while conforming to pion's mediadevices API.

- **Module**: `github.com/hypercamio/mediadevices-ffmpeg-go`
- **Go version**: 1.25.0

## Build Commands

```bash
go build ./...       # Build all packages
go test ./...        # Run all tests
go test ./pkg/...    # Run tests for a specific package subtree
go test -run TestFoo ./path/to/pkg  # Run a single test
go vet ./...         # Static analysis
```

## Key Dependencies (Expected)

- [pion/mediadevices](https://github.com/pion/mediadevices/) — the interface this library implements
- FFmpeg — system-level dependency required at runtime; Go bindings TBD
