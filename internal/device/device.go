package device

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/gb28181"
	"github.com/local/gb28181-device/internal/media"
	"github.com/local/gb28181-device/internal/sip"
)

type Device struct {
	cfg *config.Config
	// 运行时修改通道/媒体绑定时用
	cfgMu sync.RWMutex
	ua    *sip.UA
	ms    *media.SessionManager

	sn        int32
	kaFail    int32
	stopCh    chan struct{}
	stopped   atomic.Bool
	srcClose  func() error
	startedAt time.Time
	logs      *LogBuffer

	OnConfigChanged func(cfg *config.Config)
}

func New(cfg *config.Config) *Device {
	ua := sip.NewUA(
		cfg.SIP.LocalIP, cfg.SIP.LocalPort,
		cfg.SIP.ServerIP, cfg.SIP.ServerPort,
		cfg.SIP.Transport,
		cfg.SIP.Username, cfg.SIP.Password,
	)
	d := &Device{
		cfg:    cfg,
		ua:     ua,
		stopCh: make(chan struct{}),
		logs:   NewLogBuffer(800),
	}
	srcFactory := func(channelID string) (media.H264Source, error) {
		d.cfgRLock()
		v := cfg.Media.OptionsFor(channelID)
		d.cfgRUnlock()
		log.Printf("[media] channel %s source=%s mp4=%s h264=%s",
			channelID, v.Kind, v.MP4, v.H264)
		return media.NewSource(media.SourceOptions{
			Kind:   v.Kind,
			H264:   v.H264,
			MP4:    v.MP4,
			Width:  v.Width,
			Height: v.Height,
			FPS:    v.FPS,
		})
	}
	d.ms = media.NewSessionManager(cfg.Media.LocalIP, cfg.Media.FPS, cfg.Media.RTPPayloadMax, srcFactory)
	d.ms.OnComplete = func(ch, callID string) {
		_ = d.SendMediaStatusNotify(ch, "121")
	}
	return d
}

func (d *Device) nextSN() string {
	return strconv.Itoa(int(atomic.AddInt32(&d.sn, 1)))
}

func (d *Device) Start() error {
	d.startedAt = time.Now()
	if err := d.ua.Start(); err != nil {
		return err
	}

	// 入站 SIP 请求
	d.ua.Handle("MESSAGE", d.onMessage)
	d.ua.Handle("INVITE", d.onInvite)
	d.ua.Handle("BYE", d.onBye)
	d.ua.Handle("INFO", d.onInfo)
	d.ua.Handle("ACK", d.onAck)
	d.ua.Handle("OPTIONS", d.onOptions)
	d.ua.Handle("SUBSCRIBE", d.onSubscribe)
	d.ua.Handle("NOTIFY", d.onNotify)

	// 注册循环
	go d.registerLoop()
	// 心跳
	go d.keepaliveLoop()

	return nil
}

func (d *Device) Stop() {
	if d.stopped.Swap(true) {
		return
	}
	close(d.stopCh)
	d.ms.StopAll()
	_ = d.ua.Unregister()
	d.ua.Close()
}

func (d *Device) Status() Status {
	sess := d.ms.ListSessions()
	sessions := make([]any, 0, len(sess))
	for _, x := range sess {
		sessions = append(sessions, x)
	}

	d.cfgRLock()
	chs := make([]ChannelStatus, 0, len(d.cfg.Device.Channels))
	for _, ch := range d.cfg.Device.Channels {
		v := d.cfg.Media.OptionsFor(ch.ID)
		chs = append(chs, ChannelStatus{
			ID:     ch.ID,
			Name:   ch.Name,
			Status: ch.Status,
			MP4:    v.MP4,
			Source: v.Kind,
			H264:   v.H264,
		})
	}
	mode := d.cfg.Media.Mode
	src := d.cfg.Media.Source
	deviceID := d.cfg.Device.ID
	deviceName := d.cfg.Device.Name
	server := fmt.Sprintf("%s:%d", d.cfg.SIP.ServerIP, d.cfg.SIP.ServerPort)
	local := fmt.Sprintf("%s:%d", d.cfg.SIP.LocalIP, d.cfg.SIP.LocalPort)
	transport := d.cfg.SIP.Transport
	d.cfgRUnlock()

	up := int64(0)
	if !d.startedAt.IsZero() {
		up = int64(time.Since(d.startedAt).Seconds())
	}
	return Status{
		Registered:  d.ua.IsRegistered(),
		DeviceID:    deviceID,
		DeviceName:  deviceName,
		Server:      server,
		Local:       local,
		Transport:   transport,
		MediaMode:   mode,
		MediaSource: src,
		Channels:    chs,
		Sessions:    sessions,
		UptimeSec:   up,
		StartedAt:   d.startedAt,
	}
}

