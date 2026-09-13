package gb28181

import "testing"

func TestParseCatalogQuery(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Query>
  <CmdType>Catalog</CmdType>
  <SN>17</SN>
  <DeviceID>34020000001180000001</DeviceID>
</Query>`)
	root, err := ParseRoot(body)
	if err != nil {
		t.Fatal(err)
	}
	if root.CmdType != "Catalog" || root.SN != "17" || root.DeviceID != "34020000001180000001" {
		t.Fatalf("%+v", root)
	}
}

func TestParseGB2312Query(t *testing.T) {
	// WVP 等平台常见 encoding="gb2312"
	body := []byte(`<?xml version="1.0" encoding="gb2312"?>
<Query>
  <CmdType>Catalog</CmdType>
  <SN>5</SN>
  <DeviceID>34020000001180000001</DeviceID>
</Query>`)
	root, err := ParseRoot(body)
	if err != nil {
		t.Fatal(err)
	}
	if root.CmdType != "Catalog" || root.SN != "5" {
		t.Fatalf("%+v", root)
	}
}

func TestDecodePTZ(t *testing.T) {
	// stop
	if DecodePTZ("A50F0100000000B5") == "unknown" {
		// 构造一个 zoom-in: A5 0F 01 ...
	}
	// 典型水平右转指令片段（byte2 高4位=1）
	name := DecodePTZ("A50F0110000000B5")
	if name == "unknown" {
		t.Fatal("should decode")
	}
}

func TestParseRecordInfoQuery(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Query>
  <CmdType>RecordInfo</CmdType>
  <SN>12345</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <StartTime>2026-09-11T08:00:00</StartTime>
  <EndTime>2026-09-11T18:00:00</EndTime>
  <Type>all</Type>
</Query>`)
	root, err := ParseRoot(body)
	if err != nil {
		t.Fatal(err)
	}
	if root.CmdType != "RecordInfo" || root.SN != "12345" || root.DeviceID != "34020000001320000001" {
		t.Fatalf("unexpected root: %+v", root)
	}

	var req RecordInfoReq
	if err := Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req.StartTime != "2026-09-11T08:00:00" || req.EndTime != "2026-09-11T18:00:00" {
		t.Fatalf("unexpected query req: %+v", req)
	}
}

func TestBuildXMLGB2312(t *testing.T) {
	fields := map[string]any{
		"CmdType":          "Alarm",
		"SN":               "123",
		"DeviceID":         "34020000001320000001",
		"AlarmPriority":    "3",
		"AlarmMethod":      "5",
		"AlarmTime":        "2026-09-13T00:00:00",
		"AlarmDescription": "通道1画面移动侦测报警",
		"Info":             "<AlarmType>2</AlarmType>",
	}
	xmlBytes, err := BuildXML("Notify", fields, "GB2312")
	if err != nil {
		t.Fatalf("BuildXML failed: %v", err)
	}

	// 验证可以被 unmarshalXML (结合 GBK/GB2312 decoder) 正常解析
	var notify AlarmNotify
	if err := Unmarshal(xmlBytes, &notify); err != nil {
		t.Fatalf("Unmarshal GB2312 XML failed: %v", err)
	}
	if notify.CmdType != "Alarm" || notify.AlarmMethod != "5" || notify.Info.AlarmType != "2" {
		t.Fatalf("parsed unexpected notify: %+v", notify)
	}
	if notify.AlarmDescription != "通道1画面移动侦测报警" {
		t.Fatalf("expected Chinese description, got: %s", notify.AlarmDescription)
	}
}

func TestDeviceControlReqParsingLenient(t *testing.T) {
	// 测试包含 Info 嵌套和带空格的 DeviceControl
	body := []byte(`<?xml version="1.0" encoding="gb2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>9988</SN>
  <DeviceID> 34020000001320000001 </DeviceID>
  <GuardCmd> SetGuard </GuardCmd>
  <Info>
    <AlarmCmd> ResetAlarm </AlarmCmd>
  </Info>
</Control>`)
	var ctrl DeviceControlReq
	if err := Unmarshal(body, &ctrl); err != nil {
		t.Fatalf("Unmarshal DeviceControlReq failed: %v", err)
	}
	if ctrl.CmdType != "DeviceControl" {
		t.Fatalf("expected DeviceControl, got %s", ctrl.CmdType)
	}
	if ctrl.Info.AlarmCmd != " ResetAlarm " {
		t.Fatalf("expected Info.AlarmCmd, got %s", ctrl.Info.AlarmCmd)
	}
}

