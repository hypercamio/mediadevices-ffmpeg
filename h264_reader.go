package mediadevices

import (
	"fmt"
	"io"
	"net"

	"github.com/pion/rtp"
)

// H264NaluType represents the type of an H264 NAL unit.
type H264NaluType uint8

const (
	// NALU types
	NALUTypeUnknown     H264NaluType = 0
	NALUTypeSlice      H264NaluType = 1
	NALUTypeDPA        H264NaluType = 2
	NALUTypeDPB        H264NaluType = 3
	NALUTypeIDC        H264NaluType = 4
	NALUTypeSEI        H264NaluType = 5
	NALUTypeSPS        H264NaluType = 7
	NALUTypePPS        H264NaluType = 8
)

// IsKeyframe returns true if the NAL unit is a keyframe.
func (t H264NaluType) IsKeyframe() bool {
	return t == NALUTypeSPS || t == NALUTypePPS || t == 5 // 5 = IDR slice
}

// NALUnit represents a single H264 Network Abstraction Layer Unit.
type NALUnit struct {
	Type      H264NaluType
	Data      []byte
	Keyframe  bool
}

// String returns a string representation of the NAL unit type.
func (n *NALUnit) String() string {
	return fmt.Sprintf("NALU(type=%d, size=%d, keyframe=%v)", n.Type, len(n.Data), n.Keyframe)
}

// H264ReaderConfig holds configuration for creating an H264 video reader.
type H264ReaderConfig struct {
	DeviceName  string // Original device name for FFmpeg (e.g., "USB2.0 HD UVC WebCam")
	DeviceID    string // UUID (kept for backwards compatibility)
	Width       int
	Height      int
	FrameRate   float64
	BitRate     int // in kbps, 0 for default
	KeyInterval int // GOP size, 0 for auto (default 60)
	Profile     string // "baseline", "main", "high"
	Preset      string // "ultrafast", "fast", "medium", "slow"
}

// buildH264Args builds FFmpeg arguments for H264 video capture.
func buildH264Args(cfg H264ReaderConfig) []string {
	args := []string{}

	// Use DeviceName if available, otherwise fallback to DeviceID
	deviceName := cfg.DeviceName
	if deviceName == "" {
		deviceName = cfg.DeviceID
	}

	// Input from DirectShow (Windows)
	args = append(args, "-f", "dshow")
	// For MJPEG cameras, increase analyzeduration and probesize to properly detect stream parameters
	args = append(args, "-analyzeduration", "10000000", "-probesize", "10000000")
	args = append(args, "-i", fmt.Sprintf("video=%s", deviceName))

	// Video encoding settings
	args = append(args, "-c:v", "libx264")

	// Preset for encoding speed vs compression
	preset := cfg.Preset
	if preset == "" {
		preset = "ultrafast"
	}
	args = append(args, "-preset", preset)

	// Tune for low latency streaming
	args = append(args, "-tune", "zerolatency")

	// Resolution
	if cfg.Width > 0 && cfg.Height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", cfg.Width, cfg.Height))
	}

	// Frame rate
	if cfg.FrameRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%.2f", cfg.FrameRate))
	}

	// Bit rate (target)
	if cfg.BitRate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%dk", cfg.BitRate))
	}

	// Key frame interval (GOP size)
	keyInt := cfg.KeyInterval
	if keyInt == 0 {
		keyInt = 60
	}
	args = append(args, "-g", fmt.Sprintf("%d", keyInt))

	// Force IDR frame generation every 30 frames to trigger PPS output
	// This ensures SPS/PPS are output more frequently for proper stream initialization
	args = append(args, "-force_key_frames", "expr:not(mod(n,30))")

	// Profile
	profile := cfg.Profile
	if profile == "" {
		profile = "main"
	}
	args = append(args, "-profile:v", profile)

	// Additional options for low latency
	args = append(args, "-pix_fmt", "yuv420p")
	args = append(args, "-an") // no audio
	args = append(args, "-sn")  // no subtitles

	// Ensure SPS/PPS are sent with every IDR frame for proper stream decoding
	// This is critical for RTSP servers to properly announce the stream
	args = append(args, "-x264-params", "repeatheaders=1")

	// Output format: H264 raw bitstream (annexb) - this ensures SPS/PPS are output as NAL units
	// Using annexb format instead of mpegts to make SPS/PPS extraction easier
	args = append(args, "-f", "h264")
	args = append(args, "pipe:1")

	return args
}

// H264VideoReader reads H264 encoded video frames from an FFmpeg subprocess.
type H264VideoReader struct {
	proc   *ffmpegProcess
	width  int
	height int
}

