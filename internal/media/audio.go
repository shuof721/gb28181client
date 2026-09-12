package media

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
)

// G.711 A-law / U-law 编解码查表
var (
	pcmToALawTable [65536]byte
	aLawToPCMTable [256]int16
	pcmToULawTable [65536]byte
	uLawToPCMTable [256]int16
	tablesOnce     sync.Once
)

func init() {
	initG711Tables()
}

func initG711Tables() {
	tablesOnce.Do(func() {
		// 初始化 A-law 表
		for i := 0; i < 65536; i++ {
			pcm := int16(uint16(i))
			pcmToALawTable[i] = linearToALaw(pcm)
		}
		for i := 0; i < 256; i++ {
			aLawToPCMTable[i] = aLawToLinear(byte(i))
		}

		// 初始化 U-law 表
		for i := 0; i < 65536; i++ {
			pcm := int16(uint16(i))
			pcmToULawTable[i] = linearToULaw(pcm)
		}
		for i := 0; i < 256; i++ {
			uLawToPCMTable[i] = uLawToLinear(byte(i))
		}
	})
}

// linearToALaw 将 16 位线性 PCM 样本转为 G.711 A-law (PCMA) 字节 (遵循 ITU-T G.711 标准)
func linearToALaw(pcm int16) byte {
	mask := byte(0xD5)
	val := int(pcm) >> 3

	if val < 0 {
		mask = 0x55
		val = -val - 1
	}

	var seg byte
	if val <= 0x1F {
		seg = 0
	} else if val <= 0x3F {
		seg = 1
	} else if val <= 0x7F {
		seg = 2
	} else if val <= 0xFF {
		seg = 3
	} else if val <= 0x1FF {
		seg = 4
	} else if val <= 0x3FF {
		seg = 5
	} else if val <= 0x7FF {
		seg = 6
	} else if val <= 0xFFF {
		seg = 7
	} else {
		return 0x7F ^ mask
	}

	aval := seg << 4
	if seg < 2 {
		aval |= byte((val >> 1) & 0x0F)
	} else {
		aval |= byte((val >> seg) & 0x0F)
	}
	return aval ^ mask
}

// aLawToLinear 将 G.711 A-law 字节转为 16 位线性 PCM (遵循 ITU-T G.711 标准)
func aLawToLinear(alaw byte) int16 {
	alaw ^= 0x55
	t := (int(alaw&0x0F) << 4)
	seg := (alaw & 0x70) >> 4

	switch seg {
	case 0:
		t += 8
	case 1:
		t += 0x108
	default:
		t += 0x108
		t <<= (seg - 1)
	}

	if (alaw & 0x80) != 0 {
		return int16(t)
	}
	return -int16(t)
}

// linearToULaw 将 16 位线性 PCM 转为 G.711 U-law (PCMU) 字节
func linearToULaw(pcm int16) byte {
	const bias = 0x84
	const clip = 32635

	sign := byte(0)
	sample := int(pcm)
	if sample < 0 {
		sample = -sample
		sign = 0x80
	}
	if sample > clip {
		sample = clip
	}
	sample += bias

	exponent := 7
	for expMask := 0x4000; (sample & expMask) == 0 && exponent > 0; expMask >>= 1 {
		exponent--
	}
	mantissa := (sample >> (exponent + 3)) & 0x0F
	uval := sign | byte(exponent<<4) | byte(mantissa)
	return ^uval
}

// uLawToLinear 将 G.711 U-law 字节转为 16 位线性 PCM
func uLawToLinear(ulaw byte) int16 {
	ulaw = ^ulaw
	sign := ulaw & 0x80
	exponent := int((ulaw >> 4) & 0x07)
	mantissa := int(ulaw & 0x0F)
	sample := ((mantissa << 3) + 0x84) << exponent
	sample -= 0x84
	if sign != 0 {
		return -int16(sample)
	}
	return int16(sample)
}

// PCM16ToALaw 批量转换 PCM16 -> G.711A (PCMA)
func PCM16ToALaw(pcm []int16) []byte {
	out := make([]byte, len(pcm))
	for i, s := range pcm {
		out[i] = pcmToALawTable[uint16(s)]
	}
	return out
}

// ALawToPCM16 批量转换 G.711A (PCMA) -> PCM16
func ALawToPCM16(alaw []byte) []int16 {
	out := make([]int16, len(alaw))
	for i, b := range alaw {
		out[i] = aLawToPCMTable[b]
	}
	return out
}

// PCM16ToULaw 批量转换 PCM16 -> G.711U (PCMU)
func PCM16ToULaw(pcm []int16) []byte {
	out := make([]byte, len(pcm))
	for i, s := range pcm {
		out[i] = pcmToULawTable[uint16(s)]
	}
	return out
}

// ULawToPCM16 批量转换 G.711U (PCMU) -> PCM16
func ULawToPCM16(ulaw []byte) []int16 {
	out := make([]int16, len(ulaw))
	for i, b := range ulaw {
		out[i] = uLawToPCMTable[b]
	}
	return out
}

