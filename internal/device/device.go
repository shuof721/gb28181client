package device

import (
	"fmt"
	"log"
	"net"
	"os"
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
	tm    *media.TalkManager

	sn        int32
	kaFail    int32
	stopCh    chan struct{}
	stopped   atomic.Bool
	srcClose  func() error
	startedAt time.Time
	logs      *LogBuffer

	ptzMu  sync.RWMutex
	ptzMap map[string]*ChannelPTZ

	alarmMgr *AlarmManager
	gpsMgr   *GPSManager
	subMgr   *SubscriptionManager

	timeOffset  atomic.Int64 // 虚拟时钟校时偏移 (纳秒)
	isRecording atomic.Bool  // 是否处于录像状态
	isRebooting atomic.Bool  // 是否处于模拟重启中

	controlEventsMu sync.RWMutex
	controlEvents   []ControlEventRecord

	OnConfigChanged func(cfg *config.Config)
}

// ControlEventRecord 平台远程控制与配置下发审计日志记录
type ControlEventRecord struct {
	ID       string `json:"id"`
	Time     string `json:"time"`
	CmdType  string `json:"cmdType"`
	DeviceID string `json:"deviceId"`
	Action   string `json:"action"`
	Detail   string `json:"detail"`
	Source   string `json:"source"`
}

func New(cfg *config.Config) *Device {
	ua := sip.NewUA(
		cfg.SIP.LocalIP, cfg.SIP.LocalPort,
		cfg.SIP.ServerIP, cfg.SIP.ServerPort,
		cfg.SIP.Transport,
		cfg.SIP.Username, cfg.SIP.Password,
	)
	if cfg.SIP.ServerID != "" {
		ua.SetServerID(cfg.SIP.ServerID)
	}
	d := &Device{
		cfg:      cfg,
		ua:       ua,
		stopCh:   make(chan struct{}),
		logs:     NewLogBuffer(800),
		ptzMap:   make(map[string]*ChannelPTZ),
		alarmMgr: NewAlarmManager(),
	}
	srcFactory := func(channelID string) (media.H264Source, error) {
		d.cfgRLock()
		v := cfg.Media.OptionsFor(channelID)
		d.cfgRUnlock()
		log.Printf("[media] channel %s source=%s mp4=%s h264=%s",
			channelID, v.Kind, v.MP4, v.H264)
		var getPTZ func() media.PTZInfo
		ptz := d.GetChannelPTZ(channelID)
		if ptz != nil {
			getPTZ = func() media.PTZInfo {
				st := ptz.Status()
				return media.PTZInfo{
					Pan:            st.Pan,
					Tilt:           st.Tilt,
					Zoom:           st.Zoom,
					Focus:          st.Focus,
					Iris:           st.Iris,
					IsMoving:       st.IsMoving,
					ActivePresetID: st.ActivePresetID,
					ActivePreset:   st.ActivePreset,
					StatusDesc:     st.StatusDesc,
				}
			}
		}

		return media.NewSource(media.SourceOptions{
			Kind:      v.Kind,
			H264:      v.H264,
			MP4:       v.MP4,
			Width:     v.Width,
			Height:    v.Height,
			FPS:       v.FPS,
			ChannelID: channelID,
			GetPTZ:    getPTZ,
		})
	}
	d.ms = media.NewSessionManager(cfg.Media.LocalIP, cfg.Media.FPS, cfg.Media.RTPPayloadMax, srcFactory)
	d.ms.SetAudioSourceFactory(func(channelID string) (media.AudioSource, string, string, error) {
		return d.CreateAudioSourceForChannel(channelID)
	})
	d.ms.OnComplete = func(ch, callID string) {
		_ = d.SendMediaStatusNotify(ch, "121")
	}
	d.tm = media.NewTalkManager(cfg.SIP.LocalIP)
	d.subMgr = NewSubscriptionManager()
	d.gpsMgr = NewGPSManager(cfg.MobilePosition, cfg.Device.Channels, d.sendMobilePosition)
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
	// GPS 轨迹上报
	d.gpsMgr.Start()

	return nil
}

func (d *Device) Stop() {
	if d.stopped.Swap(true) {
		return
	}
	close(d.stopCh)
	d.gpsMgr.Stop()
	d.ms.StopAll()
	d.tm.StopAll()

	// 优雅注销：Best-effort，最多等 1 秒；网络异常或无响应时直接超时退出
	if d.ua != nil && d.ua.IsRegistered() {
		unregDone := make(chan struct{})
		go func() {
			_ = d.ua.Unregister()
			close(unregDone)
		}()
		select {
		case <-unregDone:
		case <-time.After(1 * time.Second):
			log.Printf("[device] %s unregister timed out (1s), closing immediately", d.cfg.Device.ID)
		}
	}
	if d.ua != nil {
		d.ua.Close()
	}
}

func (d *Device) AlarmManager() *AlarmManager {
	return d.alarmMgr
}

func (d *Device) GPSManager() *GPSManager {
	return d.gpsMgr
}

func (d *Device) SubscriptionManager() *SubscriptionManager {
	return d.subMgr
}

// Now 返回考虑虚拟时钟校准偏移后的设备当前时刻
func (d *Device) Now() time.Time {
	offset := d.timeOffset.Load()
	if offset == 0 {
		return time.Now()
	}
	return time.Now().Add(time.Duration(offset))
}

// SetTimeOffset 设置虚拟时钟校准偏移
func (d *Device) SetTimeOffset(offset time.Duration) {
	d.timeOffset.Store(int64(offset))
	log.Printf("[device] %s virtual clock offset set to %.1fs (target time: %s)",
		d.cfg.Device.ID, offset.Seconds(), d.Now().Format("2006-01-02 15:04:05"))
}

func (d *Device) IsRecording() bool {
	return d.isRecording.Load()
}

func (d *Device) SetRecording(rec bool) {
	d.isRecording.Store(rec)
	log.Printf("[device] %s recording status set to %v", d.cfg.Device.ID, rec)
}

func (d *Device) IsRebooting() bool {
	return d.isRebooting.Load()
}

func (d *Device) recordControlEvent(cmdType, devID, action, detail, source string) {
	d.controlEventsMu.Lock()
	defer d.controlEventsMu.Unlock()
	evt := ControlEventRecord{
		ID:       fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Time:     d.Now().Format("2006-01-02 15:04:05"),
		CmdType:  cmdType,
		DeviceID: devID,
		Action:   action,
		Detail:   detail,
		Source:   source,
	}
	d.controlEvents = append(d.controlEvents, evt)
	if len(d.controlEvents) > 50 {
		d.controlEvents = d.controlEvents[len(d.controlEvents)-50:]
	}
}

func (d *Device) GetControlEvents() []ControlEventRecord {
	d.controlEventsMu.RLock()
	defer d.controlEventsMu.RUnlock()
	res := make([]ControlEventRecord, len(d.controlEvents))
	copy(res, d.controlEvents)
	// 倒序排列，最新事件在前
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	return res
}

func (d *Device) triggerSimulatedReboot() {
	if d.isRebooting.Swap(true) {
		return
	}
	defer d.isRebooting.Store(false)

	log.Printf("[device] %s simulated reboot initiating...", d.cfg.Device.ID)
	// 1. 停止当前所有正在推流的媒体会话
	if d.ms != nil {
		d.ms.StopAll()
	}
	// 2. 向平台注销 (Expires: 0)
	if d.ua != nil {
		_ = d.ua.Unregister()
	}

	// 模拟硬件关机重启耗时 (3秒)
	time.Sleep(3 * time.Second)

	// 3. 重新向平台发起注册恢复服务
	if d.ua != nil {
		log.Printf("[device] %s simulated reboot recovering, re-registering...", d.cfg.Device.ID)
		_ = d.ua.Register(d.cfg.SIP.Expires)
	}
	log.Printf("[device] %s simulated reboot completed", d.cfg.Device.ID)
}

