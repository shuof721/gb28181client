package device

import (
	"math"
	"sync"
	"time"

	"github.com/local/gb28181-device/internal/config"
)

// GPSStatus 表示设备/通道当前最新的位置与运动信息
type GPSStatus struct {
	Enabled   bool    `json:"enabled"`
	ChannelID string  `json:"channelId"`
	Time      string  `json:"time"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Altitude  float64 `json:"altitude"`
	Speed     float64 `json:"speed"`     // km/h
	Direction float64 `json:"direction"` // 0~360° (0: 正北, 90: 正东, 180: 正南, 270: 正西)
	Pattern   string  `json:"pattern"`
	Interval  int     `json:"interval"`
	Mode      string  `json:"mode"`
}

// SingleTracker 跟踪单个实体（主设备或具体通道）的轨迹
type SingleTracker struct {
	cfg       config.MobilePositionConfig
	startTime time.Time
	lastSent  time.Time
}

// GPSManager 负责管理移动设备及其下挂通道的多路位置仿真与上报调度
type GPSManager struct {
	mu       sync.RWMutex
	master   *SingleTracker
	channels map[string]*SingleTracker // key: channelID
	sendFn   func(status GPSStatus) error

	ticker *time.Ticker
	stopCh chan struct{}
}

func sanitizeGPSConfig(cfg config.MobilePositionConfig) config.MobilePositionConfig {
	if cfg.Interval <= 0 {
		cfg.Interval = 5
	}
	if cfg.Pattern == "" {
		cfg.Pattern = "circle"
	}
	if cfg.Longitude == 0 && cfg.Latitude == 0 {
		cfg.Longitude = 116.397428
		cfg.Latitude = 39.909230
	}
	if cfg.Radius <= 0 {
		cfg.Radius = 500.0
	}
	if cfg.Speed <= 0 {
		cfg.Speed = 30.0
	}
	return cfg
}

// NewGPSManager 创建多通道 GPS 轨迹管理器
func NewGPSManager(masterCfg config.MobilePositionConfig, channels []config.ChannelConfig, sendFn func(status GPSStatus) error) *GPSManager {
	masterCfg = sanitizeGPSConfig(masterCfg)
	gm := &GPSManager{
		master: &SingleTracker{
			cfg:       masterCfg,
			startTime: time.Now(),
		},
		channels: make(map[string]*SingleTracker),
		sendFn:   sendFn,
		stopCh:   make(chan struct{}),
	}

	for _, ch := range channels {
		if ch.ID == "" {
			continue
		}
		var chCfg config.MobilePositionConfig
		if ch.MobilePosition != nil {
			chCfg = *ch.MobilePosition
		} else {
			chCfg = masterCfg
			chCfg.ChannelID = ch.ID
			chCfg.Pattern = "follow"
		}
		chCfg.ChannelID = ch.ID
		chCfg = sanitizeGPSConfig(chCfg)
		gm.channels[ch.ID] = &SingleTracker{
			cfg:       chCfg,
			startTime: time.Now(),
		}
	}

	return gm
}

// UpdateMasterConfig 更新主设备/全局 GPS 模板
func (g *GPSManager) UpdateMasterConfig(newCfg config.MobilePositionConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	newCfg = sanitizeGPSConfig(newCfg)
	g.master.cfg = newCfg
}

// UpdateChannelConfig 更新指定通道的 GPS 配置
func (g *GPSManager) UpdateChannelConfig(channelID string, newCfg config.MobilePositionConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if channelID == "" || channelID == "__master__" {
		newCfg = sanitizeGPSConfig(newCfg)
		g.master.cfg = newCfg
		return
	}

	newCfg.ChannelID = channelID
	newCfg = sanitizeGPSConfig(newCfg)
	tracker, ok := g.channels[channelID]
	if !ok {
		tracker = &SingleTracker{
			cfg:       newCfg,
			startTime: time.Now(),
		}
		g.channels[channelID] = tracker
	} else {
		tracker.cfg = newCfg
		tracker.startTime = time.Now()
	}
}

// SyncAllChannels 将配置同步到所有通道
func (g *GPSManager) SyncAllChannels(baseCfg config.MobilePositionConfig, followMode bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	baseCfg = sanitizeGPSConfig(baseCfg)
	g.master.cfg = baseCfg

	now := time.Now()
	for chID, tr := range g.channels {
		if followMode {
			tr.cfg.Pattern = "follow"
			tr.cfg.Enabled = baseCfg.Enabled
			tr.cfg.Interval = baseCfg.Interval
			tr.cfg.Mode = baseCfg.Mode
		} else {
			chCfg := baseCfg
			chCfg.ChannelID = chID
			tr.cfg = chCfg
			tr.startTime = now
		}
	}
}

// GetMasterConfig 获取主设备 GPS 配置
func (g *GPSManager) GetMasterConfig() config.MobilePositionConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.master.cfg
}

// GetChannelConfig 获取指定通道的 GPS 配置
func (g *GPSManager) GetChannelConfig(channelID string) config.MobilePositionConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if channelID == "" || channelID == "__master__" {
		return g.master.cfg
	}
	if tr, ok := g.channels[channelID]; ok {
		return tr.cfg
	}
	cfg := g.master.cfg
	cfg.ChannelID = channelID
	return cfg
}

// GetAllChannelConfigs 获取所有通道的 GPS 配置映射
func (g *GPSManager) GetAllChannelConfigs() map[string]config.MobilePositionConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	res := make(map[string]config.MobilePositionConfig, len(g.channels))
	for id, tr := range g.channels {
		res[id] = tr.cfg
	}
	return res
}

// Current 计算并返回指定通道（或主设备）当前时刻的位置状态
func (g *GPSManager) Current(channelID ...string) GPSStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()

	now := time.Now()
	chID := ""
	if len(channelID) > 0 {
		chID = channelID[0]
	}

	if chID == "" || chID == "__master__" {
		return g.calcTrackerStatus(g.master, chID, now)
	}

	if tr, ok := g.channels[chID]; ok {
		return g.calcTrackerStatus(tr, chID, now)
	}

	return g.calcTrackerStatus(g.master, chID, now)
}

// CurrentAll 返回所有已启用位置仿真的通道（若通道均未启用但主设备启用则返回主设备）
func (g *GPSManager) CurrentAll() []GPSStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()

	now := time.Now()
	var list []GPSStatus
	for chID, tr := range g.channels {
		if tr.cfg.Enabled {
			list = append(list, g.calcTrackerStatus(tr, chID, now))
		}
	}

	if len(list) == 0 && g.master.cfg.Enabled {
		list = append(list, g.calcTrackerStatus(g.master, g.master.cfg.ChannelID, now))
	}
	return list
}

// GetAllChannelStatuses 返回所有通道的当前实时状态
func (g *GPSManager) GetAllChannelStatuses() map[string]GPSStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()

	now := time.Now()
	res := make(map[string]GPSStatus, len(g.channels))
	for chID, tr := range g.channels {
		res[chID] = g.calcTrackerStatus(tr, chID, now)
	}
	return res
}

func (g *GPSManager) calcTrackerStatus(tr *SingleTracker, defaultChID string, now time.Time) GPSStatus {
	cfg := tr.cfg
	targetCh := cfg.ChannelID
	if targetCh == "" {
		targetCh = defaultChID
	}
	nowStr := now.Format("2006-01-02T15:04:05")

	if cfg.Pattern == "follow" {
		masterSt := g.calcTrackerStatusNoLock(g.master, targetCh, now)
		masterSt.ChannelID = targetCh
		masterSt.Enabled = cfg.Enabled
		masterSt.Pattern = "follow"
		masterSt.Interval = cfg.Interval
		masterSt.Mode = cfg.Mode
		return masterSt
	}

	if !cfg.Enabled || cfg.Pattern == "static" {
		return GPSStatus{
			Enabled:   cfg.Enabled,
			ChannelID: targetCh,
			Time:      nowStr,
			Longitude: cfg.Longitude,
			Latitude:  cfg.Latitude,
			Altitude:  cfg.Altitude,
			Speed:     0,
			Direction: cfg.Direction,
			Pattern:   cfg.Pattern,
			Interval:  cfg.Interval,
			Mode:      cfg.Mode,
		}
	}

	elapsed := now.Sub(tr.startTime).Seconds()
	speedMps := cfg.Speed / 3.6
	lon, lat, dir := calculateTrajectory(cfg.Pattern, cfg.Longitude, cfg.Latitude, cfg.Radius, speedMps, cfg.Direction, elapsed)

	return GPSStatus{
		Enabled:   cfg.Enabled,
		ChannelID: targetCh,
		Time:      nowStr,
		Longitude: lon,
		Latitude:  lat,
		Altitude:  cfg.Altitude,
		Speed:     cfg.Speed,
		Direction: dir,
		Pattern:   cfg.Pattern,
		Interval:  cfg.Interval,
		Mode:      cfg.Mode,
	}
}

func (g *GPSManager) calcTrackerStatusNoLock(tr *SingleTracker, defaultChID string, now time.Time) GPSStatus {
	cfg := tr.cfg
	targetCh := cfg.ChannelID
	if targetCh == "" {
		targetCh = defaultChID
	}
	nowStr := now.Format("2006-01-02T15:04:05")

	if !cfg.Enabled || cfg.Pattern == "static" {
		return GPSStatus{
			Enabled:   cfg.Enabled,
			ChannelID: targetCh,
			Time:      nowStr,
			Longitude: cfg.Longitude,
			Latitude:  cfg.Latitude,
			Altitude:  cfg.Altitude,
			Speed:     0,
			Direction: cfg.Direction,
			Pattern:   cfg.Pattern,
			Interval:  cfg.Interval,
			Mode:      cfg.Mode,
		}
	}

	elapsed := now.Sub(tr.startTime).Seconds()
	speedMps := cfg.Speed / 3.6
	lon, lat, dir := calculateTrajectory(cfg.Pattern, cfg.Longitude, cfg.Latitude, cfg.Radius, speedMps, cfg.Direction, elapsed)

	return GPSStatus{
		Enabled:   cfg.Enabled,
		ChannelID: targetCh,
		Time:      nowStr,
		Longitude: lon,
		Latitude:  lat,
		Altitude:  cfg.Altitude,
		Speed:     cfg.Speed,
		Direction: dir,
		Pattern:   cfg.Pattern,
		Interval:  cfg.Interval,
		Mode:      cfg.Mode,
	}
}

// calculateTrajectory 根据不同模式计算 (经度, 纬度, 方向角)
func calculateTrajectory(pattern string, originLon, originLat, radius, speedMps, initDir float64, elapsed float64) (float64, float64, float64) {
	// WGS-84 地球参考半径常数
	const earthRadius = 6378137.0
	metersPerLat := (math.Pi / 180.0) * earthRadius
	radLat := originLat * (math.Pi / 180.0)
	metersPerLon := (math.Pi / 180.0) * earthRadius * math.Cos(radLat)
	if metersPerLon <= 0 {
		metersPerLon = metersPerLat
	}

	var dx, dy float64 // 相对中心点米数 (东向为正 dx, 北向为正 dy)
	var dir float64    // 顺时针方向角 (0:北, 90:东, 180:南, 270:西)

	switch pattern {
	case "circle":
		// 圆周巡逻：角速度 omega = v / R
		if radius <= 0 {
			radius = 500.0
		}
		omega := speedMps / radius
		theta := omega * elapsed // 弧度

		dx = radius * math.Cos(theta)
		dy = radius * math.Sin(theta)

		// 瞬时运动切线方向：vx = -v*sin(theta), vy = v*cos(theta)
		vx := -speedMps * math.Sin(theta)
		vy := speedMps * math.Cos(theta)
		// 转换为航向角 (以正北为0度，顺时针方向)
		headingRad := math.Atan2(vx, vy)
		headingDeg := headingRad * (180.0 / math.Pi)
		if headingDeg < 0 {
			headingDeg += 360.0
		}
		dir = headingDeg

	case "linear":
		// 线性往返折返巡逻：沿初始方向往返
		if radius <= 0 {
			radius = 500.0
		}
		halfPeriod := radius / speedMps
		if halfPeriod <= 0 {
			halfPeriod = 10.0
		}
		totalPeriod := halfPeriod * 2.0
		cycleTime := math.Mod(elapsed, totalPeriod)

		dirRad := initDir * (math.Pi / 180.0)
		unitX := math.Sin(dirRad) // 顺时针，东向分量
		unitY := math.Cos(dirRad) // 顺时针，北向分量

		if cycleTime < halfPeriod {
			dist := speedMps * cycleTime
			dx = unitX * dist
			dy = unitY * dist
			dir = initDir
		} else {
			dist := radius - speedMps*(cycleTime-halfPeriod)
			dx = unitX * dist
			dy = unitY * dist
			dir = math.Mod(initDir+180.0, 360.0)
		}

	case "sine":
		// 正弦波动轨迹：沿主航向前进，横向正弦偏摆
		if radius <= 0 {
			radius = 500.0
		}
		halfPeriod := radius / speedMps
		if halfPeriod <= 0 {
			halfPeriod = 10.0
		}
		totalPeriod := halfPeriod * 2.0
		cycleTime := math.Mod(elapsed, totalPeriod)

		var mainDist float64
		var isForward bool
		if cycleTime < halfPeriod {
			mainDist = speedMps * cycleTime
			isForward = true
		} else {
			mainDist = radius - speedMps*(cycleTime-halfPeriod)
			isForward = false
		}

		// 横向振幅为半径的 1/4
		amplitude := radius / 4.0
		sineFreq := (2.0 * math.Pi) / halfPeriod
		transverseOffset := amplitude * math.Sin(sineFreq*cycleTime)

		dirRad := initDir * (math.Pi / 180.0)
		mainX := math.Sin(dirRad)
		mainY := math.Cos(dirRad)

		// 垂直于主方向的单位向量 (旋转 90 度)
		perpX := math.Cos(dirRad)
		perpY := -math.Sin(dirRad)

		dx = mainX*mainDist + perpX*transverseOffset
		dy = mainY*mainDist + perpY*transverseOffset

		if isForward {
			dir = initDir
		} else {
			dir = math.Mod(initDir+180.0, 360.0)
		}

	default:
		return originLon, originLat, initDir
	}

	dLon := dx / metersPerLon
	dLat := dy / metersPerLat

	return originLon + dLon, originLat + dLat, math.Round(dir*10) / 10
}

// Start 启动后台多通道独立定时上报协程
func (g *GPSManager) Start() {
	g.mu.Lock()
	g.ticker = time.NewTicker(1 * time.Second)
	g.mu.Unlock()

	go func() {
		for {
			select {
			case <-g.stopCh:
				return
			case <-g.ticker.C:
				g.dispatchTick()
			}
		}
	}()
}

func (g *GPSManager) dispatchTick() {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()

	// 1. 检查各通道
	dispatched := false
	for chID, tr := range g.channels {
		if !tr.cfg.Enabled {
			continue
		}
		if tr.cfg.Mode != "active" && tr.cfg.Mode != "both" {
			continue
		}
		intervalSec := tr.cfg.Interval
		if intervalSec <= 0 {
			intervalSec = 5
		}
		if now.Sub(tr.lastSent) >= time.Duration(intervalSec)*time.Second {
			tr.lastSent = now
			st := g.calcTrackerStatus(tr, chID, now)
			if g.sendFn != nil {
				go g.sendFn(st)
			}
			dispatched = true
		}
	}

	// 2. 若没有配置通道或未分发任何通道，检查主设备
	if !dispatched && g.master.cfg.Enabled {
		if g.master.cfg.Mode == "active" || g.master.cfg.Mode == "both" {
			intervalSec := g.master.cfg.Interval
			if intervalSec <= 0 {
				intervalSec = 5
			}
			if now.Sub(g.master.lastSent) >= time.Duration(intervalSec)*time.Second {
				g.master.lastSent = now
				st := g.calcTrackerStatus(g.master, g.master.cfg.ChannelID, now)
				if g.sendFn != nil {
					go g.sendFn(st)
				}
			}
		}
	}
}

// Stop 停止定时器
func (g *GPSManager) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.ticker != nil {
		g.ticker.Stop()
	}
	select {
	case <-g.stopCh:
	default:
		close(g.stopCh)
	}
}

// ReportNow 立即触发单次上报（可指定通道ID，空则上报所有开启通道或主设备）
func (g *GPSManager) ReportNow(channelID ...string) (GPSStatus, error) {
	chID := ""
	if len(channelID) > 0 {
		chID = channelID[0]
	}

	st := g.Current(chID)
	var err error
	if g.sendFn != nil {
		err = g.sendFn(st)
	}
	return st, err
}
