package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/local/gb28181-device/internal/config"
)

func TestStoreCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gb-store-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s := New(tempDir)

	// Initially empty
	list, err := s.List()
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0, got %d", len(list))
	}

	// Save profile
	p := &config.DeviceProfile{
		Enabled: true,
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: 5060,
			LocalIP:    "127.0.0.1",
			LocalPort:  5070,
			Transport:  "udp",
			Password:   "123456",
		},
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "TestDevice1",
			Channels: []config.ChannelConfig{
				{
					ID:   "34020000001320000001",
					Name: "Camera1",
				},
			},
		},
	}

	if err := s.Save(p); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Verify target file exists
	targetPath := filepath.Join(tempDir, "34020000001180000001.json")
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected file %s: %v", targetPath, err)
	}

	// Get
	got, err := s.Get("34020000001180000001")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if got.Device.Name != "TestDevice1" {
		t.Errorf("expected TestDevice1, got %s", got.Device.Name)
	}
	if got.SIP.LocalPort != 5070 {
		t.Errorf("expected 5070, got %d", got.SIP.LocalPort)
	}

	// Save another
	p2 := &config.DeviceProfile{
		Enabled: false,
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: 5060,
			LocalIP:    "127.0.0.1",
			LocalPort:  5071,
			Transport:  "udp",
			Password:   "123456",
		},
		Device: config.DeviceConfig{
			ID:   "34020000001180000002",
			Name: "TestDevice2",
			Channels: []config.ChannelConfig{
				{
					ID:   "34020000001320000002",
					Name: "Camera2",
				},
			},
		},
	}
	if err := s.Save(p2); err != nil {
		t.Fatalf("save p2: %v", err)
	}

	list, err = s.List()
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}

	// Delete
	if err := s.Delete("34020000001180000001"); err != nil {
		t.Fatalf("delete error: %v", err)
	}
	list, err = s.List()
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}
}
