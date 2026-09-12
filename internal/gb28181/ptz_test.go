package gb28181

import (
	"testing"
)

func TestParsePTZCmd(t *testing.T) {
	tests := []struct {
		name        string
		hex         string
		wantAction  PTZAction
		wantPanDir  int
		wantTiltDir int
		wantZoomDir int
		wantZoomSpd int
		wantFocus   int
		wantIris    int
		wantPreset  int
	}{
		{
			name:       "Standard GB WVP Stop command",
			hex:        "A50F0100000000B5",
			wantAction: PTZActionStop,
		},
		{
			name:        "Standard GB WVP Move Right (0x01)",
			hex:         "A50F01011F0000D5",
			wantAction:  PTZActionMove,
			wantPanDir:  1,
			wantTiltDir: 0,
		},
		{
			name:        "Standard GB WVP Move Left (0x02)",
			hex:         "A50F01021F0000D6",
			wantAction:  PTZActionMove,
			wantPanDir:  -1,
			wantTiltDir: 0,
		},
		{
			name:        "Standard GB WVP Move Up (0x08)",
			hex:         "A50F0108001F00D7",
			wantAction:  PTZActionMove,
			wantPanDir:  0,
			wantTiltDir: 1,
		},
		{
			name:        "Standard GB WVP Move Down (0x04)",
			hex:         "A50F0104001F00D3",
			wantAction:  PTZActionMove,
			wantPanDir:  0,
			wantTiltDir: -1,
		},
		{
			name:        "Standard GB WVP Move Up-Right (0x09)",
			hex:         "A50F01091F1F00F7",
			wantAction:  PTZActionMove,
			wantPanDir:  1,
			wantTiltDir: 1,
		},
		{
			name:        "Standard GB WVP Move Down-Left (0x06)",
			hex:         "A50F01061F1F00F4",
			wantAction:  PTZActionMove,
			wantPanDir:  -1,
			wantTiltDir: -1,
		},
		{
			name:        "Standard GB Zoom In 变倍放大 (0x10)",
			hex:         "A50F0110000020EB",
			wantAction:  PTZActionMove,
			wantZoomDir: 1,
			wantZoomSpd: 2,
		},
		{
			name:        "Standard GB Zoom Out 变倍缩小 (0x20)",
			hex:         "A50F0120000020FB",
			wantAction:  PTZActionMove,
			wantZoomDir: -1,
			wantZoomSpd: 2,
		},
		{
			name:        "Standard GB Combined Move Up-Right + Zoom In (0x19)",
			hex:         "A50F01191F1F202C",
			wantAction:  PTZActionMove,
			wantPanDir:  1,
			wantTiltDir: 1,
			wantZoomDir: 1,
			wantZoomSpd: 2,
		},
		{
			name:        "Standard GB Focus Near 聚焦近 (0x42)",
			hex:         "A50F01421F000016",
			wantAction:  PTZActionMove,
			wantFocus:   -1,
		},
		{
			name:        "Standard GB Iris Open 光圈开 (0x48)",
			hex:         "A50F0148001F0017",
			wantAction:  PTZActionMove,
			wantIris:    1,
		},
		{
			name:       "Standard GB Preset Set #5 (0x81)",
			hex:        "A50F01810000053A",
			wantAction: PTZActionPresetSet,
			wantPreset: 5,
		},
		{
			name:       "Standard GB Preset Call #3 (0x82)",
			hex:        "A50F018200000339",
			wantAction: PTZActionPresetCall,
			wantPreset: 3,
		},
		{
			name:       "Standard GB Preset Delete #2 (0x83)",
			hex:        "A50F018300000239",
			wantAction: PTZActionPresetDelete,
			wantPreset: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParsePTZCmd(tt.hex)
			if err != nil {
				t.Fatalf("ParsePTZCmd error: %v", err)
			}
			if cmd.Action != tt.wantAction {
				t.Errorf("Action = %v, want %v", cmd.Action, tt.wantAction)
			}
			if tt.wantAction == PTZActionMove {
				if cmd.PanDir != tt.wantPanDir {
					t.Errorf("PanDir = %v, want %v", cmd.PanDir, tt.wantPanDir)
				}
				if cmd.TiltDir != tt.wantTiltDir {
					t.Errorf("TiltDir = %v, want %v", cmd.TiltDir, tt.wantTiltDir)
				}
				if cmd.ZoomDir != tt.wantZoomDir {
					t.Errorf("ZoomDir = %v, want %v", cmd.ZoomDir, tt.wantZoomDir)
				}
				if tt.wantZoomSpd > 0 && cmd.ZoomSpeed != tt.wantZoomSpd {
					t.Errorf("ZoomSpeed = %v, want %v", cmd.ZoomSpeed, tt.wantZoomSpd)
				}
				if tt.wantFocus != 0 && cmd.FocusDir != tt.wantFocus {
					t.Errorf("FocusDir = %v, want %v", cmd.FocusDir, tt.wantFocus)
				}
				if tt.wantIris != 0 && cmd.IrisDir != tt.wantIris {
					t.Errorf("IrisDir = %v, want %v", cmd.IrisDir, tt.wantIris)
				}
			}
			if tt.wantPreset > 0 && cmd.PresetID != tt.wantPreset {
				t.Errorf("PresetID = %v, want %v", cmd.PresetID, tt.wantPreset)
			}
		})
	}
}

func TestEncodePTZCmd(t *testing.T) {
	// 1. 移动与变倍编码测试
	cmd := &PTZCommand{
		Action:    PTZActionMove,
		PanDir:    1,
		TiltDir:   1,
		ZoomDir:   1,
		PanSpeed:  31,
		TiltSpeed: 31,
		ZoomSpeed: 2,
	}
	hexStr := EncodePTZCmd(cmd)
	if len(hexStr) != 16 {
		t.Fatalf("EncodePTZCmd length = %d, want 16", len(hexStr))
	}

	parsed, err := ParsePTZCmd(hexStr)
	if err != nil {
		t.Fatalf("Parse encoded hex error: %v", err)
	}
	if parsed.Action != PTZActionMove || parsed.PanDir != 1 || parsed.TiltDir != 1 || parsed.ZoomDir != 1 || parsed.ZoomSpeed != 2 {
		t.Errorf("Decoded mismatch: %+v", parsed)
	}

	// 2. 变倍缩小编码测试
	zoomOutCmd := &PTZCommand{
		Action:    PTZActionMove,
		ZoomDir:   -1,
		ZoomSpeed: 2,
	}
	zHex := EncodePTZCmd(zoomOutCmd)
	zParsed, err := ParsePTZCmd(zHex)
	if err != nil {
		t.Fatalf("Parse zoomout hex error: %v", err)
	}
	if zParsed.Action != PTZActionMove || zParsed.ZoomDir != -1 || zParsed.ZoomSpeed != 2 {
		t.Errorf("Decoded zoomout mismatch: %+v", zParsed)
	}

	// 3. 预置位调用编码测试
	presetCmd := &PTZCommand{
		Action:   PTZActionPresetCall,
		PresetID: 8,
	}
	pHex := EncodePTZCmd(presetCmd)
	pParsed, err := ParsePTZCmd(pHex)
	if err != nil {
		t.Fatalf("Parse preset hex error: %v", err)
	}
	if pParsed.Action != PTZActionPresetCall || pParsed.PresetID != 8 {
		t.Errorf("Decoded preset mismatch: %+v", pParsed)
	}
}
