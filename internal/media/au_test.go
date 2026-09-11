package media

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// 构造带 SPS/PPS/双 slice IDR/P 的 Annex-B，验证 AU 切分不把 IDR 切碎。
func TestFileSourceAUSplit(t *testing.T) {
	// 简化 NAL：仅用 header 字节，first_mb 用真实 Exp-Golomb 0 = 0x80
	// slice NAL: nal_ref=3, type=5 → 0x65; first_mb ue(0)=1bit '1' → 0x80...
	sps := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1e}
	pps := []byte{0x00, 0x00, 0x00, 0x01, 0x68, 0xce, 0x38, 0x80}
	// IDR slice0: first_mb=0 → ue bitstream starts with 1 → 0x80
	idr0 := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0x88, 0x84, 0x21, 0xa0}
	// IDR slice1 (same pic, first_mb!=0): ue(1)='010' → 0x40...
	idr1 := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0x40, 0x00, 0x11, 0x22}
	// P slice first_mb=0
	p0 := []byte{0x00, 0x00, 0x00, 0x01, 0x41, 0x9a, 0x00, 0x33}

	var raw bytes.Buffer
	raw.Write(sps)
	raw.Write(pps)
	raw.Write(idr0)
	raw.Write(idr1)
	raw.Write(p0)

	dir := t.TempDir()
	path := filepath.Join(dir, "t.h264")
	if err := os.WriteFile(path, raw.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	src, err := NewFileSource(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.aus) != 2 {
		t.Fatalf("want 2 AUs, got %d: %v", len(src.aus), auSizes(src.aus))
	}
	// AU0 应含 SPS+PPS+两 slice
	f1, err := src.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(f1, sps) || !bytes.Contains(f1, idr0) || !bytes.Contains(f1, idr1) {
		t.Fatalf("AU0 broken, len=%d", len(f1))
	}
	f2, err := src.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(f2, p0) {
		t.Fatalf("AU1 should be P, len=%d", len(f2))
	}
}

func auSizes(aus [][]byte) []int {
	out := make([]int, len(aus))
	for i, a := range aus {
		out[i] = len(a)
	}
	return out
}

func TestPackVideoPESStructure(t *testing.T) {
	au := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1e, 0x00, 0x00, 0x00, 0x01, 0x65, 0x88}
	ps := PackVideoPES(au, 90000)
	if !bytes.HasPrefix(ps, []byte{0x00, 0x00, 0x01, 0xBA}) {
		t.Fatal("missing pack header")
	}
	if !bytes.Contains(ps, []byte{0x00, 0x00, 0x01, 0xBC}) {
		t.Fatal("missing PSM")
	}
	if !bytes.Contains(ps, []byte{0x00, 0x00, 0x01, 0xE0}) {
		t.Fatal("missing PES")
	}
	if !bytes.Contains(ps, []byte{0x00, 0x00, 0x01, 0xBB}) {
		t.Fatal("missing system header")
	}
	// PSM 中应含 H264 stream_type 0x1B
	if !bytes.Contains(ps, []byte{0x1B, 0xE0}) {
		t.Fatal("PSM missing h264 type")
	}
}
