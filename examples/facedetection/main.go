package main

import (
	"image"
	"log"
	"os"
	"time"

	pigo "github.com/esimov/pigo/core"
	"github.com/hypercamio/mediadevices-ffmpeg"
)

const (
	confidenceLevel = 5.0
)

var (
	cascade    []byte
	classifier *pigo.Pigo
)

func detectFace(frame *image.YCbCr) bool {
	bounds := frame.Bounds()
	cascadeParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     600,
		ShiftFactor: 0.15,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: frame.Y, // Y in YCbCr should be enough to detect faces
			Rows:   bounds.Dy(),
			Cols:   bounds.Dx(),
			Dim:    bounds.Dx(),
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cascadeParams, 0.0)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0)

	for _, det := range dets {
		if det.Q >= confidenceLevel {
			return true
		}
	}

	return false
}

func main() {
	// prepare face detector
	var err error
	cascade, err = os.ReadFile("facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
	}
	p := pigo.NewPigo()

	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = p.Unpack(cascade)
	if err != nil {
		log.Fatalf("Error unpacking the cascade file: %s", err)
	}

	// Discover available video devices
	devices, err := mediadevices.VideoInputDevices()
	if err != nil {
		log.Fatalf("Error discovering video devices: %s", err)
	}
	if len(devices) == 0 {
		log.Fatal("No video devices found")
	}
	log.Printf("Using video device: %s", devices[0].Label)

	// Request video access using GetUserMedia
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaTrackConstraints{
		Video: &mediadevices.VideoTrackConstraints{
			Width:    mediadevices.IntPtr(640),
			Height:   mediadevices.IntPtr(480),
			FrameRate: mediadevices.Float64Ptr(30.0),
		},
	})
	if err != nil {
		log.Fatalf("Error creating video stream: %s", err)
	}
	defer stream.Close()

	// Get the first video track
	videoTracks := stream.GetVideoTracks()
	if len(videoTracks) == 0 {
		log.Fatal("No video tracks found")
	}
	track := videoTracks[0]

	// To save resources, we can simply use 4 fps to detect faces.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		frame, err := track.Read()
		if err != nil {
			log.Fatalf("Error reading video frame: %s", err)
		}

		if detectFace(frame.(*image.YCbCr)) {
			log.Println("Detect a face")
		}
	}
}