func (d *Device) CreateAudioSourceForChannel(channelID string) (media.AudioSource, string, string, error) {
	d.cfgRLock()
	opts := d.cfg.Media.OptionsFor(channelID)
	d.cfgRUnlock()
	if !opts.AudioEnabled {
		log.Printf("[media] channel %s audio disabled (audioEnabled=false)", channelID)
		return nil, opts.AudioCodec, opts.AudioSource, nil
	}
	audioKind := strings.ToLower(strings.TrimSpace(opts.AudioSource))
	audioFile := opts.AudioFile

	// 当通道绑定的是 MP4 视频，且音源为默认/mp4/auto 时，优先使用该 MP4 提取出的原声音频流
	if (audioKind == "" || audioKind == "mp4" || audioKind == "auto") && opts.Kind == "mp4" && opts.MP4 != "" {
		if cache, err := media.EnsureMP4Audio(opts.MP4); err == nil && cache != "" {
			if st, err := os.Stat(cache); err == nil && st.Size() > 0 {
				audioKind = "file"
				audioFile = cache
				log.Printf("[media] channel %s using MP4 original audio from %s (%d bytes)", channelID, cache, st.Size())
			}
		} else if err != nil {
			log.Printf("[media] channel %s failed to extract MP4 audio: %v, fallback to ambient", channelID, err)
			audioKind = "ambient"
		}
	}

	log.Printf("[media] channel %s creating audio source: kind=%s file=%s codec=%s sr=%d",
		channelID, audioKind, audioFile, opts.AudioCodec, opts.AudioSampleRate)

	src, err := media.NewAudioSource(media.AudioSourceOptions{
		Kind:       audioKind,
		FilePath:   audioFile,
		SampleRate: opts.AudioSampleRate,
		Codec:      opts.AudioCodec,
	})
	return src, opts.AudioCodec, audioKind, err
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
		gSt := "ResetGuard"
		dSt := "OFFDUTY"
		isAlm := false
		if d.alarmMgr != nil {
			gSt = d.alarmMgr.GetGuard(ch.ID)
			dSt = d.alarmMgr.DutyStatus(ch.ID)
			isAlm = d.alarmMgr.IsAlarming(ch.ID)
		}
		chs = append(chs, ChannelStatus{
			ID:           ch.ID,
			Name:         ch.Name,
			Status:       ch.Status,
			GuardStatus:  gSt,
			DutyStatus:   dSt,
			IsAlarming:   isAlm,
			MP4:          v.MP4,
			Source:       v.Kind,
			H264:         v.H264,
			AudioEnabled: v.AudioEnabled,
			AudioSource:  v.AudioSource,
			AudioCodec:   v.AudioCodec,
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

	talkSessions := d.tm.ListSessions()

	guardStatus := "ResetGuard"
	dutyStatus := "OFFDUTY"
	autoAlarm := false
	var recentAlarms []AlarmEventRecord
	if d.alarmMgr != nil {
		guardStatus = d.alarmMgr.GetGuard("")
		dutyStatus = d.alarmMgr.DutyStatus(deviceID)
		autoAlarm = d.alarmMgr.IsAutoAlarmRunning()
		recentAlarms = d.alarmMgr.ListRecords(50)
	}

	up := int64(0)
	if !d.startedAt.IsZero() {
		up = int64(time.Since(d.startedAt).Seconds())
	}
	gpsSt := GPSStatus{}
	if d.gpsMgr != nil {
		gpsSt = d.gpsMgr.Current()
	}
	subCount := 0
	if d.subMgr != nil {
		subCount = len(d.subMgr.ListAll())
	}

	return Status{
		Registered:    d.ua.IsRegistered(),
		DeviceID:      deviceID,
		DeviceName:    deviceName,
		Server:        server,
		Local:         local,
		Transport:     transport,
		MediaMode:     mode,
		MediaSource:   src,
		GuardStatus:   guardStatus,
		DutyStatus:    dutyStatus,
		AutoAlarm:     autoAlarm,
		Alarms:        recentAlarms,
		Channels:      chs,
		Sessions:      sessions,
		TalkSessions:  talkSessions,
		GPS:           gpsSt,
		Subscribers:   subCount,
		UptimeSec:     up,
		StartedAt:     d.startedAt,
		Recording:     d.isRecording.Load(),
		Rebooting:     d.isRebooting.Load(),
		TimeOffsetSec: int64(time.Duration(d.timeOffset.Load()).Seconds()),
		DeviceTime:    d.Now().Format("2006-01-02 15:04:05"),
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
	d.tm.StopByCallID(callID)
}

func (d *Device) TalkManager() *media.TalkManager {
	return d.tm
}

func (d *Device) MediaManager() *media.SessionManager {
	return d.ms
}

// GetChannelPTZ 获取或初始化指定通道的虚拟云台控制器。
func (d *Device) GetChannelPTZ(channelID string) *ChannelPTZ {
	d.ptzMu.Lock()
	defer d.ptzMu.Unlock()

	if p, ok := d.ptzMap[channelID]; ok {
		return p
	}

	d.cfgRLock()
	var ptzCfg *config.PTZConfig
	for _, ch := range d.cfg.Device.Channels {
		if ch.ID == channelID {
			ptzCfg = ch.PTZ
			break
		}
	}
	d.cfgRUnlock()

	if ptzCfg == nil {
		ptzCfg = config.DefaultPTZConfig()
	}

	onChange := func(chID string, latest *config.PTZConfig) {
		d.cfgLock()
		for i := range d.cfg.Device.Channels {
			if d.cfg.Device.Channels[i].ID == chID {
				d.cfg.Device.Channels[i].PTZ = latest
				break
			}
		}
		d.cfgUnlock()
		if d.OnConfigChanged != nil {
			d.OnConfigChanged(d.cfg)
		}
	}

	p := NewChannelPTZ(channelID, ptzCfg, onChange)
	d.ptzMap[channelID] = p
	return p
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
	if interval <= 0 {
		interval = 60 * time.Second
	}
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
	body, err := gb28181.MarshalXMLWithCharset(gb28181.KeepaliveNotify{
		CmdType:  "Keepalive",
		SN:       d.nextSN(),
		DeviceID: d.cfg.Device.ID,
		Status:   "OK",
	}, d.getCharset())
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
	ev := m.GetHeader("Event")
	callID := m.CallID()
	log.Printf("[sip] recv SUBSCRIBE Event=%s from=%s Call-ID=%s", ev, src, callID)

	expiresStr := m.GetHeader("Expires")
	expiresSec := 3600
	if expiresStr != "" {
		if s, err := strconv.Atoi(expiresStr); err == nil {
			expiresSec = s
		}
	}

	// 动态学习对端平台 SIP ID
	fromUser := m.FromUser()
	if fromUser != "" && len(fromUser) >= 10 {
		d.ua.SetServerID(fromUser)
	}

	contactURI := ""
	if contact := m.GetHeader("Contact"); contact != "" {
		contactURI = strings.Trim(contact, "<>")
		if idx := strings.Index(contactURI, ">"); idx != -1 {
			contactURI = contactURI[:idx]
		}
	}

	fromHdr := m.GetHeader("From")
	toHdr := m.GetHeader("To")
	localTag := ""
	if strings.Contains(strings.ToLower(toHdr), "tag=") {
		parts := strings.Split(toHdr, "tag=")
		if len(parts) > 1 {
			localTag = strings.Split(parts[1], ";")[0]
		}
	}
	if localTag == "" {
		localTag = sip.RandomToken(6)
		toHdr = toHdr + ";tag=" + localTag
	}

	extraHeaders := func(resp *sip.Message) {
		if ev != "" {
			resp.SetHeader("Event", ev)
		}
		resp.SetHeader("Expires", strconv.Itoa(expiresSec))
		resp.SetHeader("To", toHdr)
		resp.SetHeader("Contact", d.ua.ContactURI())
	}

	_ = d.ua.ReplyExtra(m, src, 200, "OK", extraHeaders, nil, "")

	// 订阅解除 (Expires=0)
	if expiresSec <= 0 {
		d.subMgr.Remove(callID)
		log.Printf("[sub] subscriber removed (Expires=0): callID=%s", callID)
		return
	}

	// 建立/续订活跃订阅
	sub := &Subscriber{
		ID:           callID,
		Event:        ev,
		CallID:       callID,
		FromHeader:   fromHdr,
		ToHeader:     toHdr,
		ContactURI:   contactURI,
		PlatformID:   fromUser,
		SubscribedAt: time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiresSec) * time.Second),
		ExpiresSec:   expiresSec,
		Interval:     5,
	}

	if len(m.Body) > 0 {
		var q gb28181.MobilePositionQueryReq
		if err := unmarshalBody(m.Body, &q); err == nil && q.Interval > 0 {
			sub.Interval = q.Interval
		}
	}

	d.subMgr.AddOrUpdate(sub)
	log.Printf("[sub] subscription active: Event=%s Platform=%s Expires=%ds Interval=%ds",
		ev, fromUser, expiresSec, sub.Interval)

	// RFC 3265 规范：订阅建立成功后，需立即向订阅方发送初始状态 NOTIFY
	go func() {
		time.Sleep(100 * time.Millisecond)
		lowerEv := strings.ToLower(ev)
		if lowerEv == "presence" || lowerEv == "mobileposition" {
			for _, st := range d.gpsMgr.CurrentAll() {
				_ = d.sendMobilePosition(st)
			}
		} else if lowerEv == "catalog" {
			d.sendInitialCatalogNotify(sub)
		}
	}()
}

func (d *Device) sendMobilePosition(st GPSStatus) error {
	targetChannelID := st.ChannelID
	if targetChannelID == "" {
		d.cfgRLock()
		if len(d.cfg.Device.Channels) > 0 {
			targetChannelID = d.cfg.Device.Channels[0].ID
		} else {
			targetChannelID = d.cfg.Device.ID
		}
		d.cfgRUnlock()
	}
	sn := d.nextSN()
	charset := d.getCharset()

	body, err := gb28181.BuildXML("Notify", map[string]any{
		"CmdType":   "MobilePosition",
		"SN":        sn,
		"DeviceID":  targetChannelID,
		"Time":      st.Time,
		"Longitude": fmt.Sprintf("%.6f", st.Longitude),
		"Latitude":  fmt.Sprintf("%.6f", st.Latitude),
		"Speed":     fmt.Sprintf("%.1f", st.Speed),
		"Direction": fmt.Sprintf("%.1f", st.Direction),
		"Altitude":  fmt.Sprintf("%.1f", st.Altitude),
	}, charset)
	if err != nil {
		log.Printf("[gps] build MobilePosition xml failed: %v", err)
		return err
	}

	// 1. 向活跃的 presence / MobilePosition 订阅者发送 NOTIFY
	subscribers := d.subMgr.ListByEvent("presence")
	if len(subscribers) == 0 {
		subscribers = d.subMgr.ListByEvent("MobilePosition")
	}

	for _, sub := range subscribers {
		subState := fmt.Sprintf("active;expires=%d", sub.RemainingSec())
		reqURI := sub.ContactURI
		if reqURI == "" {
			reqURI = fmt.Sprintf("sip:%s@%s:%d", sub.PlatformID, d.cfg.SIP.ServerIP, d.cfg.SIP.ServerPort)
		}
		fromHeader := sub.ToHeader
		toHeader := sub.FromHeader
		callID := sub.CallID
		event := sub.Event

		go func(rURI, from, to, cid, ev, state string) {
			_, nerr := d.ua.SendNotify(rURI, from, to, cid, ev, state, body, "Application/MANSCDP+xml")
			if nerr != nil {
				log.Printf("[gps] send NOTIFY to %s failed: %v", rURI, nerr)
			}
		}(reqURI, fromHeader, toHeader, callID, event, subState)
	}

	// 2. 若处于 active 主动上报模式，或 both 模式下无有效订阅者，则以 SIP MESSAGE 发往平台
	if st.Mode == "active" || (st.Mode == "both" && len(subscribers) == 0) {
		targetID := d.ua.GetServerID()
		go func() {
			_, merr := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml")
			if merr != nil {
				log.Printf("[gps] send active MESSAGE to %s failed: %v", targetID, merr)
			}
		}()
	}

	log.Printf("[gps] position reported ch=%s lon=%.6f lat=%.6f speed=%.1f dir=%.1f (subs=%d mode=%s)",
		targetChannelID, st.Longitude, st.Latitude, st.Speed, st.Direction, len(subscribers), st.Mode)
	return nil
}

// NotifyCatalogChange 当通道发生变更 (ON/OFF/ADD/DEL/UPDATE) 时向所有活跃 Catalog 订阅者广播 NOTIFY
func (d *Device) NotifyCatalogChange(event string, ch config.ChannelConfig) {
	subscribers := d.subMgr.ListByEvent("Catalog")

	ptzType := ch.PTZType
	if ptzType == 0 {
		ptzType = 1
	}
	item := gb28181.CatalogNotifyItem{
		DeviceID:     ch.ID,
		Event:        strings.ToUpper(event),
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
		PTZType:      ptzType,
	}

	sn := d.nextSN()
	charset := d.getCharset()
	body, err := gb28181.BuildCatalogNotifyXML(sn, d.cfg.Device.ID, []gb28181.CatalogNotifyItem{item}, charset)
	if err != nil {
		log.Printf("[catalog] build Catalog Notify xml failed: %v", err)
		return
	}

	// 1. 发送给所有订阅者
	for _, sub := range subscribers {
		subState := fmt.Sprintf("active;expires=%d", sub.RemainingSec())
		reqURI := sub.ContactURI
		if reqURI == "" {
			reqURI = fmt.Sprintf("sip:%s@%s:%d", sub.PlatformID, d.cfg.SIP.ServerIP, d.cfg.SIP.ServerPort)
		}
		fromHeader := sub.ToHeader
		toHeader := sub.FromHeader
		callID := sub.CallID

		go func(rURI, from, to, cid, state string) {
			_, nerr := d.ua.SendNotify(rURI, from, to, cid, "Catalog", state, body, "Application/MANSCDP+xml")
			if nerr != nil {
				log.Printf("[catalog] send NOTIFY to %s failed: %v", rURI, nerr)
			}
		}(reqURI, fromHeader, toHeader, callID, subState)
	}

	// 2. 若当前没有订阅者，以 MESSAGE 发送兜底
	if len(subscribers) == 0 {
		targetID := d.ua.GetServerID()
		go func() {
			_, merr := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml")
			if merr != nil {
				log.Printf("[catalog] send incremental MESSAGE to %s failed: %v", targetID, merr)
			}
		}()
	}

	log.Printf("[catalog] Catalog Notify sent: ch=%s event=%s subscribers=%d", ch.ID, item.Event, len(subscribers))
}

func (d *Device) sendInitialCatalogNotify(sub *Subscriber) {
	d.cfgRLock()
	channels := d.cfg.Device.Channels
	devID := d.cfg.Device.ID
	d.cfgRUnlock()

	items := make([]gb28181.CatalogNotifyItem, 0, len(channels))
	for _, ch := range channels {
		ptzType := ch.PTZType
		if ptzType == 0 {
			ptzType = 1
		}
		items = append(items, gb28181.CatalogNotifyItem{
			DeviceID:     ch.ID,
			Event:        "ON",
			Name:         ch.Name,
			Manufacturer: ch.Manufacturer,
			Model:        ch.Model,
			Status:       ch.Status,
			ParentID:     devID,
			PTZType:      ptzType,
		})
	}

	sn := d.nextSN()
	charset := d.getCharset()
	body, err := gb28181.BuildCatalogNotifyXML(sn, devID, items, charset)
	if err != nil {
		log.Printf("[catalog] build initial Catalog Notify xml failed: %v", err)
		return
	}

	subState := fmt.Sprintf("active;expires=%d", sub.RemainingSec())
	reqURI := sub.ContactURI
	if reqURI == "" {
		reqURI = fmt.Sprintf("sip:%s@%s:%d", sub.PlatformID, d.cfg.SIP.ServerIP, d.cfg.SIP.ServerPort)
	}
	_, _ = d.ua.SendNotify(reqURI, sub.ToHeader, sub.FromHeader, sub.CallID, "Catalog", subState, body, "Application/MANSCDP+xml")
	log.Printf("[catalog] initial Catalog Notify sent to %s (channels=%d)", sub.PlatformID, len(items))
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
	// 动态学习对端平台 SIP ID（如 34020000002000000001）
	if fromUser := m.FromUser(); fromUser != "" && len(fromUser) >= 10 {
		d.ua.SetServerID(fromUser)
	}

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
	case "presetquery":
		d.respPresetQuery(root, m, src)
	case "recordinfo":
		d.respRecordInfo(root, m, src)
	case "broadcast":
		log.Printf("[gb] broadcast notify from=%s SourceID=%s TargetID=%s", src, root.SourceID, root.TargetID)
		go d.replyMANSCDP(root, m, src, map[string]any{
			"CmdType":  "Broadcast",
			"SN":       root.SN,
			"DeviceID": d.cfg.Device.ID,
			"Result":   "OK",
		})
		// 国标 4.3.4 规范：收到 Broadcast Notify 并回复 200 OK 后，设备作为 UAC 主动向平台发起 INVITE
		go d.startBroadcastInvite(root.SourceID, root.TargetID)
	case "configdownload":
		d.respConfigDownload(root, m, src)
	case "deviceconfig":
		d.onDeviceConfig(root, m, src)
	default:
		log.Printf("[gb] unhandled CmdType=%s", root.CmdType)
	}
}

