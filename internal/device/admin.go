package device

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/local/gb28181-device/internal/config"
)

// AssetsDir 视频库目录（与默认 mp4 路径约定一致）。
const AssetsDir = "assets"

// VideoItem Web 视频库条目。
type VideoItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// AddChannelRequest 新增通道。
type AddChannelRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// 可选绑定视频（相对或相对 assets）
	MP4 string `json:"mp4"`
}

// BindChannelRequest 绑定通道媒体。
type BindChannelRequest struct {
	ChannelID    string `json:"channelId"`
	Source       string `json:"source"` // mp4 | file | synthetic
	MP4          string `json:"mp4"`
	H264         string `json:"h264"`
	AudioEnabled *bool  `json:"audioEnabled,omitempty"`
	AudioSource  string `json:"audioSource,omitempty"`
	AudioFile    string `json:"audioFile,omitempty"`
}

func (d *Device) cfgLock() { d.cfgMu.Lock() }
func (d *Device) cfgUnlock() { d.cfgMu.Unlock() }
func (d *Device) cfgRLock() { d.cfgMu.RLock() }
func (d *Device) cfgRUnlock() { d.cfgMu.RUnlock() }

// ListVideos 列出 assets 下可播放视频。
func ListVideos() ([]VideoItem, error) {
	entries, err := os.ReadDir(AssetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []VideoItem{}, nil
		}
		return nil, err
	}
	var out []VideoItem
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".mp4" && ext != ".h264" && ext != ".264" {
			continue
		}
		// 跳过抽流缓存
		if strings.Contains(name, ".cache") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, VideoItem{
			Name: name,
			Path: filepath.ToSlash(filepath.Join(AssetsDir, name)),
			Size: info.Size(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (d *Device) ListVideos() ([]VideoItem, error) {
	return ListVideos()
}

// AddChannel 运行时新增通道，并可选绑定视频。
func (d *Device) AddChannel(req AddChannelRequest) error {
	id := config.NormalizeGBID(req.ID)
	if len(id) != 20 {
		return fmt.Errorf("通道编号必须是 20 位国标编码")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "通道" + id[18:]
	}

	d.cfgLock()

	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
			d.cfgUnlock()
			return fmt.Errorf("通道已存在: %s", id)
		}
	}

	d.cfg.Device.Channels = append(d.cfg.Device.Channels, config.ChannelConfig{
		ID:           id,
		Name:         name,
		Manufacturer: d.cfg.Device.Manufacturer,
		Model:        "SIM-IPC",
		Address:      name,
		Status:       "ON",
		Parental:     0,
		ParentID:     d.cfg.Device.ID,
		RegisterWay:  1,
		CivilCode:    "340200",
	})

	if d.cfg.Media.Channels == nil {
		d.cfg.Media.Channels = map[string]config.ChannelMediaConfig{}
	}
	if mp4 := normalizeVideoPath(req.MP4); mp4 != "" {
		d.cfg.Media.Mode = "per_channel"
		d.cfg.Media.Channels[id] = config.ChannelMediaConfig{
			Source: "mp4",
			MP4File: mp4,
		}
	}

	newCh := d.cfg.Device.Channels[len(d.cfg.Device.Channels)-1]
	d.cfgUnlock()

	log.Printf("[ui] channel added id=%s name=%s mp4=%q", id, name, req.MP4)
	go d.NotifyCatalogChange("ADD", newCh)
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
	return nil
}

