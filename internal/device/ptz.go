package device

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/gb28181"
)

const (
	maxPanSpeedDegPerSec  = 60.0 // 最大水平旋转角速度 (度/秒)
	maxTiltSpeedDegPerSec = 30.0 // 最大垂直旋转角速度 (度/秒)
	maxZoomRatePerSec     = 2.0  // 最大变倍速度 (倍/秒)
	presetTransitDegSec   = 90.0 // 调用预置位时的平滑过渡速度 (度/秒)
)

// PTZStatus 供 Web 交互与状态呈现。
type PTZStatus struct {
	ChannelID      string             `json:"channelId"`
	Pan            float64            `json:"pan"`            // 0.0 ~ 360.0°
	Tilt           float64            `json:"tilt"`           // -90.0 ~ +90.0°
	Zoom           float64            `json:"zoom"`           // 1.0 ~ 30.0x
	Focus          float64            `json:"focus"`          // 0.0 ~ 100.0%
	Iris           float64            `json:"iris"`           // 0.0 ~ 100.0%
	IsMoving       bool               `json:"isMoving"`       // 是否在运动中
	PanDir         int                `json:"panDir"`         // -1: 左, 0: 停, 1: 右
	TiltDir        int                `json:"tiltDir"`        // -1: 下, 0: 停, 1: 上
	ZoomDir        int                `json:"zoomDir"`        // -1: 缩小, 0: 停, 1: 放大
	FocusDir       int                `json:"focusDir"`       // -1: 近, 0: 停, 1: 远
	IrisDir        int                `json:"irisDir"`        // -1: 关, 0: 停, 1: 开
	PanSpeed       int                `json:"panSpeed"`
	TiltSpeed      int                `json:"tiltSpeed"`
	ZoomSpeed      int                `json:"zoomSpeed"`
	FocusSpeed     int                `json:"focusSpeed"`
	IrisSpeed      int                `json:"irisSpeed"`
	ActivePresetID int                `json:"activePresetId"` // 当前定格的预置位 ID (若在预置位上)
	ActivePreset   string             `json:"activePreset"`   // 预置位名称
	StatusDesc     string             `json:"statusDesc"`     // 人类可读状态描述
	Presets        []config.PTZPreset `json:"presets"`
}

// ChannelPTZ 单个通道的虚拟云台状态机。
type ChannelPTZ struct {
	mu        sync.RWMutex
	channelID string

	pan   float64 // 0.0 ~ 360.0°
	tilt  float64 // -90.0 ~ +90.0°
	zoom  float64 // 1.0 ~ 30.0x
	focus float64 // 0.0 ~ 100.0%
	iris  float64 // 0.0 ~ 100.0%

	isMoving   bool
	panDir     int
	tiltDir    int
	zoomDir    int
	focusDir   int
	irisDir    int
	panSpeed   int
	tiltSpeed  int
	zoomSpeed  int
	focusSpeed int
	irisSpeed  int

	// 平滑跳转至预置位状态
	isTransitioning bool
	targetPan       float64
	targetTilt      float64
	targetZoom      float64
	targetPresetID  int
	targetPresetNm  string

	activePresetID int
	activePresetNm string

	lastUpdate time.Time
	presets    map[int]config.PTZPreset

	onPresetChanged func(ch string, ptzCfg *config.PTZConfig)
}

// NewChannelPTZ 创建通道虚拟云台。
func NewChannelPTZ(chID string, initial *config.PTZConfig, onChange func(string, *config.PTZConfig)) *ChannelPTZ {
	if initial == nil {
		initial = config.DefaultPTZConfig()
	}
	p := &ChannelPTZ{
		channelID:       chID,
		pan:             wrapPan(initial.Pan),
		tilt:            clampTilt(initial.Tilt),
		zoom:            clampZoom(initial.Zoom),
		focus:           50.0,
		iris:            50.0,
		lastUpdate:      time.Now(),
		presets:         make(map[int]config.PTZPreset),
		onPresetChanged: onChange,
	}
	if p.zoom < 1.0 {
		p.zoom = 1.0
	}

	for _, pr := range initial.Presets {
		p.presets[pr.ID] = pr
	}

	// 检查是否恰好停在某个初始预置位
	p.matchActivePresetLocked()
	return p
}

func wrapPan(val float64) float64 {
	val = math.Mod(val, 360.0)
	if val < 0 {
		val += 360.0
	}
	return math.Round(val*10) / 10
}

func clampTilt(val float64) float64 {
	if val > 90.0 {
		val = 90.0
	} else if val < -90.0 {
		val = -90.0
	}
	return math.Round(val*10) / 10
}

