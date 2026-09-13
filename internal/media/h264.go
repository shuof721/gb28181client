package media

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

// H264Source 输出 Annex-B 访问单元（AU）：一帧画面对应的全部 NAL。
// 每个 NAL 以 00 00 00 01 开头。
type H264Source interface {
	Next() ([]byte, error)
	Seek(offsetSec float64, fps int) (actualSec float64, err error)
	Close() error
}

// nalUnit 单个 NAL（含起始码）。
type nalUnit struct {
	start int // 在原始 data 中的起始下标（含 start code）
	end   int
	typ   byte
}

// FileSource 循环读取 Annex-B H.264 文件。
type FileSource struct {
	path string
	data []byte
	nals []nalUnit
	// 预解析好的访问单元（每个 AU 是 data 切片）
	aus        [][]byte
	sps        []byte
	pps        []byte
	idrIndices []int // 包含 IDR 帧的 AU 下标
	pos        int   // 下一个 AU 下标
	forceI     bool
	mu         sync.Mutex
}

func NewFileSource(path string) (*FileSource, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) < 8 {
		return nil, fmt.Errorf("h264 file too small: %s", path)
	}
	s := &FileSource{path: path, data: b}
	s.nals = splitNals(b)
	if len(s.nals) == 0 {
		return nil, fmt.Errorf("no NAL units found in %s", path)
	}
	s.aus = buildAccessUnits(b, s.nals)
	if len(s.aus) == 0 {
		return nil, fmt.Errorf("no access units found in %s", path)
	}
	for i, au := range s.aus {
		if auHasIDR(au) {
			s.idrIndices = append(s.idrIndices, i)
		}
	}
	s.captureParamSets()
	// 会话从 IDR 开始，避免开头就是 P 帧导致花屏
	s.skipToIDR()
	return s, nil
}

func (s *FileSource) Close() error { return nil }

func (s *FileSource) captureParamSets() {
	for _, n := range s.nals {
		nal := s.data[n.start:n.end]
		switch n.typ {
		case 7:
			if s.sps == nil {
				s.sps = append([]byte(nil), nal...)
			}
		case 8:
			if s.pps == nil {
				s.pps = append([]byte(nil), nal...)
			}
		}
	}
}

func (s *FileSource) skipToIDR() {
	if len(s.idrIndices) > 0 {
		s.pos = s.idrIndices[0]
		return
	}
	for i, au := range s.aus {
		if auHasIDR(au) {
			s.pos = i
			return
		}
	}
	s.pos = 0
}

// Seek 根据 offsetSec（秒）定位到前置最近的 IDR 关键帧。
// 若素材时长小于请求偏移量，会按素材总时长循环取模。
func (s *FileSource) Seek(offsetSec float64, fps int) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	totalFrames := len(s.aus)
	if totalFrames == 0 {
		return 0, io.EOF
	}
	if fps <= 0 {
		fps = 25
	}
	if offsetSec < 0 {
		offsetSec = 0
	}

	durationSec := float64(totalFrames) / float64(fps)
	targetSec := offsetSec
	if durationSec > 0 && offsetSec >= durationSec {
		targetSec = math.Mod(offsetSec, durationSec)
	}

	targetFrame := int(targetSec * float64(fps))
	if targetFrame >= totalFrames {
		targetFrame = totalFrames - 1
	}

	// 查找 <= targetFrame 的最近 IDR 帧
	chosenIDR := 0
	if len(s.idrIndices) > 0 {
		chosenIDR = s.idrIndices[0]
		for _, idx := range s.idrIndices {
			if idx <= targetFrame {
				chosenIDR = idx
			} else {
				break
			}
		}
	}

	s.pos = chosenIDR
	s.forceI = true // 标记在跳转后的下一次 Next() 强制带上 SPS/PPS

	// 计算实际对齐的相对秒数
	deltaSec := float64(targetFrame-chosenIDR) / float64(fps)
	actualSec := offsetSec - deltaSec
	if actualSec < 0 {
		actualSec = 0
	}
	return actualSec, nil
}

// ForceIFrame 寻找当前播放位置之后的下一个最近 IDR 帧（若无则循环至首个 IDR），并标记下一次 Next() 强制输出 SPS/PPS。
func (s *FileSource) ForceIFrame() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.idrIndices) > 0 {
		found := false
		for _, idx := range s.idrIndices {
			if idx >= s.pos {
				s.pos = idx
				found = true
				break
			}
		}
		if !found {
			s.pos = s.idrIndices[0]
		}
	} else {
		s.skipToIDR()
	}
	s.forceI = true
}