func (d *Device) RegisterNow() error {
	return d.ua.Register(d.cfg.SIP.Expires)
}

func (d *Device) UnregisterNow() error {
	return d.ua.Unregister()
}

func (d *Device) KeepaliveNow() error {
	return d.sendKeepalive()
}

func (d *Device) StopSession(callID string) {
	d.ms.StopByCallID(callID)
}

func (d *Device) Logs(n int) []string {
	if d.logs == nil {
		return nil
	}
	return d.logs.Tail(n)
}

func (d *Device) AppendLog(line string) {
	if d.logs != nil {
		d.logs.Append(line)
	}
}

func (d *Device) registerLoop() {
	expires := d.cfg.SIP.Expires
	interval := time.Duration(expires) * time.Second * 2 / 3
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}
	for {
		if err := d.ua.Register(expires); err != nil {
			log.Printf("[device] register failed: %v, retry in 5s", err)
			select {
			case <-d.stopCh:
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		log.Printf("[device] registered as %s", d.cfg.Device.ID)
		select {
		case <-d.stopCh:
			return
		case <-time.After(interval):
		}
	}
}

func (d *Device) keepaliveLoop() {
	interval := time.Duration(d.cfg.SIP.KeepaliveInterval) * time.Second
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-d.stopCh:
			return
		case <-t.C:
			if !d.ua.IsRegistered() {
				continue
			}
			if err := d.sendKeepalive(); err != nil {
				n := atomic.AddInt32(&d.kaFail, 1)
				log.Printf("[device] keepalive failed (%d): %v", n, err)
				if int(n) >= d.cfg.SIP.KeepaliveTimeoutCount {
					log.Printf("[device] keepalive timeout, will re-register")
					atomic.StoreInt32(&d.kaFail, 0)
					// 触发重注册：直接再 Register 一次
					go func() {
						if err := d.ua.Register(d.cfg.SIP.Expires); err != nil {
							log.Printf("[device] re-register failed: %v", err)
						}
					}()
				}
			} else {
				atomic.StoreInt32(&d.kaFail, 0)
			}
		}
	}
}

func (d *Device) sendKeepalive() error {
	body, err := gb28181.MarshalXML(gb28181.KeepaliveNotify{
		CmdType:  "Keepalive",
		SN:       d.nextSN(),
		DeviceID: d.cfg.Device.ID,
		Status:   "OK",
	})
	if err != nil {
		return err
	}
	_, err = d.ua.SendMessage(body, "Application/MANSCDP+xml")
	return err
}

// ---------- 入站请求 ----------

func (d *Device) onOptions(m *sip.Message, src net.Addr) {
	_ = d.ua.Reply(m, src, 200, "OK", nil, "")
}

func (d *Device) onAck(m *sip.Message, src net.Addr) {
	// INVITE 的 ACK，无需回复
}

func (d *Device) onNotify(m *sip.Message, src net.Addr) {
	_ = d.ua.Reply(m, src, 200, "OK", nil, "")
}

func (d *Device) onSubscribe(m *sip.Message, src net.Addr) {
	// 平台订阅报警/目录等，回复 200
	resp := d.ua.Reply(m, src, 200, "OK", nil, "")
	_ = resp
	// 有些平台期望 Event 头在响应里
}

