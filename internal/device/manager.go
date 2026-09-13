package device

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/storage"
)

// DeviceSummary 概览展示用的设备简明状态。
type DeviceSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Running        bool   `json:"running"`
	Registered     bool   `json:"registered"`
	Server         string `json:"server"`
	LocalPort      int    `json:"localPort"`
	Transport      string `json:"transport"`
	ChannelCount   int    `json:"channelCount"`
	ActiveSessions int    `json:"activeSessions"`
	UptimeSec      int64  `json:"uptimeSec"`
	Error          string `json:"error,omitempty"`
	Enabled        bool   `json:"enabled"`
}

// ManagedDevice 托管的设备运行态。
type ManagedDevice struct {
	Profile *config.DeviceProfile
	Dev     *Device
	Running bool
	Err     string
}

// Manager 多设备生命周期管理器。
type Manager struct {
	store   *storage.Store
	mu      sync.RWMutex
	devs    map[string]*ManagedDevice
	sysLogs *LogBuffer
}

func NewManager(store *storage.Store) *Manager {
	return &Manager{
		store:   store,
		devs:    make(map[string]*ManagedDevice),
		sysLogs: NewLogBuffer(1000),
	}
}

// AppendLog 全局日志写入（同时写入系统日志及当前运行的设备）。
func (m *Manager) AppendLog(line string) {
	m.sysLogs.Append(line)
	if !m.mu.TryRLock() {
		return
	}
	defer m.mu.RUnlock()
	for _, md := range m.devs {
		if md.Dev != nil {
			md.Dev.AppendLog(line)
		}
	}
}

// Logs 返回最近系统日志。
func (m *Manager) Logs(n int) []string {
	return m.sysLogs.Tail(n)
}

// Init 从存储中加载所有设备并自动启动 enabled=true 的设备。
func (m *Manager) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	profiles, err := m.store.List()
	if err != nil {
		return fmt.Errorf("load device profiles: %w", err)
	}

	for _, p := range profiles {
		id := p.Device.ID
		md := &ManagedDevice{
			Profile: p,
			Running: false,
		}
		m.devs[id] = md
		if p.Enabled {
			if err := m.startDeviceLocked(md); err != nil {
				log.Printf("[manager] auto-start device %s (%s) failed: %v", id, p.Device.Name, err)
				md.Err = err.Error()
			}
		}
	}
	log.Printf("[manager] initialized %d devices from store", len(m.devs))
	return nil
}

func (m *Manager) startDeviceLocked(md *ManagedDevice) error {
	if md.Running && md.Dev != nil {
		return nil
	}
	cfg := md.Profile.ToConfig()
	d := New(cfg)

	// 当通道或媒体模式变更时，自动同步并存入 JSON
	pRef := md.Profile
	d.OnConfigChanged = func(latestCfg *config.Config) {
		m.mu.Lock()
		defer m.mu.Unlock()
		pRef.Device.Channels = latestCfg.Device.Channels
		pRef.Media = latestCfg.Media
		if err := m.store.Save(pRef); err != nil {
			log.Printf("[manager] auto-save device %s config failed: %v", pRef.Device.ID, err)
		}
	}

	if err := d.Start(); err != nil {
		md.Err = err.Error()
		return err
	}
	md.Dev = d
	md.Running = true
	md.Err = ""
	log.Printf("[manager] device %s (%s) started on port %d", md.Profile.Device.ID, md.Profile.Device.Name, md.Profile.SIP.LocalPort)
	return nil
}

// Start 启动指定设备。
func (m *Manager) Start(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = config.NormalizeGBID(id)
	md, ok := m.devs[id]
	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}
	if md.Running {
		return nil
	}

	// 检查当前端口是否有冲突
	if err := m.checkPortConflictLocked(md.Profile.SIP.LocalPort, id); err != nil {
		return err
	}

	if err := m.startDeviceLocked(md); err != nil {
		return err
	}

	md.Profile.Enabled = true
	_ = m.store.Save(md.Profile)
	return nil
}

