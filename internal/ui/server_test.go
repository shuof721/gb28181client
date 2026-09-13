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
	"github.com/local/gb28181-device/internal/media"
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

func TestServerChannelAudioAPI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gb-ui-audio-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.New(tempDir)
	profile := config.DeviceProfile{
		Enabled: false,
		SIP: config.SIPConfig{
			ServerIP: "127.0.0.1", ServerPort: 5060,
			LocalIP: "127.0.0.1", LocalPort: 5074,
			Transport: "udp", Password: "pass",
		},
		Device: config.DeviceConfig{
			ID: "34020000001180000001", Name: "NVR-Audio",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "IPC-1"},
				{ID: "34020000001320000002", Name: "IPC-2"},
			},
		},
		Media: config.MediaConfig{
			Source: "synthetic",
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

	// 1. 设置伴音为开启 + 切换为 sine
	enable := true
	body := `{"channelId":"34020000001320000001","audioEnabled":true,"audioSource":"sine"}`
	req := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/channels/audio", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on channel audio set, got %d: %s", rec.Code, rec.Body.String())
	}

	// 检查持久化配置
	loaded, err := store.Get("34020000001180000001")
	if err != nil {
		t.Fatalf("get profile failed: %v", err)
	}
	chCfg := loaded.Media.Channels["34020000001320000001"]
	if chCfg.AudioEnabled == nil || *chCfg.AudioEnabled != enable {
		t.Errorf("expected AudioEnabled=true, got %v", chCfg.AudioEnabled)
	}
	if chCfg.AudioSource != "sine" {
		t.Errorf("expected AudioSource=sine, got %s", chCfg.AudioSource)
	}

	// 2. 停用伴音
	disableBody := `{"channelId":"34020000001320000001","audioEnabled":false}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/channels/audio", strings.NewReader(disableBody))
	rec2 := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 on channel audio disable, got %d: %s", rec2.Code, rec2.Body.String())
	}

	loaded2, _ := store.Get("34020000001180000001")
	chCfg2 := loaded2.Media.Channels["34020000001320000001"]
	if chCfg2.AudioEnabled == nil || *chCfg2.AudioEnabled != false {
		t.Errorf("expected AudioEnabled=false, got %v", chCfg2.AudioEnabled)
	}

	// 3. 测试活跃推流中的热切换与无死锁验证
	_ = mgr.Start("34020000001180000001")
	dev, err := mgr.GetDevice("34020000001180000001")
	if err != nil {
		t.Fatalf("get device failed: %v", err)
	}
	sdp1 := &media.SDPInfo{IP: "127.0.0.1", VideoPort: 45000, SSRC: "0200001111"}
	_, err = dev.MediaManager().StartLive("34020000001320000001", "live-call-1", sdp1)
	if err != nil {
		t.Fatalf("start live session 1 failed: %v", err)
	}
	defer dev.MediaManager().StopAll()

	// 在推流活跃期间切换伴音为 beep
	beepBody := `{"channelId":"34020000001320000001","audioEnabled":true,"audioSource":"beep"}`
	req3 := httptest.NewRequest(http.MethodPost, "/api/devices/34020000001180000001/channels/audio", strings.NewReader(beepBody))
	rec3 := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 on hot-swap to beep, got %d: %s", rec3.Code, rec3.Body.String())
	}

	// 验证通道 1 会话依然存活且音源已热替换为 beep
	sessList := dev.MediaManager().ListSessions()
	if len(sessList) != 1 {
		t.Fatalf("expected 1 active session after hot-swap, got %d", len(sessList))
	}
	if sessList[0].AudioSource != "beep" {
		t.Errorf("expected hot-swapped audio source beep, got %s", sessList[0].AudioSource)
	}

	// 验证通道 2 发起点播不会被锁阻塞
	sdp2 := &media.SDPInfo{IP: "127.0.0.1", VideoPort: 45002, SSRC: "0200002222"}
	ans2, err := dev.MediaManager().StartLive("34020000001320000002", "live-call-2", sdp2)
	if err != nil || ans2 == "" {
		t.Fatalf("start live session 2 failed: %v", err)
	}
	if len(dev.MediaManager().ListSessions()) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(dev.MediaManager().ListSessions()))
	}
}


