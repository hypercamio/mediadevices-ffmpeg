package mediadevices

import (
	"fmt"
)

// GetUserMedia 请求用户授权并访问摄像头和/或麦克风。
// 对应 MDN 的 navigator.mediaDevices.getUserMedia()。
//
// 参数 constraints 指定请求的媒体类型和约束：
//   - Video: 设置 VideoTrackConstraints 来请求视频
//   - Audio: 设置 AudioTrackConstraints 来请求音频
//   - 同时设置两者可以同时获取音视频
//
// 返回包含请求轨道的 MediaStream。
// 调用方应在使用完毕后调用 stream.Close() 释放资源。
//
// 示例：
//
//	// 仅获取视频
//	stream, err := mediadevices.GetUserMedia(mediadevices.MediaTrackConstraints{
//	    Video: &mediadevices.VideoTrackConstraints{
//	        Width:    IntPtr(1280),
//	        Height:   IntPtr(720),
//	        FrameRate: Float64Ptr(30.0),
//	    },
//	})
//
//	// 同时获取音视频
//	stream, err := mediadevices.GetUserMedia(mediadevices.MediaTrackConstraints{
//	    Video: &mediadevices.VideoTrackConstraints{...},
//	    Audio: &mediadevices.AudioTrackConstraints{...},
//	})
func GetUserMedia(constraints MediaTrackConstraints) (*MediaStream, error) {
	var tracks []*MediaStreamTrack

	// 请求视频
	if constraints.Video != nil {
		track, err := getVideoTrack(constraints.Video)
		if err != nil {
			// 清理已创建的轨道
			for _, t := range tracks {
				t.Stop()
			}
			return nil, fmt.Errorf("getUserMedia video: %w", err)
		}
		tracks = append(tracks, track)
	}

	// 请求音频
	if constraints.Audio != nil {
		track, err := getAudioTrack(constraints.Audio)
		if err != nil {
			// 清理已创建的轨道
			for _, t := range tracks {
				t.Stop()
			}
			return nil, fmt.Errorf("getUserMedia audio: %w", err)
		}
		tracks = append(tracks, track)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("getUserMedia: no constraints specified (neither video nor audio)")
	}

	return newMediaStreamWithTracks(tracks...), nil
}

// getVideoTrack 根据约束创建视频轨道。
func getVideoTrack(constraints *VideoTrackConstraints) (*MediaStreamTrack, error) {
	// 获取设备
	var deviceInfo MediaDeviceInfo
	if constraints.DeviceID != nil {
		// 使用指定的设备
		devices, err := VideoInputDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to get video devices: %w", err)
		}
		found := false
		for _, d := range devices {
			if d.DeviceID == *constraints.DeviceID {
				deviceInfo = d
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("video device not found: %s", *constraints.DeviceID)
		}
	} else {
		// 使用默认设备（第一个可用的视频输入设备）
		devices, err := VideoInputDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to get video devices: %w", err)
		}
		if len(devices) == 0 {
			return nil, fmt.Errorf("no video input devices available")
		}
		deviceInfo = devices[0]
	}

	// 解析约束
	width := 640
	height := 480
	frameRate := 30.0

	if constraints.Width != nil {
		width = *constraints.Width
	}
	if constraints.Height != nil {
		height = *constraints.Height
	}
	if constraints.FrameRate != nil {
		frameRate = *constraints.FrameRate
	}

	return newVideoTrack(deviceInfo, width, height, frameRate)
}

// getAudioTrack 根据约束创建音频轨道。
func getAudioTrack(constraints *AudioTrackConstraints) (*MediaStreamTrack, error) {
	// 获取设备
	var deviceInfo MediaDeviceInfo
	if constraints.DeviceID != nil {
		// 使用指定的设备
		devices, err := AudioInputDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to get audio devices: %w", err)
		}
		found := false
		for _, d := range devices {
			if d.DeviceID == *constraints.DeviceID {
				deviceInfo = d
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("audio device not found: %s", *constraints.DeviceID)
		}
	} else {
		// 使用默认设备（第一个可用的音频输入设备）
		devices, err := AudioInputDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to get audio devices: %w", err)
		}
		if len(devices) == 0 {
			return nil, fmt.Errorf("no audio input devices available")
		}
		deviceInfo = devices[0]
	}

	// 解析约束
	sampleRate := 48000
	channels := 2

	if constraints.SampleRate != nil {
		sampleRate = *constraints.SampleRate
	}
	if constraints.Channels != nil {
		channels = *constraints.Channels
	}

	return newAudioTrack(deviceInfo, sampleRate, channels)
}

// IntPtr 返回指向整数的指针。
// 用于设置约束中的可选整数字段。
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr 返回指向 float64 的指针。
// 用于设置约束中的可选浮点数字段。
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr 返回指向 bool 的指针。
// 用于设置约束中的可选布尔字段。
func BoolPtr(b bool) *bool {
	return &b
}
