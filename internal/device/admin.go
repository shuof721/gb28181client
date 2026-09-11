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
func (d *Device) ListVideos() ([]VideoItem, error) {
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
	defer d.cfgUnlock()

	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
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

	log.Printf("[ui] channel added id=%s name=%s mp4=%q", id, name, req.MP4)
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
	log.Printf("[ui] bind channel %s source=%s mp4=%s h264=%s", id, cfg.Source, cfg.MP4File, cfg.H264File)

	// 若该通道正在播，停掉让下次点播用新源
	go d.ms.StopByChannel(id)
	return nil
}

// RemoveChannel 删除通道。
func (d *Device) RemoveChannel(id string) error {
	id = config.NormalizeGBID(id)
	d.cfgLock()
	defer d.cfgUnlock()
	for i, ch := range d.cfg.Device.Channels {
		if ch.ID == id {
			d.cfg.Device.Channels = append(d.cfg.Device.Channels[:i], d.cfg.Device.Channels[i+1:]...)
			if d.cfg.Media.Channels != nil {
				delete(d.cfg.Media.Channels, id)
			}
			log.Printf("[ui] channel removed %s", id)
			go d.ms.StopByChannel(id)
			return nil
		}
	}
	return fmt.Errorf("通道不存在: %s", id)
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