// BindChannelVideo 绑定/更新通道视频。
func (d *Device) BindChannelVideo(req BindChannelRequest) error {
	id := config.NormalizeGBID(req.ChannelID)
	if len(id) != 20 {
		return fmt.Errorf("通道编号必须是 20 位国标编码")
	}

	source := strings.ToLower(strings.TrimSpace(req.Source))
	if source != "" {
		switch source {
		case "mp4", "file", "synthetic", "ptz":
		default:
			return fmt.Errorf("source 必须是 mp4/file/synthetic/ptz")
		}
	}

	d.cfgLock()

	found := false
	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
			found = true
			break
		}
	}
	if !found {
		d.cfgUnlock()
		return fmt.Errorf("通道不存在: %s", id)
	}

	if d.cfg.Media.Channels == nil {
		d.cfg.Media.Channels = map[string]config.ChannelMediaConfig{}
	}
	// 绑定即切到 per_channel，否则共享模式下改绑定无效
	d.cfg.Media.Mode = "per_channel"

	// 从原有配置继承，避免仅更新伴音或单独字段时把视频文件或源类型冲掉
	cfg := d.cfg.Media.Channels[id]
	if source != "" {
		cfg.Source = source
	}
	if mp4 := normalizeVideoPath(req.MP4); mp4 != "" {
		cfg.MP4File = mp4
		if cfg.Source == "" {
			cfg.Source = "mp4"
		}
	}
	if h := normalizeVideoPath(req.H264); h != "" {
		cfg.H264File = h
		if cfg.Source == "" {
			cfg.Source = "file"
		}
	}
	if req.AudioEnabled != nil {
		cfg.AudioEnabled = req.AudioEnabled
	}
	if req.AudioSource != "" {
		cfg.AudioSource = req.AudioSource
	}
	if req.AudioFile != "" {
		cfg.AudioFile = req.AudioFile
	}

	d.cfg.Media.Channels[id] = cfg
	log.Printf("[ui] bind channel %s source=%s mp4=%s h264=%s audioEnabled=%v audioSource=%s",
		id, cfg.Source, cfg.MP4File, cfg.H264File, cfg.AudioEnabled, cfg.AudioSource)

	videoChanged := req.Source != "" || req.MP4 != "" || req.H264 != ""
	onChanged := d.OnConfigChanged
	cfgCopy := *d.cfg
	d.cfgUnlock()

	if onChanged != nil {
		onChanged(&cfgCopy)
	}

	if videoChanged {
		// 视频源发生变更：停掉旧推流，让下次点播使用新视频源
		if d.ms != nil {
			go d.ms.StopByChannel(id)
		}
	} else if d.ms != nil {
		// 仅伴音配置变更：对当前正在活跃推流的会话执行无缝热更新，不停视频流！
		opts := cfgCopy.Media.OptionsFor(id)
		src, codec, kind, _ := d.CreateAudioSourceForChannel(id)
		d.ms.UpdateChannelAudio(id, opts.AudioEnabled, src, codec, kind)
	}
	return nil
}

// RemoveChannel 删除通道。
func (d *Device) RemoveChannel(id string) error {
	id = config.NormalizeGBID(id)
	d.cfgLock()
	found := false
	var removedCh config.ChannelConfig
	for i, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
			removedCh = ch
			d.cfg.Device.Channels = append(d.cfg.Device.Channels[:i], d.cfg.Device.Channels[i+1:]...)
			if d.cfg.Media.Channels != nil {
				delete(d.cfg.Media.Channels, id)
			}
			found = true
			break
		}
	}
	d.cfgUnlock()
	if !found {
		return fmt.Errorf("通道不存在: %s", id)
	}
	log.Printf("[ui] channel removed %s", id)
	go d.ms.StopByChannel(id)
	go d.NotifyCatalogChange("DEL", removedCh)
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
	return nil
}

// SetMediaMode shared | per_channel
func (d *Device) SetMediaMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "shared" && mode != "per_channel" {
		return fmt.Errorf("mode 必须是 shared 或 per_channel")
	}
	d.cfgLock()
	d.cfg.Media.Mode = mode
	d.cfgUnlock()
	log.Printf("[ui] media mode -> %s", mode)
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
	return nil
}

// DeleteVideo 从 assets 中删除视频及抽流缓存。
func DeleteVideo(name string) error {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\\", "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if name == "" || name == "." || name == ".." {
		return fmt.Errorf("非法文件名")
	}
	p := filepath.Join(AssetsDir, name)
	if _, err := os.Stat(p); err != nil {
		return fmt.Errorf("视频不存在: %s", name)
	}
	_ = os.Remove(p)
	base := strings.TrimSuffix(name, filepath.Ext(name))
	cache := filepath.Join(AssetsDir, base+".h264.v3.cache")
	_ = os.Remove(cache)
	log.Printf("[ui] video deleted: %s", name)
	return nil
}

func (d *Device) DeleteVideo(name string) error {
	return DeleteVideo(name)
}

