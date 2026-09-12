package device

import (
	"os"
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/storage"
)

func TestManagerPortAndCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gb-mgr-test-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.New(tempDir)
	mgr := NewManager(store)
	if err := mgr.Init(); err != nil {
		t.Fatalf("init error: %v", err)
	}

	p1 := mgr.NextAvailablePort()
	if p1 != 5070 {
		t.Errorf("expected 5070, got %d", p1)
	}

	// Add dev 1
	dev1 := &config.DeviceProfile{
		Enabled: false,
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
			Name: "NVR-1",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "IPC-1"},
			},
		},
	}

	if err := mgr.AddDevice(dev1); err != nil {
		t.Fatalf("add dev1: %v", err)
	}

	p2 := mgr.NextAvailablePort()
	if p2 != 5071 {
		t.Errorf("expected 5071, got %d", p2)
	}

	// Add dev 2 with port 5071
	dev2 := &config.DeviceProfile{
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
			Name: "NVR-2",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000002", Name: "IPC-2"},
			},
		},
	}

	if err := mgr.AddDevice(dev2); err != nil {
		t.Fatalf("add dev2: %v", err)
	}

	list := mgr.ListSummaries()
	if len(list) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(list))
	}

	// Try adding duplicate ID
	if err := mgr.AddDevice(dev1); err == nil {
		t.Errorf("expected duplicate ID error")
	}

	// Delete dev1
	if err := mgr.DeleteDevice("34020000001180000001"); err != nil {
		t.Fatalf("delete error: %v", err)
	}

	list = mgr.ListSummaries()
	if len(list) != 1 {
		t.Fatalf("expected 1 device, got %d", len(list))
	}
}

func TestManagerLiveBindAndShutdown(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gb-mgr-live-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.New(tempDir)
	mgr := NewManager(store)
	if err := mgr.Init(); err != nil {
		t.Fatalf("init error: %v", err)
	}

	dev := &config.DeviceProfile{
		Enabled: true,
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: 5060,
			LocalIP:    "127.0.0.1",
			LocalPort:  5170,
			Transport:  "udp",
			Password:   "123456",
		},
		Device: config.DeviceConfig{
			ID:   "34020000001180000003",
			Name: "NVR-Live",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "IPC-1"},
			},
		},
	}

	if err := mgr.AddDevice(dev); err != nil {
		t.Fatalf("add dev: %v", err)
	}

	// 模拟运行态绑定视频源（触发 OnConfigChanged），验证绝不自死锁
	doneCh := make(chan error, 1)
	go func() {
		err := mgr.BindChannelVideo("34020000001180000003", BindChannelRequest{
			ChannelID: "34020000001320000001",
			Source:    "ptz",
		})
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("BindChannelVideo error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("BindChannelVideo deadlocked!")
	}

	// 验证秒级退出
	stopDone := make(chan struct{})
	go func() {
		mgr.StopAll()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		// 成功退出
	case <-time.After(2 * time.Second):
		t.Fatalf("StopAll deadlocked or took too long!")
	}
}