func (d *Device) onInfo(m *sip.Message, src net.Addr) {
	ct := strings.ToLower(m.GetHeader("Content-Type"))
	bodyStr := strings.TrimSpace(string(m.Body))
	isMANSRTSP := strings.Contains(ct, "mansrtsp") ||
		strings.HasPrefix(bodyStr, "PLAY") ||
		strings.HasPrefix(bodyStr, "PAUSE") ||
		strings.HasPrefix(bodyStr, "TEARDOWN")

	if isMANSRTSP {
		req, err := gb28181.ParseMANSRTSP(m.Body)
		if err != nil {
			log.Printf("[device] bad MANSRTSP: %v", err)
			_ = d.ua.Reply(m, src, 400, "Bad Request", nil, "")
			return
		}

		callID := m.CallID()
		sess := d.ms.GetSession(callID)
		scale := 1.0
		rangeNPT := 0.0
		isPause := false

		if sess != nil {
			switch req.Method {
			case "PLAY":
				if req.HasScale {
					sess.SetScale(req.Scale)
				}
				if req.HasRange {
					actualSec := sess.Seek(req.RangeNPT)
					rangeNPT = actualSec
				} else {
					rangeNPT = sess.CurrentOffset()
				}
				sess.Resume()
				scale = sess.GetScale()
			case "PAUSE":
				sess.Pause()
				isPause = true
			case "TEARDOWN":
				sess.Stop()
				isPause = true
			}
			log.Printf("[gb] MANSRTSP %s callID=%s scale=%.2f range=%.2f pause=%v",
				req.Method, callID, scale, rangeNPT, isPause)
		} else {
			log.Printf("[gb] MANSRTSP %s callID=%s (session not found, replying OK)", req.Method, callID)
		}

		respBody := gb28181.BuildMANSRTSPResponse(req.CSeq, scale, rangeNPT, isPause)
		_ = d.ua.Reply(m, src, 200, "OK", respBody, "Application/MANSRTSP")
		return
	}

	// 其他 INFO（如语音对讲等）
	log.Printf("[device] INFO content-type=%s body=%s", m.GetHeader("Content-Type"), truncate(string(m.Body), 200))
	_ = d.ua.Reply(m, src, 200, "OK", nil, "")
}

func (d *Device) onMessage(m *sip.Message, src net.Addr) {
	// 先 200
	if err := d.ua.Reply(m, src, 200, "OK", nil, ""); err != nil {
		log.Printf("[device] reply MESSAGE 200 failed: %v", err)
	}

	ct := strings.ToLower(m.GetHeader("Content-Type"))
	if strings.Contains(ct, "xml") || len(m.Body) > 0 {
		root, err := gb28181.ParseRoot(m.Body)
		if err != nil {
			log.Printf("[device] bad manscdp body: %v", err)
			return
		}
		d.handleMANSCDP(root, m, src)
	}
}

func (d *Device) handleMANSCDP(root *gb28181.Root, m *sip.Message, src net.Addr) {
	cmd := strings.ToLower(root.CmdType)
	log.Printf("[gb] recv %s SN=%s DeviceID=%s from=%s", root.CmdType, root.SN, root.DeviceID, src)

	switch cmd {
	case "catalog":
		d.respCatalog(root, m, src)
	case "deviceinfo":
		d.respDeviceInfo(root, m, src)
	case "devicestatus":
		d.respDeviceStatus(root, m, src)
	case "devicecontrol":
		d.onDeviceControl(root, m, src)
	case "recordinfo":
		d.respRecordInfo(root, m, src)
	case "configdownload":
		// 简单回复空
		d.replyMANSCDP(root, m, src, map[string]any{
			"CmdType": root.CmdType,
			"SN":      root.SN,
			"DeviceID": d.cfg.Device.ID,
			"Result":  "OK",
		})
	default:
		log.Printf("[gb] unhandled CmdType=%s", root.CmdType)
	}
}

