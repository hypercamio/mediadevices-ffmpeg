package ffmpeg

import (
	"image"
	"testing"
)

func TestParseYUV420pFrame(t *testing.T) {
	width, height := 4, 2
	ySize := width * height   // 8
	cSize := ySize / 4        // 2
	totalSize := ySize + 2*cSize // 12

	// Build known data: Y plane all 128, Cb all 64, Cr all 192.
	data := make([]byte, totalSize)
	for i := 0; i < ySize; i++ {
		data[i] = 128
	}
	for i := ySize; i < ySize+cSize; i++ {
		data[i] = 64
	}
	for i := ySize + cSize; i < totalSize; i++ {
		data[i] = 192
	}

	img, err := parseYUV420pFrame(data, width, height)
	if err != nil {
		t.Fatalf("parseYUV420pFrame: %v", err)
	}

	if img.Rect != image.Rect(0, 0, width, height) {
		t.Errorf("rect = %v, want %v", img.Rect, image.Rect(0, 0, width, height))
	}
	if img.SubsampleRatio != image.YCbCrSubsampleRatio420 {
		t.Errorf("subsample = %v, want 420", img.SubsampleRatio)
	}

	// Check Y plane values.
	for i, v := range img.Y {
		if v != 128 {
			t.Errorf("Y[%d] = %d, want 128", i, v)
			break
		}
	}
	// Check Cb plane values.
	for i, v := range img.Cb {
		if v != 64 {
			t.Errorf("Cb[%d] = %d, want 64", i, v)
			break
		}
	}
	// Check Cr plane values.
	for i, v := range img.Cr {
		if v != 192 {
			t.Errorf("Cr[%d] = %d, want 192", i, v)
			break
		}
	}
}

func TestParseYUV420pFrame_WrongSize(t *testing.T) {
	_, err := parseYUV420pFrame([]byte{1, 2, 3}, 4, 2)
	if err == nil {
		t.Fatal("expected error for wrong data size")
	}
}

func TestParseYUV420pFrame_LargerFrame(t *testing.T) {
	width, height := 320, 240
	ySize := width * height
	cSize := ySize / 4
	totalSize := ySize + 2*cSize

	data := make([]byte, totalSize)
	// Fill with a gradient pattern.
	for i := range data {
		data[i] = byte(i % 256)
	}

	img, err := parseYUV420pFrame(data, width, height)
	if err != nil {
		t.Fatalf("parseYUV420pFrame: %v", err)
	}

	if img.YStride != width {
		t.Errorf("YStride = %d, want %d", img.YStride, width)
	}
	if img.CStride != (width+1)/2 {
		t.Errorf("CStride = %d, want %d", img.CStride, (width+1)/2)
	}
	if len(img.Y) != ySize {
		t.Errorf("len(Y) = %d, want %d", len(img.Y), ySize)
	}
	if len(img.Cb) != cSize {
		t.Errorf("len(Cb) = %d, want %d", len(img.Cb), cSize)
	}
}
