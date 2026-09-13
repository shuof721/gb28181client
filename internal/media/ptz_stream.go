package media

import (
	"bytes"
	"fmt"
	"math"
	"sync"
	"time"
)

// PTZInfo 云台姿态信息。
type PTZInfo struct {
	Pan            float64 // 0.0 ~ 360.0°
	Tilt           float64 // -90.0 ~ +90.0°
	Zoom           float64 // 1.0 ~ 30.0x
	Focus          float64 // 0.0 ~ 100.0%
	Iris           float64 // 0.0 ~ 100.0%
	IsMoving       bool
	ActivePresetID int
	ActivePreset   string
	StatusDesc     string
}

// PTZSourceOptions PTZ 动态视频源参数。
type PTZSourceOptions struct {
	ChannelID string
	Width     int
	Height    int
	FPS       int
	GetPTZ    func() PTZInfo
}

// ptzLandmark 虚拟园区 360° 全景地标。
type ptzLandmark struct {
	az   float64 // 方位角 0~360°
	dist float64 // 距离（米）
	name string  // 地标名称
	kind string  // gate, tower, highway, parking, datacenter, fence, logistics, substation
	y    byte    // YUV 颜色
	u    byte
	v    byte
}

var defaultLandmarks = []ptzLandmark{
	{az: 0, dist: 70, name: "GATE-01", kind: "gate", y: 170, u: 200, v: 50},           // 园区正门 (Cyan)
	{az: 45, dist: 160, name: "TOWER-A", kind: "tower", y: 150, u: 210, v: 100},       // 综合科研大厦 A座 (Indigo)
	{az: 90, dist: 120, name: "HIGHWAY-E", kind: "highway", y: 160, u: 90, v: 45},      // 东环主干道 (Green)
	{az: 135, dist: 85, name: "PARK-P1", kind: "parking", y: 190, u: 50, v: 150},       // 智能生态停车场 (Yellow)
	{az: 180, dist: 130, name: "IDC-BLDG-B", kind: "datacenter", y: 140, u: 190, v: 90}, // 核心数据机房 B座 (Blue)
	{az: 225, dist: 100, name: "FENCE-W", kind: "fence", y: 120, u: 90, v: 220},        // 西周界电子防区 (Red)
	{az: 270, dist: 140, name: "LOGISTICS", kind: "logistics", y: 160, u: 50, v: 180},   // 智慧物流装卸平台 (Orange)
	{az: 315, dist: 180, name: "TEL-TOWER", kind: "substation", y: 150, u: 180, v: 170}, // 变电站与微波铁塔 (Purple)
}

// PTZSource 纯 Go 实现的 PTZ 动态全景 H.264 视频源。
type PTZSource struct {
	chID     string
	w, h     int
	fps      int
	sps, pps []byte
	getPTZ   func() PTZInfo

	mu       sync.Mutex
	frameIdx int
	lastTime time.Time

	// 预分配 YUV 内存画布
	yBuf  []byte
	cbBuf []byte
	crBuf []byte
}

// NewPTZSource 创建 PTZ 动态视频流源。
func NewPTZSource(opts PTZSourceOptions) (*PTZSource, error) {
	w := opts.Width
	h := opts.Height
	fps := opts.FPS

	// PTZ 虚拟流采用纯 Go 原生 I_PCM 宏块光栅化（无外部硬件/FFmpeg 依赖）。
	// I_PCM 宏块包含无压缩原始 YUV 采样。若采用 720P/1080P 会导致单帧高达 1.4MB、产生 280Mbps 突发流，
	// 导致操作系统 UDP 缓冲区溢出丢包（表现为播放器只能解出顶部 20% 画面、下半部分断片）。
	// 经过实测，480x272 (16:9) @ 15fps 是纯 Go 软件合成画面的黄金尺寸（单帧仅 ~195KB），
	// 既能保证全景园区、罗盘、准星、建筑群与 OSD 100% 完整流畅显示，又能彻底杜绝任何网络丢包。
	if w <= 0 || w > 640 {
		w = 480
	}
	if h <= 0 || h > 368 {
		h = 272
	}
	if fps <= 0 || fps > 20 {
		fps = 15
	}
	w = (w / 16) * 16
	h = (h / 16) * 16

	s := &PTZSource{
		chID:   opts.ChannelID,
		w:      w,
		h:      h,
		fps:    fps,
		getPTZ: opts.GetPTZ,
		yBuf:   make([]byte, w*h),
		cbBuf:  make([]byte, (w/2)*(h/2)),
		crBuf:  make([]byte, (w/2)*(h/2)),
	}
	s.sps = buildSPS(w, h)
	s.pps = buildPPS()
	return s, nil
}

