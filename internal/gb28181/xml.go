package gb28181

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 通用 MANSCDP XML

type Root struct {
	XMLName  xml.Name
	CmdType  string `xml:"CmdType"`
	SN       string `xml:"SN"`
	DeviceID string `xml:"DeviceID"`
	SourceID string `xml:"SourceID"`
	TargetID string `xml:"TargetID"`
}

// charsetReader 支持国标设备常见的 GB2312/GBK/GB18030 声明。
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "", "utf-8", "utf8", "us-ascii", "ascii":
		return input, nil
	case "gb2312", "gb-2312", "gbk", "gb18030", "gb-18030":
		return transform.NewReader(input, simplifiedchinese.GBK.NewDecoder()), nil
	default:
		// 尝试按 GBK 解，很多设备 charset 写法不规范
		return transform.NewReader(input, simplifiedchinese.GBK.NewDecoder()), nil
	}
}

func unmarshalXML(body []byte, v any) error {
	dec := xml.NewDecoder(bytes.NewReader(body))
	dec.CharsetReader = charsetReader
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

func ParseRoot(body []byte) (*Root, error) {
	var r Root
	if err := unmarshalXML(body, &r); err != nil {
		return nil, fmt.Errorf("parse manscdp: %w", err)
	}
	return &r, nil
}

func Unmarshal(body []byte, v any) error {
	return unmarshalXML(body, v)
}

// ----- Catalog -----

type CatalogReq struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

type CatalogItem struct {
	DeviceID     string `xml:"DeviceID"`
	Name         string `xml:"Name"`
	Manufacturer string `xml:"Manufacturer"`
	Model        string `xml:"Model"`
	Owner        string `xml:"Owner"`
	CivilCode    string `xml:"CivilCode,omitempty"`
	Address      string `xml:"Address,omitempty"`
	Parental     int    `xml:"Parental"`
	ParentID     string `xml:"ParentID,omitempty"`
	SafetyWay    int    `xml:"SafetyWay"`
	RegisterWay  int    `xml:"RegisterWay"`
	Secrecy      int    `xml:"Secrecy"`
	Status       string `xml:"Status"`
	// 可选
	IPAddress    string `xml:"IPAddress,omitempty"`
	Port         int    `xml:"Port,omitempty"`
	Longitude    float64 `xml:"Longitude,omitempty"`
	Latitude     float64 `xml:"Latitude,omitempty"`
	PTZType      int     `xml:"PTZType,omitempty"`
}

type CatalogResp struct {
	XMLName    xml.Name `xml:"Response"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	SumNum     int      `xml:"SumNum"`
	DeviceList struct {
		Num  int          `xml:"Num,attr"`
		Items []CatalogItem `xml:"Item"`
	} `xml:"DeviceList"`
}

// ----- DeviceInfo -----

type DeviceInfoReq struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

type DeviceInfoResp struct {
	XMLName      xml.Name `xml:"Response"`
	CmdType      string   `xml:"CmdType"`
	SN           string   `xml:"SN"`
	DeviceID     string   `xml:"DeviceID"`
	DeviceName   string   `xml:"DeviceName"`
	Manufacturer string   `xml:"Manufacturer"`
	Model        string   `xml:"Model"`
	Firmware     string   `xml:"Firmware"`
	Channel      int      `xml:"Channel"`
}

// ----- DeviceStatus -----

type DeviceStatusReq struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

type DeviceStatusResp struct {
	XMLName    xml.Name `xml:"Response"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	Result     string   `xml:"Result"`
	Online     string   `xml:"Online"`
	Status     string   `xml:"Status"`
	Encode     string   `xml:"Encode"`
	Record     string   `xml:"Record"`
	DeviceTime string   `xml:"DeviceTime"`
}

// ----- DeviceControl (PTZ 等) -----

type DragZoomParam struct {
	Length    int `xml:"Length"`
	Width     int `xml:"Width"`
	MidPointX int `xml:"MidPointX"`
	MidPointY int `xml:"MidPointY"`
	LengthX   int `xml:"LengthX"`
	LengthY   int `xml:"LengthY"`
}

type DeviceControlReq struct {
	XMLName     xml.Name       `xml:"Control"`
	CmdType     string         `xml:"CmdType"`
	SN          string         `xml:"SN"`
	DeviceID    string         `xml:"DeviceID"`
	PTZCmd      string         `xml:"PTZCmd,omitempty"`
	DragZoomIn  *DragZoomParam `xml:"DragZoomIn,omitempty"`
	DragZoomOut *DragZoomParam `xml:"DragZoomOut,omitempty"`
	Info        struct {
		ControlPriority string `xml:"ControlPriority,omitempty"`
	} `xml:"Info,omitempty"`
	// 雨刷/灯光等可扩展
}

// ----- Keepalive -----

type KeepaliveNotify struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Status   string   `xml:"Status"`
}

