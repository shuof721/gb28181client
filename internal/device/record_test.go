package device

import (
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
)

func TestParseFlexibleTime(t *testing.T) {
	cases := []struct {
		input string
		year  int
		month time.Month
		day   int
		hour  int
		min   int
		sec   int
	}{
		{"2026-09-12T14:30:00", 2026, 9, 12, 14, 30, 0},
		{"2026-09-12 14:30:00", 2026, 9, 12, 14, 30, 0},
		{"2026-09-12 14:30:00.123", 2026, 9, 12, 14, 30, 0},
		{"20260912143000", 2026, 9, 12, 14, 30, 0},
		{"2026-09-12", 2026, 9, 12, 0, 0, 0},
	}

	for _, c := range cases {
		tm, err := ParseFlexibleTime(c.input)
		if err != nil {
			t.Fatalf("parse %q failed: %v", c.input, err)
		}
		if tm.Year() != c.year || tm.Month() != c.month || tm.Day() != c.day ||
			tm.Hour() != c.hour || tm.Minute() != c.min || tm.Second() != c.sec {
			t.Errorf("expected %v-%v-%v %v:%v:%v, got %v", c.year, c.month, c.day, c.hour, c.min, c.sec, tm)
		}
	}

	// Test Unix timestamp
	ts := time.Date(2026, 9, 12, 12, 0, 0, 0, time.Local)
	secStr := string(rune(ts.Unix())) // test with ParseFlexibleTime using string
	tmSec, err := ParseFlexibleTime("1789214400")
	if err != nil {
		t.Fatalf("parse unix sec failed: %v", err)
	}
	if tmSec.Unix() != 1789214400 {
		t.Errorf("expected 1789214400, got %d", tmSec.Unix())
	}
	_ = secStr
}

func TestGenerateRecordItems(t *testing.T) {
	cfg := config.RecordConfig{
		Enabled:      true,
		Mode:         "continuous",
		SliceMinutes: 60,
		RetainDays:   7,
		RecordType:   "time",
		FileSizeMB:   100,
	}

	now := time.Now()
	// Yesterday 00:00:00 to 06:00:00 (6 hours => 6 slices of 60m)
	yest := now.AddDate(0, 0, -1)
	startStr := time.Date(yest.Year(), yest.Month(), yest.Day(), 0, 0, 0, 0, yest.Location()).Format("2006-01-02T15:04:05")
	endStr := time.Date(yest.Year(), yest.Month(), yest.Day(), 6, 0, 0, 0, yest.Location()).Format("2006-01-02T15:04:05")

	items := GenerateRecordItems(cfg, "34020000001320000001", "Camera-1", "34020000001180000001", startStr, endStr, "all")
	if len(items) != 6 {
		t.Errorf("expected 6 slices for 6 hours, got %d", len(items))
	}

	// Test expiration outside retain days (e.g. 30 days ago with RetainDays=7)
	past := now.AddDate(0, 0, -30)
	pastStart := time.Date(past.Year(), past.Month(), past.Day(), 0, 0, 0, 0, past.Location()).Format("2006-01-02T15:04:05")
	pastEnd := time.Date(past.Year(), past.Month(), past.Day(), 6, 0, 0, 0, past.Location()).Format("2006-01-02T15:04:05")

	expiredItems := GenerateRecordItems(cfg, "34020000001320000001", "Camera-1", "34020000001180000001", pastStart, pastEnd, "all")
	if len(expiredItems) != 0 {
		t.Errorf("expected 0 slices for expired date, got %d", len(expiredItems))
	}

	// Test disabled
	cfg.Enabled = false
	disabledItems := GenerateRecordItems(cfg, "34020000001320000001", "Camera-1", "34020000001180000001", startStr, endStr, "all")
	if len(disabledItems) != 0 {
		t.Errorf("expected 0 slices when disabled, got %d", len(disabledItems))
	}
}
