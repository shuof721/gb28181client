package device

import (
	"math"
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
)

func TestGPSManagerTrajectory(t *testing.T) {
	cfg := config.MobilePositionConfig{
		Enabled:   true,
		ChannelID: "34020000001320000001",
		Interval:  1,
		Mode:      "both",
		Pattern:   "circle",
		Longitude: 116.397428,
		Latitude:  39.909230,
		Altitude:  60.0,
		Speed:     36.0, // 36 km/h = 10 m/s
		Direction: 90.0,
		Radius:    500.0,
	}

	channels := []config.ChannelConfig{
		{ID: "34020000001320000001"},
	}

	var reported []GPSStatus
	gm := NewGPSManager(cfg, channels, func(st GPSStatus) error {
		reported = append(reported, st)
		return nil
	})

	// 1. 测试静态与初始位置
	st0 := gm.Current("34020000001320000001")
	if st0.ChannelID != cfg.ChannelID {
		t.Fatalf("expected ChannelID %s, got %s", cfg.ChannelID, st0.ChannelID)
	}
	if math.Abs(st0.Altitude-60.0) > 0.001 {
		t.Fatalf("expected Altitude 60, got %f", st0.Altitude)
	}
	if st0.Speed != 36.0 {
		t.Fatalf("expected Speed 36, got %f", st0.Speed)
	}

	// 2. 测试计算圆周运动
	lon, lat, dir := calculateTrajectory("circle", 116.397428, 39.909230, 500.0, 10.0, 0, 10.0)
	if lon == 116.397428 && lat == 39.909230 {
		t.Fatal("expected coordinates to change on circle trajectory")
	}
	if dir < 0 || dir > 360 {
		t.Fatalf("invalid heading direction %f", dir)
	}

	// 3. 测试线性往返运动
	lon1, _, dir1 := calculateTrajectory("linear", 116.397428, 39.909230, 100.0, 10.0, 90.0, 2.0)
	if lon1 <= 116.397428 {
		t.Fatal("expected eastward movement for 90 degree heading")
	}
	if dir1 != 90.0 {
		t.Fatalf("expected forward dir 90, got %f", dir1)
	}
	// 往返折返：周期为 20s (halfPeriod = 10s)，在 15s 时应该折返向西 (270度)
	_, _, dir2 := calculateTrajectory("linear", 116.397428, 39.909230, 100.0, 10.0, 90.0, 15.0)
	if dir2 != 270.0 {
		t.Fatalf("expected backward dir 270, got %f", dir2)
	}

	// 4. 测试手动单次上报
	stReport, err := gm.ReportNow("34020000001320000001")
	if err != nil {
		t.Fatalf("ReportNow error: %v", err)
	}
	if len(reported) != 1 {
		t.Fatalf("expected 1 reported item, got %d", len(reported))
	}
	if reported[0].ChannelID != stReport.ChannelID {
		t.Fatalf("mismatched reported channel %s vs %s", reported[0].ChannelID, stReport.ChannelID)
	}

	// 5. 启停测试
	gm.Start()
	time.Sleep(50 * time.Millisecond)
	gm.Stop()
}

func TestGPSManagerMultiChannelAndFollow(t *testing.T) {
	masterCfg := config.MobilePositionConfig{
		Enabled:   true,
		Pattern:   "circle",
		Longitude: 120.123456,
		Latitude:  30.123456,
		Radius:    200.0,
		Speed:     72.0,
	}

	ch1Cfg := config.MobilePositionConfig{
		Enabled:   true,
		Pattern:   "static",
		Longitude: 110.0,
		Latitude:  20.0,
	}

	channels := []config.ChannelConfig{
		{
			ID:             "ch1",
			MobilePosition: &ch1Cfg,
		},
		{
			ID: "ch2", // 未指定，自动进入 follow 模式
		},
	}

	var reports []GPSStatus
	gm := NewGPSManager(masterCfg, channels, func(st GPSStatus) error {
		reports = append(reports, st)
		return nil
	})

	// 检查通道 1 (独立静态配置)
	st1 := gm.Current("ch1")
	if st1.ChannelID != "ch1" {
		t.Fatalf("expected ch1, got %s", st1.ChannelID)
	}
	if st1.Longitude != 110.0 || st1.Latitude != 20.0 {
		t.Fatalf("ch1 should retain static coords, got %f, %f", st1.Longitude, st1.Latitude)
	}

	// 检查通道 2 (跟随主车模式)
	st2 := gm.Current("ch2")
	if st2.ChannelID != "ch2" {
		t.Fatalf("expected ch2, got %s", st2.ChannelID)
	}
	if st2.Pattern != "follow" {
		t.Fatalf("ch2 should be follow pattern, got %s", st2.Pattern)
	}
	// 跟随主车的坐标应该在主车附近（120.123456, 30.123456）
	if math.Abs(st2.Longitude-120.123456) > 0.05 || math.Abs(st2.Latitude-30.123456) > 0.05 {
		t.Fatalf("ch2 follow coords too far from master: %f, %f", st2.Longitude, st2.Latitude)
	}

	// 测试 CurrentAll()
	all := gm.CurrentAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 active channel statuses, got %d", len(all))
	}

	// 测试 SyncAllChannels 为 follow 模式
	gm.SyncAllChannels(masterCfg, true)
	ch1Updated := gm.GetChannelConfig("ch1")
	if ch1Updated.Pattern != "follow" {
		t.Fatalf("expected ch1 pattern to be follow after sync, got %s", ch1Updated.Pattern)
	}
}