func TestConfigDownloadAndDeviceConfigParsing(t *testing.T) {
	// 1. 测试 ConfigDownload 基本参数查询
	queryXML := []byte(`<?xml version="1.0" encoding="gb2312"?>
<Query>
  <CmdType>ConfigDownload</CmdType>
  <SN>123456</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <ConfigType>BasicParam</ConfigType>
</Query>`)

	var q ConfigDownloadReq
	if err := Unmarshal(queryXML, &q); err != nil {
		t.Fatalf("Unmarshal ConfigDownload failed: %v", err)
	}
	if q.CmdType != "ConfigDownload" || q.ConfigType != "BasicParam" || q.DeviceID != "34020000001320000001" {
		t.Fatalf("unexpected query content: %+v", q)
	}

	// 2. 测试 DeviceConfig 下发与校时 (使用 EncodeXML 转换为真实 GB2312 编码)
	rawConfigXML := `<?xml version="1.0" encoding="GB2312"?>
<Control>
  <CmdType>DeviceConfig</CmdType>
  <SN>887766</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <BasicParam>
    <Name>东门主路球机</Name>
    <HeartBeatInterval>30</HeartBeatInterval>
    <HeartBeatCount>5</HeartBeatCount>
  </BasicParam>
  <Time>2026-09-13T12:00:00</Time>
</Control>`
	configXML, err := EncodeXML(rawConfigXML, "GB2312")
	if err != nil {
		t.Fatalf("EncodeXML failed: %v", err)
	}

	var cfg DeviceConfigReq
	if err := Unmarshal(configXML, &cfg); err != nil {
		t.Fatalf("Unmarshal DeviceConfig failed: %v", err)
	}
	if cfg.CmdType != "DeviceConfig" || cfg.BasicParam == nil || cfg.BasicParam.Name != "东门主路球机" {
		t.Fatalf("unexpected DeviceConfig: %+v", cfg)
	}
	if cfg.BasicParam.HeartBeatInterval != 30 || cfg.Time != "2026-09-13T12:00:00" {
		t.Fatalf("unexpected params: interval=%d, time=%s", cfg.BasicParam.HeartBeatInterval, cfg.Time)
	}

	// 3. 测试高级 DeviceControl (IFrame, Record, HomePosition)
	ctrlXML := []byte(`<?xml version="1.0" encoding="gb2312"?>
<Control>
  <CmdType>DeviceControl</CmdType>
  <SN>554433</SN>
  <DeviceID>34020000001320000001</DeviceID>
  <IFameCmd>Send</IFameCmd>
  <RecordCmd>Record</RecordCmd>
  <HomePosition>
    <Enabled>1</Enabled>
    <ResetTime>20</ResetTime>
    <PresetIndex>2</PresetIndex>
  </HomePosition>
</Control>`)

	var ctrl DeviceControlReq
	if err := Unmarshal(ctrlXML, &ctrl); err != nil {
		t.Fatalf("Unmarshal advanced DeviceControl failed: %v", err)
	}
	if !ctrl.IsForceIFrame() {
		t.Fatalf("expected IsForceIFrame to be true")
	}
	if ctrl.RecordCmd != "Record" {
		t.Fatalf("expected RecordCmd Record, got %s", ctrl.RecordCmd)
	}
	if ctrl.HomePosition == nil || ctrl.HomePosition.Enabled != "1" || ctrl.HomePosition.ResetTime != 20 || ctrl.HomePosition.PresetIndex != 2 {
		t.Fatalf("unexpected HomePosition: %+v", ctrl.HomePosition)
	}
}