// newH264VideoReader creates a new H264VideoReader.
func newH264VideoReader(cfg H264ReaderConfig) (*H264VideoReader, error) {
	// Use DeviceName if available, otherwise use DeviceID
	deviceName := cfg.DeviceName
	if deviceName == "" {
		deviceName = cfg.DeviceID
	}
	if deviceName == "" {
		return nil, fmt.Errorf("DeviceName or DeviceID is required")
	}

	args := buildH264Args(cfg)
	gcfg := GetConfig()

	proc, err := startProcess(gcfg.FFmpegPath, args)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg start H264 capture: %w", err)
	}

	return &H264VideoReader{
		proc:  proc,
		width: cfg.Width,
		height: cfg.Height,
	}, nil
}

// Read reads the next H264 NAL unit from the stream.
// Returns nil when the stream ends.
func (r *H264VideoReader) Read() (*NALUnit, error) {
	// Read H.264 NAL units from raw bitstream (annexb format)
	// Each NAL unit is preceded by start code: 0x00 0x00 0x00 0x01 or 0x00 0x00 0x01

	// Read a buffer to find NAL units
	buf := make([]byte, 4096)
	n, err := io.ReadFull(r.proc, buf)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to read H264 data: %w", err)
	}

	// Parse NAL units from the buffer
	nalus := parseH264Bitstream(buf[:n])
	if len(nalus) == 0 {
		return nil, nil
	}

	// Return the first NAL unit
	return nalus[0], nil
}

// parseH264Bitstream parses H.264 raw bitstream (annexb format) and extracts NAL units.
func parseH264Bitstream(data []byte) []*NALUnit {
	var nalus []*NALUnit
	i := 0

	for i < len(data) {
		// Find start code (0x00 0x00 0x00 0x01 or 0x00 0x00 0x01)
		startCodeLen := 0
		for i < len(data)-3 {
			if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
				startCodeLen = 3
				break
			}
			if i < len(data)-4 && data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 {
				startCodeLen = 4
				break
			}
			i++
		}

		if startCodeLen == 0 {
			break
		}

		i += startCodeLen
		if i >= len(data) {
			break
		}

		// Find next start code or end of data
		j := i
		for j < len(data)-4 {
			if data[j] == 0x00 && data[j+1] == 0x00 && data[j+2] == 0x00 && data[j+3] == 0x01 {
				break
			}
			if data[j] == 0x00 && data[j+1] == 0x00 && data[j+2] == 0x01 {
				break
			}
			j++
		}

		nalData := data[i:j]
		if len(nalData) > 0 {
			nalType := H264NaluType(nalData[0] & 0x1F)
			nalus = append(nalus, &NALUnit{
				Type:     nalType,
				Data:     nalData,
				Keyframe: nalType.IsKeyframe(),
			})
		}

		i = j
	}

	return nalus
}

// parseTSPacket parses an MPEG-TS packet and extracts H264 NAL units.
func parseTSPacket(data []byte) ([]*NALUnit, error) {
	if len(data) < 188 {
		return nil, fmt.Errorf("invalid TS packet: too short (%d bytes)", len(data))
	}

	// Check sync byte
	if data[0] != 0x47 {
		return nil, fmt.Errorf("invalid TS sync byte: 0x%02x", data[0])
	}

	var nalus []*NALUnit

	// Parse TS header (first 4 bytes)
	pid := int(data[1]&0x1F)<<8 | int(data[2])
	adaptationFieldControl := (data[3] >> 0) & 0x03

	// Debug: print all PIDs
	_ = pid
	// fmt.Printf("[DEBUG parseTSPacket] PID: 0x%x, adaptation: %d\n", pid, adaptationFieldControl)

	// Skip non-video packets
	// Note: SPS/PPS may be in different PIDs than video
	// Let's be more permissive for debugging
	if pid < 0x10 || pid > 0x1FFE {
		return nil, nil
	}

	// Skip adaptation field if present
	offset := 4
	if adaptationFieldControl&0x02 != 0 {
		if offset >= len(data) {
			return nil, nil
		}
		adaptationFieldLength := int(data[offset])
		offset += 1 + adaptationFieldLength
	}

	if offset >= len(data) {
		return nil, nil
	}

	// Look for PES header start (0x00 0x00 0x01)
	pesStart := -1
	for i := offset; i < len(data)-3; i++ {
		if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
			pesStart = i + 3
			break
		}
	}

	if pesStart == -1 {
		return nil, nil // No PES header found
	}

	// PES header: skip stream_id and PES length
	if pesStart+2 >= len(data) {
		return nil, nil
	}
	pesStart += 2

	// Skip PES optional header if present
	if pesStart >= len(data) {
		return nil, nil
	}
	pesHeaderLength := int(data[pesStart])
	pesStart += 1 + pesHeaderLength

	if pesStart >= len(data) {
		return nil, nil
	}

	// Extract NAL units from PES payload
	pesPayload := data[pesStart:]
	nalus = append(nalus, parseNALUnits(pesPayload)...)

	return nalus, nil
}

