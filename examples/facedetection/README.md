## Face Detection Example

Real-time face detection using [mediadevices-ffmpeg-go](https://github.com/hypercamio/mediadevices-ffmpeg-go) for video capture and [pigo](https://github.com/esimov/pigo) for face detection.

### Prerequisites

- Go 1.25+
- FFmpeg 8.x installed and available in PATH
- A connected webcam

### Build and Run

```
cd examples/facedetection
go build
./facedetection
```

The program captures video at 640x480 from the first available camera and checks for faces at ~4 fps. You should see log output when a face is detected.