func clampZoom(val float64) float64 {
	if val < 1.0 {
		val = 1.0
	} else if val > 30.0 {
		val = 30.0
	}
	return math.Round(val*10) / 10
}

// updatePositionLocked 根据经过的时间动态平滑推进姿态。
func (p *ChannelPTZ) updatePositionLocked(now time.Time) {
	dt := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	if dt <= 0 {
		return
	}
	// 避免异常长时间未更新导致的物理突变（最多按 2 秒计）
	if dt > 2.0 {
		dt = 2.0
	}

	if p.isTransitioning {
		p.stepTransitionLocked(dt)
		return
	}

	if !p.isMoving {
		return
	}

	// 水平转动
	if p.panDir != 0 {
		spd := float64(p.panSpeed)
		if spd <= 0 {
			spd = 64
		}
		degPerSec := (spd / 255.0) * maxPanSpeedDegPerSec
		p.pan = wrapPan(p.pan + float64(p.panDir)*degPerSec*dt)
	}

	// 垂直转动
	if p.tiltDir != 0 {
		spd := float64(p.tiltSpeed)
		if spd <= 0 {
			spd = 64
		}
		degPerSec := (spd / 255.0) * maxTiltSpeedDegPerSec
		p.tilt = clampTilt(p.tilt + float64(p.tiltDir)*degPerSec*dt)
	}

	// 镜头变倍
	if p.zoomDir != 0 {
		spd := float64(p.zoomSpeed)
		if spd <= 0 {
			spd = 4
		}
		// 基础速率 0.4x/s + 速度等级 0~15 动态映射，spd=2 约 0.88x/s，spd=15 约 4.0x/s
		ratePerSec := 0.4 + (spd/15.0)*3.6
		p.zoom = clampZoom(p.zoom + float64(p.zoomDir)*ratePerSec*dt)
	}

	// 镜头聚焦 (0.0 ~ 100.0%)
	if p.focusDir != 0 {
		spd := float64(p.focusSpeed)
		if spd <= 0 {
			spd = 64
		}
		rate := 10.0 + (spd/255.0)*30.0
		p.focus = math.Max(0.0, math.Min(100.0, p.focus+float64(p.focusDir)*rate*dt))
	}

	// 镜头光圈 (0.0 ~ 100.0%)
	if p.irisDir != 0 {
		spd := float64(p.irisSpeed)
		if spd <= 0 {
			spd = 64
		}
		rate := 10.0 + (spd/255.0)*30.0
		p.iris = math.Max(0.0, math.Min(100.0, p.iris+float64(p.irisDir)*rate*dt))
	}

	p.matchActivePresetLocked()
}

// stepTransitionLocked 平滑推进预置位过渡。
func (p *ChannelPTZ) stepTransitionLocked(dt float64) {
	step := presetTransitDegSec * dt

	// 水平角最短路径过渡 (跨 0/360 度保护)
	diffPan := p.targetPan - p.pan
	if diffPan > 180 {
		diffPan -= 360
	} else if diffPan < -180 {
		diffPan += 360
	}

	panDone := math.Abs(diffPan) <= step
	if panDone {
		p.pan = p.targetPan
	} else {
		if diffPan > 0 {
			p.pan = wrapPan(p.pan + step)
		} else {
			p.pan = wrapPan(p.pan - step)
		}
	}

	// 垂直角过渡
	diffTilt := p.targetTilt - p.tilt
	tiltDone := math.Abs(diffTilt) <= step
	if tiltDone {
		p.tilt = p.targetTilt
	} else {
		if diffTilt > 0 {
			p.tilt = clampTilt(p.tilt + step)
		} else {
			p.tilt = clampTilt(p.tilt - step)
		}
	}

	// 变倍过渡
	diffZoom := p.targetZoom - p.zoom
	zoomStep := maxZoomRatePerSec * dt
	zoomDone := math.Abs(diffZoom) <= zoomStep
	if zoomDone {
		p.zoom = p.targetZoom
	} else {
		if diffZoom > 0 {
			p.zoom = clampZoom(p.zoom + zoomStep)
		} else {
			p.zoom = clampZoom(p.zoom - zoomStep)
		}
	}

	if panDone && tiltDone && zoomDone {
		p.isTransitioning = false
		p.isMoving = false
		p.activePresetID = p.targetPresetID
		p.activePresetNm = p.targetPresetNm
	}
}