func (s *PTZSource) Close() error {
	return nil
}

func (s *PTZSource) Seek(offsetSec float64, fps int) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if offsetSec < 0 {
		offsetSec = 0
	}
	if fps <= 0 {
		fps = s.fps
	}
	s.frameIdx = int(offsetSec * float64(fps))
	s.lastTime = time.Time{}
	return offsetSec, nil
}

// ForceIFrame 强制下一帧立即生成并输出 SPS/PPS + IDR 帧
func (s *PTZSource) ForceIFrame() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastTime = time.Time{} // 清空节流时间，立即生成下一帧
}

func (s *PTZSource) Next() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 严格按帧率节流
	if !s.lastTime.IsZero() {
		interval := time.Second / time.Duration(s.fps)
		d := interval - time.Since(s.lastTime)
		if d > 0 {
			time.Sleep(d)
		}
	}
	s.lastTime = time.Now()
	s.frameIdx++

	// 获取实时云台姿态
	var ptz PTZInfo
	if s.getPTZ != nil {
		ptz = s.getPTZ()
	}
	if ptz.Zoom < 1.0 {
		ptz.Zoom = 1.0
	}

	// 渲染当前帧到 YUV 画布
	s.renderScene(ptz)

	// 编码为 H.264 I_PCM 宏块切片
	idr := s.encodeIPCMFrame()

	var buf bytes.Buffer
	buf.Write(s.sps)
	buf.Write(s.pps)
	buf.Write(idr)
	return buf.Bytes(), nil
}

// renderScene 执行纯 Go YUV 光栅化。
func (s *PTZSource) renderScene(ptz PTZInfo) {
	w := s.w
	h := s.h

	pan := ptz.Pan
	for pan < 0 {
		pan += 360
	}
	pan = math.Mod(pan, 360.0)

	tilt := ptz.Tilt
	if tilt > 90 {
		tilt = 90
	} else if tilt < -90 {
		tilt = -90
	}

	zoom := ptz.Zoom
	if zoom < 1.0 {
		zoom = 1.0
	} else if zoom > 30.0 {
		zoom = 30.0
	}

	baseFovH := 60.0
	fovH := baseFovH / zoom
	fovV := fovH * (float64(h) / float64(w))

	// 地平线 Y 坐标：tilt > 0 仰视（地平线下降），tilt < 0 俯视（地平线上升）
	horizonY := int(float64(h)*0.5 + (tilt/(fovV*0.5))*float64(h)*0.5)

	// 1. 绘制深空背景 (Sky)
	skyYEnd := horizonY
	if skyYEnd > h {
		skyYEnd = h
	}
	if skyYEnd > 0 {
		s.fillRect(0, 0, w, skyYEnd, 26, 158, 112) // 深邃靛蓝天空
		// 绘制星空
		for i := 0; i < 40; i++ {
			starAz := float64((i * 47) % 360)
			starElev := float64(5 + (i*17)%75)
			diff := normDegDiff(starAz, pan)
			if math.Abs(diff) < fovH*0.55 {
				sx := int(float64(w)*0.5 + (diff/(fovH*0.5))*float64(w)*0.5)
				sy := horizonY - int((starElev/(fovV*0.5))*float64(h)*0.5)
				if sy >= 2 && sy < horizonY-3 {
					s.setPixel(sx, sy, 220, 128, 128)
				}
			}
		}
	}

	// 2. 绘制大地与透视网格 (Ground)
	if horizonY < h {
		gTop := horizonY
		if gTop < 0 {
			gTop = 0
		}
		s.fillRect(0, gTop, w, h-gTop, 36, 116, 120) // 深暗地面灰绿

		// 地面经向辐射线（每隔 15°）
		panFloor := math.Floor(pan/15.0) * 15.0
		for a := panFloor - 45; a <= panFloor + 45; a += 15.0 {
			diff := normDegDiff(a, pan)
			if math.Abs(diff) < fovH*0.65 {
				startX := int(float64(w)*0.5 + (diff/(fovH*0.5))*float64(w)*0.5)
				bottomX := int(float64(w)*0.5 + (diff/(fovH*0.5))*float64(w)*0.5*2.6)
				s.drawLine(startX, horizonY, bottomX, h, 68, 148, 92)
			}
		}

		// 地面距离虚线圈 (60m, 100m, 160m)
		rings := []struct {
			label string
			ratio float64
		}{
			{"160M", 0.22},
			{"100M", 0.48},
			{"60M", 0.78},
		}
		for _, rg := range rings {
			ry := horizonY + int(float64(h-horizonY)*rg.ratio)
			if ry > 0 && ry < h {
				for x := 0; x < w; x += 8 {
					s.drawHLine(x, x+4, ry, 60, 155, 95)
				}
				s.drawString(w-60, ry-9, rg.label, 80, 160, 90)
			}
		}
	}

	// 地平线荧光高亮线
	if horizonY >= 0 && horizonY < h {
		s.drawHLine(0, w, horizonY, 160, 185, 60)
	}

	// 3. 绘制 360° 园区地标群
	var lockedLM *ptzLandmark
	minDiff := 999.0

	for i := range defaultLandmarks {
		lm := &defaultLandmarks[i]
		diff := normDegDiff(lm.az, pan)
		if math.Abs(diff) < fovH*0.65 {
			sx := int(float64(w)*0.5 + (diff/(fovH*0.5))*float64(w)*0.5)
			scale := math.Min(3.2, math.Max(0.4, (80.0/lm.dist)*(0.6+zoom*0.35)))
			groundRatio := math.Min(0.9, math.Max(0.15, 65.0/lm.dist))
			baseY := horizonY + int(math.Max(10, float64(h-horizonY)*groundRatio))

			if math.Abs(diff) < minDiff {
				minDiff = math.Abs(diff)
				if math.Abs(diff) < 5.0 {
					lockedLM = lm
				}
			}

			isLocked := (lockedLM == lm)
			s.drawLandmark(lm, sx, baseY, scale, isLocked)
		}
	}

	// 4. 战术顶部罗盘天际刻度标尺 (Top Compass Ribbon)
	s.drawCompassRibbon(pan, fovH)

	// 5. 战术中央十字准星 (Tactical Crosshair)
	s.drawCrosshair(w/2, h/2, lockedLM != nil)

	// 6. OSD 水印烧录 (On-Screen Display)
	s.drawOSD(ptz, lockedLM)
}

