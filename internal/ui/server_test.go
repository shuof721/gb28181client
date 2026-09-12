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
