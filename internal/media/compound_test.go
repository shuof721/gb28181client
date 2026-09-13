package media

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// TestCompoundPESStructure 测试音视频复合 PS 包的头部标志、System Header、PSM 映射及双 PES 数据
func TestCompoundPESStructure(t *testing.T) {
	fakeVideo := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0x88, 0x84, 0x00, 0x10, 0x20}
	fakeAudio := make([]byte, 320)
	for i := range fakeAudio {
		fakeAudio[i] = byte(i % 256)
	}
	ts := uint64(90000)

	// 1. 生成音视频复合流
	compoundPS := PackCompoundPES(fakeVideo, fakeAudio, ts)

	// 验证 Pack Header: 00 00 01 BA
	if len(compoundPS) < 14 || !bytes.Equal(compoundPS[:4], []byte{0x00, 0x00, 0x01, 0xBA}) {
		t.Fatalf("invalid pack header prefix: %x", compoundPS[:4])
	}

	// 验证 System Header: 00 00 01 BB
	sysIdx := bytes.Index(compoundPS, []byte{0x00, 0x00, 0x01, 0xBB})
	if sysIdx == -1 {
		t.Fatalf("system header not found in compound PS")
	}
	// audio_bound 应为 1 -> byte 3 (offset + 9) 应包含 audio_bound
	// b3 = (1 << 2) | 0x01 = 0x05
	if compoundPS[sysIdx+9] != 0x05 {
		t.Fatalf("expected audio_bound=1 (0x05), got 0x%02X", compoundPS[sysIdx+9])
	}

	// 验证 PSM Header: 00 00 01 BC
	psmIdx := bytes.Index(compoundPS, []byte{0x00, 0x00, 0x01, 0xBC})
	if psmIdx == -1 {
		t.Fatalf("PSM not found in compound PS")
	}
	// PSM elementary_stream_map_length 应为 8 (2条流: 0x1B/0xE0 与 0x90/0xC0)
	mapLen := binary.BigEndian.Uint16(compoundPS[psmIdx+10 : psmIdx+12])
	if mapLen != 8 {
		t.Fatalf("expected PSM mapLen=8, got %d", mapLen)
	}
	// 验证视频轨与音频轨映射
	if !bytes.Contains(compoundPS[psmIdx:], []byte{0x1B, 0xE0, 0x00, 0x00}) {
		t.Errorf("PSM missing video track 0x1B / 0xE0")
	}
	if !bytes.Contains(compoundPS[psmIdx:], []byte{0x90, 0xC0, 0x00, 0x00}) {
		t.Errorf("PSM missing audio track 0x90 / 0xC0")
	}

	// 验证 Audio PES: 00 00 01 C0
	aIdx := bytes.Index(compoundPS, []byte{0x00, 0x00, 0x01, 0xC0})
	if aIdx == -1 {
		t.Fatalf("Audio PES (0xC0) not found")
	}
	if !bytes.Contains(compoundPS[aIdx:], fakeAudio) {
		t.Errorf("Audio PES does not contain fake audio payload")
	}

	// 验证 Video PES: 00 00 01 E0，且必须在 Audio PES 之后（避免大帧截断）
	vIdx := bytes.Index(compoundPS, []byte{0x00, 0x00, 0x01, 0xE0})
	if vIdx == -1 {
		t.Fatalf("Video PES (0xE0) not found")
	}
	if aIdx >= vIdx {
		t.Errorf("Audio PES should precede Video PES: aIdx=%d, vIdx=%d", aIdx, vIdx)
	}
	if !bytes.Contains(compoundPS[vIdx:], fakeVideo) {
		t.Errorf("Video PES does not contain fake video payload")
	}

	// 2. 验证纯视频 PS 包（音频为空）
	videoOnlyPS := PackVideoPES(fakeVideo, ts)
	if bytes.Contains(videoOnlyPS, []byte{0x00, 0x00, 0x01, 0xC0}) {
		t.Errorf("pure video PS should not contain audio PES (0xC0)")
	}
	vSysIdx := bytes.Index(videoOnlyPS, []byte{0x00, 0x00, 0x01, 0xBB})
	if videoOnlyPS[vSysIdx+9] != 0x01 {
		t.Errorf("pure video PS expected audio_bound=0 (0x01), got 0x%02X", videoOnlyPS[vSysIdx+9])
	}
}