// Stop 停止指定设备。
func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	id = config.NormalizeGBID(id)
	md, ok := m.devs[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("device not found: %s", id)
	}
	if !md.Running || md.Dev == nil {
		md.Running = false
		md.Profile.Enabled = false
		_ = m.store.Save(md.Profile)
		m.mu.Unlock()
		return nil
	}

	dev := md.Dev
	md.Running = false
	md.Dev = nil
	md.Profile.Enabled = false
	_ = m.store.Save(md.Profile)
	m.mu.Unlock()

	dev.Stop()
	log.Printf("[manager] device %s stopped", id)
	return nil
}

// Restart 重启指定设备。
func (m *Manager) Restart(id string) error {
	_ = m.Stop(id)
	return m.Start(id)
}

// StartAll 启动所有设备。
func (m *Manager) StartAll() []error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, md := range m.devs {
		if !md.Running {
			if err := m.checkPortConflictLocked(md.Profile.SIP.LocalPort, md.Profile.Device.ID); err != nil {
				errs = append(errs, fmt.Errorf("device %s port conflict: %w", md.Profile.Device.ID, err))
				continue
			}
			if err := m.startDeviceLocked(md); err != nil {
				errs = append(errs, fmt.Errorf("device %s start failed: %w", md.Profile.Device.ID, err))
			} else {
				md.Profile.Enabled = true
				_ = m.store.Save(md.Profile)
			}
		}
	}
	return errs
}

// StopAll 停止所有设备。
func (m *Manager) StopAll() {
	m.mu.Lock()
	var toStop []*Device
	for _, md := range m.devs {
		if md.Running && md.Dev != nil {
			toStop = append(toStop, md.Dev)
			md.Running = false
			md.Dev = nil
		}
	}
	m.mu.Unlock()

	var wg sync.WaitGroup
	for _, dev := range toStop {
		wg.Add(1)
		go func(d *Device) {
			defer wg.Done()
			d.Stop()
		}(dev)
	}
	wg.Wait()
	log.Printf("[manager] all devices stopped")
}


// AddDevice 新增模拟设备。
func (m *Manager) AddDevice(p *config.DeviceProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p.ApplyDefaults()
	if err := p.Validate(); err != nil {
		return err
	}
	id := p.Device.ID
	if _, exists := m.devs[id]; exists {
		return fmt.Errorf("device already exists: %s", id)
	}

	// 端口冲突检查
	if err := m.checkPortConflictLocked(p.SIP.LocalPort, id); err != nil {
		return err
	}

	if err := m.store.Save(p); err != nil {
		return fmt.Errorf("save device: %w", err)
	}

	md := &ManagedDevice{
		Profile: p,
		Running: false,
	}
	m.devs[id] = md

	if p.Enabled {
		if err := m.startDeviceLocked(md); err != nil {
			return fmt.Errorf("device saved but failed to start: %w", err)
		}
	}
	return nil
}

// UpdateProfile 更新模拟设备全量配置。
func (m *Manager) UpdateProfile(id string, p *config.DeviceProfile, restart bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = config.NormalizeGBID(id)
	md, ok := m.devs[id]
	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}

	p.ApplyDefaults()
	if err := p.Validate(); err != nil {
		return err
	}

	if p.Device.ID != id {
		return fmt.Errorf("changing device ID is not allowed")
	}

	// 端口检查
	if err := m.checkPortConflictLocked(p.SIP.LocalPort, id); err != nil {
		return err
	}

	if err := m.store.Save(p); err != nil {
		return fmt.Errorf("save device profile: %w", err)
	}

	md.Profile = p

	if md.Running && restart {
		if md.Dev != nil {
			md.Dev.Stop()
			md.Dev = nil
			md.Running = false
		}
		if err := m.startDeviceLocked(md); err != nil {
			return fmt.Errorf("updated, but restart failed: %w", err)
		}
	}
	return nil
}