func (d *Device) replyMANSCDP(root *gb28181.Root, req *sip.Message, src net.Addr, fields map[string]any) {
	// 构造 Response 根
	type response struct {
		XMLName  struct{} `xml:"Response"`
		CmdType  string   `xml:"CmdType"`
		SN       string   `xml:"SN"`
		DeviceID string   `xml:"DeviceID"`
	}
	// 直接用 map 不支持 xml 根名，改为手写更通用
	body := buildXMLResponse(fields)
	// 平台期望响应走 SIP MESSAGE（不是对原 MESSAGE 的 200 里带 body）
	// GB28181：查询响应通过新的 MESSAGE 发给平台
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] send response failed: %v", err)
	}
}

func buildXMLResponse(fields map[string]any) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\r\n")
	b.WriteString("<Response>\r\n")
	// 稳定字段顺序
	order := []string{"CmdType", "SN", "DeviceID", "Result", "SumNum", "DeviceName", "Manufacturer", "Model", "Firmware", "Channel", "Online", "Status", "Encode", "Record", "DeviceTime"}
	written := map[string]bool{}
	for _, k := range order {
		if v, ok := fields[k]; ok {
			fmt.Fprintf(&b, "  <%s>%v</%s>\r\n", k, v, k)
			written[k] = true
		}
	}
	for k, v := range fields {
		if written[k] || k == "DeviceList" || k == "RecordList" {
			continue
		}
		fmt.Fprintf(&b, "  <%s>%v</%s>\r\n", k, v, k)
	}
	if dl, ok := fields["DeviceList"].(string); ok {
		b.WriteString(dl)
	}
	if rl, ok := fields["RecordList"].(string); ok {
		b.WriteString(rl)
	}
	b.WriteString("</Response>\r\n")
	return []byte(b.String())
}

func (d *Device) respCatalog(root *gb28181.Root, req *sip.Message, src net.Addr) {
	d.cfgRLock()
	items := make([]gb28181.CatalogItem, 0, len(d.cfg.Device.Channels))
	for _, ch := range d.cfg.Device.Channels {
		items = append(items, gb28181.CatalogItem{
			DeviceID:     ch.ID,
			Name:         ch.Name,
			Manufacturer: ch.Manufacturer,
			Model:        ch.Model,
			Owner:        "Owner",
			CivilCode:    ch.CivilCode,
			Address:      ch.Address,
			Parental:     ch.Parental,
			ParentID:     ch.ParentID,
			SafetyWay:    ch.SafetyWay,
			RegisterWay:  ch.RegisterWay,
			Secrecy:      ch.Secrecy,
			Status:       ch.Status,
		})
	}
	d.cfgRUnlock()

	var listB strings.Builder
	fmt.Fprintf(&listB, "  <DeviceList Num=\"%d\">\r\n", len(items))
	for _, it := range items {
		listB.WriteString("    <Item>\r\n")
		fmt.Fprintf(&listB, "      <DeviceID>%s</DeviceID>\r\n", it.DeviceID)
		fmt.Fprintf(&listB, "      <Name>%s</Name>\r\n", it.Name)
		fmt.Fprintf(&listB, "      <Manufacturer>%s</Manufacturer>\r\n", it.Manufacturer)
		fmt.Fprintf(&listB, "      <Model>%s</Model>\r\n", it.Model)
		fmt.Fprintf(&listB, "      <Owner>%s</Owner>\r\n", it.Owner)
		if it.CivilCode != "" {
			fmt.Fprintf(&listB, "      <CivilCode>%s</CivilCode>\r\n", it.CivilCode)
		}
		if it.Address != "" {
			fmt.Fprintf(&listB, "      <Address>%s</Address>\r\n", it.Address)
		}
		fmt.Fprintf(&listB, "      <Parental>%d</Parental>\r\n", it.Parental)
		fmt.Fprintf(&listB, "      <ParentID>%s</ParentID>\r\n", it.ParentID)
		fmt.Fprintf(&listB, "      <SafetyWay>%d</SafetyWay>\r\n", it.SafetyWay)
		fmt.Fprintf(&listB, "      <RegisterWay>%d</RegisterWay>\r\n", it.RegisterWay)
		fmt.Fprintf(&listB, "      <Secrecy>%d</Secrecy>\r\n", it.Secrecy)
		fmt.Fprintf(&listB, "      <Status>%s</Status>\r\n", it.Status)
		listB.WriteString("    </Item>\r\n")
	}
	listB.WriteString("  </DeviceList>\r\n")

	body := buildXMLResponse(map[string]any{
		"CmdType":    "Catalog",
		"SN":         root.SN,
		"DeviceID":   root.DeviceID, // 平台请求里的 DeviceID，通常是设备 ID
		"SumNum":     len(items),
		"DeviceList": listB.String(),
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] catalog response failed: %v", err)
	} else {
		log.Printf("[gb] catalog response sent, channels=%d", len(items))
	}
}

