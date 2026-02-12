package mediadevices

import (
	"fmt"
	"image"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// MediaStreamTrackState 表示轨道的当前状态。
// 对应 MDN 的 MediaStreamTrack.state。
type MediaStreamTrackState string

const (
	// MediaStreamTrackStateLive 表示轨道正在积极产生数据。
	MediaStreamTrackStateLive MediaStreamTrackState = "live"
	// MediaStreamTrackStateEnded 表示轨道已结束，不再产生数据。
	MediaStreamTrackStateEnded MediaStreamTrackState = "ended"
)

// MediaStreamTrack 表示媒体流中的单个轨道。
// 对应 MDN 的 MediaStreamTrack 接口。
// 每个轨道可以是视频或音频。
type MediaStreamTrack struct {
	id          string
	kind        MediaDeviceKind
	label       string
	enabled     atomic.Bool
	readyState  MediaStreamTrackState

	// 内部：实际读取器
	videoReader *VideoReader
	audioReader *AudioReader

	// 用于同步访问
	mu sync.Mutex
}

// newVideoTrack 创建一个新的视频轨道。
func newVideoTrack(deviceInfo MediaDeviceInfo, width, height int, frameRate float64) (*MediaStreamTrack, error) {
	reader, err := newVideoReaderInternal(deviceInfo.DeviceID, width, height, frameRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create video reader: %w", err)
	}

	return &MediaStreamTrack{
		id:          generateTrackID(),
		kind:        MediaDeviceKindVideoInput,
		label:       deviceInfo.Label,
		readyState:  MediaStreamTrackStateLive,
		videoReader:  reader,
	}, nil
}

// newAudioTrack 创建一个新的音频轨道。
func newAudioTrack(deviceInfo MediaDeviceInfo, sampleRate, channels int) (*MediaStreamTrack, error) {
	reader, err := newAudioReaderInternal(deviceInfo.DeviceID, sampleRate, channels)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio reader: %w", err)
	}

	return &MediaStreamTrack{
		id:          generateTrackID(),
		kind:        MediaDeviceKindAudioInput,
		label:       deviceInfo.Label,
		readyState:  MediaStreamTrackStateLive,
		audioReader: reader,
	}, nil
}

// ID 返回轨道的唯一标识符。
// 对应 MDN 的 MediaStreamTrack.id。
func (t *MediaStreamTrack) ID() string {
	return t.id
}

// Kind 返回轨道的类型。
// 对应 MDN 的 MediaStreamTrack.kind。
func (t *MediaStreamTrack) Kind() MediaDeviceKind {
	return t.kind
}

// Label 返回轨道的标签。
// 对应 MDN 的 MediaStreamTrack.label。
func (t *MediaStreamTrack) Label() string {
	return t.label
}

// Enabled 返回轨道是否启用。
// 对应 MDN 的 MediaStreamTrack.enabled。
func (t *MediaStreamTrack) Enabled() bool {
	return t.enabled.Load()
}

// SetEnabled 设置轨道是否启用。
// 对应 MDN 的 MediaStreamTrack.enabled。
// 禁用轨道会产生静音音频或黑帧。
func (t *MediaStreamTrack) SetEnabled(enabled bool) {
	t.enabled.Store(enabled)
}

// ReadyState 返回轨道的就绪状态。
// 对应 MDN 的 MediaStreamTrack.readyState。
func (t *MediaStreamTrack) ReadyState() MediaStreamTrackState {
	return t.readyState
}

// Stop 停止轨道。
// 对应 MDN 的 MediaStreamTrack.stop()。
// 停止后轨道进入 ended 状态。
func (t *MediaStreamTrack) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.readyState == MediaStreamTrackStateEnded {
		return
	}

	if t.videoReader != nil {
		t.videoReader.Close()
		t.videoReader = nil
	}
	if t.audioReader != nil {
		t.audioReader.Close()
		t.audioReader = nil
	}

	t.readyState = MediaStreamTrackStateEnded
}

// Close 是 Stop 的别名，用于与 io.Closer 接口兼容。
func (t *MediaStreamTrack) Close() error {
	t.Stop()
	return nil
}

// Read 读取一帧视频数据。
// 仅在视频轨道上有效。
// 返回 io.EOF 当流结束时。
func (t *MediaStreamTrack) Read() (image.Image, error) {
	if t.kind != MediaDeviceKindVideoInput {
		return nil, fmt.Errorf("cannot read video from non-video track")
	}
	if t.videoReader == nil {
		return nil, io.EOF
	}
	return t.videoReader.Read()
}

// ReadAudio 读取一段音频数据。
// 仅在音频轨道上有效。
// 返回 io.EOF 当流结束时。
func (t *MediaStreamTrack) ReadAudio() (*AudioChunk, error) {
	if t.kind != MediaDeviceKindAudioInput {
		return nil, fmt.Errorf("cannot read audio from non-audio track")
	}
	if t.audioReader == nil {
		return nil, io.EOF
	}
	return t.audioReader.Read()
}