func (s *FileSource) Next() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.aus) == 0 {
		return nil, io.EOF
	}
	if s.pos >= len(s.aus) {
		s.pos = 0
		// 循环回开头时再次对齐 IDR
		s.skipToIDR()
	}
	au := s.aus[s.pos]
	s.pos++

	// 每个 IDR 前强制补 SPS/PPS，解决中途进流花屏；或经过 Seek 后 forceI 标记
	out := au
	isIDR := auHasIDR(au)
	if (isIDR || s.forceI) && s.sps != nil && s.pps != nil {
		if !bytes.Contains(au, s.sps) {
			out = concatNALs(s.sps, s.pps, au)
		}
		s.forceI = false
	}
	return out, nil
}

func concatNALs(parts ...[]byte) []byte {
	var b bytes.Buffer
	for _, p := range parts {
		b.Write(p)
	}
	return b.Bytes()
}

func auHasIDR(au []byte) bool {
	for _, n := range splitNals(au) {
		if n.typ == 5 {
			return true
		}
	}
	return false
}

// splitNals 扫描 Annex-B，返回全部 NAL。
func splitNals(data []byte) []nalUnit {
	var nals []nalUnit
	i := 0
	for i < len(data) {
		_, n := startCodeLen(data[i:])
		if n == 0 {
			i++
			continue
		}
		nalStart := i + n
		if nalStart >= len(data) {
			break
		}
		// 找下一个 start code
		j := nalStart + 1
		for j < len(data) {
			if _, n2 := startCodeLen(data[j:]); n2 > 0 {
				break
			}
			j++
		}
		nals = append(nals, nalUnit{
			start: i, // 含起始码
			end:   j,
			typ:   data[nalStart] & 0x1F,
		})
		i = j
	}
	return nals
}

// buildAccessUnits 把 NAL 合成访问单元。
// 规则（够用且稳）：
//   - SPS/PPS/SEI/AUD 等非 VCL 挂到「下一幅」画面
//   - 遇到 VCL(type 1/5) 时，若 first_mb_in_slice==0 且已有 VCL，则开新 AU
//   - 同一画面的多个 slice（first_mb!=0）并入当前 AU
func buildAccessUnits(data []byte, nals []nalUnit) [][]byte {
	var aus [][]byte
	var cur []byte
	hasVCL := false

	flush := func() {
		if len(cur) > 0 {
			aus = append(aus, cur)
		}
		cur = nil
		hasVCL = false
	}

	for _, n := range nals {
		nal := data[n.start:n.end]
		isVCL := n.typ == 1 || n.typ == 5

		if isVCL {
			newPic := !hasVCL || firstMBIsZero(nal)
			if hasVCL && newPic {
				flush()
			}
			cur = append(cur, nal...)
			hasVCL = true
			continue
		}

		// 非 VCL：若当前 AU 已有 VCL，且遇到新的参数集/AUD，通常属于下一幅
		// 这里简单处理：已有 VCL 时，SPS/PPS/AUD 触发切分；SEI 跟在当前帧后面问题不大
		if hasVCL && (n.typ == 7 || n.typ == 8 || n.typ == 9) {
			flush()
		}
		cur = append(cur, nal...)
	}
	flush()
	return aus
}

// firstMBIsZero 判断 slice header 的 first_mb_in_slice 是否为 0（新画面起始）。
// 需要跳过 NAL header 后做无符号 Exp-Golomb。
func firstMBIsZero(nalWithSC []byte) bool {
	// 去掉起始码
	_, sc := startCodeLen(nalWithSC)
	if sc == 0 || sc >= len(nalWithSC) {
		return false
	}
	nalHeader := nalWithSC[sc]
	nalType := nalHeader & 0x1F
	if nalType != 1 && nalType != 5 {
		return false
	}
	rbsp := nalWithSC[sc+1:]
	// 去 emulation prevention
	rbsp = removeEmulation(rbsp)
	if len(rbsp) == 0 {
		return false
	}
	// first_mb_in_slice = ue(v)
	v, _ := readUE(rbsp)
	return v == 0
}

func removeEmulation(b []byte) []byte {
	if len(b) < 3 {
		return b
	}
	out := make([]byte, 0, len(b))
	zeros := 0
	for i := 0; i < len(b); i++ {
		if zeros >= 2 && b[i] == 0x03 && i+1 < len(b) && b[i+1] <= 0x03 {
			zeros = 0
			continue // 跳过 0x03
		}
		out = append(out, b[i])
		if b[i] == 0 {
			zeros++
		} else {
			zeros = 0
		}
	}
	return out
}