// drawLandmark 绘制单个标志建筑物剪影。
func (s *PTZSource) drawLandmark(lm *ptzLandmark, cx, baseY int, scale float64, isLocked bool) {
	sc := scale
	switch lm.kind {
	case "gate": // 园区大门
		pw := int(12 * sc)
		ph := int(45 * sc)
		s.fillRect(cx-int(40*sc), baseY-ph, pw, ph, lm.y, lm.u, lm.v)
		s.fillRect(cx+int(28*sc), baseY-ph, pw, ph, lm.y, lm.u, lm.v)
		// 门头横梁
		s.fillRect(cx-int(45*sc), baseY-ph-int(12*sc), int(85*sc), int(12*sc), lm.y+30, lm.u, lm.v)
		// 红色道闸杆
		s.fillRect(cx-int(28*sc), baseY-int(10*sc), int(56*sc), int(3*sc), 80, 90, 220)

	case "tower": // 科研大厦 A座
		tw := int(55 * sc)
		th := int(120 * sc)
		s.fillRect(cx-tw/2, baseY-th, tw, th, lm.y, lm.u, lm.v)
		// 阶梯顶部
		s.fillRect(cx-int(18*sc), baseY-th-int(20*sc), int(36*sc), int(20*sc), lm.y+20, lm.u, lm.v)
		// 楼顶点天线
		s.drawVLine(cx, baseY-th-int(38*sc), baseY-th-int(20*sc), 200, 128, 128)
		// 航空障碍警示灯（周期闪烁）
		if (time.Now().UnixMilli()/500)%2 == 0 {
			s.fillRect(cx-1, baseY-th-int(40*sc), 3, 3, 120, 90, 240)
		}

	case "highway": // 东环主干道
		rw := int(60 * sc)
		rh := int(30 * sc)
		s.fillRect(cx-rw/2, baseY-rh, rw, rh, 50, 128, 128)
		// 黄色车道虚线
		for y := baseY - rh + 2; y < baseY; y += 6 {
			s.drawVLine(cx, y, y+3, 190, 40, 150)
		}

	case "parking": // 智能生态停车场
		pw := int(80 * sc)
		ph := int(24 * sc)
		s.fillRect(cx-pw/2, baseY-ph, pw, ph, lm.y, lm.u, lm.v)
		// 光伏雨棚
		s.fillRect(cx-pw/2-2, baseY-ph-int(8*sc), pw+4, int(4*sc), lm.y+40, lm.u, lm.v)
		// 充电桩绿灯
		s.setPixel(cx-pw/2+5, baseY-ph-int(4*sc), 160, 40, 40)

	case "datacenter": // 数据机房 B座
		dw := int(90 * sc)
		dh := int(50 * sc)
		s.fillRect(cx-dw/2, baseY-dh, dw, dh, lm.y, lm.u, lm.v)
		// 冷光机柜通槽
		for k := -1; k <= 1; k++ {
			s.fillRect(cx+k*int(25*sc)-2, baseY-dh+int(8*sc), 4, dh-int(16*sc), 180, 190, 60)
		}

	case "fence": // 周界防区
		fw := int(100 * sc)
		fh := int(35 * sc)
		// 防区立柱
		for f := -2; f <= 2; f++ {
			s.fillRect(cx+f*int(22*sc)-1, baseY-fh, 3, fh, 90, 128, 128)
		}
		// 红色激光射线（呼吸变光）
		laserY := 90 + byte(30.0*math.Sin(float64(time.Now().UnixMilli())*0.008))
		s.drawHLine(cx-fw/2, cx+fw/2, baseY-fh+int(4*sc), laserY, 80, 230)

	case "logistics": // 物流装卸平台
		lw := int(95 * sc)
		lh := int(45 * sc)
		s.fillRect(cx-lw/2, baseY-lh, lw, lh, lm.y, lm.u, lm.v)
		// 货柜箱堆叠（蓝色与橙色）
		s.fillRect(cx-int(35*sc), baseY-int(16*sc), int(28*sc), int(16*sc), 120, 200, 60)
		s.fillRect(cx-int(32*sc), baseY-int(28*sc), int(22*sc), int(12*sc), 140, 50, 200)

	case "substation": // 通讯铁塔
		th := int(130 * sc)
		// A字塔架两翼
		s.drawLine(cx-int(16*sc), baseY, cx, baseY-th, 180, 128, 128)
		s.drawLine(cx+int(16*sc), baseY, cx, baseY-th, 180, 128, 128)
		// 塔顶微波天线鼓
		s.fillRect(cx-int(6*sc), baseY-th+int(15*sc), int(12*sc), int(8*sc), 140, 140, 120)
		// 闪烁警示灯
		if (time.Now().UnixMilli()/400)%2 == 0 {
			s.fillRect(cx-1, baseY-th-3, 3, 3, 130, 90, 230)
		}
	}

	// 顶部地标标签
	tagY := baseY - int(60*sc)
	if lm.kind == "tower" {
		tagY = baseY - int(155*sc)
	} else if lm.kind == "substation" {
		tagY = baseY - int(140*sc)
	}
	s.drawString(cx-len(lm.name)*4, tagY, lm.name, 220, 128, 128)

	// 目标锁定瞄准框
	if isLocked {
		bw := int(35 * sc)
		bh := int(25 * sc)
		if bw < 30 {
			bw = 30
		}
		if bh < 25 {
			bh = 25
		}
		pulse := int(2.0 * math.Sin(float64(time.Now().UnixMilli())*0.008))
		bw += pulse
		bh += pulse
		y1 := baseY - bh*2
		y2 := baseY + 2
		x1 := cx - bw
		x2 := cx + bw
		clen := 6

		// 四角取景框
		s.drawHLine(x1, x1+clen, y1, 190, 40, 40)
		s.drawVLine(x1, y1, y1+clen, 190, 40, 40)

		s.drawHLine(x2-clen, x2, y1, 190, 40, 40)
		s.drawVLine(x2, y1, y1+clen, 190, 40, 40)

		s.drawHLine(x1, x1+clen, y2, 190, 40, 40)
		s.drawVLine(x1, y2-clen, y2, 190, 40, 40)

		s.drawHLine(x2-clen, x2, y2, 190, 40, 40)
		s.drawVLine(x2, y2-clen, y2, 190, 40, 40)

		s.drawString(cx-28, y1-11, "TARGET LOCK", 190, 40, 40)
	}
}