// GetSettings 返回轨道的当前设置。
// 对应 MDN 的 MediaStreamTrack.getSettings()。
func (t *MediaStreamTrack) GetSettings() MediaTrackSettings {
	t.mu.Lock()
	defer t.mu.Unlock()

	settings := MediaTrackSettings{}

	if t.videoReader != nil {
		settings.Width = t.videoReader.Width()
		settings.Height = t.videoReader.Height()
		// FrameRate 需要额外计算或存储
		settings.AspectRatio = float64(settings.Width) / float64(settings.Height)
	}
	if t.audioReader != nil {
		settings.SampleRate = t.audioReader.SampleRate()
		// SampleSize 固定为 16 (S16LE)
		settings.SampleSize = 16
	}

	return settings
}

// MediaStream 表示包含零个或多个 MediaStreamTrack 的媒体流。
// 对应 MDN 的 MediaStream 接口。
type MediaStream struct {
	id     string
	tracks map[string]*MediaStreamTrack
	active atomic.Bool
	mu     sync.RWMutex
}

// NewMediaStream 创建一个新的空媒体流。
func NewMediaStream() *MediaStream {
	return &MediaStream{
		id:     generateStreamID(),
		tracks: make(map[string]*MediaStreamTrack),
	}
}

// newMediaStreamWithTracks 使用指定的轨道创建一个媒体流。
func newMediaStreamWithTracks(tracks ...*MediaStreamTrack) *MediaStream {
	s := NewMediaStream()
	for _, track := range tracks {
		s.tracks[track.id] = track
	}
	if len(tracks) > 0 {
		s.active.Store(true)
	}
	return s
}

// ID 返回流的唯一标识符。
// 对应 MDN 的 MediaStream.id。
func (s *MediaStream) ID() string {
	return s.id
}

// Active 返回流是否活跃。
// 对应 MDN 的 MediaStream.active。
func (s *MediaStream) Active() bool {
	return s.active.Load()
}

// GetTracks 返回流中的所有轨道。
// 对应 MDN 的 MediaStream.getTracks()。
func (s *MediaStream) GetTracks() []*MediaStreamTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tracks := make([]*MediaStreamTrack, 0, len(s.tracks))
	for _, track := range s.tracks {
		tracks = append(tracks, track)
	}
	return tracks
}

// GetVideoTracks 返回流中的所有视频轨道。
// 对应 MDN 的 MediaStream.getVideoTracks()。
func (s *MediaStream) GetVideoTracks() []*MediaStreamTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tracks []*MediaStreamTrack
	for _, track := range s.tracks {
		if track.kind == MediaDeviceKindVideoInput {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

// GetAudioTracks 返回流中的所有音频轨道。
// 对应 MDN 的 MediaStream.getAudioTracks()。
func (s *MediaStream) GetAudioTracks() []*MediaStreamTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tracks []*MediaStreamTrack
	for _, track := range s.tracks {
		if track.kind == MediaDeviceKindAudioInput {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

// GetTrackByID 返回指定 ID 的轨道。
// 对应 MDN 的 MediaStream.getTrackById()。
func (s *MediaStream) GetTrackByID(id string) *MediaStreamTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.tracks[id]
}

// AddTrack 向流中添加轨道。
// 对应 MDN 的 MediaStream.addTrack()。
func (s *MediaStream) AddTrack(track *MediaStreamTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tracks[track.id] = track
	s.active.Store(true)
}

// RemoveTrack 从流中移除轨道。
// 对应 MDN 的 MediaStream.removeTrack()。
func (s *MediaStream) RemoveTrack(track *MediaStreamTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tracks, track.id)
	if len(s.tracks) == 0 {
		s.active.Store(false)
	}
}

// Clone 创建流的副本，包含所有轨道的克隆。
// 对应 MDN 的 MediaStream.clone()。
func (s *MediaStream) Clone() *MediaStream {
	clone := NewMediaStream()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, track := range s.tracks {
		clone.tracks[track.id] = track
	}
	clone.active.Store(s.active.Load())

	return clone
}

// Close 关闭流，停止所有轨道并释放资源。
// 对应 MDN 的 MediaStream 生命周期管理。
func (s *MediaStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, track := range s.tracks {
		track.Stop()
	}
	s.tracks = make(map[string]*MediaStreamTrack)
	s.active.Store(false)
	return nil
}

// generateTrackID 生成唯一的轨道 ID。
func generateTrackID() string {
	return fmt.Sprintf("track-%d", time.Now().UnixNano())
}

// generateStreamID 生成唯一的流 ID。
func generateStreamID() string {
	return fmt.Sprintf("stream-%d", time.Now().UnixNano())
}

// 确保 MediaStreamTrack 满足 io.Closer 接口
var _ io.Closer = (*MediaStreamTrack)(nil)
