package media

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AudioSource 伴音源接口，用于实时视频点播/回放时的音频复合流。
type AudioSource interface {
	// NextChunk 生成接下来一段音频样本（按视频帧间隔所需样本数，如 25fps 下为 320 采样点）。
	// 返回：G.711 编码后的字节流、原始 PCM16 采样（供计算 VU 电平）、错误信息
	NextChunk(samples int) ([]byte, []int16, error)
	// Close 释放资源
	Close() error
}

// AudioSourceOptions 伴音源创建参数
type AudioSourceOptions struct {
	Kind       string // "beep", "sine", "ambient", "silence", "file"
	FilePath   string // 本地音频文件路径 (Kind="file" 时有效)
	SampleRate int    // 采样率，默认 8000
	Codec      string // 编码格式，默认 "G.711A"
}

// NewAudioSource 创建伴音源实例
func NewAudioSource(opts AudioSourceOptions) (AudioSource, error) {
	if opts.SampleRate <= 0 {
		opts.SampleRate = 8000
	}
	kind := strings.ToLower(strings.TrimSpace(opts.Kind))
	switch kind {
	case "beep", "":
		return newBeepAudioSource(opts.SampleRate), nil
	case "sine":
		return newSineAudioSource(1000, opts.SampleRate), nil
	case "ambient", "noise":
		return newAmbientAudioSource(opts.SampleRate), nil
	case "silence":
		return newSilenceAudioSource(), nil
	case "file":
		if opts.FilePath != "" {
			src, err := newFileAudioSource(opts.FilePath, opts.SampleRate)
			if err == nil {
				return src, nil
			}
		}
		// 若文件加载失败，降级为环境底噪
		return newAmbientAudioSource(opts.SampleRate), nil
	default:
		return newBeepAudioSource(opts.SampleRate), nil
	}
}

// ----- Beep 发生器（安防特征双音调蜂鸣：1000Hz/1400Hz 蜂鸣与停顿） -----
type beepAudioSource struct {
	mu         sync.Mutex
	gen        *BeepGenerator
	sampleRate int
}

func newBeepAudioSource(sampleRate int) *beepAudioSource {
	return &beepAudioSource{
		gen:        NewBeepGenerator(),
		sampleRate: sampleRate,
	}
}

func (s *beepAudioSource) NextChunk(samples int) ([]byte, []int16, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if samples <= 0 {
		samples = 320
	}
	pcm := make([]int16, 0, samples)
	for len(pcm) < samples {
		f := s.gen.NextFrame()
		pcm = append(pcm, f...)
	}
	if len(pcm) > samples {
		pcm = pcm[:samples]
	}
	alaw := PCM16ToALaw(pcm)
	return alaw, pcm, nil
}

func (s *beepAudioSource) Close() error {
	return nil
}

// ----- Sine 发生器（1000Hz 标准纯正弦波） -----
type sineAudioSource struct {
	mu         sync.Mutex
	gen        *ToneGenerator
	sampleRate int
}

func newSineAudioSource(freq float64, sampleRate int) *sineAudioSource {
	return &sineAudioSource{
		gen:        NewToneGenerator(freq, sampleRate),
		sampleRate: sampleRate,
	}
}

func (s *sineAudioSource) NextChunk(samples int) ([]byte, []int16, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if samples <= 0 {
		samples = 320
	}
	pcm := s.gen.NextFrame(samples, 0.4)
	alaw := PCM16ToALaw(pcm)
	return alaw, pcm, nil
}

func (s *sineAudioSource) Close() error {
	return nil
}

// ----- Ambient 发生器（安防监控摄像头真实环境底噪与气流微噪） -----
type ambientAudioSource struct {
	mu         sync.Mutex
	rnd        *rand.Rand
	lastSample float64
	sampleRate int
}

func newAmbientAudioSource(sampleRate int) *ambientAudioSource {
	return &ambientAudioSource{
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
		sampleRate: sampleRate,
	}
}

func (s *ambientAudioSource) NextChunk(samples int) ([]byte, []int16, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if samples <= 0 {
		samples = 320
	}
	pcm := make([]int16, samples)
	// 一阶低通滤波模拟自然微风环境声，振幅限制在 1200 以内 (约 -28 dBFS)
	for i := 0; i < samples; i++ {
		white := (s.rnd.Float64()*2.0 - 1.0) * 1100.0
		// 低通滤波平滑化
		s.lastSample = s.lastSample*0.65 + white*0.35
		val := int16(s.lastSample)
		pcm[i] = val
	}
	alaw := PCM16ToALaw(pcm)
	return alaw, pcm, nil
}

func (s *ambientAudioSource) Close() error {
	return nil
}

// ----- Silence 发生器（静音帧） -----
type silenceAudioSource struct{}

func newSilenceAudioSource() *silenceAudioSource {
	return &silenceAudioSource{}
}

func (s *silenceAudioSource) NextChunk(samples int) ([]byte, []int16, error) {
	if samples <= 0 {
		samples = 320
	}
	pcm := make([]int16, samples)
	alaw := PCM16ToALaw(pcm)
	return alaw, pcm, nil
}

func (s *silenceAudioSource) Close() error {
	return nil
}

// ----- File 发生器（本地音频文件循环读取） -----
type fileAudioSource struct {
	mu         sync.Mutex
	data       []byte // 存储 raw ALaw 或 PCM16
	isALaw     bool
	cursor     int
	sampleRate int
}

func newFileAudioSource(path string, sampleRate int) (*fileAudioSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil || len(data) == 0 {
		return nil, fmt.Errorf("empty or unreadable audio file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	isALaw := ext == ".alaw" || ext == ".g711a" || ext == ".pcma" || strings.Contains(strings.ToLower(path), ".alaw.") || strings.Contains(strings.ToLower(path), ".alaw")

	// 若为标准 WAV，跳过 44 字节头
	if ext == ".wav" && len(data) > 44 {
		data = data[44:]
	}

	return &fileAudioSource{
		data:       data,
		isALaw:     isALaw,
		sampleRate: sampleRate,
	}, nil
}

func (s *fileAudioSource) NextChunk(samples int) ([]byte, []int16, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if samples <= 0 {
		samples = 320
	}

	if len(s.data) == 0 {
		pcm := make([]int16, samples)
		return PCM16ToALaw(pcm), pcm, nil
	}

	if s.isALaw {
		// 每次取 samples 字节的 ALaw
		alaw := make([]byte, samples)
		for i := 0; i < samples; i++ {
			alaw[i] = s.data[s.cursor]
			s.cursor = (s.cursor + 1) % len(s.data)
		}
		pcm := ALawToPCM16(alaw)
		return alaw, pcm, nil
	}

	// PCM16 (每次取 samples * 2 字节)
	pcm := make([]int16, samples)
	dataLen := len(s.data) / 2 * 2
	if dataLen == 0 {
		return PCM16ToALaw(pcm), pcm, nil
	}
	for i := 0; i < samples; i++ {
		idx := (s.cursor + i*2) % dataLen
		sample := int16(uint16(s.data[idx]) | uint16(s.data[idx+1])<<8)
		pcm[i] = sample
	}
	s.cursor = (s.cursor + samples*2) % dataLen
	alaw := PCM16ToALaw(pcm)
	return alaw, pcm, nil
}

func (s *fileAudioSource) Close() error {
	return nil
}
