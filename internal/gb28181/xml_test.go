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

