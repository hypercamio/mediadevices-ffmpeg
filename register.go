package mediadevices

import (
	"log"
	"sync"
)

var (
	initOnce       sync.Once
	cachedDevices  []MediaDeviceInfo
	cachedDevErr   error
)

// EnumerateDevices 返回系统中所有可用的媒体设备。
// 对应 MDN 的 navigator.mediaDevices.enumerateDevices()。
//
// 使用 FFmpeg 进行设备发现：
// - Windows: 使用 dshow 列出 DirectShow 设备
// - macOS: 使用 avfoundation 列出 AVFoundation 设备
// - Linux: 使用 v4l2 列出视频设备，ALSA 列出音频设备
//
// 如果 FFmpeg 未找到或没有检测到设备，返回空切片而非错误。
func EnumerateDevices() ([]MediaDeviceInfo, error) {
	initOnce.Do(func() {
		cfg := GetConfig()
		cachedDevices, cachedDevErr = discoverDevices(cfg.FFmpegPath)
		if cachedDevErr != nil && cfg.Verbose {
			log.Printf("ffmpeg: device discovery failed: %v", cachedDevErr)
		}
		if cfg.Verbose && cachedDevErr == nil {
			log.Printf("ffmpeg: discovered %d devices", len(cachedDevices))
			for _, d := range cachedDevices {
				log.Printf("ffmpeg:   [%s] %s (id=%s, default=%v)", d.Kind, d.Label, d.DeviceID, d.IsDefault)
			}
		}
	})
	return cachedDevices, cachedDevErr
}

// VideoInputDevices 返回所有可用的视频输入设备。
func VideoInputDevices() ([]MediaDeviceInfo, error) {
	all, err := EnumerateDevices()
	if err != nil {
		return nil, err
	}
	var result []MediaDeviceInfo
	for _, d := range all {
		if d.Kind == MediaDeviceKindVideoInput {
			result = append(result, d)
		}
	}
	return result, nil
}

// AudioInputDevices 返回所有可用的音频输入设备。
func AudioInputDevices() ([]MediaDeviceInfo, error) {
	all, err := EnumerateDevices()
	if err != nil {
		return nil, err
	}
	var result []MediaDeviceInfo
	for _, d := range all {
		if d.Kind == MediaDeviceKindAudioInput {
			result = append(result, d)
		}
	}
	return result, nil
}

// AudioOutputDevices 返回所有可用的音频输出设备。
// 注意：当前实现中 FFmpeg 不支持列出音频输出设备，此函数可能返回空切片。
func AudioOutputDevices() ([]MediaDeviceInfo, error) {
	all, err := EnumerateDevices()
	if err != nil {
		return nil, err
	}
	var result []MediaDeviceInfo
	for _, d := range all {
		if d.Kind == MediaDeviceKindAudioOutput {
			result = append(result, d)
		}
	}
	return result, nil
}