// readUE 读无符号 Exp-Golomb。
func readUE(b []byte) (uint, int) {
	// bit reader
	val := uint(0)
	n := 0
	for i := 0; i < len(b); i++ {
		for k := 7; k >= 0; k-- {
			bit := (b[i] >> k) & 1
			n++
			if bit == 1 {
				// 读 n-1 个后缀位
				suffix := uint(0)
				need := n - 1
				read := 0
				for j := i; j < len(b) && read < need; j++ {
					startBit := 7
					if j == i {
						startBit = k - 1
						if startBit < 0 {
							continue
						}
					}
					for kk := startBit; kk >= 0 && read < need; kk-- {
						suffix = (suffix << 1) | uint((b[j]>>kk)&1)
						read++
					}
				}
				val = (uint(1) << uint(need)) - 1 + suffix
				return val, n
			}
			if n > 32 {
				return 0, n
			}
		}
	}
	return 0, n
}

func startCodeLen(b []byte) (int, int) {
	if len(b) >= 4 && b[0] == 0 && b[1] == 0 && b[2] == 0 && b[3] == 1 {
		return 4, 4
	}
	if len(b) >= 3 && b[0] == 0 && b[1] == 0 && b[2] == 1 {
		return 3, 3
	}
	return 0, 0
}

// SyntheticSource 生成 I_PCM 宏块的 H.264 IDR 流。
type SyntheticSource struct {
	w, h, fps int
	frameIdx  int
	sps, pps  []byte
	mu        sync.Mutex
	lastTime  time.Time
}

func NewSyntheticSource(w, h, fps int) (*SyntheticSource, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid size")
	}
	if fps <= 0 {
		fps = 25
	}
	w = (w / 16) * 16
	h = (h / 16) * 16
	if w < 16 || h < 16 {
		return nil, fmt.Errorf("size too small")
	}
	s := &SyntheticSource{w: w, h: h, fps: fps}
	s.sps = buildSPS(w, h)
	s.pps = buildPPS()
	return s, nil
}

func (s *SyntheticSource) Close() error { return nil }

func (s *SyntheticSource) Seek(offsetSec float64, fps int) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if offsetSec < 0 {
		offsetSec = 0
	}
	if fps <= 0 {
		fps = s.fps
	}
	s.frameIdx = int(offsetSec * float64(fps))
	s.lastTime = time.Time{}
	return offsetSec, nil
}

func (s *SyntheticSource) Next() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.lastTime.IsZero() {
		interval := time.Second / time.Duration(s.fps)
		d := interval - time.Since(s.lastTime)
		if d > 0 {
			time.Sleep(d)
		}
	}
	s.lastTime = time.Now()
	s.frameIdx++
	idr := buildIPCMIDR(s.w, s.h, s.frameIdx)
	var buf bytes.Buffer
	buf.Write(s.sps)
	buf.Write(s.pps)
	buf.Write(idr)
	return buf.Bytes(), nil
}

func buildSPS(w, h int) []byte {
	e := newBitWriter()
	e.writeUint(66, 8)
	e.writeUint(0, 8)
	e.writeUint(30, 8)
	e.writeUE(0)
	e.writeUE(0)
	e.writeUE(0)
	e.writeUE(0)
	e.writeUE(1)
	e.writeBit(0)
	e.writeUE(uint(w/16 - 1))
	e.writeUE(uint(h/16 - 1))
	e.writeBit(1)
	e.writeBit(1)
	e.writeBit(0)
	e.writeBit(0)
	e.rbspTrailing()
	return annexB(7, e.bytes())
}

func buildPPS() []byte {
	e := newBitWriter()
	e.writeUE(0)
	e.writeUE(0)
	e.writeBit(0)
	e.writeBit(0)
	e.writeUE(0)
	e.writeUE(0)
	e.writeUE(0)
	e.writeBit(0)
	e.writeUint(0, 2)
	e.writeSE(0)
	e.writeSE(0)
	e.writeSE(0)
	e.writeBit(0)
	e.writeBit(0)
	e.writeBit(0)
	e.rbspTrailing()
	return annexB(8, e.bytes())
}