// drawCompassRibbon 绘制顶部滚动战术罗盘。
func (s *PTZSource) drawCompassRibbon(pan, fovH float64) {
	w := s.w
	rw := 280
	rx1 := (w - rw) / 2
	rx2 := rx1 + rw
	ry := 8
	rh := 16

	s.fillRect(rx1, ry, rw, rh, 20, 128, 128)
	s.drawHLine(rx1, rx2, ry, 70, 160, 100)
	s.drawHLine(rx1, rx2, ry+rh, 70, 160, 100)

	halfSpan := 35.0
	startDeg := math.Floor((pan-halfSpan)/5.0) * 5.0
	endDeg := math.Ceil((pan+halfSpan)/5.0) * 5.0

	for deg := startDeg; deg <= endDeg; deg += 5.0 {
		dDiff := normDegDiff(deg, pan)
		x := int(float64(w)*0.5 + (dDiff/halfSpan)*float64(rw)*0.45)
		if x < rx1+2 || x > rx2-2 {
			continue
		}
		norm := int(math.Mod(deg+3600.0, 360.0))
		isMajor := (norm % 45 == 0)
		isMid := (!isMajor && norm%15 == 0)

		tickH := 3
		if isMajor {
			tickH = 7
		} else if isMid {
			tickH = 5
		}
		s.drawVLine(x, ry+rh-tickH, ry+rh, 160, 180, 70)

		if isMajor || isMid {
			lbl := fmt.Sprintf("%d", norm)
			if norm == 0 {
				lbl = "N"
			} else if norm == 45 {
				lbl = "NE"
			} else if norm == 90 {
				lbl = "E"
			} else if norm == 135 {
				lbl = "SE"
			} else if norm == 180 {
				lbl = "S"
			} else if norm == 225 {
				lbl = "SW"
			} else if norm == 270 {
				lbl = "W"
			} else if norm == 315 {
				lbl = "NW"
			}
			yColor := byte(180)
			if norm == 0 {
				yColor = 220
			}
			s.drawString(x-len(lbl)*3, ry+2, lbl, yColor, 128, 128)
		}
	}

	// 中心指示黄色倒三角
	cx := w / 2
	s.drawHLine(cx-3, cx+4, ry, 210, 30, 150)
	s.drawHLine(cx-2, cx+3, ry+1, 210, 30, 150)
	s.drawHLine(cx-1, cx+2, ry+2, 210, 30, 150)
	s.setPixel(cx, ry+3, 210, 30, 150)
}