func (d *Device) replyMANSCDP(root *gb28181.Root, req *sip.Message, src net.Addr, fields map[string]any) {
	body := d.buildXMLResponse(fields)
	targetID := ""
	if req != nil {
		targetID = req.FromUser()
	}
	if targetID == "" {
		targetID = d.ua.GetServerID()
	}
	// 平台期望响应走 SIP MESSAGE（通过新事务发往对端平台 ID）
	_, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] send response to %s failed: %v", targetID, err)
	}
}

func (d *Device) getCharset() string {
	d.cfgRLock()
	defer d.cfgRUnlock()
	if d.cfg != nil && d.cfg.SIP.Charset != "" {
		return d.cfg.SIP.Charset
	}
	return "GB2312"
}

func (d *Device) buildXMLResponse(fields map[string]any) []byte {
	data, err := gb28181.BuildXML("Response", fields, d.getCharset())
	if err != nil {
		log.Printf("[gb] buildXMLResponse failed: %v", err)
		return nil
	}
	return data
}

func (d *Device) buildXMLNotify(fields map[string]any) []byte {
	data, err := gb28181.BuildXML("Notify", fields, d.getCharset())
	if err != nil {
		log.Printf("[gb] buildXMLNotify failed: %v", err)
		return nil
	}
	return data
}