func (d *Device) respDeviceInfo(root *gb28181.Root, req *sip.Message, src net.Addr) {
	body := buildXMLResponse(map[string]any{
		"CmdType":      "DeviceInfo",
		"SN":           root.SN,
		"DeviceID":     d.cfg.Device.ID,
		"DeviceName":   d.cfg.Device.Name,
		"Manufacturer": d.cfg.Device.Manufacturer,
		"Model":        d.cfg.Device.Model,
		"Firmware":     d.cfg.Device.Firmware,
		"Channel":      len(d.cfg.Device.Channels),
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] deviceinfo response failed: %v", err)
	} else {
		log.Printf("[gb] deviceinfo response sent")
	}
}

func (d *Device) respDeviceStatus(root *gb28181.Root, req *sip.Message, src net.Addr) {
	body := buildXMLResponse(map[string]any{
		"CmdType":    "DeviceStatus",
		"SN":         root.SN,
		"DeviceID":   d.cfg.Device.ID,
		"Result":     "OK",
		"Online":     "ONLINE",
		"Status":     "OK",
		"Encode":     "ON",
		"Record":     "OFF",
		"DeviceTime": time.Now().Format("2006-01-02T15:04:05"),
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] devicestatus response failed: %v", err)
	}
}

func (d *Device) onDeviceControl(root *gb28181.Root, req *sip.Message, src net.Addr) {
	var ctrl gb28181.DeviceControlReq
	if err := unmarshalBody(req.Body, &ctrl); err != nil {
		log.Printf("[gb] devicecontrol parse: %v", err)
	}
	action := gb28181.DecodePTZ(ctrl.PTZCmd)
	log.Printf("[gb] DeviceControl CmdType=%s DeviceID=%s PTZ=%s => %s",
		root.CmdType, root.DeviceID, ctrl.PTZCmd, action)
	// 控制类一般无需 MESSAGE 回业务响应，仅 SIP 200（已在 onMessage 回过）
}