// drawCrosshair 绘制战术十字瞄准准星。
func (s *PTZSource) drawCrosshair(cx, cy int, locked bool) {
	yColor := byte(140)
	uColor := byte(180)
	vColor := byte(70)
	if locked {
		yColor = 190
		uColor = 40
		vColor = 40
	}

	// 十字准星（带中心空隙）
	s.drawHLine(cx-30, cx-6, cy, yColor, uColor, vColor)
	s.drawHLine(cx+6, cx+30, cy, yColor, uColor, vColor)
	s.drawVLine(cx, cy-22, cy-6, yColor, uColor, vColor)
	s.drawVLine(cx, cy+6, cy+22, yColor, uColor, vColor)

	// 毫弧度微调刻度
	s.drawVLine(cx-18, cy-2, cy+3, yColor, uColor, vColor)
	s.drawVLine(cx+18, cy-2, cy+3, yColor, uColor, vColor)
	s.drawHLine(cx-2, cx+3, cy-12, yColor, uColor, vColor)
	s.drawHLine(cx-2, cx+3, cy+12, yColor, uColor, vColor)

	// 四角战术取景框
	pad := 10
	blen := 12
	// 左上
	s.drawHLine(pad, pad+blen, pad, 120, 128, 128)
	s.drawVLine(pad, pad, pad+blen, 120, 128, 128)
	// 右上
	s.drawHLine(s.w-pad-blen, s.w-pad, pad, 120, 128, 128)
	s.drawVLine(s.w-pad, pad, pad+blen, 120, 128, 128)
	// 左下
	s.drawHLine(pad, pad+blen, s.h-pad, 120, 128, 128)
	s.drawVLine(pad, s.h-pad-blen, s.h-pad, 120, 128, 128)
	// 右下
	s.drawHLine(s.w-pad-blen, s.w-pad, s.h-pad, 120, 128, 128)
	s.drawVLine(s.w-pad, s.h-pad-blen, s.h-pad, 120, 128, 128)
}