// buildXMLResponse 兼容包级调用
func buildXMLResponse(fields map[string]any) []byte {
	data, _ := gb28181.BuildXML("Response", fields, "GB2312")
	return data
}

// buildXMLNotify 兼容包级调用
func buildXMLNotify(fields map[string]any) []byte {
	data, _ := gb28181.BuildXML("Notify", fields, "GB2312")
	return data
}

func (d *Device) respCatalog(root *gb28181.Root, req *sip.Message, src net.Addr) {
	d.cfgRLock()
	items := make([]gb28181.CatalogItem, 0, len(d.cfg.Device.Channels))
	for _, ch := range d.cfg.Device.Channels {
		ptzType := ch.PTZType
		if ptzType == 0 {
			ptzType = 1 // 默认作为球机 (支持 PTZ 控制)
		}
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
			PTZType:      ptzType,
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
		if it.PTZType > 0 {
			fmt.Fprintf(&listB, "      <PTZType>%d</PTZType>\r\n", it.PTZType)
		}
		listB.WriteString("    </Item>\r\n")
	}
	listB.WriteString("  </DeviceList>\r\n")

	body := d.buildXMLResponse(map[string]any{
		"CmdType":    "Catalog",
		"SN":         root.SN,
		"DeviceID":   root.DeviceID, // 平台请求里的 DeviceID，通常是设备 ID
		"SumNum":     len(items),
		"DeviceList": listB.String(),
	})
	targetID := req.FromUser()
	if targetID == "" {
		targetID = d.ua.GetServerID()
	}
	_, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] catalog response failed: %v", err)
	} else {
		log.Printf("[gb] catalog response sent to %s, channels=%d", targetID, len(items))
	}
}

func (d *Device) respDeviceInfo(root *gb28181.Root, req *sip.Message, src net.Addr) {
	body := d.buildXMLResponse(map[string]any{
		"CmdType":      "DeviceInfo",
		"SN":           root.SN,
		"DeviceID":     d.cfg.Device.ID,
		"DeviceName":   d.cfg.Device.Name,
		"Manufacturer": d.cfg.Device.Manufacturer,
		"Model":        d.cfg.Device.Model,
		"Firmware":     d.cfg.Device.Firmware,
		"Channel":      len(d.cfg.Device.Channels),
	})
	targetID := req.FromUser()
	if targetID == "" {
		targetID = d.ua.GetServerID()
	}
	_, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] deviceinfo response failed: %v", err)
	} else {
		log.Printf("[gb] deviceinfo response sent to %s", targetID)
	}
}

func (d *Device) respDeviceStatus(root *gb28181.Root, req *sip.Message, src net.Addr) {
	d.cfgRLock()
	channels := d.cfg.Device.Channels
	devID := d.cfg.Device.ID
	d.cfgRUnlock()

	var asBuilder strings.Builder
	if len(channels) == 0 {
		duty := "OFFDUTY"
		if d.alarmMgr != nil {
			duty = d.alarmMgr.DutyStatus(devID)
		}
		asBuilder.WriteString("  <Alarmstatus Num=\"1\">\r\n")
		asBuilder.WriteString("    <Item>\r\n")
		fmt.Fprintf(&asBuilder, "      <DeviceID>%s</DeviceID>\r\n", devID)
		fmt.Fprintf(&asBuilder, "      <DutyStatus>%s</DutyStatus>\r\n", duty)
		asBuilder.WriteString("    </Item>\r\n")
		asBuilder.WriteString("  </Alarmstatus>\r\n")
	} else {
		fmt.Fprintf(&asBuilder, "  <Alarmstatus Num=\"%d\">\r\n", len(channels))
		for _, ch := range channels {
			duty := "OFFDUTY"
			if d.alarmMgr != nil {
				duty = d.alarmMgr.DutyStatus(ch.ID)
			}
			asBuilder.WriteString("    <Item>\r\n")
			fmt.Fprintf(&asBuilder, "      <DeviceID>%s</DeviceID>\r\n", ch.ID)
			fmt.Fprintf(&asBuilder, "      <DutyStatus>%s</DutyStatus>\r\n", duty)
			asBuilder.WriteString("    </Item>\r\n")
		}
		asBuilder.WriteString("  </Alarmstatus>\r\n")
	}

	recStatus := "OFF"
	if d.isRecording.Load() {
		recStatus = "ON"
	}

	body := d.buildXMLResponse(map[string]any{
		"CmdType":     "DeviceStatus",
		"SN":          root.SN,
		"DeviceID":    devID,
		"Result":      "OK",
		"Online":      "ONLINE",
		"Status":      "OK",
		"Encode":      "ON",
		"Record":      recStatus,
		"DeviceTime":  d.Now().Format("2006-01-02T15:04:05"),
		"Alarmstatus": asBuilder.String(),
	})
	targetID := req.FromUser()
	if targetID == "" {
		targetID = d.ua.GetServerID()
	}
	_, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml")
	if err != nil {
		log.Printf("[gb] devicestatus response failed: %v", err)
	} else {
		log.Printf("[gb] devicestatus response sent to %s (channels=%d record=%s)", targetID, len(channels), recStatus)
	}
}