// CalculateRMSLevel 计算 PCM16 音频帧的人耳感知电平（0.0 ~ 100.0），供前端 VU 表绘制
// 采用声学工程标准的对数分贝 dBFS 标度映射（-52 dBFS ~ -2 dBFS -> 0% ~ 100%），
// 并结合底噪门限截断，使对讲普通人声、喊话及提示音的跳动幅度鲜明灵敏。
func CalculateRMSLevel(pcm []int16) float64 {
	if len(pcm) == 0 {
		return 0
	}
	var sumSquares float64
	for _, s := range pcm {
		f := float64(s)
		sumSquares += f * f
	}
	rms := math.Sqrt(sumSquares / float64(len(pcm)))
	if rms < 12.0 {
		return 0
	}

	// 转换为相对于满量程 (32768) 的分贝值 dBFS
	db := 20.0 * math.Log10(rms/32768.0)

	// 人声对讲动态范围：-52 dBFS (极轻微背景声) 到 -2 dBFS (近场满载)
	const minDB = -52.0
	const maxDB = -2.0
	if db <= minDB {
		return 0
	}
	if db >= maxDB {
		return 100
	}

	level := ((db - minDB) / (maxDB - minDB)) * 100.0
	return math.Round(level*10) / 10
}

// RTPPacketizeAudio 将音频原始字节（如 160 字节的 PCMA 帧）封装为单个 RTP 数据包
// pt: Payload Type (如 8 为 PCMA, 0 为 PCMU)
// ssrc: 会话 SSRC
// seq: 序列号计数指针
// ts: 时间戳
func RTPPacketizeAudio(payload []byte, pt uint8, ssrc uint32, seq *uint16, ts uint32) []byte {
	pkt := make([]byte, 12+len(payload))
	// V=2, P=0, X=0, CC=0
	pkt[0] = 0x80
	// M=0, PT
	pkt[1] = pt & 0x7F
	binary.BigEndian.PutUint16(pkt[2:4], *seq)
	*seq++
	binary.BigEndian.PutUint32(pkt[4:8], ts)
	binary.BigEndian.PutUint32(pkt[8:12], ssrc)
	copy(pkt[12:], payload)
	return pkt
}

// RTPDepacketizeAudio 解析单个音频 RTP 数据包，返回 payload、seq、timestamp
func RTPDepacketizeAudio(pkt []byte) (payload []byte, pt uint8, seq uint16, ts uint32, err error) {
	if len(pkt) < 12 {
		return nil, 0, 0, 0, errors.New("rtp packet too short")
	}
	v := pkt[0] >> 6
	if v != 2 {
		return nil, 0, 0, 0, errors.New("unsupported rtp version")
	}
	cc := int(pkt[0] & 0x0F)
	hasExt := (pkt[0] & 0x10) != 0
	pt = pkt[1] & 0x7F
	seq = binary.BigEndian.Uint16(pkt[2:4])
	ts = binary.BigEndian.Uint32(pkt[4:8])

	offset := 12 + cc*4
	if len(pkt) < offset {
		return nil, 0, 0, 0, errors.New("invalid csrc length")
	}
	if hasExt {
		if len(pkt) < offset+4 {
			return nil, 0, 0, 0, errors.New("invalid extension header")
		}
		extLen := int(binary.BigEndian.Uint16(pkt[offset+2:offset+4])) * 4
		offset += 4 + extLen
		if len(pkt) < offset {
			return nil, 0, 0, 0, errors.New("extension header out of range")
		}
	}
	return pkt[offset:], pt, seq, ts, nil
}

// ToneGenerator 生成标准单音频或蜂鸣序列 PCM16 帧 (8000Hz)
type ToneGenerator struct {
	sampleRate int
	phase      float64
	step       float64
}

// NewToneGenerator 创建正弦波音频发生器
func NewToneGenerator(freq float64, sampleRate int) *ToneGenerator {
	if sampleRate <= 0 {
		sampleRate = 8000
	}
	return &ToneGenerator{
		sampleRate: sampleRate,
		step:       2.0 * math.Pi * freq / float64(sampleRate),
	}
}

// NextFrame 生成一段 20ms（160 samples）的 PCM16 数据
func (g *ToneGenerator) NextFrame(numSamples int, amplitude float64) []int16 {
	if numSamples <= 0 {
		numSamples = 160 // 8000Hz * 0.02s
	}
	if amplitude <= 0 {
		amplitude = 0.5
	}
	out := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		sample := math.Sin(g.phase) * amplitude * 32767.0
		out[i] = int16(sample)
		g.phase += g.step
		if g.phase >= 2.0*math.Pi {
			g.phase -= 2.0 * math.Pi
		}
	}
	return out
}

// BeepGenerator 生成具有安防对讲连通特征的间歇蜂鸣信号（Beep-Beep-Pause）
type BeepGenerator struct {
	g1         *ToneGenerator // 1000Hz
	g2         *ToneGenerator // 1400Hz
	frameIndex int
}

// NewBeepGenerator 创建对讲提示音发生器
func NewBeepGenerator() *BeepGenerator {
	return &BeepGenerator{
		g1: NewToneGenerator(1000, 8000),
		g2: NewToneGenerator(1400, 8000),
	}
}

// NextFrame 生成 20ms (160 samples) 循环对讲提示音
// 周期为 1.5 秒 (75 个 20ms 帧):
// 帧 0-8: 1000Hz (160ms)
// 帧 9-12: 静音 (80ms)
// 帧 13-21: 1400Hz (160ms)
// 帧 22-74: 静音 (1060ms)
func (b *BeepGenerator) NextFrame() []int16 {
	idx := b.frameIndex % 75
	b.frameIndex++

	if idx >= 0 && idx < 9 {
		return b.g1.NextFrame(160, 0.4)
	} else if idx >= 13 && idx < 22 {
		return b.g2.NextFrame(160, 0.4)
	}
	return make([]int16, 160) // 静音帧
}