func (d *Device) respRecordInfo(root *gb28181.Root, req *sip.Message, src net.Addr) {
	var query gb28181.RecordInfoReq
	if err := unmarshalBody(req.Body, &query); err != nil {
		log.Printf("[gb] recordinfo parse query failed: %v", err)
	}

	channelID := root.DeviceID
	if query.DeviceID != "" {
		channelID = query.DeviceID
	}
	channelName := "通道录像"
	d.cfgRLock()
	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == channelID {
			channelName = ch.Name
			break
		}
	}
	d.cfgRUnlock()

	startTimeStr := query.StartTime
	endTimeStr := query.EndTime
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if t, err := time.Parse("2006-01-02T15:04:05", startTimeStr); err == nil && !t.IsZero() {
		startTime = t
	}
	if t, err := time.Parse("2006-01-02T15:04:05", endTimeStr); err == nil && !t.IsZero() {
		endTime = t
	}

	// 动态生成录像切片列表：在起止时间范围内切成若个模拟录像段（每 2 小时一段）
	items := make([]gb28181.RecordItem, 0)
	step := 2 * time.Hour
	cur := startTime
	for cur.Before(endTime) {
		segEnd := cur.Add(step)
		if segEnd.After(endTime) {
			segEnd = endTime
		}
		items = append(items, gb28181.RecordItem{
			DeviceID:   channelID,
			Name:       channelName,
			FilePath:   fmt.Sprintf("/record/%s/%s_%s.mp4", cur.Format("20060102"), cur.Format("150405"), segEnd.Format("150405")),
			Address:    "LocalDisk",
			StartTime:  cur.Format("2006-01-02T15:04:05"),
			EndTime:    segEnd.Format("2006-01-02T15:04:05"),
			Secrecy:    0,
			Type:       "time",
			RecorderID: d.cfg.Device.ID,
			FileSize:   "104857600",
		})
		cur = segEnd
	}

	var listB strings.Builder
	fmt.Fprintf(&listB, "  <RecordList Num=\"%d\">\r\n", len(items))
	for _, it := range items {
		listB.WriteString("    <Item>\r\n")
		fmt.Fprintf(&listB, "      <DeviceID>%s</DeviceID>\r\n", it.DeviceID)
		fmt.Fprintf(&listB, "      <Name>%s</Name>\r\n", it.Name)
		fmt.Fprintf(&listB, "      <FilePath>%s</FilePath>\r\n", it.FilePath)
		fmt.Fprintf(&listB, "      <Address>%s</Address>\r\n", it.Address)
		fmt.Fprintf(&listB, "      <StartTime>%s</StartTime>\r\n", it.StartTime)
		fmt.Fprintf(&listB, "      <EndTime>%s</EndTime>\r\n", it.EndTime)
		fmt.Fprintf(&listB, "      <Secrecy>%d</Secrecy>\r\n", it.Secrecy)
		fmt.Fprintf(&listB, "      <Type>%s</Type>\r\n", it.Type)
		fmt.Fprintf(&listB, "      <RecorderID>%s</RecorderID>\r\n", it.RecorderID)
		fmt.Fprintf(&listB, "      <FileSize>%s</FileSize>\r\n", it.FileSize)
		listB.WriteString("    </Item>\r\n")
	}
	listB.WriteString("  </RecordList>\r\n")

	body := buildXMLResponse(map[string]any{
		"CmdType":    "RecordInfo",
		"SN":         root.SN,
		"DeviceID":   channelID,
		"Name":       channelName,
		"SumNum":     len(items),
		"RecordList": listB.String(),
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] recordinfo response failed: %v", err)
	} else {
		log.Printf("[gb] recordinfo response sent ch=%s records=%d", channelID, len(items))
	}
}

func (d *Device) SendMediaStatusNotify(channelID, notifyType string) error {
	if notifyType == "" {
		notifyType = "121"
	}
	body := buildXMLNotify(map[string]any{
		"CmdType":    "MediaStatus",
		"SN":         d.nextSN(),
		"DeviceID":   channelID,
		"NotifyType": notifyType,
	})
	_, err := d.ua.SendMessage(body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] send MediaStatus %s failed: %v", notifyType, err)
		return err
	}
	log.Printf("[gb] MediaStatus Notify sent ch=%s type=%s", channelID, notifyType)
	return nil
}