func (d *Device) respConfigDownload(root *gb28181.Root, req *sip.Message, src net.Addr) {
	var query gb28181.ConfigDownloadReq
	if err := unmarshalBody(req.Body, &query); err != nil {
		log.Printf("[gb] configdownload parse query failed: %v", err)
	}

	targetID := strings.TrimSpace(query.DeviceID)
	if targetID == "" {
		targetID = strings.TrimSpace(root.DeviceID)
	}
	targetID = config.NormalizeGBID(targetID)
	if targetID == "" {
		targetID = d.cfg.Device.ID
	}

	cfgType := strings.TrimSpace(query.ConfigType)
	d.recordControlEvent("ConfigDownload", targetID, "配置查询", fmt.Sprintf("ConfigType=%s", cfgType), src.String())

	fields := map[string]any{
		"CmdType":  "ConfigDownload",
		"SN":       root.SN,
		"DeviceID": targetID,
		"Result":   "OK",
	}

	switch strings.ToLower(cfgType) {
	case "videoparamopt":
		fields["VideoParamOpt"] = "\r\n    <DownloadSpeed>1/2/4</DownloadSpeed>\r\n    <Resolution>1920*1080/1280*720/704*576</Resolution>\r\n  "
	case "audioparamopt":
		fields["AudioParamOpt"] = "\r\n    <AudioFormat>G.711A/G.711U/AAC</AudioFormat>\r\n    <SamplingRate>8/16/32</SamplingRate>\r\n  "
	default: // "basicparam" 或空
		d.cfgRLock()
		devName := d.cfg.Device.Name
		expires := d.cfg.SIP.Expires
		interval := d.cfg.SIP.KeepaliveInterval
		if interval <= 0 {
			interval = 60
		}
		timeoutCnt := d.cfg.SIP.KeepaliveTimeoutCount
		if timeoutCnt <= 0 {
			timeoutCnt = 3
		}
		if targetID != d.cfg.Device.ID {
			for _, ch := range d.cfg.Device.Channels {
				if ch.ID == targetID {
					devName = ch.Name
					break
				}
			}
		}
		d.cfgRUnlock()

		gpsSt := GPSStatus{Longitude: 116.397428, Latitude: 39.909230}
		if d.gpsMgr != nil {
			gpsSt = d.gpsMgr.Current(targetID)
		}

		var bp strings.Builder
		bp.WriteString("\r\n")
		fmt.Fprintf(&bp, "    <Name>%s</Name>\r\n", devName)
		fmt.Fprintf(&bp, "    <Expiration>%d</Expiration>\r\n", expires)
		fmt.Fprintf(&bp, "    <HeartBeatInterval>%d</HeartBeatInterval>\r\n", interval)
		fmt.Fprintf(&bp, "    <HeartBeatCount>%d</HeartBeatCount>\r\n", timeoutCnt)
		bp.WriteString("    <PositionCapability>1</PositionCapability>\r\n")
		fmt.Fprintf(&bp, "    <Longitude>%.6f</Longitude>\r\n", gpsSt.Longitude)
		fmt.Fprintf(&bp, "    <Latitude>%.6f</Latitude>\r\n  ", gpsSt.Latitude)

		fields["BasicParam"] = bp.String()
	}

	d.replyMANSCDP(root, req, src, fields)
}

func (d *Device) onDeviceConfig(root *gb28181.Root, req *sip.Message, src net.Addr) {
	var cfgReq gb28181.DeviceConfigReq
	if err := unmarshalBody(req.Body, &cfgReq); err != nil {
		log.Printf("[gb] deviceconfig parse error: %v", err)
	}

	targetID := strings.TrimSpace(cfgReq.DeviceID)
	if targetID == "" {
		targetID = strings.TrimSpace(root.DeviceID)
	}
	targetID = config.NormalizeGBID(targetID)
	if targetID == "" {
		targetID = d.cfg.Device.ID
	}

	var actions []string

	// 1. 处理时间校准 (Time / Date)
	timeStr := strings.TrimSpace(cfgReq.Time)
	dateStr := strings.TrimSpace(cfgReq.Date)
	var fullTimeStr string
	if dateStr != "" && timeStr != "" && !strings.Contains(timeStr, "-") {
		fullTimeStr = dateStr + "T" + timeStr
	} else if timeStr != "" {
		fullTimeStr = timeStr
	} else if dateStr != "" {
		fullTimeStr = dateStr
	}

	if fullTimeStr != "" {
		var targetTime time.Time
		var parseErr error
		layouts := []string{
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05.000",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, fullTimeStr, time.Local); err == nil {
				targetTime = t
				parseErr = nil
				break
			} else {
				parseErr = err
			}
		}
		if parseErr == nil && !targetTime.IsZero() {
			offset := targetTime.Sub(time.Now())
			d.timeOffset.Store(int64(offset))
			actions = append(actions, fmt.Sprintf("校时至 %s (偏差 %.1fs)", targetTime.Format("15:04:05"), offset.Seconds()))
			log.Printf("[gb] DeviceConfig Time synced to %s (offset=%.1fs)", targetTime.Format("2006-01-02 15:04:05"), offset.Seconds())
		}
	}

	// 2. 处理 BasicParam
	if cfgReq.BasicParam != nil {
		bp := cfgReq.BasicParam
		d.cfgLock()
		changed := false
		if bp.Name != "" {
			if targetID == d.cfg.Device.ID {
				d.cfg.Device.Name = bp.Name
				changed = true
				actions = append(actions, fmt.Sprintf("修改设备名: %s", bp.Name))
			} else {
				for i := range d.cfg.Device.Channels {
					if d.cfg.Device.Channels[i].ID == targetID {
						d.cfg.Device.Channels[i].Name = bp.Name
						changed = true
						actions = append(actions, fmt.Sprintf("修改通道名: %s", bp.Name))
						break
					}
				}
			}
		}
		if bp.HeartBeatInterval > 0 {
			d.cfg.SIP.KeepaliveInterval = bp.HeartBeatInterval
			changed = true
			actions = append(actions, fmt.Sprintf("心跳周期: %ds", bp.HeartBeatInterval))
		}
		if bp.HeartBeatCount > 0 {
			d.cfg.SIP.KeepaliveTimeoutCount = bp.HeartBeatCount
			changed = true
			actions = append(actions, fmt.Sprintf("心跳超时次数: %d", bp.HeartBeatCount))
		}
		onCh := d.OnConfigChanged
		curCfg := d.cfg
		d.cfgUnlock()

		if changed && onCh != nil {
			onCh(curCfg)
		}

		// 若修改了通道名称，自动向已订阅平台推送 Catalog UPDATE 增量通知！
		if bp.Name != "" && targetID != d.cfg.Device.ID {
			var updatedCh config.ChannelConfig
			for _, c := range curCfg.Device.Channels {
				if c.ID == targetID {
					updatedCh = c
					break
				}
			}
			if updatedCh.ID != "" {
				go d.NotifyCatalogChange("UPDATE", updatedCh)
			}
		}
	}

	// 3. 处理 VideoParamOpt
	if cfgReq.VideoParamOpt != nil {
		vp := cfgReq.VideoParamOpt
		actions = append(actions, fmt.Sprintf("视频参数下发: res=%s speed=%s", vp.Resolution, vp.DownloadSpeed))
	}

	detail := strings.Join(actions, ", ")
	if detail == "" {
		detail = "参数无变更"
	}
	d.recordControlEvent("DeviceConfig", targetID, "配置下发", detail, src.String())

	// 回复 Response 200 OK
	go d.replyMANSCDP(root, req, src, map[string]any{
		"CmdType":  "DeviceConfig",
		"SN":       root.SN,
		"DeviceID": targetID,
		"Result":   "OK",
	})
}