// parseNALUnits parses a slice of PES payload data and extracts NAL units.
func parseNALUnits(data []byte) []*NALUnit {
	var nalus []*NALUnit
	i := 0

	for i < len(data) {
		// Find start code (0x00 0x00 0x00 0x01 or 0x00 0x00 0x01)
		startCodeLen := 0
		for i < len(data)-3 {
			if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
				startCodeLen = 3
				break
			}
			if data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 {
				startCodeLen = 4
				break
			}
			i++
		}

		if startCodeLen == 0 {
			break
		}

		i += startCodeLen
		if i >= len(data) {
			break
		}

		// Find next start code or end of data
		j := i
		for j < len(data)-4 {
			if data[j] == 0x00 && data[j+1] == 0x00 && data[j+2] == 0x00 && data[j+3] == 0x01 {
				break
			}
			if data[j] == 0x00 && data[j+1] == 0x00 && data[j+2] == 0x01 {
				break
			}
			j++
		}

		nalData := data[i:j]
		if len(nalData) > 0 {
			nalType := H264NaluType(nalData[0] & 0x1F)
			nalus = append(nalus, &NALUnit{
				Type:     nalType,
				Data:     nalData,
				Keyframe: nalType.IsKeyframe(),
			})
		}

		i = j
	}

	return nalus
}

// Width returns the video width in pixels.
func (r *H264VideoReader) Width() int {
	return r.width
}

// Height returns the video height in pixels.
func (r *H264VideoReader) Height() int {
	return r.height
}

// Close stops the FFmpeg subprocess and releases resources.
func (r *H264VideoReader) Close() error {
	if r.proc != nil {
		return r.proc.Stop()
	}
	return nil
}

// RTPReader reads H264 data and packages it into RTP packets.
type RTPReader struct {
	reader *H264VideoReader
	ssrc   uint32
	seq    uint16
	ts     uint32
	mtu    int

	// Cached SPS/PPS for keyframe injection
	sps []byte
	pps []byte
}

// NewRTPReader creates a new RTP reader for H264 video streaming.
func NewRTPReader(cfg H264ReaderConfig, initialSSRC uint32, mtu int) (*RTPReader, error) {
	reader, err := newH264VideoReader(cfg)
	if err != nil {
		return nil, err
	}

	if mtu <= 0 || mtu > 1500 {
		mtu = 1200 // Safe default for RTP over UDP
	}

	return &RTPReader{
		reader: reader,
		ssrc:   initialSSRC,
		seq:    uint16(initialSSRC),
		ts:     0,
		mtu:    mtu,
	}, nil
}

// Read reads the next RTP packet.
func (r *RTPReader) Read() (*rtp.Packet, error) {
	for {
		nal, err := r.reader.Read()
		if err != nil {
			return nil, err
		}
		if nal == nil {
			continue
		}

		return r.nalToRTP(nal)
	}
}

// ReadMultiple reads all RTP packets for the current NAL unit.
func (r *RTPReader) ReadMultiple() ([]*rtp.Packet, error) {
	for {
		nal, err := r.reader.Read()
		if err != nil {
			return nil, err
		}
		if nal == nil {
			continue
		}

		// Cache SPS/PPS when found
		if r.sps == nil && nal.Type == NALUTypeSPS {
			r.sps = make([]byte, len(nal.Data))
			copy(r.sps, nal.Data)
		}
		if r.pps == nil && nal.Type == NALUTypePPS {
			r.pps = make([]byte, len(nal.Data))
			copy(r.pps, nal.Data)
		}

		return r.nalToRTPMultiple(nal)
	}
}

// PeekNAL returns the current NAL unit without consuming it.
// Returns nil if no NAL unit is available.
func (r *RTPReader) PeekNAL() (*NALUnit, error) {
	// Note: This is a simplified implementation that reads and caches the NAL
	// In a production implementation, you might want to use a buffer
	return r.reader.Read()
}

// GetSPSPPS returns the cached SPS and PPS.
// Returns nil if not yet extracted.
func (r *RTPReader) GetSPSPPS() ([]byte, []byte) {
	return r.sps, r.pps
}

// nalToRTP converts an H264 NAL unit to RTP packet.
func (r *RTPReader) nalToRTP(nal *NALUnit) (*rtp.Packet, error) {
	nalLen := len(nal.Data)
	maxPayloadSize := r.mtu - 20 // Reserve space for IP/UDP headers

	if nalLen <= maxPayloadSize-12 {
		// Single NAL unit packet
		return &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				Marker:         true,
				PayloadType:    96,
				SequenceNumber: r.nextSeq(),
				Timestamp:      r.nextTS(),
				SSRC:          r.ssrc,
			},
			Payload: nal.Data,
		}, nil
	}

	// Fragmentation Unit (FU) for large NAL units
	packets, err := r.nalToRTPMultiple(nal)
	if err != nil {
		return nil, err
	}

	if len(packets) > 0 {
		return packets[0], nil
	}

	return nil, fmt.Errorf("failed to create RTP packet")
}