// drawOSD 烧录字符水印。
func (s *PTZSource) drawOSD(ptz PTZInfo, locked *ptzLandmark) {
	// 左上角：通道名称与状态
	chTag := s.chID
	if chTag == "" {
		chTag = "CAM-01"
	}
	s.drawString(14, 12, fmt.Sprintf("%s [LIVE PTZ]", chTag), 210, 50, 50) // 荧光绿

	// 左下角：实时安全时间戳（含毫秒）
	now := time.Now()
	timeStr := now.Format("2006-01-02 15:04:05.000")
	s.drawString(14, s.h-22, timeStr, 220, 128, 128)

	// 右下角：实时姿态遥测
	telemetry := fmt.Sprintf("P:%.1f T:%+.1f Z:%.1fX", ptz.Pan, ptz.Tilt, ptz.Zoom)
	if ptz.Iris > 0 && ptz.Iris != 50.0 {
		telemetry += fmt.Sprintf(" I:%.0f%%", ptz.Iris)
	}
	if ptz.Focus > 0 && ptz.Focus != 50.0 {
		telemetry += fmt.Sprintf(" F:%.0f%%", ptz.Focus)
	}
	s.drawString(s.w-len(telemetry)*8-14, s.h-22, telemetry, 220, 180, 60)

	// 中下方：当前状态 / 预置位 / 目标锁定
	if locked != nil {
		lockStr := fmt.Sprintf("LOCK: %s (%.0f DEG / %.0fM)", locked.name, locked.az, locked.dist)
		s.drawString(s.w/2-len(lockStr)*4, s.h-22, lockStr, 210, 40, 40)
	} else if ptz.ActivePresetID > 0 {
		psStr := fmt.Sprintf("PRESET: #%d %s", ptz.ActivePresetID, ptz.ActivePreset)
		s.drawString(s.w/2-len(psStr)*4, s.h-22, psStr, 200, 190, 70)
	} else if ptz.IsMoving {
		s.drawString(s.w/2-40, s.h-22, "PTZ ROTATING...", 220, 70, 210)
	}
}

// encodeIPCMFrame 将 YUV 画布直接编码为 H.264 I_PCM 宏块切片。
func (s *PTZSource) encodeIPCMFrame() []byte {
	w := s.w
	h := s.h
	mbsX := w / 16
	mbsY := h / 16

	e2 := newBitWriter()
	e2.writeUE(0) // first_mb_in_slice = 0
	e2.writeUE(7) // slice_type = 7 (I_ALL)
	e2.writeUE(0) // pic_parameter_set_id = 0
	e2.writeUint(uint(s.frameIdx&0xF), 4)
	e2.writeUE(0) // idr_pic_id = 0
	e2.writeUint(uint((s.frameIdx*2)&0xF), 4)
	e2.writeBit(0) // redundant_pic_cnt_present_flag
	e2.writeBit(0) // no_output_of_prior_pics_flag
	e2.writeSE(0) // slice_qp_delta

	for my := 0; my < mbsY; my++ {
		for mx := 0; mx < mbsX; mx++ {
			e2.writeUE(25) // mb_type = 25 (I_PCM)
			e2.byteAlign()

			// 256 字节 Y 数据（16x16）
			for r := 0; r < 16; r++ {
				rowOffset := (my*16+r)*w + mx*16
				e2.buf = append(e2.buf, s.yBuf[rowOffset:rowOffset+16]...)
			}
			// 64 字节 Cb 数据（8x8）
			for r := 0; r < 8; r++ {
				rowOffset := (my*8+r)*(w/2) + mx*8
				e2.buf = append(e2.buf, s.cbBuf[rowOffset:rowOffset+8]...)
			}
			// 64 字节 Cr 数据（8x8）
			for r := 0; r < 8; r++ {
				rowOffset := (my*8+r)*(w/2) + mx*8
				e2.buf = append(e2.buf, s.crBuf[rowOffset:rowOffset+8]...)
			}
		}
	}
	e2.rbspTrailing()
	return annexB(5, e2.bytes())
}

// ---------------- 基础绘图函数 ----------------

func (s *PTZSource) setPixel(x, y int, yVal, uVal, vVal byte) {
	if x < 0 || x >= s.w || y < 0 || y >= s.h {
		return
	}
	s.yBuf[y*s.w+x] = yVal
	cx := x / 2
	cy := y / 2
	uvIdx := cy*(s.w/2) + cx
	s.cbBuf[uvIdx] = uVal
	s.crBuf[uvIdx] = vVal
}