func (d *Device) onDeviceControl(root *gb28181.Root, req *sip.Message, src net.Addr) {
	var ctrl gb28181.DeviceControlReq
	if err := unmarshalBody(req.Body, &ctrl); err != nil {
		log.Printf("[gb] devicecontrol parse: %v", err)
	}

	channelID := strings.TrimSpace(ctrl.DeviceID)
	if channelID == "" {
		channelID = strings.TrimSpace(root.DeviceID)
	}
	channelID = config.NormalizeGBID(channelID)
	if channelID == "" {
		channelID = d.cfg.Device.ID
	}

	d.cfgRLock()
	allChs := make([]string, 0, len(d.cfg.Device.Channels))
	for _, c := range d.cfg.Device.Channels {
		allChs = append(allChs, c.ID)
	}
	rootDevID := d.cfg.Device.ID
	d.cfgRUnlock()

	guardCmd := strings.TrimSpace(ctrl.GuardCmd)
	if guardCmd == "" {
		guardCmd = strings.TrimSpace(ctrl.Info.GuardCmd)
	}
	alarmCmd := strings.TrimSpace(ctrl.AlarmCmd)
	if alarmCmd == "" {
		alarmCmd = strings.TrimSpace(ctrl.Info.AlarmCmd)
	}

	if guardCmd != "" {
		isGuard := strings.EqualFold(guardCmd, "SetGuard") || guardCmd == "1" || strings.EqualFold(guardCmd, "On")
		if d.alarmMgr != nil {
			if channelID == rootDevID || channelID == "" {
				d.alarmMgr.SetGuard("", isGuard, allChs)
				d.alarmMgr.SetGuard(rootDevID, isGuard, allChs)
			} else {
				d.alarmMgr.SetGuard(channelID, isGuard)
			}
		}
		log.Printf("[gb] DeviceControl GuardCmd=%s on %s => isGuard=%v (cascaded=%v)", guardCmd, channelID, isGuard, channelID == rootDevID)
		d.recordControlEvent("DeviceControl", channelID, "布撤防", fmt.Sprintf("GuardCmd=%s", guardCmd), src.String())
		go d.replyMANSCDP(root, req, src, map[string]any{
			"CmdType":  "DeviceControl",
			"SN":       root.SN,
			"DeviceID": channelID,
			"Result":   "OK",
		})
	} else if alarmCmd != "" {
		isReset := strings.EqualFold(alarmCmd, "ResetAlarm") || alarmCmd == "1" || strings.EqualFold(alarmCmd, "Reset")
		if isReset {
			if d.alarmMgr != nil {
				if channelID == rootDevID || channelID == "" {
					d.alarmMgr.ResetAlarm("", allChs)
					d.alarmMgr.ResetAlarm(rootDevID, allChs)
				} else {
					d.alarmMgr.ResetAlarm(channelID)
				}
			}
			log.Printf("[gb] DeviceControl AlarmCmd=%s on %s => Reset OK", alarmCmd, channelID)
			d.recordControlEvent("DeviceControl", channelID, "报警复位", fmt.Sprintf("AlarmCmd=%s", alarmCmd), src.String())
			go d.replyMANSCDP(root, req, src, map[string]any{
				"CmdType":  "DeviceControl",
				"SN":       root.SN,
				"DeviceID": channelID,
				"Result":   "OK",
			})
		}
	} else if ctrl.IsForceIFrame() {
		applied := false
		if d.ms != nil {
			applied = d.ms.ForceIFrame(channelID)
		}
		log.Printf("[gb] DeviceControl Force I-Frame on %s => applied=%v", channelID, applied)
		d.recordControlEvent("DeviceControl", channelID, "强制关键帧", fmt.Sprintf("Force I-Frame (stream=%v)", applied), src.String())
		go d.replyMANSCDP(root, req, src, map[string]any{
			"CmdType":  "DeviceControl",
			"SN":       root.SN,
			"DeviceID": channelID,
			"Result":   "OK",
		})
	} else if ctrl.RecordCmd != "" {
		isRec := strings.EqualFold(ctrl.RecordCmd, "Record") || ctrl.RecordCmd == "1" || strings.EqualFold(ctrl.RecordCmd, "StartRecord")
		d.isRecording.Store(isRec)
		log.Printf("[gb] DeviceControl RecordCmd=%s on %s => isRecording=%v", ctrl.RecordCmd, channelID, isRec)
		d.recordControlEvent("DeviceControl", channelID, "录像控制", fmt.Sprintf("RecordCmd=%s => %v", ctrl.RecordCmd, isRec), src.String())
		go d.replyMANSCDP(root, req, src, map[string]any{
			"CmdType":  "DeviceControl",
			"SN":       root.SN,
			"DeviceID": channelID,
			"Result":   "OK",
		})
	} else if ctrl.HomePosition != nil {
		hp := ctrl.HomePosition
		enabled := hp.Enabled == "1" || strings.EqualFold(hp.Enabled, "true") || hp.HomePositionReset == "1"
		ptz := d.GetChannelPTZ(channelID)
		st := ptz.SetHomePosition(enabled, hp.PresetIndex, hp.ResetTime)
		log.Printf("[gb] DeviceControl HomePosition on %s => enabled=%v preset=%d resetTime=%ds",
			channelID, enabled, hp.PresetIndex, hp.ResetTime)
		d.recordControlEvent("DeviceControl", channelID, "守望位设置",
			fmt.Sprintf("enabled=%v preset=%d resetTime=%ds (status=%s)", enabled, hp.PresetIndex, hp.ResetTime, st.StatusDesc),
			src.String())
		go d.replyMANSCDP(root, req, src, map[string]any{
			"CmdType":  "DeviceControl",
			"SN":       root.SN,
			"DeviceID": channelID,
			"Result":   "OK",
		})
	} else if ctrl.TeleBoot != "" {
		log.Printf("[gb] DeviceControl TeleBoot=%s on %s => Reboot OK, starting reboot sequence...", ctrl.TeleBoot, channelID)
		d.recordControlEvent("DeviceControl", channelID, "远程重启", "TeleBoot=Boot", src.String())
		go d.replyMANSCDP(root, req, src, map[string]any{
			"CmdType":  "DeviceControl",
			"SN":       root.SN,
			"DeviceID": channelID,
			"Result":   "OK",
		})
		go d.triggerSimulatedReboot()
	} else if ctrl.PTZCmd != "" {
		ptzCmd, err := gb28181.ParsePTZCmd(ctrl.PTZCmd)
		if err != nil {
			log.Printf("[gb] DeviceControl parse PTZCmd '%s' failed: %v", ctrl.PTZCmd, err)
		} else {
			ptz := d.GetChannelPTZ(channelID)
			st, execErr := ptz.ExecuteCommand(ptzCmd)
			if execErr != nil {
				log.Printf("[gb] PTZ execute error on %s: %v", channelID, execErr)
			} else {
				log.Printf("[gb] PTZ %s on %s => action=%s (%s), current pos (pan=%.1f° tilt=%.1f° zoom=%.1fx focus=%.0f%% iris=%.0f%%)",
					ctrl.PTZCmd, channelID, ptzCmd.Action, ptzCmd.Description, st.Pan, st.Tilt, st.Zoom, st.Focus, st.Iris)
				if ptzCmd.Action == gb28181.PTZActionPresetCall || ptzCmd.Action == gb28181.PTZActionPresetSet || ptzCmd.Action == gb28181.PTZActionPresetDelete {
					d.recordControlEvent("DeviceControl", channelID, "预置位控制", fmt.Sprintf("%s (%s)", ptzCmd.Description, ctrl.PTZCmd), src.String())
				}
			}
		}
	} else if ctrl.DragZoomIn != nil {
		dz := ctrl.DragZoomIn
		ptz := d.GetChannelPTZ(channelID)
		st := ptz.DragZoom(true, dz.Length, dz.Width, dz.MidPointX, dz.MidPointY, dz.LengthX, dz.LengthY)
		log.Printf("[gb] DragZoomIn on %s => window(%dx%d) center(%d,%d) box(%dx%d) => target (pan=%.1f° tilt=%.1f° zoom=%.1fx)",
			channelID, dz.Length, dz.Width, dz.MidPointX, dz.MidPointY, dz.LengthX, dz.LengthY, st.Pan, st.Tilt, st.Zoom)
		d.recordControlEvent("DeviceControl", channelID, "3D拉框放大", fmt.Sprintf("center=(%d,%d) target=(pan=%.1f° tilt=%.1f° zoom=%.1fx)", dz.MidPointX, dz.MidPointY, st.Pan, st.Tilt, st.Zoom), src.String())
	} else if ctrl.DragZoomOut != nil {
		dz := ctrl.DragZoomOut
		ptz := d.GetChannelPTZ(channelID)
		st := ptz.DragZoom(false, dz.Length, dz.Width, dz.MidPointX, dz.MidPointY, dz.LengthX, dz.LengthY)
		log.Printf("[gb] DragZoomOut on %s => window(%dx%d) center(%d,%d) box(%dx%d) => target (pan=%.1f° tilt=%.1f° zoom=%.1fx)",
			channelID, dz.Length, dz.Width, dz.MidPointX, dz.MidPointY, dz.LengthX, dz.LengthY, st.Pan, st.Tilt, st.Zoom)
		d.recordControlEvent("DeviceControl", channelID, "3D拉框缩小", fmt.Sprintf("center=(%d,%d) target=(pan=%.1f° tilt=%.1f° zoom=%.1fx)", dz.MidPointX, dz.MidPointY, st.Pan, st.Tilt, st.Zoom), src.String())
	} else {
		log.Printf("[gb] DeviceControl CmdType=%s DeviceID=%s", root.CmdType, root.DeviceID)
	}
}