// nalToRTPMultiple converts an H264 NAL unit to multiple RTP packets.
func (r *RTPReader) nalToRTPMultiple(nal *NALUnit) ([]*rtp.Packet, error) {
	nalLen := len(nal.Data)
	maxPayloadSize := r.mtu - 20

	if nalLen <= maxPayloadSize-12 {
		return []*rtp.Packet{
			{
				Header: rtp.Header{
					Version:        2,
					Marker:         true,
					PayloadType:    96,
					SequenceNumber: r.nextSeq(),
					Timestamp:      r.nextTS(),
					SSRC:          r.ssrc,
				},
				Payload: nal.Data,
			},
		}, nil
	}

	// Fragmentation Unit
	fuIndicator := uint8(28) | (nal.Data[0] & 0xE0) // NRI from original NAL
	fuHeader := uint8(0x00)
	payloadData := nal.Data[1:]
	offset := 0
	var packets []*rtp.Packet

	for offset < len(payloadData) {
		isLast := offset+maxPayloadSize-14 >= len(payloadData)
		fuH := fuHeader
		if offset == 0 {
			fuH |= 0x80 // S bit (start)
		}
		if isLast {
			fuH |= 0x40 // E bit (end)
		}

		chunkSize := len(payloadData) - offset
		if chunkSize > maxPayloadSize-14 {
			chunkSize = maxPayloadSize - 14
		}

		payload := []byte{fuIndicator, fuH}
		payload = append(payload, payloadData[offset:offset+chunkSize]...)

		packets = append(packets, &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				Marker:         isLast && nal.Keyframe,
				PayloadType:    96,
				SequenceNumber: r.nextSeq(),
				Timestamp:      r.nextTS(),
				SSRC:          r.ssrc,
			},
			Payload: payload,
		})

		offset += chunkSize
	}

	return packets, nil
}

func (r *RTPReader) nextSeq() uint16 {
	r.seq++
	return r.seq
}

func (r *RTPReader) nextTS() uint32 {
	// 90kHz timestamp clock (standard for MPEG)
	r.ts += 3000 // 30fps = 3000 ticks per frame
	return r.ts
}

// Close closes the RTP reader and underlying video reader.
func (r *RTPReader) Close() error {
	return r.reader.Close()
}

// Width returns the video width.
func (r *RTPReader) Width() int {
	return r.reader.Width()
}

// Height returns the video height.
func (r *RTPReader) Height() int {
	return r.reader.Height()
}

// UDPWriter is a helper for writing RTP packets over UDP.
type UDPWriter struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	payload int
}

// NewUDPWriter creates a new UDP writer for RTP streaming.
func NewUDPWriter(addr string, mtu int) (*UDPWriter, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve UDP addr: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("dial UDP: %w", err)
	}

	if mtu <= 0 {
		mtu = 1500
	}

	return &UDPWriter{
		conn:    conn,
		addr:    udpAddr,
		payload: mtu - 20 - 8, // MTU - IP header - UDP header
	}, nil
}

// WritePacket writes an RTP packet over UDP.
func (w *UDPWriter) WritePacket(pkt *rtp.Packet) error {
	data, err := pkt.Marshal()
	if err != nil {
		return err
	}
	_, err = w.conn.WriteToUDP(data, w.addr)
	return err
}

// Write writes raw data over UDP.
func (w *UDPWriter) Write(data []byte) error {
	_, err := w.conn.WriteToUDP(data, w.addr)
	return err
}

// Close closes the UDP connection.
func (w *UDPWriter) Close() error {
	return w.conn.Close()
}

// LocalAddr returns the local UDP address.
func (w *UDPWriter) LocalAddr() *net.UDPAddr {
	return w.conn.LocalAddr().(*net.UDPAddr)
}

// H264CodecInfo contains H264 codec parameters.
type H264CodecInfo struct {
	Profile     string
	Level       string
	Width       int
	Height      int
	PixelFormat string
	SPS         []byte
	PPS         []byte
}

// ExtractH264Info extracts SPS and PPS from the first keyframes.
func ExtractH264Info(data []byte) *H264CodecInfo {
	return &H264CodecInfo{
		Profile:     "main",
		Level:       "4.0",
		PixelFormat: "yuv420p",
	}
}

// IsKeyframe checks if the NAL unit is a keyframe.
func IsKeyframe(data []byte) bool {
	if len(data) < 1 {
		return false
	}
	nalType := H264NaluType(data[0] & 0x1F)
	return nalType.IsKeyframe()
}

// nalTypeString returns a string representation of the NAL unit type.
