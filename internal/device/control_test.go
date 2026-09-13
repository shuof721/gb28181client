package device

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/gb28181"
	"github.com/local/gb28181-device/internal/sip"
)

func encXML(s string) []byte {
	b, _ := gb28181.EncodeXML(s, "GB2312")
	return b
}

func TestConfigDownloadOverSIP(t *testing.T) {
	serverUA := sip.NewUA("127.0.0.1", 59281, "127.0.0.1", 59282, "udp", "", "")
	if err := serverUA.Start(); err != nil {
		t.Fatal(err)
	}
	defer serverUA.Close()

	respChan := make(chan string, 10)
	serverUA.Handle("MESSAGE", func(m *sip.Message, src net.Addr) {
		_ = serverUA.Reply(m, src, 200, "OK", nil, "")
		respChan <- string(m.Body)
	})

	cfg := &config.Config{
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "测试NVR设备",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "一号通道", PTZType: 1},
			},
		},
		SIP: config.SIPConfig{
			LocalIP:               "127.0.0.1",
			LocalPort:             59282,
			ServerIP:              "127.0.0.1",
			ServerPort:            59281,
			Transport:             "udp",
			Expires:               3600,
			KeepaliveInterval:     60,
			KeepaliveTimeoutCount: 3,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatal(err)
	}
	defer dev.Stop()

	// 1. BasicParam 查询
	q1 := `<?xml version="1.0" encoding="GB2312"?>
<Query>
  <CmdType>ConfigDownload</CmdType>
  <SN>1001</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <ConfigType>BasicParam</ConfigType>
</Query>`
	_, err := serverUA.SendMessage(encXML(q1), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<CmdType>ConfigDownload</CmdType>") {
			t.Fatalf("expected ConfigDownload response, got: %s", resp)
		}
		if !strings.Contains(resp, "<BasicParam>") || !strings.Contains(resp, "<HeartBeatInterval>60</HeartBeatInterval>") {
			t.Fatalf("missing BasicParam or HeartBeatInterval in: %s", resp)
		}
		if !strings.Contains(resp, "<PositionCapability>1</PositionCapability>") {
			t.Fatalf("missing PositionCapability in: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for BasicParam response")
	}

	// 2. VideoParamOpt 查询
	q2 := `<?xml version="1.0" encoding="GB2312"?>
<Query>
  <CmdType>ConfigDownload</CmdType>
  <SN>1002</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <ConfigType>VideoParamOpt</ConfigType>
</Query>`
	_, err = serverUA.SendMessage(encXML(q2), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<VideoParamOpt>") || !strings.Contains(resp, "<DownloadSpeed>") {
			t.Fatalf("missing VideoParamOpt in: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for VideoParamOpt response")
	}

	// 3. AudioParamOpt 查询
	q3 := `<?xml version="1.0" encoding="GB2312"?>
<Query>
  <CmdType>ConfigDownload</CmdType>
  <SN>1003</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <ConfigType>AudioParamOpt</ConfigType>
</Query>`
	_, err = serverUA.SendMessage(encXML(q3), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<AudioParamOpt>") || !strings.Contains(resp, "<AudioFormat>") {
			t.Fatalf("missing AudioParamOpt in: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for AudioParamOpt response")
	}

	// 4. 验证审计日志记录
	events := dev.GetControlEvents()
	if len(events) < 3 {
		t.Fatalf("expected >= 3 audit events, got %d", len(events))
	}
	if events[0].CmdType != "ConfigDownload" {
		t.Fatalf("expected ConfigDownload event, got %s", events[0].CmdType)
	}
}

func TestDeviceConfigOverSIP(t *testing.T) {
	serverUA := sip.NewUA("127.0.0.1", 59283, "127.0.0.1", 59284, "udp", "", "")
	if err := serverUA.Start(); err != nil {
		t.Fatal(err)
	}
	defer serverUA.Close()

	respChan := make(chan string, 10)
	serverUA.Handle("MESSAGE", func(m *sip.Message, src net.Addr) {
		_ = serverUA.Reply(m, src, 200, "OK", nil, "")
		respChan <- string(m.Body)
	})

	cfg := &config.Config{
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "旧设备名",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "一号通道"},
			},
		},
		SIP: config.SIPConfig{
			LocalIP:    "127.0.0.1",
			LocalPort:  59284,
			ServerIP:   "127.0.0.1",
			ServerPort: 59283,
			Transport:  "udp",
			Expires:    3600,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatal(err)
	}
	defer dev.Stop()

	// 1. 平台校时 (Time)
	c1 := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceConfig</CmdType>
  <SN>2001</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <Time>2028-11-20T15:30:00</Time>