func (p *ChannelPTZ) matchActivePresetLocked() {
	p.activePresetID = 0
	p.activePresetNm = ""
	for id, pr := range p.presets {
		if math.Abs(p.pan-pr.Pan) < 0.5 && math.Abs(p.tilt-pr.Tilt) < 0.5 && math.Abs(p.zoom-pr.Zoom) < 0.2 {
			p.activePresetID = id
			p.activePresetNm = pr.Name
			break
		}
	}
}

// ExecuteCommand 执行国标 PTZ 指令。
func (p *ChannelPTZ) ExecuteCommand(cmd *gb28181.PTZCommand) (*PTZStatus, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	p.updatePositionLocked(now)

	switch cmd.Action {
	case gb28181.PTZActionStop:
		p.isMoving = false
		p.isTransitioning = false
		p.panDir = 0
		p.tiltDir = 0
		p.zoomDir = 0
		p.focusDir = 0
		p.irisDir = 0
		p.matchActivePresetLocked()

	case gb28181.PTZActionMove:
		p.isTransitioning = false
		p.isMoving = (cmd.PanDir != 0 || cmd.TiltDir != 0 || cmd.ZoomDir != 0 || cmd.FocusDir != 0 || cmd.IrisDir != 0)
		p.panDir = cmd.PanDir
		p.tiltDir = cmd.TiltDir
		p.zoomDir = cmd.ZoomDir
		p.focusDir = cmd.FocusDir
		p.irisDir = cmd.IrisDir
		p.panSpeed = cmd.PanSpeed
		p.tiltSpeed = cmd.TiltSpeed
		p.zoomSpeed = cmd.ZoomSpeed
		p.focusSpeed = cmd.FocusSpeed
		p.irisSpeed = cmd.IrisSpeed
		p.activePresetID = 0
		p.activePresetNm = ""

	case gb28181.PTZActionPresetSet:
		id := cmd.PresetID
		if id <= 0 || id > 255 {
			return nil, fmt.Errorf("invalid preset id %d", id)
		}
		name := fmt.Sprintf("预置位%d", id)
		if existing, ok := p.presets[id]; ok && existing.Name != "" {
			name = existing.Name
		}
		p.presets[id] = config.PTZPreset{
			ID:   id,
			Name: name,
			Pan:  p.pan,
			Tilt: p.tilt,
			Zoom: p.zoom,
		}
		p.activePresetID = id
		p.activePresetNm = name
		p.notifyConfigChangedLocked()

	case gb28181.PTZActionPresetCall:
		id := cmd.PresetID
		pr, ok := p.presets[id]
		if !ok {
			return nil, fmt.Errorf("preset %d not found", id)
		}
		p.isTransitioning = true
		p.isMoving = true
		p.targetPan = wrapPan(pr.Pan)
		p.targetTilt = clampTilt(pr.Tilt)
		p.targetZoom = clampZoom(pr.Zoom)
		p.targetPresetID = id
		p.targetPresetNm = pr.Name

	case gb28181.PTZActionPresetDelete:
		id := cmd.PresetID
		delete(p.presets, id)
		if p.activePresetID == id {
			p.activePresetID = 0
			p.activePresetNm = ""
		}
		p.notifyConfigChangedLocked()

	default:
		// 巡航或其他扩展指令暂以静止状态响应
	}

	return p.statusLocked(), nil
}