// TestAudioSources 测试各伴音源（Beep, Sine, Ambient, Silence）的正常产出
func TestAudioSources(t *testing.T) {
	sources := []string{"beep", "sine", "ambient", "silence"}
	for _, kind := range sources {
		src, err := NewAudioSource(AudioSourceOptions{
			Kind:       kind,
			SampleRate: 8000,
			Codec:      "G.711A",
		})
		if err != nil {
			t.Fatalf("NewAudioSource(%s) failed: %v", kind, err)
		}
		alaw, pcm, err := src.NextChunk(320)
		if err != nil {
			t.Fatalf("%s NextChunk failed: %v", kind, err)
		}
		if len(alaw) != 320 || len(pcm) != 320 {
			t.Fatalf("%s expected 320 samples, got alaw=%d pcm=%d", kind, len(alaw), len(pcm))
		}
		lvl := CalculateRMSLevel(pcm)
		if kind == "silence" {
			if lvl != 0 {
				t.Errorf("silence expected lvl=0, got %.2f", lvl)
			}
			// G.711A 静音为 0xD5
			if alaw[0] != 0xD5 {
				t.Errorf("silence G.711A byte expected 0xD5, got 0x%02X", alaw[0])
			}
		} else if kind == "sine" {
			if lvl < 50.0 {
				t.Errorf("sine wave expected active RMS lvl > 50, got %.2f", lvl)
			}
		}
		_ = src.Close()
	}
}

// TestSessionWithCompoundAudio 测试 SessionManager 启用音频源后的推流与 VU 电平实时监控
func TestSessionWithCompoundAudio(t *testing.T) {
	recvUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen udp failed: %v", err)
	}
	defer recvUDP.Close()
	recvPort := recvUDP.LocalAddr().(*net.UDPAddr).Port

	mgr := NewSessionManager("127.0.0.1", 25, 1400, func(channelID string) (H264Source, error) {
		return NewSyntheticSource(320, 240, 25)
	})
	mgr.SetAudioSourceFactory(func(channelID string) (AudioSource, string, string, error) {
		src, err := NewAudioSource(AudioSourceOptions{Kind: "sine", SampleRate: 8000})
		return src, "G.711A", "sine", err
	})

	sdp := &SDPInfo{
		IP:          "127.0.0.1",
		VideoPort:   recvPort,
		SSRC:        "0100000008",
		SessionName: "Play",
	}

	callID := "compound-test-callid"
	ans, err := mgr.StartLive("34020000001320000001", callID, sdp)
	if err != nil {
		t.Fatalf("StartLive failed: %v", err)
	}
	if ans == "" {
		t.Fatalf("empty answer SDP")
	}

	// 等待接收至少 1 个包含音频 PES 的复合包
	_ = recvUDP.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 2048)
	foundAudioPES := false

	for start := time.Now(); time.Since(start) < 2*time.Second; {
		n, _, err := recvUDP.ReadFrom(buf)
		if err != nil {
			break
		}
		if n > 12 {
			payload := buf[12:n]
			if bytes.Contains(payload, []byte{0x00, 0x00, 0x01, 0xC0}) {
				foundAudioPES = true
				break
			}
		}
	}

	if !foundAudioPES {
		t.Errorf("did not receive Audio PES (0x000001C0) in RTP stream")
	}

	// 检查会话状态回显与 VU 电平
	sessions := mgr.ListSessions()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	sess := sessions[0]
	if !sess.AudioEnabled {
		t.Errorf("expected AudioEnabled=true")
	}
	if sess.AudioCodec != "G.711A" {
		t.Errorf("expected AudioCodec=G.711A, got %s", sess.AudioCodec)
	}
	if sess.AudioLevel <= 0 {
		t.Errorf("expected positive audio VU level from sine wave, got %.2f", sess.AudioLevel)
	}

	mgr.StopAll()
}
