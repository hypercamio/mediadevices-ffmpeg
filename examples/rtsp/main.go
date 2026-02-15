package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"

	mediadevices "github.com/hypercamio/mediadevices-ffmpeg"
)

func main() {
	server := flag.String("server", "rtsp://media.hypercam.local:8554", "RTSP server URL")
	width := flag.Int("width", 1280, "Video width")
	height := flag.Int("height", 720, "Video height")
	fps := flag.Float64("fps", 30, "Frame rate")
	bitrate := flag.Int("bitrate", 2000, "Video bitrate in kbps")
	preset := flag.String("preset", "ultrafast", "x264 preset (ultrafast, fast, medium, slow)")
	profile := flag.String("profile", "baseline", "x264 profile (baseline, main, high)")
	deviceFlag := flag.String("device", "", "Device ID or index (skips interactive selection)")
	ffmpegFlag := flag.String("ffmpeg", "", "Path to ffmpeg binary")
	flag.Parse()

	if *ffmpegFlag != "" {
		mediadevices.SetConfig(mediadevices.Config{FFmpegPath: *ffmpegFlag})
	}

	devices, err := mediadevices.VideoInputDevices()
	if err != nil {
		log.Fatalf("Failed to enumerate devices: %v", err)
	}
	if len(devices) == 0 {
		log.Fatal("No video input devices found")
	}

	fmt.Println("Available video devices:")
	for i, d := range devices {
		fmt.Printf("  [%d] %s (ID: %s)\n", i, d.Label, d.DeviceID)
	}

	var device mediadevices.MediaDeviceInfo
	if *deviceFlag != "" {
		found := false
		for _, d := range devices {
			if d.DeviceID == *deviceFlag {
				device = d
				found = true
				break
			}
		}
		if !found {
			if idx, err := strconv.Atoi(*deviceFlag); err == nil && idx >= 0 && idx < len(devices) {
				device = devices[idx]
				found = true
			}
		}
		if !found {
			log.Fatalf("Device not found: %s", *deviceFlag)
		}
	} else if len(devices) == 1 {
		device = devices[0]
	} else {
		fmt.Print("Select device [0]: ")
		var input string
		fmt.Scanln(&input)
		idx := 0
		if input != "" {
			idx, err = strconv.Atoi(input)
			if err != nil || idx < 0 || idx >= len(devices) {
				log.Fatalf("Invalid device index: %s", input)
			}
		}
		device = devices[idx]
	}

	rtspURL := fmt.Sprintf("%s/%s", strings.TrimRight(*server, "/"), device.DeviceID)
	fmt.Printf("\nStreaming: %s -> %s\n", device.Label, rtspURL)

	args := buildRTSPArgs(device.DeviceName, rtspURL, *width, *height, *fps, *bitrate, *preset, *profile)

	cfg := mediadevices.GetConfig()
	ffmpegPath := cfg.FFmpegPath
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	fmt.Printf("FFmpeg: %s %s\n\n", ffmpegPath, strings.Join(args, " "))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			fmt.Println("Stream stopped.")
			return
		}
		log.Fatalf("FFmpeg exited with error: %v", err)
	}
}

func buildRTSPArgs(deviceName, rtspURL string, width, height int, fps float64, bitrate int, preset, profile string) []string {
	var args []string

	// Platform-specific input format
	switch runtime.GOOS {
	case "windows":
		args = append(args, "-f", "dshow")
		args = append(args, "-analyzeduration", "10000000", "-probesize", "10000000")
		args = append(args, "-i", fmt.Sprintf("video=%s", deviceName))
	case "linux":
		args = append(args, "-f", "v4l2")
		args = append(args, "-i", deviceName)
	case "darwin":
		args = append(args, "-f", "avfoundation")
		args = append(args, "-i", deviceName)
	default:
		args = append(args, "-i", deviceName)
	}

	// H.264 encoding with low-latency tuning
	args = append(args, "-c:v", "libx264")
	args = append(args, "-preset", preset)
	args = append(args, "-tune", "zerolatency")

	if width > 0 && height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", width, height))
	}
	if fps > 0 {
		args = append(args, "-r", fmt.Sprintf("%.2f", fps))
	}
	if bitrate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%dk", bitrate))
	}

	// GOP size: 2x fps for ~2 second keyframe interval
	gopSize := int(fps * 2)
	if gopSize <= 0 {
		gopSize = 60
	}
	args = append(args, "-g", fmt.Sprintf("%d", gopSize))

	args = append(args, "-profile:v", profile)
	args = append(args, "-pix_fmt", "yuv420p")
	args = append(args, "-an")
	args = append(args, "-sn")

	// RTSP output over TCP
	args = append(args, "-f", "rtsp")
	args = append(args, "-rtsp_transport", "tcp")
	args = append(args, rtspURL)

	return args
}