</Control>`
	_, err := serverUA.SendMessage(encXML(c1), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for DeviceConfig response")
	}

	devTime := dev.Now()
	if devTime.Year() != 2028 || devTime.Month() != time.November || devTime.Day() != 20 {
		t.Fatalf("unexpected devTime after calibration: %v", devTime)
	}
	if dev.Status().TimeOffsetSec == 0 {
		t.Fatalf("expected non-zero TimeOffsetSec")
	}

	// 复位时间偏移
	dev.ResetTimeOffset()
	if dev.Status().TimeOffsetSec != 0 {
		t.Fatalf("expected 0 TimeOffsetSec after reset")
	}

	// 2. 修改设备名称与心跳周期
	c2 := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceConfig</CmdType>
  <SN>2002</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <BasicParam>
    <Name>远程更名测试设备</Name>
    <HeartBeatInterval>45</HeartBeatInterval>
    <HeartBeatCount>5</HeartBeatCount>
  </BasicParam>
</Control>`
	_, err = serverUA.SendMessage(encXML(c2), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for DeviceConfig response")
	}

	if dev.cfg.Device.Name != "远程更名测试设备" {
		t.Fatalf("expected new device name, got: %s", dev.cfg.Device.Name)
	}
	if dev.cfg.SIP.KeepaliveInterval != 45 {
		t.Fatalf("expected keepalive interval 45, got: %d", dev.cfg.SIP.KeepaliveInterval)
	}
}

func TestDeviceControlOverSIP(t *testing.T) {
	serverUA := sip.NewUA("127.0.0.1", 59285, "127.0.0.1", 59286, "udp", "", "")
	if err := serverUA.Start(); err != nil {
		t.Fatal(err)
	}
	defer serverUA.Close()

	respChan := make(chan string, 10)
	serverUA.Handle("MESSAGE", func(m *sip.Message, src net.Addr) {
		_ = serverUA.Reply(m, src, 200, "OK", nil, "")
		respChan <- string(m.Body)
	})

	cfg := &config.Config{
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "测试NVR设备",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "一号通道", PTZType: 1},
			},
		},
		SIP: config.SIPConfig{
			LocalIP:    "127.0.0.1",
			LocalPort:  59286,
			ServerIP:   "127.0.0.1",
			ServerPort: 59285,
			Transport:  "udp",
			Expires:    3600,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatal(err)
	}
	defer dev.Stop()

	// 1. 远程启动录像 (RecordCmd: Record)
	ctrlRec := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>3001</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <RecordCmd>Record</RecordCmd>
</Control>`
	_, err := serverUA.SendMessage(encXML(ctrlRec), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for RecordCmd response")
	}

	if !dev.IsRecording() || !dev.Status().Recording {
		t.Fatalf("expected device recording to be true")
	}

	// 2. 状态查询应联动输出 <Record>ON</Record>
	stQuery := `<?xml version="1.0" encoding="GB2312"?>
<Query>
  <CmdType>DeviceStatus</CmdType>
  <SN>3002</SN>
  <DeviceID>34020000001180000001</DeviceID>
</Query>`
	_, err = serverUA.SendMessage(encXML(stQuery), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Record>ON</Record>") {
			t.Fatalf("expected <Record>ON</Record> in DeviceStatus: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for DeviceStatus response")
	}

	// 3. 停止录像 (RecordCmd: StopRecord)
	ctrlStopRec := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>3003</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <RecordCmd>StopRecord</RecordCmd>
</Control>`
	_, err = serverUA.SendMessage(encXML(ctrlStopRec), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for StopRecord response")
	}

	if dev.IsRecording() || dev.Status().Recording {
		t.Fatalf("expected device recording to be false")
	}

	// 4. 强制请求关键帧 (WVP 兼容 IFameCmd)
	ctrlIFrame := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>3004</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <IFameCmd>1</IFameCmd>
</Control>`
	_, err = serverUA.SendMessage(encXML(ctrlIFrame), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for IFameCmd response")
	}

	// 5. 设置云台看守位 (HomePosition)
	ctrlHome := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>3005</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <HomePosition>
    <Enabled>1</Enabled>
    <ResetTime>45</ResetTime>
    <PresetIndex>3</PresetIndex>
  </HomePosition>
</Control>`
	_, err = serverUA.SendMessage(encXML(ctrlHome), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for HomePosition response")
	}

	ptz := dev.GetChannelPTZ("34020000001320000001")
	if ptz == nil {
		t.Fatalf("GetChannelPTZ returned nil")
	}
	ptzSt := ptz.Status()
	if !ptzSt.HomePositionEnabled || ptzSt.HomePositionPreset != 3 || ptzSt.HomePositionResetSec != 45 {
		t.Fatalf("unexpected PTZ HomePosition status: %+v", ptzSt)
	}

	// 6. 模拟远程重启 (TeleBoot)
	ctrlBoot := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>3006</SN>
  <DeviceID>34020000001180000001</DeviceID>
  <TeleBoot>Boot</TeleBoot>
</Control>`
	_, err = serverUA.SendMessage(encXML(ctrlBoot), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK, got: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for TeleBoot response")
	}

	if !dev.IsRebooting() {
		t.Fatalf("expected IsRebooting to be true")
	}
}
