package mediadevices

import "encoding/json"

// MediaDeviceKind 表示设备类型，对应 MDN 的 kind 属性。
// 用于标识设备是视频输入、音频输入还是音频输出设备。
type MediaDeviceKind string

const (
	// MediaDeviceKindVideoInput 表示视频输入设备，如摄像头。
	MediaDeviceKindVideoInput MediaDeviceKind = "videoinput"
	// MediaDeviceKindAudioInput 表示音频输入设备，如麦克风。
	MediaDeviceKindAudioInput MediaDeviceKind = "audioinput"
	// MediaDeviceKindAudioOutput 表示音频输出设备，如扬声器。
	MediaDeviceKindAudioOutput MediaDeviceKind = "audiooutput"
)

// MediaDeviceInfo 表示单个媒体设备的信息，对应 MDN 的 MediaDeviceInfo 接口。
// 包含设备的唯一标识、类型、标签等信息。
type MediaDeviceInfo struct {
	// DeviceID 是设备的唯一标识符。
	// 在 Windows (dshow): 设备名称字符串。
	// 在 Linux: 设备路径 (如 "/dev/video0") 或 ALSA ID (如 "hw:0,0")。
	// 在 macOS (avfoundation): 设备索引字符串 (如 "0", "1")。
	DeviceID string

	// GroupID 是同属一个物理设备的组 ID。
	// 相同物理设备的不同捕获点（如同一摄像头的不同焦距）会有相同的 GroupID。
	GroupID string

	// Kind 表示设备类型（视频输入、音频输入或音频输出）。
	Kind MediaDeviceKind

	// Label 是设备的可读名称。
	// 如果隐私设置阻止访问设备信息，Label 可能为空字符串。
	Label string

	// IsDefault 表示该设备是否是系统默认设备。
	IsDefault bool
}

// ToJSON 将 MediaDeviceInfo 转换为 JSON 兼容的 map。
// 适用于调试日志或与其他系统集成。
func (m *MediaDeviceInfo) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"deviceId":  m.DeviceID,
		"groupId":   m.GroupID,
		"kind":      string(m.Kind),
		"label":     m.Label,
		"isDefault": m.IsDefault,
	}
}

// MarshalJSON 实现 json.Marshaler 接口。
func (m *MediaDeviceInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.ToJSON())
}