func (d *Device) respPresetQuery(root *gb28181.Root, req *sip.Message, src net.Addr) {
	var query gb28181.PresetQueryReq
	if err := unmarshalBody(req.Body, &query); err != nil {
		log.Printf("[gb] presetquery parse query failed: %v", err)
	}

	channelID := root.DeviceID
	if query.DeviceID != "" {
		channelID = query.DeviceID
	}
	ptz := d.GetChannelPTZ(channelID)
	prs := ptz.PresetsList()

	resp := gb28181.PresetQueryResp{
		CmdType:  "PresetQuery",
		SN:       root.SN,
		DeviceID: channelID,
	}
	resp.PresetList.Num = len(prs)
	for _, pr := range prs {
		resp.PresetList.Items = append(resp.PresetList.Items, gb28181.PresetItem{
			PresetID:   strconv.Itoa(pr.ID),
			PresetName: pr.Name,
		})
	}

	body, err := gb28181.MarshalXMLWithCharset(&resp, d.getCharset())
	if err != nil {
		log.Printf("[gb] marshal presetquery response failed: %v", err)
		return
	}
	targetID := req.FromUser()
	if targetID == "" {
		targetID = d.ua.GetServerID()
	}
	if _, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml"); err != nil {
		log.Printf("[gb] presetquery response failed: %v", err)
	} else {
		log.Printf("[gb] presetquery response sent to %s for %s, presets=%d", targetID, channelID, len(prs))
	}
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
	recCfg := d.cfg.Record
	nvrID := d.cfg.Device.ID
	d.cfgRUnlock()

	targetID := req.FromUser()
	if targetID == "" {
		targetID = d.ua.GetServerID()
	}

	items := GenerateRecordItems(recCfg, channelID, channelName, nvrID, query.StartTime, query.EndTime, query.Type)
	total := len(items)

	const maxItemsPerPacket = 30
	if total == 0 {
		body := d.buildXMLResponse(map[string]any{
			"CmdType":    "RecordInfo",
			"SN":         root.SN,
			"DeviceID":   channelID,
			"Name":       channelName,
			"SumNum":     0,
			"RecordList": "  <RecordList Num=\"0\">\r\n  </RecordList>\r\n",
		})
		if _, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml"); err != nil {
			log.Printf("[gb] recordinfo response (0 items) failed: %v", err)
		} else {
			log.Printf("[gb] recordinfo response sent to %s ch=%s records=0", targetID, channelID)
		}
		return
	}

	packetIdx := 0
	for i := 0; i < total; i += maxItemsPerPacket {
		end := i + maxItemsPerPacket
		if end > total {
			end = total
		}
		chunk := items[i:end]
		packetIdx++

		var listB strings.Builder
		fmt.Fprintf(&listB, "  <RecordList Num=\"%d\">\r\n", len(chunk))
		for _, it := range chunk {
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

		body := d.buildXMLResponse(map[string]any{
			"CmdType":    "RecordInfo",
			"SN":         root.SN,
			"DeviceID":   channelID,
			"Name":       channelName,
			"SumNum":     total,
			"RecordList": listB.String(),
		})

		if _, err := d.ua.SendMessageTo(targetID, body, "Application/MANSCDP+xml"); err != nil {
			log.Printf("[gb] recordinfo response packet %d failed: %v", packetIdx, err)
		}
		if end < total {
			time.Sleep(15 * time.Millisecond) // 避免 UDP 连发丢包
		}
	}
	log.Printf("[gb] recordinfo response sent to %s ch=%s total=%d packets=%d", targetID, channelID, total, packetIdx)
}

