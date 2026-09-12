package gb28181

import (
	"strings"
	"testing"
)

func TestParseMANSRTSP(t *testing.T) {
	playReq := []byte("PLAY RTSP/1.0\r\nCSeq: 1\r\nScale: 2.000000\r\nRange: npt=120.500000-\r\n\r\n")
	req, err := ParseMANSRTSP(playReq)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if req.Method != "PLAY" {
		t.Errorf("expected PLAY, got %s", req.Method)
	}
	if req.CSeq != "1" {
		t.Errorf("expected CSeq 1, got %s", req.CSeq)
	}
	if !req.HasScale || req.Scale != 2.0 {
		t.Errorf("expected Scale 2.0, got %v", req.Scale)
	}
	if !req.HasRange || req.RangeNPT != 120.5 {
		t.Errorf("expected Range 120.5, got %v", req.RangeNPT)
	}

	resp := BuildMANSRTSPResponse(req.CSeq, req.Scale, req.RangeNPT, false)
	if !strings.Contains(string(resp), "RTSP/1.0 200 OK") || !strings.Contains(string(resp), "Scale: 2.000000") {
		t.Errorf("bad response: %s", string(resp))
	}

	pauseReq := []byte("PAUSE RTSP/1.0\r\nCSeq: 2\r\nPauseTime: 120\r\n\r\n")
	req2, err := ParseMANSRTSP(pauseReq)
	if err != nil {
		t.Fatalf("parse pause failed: %v", err)
	}
	if req2.Method != "PAUSE" || req2.CSeq != "2" {
		t.Errorf("bad pause req: %+v", req2)
	}
	resp2 := BuildMANSRTSPResponse(req2.CSeq, 0, 0, true)
	if !strings.Contains(string(resp2), "RTSP/1.0 200 OK") || !strings.Contains(string(resp2), "CSeq: 2") {
		t.Errorf("bad pause response: %s", string(resp2))
	}
}
