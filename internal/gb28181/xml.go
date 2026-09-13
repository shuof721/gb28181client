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
		// Fallback: 如果初始解码失败（比如平台发来的 GB2312 报文未带 xml 声明导致 Go 默认按 utf-8 报 invalid UTF-8），
		// 则尝试先整体将 GB18030 转为 UTF-8 后再次解码
		reader := transform.NewReader(bytes.NewReader(body), simplifiedchinese.GB18030.NewDecoder())
		utf8Body, rerr := io.ReadAll(reader)
		if rerr == nil {
			dec2 := xml.NewDecoder(bytes.NewReader(utf8Body))
			dec2.CharsetReader = charsetReader
			if err2 := dec2.Decode(v); err2 == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

// EncodeXML 将 UTF-8 字符串转换为目标字符集的字节切片。
// 若 charset 为 GB2312/GBK/GB18030，则使用 GB18030 编码（兼容 GB2312 与 GBK 全部汉字与字符）；
// 若 charset 为 UTF-8，则返回 UTF-8 字节。
func EncodeXML(xmlStr string, charset string) ([]byte, error) {
	cs := strings.ToUpper(strings.TrimSpace(charset))
	if cs == "" || cs == "GB2312" || cs == "GBK" || cs == "GB18030" {
		reader := transform.NewReader(strings.NewReader(xmlStr), simplifiedchinese.GB18030.NewEncoder())
		return io.ReadAll(reader)
	}
	return []byte(xmlStr), nil
}

// BuildXML 统一构造符合 GB/T 28181 规范的 XML 报文（标准头部声明、字段顺序与字符编码）。
// rootTag: "Response", "Notify", "Control", "Query"
func BuildXML(rootTag string, fields map[string]any, charset string) ([]byte, error) {
	cs := strings.ToUpper(strings.TrimSpace(charset))
	if cs == "" {
		cs = "GB2312"
	}
	encHeader := cs
	if cs == "GBK" || cs == "GB18030" {
		encHeader = "GB2312" // 国标规范通常要求声明 GB2312
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<?xml version=\"1.0\" encoding=\"%s\"?>\r\n", encHeader)
	fmt.Fprintf(&b, "<%s>\r\n", rootTag)

	// GB28181 标准字段排序，保证各平台 XML 解析器顺序一致
	order := []string{
		"CmdType", "SN", "DeviceID", "Result",
		"Time",
		"AlarmPriority", "AlarmMethod", "AlarmTime", "AlarmDescription",
		"Longitude", "Latitude", "Speed", "Direction", "Altitude",
		"SumNum", "DeviceName", "Manufacturer", "Model", "Firmware",
		"Channel", "Online", "Status", "Encode", "Record", "DeviceTime",
		"NotifyType",
	}
	written := map[string]bool{}
	for _, k := range order {
		if v, ok := fields[k]; ok {
			fmt.Fprintf(&b, "  <%s>%v</%s>\r\n", k, v, k)
			written[k] = true
		}
	}
	for k, v := range fields {
		if written[k] || k == "DeviceList" || k == "RecordList" || k == "Alarmstatus" || k == "Info" {
			continue
		}
		fmt.Fprintf(&b, "  <%s>%v</%s>\r\n", k, v, k)
	}
	if info, ok := fields["Info"].(string); ok && strings.TrimSpace(info) != "" {
		trimmed := strings.TrimSpace(info)
		if strings.HasPrefix(trimmed, "<Info>") && strings.HasSuffix(trimmed, "</Info>") {
			b.WriteString("  " + trimmed + "\r\n")
		} else {
			b.WriteString(fmt.Sprintf("  <Info>\r\n    %s\r\n  </Info>\r\n", trimmed))
		}
	}
	if as, ok := fields["Alarmstatus"].(string); ok && as != "" {
		b.WriteString(as)
	}
	if dl, ok := fields["DeviceList"].(string); ok && dl != "" {
		b.WriteString(dl)
	}
	if rl, ok := fields["RecordList"].(string); ok && rl != "" {
		b.WriteString(rl)
	}
	fmt.Fprintf(&b, "</%s>\r\n", rootTag)

	return EncodeXML(b.String(), cs)
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

type HomePositionParam struct {
	Enabled           string `xml:"Enabled,omitempty"`           // 1 / 0
	HomePositionReset string `xml:"HomePositionReset,omitempty"` // 1 / 0 (GB28181 标准别名)
	ResetTime         int    `xml:"ResetTime,omitempty"`         // 归位时间(秒)
	PresetIndex       int    `xml:"PresetIndex,omitempty"`       // 预置位编号 0~255
}

type DeviceControlReq struct {
	XMLName      xml.Name           // 宽松匹配任何根标签 (Control / Notify 等)
	CmdType      string             `xml:"CmdType"`
	SN           string             `xml:"SN"`
	DeviceID     string             `xml:"DeviceID"`
	PTZCmd       string             `xml:"PTZCmd,omitempty"`
	DragZoomIn   *DragZoomParam     `xml:"DragZoomIn,omitempty"`
	DragZoomOut  *DragZoomParam     `xml:"DragZoomOut,omitempty"`
	GuardCmd     string             `xml:"GuardCmd,omitempty"`     // SetGuard (布防) | ResetGuard (撤防)
	AlarmCmd     string             `xml:"AlarmCmd,omitempty"`     // ResetAlarm (报警复位)
	TeleBoot     string             `xml:"TeleBoot,omitempty"`     // Boot (远程重启)
	RecordCmd    string             `xml:"RecordCmd,omitempty"`    // Record / StopRecord
	IFrameCmd    string             `xml:"IFrameCmd,omitempty"`    // Send / 1 (强制关键帧)
	IFameCmd     string             `xml:"IFameCmd,omitempty"`     // WVP 历史拼写兼容
	IFCDCmd      string             `xml:"IFCDCmd,omitempty"`      // 国标别名
	HomePosition *HomePositionParam `xml:"HomePosition,omitempty"` // 看守位
	Info         struct {
		ControlPriority string `xml:"ControlPriority,omitempty"`
		AlarmMethod     string `xml:"AlarmMethod,omitempty"`
		AlarmType       string `xml:"AlarmType,omitempty"`
		GuardCmd        string `xml:"GuardCmd,omitempty"`
		AlarmCmd        string `xml:"AlarmCmd,omitempty"`
		IFrameCmd       string `xml:"IFrameCmd,omitempty"`
		RecordCmd       string `xml:"RecordCmd,omitempty"`
	} `xml:"Info,omitempty"`
	// 雨刷/灯光等可扩展
}

// IsForceIFrame 判断是否为强制关键帧指令 (兼顾各厂商和平台不同命名)
func (r *DeviceControlReq) IsForceIFrame() bool {
	return r.IFrameCmd != "" || r.IFameCmd != "" || r.IFCDCmd != "" || r.Info.IFrameCmd != ""
}

// ----- ConfigDownload (配置查询) -----

type ConfigDownloadReq struct {
	XMLName    xml.Name `xml:"Query"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	ConfigType string   `xml:"ConfigType"` // BasicParam, VideoParamOpt, AudioParamOpt, SVACEncodeConfig 等
}

type BasicParamConfig struct {
	Name               string  `xml:"Name,omitempty"`
	Expiration         string  `xml:"Expiration,omitempty"`
	HeartBeatInterval  int     `xml:"HeartBeatInterval,omitempty"`
	HeartBeatCount     int     `xml:"HeartBeatCount,omitempty"`
	PositionCapability int     `xml:"PositionCapability,omitempty"` // 0:不支持, 1:GPS, 2:北斗
	Longitude          float64 `xml:"Longitude,omitempty"`
	Latitude           float64 `xml:"Latitude,omitempty"`
}

type VideoParamOptConfig struct {
	DownloadSpeed string `xml:"DownloadSpeed,omitempty"` // 各可选参数以 '/' 分隔，如 "1/2/4"
	Resolution    string `xml:"Resolution,omitempty"`    // 如 "1920*1080/1280*720/704*576"
}

type AudioParamOptConfig struct {
	AudioFormat  string `xml:"AudioFormat,omitempty"`  // 如 "G.711A/G.711U/AAC"
	SamplingRate string `xml:"SamplingRate,omitempty"` // 如 "8/16/32"
}

// ----- DeviceConfig (配置下发与校时) -----

type DeviceConfigReq struct {
	XMLName       xml.Name             // 宽松匹配 Control / Query 等
	CmdType       string               `xml:"CmdType"`
	SN            string               `xml:"SN"`
	DeviceID      string               `xml:"DeviceID"`
	BasicParam    *BasicParamConfig    `xml:"BasicParam,omitempty"`
	VideoParamOpt *VideoParamOptConfig `xml:"VideoParamOpt,omitempty"`
	Time          string               `xml:"Time,omitempty"` // 2026-09-13T11:40:00 或 11:40:00
	Date          string               `xml:"Date,omitempty"` // 2026-09-13
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

// ----- MobilePosition (移动设备位置数据通知与订阅) -----

type MobilePositionQueryReq struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Interval int      `xml:"Interval,omitempty"`
}

type MobilePositionNotify struct {
	XMLName   xml.Name `xml:"Notify"`
	CmdType   string   `xml:"CmdType"`
	SN        string   `xml:"SN"`
	DeviceID  string   `xml:"DeviceID"`
	Time      string   `xml:"Time"`
	Longitude float64  `xml:"Longitude"`
	Latitude  float64  `xml:"Latitude"`
	Speed     float64  `xml:"Speed,omitempty"`
	Direction float64  `xml:"Direction,omitempty"`
	Altitude  float64  `xml:"Altitude,omitempty"`
}

// ----- CatalogNotify (目录订阅增量通知) -----

type CatalogNotifyItem struct {
	DeviceID     string `xml:"DeviceID"`
	Event        string `xml:"Event"` // ON, OFF, VLOST, DEFECT, ADD, DEL, UPDATE
	Name         string `xml:"Name,omitempty"`
	Manufacturer string `xml:"Manufacturer,omitempty"`
	Model        string `xml:"Model,omitempty"`
	Owner        string `xml:"Owner,omitempty"`
	CivilCode    string `xml:"CivilCode,omitempty"`
	Address      string `xml:"Address,omitempty"`
	Parental     int    `xml:"Parental,omitempty"`
	ParentID     string `xml:"ParentID,omitempty"`
	SafetyWay    int    `xml:"SafetyWay,omitempty"`
	RegisterWay  int    `xml:"RegisterWay,omitempty"`
	Secrecy      int    `xml:"Secrecy,omitempty"`
	Status       string `xml:"Status,omitempty"`
	PTZType      int    `xml:"PTZType,omitempty"`
}

type CatalogNotify struct {
	XMLName    xml.Name `xml:"Notify"`
	CmdType    string   `xml:"CmdType"`
	SN         string   `xml:"SN"`
	DeviceID   string   `xml:"DeviceID"`
	SumNum     int      `xml:"SumNum"`
	DeviceList struct {
		Num   int                 `xml:"Num,attr"`
		Items []CatalogNotifyItem `xml:"Item"`
	} `xml:"DeviceList"`
}

// BuildCatalogNotifyXML 构造标准目录增量通知报文
func BuildCatalogNotifyXML(sn, rootDeviceID string, items []CatalogNotifyItem, charset string) ([]byte, error) {
	var listB strings.Builder
	fmt.Fprintf(&listB, "  <DeviceList Num=\"%d\">\r\n", len(items))
	for _, it := range items {
		listB.WriteString("    <Item>\r\n")
		fmt.Fprintf(&listB, "      <DeviceID>%s</DeviceID>\r\n", it.DeviceID)
		fmt.Fprintf(&listB, "      <Event>%s</Event>\r\n", it.Event)
		if it.Name != "" {
			fmt.Fprintf(&listB, "      <Name>%s</Name>\r\n", it.Name)
		}
		if it.Manufacturer != "" {
			fmt.Fprintf(&listB, "      <Manufacturer>%s</Manufacturer>\r\n", it.Manufacturer)
		}
		if it.Model != "" {
			fmt.Fprintf(&listB, "      <Model>%s</Model>\r\n", it.Model)
		}
		if it.Owner != "" {
			fmt.Fprintf(&listB, "      <Owner>%s</Owner>\r\n", it.Owner)
		}
		if it.CivilCode != "" {
			fmt.Fprintf(&listB, "      <CivilCode>%s</CivilCode>\r\n", it.CivilCode)
		}
		if it.Address != "" {
			fmt.Fprintf(&listB, "      <Address>%s</Address>\r\n", it.Address)
		}
		if it.ParentID != "" {
			fmt.Fprintf(&listB, "      <ParentID>%s</ParentID>\r\n", it.ParentID)
		}
		if it.Status != "" {
			fmt.Fprintf(&listB, "      <Status>%s</Status>\r\n", it.Status)
		}
		if it.RegisterWay > 0 {
			fmt.Fprintf(&listB, "      <RegisterWay>%d</RegisterWay>\r\n", it.RegisterWay)
		}
		if it.PTZType > 0 {
			fmt.Fprintf(&listB, "      <PTZType>%d</PTZType>\r\n", it.PTZType)
		}
		listB.WriteString("    </Item>\r\n")
	}
	listB.WriteString("  </DeviceList>\r\n")

	return BuildXML("Notify", map[string]any{
		"CmdType":    "Catalog",
		"SN":         sn,
		"DeviceID":   rootDeviceID,
		"SumNum":     len(items),
		"DeviceList": listB.String(),
	}, charset)
}

// ----- DeviceConfig 等可按需扩展 -----

func MarshalXML(v any) ([]byte, error) {
	return MarshalXMLWithCharset(v, "GB2312")
}

func MarshalXMLWithCharset(v any, charset string) ([]byte, error) {
	cs := strings.ToUpper(strings.TrimSpace(charset))
	if cs == "" {
		cs = "GB2312"
	}
	out, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	encHeader := cs
	if cs == "GBK" || cs == "GB18030" {
		encHeader = "GB2312"
	}
	header := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"%s\"?>\r\n", encHeader)
	fullXML := header + string(out) + "\r\n"
	return EncodeXML(fullXML, cs)
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
