package media

import (
	"testing"
)

func TestG711ALawCodec(t *testing.T) {
	// 测试正弦波数据转 A-law 后再转回 PCM，验证保真度与量化误差在合理范围 (< 5%)
	gen := NewToneGenerator(440, 8000)
	orig := gen.NextFrame(160, 0.8)

	alaw := PCM16ToALaw(orig)
	if len(alaw) != 160 {
		t.Fatalf("expected 160 bytes, got %d", len(alaw))
	}

	decoded := ALawToPCM16(alaw)
	if len(decoded) != 160 {
		t.Fatalf("expected 160 samples, got %d", len(decoded))
	}

	for i := 0; i < len(orig); i++ {
		diff := int(orig[i]) - int(decoded[i])
		if diff < 0 {
			diff = -diff
		}
		// A-law 为非线性对数量化，较大幅度下量化步长稍大，diff 应当在 500 以内
		if diff > 500 {
			t.Errorf("sample %d: orig %d, decoded %d, diff %d too large", i, orig[i], decoded[i], diff)
		}
	}
}

func TestG711ULawCodec(t *testing.T) {
	gen := NewToneGenerator(1000, 8000)
	orig := gen.NextFrame(160, 0.6)

	ulaw := PCM16ToULaw(orig)
	if len(ulaw) != 160 {
		t.Fatalf("expected 160 bytes, got %d", len(ulaw))
	}

	decoded := ULawToPCM16(ulaw)
	if len(decoded) != 160 {
		t.Fatalf("expected 160 samples, got %d", len(decoded))
	}
}

func TestAudioRTPPacketizeDepacketize(t *testing.T) {
	payload := make([]byte, 160)
	for i := range payload {
		payload[i] = byte(i)
	}

	seq := uint16(100)
	ts := uint32(96000)
	ssrc := uint32(0x12345678)

	pkt := RTPPacketizeAudio(payload, 8, ssrc, &seq, ts)
	if len(pkt) != 12+160 {
		t.Fatalf("expected packet length %d, got %d", 12+160, len(pkt))
	}
	if seq != 101 {
		t.Fatalf("expected seq increment to 101, got %d", seq)
	}

	decPayload, pt, decSeq, decTs, err := RTPDepacketizeAudio(pkt)
	if err != nil {
		t.Fatalf("depacketize error: %v", err)
	}
	if pt != 8 {
		t.Errorf("expected pt 8, got %d", pt)
	}
	if decSeq != 100 {
		t.Errorf("expected seq 100, got %d", decSeq)
	}
	if decTs != ts {
		t.Errorf("expected ts %d, got %d", ts, decTs)
	}
	for i := range payload {
		if decPayload[i] != payload[i] {
			t.Fatalf("payload byte %d mismatch", i)
		}
	}
}

func TestBeepGenerator(t *testing.T) {
	bg := NewBeepGenerator()
	// 生成 100 帧（覆盖 2 秒周期），验证无 panic 且数据长度正常
	for i := 0; i < 100; i++ {
		frame := bg.NextFrame()
		if len(frame) != 160 {
			t.Fatalf("expected frame len 160, got %d", len(frame))
		}
		vol := CalculateRMSLevel(frame)
		if vol < 0 || vol > 100 {
			t.Fatalf("invalid volume level: %f", vol)
		}
	}
}
