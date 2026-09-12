package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/device"
	"github.com/local/gb28181-device/internal/storage"
)

func TestServerDeviceRecordsAPI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gb-ui-server-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.New(tempDir)
	profile := config.DeviceProfile{
		Enabled: false,
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: 5060,
			LocalIP:    "127.0.0.1",
			LocalPort:  5070,
			Transport:  "udp",
			Password:   "12345678",
		},
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "NVR-1",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "IPC-1"},
			},
		},
		Record: config.RecordConfig{
			Enabled:      true,
			Mode:         "continuous",
			SliceMinutes: 60,
			RetainDays:   7,
			RecordType:   "time",
			FileSizeMB:   100,
		},
	}
	if err := store.Save(&profile); err != nil {
		t.Fatalf("Save profile error: %v", err)
	}

	mgr := device.NewManager(store)
	if err := mgr.Init(); err != nil {
		t.Fatalf("mgr init: %v", err)
	}
	srv := New(mgr)

	// 1. Test GET /api/devices/34020000001180000001/records
	today := time.Now().Format("2006-01-02")
	req := httptest.NewRequest(http.MethodGet, "/api/devices/34020000001180000001/records?start="+today+"T00:00:00&end="+today+"T23:59:59", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Records []any               `json:"records"`
		Total   int                 `json:"total"`
		Config  config.RecordConfig `json:"config"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if res.Config.Mode != "continuous" {
		t.Errorf("expected mode continuous, got %s", res.Config.Mode)
	}
	if res.Total <= 0 {
		t.Errorf("expected > 0 records for today continuous, got %d", res.Total)
	}

	// 2. Test PUT /api/devices/34020000001180000001 to update record config
	profile.Record.Mode = "work_hours"
	profile.Record.SliceMinutes = 30
	bodyBytes, _ := json.Marshal(profile)

	putReq := httptest.NewRequest(http.MethodPut, "/api/devices/34020000001180000001", strings.NewReader(string(bodyBytes)))
	putRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d: %s", putRec.Code, putRec.Body.String())
	}

	// 3. Verify updated config reflected in records API
	req2 := httptest.NewRequest(http.MethodGet, "/api/devices/34020000001180000001/records?start="+today+"T00:00:00&end="+today+"T23:59:59", nil)
	rec2 := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec2, req2)

	var res2 struct {
		Total  int                 `json:"total"`
		Config config.RecordConfig `json:"config"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &res2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if res2.Config.Mode != "work_hours" {
		t.Errorf("expected mode work_hours, got %s", res2.Config.Mode)
	}
	if res2.Config.SliceMinutes != 30 {
		t.Errorf("expected slice_minutes 30, got %d", res2.Config.SliceMinutes)
	}
}

func TestServerDevicePTZAPI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gb-ui-server-ptz-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.New(tempDir)
	profile := config.DeviceProfile{
		Enabled: false,
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: 5060,
			LocalIP:    "127.0.0.1",
			LocalPort:  5072,
			Transport:  "udp",
			Password:   "12345678",
		},
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "NVR-PTZ",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "IPC-PTZ-1", PTZType: 1},
			},
		},
	}
	if err := store.Save(&profile); err != nil {
		t.Fatalf("Save profile error: %v", err)
	}

	mgr := device.NewManager(store)
	if err := mgr.Init(); err != nil {
		t.Fatalf("mgr init: %v", err)
	}
	if err := mgr.Start("34020000001180000001"); err != nil {
		t.Fatalf("mgr start device error: %v", err)
	}
	defer mgr.StopAll()

	srv := New(mgr)

	// 1. GET PTZ status
	getReq := httptest.NewRequest(http.MethodGet, "/api/devices/34020000001180000001/ptz?channel=34020000001320000001", nil)
	getRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var status device.PTZStatus
	if err := json.Unmarshal(getRec.Body.Bytes(), &status); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if status.ChannelID != "34020000001320000001" {
		t.Errorf("expected channelId 34020000001320000001, got %s", status.ChannelID)
	}
	if len(status.Presets) == 0 {
		t.Errorf("expected default presets, got 0")
	}

	// 2. POST PTZ move
	ctrlBody := `{"channelId":"34020000001320000001","action":"right","panSpeed":128}`
	ctrlReq := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/ptz/control", strings.NewReader(ctrlBody))
	ctrlRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(ctrlRec, ctrlReq)
	if ctrlRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on control, got %d: %s", ctrlRec.Code, ctrlRec.Body.String())
	}

	// 3. POST PTZ stop
	stopBody := `{"channelId":"34020000001320000001","action":"stop"}`
	stopReq := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/ptz/control", strings.NewReader(stopBody))
	stopRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(stopRec, stopReq)
	if stopRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on stop, got %d: %s", stopRec.Code, stopRec.Body.String())
	}

	// 4. POST add/set preset
	presetBody := `{"channelId":"34020000001320000001","presetId":8,"name":"新预置位8","useCurrent":true}`
	presetReq := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/ptz/preset", strings.NewReader(presetBody))
	presetRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(presetRec, presetReq)
	if presetRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on preset set, got %d: %s", presetRec.Code, presetRec.Body.String())
	}

	// 5. POST call preset
	callBody := `{"channelId":"34020000001320000001","presetId":1}`
	callReq := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/ptz/preset/call", strings.NewReader(callBody))
	callRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(callRec, callReq)
	if callRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on call preset, got %d: %s", callRec.Code, callRec.Body.String())
	}

	// 6. DELETE preset
	delReq := httptest.NewRequest(http.MethodDelete, "/api/devices/34020000001180000001/ptz/preset?channel=34020000001320000001&presetId=8", nil)
	delRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete preset, got %d: %s", delRec.Code, delRec.Body.String())
	}

	// 7. POST set_pose (direct angle drag/sync)
	setPoseBody := `{"channelId":"34020000001320000001","action":"set_pose","pan":135.5,"tilt":-15.0,"zoom":4.5}`
	setPoseReq := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/ptz/control", strings.NewReader(setPoseBody))
	setPoseRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(setPoseRec, setPoseReq)
	if setPoseRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on set_pose, got %d: %s", setPoseRec.Code, setPoseRec.Body.String())
	}
	var setPoseStatus device.PTZStatus
	_ = json.NewDecoder(setPoseRec.Body).Decode(&setPoseStatus)
	if setPoseStatus.Pan != 135.5 || setPoseStatus.Tilt != -15.0 || setPoseStatus.Zoom != 4.5 {
		t.Fatalf("unexpected coords after set_pose: %+v", setPoseStatus)
	}

	// 8. POST reset
	resetBody := `{"channelId":"34020000001320000001","action":"reset"}`
	resetReq := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/ptz/control", strings.NewReader(resetBody))
	resetRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(resetRec, resetReq)
	if resetRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on reset, got %d: %s", resetRec.Code, resetRec.Body.String())
	}
	var resetStatus device.PTZStatus
	_ = json.NewDecoder(resetRec.Body).Decode(&resetStatus)
	if resetStatus.Pan != 0.0 || resetStatus.Tilt != 0.0 || resetStatus.Zoom != 1.0 {
		t.Fatalf("unexpected coords after reset: %+v", resetStatus)
	}
}

