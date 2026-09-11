package config

import "testing"

func TestNormalizeGBID(t *testing.T) {
	cases := map[string]string{
		"34020000001320000001":             "34020000001320000001",
		"34020000001320000001:0200006738":  "34020000001320000001",
		" 34020000001320000001:0200003436": "34020000001320000001",
	}
	for in, want := range cases {
		if got := NormalizeGBID(in); got != want {
			t.Fatalf("NormalizeGBID(%q)=%q want %q", in, got, want)
		}
	}
}

func TestOptionsForWithSSRCSuffix(t *testing.T) {
	m := &MediaConfig{
		Mode:   "per_channel",
		Source: "synthetic",
		Channels: map[string]ChannelMediaConfig{
			"34020000001320000001": {Source: "mp4", MP4File: "ch1.mp4"},
		},
	}
	v := m.OptionsFor("34020000001320000001:0200006738")
	if v.Kind != "mp4" || v.MP4 != "ch1.mp4" {
		t.Fatalf("got %+v", v)
	}
}

func TestMediaOptionsShared(t *testing.T) {
	m := &MediaConfig{
		Mode:     "shared",
		Source:   "mp4",
		MP4File:  "a.mp4",
		H264File: "a.h264",
		FPS:      25,
		Width:    1280,
		Height:   720,
	}
	v := m.OptionsFor("34020000001320000001")
	if v.Kind != "mp4" || v.MP4 != "a.mp4" {
		t.Fatalf("%+v", v)
	}
}

func TestMediaOptionsPerChannel(t *testing.T) {
	m := &MediaConfig{
		Mode:     "per_channel",
		Source:   "mp4",
		MP4File:  "global.mp4",
		H264File: "global.h264",
		Channels: map[string]ChannelMediaConfig{
			"34020000001320000001": {Source: "mp4", MP4File: "ch1.mp4"},
			"34020000001320000002": {Source: "file", H264File: "ch2.h264"},
		},
	}
	v1 := m.OptionsFor("34020000001320000001")
	if v1.MP4 != "ch1.mp4" {
		t.Fatalf("ch1: %+v", v1)
	}
	v2 := m.OptionsFor("34020000001320000002")
	if v2.Kind != "file" || v2.H264 != "ch2.h264" {
		t.Fatalf("ch2: %+v", v2)
	}
	// 未配置 → 回退全局
	v3 := m.OptionsFor("34020000001320000099")
	if v3.MP4 != "global.mp4" {
		t.Fatalf("fallback: %+v", v3)
	}
}