func (s *PTZSource) fillRect(x, y, w, h int, yVal, uVal, vVal byte) {
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > s.w {
		w = s.w - x
	}
	if y+h > s.h {
		h = s.h - y
	}
	if w <= 0 || h <= 0 {
		return
	}

	for r := 0; r < h; r++ {
		rowStart := (y+r)*s.w + x
		for c := 0; c < w; c++ {
			s.yBuf[rowStart+c] = yVal
		}
	}

	cx1 := x / 2
	cx2 := (x + w + 1) / 2
	cy1 := y / 2
	cy2 := (y + h + 1) / 2
	if cx2 > s.w/2 {
		cx2 = s.w / 2
	}
	if cy2 > s.h/2 {
		cy2 = s.h / 2
	}

	uvW := cx2 - cx1
	for r := cy1; r < cy2; r++ {
		uvRow := r*(s.w/2) + cx1
		for c := 0; c < uvW; c++ {
			s.cbBuf[uvRow+c] = uVal
			s.crBuf[uvRow+c] = vVal
		}
	}
}

func (s *PTZSource) drawHLine(x1, x2, y int, yVal, uVal, vVal byte) {
	if y < 0 || y >= s.h {
		return
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if x1 < 0 {
		x1 = 0
	}
	if x2 >= s.w {
		x2 = s.w - 1
	}
	for x := x1; x <= x2; x++ {
		s.setPixel(x, y, yVal, uVal, vVal)
	}
}

func (s *PTZSource) drawVLine(x, y1, y2 int, yVal, uVal, vVal byte) {
	if x < 0 || x >= s.w {
		return
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if y1 < 0 {
		y1 = 0
	}
	if y2 >= s.h {
		y2 = s.h - 1
	}
	for y := y1; y <= y2; y++ {
		s.setPixel(x, y, yVal, uVal, vVal)
	}
}

func (s *PTZSource) drawLine(x1, y1, x2, y2 int, yVal, uVal, vVal byte) {
	dx := int(math.Abs(float64(x2 - x1)))
	dy := int(math.Abs(float64(y2 - y1)))
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		s.setPixel(x1, y1, yVal, uVal, vVal)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// ---------------- ASCII 8x10 点阵字库 ----------------

func (s *PTZSource) drawString(x, y int, str string, yVal, uVal, vVal byte) {
	curX := x
	for i := 0; i < len(str); i++ {
		ch := str[i]
		s.drawChar(curX, y, ch, yVal, uVal, vVal)
		curX += 7
	}
}

func (s *PTZSource) drawChar(x, y int, ch byte, yVal, uVal, vVal byte) {
	glyph, ok := font8x10[ch]
	if !ok {
		glyph = font8x10['?']
	}
	for row := 0; row < 10; row++ {
		bits := glyph[row]
		for col := 0; col < 6; col++ {
			if (bits>>(5-col))&1 == 1 {
				s.setPixel(x+col, y+row, yVal, uVal, vVal)
			}
		}
	}
}

func normDegDiff(a, b float64) float64 {
	d := a - b
	for d < -180 {
		d += 360
	}
	for d > 180 {
		d -= 360
	}
	return d
}

// 6x10 点阵字体字典 (宽6高10，低6位有效)
var font8x10 = map[byte][10]byte{
	' ': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	'.': {0, 0, 0, 0, 0, 0, 0, 0, 0x18, 0x18},
	':': {0, 0, 0x18, 0x18, 0, 0, 0x18, 0x18, 0, 0},
	'-': {0, 0, 0, 0, 0x3E, 0x3E, 0, 0, 0, 0},
	'+': {0, 0, 0x08, 0x08, 0x3E, 0x08, 0x08, 0, 0, 0},
	'/': {0, 0x02, 0x06, 0x0C, 0x18, 0x30, 0x20, 0, 0, 0},
	'[': {0, 0x1E, 0x18, 0x18, 0x18, 0x18, 0x18, 0x1E, 0, 0},
	']': {0, 0x3C, 0x0C, 0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0, 0},
	'#': {0, 0x14, 0x3E, 0x14, 0x14, 0x3E, 0x14, 0, 0, 0},
	'?': {0, 0x1C, 0x22, 0x02, 0x0C, 0x10, 0, 0x10, 0, 0},
	'0': {0, 0x1C, 0x22, 0x26, 0x2A, 0x32, 0x22, 0x1C, 0, 0},
	'1': {0, 0x08, 0x18, 0x28, 0x08, 0x08, 0x08, 0x3E, 0, 0},
	'2': {0, 0x1C, 0x22, 0x02, 0x0C, 0x10, 0x20, 0x3E, 0, 0},
	'3': {0, 0x3E, 0x04, 0x08, 0x1C, 0x02, 0x22, 0x1C, 0, 0},
	'4': {0, 0x04, 0x0C, 0x14, 0x24, 0x3E, 0x04, 0x04, 0, 0},
	'5': {0, 0x3E, 0x20, 0x3C, 0x02, 0x02, 0x22, 0x1C, 0, 0},
	'6': {0, 0x1C, 0x20, 0x20, 0x3C, 0x22, 0x22, 0x1C, 0, 0},
	'7': {0, 0x3E, 0x02, 0x04, 0x08, 0x10, 0x10, 0x10, 0, 0},
	'8': {0, 0x1C, 0x22, 0x22, 0x1C, 0x22, 0x22, 0x1C, 0, 0},
	'9': {0, 0x1C, 0x22, 0x22, 0x1E, 0x02, 0x04, 0x38, 0, 0},
	'A': {0, 0x1C, 0x22, 0x22, 0x3E, 0x22, 0x22, 0x22, 0, 0},
	'B': {0, 0x3C, 0x22, 0x22, 0x3C, 0x22, 0x22, 0x3C, 0, 0},
	'C': {0, 0x1C, 0x22, 0x20, 0x20, 0x20, 0x22, 0x1C, 0, 0},
	'D': {0, 0x38, 0x24, 0x22, 0x22, 0x22, 0x24, 0x38, 0, 0},
	'E': {0, 0x3E, 0x20, 0x20, 0x3C, 0x20, 0x20, 0x3E, 0, 0},
	'F': {0, 0x3E, 0x20, 0x20, 0x3C, 0x20, 0x20, 0x20, 0, 0},
	'G': {0, 0x1C, 0x22, 0x20, 0x2E, 0x22, 0x22, 0x1C, 0, 0},
	'H': {0, 0x22, 0x22, 0x22, 0x3E, 0x22, 0x22, 0x22, 0, 0},
	'I': {0, 0x1C, 0x08, 0x08, 0x08, 0x08, 0x08, 0x1C, 0, 0},
	'J': {0, 0x06, 0x02, 0x02, 0x02, 0x02, 0x22, 0x1C, 0, 0},
	'K': {0, 0x22, 0x24, 0x28, 0x30, 0x28, 0x24, 0x22, 0, 0},
	'L': {0, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3E, 0, 0},
	'M': {0, 0x22, 0x36, 0x2A, 0x22, 0x22, 0x22, 0x22, 0, 0},
	'N': {0, 0x22, 0x32, 0x2A, 0x26, 0x22, 0x22, 0x22, 0, 0},
	'O': {0, 0x1C, 0x22, 0x22, 0x22, 0x22, 0x22, 0x1C, 0, 0},
	'P': {0, 0x3C, 0x22, 0x22, 0x3C, 0x20, 0x20, 0x20, 0, 0},
	'Q': {0, 0x1C, 0x22, 0x22, 0x22, 0x2A, 0x24, 0x1A, 0, 0},
	'R': {0, 0x3C, 0x22, 0x22, 0x3C, 0x28, 0x24, 0x22, 0, 0},
	'S': {0, 0x1C, 0x22, 0x20, 0x1C, 0x02, 0x22, 0x1C, 0, 0},
	'T': {0, 0x3E, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0, 0},
	'U': {0, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x1C, 0, 0},
	'V': {0, 0x22, 0x22, 0x22, 0x22, 0x22, 0x14, 0x08, 0, 0},
	'W': {0, 0x22, 0x22, 0x22, 0x2A, 0x2A, 0x36, 0x22, 0, 0},
	'X': {0, 0x22, 0x22, 0x14, 0x08, 0x14, 0x22, 0x22, 0, 0},
	'Y': {0, 0x22, 0x22, 0x14, 0x08, 0x08, 0x08, 0x08, 0, 0},
	'Z': {0, 0x3E, 0x02, 0x04, 0x08, 0x10, 0x20, 0x3E, 0, 0},
}