// ManualControl 手动控制云台（Web 控制台调用）。
func (p *ChannelPTZ) ManualControl(action string, panSpeed, tiltSpeed, zoomSpeed int) *PTZStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	p.updatePositionLocked(now)

	p.isTransitioning = false
	p.activePresetID = 0
	p.activePresetNm = ""

	if panSpeed <= 0 {
		panSpeed = 64
	}
	if tiltSpeed <= 0 {
		tiltSpeed = 64
	}
	if zoomSpeed <= 0 {
		zoomSpeed = 4
	}

	switch action {
	case "stop":
		p.isMoving = false
		p.panDir = 0
		p.tiltDir = 0
		p.zoomDir = 0
		p.matchActivePresetLocked()
	case "left":
		p.isMoving = true
		p.panDir = -1
		p.tiltDir = 0
		p.zoomDir = 0
		p.panSpeed = panSpeed
	case "right":
		p.isMoving = true
		p.panDir = 1
		p.tiltDir = 0
		p.zoomDir = 0
		p.panSpeed = panSpeed
	case "up":
		p.isMoving = true
		p.panDir = 0
		p.tiltDir = 1
		p.zoomDir = 0
		p.tiltSpeed = tiltSpeed
	case "down":
		p.isMoving = true
		p.panDir = 0
		p.tiltDir = -1
		p.zoomDir = 0
		p.tiltSpeed = tiltSpeed
	case "upleft":
		p.isMoving = true
		p.panDir = -1
		p.tiltDir = 1
		p.zoomDir = 0
		p.panSpeed = panSpeed
		p.tiltSpeed = tiltSpeed
	case "upright":
		p.isMoving = true
		p.panDir = 1
		p.tiltDir = 1
		p.zoomDir = 0
		p.panSpeed = panSpeed
		p.tiltSpeed = tiltSpeed
	case "downleft":
		p.isMoving = true
		p.panDir = -1
		p.tiltDir = -1
		p.zoomDir = 0
		p.panSpeed = panSpeed
		p.tiltSpeed = tiltSpeed
	case "downright":
		p.isMoving = true
		p.panDir = 1
		p.tiltDir = -1
		p.zoomDir = 0
		p.panSpeed = panSpeed
		p.tiltSpeed = tiltSpeed
	case "zoomin":
		p.isMoving = true
		p.panDir = 0
		p.tiltDir = 0
		p.zoomDir = 1
		p.zoomSpeed = zoomSpeed
	case "zoomout":
		p.isMoving = true
		p.panDir = 0
		p.tiltDir = 0
		p.zoomDir = -1
		p.zoomSpeed = zoomSpeed
	case "reset":
		p.isMoving = false
		p.isTransitioning = false
		p.panDir = 0
		p.tiltDir = 0
		p.zoomDir = 0
		p.pan = 0
		p.tilt = 0
		p.zoom = 1.0
		p.matchActivePresetLocked()
	}

	return p.statusLocked()
}

// SetPose 直接设置云台坐标姿态。
func (p *ChannelPTZ) SetPose(pan, tilt, zoom float64) *PTZStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isMoving = false
	p.isTransitioning = false
	p.panDir = 0
	p.tiltDir = 0
	p.zoomDir = 0
	p.pan = wrapPan(pan)
	p.tilt = clampTilt(tilt)
	p.zoom = clampZoom(zoom)
	p.matchActivePresetLocked()
	return p.statusLocked()
}

// CallPreset 调用指定预置位。
func (p *ChannelPTZ) CallPreset(presetID int) (*PTZStatus, error) {
	cmd := &gb28181.PTZCommand{
		Action:   gb28181.PTZActionPresetCall,
		PresetID: presetID,
	}
	return p.ExecuteCommand(cmd)
}

// SetPreset 设置或重命名预置位。
func (p *ChannelPTZ) SetPreset(presetID int, name string, pan, tilt, zoom float64, useCurrent bool) (*PTZStatus, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.updatePositionLocked(time.Now())

	if presetID <= 0 || presetID > 255 {
		return nil, fmt.Errorf("invalid preset id %d", presetID)
	}

	if name == "" {
		name = fmt.Sprintf("预置位%d", presetID)
	}

	targetPan := wrapPan(pan)
	targetTilt := clampTilt(tilt)
	targetZoom := clampZoom(zoom)
	if useCurrent {
		targetPan = p.pan
		targetTilt = p.tilt
		targetZoom = p.zoom
	}

	p.presets[presetID] = config.PTZPreset{
		ID:   presetID,
		Name: name,
		Pan:  targetPan,
		Tilt: targetTilt,
		Zoom: targetZoom,
	}
	p.matchActivePresetLocked()
	p.notifyConfigChangedLocked()
	return p.statusLocked(), nil
}

// DeletePreset 删除预置位。
func (p *ChannelPTZ) DeletePreset(presetID int) (*PTZStatus, error) {
	cmd := &gb28181.PTZCommand{
		Action:   gb28181.PTZActionPresetDelete,
		PresetID: presetID,
	}
	return p.ExecuteCommand(cmd)
}

// Status 获取当前姿态快照。
func (p *ChannelPTZ) Status() PTZStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updatePositionLocked(time.Now())
	return *p.statusLocked()
}