// DeleteDevice 删除指定设备。
func (m *Manager) DeleteDevice(id string) error {
	m.mu.Lock()
	id = config.NormalizeGBID(id)
	md, ok := m.devs[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("device not found: %s", id)
	}

	dev := md.Dev
	if md.Running && dev != nil {
		md.Running = false
		md.Dev = nil
	}

	if err := m.store.Delete(id); err != nil {
		m.mu.Unlock()
		return err
	}

	delete(m.devs, id)
	m.mu.Unlock()

	if dev != nil {
		dev.Stop()
	}
	log.Printf("[manager] device %s deleted", id)
	return nil
}

// NextAvailablePort 自动探测下一个空闲可用的本地 SIP 端口。
func (m *Manager) NextAvailablePort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	used := make(map[int]bool)
	for _, md := range m.devs {
		used[md.Profile.SIP.LocalPort] = true
	}

	port := 5070
	for {
		if !used[port] {
			return port
		}
		port++
	}
}

func (m *Manager) checkPortConflictLocked(port int, selfID string) error {
	for otherID, md := range m.devs {
		if otherID == selfID {
			continue
		}
		if md.Profile.SIP.LocalPort == port && md.Running {
			return fmt.Errorf("local port %d is already in use by running device %s (%s)", port, otherID, md.Profile.Device.Name)
		}
	}
	return nil
}

// ListSummaries 获取所有设备的概要信息列表。
func (m *Manager) ListSummaries() []DeviceSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []DeviceSummary
	for id, md := range m.devs {
		s := DeviceSummary{
			ID:           id,
			Name:         md.Profile.Device.Name,
			Running:      md.Running,
			Server:       fmt.Sprintf("%s:%d", md.Profile.SIP.ServerIP, md.Profile.SIP.ServerPort),
			LocalPort:    md.Profile.SIP.LocalPort,
			Transport:    md.Profile.SIP.Transport,
			ChannelCount: len(md.Profile.Device.Channels),
			Enabled:      md.Profile.Enabled,
			Error:        md.Err,
		}
		if md.Running && md.Dev != nil {
			st := md.Dev.Status()
			s.Registered = st.Registered
			s.ActiveSessions = len(st.Sessions)
			s.UptimeSec = st.UptimeSec
		}
		list = append(list, s)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

// Get 获取设备托管信息。
func (m *Manager) Get(id string) (*ManagedDevice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = config.NormalizeGBID(id)
	md, ok := m.devs[id]
	if !ok {
		return nil, fmt.Errorf("device not found: %s", id)
	}
	return md, nil
}

// GetDevice 获取运行中的设备实例。
func (m *Manager) GetDevice(id string) (*Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = config.NormalizeGBID(id)
	md, ok := m.devs[id]
	if !ok {
		return nil, fmt.Errorf("device not found: %s", id)
	}
	if !md.Running || md.Dev == nil {
		return nil, fmt.Errorf("device %s is stopped", id)
	}
	return md.Dev, nil
}

// AddChannel 新增通道（运行中热增，停止态保存至配置）。
func (m *Manager) AddChannel(deviceID string, req AddChannelRequest) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}

	chID := config.NormalizeGBID(req.ID)
	if len(chID) != 20 {
		return fmt.Errorf("通道编号必须是 20 位国标编码")
	}

	if running {
		return dev.AddChannel(req)
	}

	// 离线/停止态：直接改 profile 并存盘
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range md.Profile.Device.Channels {
		if ch.ID == chID {
			return fmt.Errorf("通道已存在: %s", chID)
		}
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "通道" + chID[18:]
	}
	md.Profile.Device.Channels = append(md.Profile.Device.Channels, config.ChannelConfig{
		ID:           chID,
		Name:         name,
		Manufacturer: md.Profile.Device.Manufacturer,
		Model:        "SIM-IPC",
		Address:      name,
		Status:       "ON",
		Parental:     0,
		ParentID:     md.Profile.Device.ID,
		RegisterWay:  1,
		CivilCode:    "340200",
	})
	if md.Profile.Media.Channels == nil {
		md.Profile.Media.Channels = map[string]config.ChannelMediaConfig{}
	}
	if mp4 := normalizeVideoPath(req.MP4); mp4 != "" {
		md.Profile.Media.Mode = "per_channel"
		md.Profile.Media.Channels[chID] = config.ChannelMediaConfig{
			Source:  "mp4",
			MP4File: mp4,
		}
	}
	return m.store.Save(md.Profile)
}

