package mediadevices

// MediaTrackSupportedConstraints 表示浏览器支持的轨道约束。
// 对应 MDN 的 MediaTrackSupportedConstraints 接口。
type MediaTrackSupportedConstraints struct {
	// Width 是否支持宽度约束。
	Width bool
	// Height 是否支持高度约束。
	Height bool
	// FrameRate 是否支持帧率约束。
	FrameRate bool
	// AspectRatio 是否支持宽高比约束。
	AspectRatio bool
	// SampleRate 是否支持采样率约束（音频）。
	SampleRate bool
	// SampleSize 是否支持采样大小约束（音频）。
	SampleSize bool
	// EchoCancellation 是否支持回声消除约束（音频）。
	EchoCancellation bool
	// AutoGainControl 是否支持自动增益控制约束（音频）。
	AutoGainControl bool
	// NoiseSuppression 是否支持噪声抑制约束（音频）。
	NoiseSuppression bool
}

// GetSupportedConstraints 返回当前系统支持的轨道约束。
// 对应 MDN 的 navigator.mediaDevices.getSupportedConstraints()。
func GetSupportedConstraints() MediaTrackSupportedConstraints {
	return MediaTrackSupportedConstraints{
		Width:            true,
		Height:           true,
		FrameRate:        true,
		AspectRatio:      true,
		SampleRate:       true,
		SampleSize:       true,
		EchoCancellation: true,
		AutoGainControl:  true,
		NoiseSuppression: true,
	}
}

// VideoTrackConstraints 表示视频轨道的约束条件。
// 用于 GetUserMedia 调用时指定视频捕获参数。
type VideoTrackConstraints struct {
	// Width 指定期望的视频宽度（像素）。
	Width *int
	// Height 指定期望的视频高度（像素）。
	Height *int
	// FrameRate 指定期望的帧率。
	FrameRate *float64
	// AspectRatio 指定期望的宽高比（宽度/高度）。
	AspectRatio *float64
	// DeviceID 指定使用的设备 ID。
	// 如果为 nil，则使用默认视频设备。
	DeviceID *string
}

// AudioTrackConstraints 表示音频轨道的约束条件。
// 用于 GetUserMedia 调用时指定音频捕获参数。
type AudioTrackConstraints struct {
	// SampleRate 指定期望的采样率（Hz）。
	SampleRate *int
	// Channels 指定期望的声道数（1=单声道，2=立体声）。
	Channels *int
	// EchoCancellation 是否启用回声消除。
	EchoCancellation *bool
	// AutoGainControl 是否启用自动增益控制。
	AutoGainControl *bool
	// NoiseSuppression 是否启用噪声抑制。
	NoiseSuppression *bool
	// DeviceID 指定使用的设备 ID。
	// 如果为 nil，则使用默认音频设备。
	DeviceID *string
}

// MediaTrackConstraints 表示媒体轨道的约束条件。
// 对应 MDN 的 MediaTrackConstraints 接口。
// 可以同时指定视频和音频约束。
type MediaTrackConstraints struct {
	// Video 指定视频轨道约束。
	Video *VideoTrackConstraints
	// Audio 指定音频轨道约束。
	Audio *AudioTrackConstraints
}

// MediaTrackSettings 表示轨道的当前设置。
// 对应 MDN 的 MediaTrackSettings 接口。
// 反映应用请求的约束和设备实际能力的交集。
type MediaTrackSettings struct {
	// Width 视频的实际宽度。
	Width int
	// Height 视频的实际高度。
	Height int
	// FrameRate 视频的实际帧率。
	FrameRate float64
	// AspectRatio 视频的实际宽高比。
	AspectRatio float64
	// SampleRate 音频的实际采样率。
	SampleRate int
	// SampleSize 音频的实际采样大小（位）。
	SampleSize int
	// EchoCancellation 是否启用了回声消除。
	EchoCancellation bool
	// AutoGainControl 是否启用了自动增益控制。
	AutoGainControl bool
	// NoiseSuppression 是否启用了噪声抑制。
	NoiseSuppression bool
}
