package device

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/local/gb28181-device/internal/gb28181"
)

// AlarmEventRecord 单条报警历史记录
type AlarmEventRecord struct {
	ID          string  `json:"id"`
	Time        string  `json:"time"`
	ChannelID   string  `json:"channelId"`
	ChannelName string  `json:"channelName"`
	AlarmMethod string  `json:"alarmMethod"` // 1:电话, 2:设备, 3:短信, 4:GPS, 5:视频, 6:设备故障
	AlarmType   string  `json:"alarmType"`   // 2:移动侦测, 6:区域入侵, 11:视频遮挡, 1:视频丢失, etc.
	Priority    string  `json:"priority"`    // 1:一级 2:二级 3:三级 4:四级
	Description string  `json:"description"`
	Longitude   float64 `json:"longitude,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Status      string  `json:"status"` // "confirmed", "sent", "failed"
	LatencyMs   int64   `json:"latencyMs"`
}

// AlarmManager 管理报警历史记录、设备/通道布防状态及自动化报警发生器
type AlarmManager struct {
	mu          sync.RWMutex
	guardStatus map[string]string // key: channelID ("" 或 "root" 表示设备根节点), value: "SetGuard" | "ResetGuard"
	alarming    map[string]bool   // key: channelID, value: 是否处于活跃报警状态
	records     []AlarmEventRecord
	maxRecords  int

	autoRunning  bool
	autoStopCh   chan struct{}
	autoInterval time.Duration
	autoChannels []string
}

// NewAlarmManager 创建报警管理器
func NewAlarmManager() *AlarmManager {
	return &AlarmManager{
		guardStatus: make(map[string]string),
		alarming:    make(map[string]bool),
		records:     make([]AlarmEventRecord, 0, 100),
		maxRecords:  100,
	}
}

// SetGuard 设置指定通道或设备根节点的布防状态。
// 若提供 allChannels 且 channelID 为根节点或空，则级联更新所有子通道。
func (am *AlarmManager) SetGuard(channelID string, isGuard bool, allChannels ...[]string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	st := "ResetGuard"
	if isGuard {
		st = "SetGuard"
	}
	am.guardStatus[channelID] = st
	if channelID == "" {
		am.guardStatus["root"] = st
	}
	if (channelID == "" || channelID == "root") && len(allChannels) > 0 {
		for _, ch := range allChannels[0] {
			am.guardStatus[ch] = st
		}
	}
}

// GetGuard 获取指定通道的布防状态；若通道未单独设置，则回退继承全局/根设备状态（默认为 "ResetGuard" 撤防）
func (am *AlarmManager) GetGuard(channelID string) string {
	am.mu.RLock()
	defer am.mu.RUnlock()
	if st, ok := am.guardStatus[channelID]; ok && st != "" {
		return st
	}
	// 回退继承全局
	if channelID != "" && channelID != "root" {
		if st, ok := am.guardStatus[""]; ok && st != "" {
			return st
		}
		if st, ok := am.guardStatus["root"]; ok && st != "" {
			return st
		}
	}
	return "ResetGuard"
}

// SetAlarming 设置指定通道的报警激活状态
func (am *AlarmManager) SetAlarming(channelID string, active bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if active {
		am.alarming[channelID] = true
	} else {
		delete(am.alarming, channelID)
	}
}

// IsAlarming 获取指定通道是否正在报警
func (am *AlarmManager) IsAlarming(channelID string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.alarming[channelID]
}

// ResetAlarm 复位指定通道或全部通道的报警激活状态（国标附录 A.5.1.5）
func (am *AlarmManager) ResetAlarm(channelID string, allChannels ...[]string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if channelID == "" || channelID == "root" {
		am.alarming = make(map[string]bool)
		log.Printf("[alarm] reset all active alarms for device")
	} else {
		delete(am.alarming, channelID)
		log.Printf("[alarm] reset active alarm for channel=%s", channelID)
	}
	if len(allChannels) > 0 && (channelID == "" || channelID == "root") {
		for _, ch := range allChannels[0] {
			delete(am.alarming, ch)
		}
	}
}

// DutyStatus 获取指定通道或设备的国标防区状态（GB/T 28181 表 A.4）：
// - ALARM: 正在报警
// - ONDUTY: 已布防 (在防)
// - OFFDUTY: 已撤防 (撤防)
func (am *AlarmManager) DutyStatus(channelID string) string {
	am.mu.RLock()
	defer am.mu.RUnlock()
	if am.alarming[channelID] {
		return "ALARM"
	}
	st := "ResetGuard"
	if s, ok := am.guardStatus[channelID]; ok && s != "" {
		st = s
	} else if channelID != "" && channelID != "root" {
		if s, ok := am.guardStatus[""]; ok && s != "" {
			st = s
		} else if s, ok := am.guardStatus["root"]; ok && s != "" {
			st = s
		}
	}
	if st == "SetGuard" {
		return "ONDUTY"
	}
	return "OFFDUTY"
}

// ClearRecords 清空报警记录
func (am *AlarmManager) ClearRecords() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.records = make([]AlarmEventRecord, 0, 100)
}

// AddRecord 添加一条报警记录（环形队列保存最新记录）
func (am *AlarmManager) AddRecord(rec AlarmEventRecord) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.records = append(am.records, rec)
	if len(am.records) > am.maxRecords {
		am.records = am.records[len(am.records)-am.maxRecords:]
	}
}

// ListRecords 获取最近的报警记录列表（按时间倒序）
func (am *AlarmManager) ListRecords(limit int) []AlarmEventRecord {
	am.mu.RLock()
	defer am.mu.RUnlock()
	n := len(am.records)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]AlarmEventRecord, limit)
	for i := 0; i < limit; i++ {
		out[i] = am.records[n-1-i]
	}
	return out
}

// IsAutoAlarmRunning 判断是否正在运行自动周期报警模拟
func (am *AlarmManager) IsAutoAlarmRunning() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.autoRunning
}

// StartAutoAlarm 启动自动周期报警模拟（仅向处于布防状态的通道触发）
func (am *AlarmManager) StartAutoAlarm(interval time.Duration, channels []string, triggerFn func(channelID string)) {
	am.mu.Lock()
	if am.autoRunning {
		am.mu.Unlock()
		return
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}
	am.autoRunning = true
	am.autoInterval = interval
	am.autoChannels = channels
	am.autoStopCh = make(chan struct{})
	stopCh := am.autoStopCh
	am.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				am.mu.RLock()
				chs := am.autoChannels
				am.mu.RUnlock()
				// 自动报警仅向已布防通道触发
				armedChs := make([]string, 0, len(chs))
				for _, ch := range chs {
					if am.GetGuard(ch) == "SetGuard" {
						armedChs = append(armedChs, ch)
					}
				}
				if len(armedChs) > 0 {
					targetCh := armedChs[time.Now().UnixNano()%int64(len(armedChs))]
					triggerFn(targetCh)
				}
			}
		}
	}()
}

// StopAutoAlarm 停止自动周期报警模拟
func (am *AlarmManager) StopAutoAlarm() {
	am.mu.Lock()
	defer am.mu.Unlock()
	if !am.autoRunning {
		return
	}
	am.autoRunning = false
	if am.autoStopCh != nil {
		close(am.autoStopCh)
		am.autoStopCh = nil
	}
}

// SendAlarm 发送移动侦测报警（向后兼容简易方法）
func (d *Device) SendAlarm(channelID, description string) error {
	_, err := d.SendAlarmEvent(channelID, "5", "2", "3", description, 0, 0)
	return err
}

// SendAlarmAdvanced 发送指定类型与级别的国标报警（向后兼容方法）
func (d *Device) SendAlarmAdvanced(channelID, alarmMethod, priority, description string) error {
	alarmType := "2"
	if alarmMethod == "2" {
		alarmMethod = "5" // 映射为标准 GB28181 视频报警
	}
	_, err := d.SendAlarmEvent(channelID, alarmMethod, alarmType, priority, description, 0, 0)
	return err
}

// SendAlarmEvent 发送符合 GB/T 28181-2016 附录 D 表 D.2 规范的完整报警通知。
// 联动机制：
// 1. 检查通道布防状态：若通道处于撤防状态（ResetGuard/OFFDUTY）且非紧急警情且未指定 force，则按照国标布撤防联动机制拦截上报并记录为 suppressed；
// 2. 24小时特种/紧急警情（Priority==1 或 SOS/紧急求助）在撤防状态下具有豁免权，可正常上报；
// 3. 上报成功后，联动将通道标记为 ALARM (正在报警) 状态。
func (d *Device) SendAlarmEvent(channelID, alarmMethod, alarmType, priority, description string, lon, lat float64, force ...bool) (*AlarmEventRecord, error) {
	if channelID == "" {
		channelID = d.cfg.Device.ID
	}
	if alarmMethod == "" {
		alarmMethod = "5" // 默认：视频报警
	}
	if alarmType == "" {
		alarmType = "2" // 默认：移动侦测
	}
	if priority == "" {
		priority = "3" // 默认：三级
	}

	channelName := d.cfg.Device.Name
	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == channelID {
			channelName = ch.Name
			break
		}
	}

	if description == "" {
		description = getDefaultAlarmDescription(alarmMethod, alarmType, channelName)
	}

	nowTime := time.Now()
	nowStr := nowTime.Format("2006-01-02T15:04:05")

	// 生成随机唯一记录 ID
	var rb [4]byte
	_, _ = rand.Read(rb[:])
	recID := fmt.Sprintf("ALM-%s-%s", nowTime.Format("150405"), hex.EncodeToString(rb[:]))

	isForce := len(force) > 0 && force[0]
	// 24小时紧急防区判定：1级警情（最高紧急）或 SOS/紧急求助按钮
	isEmergency := priority == "1" || (alarmMethod == "2" && (alarmType == "5" || alarmType == "0"))

	// 检查通道当前的布撤防状态
	guardSt := "ResetGuard"
	if d.alarmMgr != nil {
		guardSt = d.alarmMgr.GetGuard(channelID)
	}

	// 联动门禁检查：撤防状态下拦截常规报警
	if guardSt == "ResetGuard" && !isEmergency && !isForce {
		errMsg := fmt.Sprintf("通道 [%s] 处于撤防状态 (ResetGuard/OFFDUTY)，常规报警已按国标联动机制拦截上报", channelID)
		log.Printf("[alarm] %s: method=%s type=%s desc=%s", errMsg, alarmMethod, alarmType, description)
		rec := AlarmEventRecord{
			ID:          recID,
			Time:        nowStr,
			ChannelID:   channelID,
			ChannelName: channelName,
			AlarmMethod: alarmMethod,
			AlarmType:   alarmType,
			Priority:    priority,
			Description: description + " (撤防拦截)",
			Longitude:   lon,
			Latitude:    lat,
			Status:      "suppressed",
			LatencyMs:   0,
		}
		if d.alarmMgr != nil {
			d.alarmMgr.AddRecord(rec)
		}
		return &rec, fmt.Errorf("%s", errMsg)
	}

	// 构造国标标准 XML
	fields := map[string]any{
		"CmdType":          "Alarm",
		"SN":               d.nextSN(),
		"DeviceID":         channelID,
		"AlarmPriority":    priority,
		"AlarmMethod":      alarmMethod,
		"AlarmTime":        nowStr,
		"AlarmDescription": description,
	}
	if lon != 0 || lat != 0 {
		fields["Longitude"] = fmt.Sprintf("%.6f", lon)
		fields["Latitude"] = fmt.Sprintf("%.6f", lat)
	}
	if alarmType != "" {
		fields["Info"] = fmt.Sprintf("<AlarmType>%s</AlarmType>", alarmType)
	}

	body := d.buildXMLNotify(fields)
	t0 := time.Now()
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	latMs := time.Since(t0).Milliseconds()

	status := "confirmed"
	if err != nil {
		status = "failed"
		log.Printf("[gb] alarm notify failed ch=%s: %v", channelID, err)
	} else {
		log.Printf("[gb] alarm notify 200 OK ch=%s method=%s type=%s priority=%s desc=%s (%dms)",
			channelID, alarmMethod, alarmType, priority, description, latMs)
		// 联动将通道防区状态切换为报警激活 (ALARM)
		if d.alarmMgr != nil {
			d.alarmMgr.SetAlarming(channelID, true)
		}
	}

	rec := AlarmEventRecord{
		ID:          recID,
		Time:        nowStr,
		ChannelID:   channelID,
		ChannelName: channelName,
		AlarmMethod: alarmMethod,
		AlarmType:   alarmType,
		Priority:    priority,
		Description: description,
		Longitude:   lon,
		Latitude:    lat,
		Status:      status,
		LatencyMs:   latMs,
	}

	if d.alarmMgr != nil {
		d.alarmMgr.AddRecord(rec)
	}
	return &rec, err
}

func getDefaultAlarmDescription(method, alarmType, name string) string {
	if method == "5" {
		switch alarmType {
		case "1":
			return fmt.Sprintf("[%s] 触发视频丢失报警 (Video Loss)", name)
		case "2":
			return fmt.Sprintf("[%s] 触发运动目标检测 (移动侦测)", name)
		case "5":
			return fmt.Sprintf("[%s] 触发警戒绊线越界报警", name)
		case "6":
			return fmt.Sprintf("[%s] 触发防区周界入侵报警", name)
		case "11":
			return fmt.Sprintf("[%s] 触发视频遮挡/镜头篡改报警", name)
		case "51":
			return fmt.Sprintf("[%s] 触发违章停车/违规停留报警", name)
		default:
			return fmt.Sprintf("[%s] 触发视频智能分析报警", name)
		}
	} else if method == "2" {
		switch alarmType {
		case "1":
			return fmt.Sprintf("[%s] 触发门禁/门磁开关报警", name)
		case "2":
			return fmt.Sprintf("[%s] 触发红外对射/防盗探测器报警", name)
		case "3":
			return fmt.Sprintf("[%s] 触发烟雾火警探测器报警", name)
		case "5", "0":
			return fmt.Sprintf("[%s] 触发人工紧急求助/报警按钮 (SOS)", name)
		default:
			return fmt.Sprintf("[%s] 触发设备开关量/探头防区报警", name)
		}
	} else if method == "6" {
		switch alarmType {
		case "21":
			return fmt.Sprintf("[%s] 触发存储介质故障/硬盘满报警", name)
		case "22":
			return fmt.Sprintf("[%s] 触发网络断开/通信链路故障报警", name)
		case "23":
			return fmt.Sprintf("[%s] 触发电源掉电/供电异常故障报警", name)
		default:
			return fmt.Sprintf("[%s] 触发设备硬件故障报警", name)
		}
	}
	return fmt.Sprintf("[%s] 触发国标模拟报警", name)
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

	body := d.buildXMLNotify(map[string]any{
		"CmdType":    "Catalog",
		"SN":         d.nextSN(),
		"DeviceID":   d.cfg.Device.ID,
		"SumNum":     len(items),
		"DeviceList": listB,
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	return err
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
