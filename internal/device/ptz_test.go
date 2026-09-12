package device

import (
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/gb28181"
)

func TestChannelPTZ(t *testing.T) {
	var savedCh string
	var savedCfg *config.PTZConfig

	onChange := func(ch string, cfg *config.PTZConfig) {
		savedCh = ch
		savedCfg = cfg
	}

	initCfg := &config.PTZConfig{
		Pan:  0.0,
		Tilt: 0.0,
		Zoom: 1.0,
		Presets: []config.PTZPreset{
			{ID: 1, Name: "预置位1", Pan: 100.0, Tilt: 20.0, Zoom: 2.0},
		},
	}

	ptz := NewChannelPTZ("test-ch", initCfg, onChange)

	// 1. Initial status
	st := ptz.Status()
	if st.Pan != 0.0 || st.Tilt != 0.0 || st.Zoom != 1.0 {
		t.Fatalf("Unexpected initial coords: %+v", st)
	}
	if len(st.Presets) != 1 {
		t.Fatalf("Expected 1 preset, got %d", len(st.Presets))
	}

	// 2. Start moving right
	moveCmd := &gb28181.PTZCommand{
		Action:   gb28181.PTZActionMove,
		PanDir:   1,
		PanSpeed: 128,
	}
	_, err := ptz.ExecuteCommand(moveCmd)
	if err != nil {
		t.Fatalf("Execute move failed: %v", err)
	}

	// Wait 100ms for simulated movement
	time.Sleep(100 * time.Millisecond)

	st = ptz.Status()
	if !st.IsMoving || st.Pan <= 0.0 {
		t.Fatalf("Expected moving right, got pan=%f, isMoving=%v", st.Pan, st.IsMoving)
	}

	// 3. Stop
	stopCmd := &gb28181.PTZCommand{Action: gb28181.PTZActionStop}
	_, err = ptz.ExecuteCommand(stopCmd)
	if err != nil {
		t.Fatalf("Execute stop failed: %v", err)
	}
	st = ptz.Status()
	if st.IsMoving {
		t.Fatalf("Expected stopped, got isMoving=%v", st.IsMoving)
	}
	savedPan := st.Pan

	// 4. Set current position as Preset #2
	_, err = ptz.SetPreset(2, "测试位置2", 0, 0, 0, true)
	if err != nil {
		t.Fatalf("SetPreset failed: %v", err)
	}
	time.Sleep(20 * time.Millisecond) // wait for async callback
	if savedCh != "test-ch" || savedCfg == nil || len(savedCfg.Presets) != 2 {
		t.Fatalf("Expected callback with 2 presets, got %v, %+v", savedCh, savedCfg)
	}

	// 5. Call Preset #1
	_, err = ptz.CallPreset(1)
	if err != nil {
		t.Fatalf("CallPreset failed: %v", err)
	}
	st = ptz.Status()
	if !st.IsMoving {
		t.Fatalf("Expected transitioning to preset, got isMoving=%v", st.IsMoving)
	}

	// 6. Delete Preset #2
	_, err = ptz.DeletePreset(2)
	if err != nil {
		t.Fatalf("DeletePreset failed: %v", err)
	}
	st = ptz.Status()
	if len(st.Presets) != 1 {
		t.Fatalf("Expected 1 preset after delete, got %d", len(st.Presets))
	}
	// 7. SetPose and Reset
	pSt := ptz.SetPose(45.5, -10.2, 3.5)
	if pSt.Pan != 45.5 || pSt.Tilt != -10.2 || pSt.Zoom != 3.5 {
		t.Fatalf("Expected SetPose to set coords, got pan=%f tilt=%f zoom=%f", pSt.Pan, pSt.Tilt, pSt.Zoom)
	}

	rSt := ptz.ManualControl("reset", 0, 0, 0)
	if rSt.Pan != 0.0 || rSt.Tilt != 0.0 || rSt.Zoom != 1.0 {
		t.Fatalf("Expected reset to set coords to 0/0/1, got pan=%f tilt=%f zoom=%f", rSt.Pan, rSt.Tilt, rSt.Zoom)
	}

	_ = savedPan
}

func TestChannelPTZMediaBinding(t *testing.T) {
	cfg := config.Default()
	cfg.Media.LocalIP = "127.0.0.1"
	cfg.Device.Channels = []config.ChannelConfig{
		{ID: "34020000001320000001", Name: "IPC-PTZ", PTZType: 1},
	}

	dev := New(cfg)
	err := dev.BindChannelVideo(BindChannelRequest{
		ChannelID: "34020000001320000001",
		Source:    "ptz",
	})
	if err != nil {
		t.Fatalf("BindChannelVideo ptz failed: %v", err)
	}

	// 模拟流媒体会话提取源
	v := dev.cfg.Media.OptionsFor("34020000001320000001")
	if v.Kind != "ptz" {
		t.Fatalf("Expected kind=ptz, got %s", v.Kind)
	}

	// 验证 PTZ 状态机存在
	ptz := dev.GetChannelPTZ("34020000001320000001")
	if ptz == nil {
		t.Fatalf("GetChannelPTZ returned nil")
	}
	ptz.SetPose(90.0, -15.0, 2.0)
	st := ptz.Status()
	if st.Pan != 90.0 || st.Tilt != -15.0 || st.Zoom != 2.0 {
		t.Fatalf("unexpected PTZ status: %+v", st)
	}
}
