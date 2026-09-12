package device

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	ChannelID string `json:"channelId"`
	Source    string `json:"source"` // mp4 | file | synthetic
	MP4       string `json:"mp4"`
	H264      string `json:"h264"`
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

	d.cfgUnlock()

	log.Printf("[ui] channel added id=%s name=%s mp4=%q", id, name, req.MP4)
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
	if source == "" {
		source = "mp4"
	}
	switch source {
	case "mp4", "file", "synthetic":
	default:
		return fmt.Errorf("source 必须是 mp4/file/synthetic")
	}

	d.cfgLock()
	defer d.cfgUnlock()

	found := false
	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("通道不存在: %s", id)
	}

	if d.cfg.Media.Channels == nil {
		d.cfg.Media.Channels = map[string]config.ChannelMediaConfig{}
	}
	// 绑定即切到 per_channel，否则共享模式下改绑定无效
	if source != "synthetic" || req.MP4 != "" || req.H264 != "" {
		d.cfg.Media.Mode = "per_channel"
	}

	cfg := config.ChannelMediaConfig{Source: source}
	if mp4 := normalizeVideoPath(req.MP4); mp4 != "" {
		cfg.MP4File = mp4
		if source == "" {
			cfg.Source = "mp4"
		}
	}
	if h := normalizeVideoPath(req.H264); h != "" {
		cfg.H264File = h
	}
	if source == "mp4" && cfg.MP4File == "" {
		// 继承全局
		if d.cfg.Media.MP4File != "" {
			cfg.MP4File = d.cfg.Media.MP4File
		}
	}

	d.cfg.Media.Channels[id] = cfg
	d.cfgUnlock()
	log.Printf("[ui] bind channel %s source=%s mp4=%s h264=%s", id, cfg.Source, cfg.MP4File, cfg.H264File)

	// 若该通道正在播，停掉让下次点播用新源
	go d.ms.StopByChannel(id)
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
	return nil
}

// RemoveChannel 删除通道。
func (d *Device) RemoveChannel(id string) error {
	id = config.NormalizeGBID(id)
	d.cfgLock()
	found := false
	for i, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
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
	name := ""
	status := ""
	for i := range d.cfg.Device.Channels {
		if d.cfg.Device.Channels[i].ID == id {
			if strings.TrimSpace(req.Name) != "" {
				d.cfg.Device.Channels[i].Name = strings.TrimSpace(req.Name)
			}
			st := strings.ToUpper(strings.TrimSpace(req.Status))
			if st == "ON" || st == "OFF" {
				d.cfg.Device.Channels[i].Status = st
				if st == "OFF" {
					go d.ms.StopByChannel(id)
				}
			}
			found = true
			name = d.cfg.Device.Channels[i].Name
			status = d.cfg.Device.Channels[i].Status
			break
		}
	}
	d.cfgUnlock()
	if !found {
		return fmt.Errorf("通道不存在: %s", id)
	}
	log.Printf("[ui] channel updated id=%s name=%s status=%s", id, name, status)
	go d.SendCatalogNotify()
	if d.OnConfigChanged != nil {
		d.OnConfigChanged(d.cfg)
	}
	return nil
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