func (d *Device) onInvite(m *sip.Message, src net.Addr) {
	sdp := string(m.Body)
	recv, err := media.ParseSDP(sdp)
	if err != nil {
		log.Printf("[media] INVITE rejected: %v", err)
		_ = d.ua.Reply(m, src, 488, "Not Acceptable Here", nil, "")
		return
	}

	// 从 SDP o= 或 Subject 取通道
	channelID := extractChannelID(m, recv)
	callID := m.CallID()
	sessName := strings.ToLower(recv.SessionName)
	isPlayback := sessName == "playback" || sessName == "download"

	log.Printf("[media] INVITE channel=%s s=%s media=%s:%d ssrc=%s t=%s %s",
		channelID, recv.SessionName, recv.IP, recv.VideoPort, recv.SSRC, recv.StartTime, recv.EndTime)

	var answer string
	if isPlayback {
		initOffset := d.calcPlaybackOffset(recv)
		answer, err = d.ms.StartPlayback(channelID, callID, recv, initOffset)
	} else {
		answer, err = d.ms.StartLive(channelID, callID, recv)
	}
	if err != nil {
		log.Printf("[media] start session failed: %v", err)
		_ = d.ua.Reply(m, src, 500, "Internal Server Error", nil, "")
		return
	}

	// 200 OK + SDP
	resp := sip.NewResponse(200, "OK")
	// 复制 Via/From/To/Call-ID/CSeq
	if v := m.GetHeader("Via"); v != "" {
		resp.SetHeader("Via", v)
	}
	from := m.GetHeader("From")
	to := m.GetHeader("To")
	if !strings.Contains(strings.ToLower(to), "tag=") {
		to = to + ";tag=" + sip.RandomToken(6)
	}
	resp.SetHeader("From", from)
	resp.SetHeader("To", to)
	resp.SetHeader("Call-ID", m.CallID())
	resp.SetHeader("CSeq", m.GetHeader("CSeq"))
	resp.SetHeader("User-Agent", "GB28181-SimDevice/1.0")
	resp.SetHeader("Contact", fmt.Sprintf("<sip:%s@%s:%d>", d.cfg.Device.ID, d.cfg.SIP.LocalIP, d.cfg.SIP.LocalPort))
	resp.SetHeader("Content-Type", "Application/SDP")
	resp.Body = []byte(answer)
	resp.SetHeader("Content-Length", strconv.Itoa(len(resp.Body)))

	if err := d.ua.SendRaw(src.String(), resp); err != nil {
		log.Printf("[media] send INVITE 200 failed: %v", err)
		d.ms.StopByCallID(callID)
	}
}

func (d *Device) onBye(m *sip.Message, src net.Addr) {
	callID := m.CallID()
	log.Printf("[media] BYE callID=%s", callID)
	d.ms.StopByCallID(callID)
	_ = d.ua.Reply(m, src, 200, "OK", nil, "")
}

func extractChannelID(m *sip.Message, sdp *media.SDPInfo) string {
	// WVP Subject 常见: channelId:ssrc,deviceId,ssrc 或 channelId,deviceId,ssrc
	if sub := m.GetHeader("Subject"); sub != "" {
		parts := strings.Split(sub, ",")
		if len(parts) >= 1 {
			id := config.NormalizeGBID(parts[0])
			if len(id) == 20 {
				return id
			}
		}
	}
	// o= 行第一字段（有时是 设备ID 或 通道ID:SSRC）
	if sdp.Owner != "" {
		f := strings.Fields(sdp.Owner)
		if len(f) > 0 {
			id := config.NormalizeGBID(f[0])
			if len(id) == 20 {
				return id
			}
		}
	}
	// Request-URI / To
	if u := m.ToUser(); u != "" {
		id := config.NormalizeGBID(u)
		if len(id) == 20 {
			return id
		}
	}
	return config.NormalizeGBID(m.ToUser())
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func unmarshalBody(body []byte, v any) error {
	return gb28181.Unmarshal(body, v)
}

func (d *Device) calcPlaybackOffset(recv *media.SDPInfo) float64 {
	if recv.StartTime == "" || recv.StartTime == "0" {
		return 0
	}
	st, err := strconv.ParseInt(recv.StartTime, 10, 64)
	if err != nil || st <= 0 {
		return 0
	}
	// 若为 Unix 秒级时间戳（例如 > 1_000_000_000）
	if st > 1_000_000_000 {
		now := time.Now()
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// 模拟切片按 2 小时（7200 秒）一段
		diff := st - dayStart.Unix()
		if diff > 0 {
			return float64(diff % 7200)
		}
		return 0
	}
	// 否则作为相对秒数
	return float64(st)
}