func (p *ChannelPTZ) statusLocked() *PTZStatus {
	var prs []config.PTZPreset
	for _, pr := range p.presets {
		prs = append(prs, pr)
	}
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].ID < prs[j].ID
	})

	var desc string
	if p.isTransitioning {
		desc = fmt.Sprintf("正在平滑移动至预置位 #%d (%s)...", p.targetPresetID, p.targetPresetNm)
	} else if p.isMoving {
		var d []string
		if p.panDir > 0 {
			d = append(d, "向右")
		} else if p.panDir < 0 {
			d = append(d, "向左")
		}
		if p.tiltDir > 0 {
			d = append(d, "向上")
		} else if p.tiltDir < 0 {
			d = append(d, "向下")
		}
		if p.zoomDir > 0 {
			d = append(d, "变倍放大")
		} else if p.zoomDir < 0 {
			d = append(d, "变倍缩小")
		}
		if p.focusDir > 0 {
			d = append(d, "聚焦远")
		} else if p.focusDir < 0 {
			d = append(d, "聚焦近")
		}
		if p.irisDir > 0 {
			d = append(d, "光圈开")
		} else if p.irisDir < 0 {
			d = append(d, "光圈关")
		}
		desc = fmt.Sprintf("运动中: %s", strings.Join(d, "+"))
	} else if p.activePresetID > 0 {
		desc = fmt.Sprintf("定位于预置位 #%d (%s)", p.activePresetID, p.activePresetNm)
	} else {
		desc = "静止"
	}

	return &PTZStatus{
		ChannelID:      p.channelID,
		Pan:            p.pan,
		Tilt:           p.tilt,
		Zoom:           p.zoom,
		Focus:          p.focus,
		Iris:           p.iris,
		IsMoving:       p.isMoving || p.isTransitioning,
		PanDir:         p.panDir,
		TiltDir:        p.tiltDir,
		ZoomDir:        p.zoomDir,
		FocusDir:       p.focusDir,
		IrisDir:        p.irisDir,
		PanSpeed:       p.panSpeed,
		TiltSpeed:      p.tiltSpeed,
		ZoomSpeed:      p.zoomSpeed,
		FocusSpeed:     p.focusSpeed,
		IrisSpeed:      p.irisSpeed,
		ActivePresetID: p.activePresetID,
		ActivePreset:   p.activePresetNm,
		StatusDesc:     desc,
		Presets:        prs,
	}
}

// DragZoom 处理国标拉框放大/拉框缩小 (3D智能定位与变倍)
func (p *ChannelPTZ) DragZoom(isZoomIn bool, length, width, midX, midY, lenX, lenY int) *PTZStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.updatePositionLocked(time.Now())

	if length <= 0 {
		length = 1920
	}
	if width <= 0 {
		width = 1080
	}
	if lenX <= 0 {
		lenX = 200
	}
	if lenY <= 0 {
		lenY = 200
	}

	baseFovH := 60.0
	fovH := baseFovH / p.zoom
	fovV := fovH * (float64(width) / float64(length))

	// 水平与垂直归一化偏移 (-1.0 ~ +1.0)
	normX := (float64(midX) - float64(length)/2.0) / (float64(length) / 2.0)
	normY := -(float64(midY) - float64(width)/2.0) / (float64(width) / 2.0)

	targetPan := wrapPan(p.pan + normX*(fovH*0.5))
	targetTilt := clampTilt(p.tilt + normY*(fovV*0.5))

	// 放大比率计算
	ratio := float64(length) / float64(lenX)
	if ratio < 1.2 {
		ratio = 1.2
	} else if ratio > 6.0 {
		ratio = 6.0
	}

	var targetZoom float64
	if isZoomIn {
		targetZoom = clampZoom(p.zoom * ratio)
	} else {
		targetZoom = clampZoom(p.zoom / ratio)
	}

	p.isTransitioning = true
	p.isMoving = true
	p.targetPan = targetPan
	p.targetTilt = targetTilt
	p.targetZoom = targetZoom
	p.targetPresetID = 0
	if isZoomIn {
		p.targetPresetNm = "拉框放大"
	} else {
		p.targetPresetNm = "拉框缩小"
	}

	return p.statusLocked()
}

// PresetsList 返回供国标 MANSCDP 响应的预置位列表。
func (p *ChannelPTZ) PresetsList() []config.PTZPreset {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var prs []config.PTZPreset
	for _, pr := range p.presets {
		prs = append(prs, pr)
	}
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].ID < prs[j].ID
	})
	return prs
}

func (p *ChannelPTZ) notifyConfigChangedLocked() {
	if p.onPresetChanged == nil {
		return
	}
	var prs []config.PTZPreset
	for _, pr := range p.presets {
		prs = append(prs, pr)
	}
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].ID < prs[j].ID
	})
	cfg := &config.PTZConfig{
		Enabled: true,
		Pan:     p.pan,
		Tilt:    p.tilt,
		Zoom:    p.zoom,
		Presets: prs,
	}
	go p.onPresetChanged(p.channelID, cfg)
}