// UpdateChannelRequest 更新通道信息。
type UpdateChannelRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"` // ON | OFF
}

// UpdateChannel 更新通道名称或在线状态。
func (d *Device) UpdateChannel(req UpdateChannelRequest) error {
	id := config.NormalizeGBID(req.ID)
	d.cfgLock()
	found := false
	var updatedCh config.ChannelConfig
	var eventType = "UPDATE"
	for i := range d.cfg.Device.Channels {
		if d.cfg.Device.Channels[i].ID == id {
			if strings.TrimSpace(req.Name) != "" {
				d.cfg.Device.Channels[i].Name = strings.TrimSpace(req.Name)
			}
			st := strings.ToUpper(strings.TrimSpace(req.Status))
			if st == "ON" || st == "OFF" {
				if d.cfg.Device.Channels[i].Status != st {
					eventType = st
				}
				d.cfg.Device.Channels[i].Status = st
				if st == "OFF" {
					go d.ms.StopByChannel(id)
				}
			}
			found = true
			updatedCh = d.cfg.Device.Channels[i]
			break
		}
	}
	d.cfgUnlock()
	if !found {
		return fmt.Errorf("通道不存在: %s", id)
	}
	log.Printf("[ui] channel updated id=%s name=%s status=%s event=%s", id, updatedCh.Name, updatedCh.Status, eventType)
	go d.NotifyCatalogChange(eventType, updatedCh)
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
	return nil
}

// SetChannelStatus 快捷切换通道在线/离线状态 (ON/OFF) 并广播通知
func (d *Device) SetChannelStatus(channelID, status string) error {
	channelID = config.NormalizeGBID(channelID)
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "ON" && status != "OFF" {
		return fmt.Errorf("status 必须是 ON 或 OFF")
	}
	return d.UpdateChannel(UpdateChannelRequest{
		ID:     channelID,
		Status: status,
	})
}

// SendManualCatalogNotify 手动模拟触发指定通道的国标增量通知 (ON, OFF, VLOST, DEFECT, ADD, DEL, UPDATE)
func (d *Device) SendManualCatalogNotify(channelID, event string) error {
	channelID = config.NormalizeGBID(channelID)
	event = strings.ToUpper(strings.TrimSpace(event))
	d.cfgRLock()
	var foundCh *config.ChannelConfig
	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == channelID {
			c := ch
			foundCh = &c
			break
		}
	}
	rootID := d.cfg.Device.ID
	d.cfgRUnlock()

	if foundCh == nil {
		foundCh = &config.ChannelConfig{
			ID:       channelID,
			Name:     "通道" + channelID,
			Status:   "ON",
			ParentID: rootID,
		}
	}
	d.NotifyCatalogChange(event, *foundCh)
	return nil
}

// UpdateGPSConfig 更新并应用移动位置模拟配置（可指定通道ID，空或__master__则更新主设备）
func (d *Device) UpdateGPSConfig(cfg config.MobilePositionConfig, channelID ...string) {
	chID := ""
	if len(channelID) > 0 {
		chID = channelID[0]
	}
	if chID == "" {
		chID = cfg.ChannelID
	}

	d.cfgLock()
	if chID == "" || chID == "__master__" || chID == d.cfg.Device.ID {
		cfg.ChannelID = ""
		d.cfg.MobilePosition = cfg
	} else {
		cfg.ChannelID = chID
		for i := range d.cfg.Device.Channels {
			if d.cfg.Device.Channels[i].ID == chID {
				c := cfg
				d.cfg.Device.Channels[i].MobilePosition = &c
				break
			}
		}
	}
	d.cfgUnlock()

	if d.gpsMgr != nil {
		if chID == "" || chID == "__master__" || chID == d.cfg.Device.ID {
			d.gpsMgr.UpdateMasterConfig(cfg)
		} else {
			d.gpsMgr.UpdateChannelConfig(chID, cfg)
		}
	}
	log.Printf("[ui] GPS config updated for target '%s': enabled=%v pattern=%s speed=%.1f interval=%d",
		chID, cfg.Enabled, cfg.Pattern, cfg.Speed, cfg.Interval)

	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
}