func (d *Device) SendMediaStatusNotify(channelID, notifyType string) error {
	if notifyType == "" {
		notifyType = "121"
	}
	body := d.buildXMLNotify(map[string]any{
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
	log.Printf("[gb] MediaStatus notify sent ch=%s type=%s", channelID, notifyType)
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
	isTalk := sessName == "talk" || sessName == "broadcast" || (recv.AudioPort > 0 && recv.VideoPort == 0)

	log.Printf("[media] INVITE channel=%s s=%s media=%s:%d audioPort=%d ssrc=%s t=%s %s tcp=%v talk=%v\n[media] raw SDP:\n%s",
		channelID, recv.SessionName, recv.IP, recv.VideoPort, recv.AudioPort, recv.SSRC, recv.StartTime, recv.EndTime, recv.IsTCP, isTalk, sdp)

	var answer string
	if isTalk {
		streamType := "talk"
		if strings.EqualFold(sessName, "broadcast") {
			streamType = "broadcast"
		}
		talkSess, err := d.tm.StartTalkSession(channelID, callID, recv, streamType)
		if err != nil {
			log.Printf("[talk] start talk session failed: %v", err)
			_ = d.ua.Reply(m, src, 500, "Internal Server Error", nil, "")
			return
		}
		answer = media.BuildTalkAnswerSDP(d.cfg.Device.ID, channelID, d.cfg.SIP.LocalIP, talkSess.LocalPort, talkSess.SSRC, recv)
	} else if isPlayback {
		initOffset := d.calcPlaybackOffset(recv)
		answer, err = d.ms.StartPlayback(channelID, callID, recv, initOffset)
		if err != nil {
			log.Printf("[media] start session failed: %v", err)
			_ = d.ua.Reply(m, src, 500, "Internal Server Error", nil, "")
			return
		}
	} else {
		answer, err = d.ms.StartLive(channelID, callID, recv)
		if err != nil {
			log.Printf("[media] start session failed: %v", err)
			_ = d.ua.Reply(m, src, 500, "Internal Server Error", nil, "")
			return
		}
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
	resp.SetHeader("Contact", d.ua.ContactURI())
	resp.SetHeader("Content-Type", "Application/SDP")
	resp.Body = []byte(answer)
	resp.SetHeader("Content-Length", strconv.Itoa(len(resp.Body)))

	if err := d.ua.SendResponse(src, resp); err != nil {
		log.Printf("[media] send INVITE 200 failed: %v", err)
		if isTalk {
			d.tm.StopByCallID(callID)
		} else {
			d.ms.StopByCallID(callID)
		}
	} else {
		log.Printf("[media] INVITE 200 OK sent to %s ch=%s (talk=%v)", src.String(), channelID, isTalk)
	}
}

func (d *Device) onBye(m *sip.Message, src net.Addr) {
	callID := m.CallID()
	log.Printf("[media] BYE callID=%s", callID)
	d.ms.StopByCallID(callID)
	d.tm.StopByCallID(callID)
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
	t, err := ParseFlexibleTime(recv.StartTime)
	if err == nil && !t.IsZero() {
		d.cfgRLock()
		sliceMins := d.cfg.Record.SliceMinutes
		d.cfgRUnlock()
		if sliceMins <= 0 {
			sliceMins = 60
		}
		sliceSec := int64(sliceMins * 60)
		return float64(t.Unix() % sliceSec)
	}

	st, err := strconv.ParseInt(recv.StartTime, 10, 64)
	if err == nil && st > 0 {
		return float64(st)
	}
	return 0
}

// startBroadcastInvite 国标语音广播：收到平台 Notify 后，设备作为 UAC 主动向平台发起 INVITE
func (d *Device) startBroadcastInvite(sourceID, targetID string) {
	d.cfgRLock()
	serverIP := d.cfg.SIP.ServerIP
	serverPort := d.cfg.SIP.ServerPort
	localIP := d.cfg.SIP.LocalIP
	transport := d.cfg.SIP.Transport
	devID := d.cfg.Device.ID
	domain := d.cfg.Device.Domain
	d.cfgRUnlock()

	if sourceID == "" {
		sourceID = domain + "2000000001"
		if len(sourceID) != 20 {
			sourceID = devID
		}
	}
	if targetID == "" {
		d.cfgRLock()
		if len(d.cfg.Device.Channels) > 0 {
			targetID = d.cfg.Device.Channels[0].ID
		} else {
			targetID = devID
		}
		d.cfgRUnlock()
	}

	callID := fmt.Sprintf("%s@%s", sip.RandomToken(12), localIP)
	isTCP := strings.EqualFold(transport, "tcp")
	ssrc := fmt.Sprintf("1%09d", time.Now().Unix()%1000000000)

	localPort := 51000
	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(localIP), Port: 0})
	if err == nil {
		localPort = c.LocalAddr().(*net.UDPAddr).Port
		_ = c.Close()
	}

	sdp := media.BuildBroadcastOfferSDP(targetID, localIP, localPort, ssrc, isTCP)
	requestURI := fmt.Sprintf("sip:%s@%s:%d", sourceID, serverIP, serverPort)

	log.Printf("[broadcast] initiating outgoing INVITE to %s ch=%s callID=%s tcp=%v\n[broadcast] offer SDP:\n%s",
		requestURI, targetID, callID, isTCP, sdp)

	sentReq, resp, err := d.ua.Invite(requestURI, func(req *sip.Message) {
		fromTag := sip.RandomToken(6)
		req.SetHeader("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", targetID, serverIP, fromTag))
		req.SetHeader("To", fmt.Sprintf("<sip:%s@%s>", sourceID, serverIP))
		req.SetHeader("Call-ID", callID)
		req.SetHeader("Subject", fmt.Sprintf("%s:0,%s:0", targetID, sourceID))
		if isTCP {
			req.SetHeader("Contact", fmt.Sprintf("<sip:%s@%s:%d;transport=tcp>", targetID, localIP, d.cfg.SIP.LocalPort))
		} else {
			req.SetHeader("Contact", fmt.Sprintf("<sip:%s@%s:%d>", targetID, localIP, d.cfg.SIP.LocalPort))
		}
	}, []byte(sdp), "Application/SDP")

	if err != nil {
		log.Printf("[broadcast] outgoing INVITE to %s failed: %v", requestURI, err)
		return
	}
	if resp == nil || resp.StatusCode != 200 {
		status := 0
		reason := "unknown"
		if resp != nil {
			status = resp.StatusCode
			reason = resp.Reason
		}
		log.Printf("[broadcast] platform rejected outgoing INVITE status=%d reason=%s", status, reason)
		return
	}

	log.Printf("[broadcast] platform accepted INVITE 200 OK (Call-ID: %s), sending ACK\n[broadcast] answer SDP:\n%s",
		callID, string(resp.Body))
	if err := d.ua.SendACK(sentReq, resp); err != nil {
		log.Printf("[broadcast] send ACK failed: %v", err)
	}

	platformSDP, err := media.ParseSDP(string(resp.Body))
	if err != nil {
		log.Printf("[broadcast] parse platform SDP error: %v", err)
		return
	}

	log.Printf("[broadcast] starting broadcast audio session ch=%s platformMedia=%s:%d ssrc=%s tcp=%v",
		targetID, platformSDP.IP, platformSDP.AudioPort, platformSDP.SSRC, platformSDP.IsTCP)

	_, err = d.tm.StartTalkSession(targetID, callID, platformSDP, "broadcast")
	if err != nil {
		log.Printf("[broadcast] start broadcast talk session failed: %v", err)
		return
	}
	log.Printf("[broadcast] broadcast active! Channel %s is now listening to platform shout", targetID)
}

// StopTalk 停止指定 Call-ID 的对讲/广播会话；若为主叫发起的广播会话，则向平台发送 BYE
func (d *Device) StopTalk(callID string) {
	sess := d.tm.GetSession(callID)
	if sess != nil {
		if strings.EqualFold(sess.StreamType, "broadcast") {
			d.cfgRLock()
			serverIP := d.cfg.SIP.ServerIP
			serverPort := d.cfg.SIP.ServerPort
			domain := d.cfg.Device.Domain
			devID := d.cfg.Device.ID
			d.cfgRUnlock()

			sourceID := domain + "2000000001"
			if len(sourceID) != 20 {
				sourceID = devID
			}
			requestURI := fmt.Sprintf("sip:%s@%s:%d", sourceID, serverIP, serverPort)
			from := fmt.Sprintf("<sip:%s@%s>", sess.ChannelID, serverIP)
			to := fmt.Sprintf("<sip:%s@%s>", sourceID, serverIP)
			_ = d.ua.SendBYE(requestURI, from, to, callID)
		}
		sess.Stop()
	}
}

// StopAllTalk 停止该设备名下的全部对讲/广播会话
func (d *Device) StopAllTalk() {
	for _, info := range d.tm.ListSessions() {
		d.StopTalk(info.CallID)
	}
}
