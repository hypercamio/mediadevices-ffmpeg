package ffmpeg

import (
	"fmt"
	"image"
)

// parseYUV420pFrame converts raw YUV420p bytes into an *image.YCbCr.
// The input must be exactly width*height*3/2 bytes (Y plane + Cb + Cr).
// The returned image owns its own memory (data is copied).
func parseYUV420pFrame(data []byte, width, height int) (*image.YCbCr, error) {
	ySize := width * height
	cSize := ySize / 4
	expected := ySize + 2*cSize
	if len(data) != expected {
		return nil, fmt.Errorf("YUV420p frame: expected %d bytes (%dx%d), got %d", expected, width, height, len(data))
	}

	chromaW := (width + 1) / 2
	chromaH := (height + 1) / 2

	img := &image.YCbCr{
		Y:              make([]byte, ySize),
		Cb:             make([]byte, cSize),
		Cr:             make([]byte, cSize),
		YStride:        width,
		CStride:        chromaW,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}

	copy(img.Y, data[:ySize])
	copy(img.Cb, data[ySize:ySize+cSize])
	copy(img.Cr, data[ySize+cSize:])

	_ = chromaH // used conceptually for chroma dimensions

	return img, nil
}