// SyncAllChannelsGPS 一键将主配置同步给所有通道（followMode 为 true 表示设置为跟随主车，false 表示克隆独立轨迹）
func (d *Device) SyncAllChannelsGPS(baseCfg config.MobilePositionConfig, followMode bool) {
	d.cfgLock()
	d.cfg.MobilePosition = baseCfg
	for i := range d.cfg.Device.Channels {
		ch := &d.cfg.Device.Channels[i]
		chCfg := baseCfg
		chCfg.ChannelID = ch.ID
		if followMode {
			chCfg.Pattern = "follow"
		}
		ch.MobilePosition = &chCfg
	}
	d.cfgUnlock()

	if d.gpsMgr != nil {
		d.gpsMgr.SyncAllChannels(baseCfg, followMode)
	}

	log.Printf("[ui] Synced all channels GPS: followMode=%v", followMode)
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
}

// ReportGPSNow 手动立即触发单次位置上报（可指定通道ID）
func (d *Device) ReportGPSNow(channelID ...string) (GPSStatus, error) {
	if d.gpsMgr == nil {
		return GPSStatus{}, fmt.Errorf("gps manager not initialized")
	}
	return d.gpsMgr.ReportNow(channelID...)
}

// ForceIFrame 触发关键帧注入
func (d *Device) ForceIFrame(channelID string) bool {
	channelID = config.NormalizeGBID(channelID)
	applied := false
	if d.ms != nil {
		applied = d.ms.ForceIFrame(channelID)
	}
	d.recordControlEvent("Manual", channelID, "手动注入关键帧", fmt.Sprintf("applied=%v", applied), "WebUI")
	return applied
}

// TriggerReboot 手动触发模拟远程重启
func (d *Device) TriggerReboot() {
	d.recordControlEvent("Manual", d.cfg.Device.ID, "手动重启", "WebUI trigger", "WebUI")
	go d.triggerSimulatedReboot()
}

// SetRecordingManual 手动切换录像状态
func (d *Device) SetRecordingManual(recording bool) {
	d.SetRecording(recording)
	d.recordControlEvent("Manual", d.cfg.Device.ID, "手动切换录像", fmt.Sprintf("recording=%v", recording), "WebUI")
}

// SetTimeManual 手动修改或校正虚拟时钟
func (d *Device) SetTimeManual(targetTimeStr string) (time.Time, error) {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "15:04:05"}
	var targetTime time.Time
	var err error
	for _, l := range layouts {
		if t, e := time.ParseInLocation(l, targetTimeStr, time.Local); e == nil {
			if l == "15:04:05" {
				now := time.Now()
				targetTime = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
			} else {
				targetTime = t
			}
			err = nil
			break
		} else {
			err = e
		}
	}
	if err != nil {
		return time.Time{}, err
	}
	offset := targetTime.Sub(time.Now())
	d.SetTimeOffset(offset)
	d.recordControlEvent("Manual", d.cfg.Device.ID, "手动校时", fmt.Sprintf("target=%s offset=%.1fs", targetTime.Format("2006-01-02 15:04:05"), offset.Seconds()), "WebUI")
	return targetTime, nil
}

// ResetTimeOffset 复位时钟偏差为系统真实时间
func (d *Device) ResetTimeOffset() {
	d.SetTimeOffset(0)
	d.recordControlEvent("Manual", d.cfg.Device.ID, "时钟复位", "同步为系统本地时间", "WebUI")
}

// SetHomePosition 设置云台看守位
func (d *Device) SetHomePosition(channelID string, enabled bool, presetIndex, resetSec int) *PTZStatus {
	channelID = config.NormalizeGBID(channelID)
	ptz := d.GetChannelPTZ(channelID)
	st := ptz.SetHomePosition(enabled, presetIndex, resetSec)
	d.recordControlEvent("Manual", channelID, "设置守望位", fmt.Sprintf("enabled=%v preset=%d resetTime=%ds", enabled, presetIndex, resetSec), "WebUI")
	return st
}

func normalizeVideoPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	// 禁止路径穿越
	if strings.Contains(p, "..") {
		return ""
	}
	// 若只是文件名，放到 assets 下
	if !strings.Contains(p, "/") {
		return filepath.ToSlash(filepath.Join(AssetsDir, p))
	}
	return p
}
