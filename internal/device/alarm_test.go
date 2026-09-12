package device

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/sip"
)

func TestAlarmManagerGuardCascadeAndDutyStatus(t *testing.T) {
	am := NewAlarmManager()

	// 1. 默认撤防 (OFFDUTY)
	if st := am.GetGuard("ch-1"); st != "ResetGuard" {
		t.Fatalf("expected ResetGuard, got %s", st)
	}
	if dt := am.DutyStatus("ch-1"); dt != "OFFDUTY" {
		t.Fatalf("expected OFFDUTY, got %s", dt)
	}

	// 2. 单通道布防 (ONDUTY)
	am.SetGuard("ch-1", true)
	if st := am.GetGuard("ch-1"); st != "SetGuard" {
		t.Fatalf("expected SetGuard, got %s", st)
	}
	if dt := am.DutyStatus("ch-1"); dt != "ONDUTY" {
		t.Fatalf("expected ONDUTY, got %s", dt)
	}
	// 未设置的通道仍然继承默认撤防
	if st := am.GetGuard("ch-2"); st != "ResetGuard" {
		t.Fatalf("expected ResetGuard for ch-2, got %s", st)
	}

	// 3. 通道产生报警 -> 变为 ALARM 状态
	am.SetAlarming("ch-1", true)
	if !am.IsAlarming("ch-1") {
		t.Fatal("expected ch-1 alarming")
	}
	if dt := am.DutyStatus("ch-1"); dt != "ALARM" {
		t.Fatalf("expected ALARM status when active, got %s", dt)
	}

	// 4. 复位报警 (ResetAlarm) -> 恢复回 ONDUTY (因为已布防)
	am.ResetAlarm("ch-1")
	if am.IsAlarming("ch-1") {
		t.Fatal("expected ch-1 alarming cleared")
	}
	if dt := am.DutyStatus("ch-1"); dt != "ONDUTY" {
		t.Fatalf("expected ONDUTY status after reset alarm, got %s", dt)
	}

	// 5. 全局根设备级联布防
	allChs := []string{"ch-1", "ch-2", "ch-3"}
	am.SetGuard("", true, allChs)
	for _, ch := range allChs {
		if st := am.GetGuard(ch); st != "SetGuard" {
			t.Fatalf("expected cascaded SetGuard for %s, got %s", ch, st)
		}
		if dt := am.DutyStatus(ch); dt != "ONDUTY" {
			t.Fatalf("expected ONDUTY for %s, got %s", ch, dt)
		}
	}

	// 6. 全局级联撤防
	am.SetGuard("", false, allChs)
	for _, ch := range allChs {
		if st := am.GetGuard(ch); st != "ResetGuard" {
			t.Fatalf("expected cascaded ResetGuard for %s, got %s", ch, st)
		}
		if dt := am.DutyStatus(ch); dt != "OFFDUTY" {
			t.Fatalf("expected OFFDUTY for %s, got %s", ch, dt)
		}
	}

	// 7. 报警历史记录添加与列表倒序
	am.AddRecord(AlarmEventRecord{ID: "1", Description: "alarm 1", Time: "2026-09-12T10:00:00"})
	am.AddRecord(AlarmEventRecord{ID: "2", Description: "alarm 2", Time: "2026-09-12T10:00:01"})
	am.AddRecord(AlarmEventRecord{ID: "3", Description: "alarm 3", Time: "2026-09-12T10:00:02"})

	list := am.ListRecords(2)
	if len(list) != 2 {
		t.Fatalf("expected 2 records, got %d", len(list))
	}
	if list[0].ID != "3" || list[1].ID != "2" {
		t.Fatalf("expected latest records first, got %s, %s", list[0].ID, list[1].ID)
	}
}