// ----- Alarm -----

type AlarmNotify struct {
	XMLName    xml.Name `xml:"Notify"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	AlarmPriority string `xml:"AlarmPriority"`
	AlarmMethod string  `xml:"AlarmMethod"`
	AlarmTime  string   `xml:"AlarmTime"`
	AlarmDescription string `xml:"AlarmDescription"`
	Longitude  float64  `xml:"Longitude,omitempty"`
	Latitude   float64  `xml:"Latitude,omitempty"`
	Info       struct {
		AlarmType string `xml:"AlarmType"`
	} `xml:"Info,omitempty"`
}

// ----- RecordInfo -----

type RecordInfoReq struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Type     string   `xml:"Type"`
	StartTime string  `xml:"StartTime"`
	EndTime  string   `xml:"EndTime"`
	Secrecy  int      `xml:"Secrecy"`
}

type RecordItem struct {
	DeviceID  string `xml:"DeviceID"`
	Name      string `xml:"Name"`
	FilePath  string `xml:"FilePath,omitempty"`
	Address   string `xml:"Address,omitempty"`
	StartTime string `xml:"StartTime"`
	EndTime   string `xml:"EndTime"`
	Secrecy   int    `xml:"Secrecy"`
	Type      string `xml:"Type"`
	RecorderID string `xml:"RecorderID,omitempty"`
	FileSize  string `xml:"FileSize,omitempty"`
}

type RecordInfoResp struct {
	XMLName  xml.Name `xml:"Response"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Name     string   `xml:"Name,omitempty"`
	SumNum   int      `xml:"SumNum"`
	RecordList struct {
		Num   int          `xml:"Num,attr"`
		Items []RecordItem `xml:"Item"`
	} `xml:"RecordList"`
}

// ----- MediaStatus -----

type MediaStatusNotify struct {
	XMLName    xml.Name `xml:"Notify"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	NotifyType string   `xml:"NotifyType"` // 121: 历史媒体发送结束, 120: 媒体流已准备就绪
}

// ----- PresetQuery -----

type PresetQueryReq struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

type PresetItem struct {
	PresetID   string `xml:"PresetID"`
	PresetName string `xml:"PresetName"`
}

type PresetQueryResp struct {
	XMLName    xml.Name `xml:"Response"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	PresetList struct {
		Num   int          `xml:"Num,attr"`
		Items []PresetItem `xml:"Item"`
	} `xml:"PresetList"`
}

// ----- DeviceConfig 等可按需扩展 -----

func MarshalXML(v any) ([]byte, error) {
	out, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	header := []byte(xml.Header)
	return append(header, out...), nil
}

// DecodePTZ 解析 GB28181 PTZCmd（A5 0F 01 ... 十六进制字符串）。
func DecodePTZ(cmd string) string {
	parsed, err := ParsePTZCmd(cmd)
	if err != nil {
		return "unknown"
	}
	if parsed.Description != "" {
		return parsed.Description
	}
	return string(parsed.Action)
}

func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("odd hex length")
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(out); i++ {
		v, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, err
		}
		out[i] = byte(v)
	}
	return out, nil
}