// UpdateChannel 更新通道信息。
func (m *Manager) UpdateChannel(deviceID string, req UpdateChannelRequest) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}
	if running {
		return dev.UpdateChannel(req)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	chID := config.NormalizeGBID(req.ID)
	found := false
	for i := range md.Profile.Device.Channels {
		if md.Profile.Device.Channels[i].ID == chID {
			if strings.TrimSpace(req.Name) != "" {
				md.Profile.Device.Channels[i].Name = strings.TrimSpace(req.Name)
			}
			st := strings.ToUpper(strings.TrimSpace(req.Status))
			if st == "ON" || st == "OFF" {
				md.Profile.Device.Channels[i].Status = st
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("通道不存在: %s", chID)
	}
	return m.store.Save(md.Profile)
}

// RemoveChannel 删除通道。
func (m *Manager) RemoveChannel(deviceID, channelID string) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}
	if running {
		return dev.RemoveChannel(channelID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	chID := config.NormalizeGBID(channelID)
	found := false
	for i, ch := range md.Profile.Device.Channels {
		if ch.ID == chID {
			md.Profile.Device.Channels = append(md.Profile.Device.Channels[:i], md.Profile.Device.Channels[i+1:]...)
			if md.Profile.Media.Channels != nil {
				delete(md.Profile.Media.Channels, chID)
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("通道不存在: %s", chID)
	}
	return m.store.Save(md.Profile)
}

// BindChannelVideo 绑定通道媒体文件。
func (m *Manager) BindChannelVideo(deviceID string, req BindChannelRequest) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}
	if running {
		return dev.BindChannelVideo(req)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	chID := config.NormalizeGBID(req.ChannelID)
	if md.Profile.Media.Channels == nil {
		md.Profile.Media.Channels = map[string]config.ChannelMediaConfig{}
	}
	source := strings.ToLower(strings.TrimSpace(req.Source))
	if source == "" {
		source = "mp4"
	}
	cfg := config.ChannelMediaConfig{Source: source}
	if mp4 := normalizeVideoPath(req.MP4); mp4 != "" {
		cfg.MP4File = mp4
	}
	if h := normalizeVideoPath(req.H264); h != "" {
		cfg.H264File = h
	}
	md.Profile.Media.Channels[chID] = cfg
	md.Profile.Media.Mode = "per_channel"
	return m.store.Save(md.Profile)
}

// SetMediaMode 设置媒体分发模式 shared | per_channel。
func (m *Manager) SetMediaMode(deviceID, mode string) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}
	if running {
		return dev.SetMediaMode(mode)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "shared" && mode != "per_channel" {
		return fmt.Errorf("mode 必须是 shared 或 per_channel")
	}
	md.Profile.Media.Mode = mode
	return m.store.Save(md.Profile)
}

// SetChannelStatus 快捷切换指定设备的通道在线/离线状态
func (m *Manager) SetChannelStatus(deviceID, channelID, status string) error {
	return m.UpdateChannel(deviceID, UpdateChannelRequest{
		ID:     channelID,
		Status: status,
	})
}

// SendCatalogNotify 手动模拟触发指定通道的国标增量通知
func (m *Manager) SendCatalogNotify(deviceID, channelID, event string) error {
	id := config.NormalizeGBID(deviceID)
	dev, err := m.GetDevice(id)
	if err != nil {
		return err
	}
	return dev.SendManualCatalogNotify(channelID, event)
}

// UpdateGPSConfig 更新指定设备的 GPS 配置
func (m *Manager) UpdateGPSConfig(deviceID string, cfg config.MobilePositionConfig, channelID ...string) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}

	chID := ""
	if len(channelID) > 0 {
		chID = channelID[0]
	}
	if chID == "" {
		chID = cfg.ChannelID
	}

	if running {
		dev.UpdateGPSConfig(cfg, chID)
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if chID == "" || chID == "__master__" || chID == md.Profile.Device.ID {
		cfg.ChannelID = ""
		md.Profile.MobilePosition = cfg
	} else {
		cfg.ChannelID = chID
		for i := range md.Profile.Device.Channels {
			if md.Profile.Device.Channels[i].ID == chID {
				c := cfg
				md.Profile.Device.Channels[i].MobilePosition = &c
				break
			}
		}
	}
	return m.store.Save(md.Profile)
}

// SyncAllChannelsGPS 一键同步所有通道 GPS 配置
func (m *Manager) SyncAllChannelsGPS(deviceID string, baseCfg config.MobilePositionConfig, followMode bool) error {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	var dev *Device
	running := false
	if ok {
		dev = md.Dev
		running = md.Running && dev != nil
	}
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device not found: %s", id)
	}

	if running {
		dev.SyncAllChannelsGPS(baseCfg, followMode)
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	md.Profile.MobilePosition = baseCfg
	for i := range md.Profile.Device.Channels {
		ch := &md.Profile.Device.Channels[i]
		chCfg := baseCfg
		chCfg.ChannelID = ch.ID
		if followMode {
			chCfg.Pattern = "follow"
		}
		ch.MobilePosition = &chCfg
	}
	return m.store.Save(md.Profile)
}

// ReportGPSNow 立即触发单次位置上报
func (m *Manager) ReportGPSNow(deviceID string, channelID ...string) (GPSStatus, error) {
	id := config.NormalizeGBID(deviceID)
	dev, err := m.GetDevice(id)
	if err != nil {
		return GPSStatus{}, err
	}
	return dev.ReportGPSNow(channelID...)
}

// GetGPSStatus 获取指定通道或主设备的 GPS 状态与配置
func (m *Manager) GetGPSStatus(deviceID string, channelID ...string) (GPSStatus, config.MobilePositionConfig, error) {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	m.mu.RUnlock()

	if !ok {
		return GPSStatus{}, config.MobilePositionConfig{}, fmt.Errorf("device not found: %s", id)
	}

	chID := ""
	if len(channelID) > 0 {
		chID = channelID[0]
	}

	if md.Running && md.Dev != nil && md.Dev.GPSManager() != nil {
		gm := md.Dev.GPSManager()
		return gm.Current(chID), gm.GetChannelConfig(chID), nil
	}

	if chID == "" || chID == "__master__" || chID == md.Profile.Device.ID {
		cfg := md.Profile.MobilePosition
		return GPSStatus{
			Enabled:   cfg.Enabled,
			ChannelID: cfg.ChannelID,
			Longitude: cfg.Longitude,
			Latitude:  cfg.Latitude,
			Altitude:  cfg.Altitude,
			Speed:     cfg.Speed,
			Direction: cfg.Direction,
			Pattern:   cfg.Pattern,
			Interval:  cfg.Interval,
			Mode:      cfg.Mode,
		}, cfg, nil
	}

	// 查找该通道
	for _, ch := range md.Profile.Device.Channels {
		if ch.ID == chID {
			if ch.MobilePosition != nil {
				cfg := *ch.MobilePosition
				return GPSStatus{
					Enabled:   cfg.Enabled,
					ChannelID: chID,
					Longitude: cfg.Longitude,
					Latitude:  cfg.Latitude,
					Altitude:  cfg.Altitude,
					Speed:     cfg.Speed,
					Direction: cfg.Direction,
					Pattern:   cfg.Pattern,
					Interval:  cfg.Interval,
					Mode:      cfg.Mode,
				}, cfg, nil
			}
			cfg := md.Profile.MobilePosition
			cfg.ChannelID = chID
			cfg.Pattern = "follow"
			return GPSStatus{
				Enabled:   cfg.Enabled,
				ChannelID: chID,
				Longitude: cfg.Longitude,
				Latitude:  cfg.Latitude,
				Altitude:  cfg.Altitude,
				Speed:     cfg.Speed,
				Direction: cfg.Direction,
				Pattern:   "follow",
				Interval:  cfg.Interval,
				Mode:      cfg.Mode,
			}, cfg, nil
		}
	}

	return GPSStatus{}, config.MobilePositionConfig{}, fmt.Errorf("channel not found: %s", chID)
}

// GetAllGPS 获取主设备及所有下挂通道的当前最新位置与配置
func (m *Manager) GetAllGPS(deviceID string) (GPSStatus, config.MobilePositionConfig, map[string]GPSStatus, map[string]config.MobilePositionConfig, error) {
	id := config.NormalizeGBID(deviceID)
	m.mu.RLock()
	md, ok := m.devs[id]
	m.mu.RUnlock()

	if !ok {
		return GPSStatus{}, config.MobilePositionConfig{}, nil, nil, fmt.Errorf("device not found: %s", id)
	}

	if md.Running && md.Dev != nil && md.Dev.GPSManager() != nil {
		gm := md.Dev.GPSManager()
		masterSt := gm.Current()
		masterCfg := gm.GetMasterConfig()
		chStatuses := gm.GetAllChannelStatuses()
		chConfigs := gm.GetAllChannelConfigs()
		return masterSt, masterCfg, chStatuses, chConfigs, nil
	}

	masterCfg := md.Profile.MobilePosition
	masterSt := GPSStatus{
		Enabled:   masterCfg.Enabled,
		Longitude: masterCfg.Longitude,
		Latitude:  masterCfg.Latitude,
		Altitude:  masterCfg.Altitude,
		Speed:     masterCfg.Speed,
		Direction: masterCfg.Direction,
		Pattern:   masterCfg.Pattern,
		Interval:  masterCfg.Interval,
		Mode:      masterCfg.Mode,
	}

	chStatuses := make(map[string]GPSStatus)
	chConfigs := make(map[string]config.MobilePositionConfig)
	for _, ch := range md.Profile.Device.Channels {
		if ch.MobilePosition != nil {
			chConfigs[ch.ID] = *ch.MobilePosition
			chStatuses[ch.ID] = GPSStatus{
				Enabled:   ch.MobilePosition.Enabled,
				ChannelID: ch.ID,
				Longitude: ch.MobilePosition.Longitude,
				Latitude:  ch.MobilePosition.Latitude,
				Altitude:  ch.MobilePosition.Altitude,
				Speed:     ch.MobilePosition.Speed,
				Direction: ch.MobilePosition.Direction,
				Pattern:   ch.MobilePosition.Pattern,
				Interval:  ch.MobilePosition.Interval,
				Mode:      ch.MobilePosition.Mode,
			}
		} else {
			cfg := masterCfg
			cfg.ChannelID = ch.ID
			cfg.Pattern = "follow"
			chConfigs[ch.ID] = cfg
			chStatuses[ch.ID] = GPSStatus{
				Enabled:   masterCfg.Enabled,
				ChannelID: ch.ID,
				Longitude: masterCfg.Longitude,
				Latitude:  masterCfg.Latitude,
				Altitude:  masterCfg.Altitude,
				Speed:     masterCfg.Speed,
				Direction: masterCfg.Direction,
				Pattern:   "follow",
				Interval:  masterCfg.Interval,
				Mode:      masterCfg.Mode,
			}
		}
	}

	return masterSt, masterCfg, chStatuses, chConfigs, nil
}

// ListSubscriptions 获取指定设备的订阅者列表
func (m *Manager) ListSubscriptions(deviceID string) ([]*Subscriber, error) {
	id := config.NormalizeGBID(deviceID)
	dev, err := m.GetDevice(id)
	if err != nil {
		return nil, err
	}
	return dev.SubscriptionManager().ListAll(), nil
}