func TestAlarmManagerAutoAlarm(t *testing.T) {
	am := NewAlarmManager()
	if am.IsAutoAlarmRunning() {
		t.Fatal("expected auto alarm not running initially")
	}

	// 布防 ch-1
	am.SetGuard("ch-1", true)

	triggerCount := 0
	triggerCh := make(chan string, 10)
	am.StartAutoAlarm(10*time.Millisecond, []string{"ch-1", "ch-2"}, func(ch string) {
		triggerCount++
		triggerCh <- ch
	})

	if !am.IsAutoAlarmRunning() {
		t.Fatal("expected auto alarm running")
	}

	// 自动报警只应触发已布防的 ch-1，ch-2 撤防不应触发
	select {
	case ch := <-triggerCh:
		if ch != "ch-1" {
			t.Fatalf("unexpected channel (should only trigger armed channel ch-1): %s", ch)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for auto alarm trigger")
	}

	am.StopAutoAlarm()
	if am.IsAutoAlarmRunning() {
		t.Fatal("expected auto alarm stopped")
	}
}

func TestDeviceSendAlarmEventWithGatekeeperLinkage(t *testing.T) {
	serverReceived := make(chan string, 5)
	serverUA := sip.NewUA("127.0.0.1", 59190, "127.0.0.1", 59191, "udp", "", "")
	if err := serverUA.Start(); err != nil {
		t.Fatal(err)
	}
	defer serverUA.Close()

	serverUA.Handle("MESSAGE", func(m *sip.Message, src net.Addr) {
		_ = serverUA.Reply(m, src, 200, "OK", nil, "")
		serverReceived <- string(m.Body)
	})

	cfg := &config.Config{
		Device: config.DeviceConfig{
			ID:   "34020000001180000001",
			Name: "测试设备",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "一号枪机"},
			},
		},
		SIP: config.SIPConfig{
			LocalIP:    "127.0.0.1",
			LocalPort:  59191,
			ServerIP:   "127.0.0.1",
			ServerPort: 59190,
			Transport:  "udp",
			Expires:    3600,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatal(err)
	}
	defer dev.Stop()

	chID := "34020000001320000001"

	// 1. 初始状态为撤防 (ResetGuard)
	// 尝试触发常规移动侦测报警 (Method=5, Type=2, Priority=3) -> 应被门禁拦截
	recSuppressed, err := dev.SendAlarmEvent(chID, "5", "2", "3", "撤防测试", 0, 0)
	if err == nil {
		t.Fatal("expected error when sending conventional alarm on disarmed channel")
	}
	if recSuppressed == nil || recSuppressed.Status != "suppressed" {
		t.Fatalf("expected suppressed status, got %+v", recSuppressed)
	}
	// 验证未向平台发送任何 SIP MESSAGE
	select {
	case body := <-serverReceived:
		t.Fatalf("unexpected message sent to platform when channel is disarmed: %s", body)
	default:
	}

	// 2. 撤防状态下触发 24 小时特种紧急报警 (Priority=1) -> 豁免放行并上报
	recEmerg, err := dev.SendAlarmEvent(chID, "2", "5", "1", "紧急求助SOS", 116.39, 39.91)
	if err != nil {
		t.Fatalf("emergency alarm should bypass disarm gatekeeper: %v", err)
	}
	if recEmerg.Status != "confirmed" {
		t.Fatalf("expected confirmed for emergency alarm, got %s", recEmerg.Status)
	}
	select {
	case body := <-serverReceived:
		if !strings.Contains(body, "<AlarmPriority>1</AlarmPriority>") {
			t.Fatalf("missing priority 1 in body: %s", body)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for emergency alarm notify")
	}

	// 3. 将通道布防 (SetGuard)
	dev.AlarmManager().SetGuard(chID, true)
	if dev.AlarmManager().GetGuard(chID) != "SetGuard" {
		t.Fatal("expected SetGuard")
	}

	// 4. 再次触发常规周界入侵报警 (Priority=2) -> 允许上报，且通道联动切换为 ALARM
	recArmed, err := dev.SendAlarmEvent(chID, "5", "6", "2", "周界入侵测试", 116.397128, 39.916527)
	if err != nil {
		t.Fatalf("SendAlarmEvent failed on armed channel: %v", err)
	}
	if recArmed.Status != "confirmed" {
		t.Fatalf("expected confirmed, got %s", recArmed.Status)
	}
	// 验证通道已进入 ALARM 状态
	if dt := dev.AlarmManager().DutyStatus(chID); dt != "ALARM" {
		t.Fatalf("expected channel to enter ALARM duty status, got %s", dt)
	}

	select {
	case body := <-serverReceived:
		if !strings.Contains(body, "<CmdType>Alarm</CmdType>") {
			t.Fatalf("missing CmdType: %s", body)
		}
		if !strings.Contains(body, "<AlarmType>6</AlarmType>") {
			t.Fatalf("missing AlarmType 6: %s", body)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for armed alarm notify")
	}

	// 5. 报警复位 (ResetAlarm) -> 通道清除报警恢复为 ONDUTY
	dev.AlarmManager().ResetAlarm(chID)
	if dt := dev.AlarmManager().DutyStatus(chID); dt != "ONDUTY" {
		t.Fatalf("expected duty status to return to ONDUTY after reset, got %s", dt)
	}
}

func TestDeviceControlAndDeviceStatusAlarmstatusXML(t *testing.T) {
	serverUA := sip.NewUA("127.0.0.1", 59194, "127.0.0.1", 59195, "udp", "", "")
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
			Name: "测试设备",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000001", Name: "一号枪机"},
				{ID: "34020000001320000002", Name: "二号球机"},
			},
		},
		SIP: config.SIPConfig{
			LocalIP:    "127.0.0.1",
			LocalPort:  59195,
			ServerIP:   "127.0.0.1",
			ServerPort: 59194,
			Transport:  "udp",
			Expires:    3600,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatal(err)
	}
	defer dev.Stop()

	// 1. 平台对通道1单独下发 SetGuard 布防
	guardBody := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>112233</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <GuardCmd>SetGuard</GuardCmd>
</Control>`
	_, err := serverUA.SendMessage([]byte(guardBody), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK in response: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for device response to SetGuard")
	}

	if st := dev.AlarmManager().GetGuard("34020000001320000001"); st != "SetGuard" {
		t.Fatalf("expected SetGuard for ch1, got %s", st)
	}
	if st := dev.AlarmManager().GetGuard("34020000001320000002"); st != "ResetGuard" {
		t.Fatalf("expected ResetGuard for ch2, got %s", st)
	}

	// 2. 模拟通道1产生报警进入 ALARM 状态
	dev.AlarmManager().SetAlarming("34020000001320000001", true)

	// 3. 平台下发 DeviceStatus 设备状态查询
	statusQueryBody := `<?xml version="1.0" encoding="GB2312"?>
<Query>
  <CmdType>DeviceStatus</CmdType>
  <SN>998877</SN>
  <DeviceID>34020000001180000001</DeviceID>
</Query>`
	_, err = serverUA.SendMessage([]byte(statusQueryBody), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<CmdType>DeviceStatus</CmdType>") {
			t.Fatalf("expected DeviceStatus response, got %s", resp)
		}
		// 校验国标 Table A.4 Alarmstatus 节点
		if !strings.Contains(resp, "<Alarmstatus Num=\"2\">") {
			t.Fatalf("missing Alarmstatus Num=2 in response: %s", resp)
		}
		// 通道1应为 ALARM
		if !strings.Contains(resp, "<DeviceID>34020000001320000001</DeviceID>") || !strings.Contains(resp, "<DutyStatus>ALARM</DutyStatus>") {
			t.Fatalf("expected ch1 to be ALARM: %s", resp)
		}
		// 通道2应为 OFFDUTY
		if !strings.Contains(resp, "<DeviceID>34020000001320000002</DeviceID>") || !strings.Contains(resp, "<DutyStatus>OFFDUTY</DutyStatus>") {
			t.Fatalf("expected ch2 to be OFFDUTY: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for DeviceStatus response")
	}

	// 4. 平台发送 AlarmCmd: ResetAlarm 复位报警
	resetAlarmBody := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>112234</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <AlarmCmd>ResetAlarm</AlarmCmd>
</Control>`
	_, err = serverUA.SendMessage([]byte(resetAlarmBody), "Application/MANSCDP+xml")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-respChan:
		if !strings.Contains(resp, "<Result>OK</Result>") {
			t.Fatalf("expected Result OK in response: %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for device response to ResetAlarm")
	}

	// 验证通道1复位后恢复为 ONDUTY
	if dt := dev.AlarmManager().DutyStatus("34020000001320000001"); dt != "ONDUTY" {
		t.Fatalf("expected ch1 duty status to be ONDUTY after reset, got %s", dt)
	}
}
