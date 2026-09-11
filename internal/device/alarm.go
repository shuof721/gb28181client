package device

import (
	"log"
	"time"

	"github.com/local/gb28181-device/internal/gb28181"
)

// SendAlarm 发送移动侦测/报警通知（Notify/Alarm）。
func (d *Device) SendAlarm(channelID, description string) error {
	return d.SendAlarmAdvanced(channelID, "2", "4", description)
}

// SendAlarmAdvanced 发送指定类型与级别的国标报警。
// alarmMethod: 1:人工视频报警 2:运动目标检测报警 3:遗留物检测报警 4:物体移除检测报警 5:绊线入侵检测报警 6:区域入侵检测报警...
// priority: 1:一级 2:二级 3:三级 4:四级
func (d *Device) SendAlarmAdvanced(channelID, alarmMethod, priority, description string) error {
	if channelID == "" {
		channelID = d.cfg.Device.ID
	}
	if alarmMethod == "" {
		alarmMethod = "2"
	}
	if priority == "" {
		priority = "4"
	}
	if description == "" {
		description = "模拟报警"
	}
	now := time.Now().Format("2006-01-02T15:04:05")

	// 手工拼 XML，避免结构体字段顺序问题
	body := buildXMLNotify(map[string]any{
		"CmdType":          "Alarm",
		"SN":               d.nextSN(),
		"DeviceID":         channelID,
		"AlarmPriority":    priority,
		"AlarmMethod":      alarmMethod,
		"AlarmTime":        now,
		"AlarmDescription": description,
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] alarm notify failed: %v", err)
		return err
	}
	log.Printf("[gb] alarm sent ch=%s method=%s priority=%s desc=%s", channelID, alarmMethod, priority, description)
	return nil
}

// SendCatalogNotify 主动上报目录变更（部分平台会 Subscribe）。
func (d *Device) SendCatalogNotify() error {
	items := d.cfg.Device.Channels
	var listB string
	listB = "  <DeviceList Num=\"" + itoa(len(items)) + "\">\r\n"
	for _, ch := range items {
		listB += "    <Item>\r\n"
		listB += "      <DeviceID>" + ch.ID + "</DeviceID>\r\n"
		listB += "      <Name>" + ch.Name + "</Name>\r\n"
		listB += "      <Manufacturer>" + ch.Manufacturer + "</Manufacturer>\r\n"
		listB += "      <Model>" + ch.Model + "</Model>\r\n"
		listB += "      <Owner>Owner</Owner>\r\n"
		listB += "      <Parental>0</Parental>\r\n"
		listB += "      <ParentID>" + ch.ParentID + "</ParentID>\r\n"
		listB += "      <SafetyWay>0</SafetyWay>\r\n"
		listB += "      <RegisterWay>1</RegisterWay>\r\n"
		listB += "      <Secrecy>0</Secrecy>\r\n"
		listB += "      <Status>" + ch.Status + "</Status>\r\n"
		listB += "    </Item>\r\n"
	}
	listB += "  </DeviceList>\r\n"

	body := buildXMLNotify(map[string]any{
		"CmdType":    "Catalog",
		"SN":         d.nextSN(),
		"DeviceID":   d.cfg.Device.ID,
		"SumNum":     len(items),
		"DeviceList": listB,
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	return err
}

func buildXMLNotify(fields map[string]any) []byte {
	// Notify 根与 Response 类似，只是标签名不同
	s := string(buildXMLResponse(fields))
	s = replaceAll(s, "<Response>", "<Notify>")
	s = replaceAll(s, "</Response>", "</Notify>")
	return []byte(s)
}

func replaceAll(s, old, new string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			out = append(out, new...)
			i += len(old)
			continue
		}
		out = append(out, s[i])
		i++
	}
	return string(out)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

var _ = gb28181.AlarmNotify{}
