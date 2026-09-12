package media

import (
	"testing"
)

func TestPTZSource(t *testing.T) {
	curPan := 0.0
	curTilt := 0.0
	curZoom := 1.0

	opts := PTZSourceOptions{
		ChannelID: "34020000001320000001",
		Width:     480,
		Height:    272,
		FPS:       25,
		GetPTZ: func() PTZInfo {
			return PTZInfo{
				Pan:            curPan,
				Tilt:           curTilt,
				Zoom:           curZoom,
				IsMoving:       true,
				ActivePresetID: 1,
				ActivePreset:   "大门全景",
			}
		},
	}

	src, err := NewPTZSource(opts)
	if err != nil {
		t.Fatalf("NewPTZSource failed: %v", err)
	}
	defer src.Close()

	// 1. 获取第一帧 (Pan 0°)
	f1, err := src.Next()
	if err != nil {
		t.Fatalf("src.Next() frame 1 failed: %v", err)
	}
	if len(f1) == 0 {
		t.Fatalf("frame 1 is empty")
	}

	// 验证包含 IDR、SPS、PPS
	nals1 := splitNals(f1)
	hasSPS := false
	hasPPS := false
	hasIDR := false
	for _, n := range nals1 {
		if n.typ == 7 {
			hasSPS = true
		}
		if n.typ == 8 {
			hasPPS = true
		}
		if n.typ == 5 {
			hasIDR = true
		}
	}
	if !hasSPS || !hasPPS || !hasIDR {
		t.Fatalf("frame missing essential NALs: sps=%v, pps=%v, idr=%v", hasSPS, hasPPS, hasIDR)
	}

	// 2. 模拟旋转并获取第二帧 (Pan 90° 东环路, Tilt -10°, Zoom 2.0x)
	curPan = 90.0
	curTilt = -10.0
	curZoom = 2.0
	f2, err := src.Next()
	if err != nil {
		t.Fatalf("src.Next() frame 2 failed: %v", err)
	}
	if len(f2) == 0 {
		t.Fatalf("frame 2 is empty")
	}

	// 3. 测试 Seek
	actual, err := src.Seek(10.5, 25)
	if err != nil {
		t.Fatalf("Seek failed: %v", err)
	}
	if actual != 10.5 {
		t.Fatalf("unexpected seek actual: %f", actual)
	}
}