func buildIPCMIDR(w, h, frameIdx int) []byte {
	mbsX := w / 16
	mbsY := h / 16
	e2 := newBitWriter()
	e2.writeUE(0)
	e2.writeUE(7)
	e2.writeUE(0)
	e2.writeUint(uint(frameIdx&0xF), 4)
	e2.writeUE(0)
	e2.writeUint(uint((frameIdx*2)&0xF), 4)
	e2.writeBit(0)
	e2.writeBit(0)
	e2.writeSE(0)

	for my := 0; my < mbsY; my++ {
		for mx := 0; mx < mbsX; mx++ {
			e2.writeUE(25)
			e2.byteAlign()
			y := byte(0x80)
			if (mx+my+frameIdx)%8 < 4 {
				y = 0x40 + byte((mx*3+frameIdx*2)%80)
			} else {
				y = 0xA0 - byte((my*2+frameIdx)%60)
			}
			for i := 0; i < 256; i++ {
				e2.writeByte(y)
			}
			cb := byte(0x80 + (mx*2)%40)
			for i := 0; i < 64; i++ {
				e2.writeByte(cb)
			}
			cr := byte(0x80 - (my*2)%40)
			for i := 0; i < 64; i++ {
				e2.writeByte(cr)
			}
		}
	}
	e2.rbspTrailing()
	return annexB(5, e2.bytes())
}

func annexB(nalType byte, rbsp []byte) []byte {
	nalHeader := byte(0x60) | nalType
	out := []byte{0x00, 0x00, 0x00, 0x01, nalHeader}
	out = append(out, emulationPrevent(rbsp)...)
	return out
}

func emulationPrevent(b []byte) []byte {
	var out bytes.Buffer
	zeros := 0
	for _, c := range b {
		if zeros >= 2 && c <= 3 {
			out.WriteByte(0x03)
			zeros = 0
		}
		out.WriteByte(c)
		if c == 0 {
			zeros++
		} else {
			zeros = 0
		}
	}
	return out.Bytes()
}

type bitWriter struct {
	buf  []byte
	cur  byte
	nbit int
}

func newBitWriter() *bitWriter { return &bitWriter{} }

func (w *bitWriter) writeBit(b int) {
	w.cur = (w.cur << 1) | byte(b&1)
	w.nbit++
	if w.nbit == 8 {
		w.buf = append(w.buf, w.cur)
		w.cur = 0
		w.nbit = 0
	}
}

func (w *bitWriter) writeUint(v uint, n int) {
	for i := n - 1; i >= 0; i-- {
		w.writeBit(int(v >> i & 1))
	}
}

func (w *bitWriter) writeByte(b byte) {
	w.byteAlign()
	w.buf = append(w.buf, b)
}

func (w *bitWriter) writeUE(v uint) {
	v++
	n := 0
	for tmp := v; tmp > 0; tmp >>= 1 {
		n++
	}
	for i := 0; i < n-1; i++ {
		w.writeBit(0)
	}
	w.writeUint(v, n)
}

func (w *bitWriter) writeSE(v int) {
	var u uint
	if v <= 0 {
		u = uint(-v) * 2
	} else {
		u = uint(v)*2 - 1
	}
	w.writeUE(u)
}

func (w *bitWriter) byteAlign() {
	for w.nbit != 0 {
		w.writeBit(0)
	}
}

func (w *bitWriter) rbspTrailing() {
	w.writeBit(1)
	w.byteAlign()
}

func (w *bitWriter) bytes() []byte {
	w.byteAlign()
	return w.buf
}

// SourceOptions 媒体源参数。
type SourceOptions struct {
	Kind      string
	H264      string
	MP4       string
	Width     int
	Height    int
	FPS       int
	ChannelID string
	GetPTZ    func() PTZInfo
}

// NewSource 按配置创建源。
func NewSource(opts SourceOptions) (H264Source, error) {
	switch strings.ToLower(opts.Kind) {
	case "file":
		return NewFileSource(opts.H264)
	case "mp4":
		return NewMP4Source(opts.MP4)
	case "ptz", "ptz-synthetic", "ptz_synthetic":
		return NewPTZSource(PTZSourceOptions{
			ChannelID: opts.ChannelID,
			Width:     opts.Width,
			Height:    opts.Height,
			FPS:       opts.FPS,
			GetPTZ:    opts.GetPTZ,
		})
	case "synthetic", "":
		return NewSyntheticSource(opts.Width, opts.Height, opts.FPS)
	default:
		return nil, fmt.Errorf("unknown media source %q (want file/mp4/synthetic/ptz)", opts.Kind)
	}
}
