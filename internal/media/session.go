package media

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SDP 描述
type SDPInfo struct {
	SessionName string // Play / Playback / Talk
	Owner       string
	IP          string // c= 媒体地址
	AudioPort   int
	VideoPort   int
	SSRC        string // y= 行
	IsTCP       bool
	TCPMode     string // passive / active
	StartTime   string // t=
	EndTime     string
	Raw         string
}

// ParseSDP 从 SDP 文本提取媒体参数（足够 GB28181 点播使用）。
func ParseSDP(s string) (*SDPInfo, error) {
	info := &SDPInfo{Raw: s}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "s="):
			info.SessionName = line[2:]
		case strings.HasPrefix(line, "o="):
			info.Owner = line[2:]
		case strings.HasPrefix(line, "c="):
			// c=IN IP4 1.2.3.4
			parts := strings.Fields(line[2:])
			if len(parts) >= 3 {
				info.IP = parts[2]
			}
		case strings.HasPrefix(line, "t="):
			parts := strings.Fields(line[2:])
			if len(parts) >= 1 {
				info.StartTime = parts[0]
			}
			if len(parts) >= 2 {
				info.EndTime = parts[1]
			}
		case strings.HasPrefix(line, "m="):
			// m=video 30000 RTP/AVP 96 97 98
			// m=video 30000 TCP/RTP/AVP 96
			parts := strings.Fields(line[2:])
			if len(parts) < 3 {
				continue
			}
			port, _ := strconv.Atoi(parts[1])
			proto := strings.ToUpper(parts[2])
			if strings.Contains(parts[0], "video") {
				info.VideoPort = port
				if strings.Contains(proto, "TCP") {
					info.IsTCP = true
				}
			} else if strings.Contains(parts[0], "audio") {
				info.AudioPort = port
			}
		case strings.HasPrefix(line, "a="):
			al := strings.ToLower(line[2:])
			if strings.Contains(al, "setup:") {
				if strings.Contains(al, "passive") {
					info.TCPMode = "passive"
				} else if strings.Contains(al, "active") {
					info.TCPMode = "active"
				}
			}
		case strings.HasPrefix(line, "y="):
			info.SSRC = strings.TrimSpace(line[2:])
		}
	}
	if info.IP == "" {
		return nil, fmt.Errorf("invalid sdp: missing c= connection IP")
	}
	if info.VideoPort == 0 {
		return nil, fmt.Errorf("invalid sdp: m=video port is 0 (platform media server/ZLM failed to open RTP port)")
	}
	return info, nil
}

// BuildAnswerSDP 构造设备 200 OK 中的 SDP（sendonly）。
func BuildAnswerSDP(deviceID, channelID, localIP string, ssrc string, recv *SDPInfo) string {
	sessName := firstNonEmpty(recv.SessionName, "Play")
	tLine := "t=0 0"
	if recv.StartTime != "" || recv.EndTime != "" {
		st := firstNonEmpty(recv.StartTime, "0")
		et := firstNonEmpty(recv.EndTime, "0")
		tLine = fmt.Sprintf("t=%s %s", st, et)
	}
	// 端口必须为有效非 0 端口（0 代表拒绝媒体流）
	port := 15060
	proto := "RTP/AVP"
	if recv.IsTCP {
		proto = "TCP/RTP/AVP"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "v=0\r\n")
	fmt.Fprintf(&b, "o=%s 0 0 IN IP4 %s\r\n", deviceID, localIP)
	fmt.Fprintf(&b, "s=%s\r\n", sessName)
	if strings.EqualFold(sessName, "playback") || strings.EqualFold(sessName, "download") {
		fmt.Fprintf(&b, "u=%s:3\r\n", channelID)
	}
	fmt.Fprintf(&b, "c=IN IP4 %s\r\n", localIP)
	fmt.Fprintf(&b, "%s\r\n", tLine)
	fmt.Fprintf(&b, "m=video %d %s 96\r\n", port, proto)
	if recv.IsTCP {
		// RFC 4145: 若接收端（WVP/ZLM）是 passive，推流端应设为 active；若对端是 active，本端设为 passive
		if recv.TCPMode == "active" {
			fmt.Fprintf(&b, "a=setup:passive\r\n")
		} else {
			fmt.Fprintf(&b, "a=setup:active\r\n")
		}
		fmt.Fprintf(&b, "a=connection:new\r\n")
	}
	fmt.Fprintf(&b, "a=sendonly\r\n")
	fmt.Fprintf(&b, "a=rtpmap:96 PS/90000\r\n")
	if ssrc != "" {
		fmt.Fprintf(&b, "y=%s\r\n", ssrc)
		fmt.Fprintf(&b, "f=\r\n")
	}
	return b.String()
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Session 一路点播/回放发送会话。
type Session struct {
	ChannelID string
	CallID    string
	SSRC      string
	ssrcNum   uint32

	StreamType string // "live", "playback", "download"
	remoteIP   string
	remotePort int
	isTCP      bool
	tcpConn    net.Conn

	// 延迟到 loop 里创建，避免抽 MP4 阻塞 INVITE 200 OK
	factory func() (H264Source, error)
	source  H264Source
	fps     int
	payload int
	localIP string

	seq uint16

	packetsSent atomic.Uint64
	bytesSent   atomic.Uint64
	startTime   time.Time

	// 回放控制
	scaleMu       sync.RWMutex
	scale         float64
	paused        atomic.Bool
	currentOffset atomic.Int64 // 毫秒
	OnComplete    func(channelID, callID string)

	initialOffset float64
	seekMu        sync.Mutex
	seekPending   bool
	seekTarget    float64
	seekDone      chan float64

	stopCh  chan struct{}
	stopped atomic.Bool
	wg      sync.WaitGroup
}

func (s *Session) GetScale() float64 {
	s.scaleMu.RLock()
	defer s.scaleMu.RUnlock()
	if s.scale <= 0 {
		return 1.0
	}
	return s.scale
}

func (s *Session) SetScale(scale float64) {
	if scale <= 0 {
		scale = 1.0
	}
	s.scaleMu.Lock()
	s.scale = scale
	s.scaleMu.Unlock()
	log.Printf("[media] session %s ch=%s scale -> %.2f", s.CallID, s.ChannelID, scale)
}

func (s *Session) Pause() {
	s.paused.Store(true)
	log.Printf("[media] session %s ch=%s paused", s.CallID, s.ChannelID)
}

func (s *Session) Resume() {
	s.paused.Store(false)
	log.Printf("[media] session %s ch=%s resumed", s.CallID, s.ChannelID)
}

func (s *Session) IsPaused() bool {
	return s.paused.Load()
}

func (s *Session) Seek(offsetSec float64) float64 {
	if offsetSec < 0 {
		offsetSec = 0
	}
	s.seekMu.Lock()
	s.seekTarget = offsetSec
	s.seekPending = true
	done := make(chan float64, 1)
	s.seekDone = done
	s.seekMu.Unlock()

	log.Printf("[media] session %s ch=%s seek requested to %.2fs", s.CallID, s.ChannelID, offsetSec)

	// 若 session 未启动 loop，直接更新并快速返回
	if s.stopCh == nil {
		s.currentOffset.Store(int64(offsetSec * 1000))
		return offsetSec
	}

	// 若 session 已经在运行，等待 loop 处理完成并返回实际定位时间；若已停止则快速返回
	select {
	case actual := <-done:
		return actual
	case <-time.After(500 * time.Millisecond):
		s.currentOffset.Store(int64(offsetSec * 1000))
		return offsetSec
	case <-s.stopCh:
		s.currentOffset.Store(int64(offsetSec * 1000))
		return offsetSec
	}
}

func (s *Session) CurrentOffset() float64 {
	return float64(s.currentOffset.Load()) / 1000.0
}

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*Session // key: callID
	byChan   map[string]*Session

	sourceFactory func(channelID string) (H264Source, error)
	fps           int
	payload       int
	localIP       string

	OnStart    func(ch string, callID string)
	OnStop     func(ch string, callID string)
	OnComplete func(ch string, callID string)
}

func NewSessionManager(localIP string, fps, payload int, factory func(channelID string) (H264Source, error)) *SessionManager {
	return &SessionManager{
		sessions:      map[string]*Session{},
		byChan:        map[string]*Session{},
		sourceFactory: factory,
		fps:           fps,
		payload:       payload,
		localIP:       localIP,
	}
}

// ParseSSRC 把 y= 的 10 位十进制转 uint32
func ParseSSRC(s string) uint32 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}

func FormatSSRC(v uint32) string {
	return fmt.Sprintf("%010d", v)
}

func (m *SessionManager) StartLive(channelID, callID string, recv *SDPInfo) (string, error) {
	return m.StartSession(channelID, callID, recv, "live", 0)
}

func (m *SessionManager) StartPlayback(channelID, callID string, recv *SDPInfo, initialOffset float64) (string, error) {
	streamType := "playback"
	if strings.EqualFold(recv.SessionName, "download") {
		streamType = "download"
	}
	return m.StartSession(channelID, callID, recv, streamType, initialOffset)
}

func (m *SessionManager) StartSession(channelID, callID string, recv *SDPInfo, streamType string, initialOffset float64) (answerSDP string, err error) {
	m.mu.Lock()
	if streamType == "live" {
		if old, ok := m.byChan[channelID]; ok {
			m.mu.Unlock()
			old.Stop()
			m.mu.Lock()
		}
	}

	ssrcNum := ParseSSRC(recv.SSRC)
	if ssrcNum == 0 {
		ssrcNum = 0x01000000 | (uint32(time.Now().UnixNano()) & 0x00FFFFFF)
	}

	s := &Session{
		ChannelID:     channelID,
		CallID:        callID,
		SSRC:          FormatSSRC(ssrcNum),
		ssrcNum:       ssrcNum,
		StreamType:    streamType,
		remoteIP:      recv.IP,
		remotePort:    recv.VideoPort,
		isTCP:         recv.IsTCP,
		factory:       func() (H264Source, error) { return m.sourceFactory(channelID) },
		fps:           m.fps,
		payload:       m.payload,
		localIP:       m.localIP,
		startTime:     time.Now(),
		scale:         1.0,
		initialOffset: initialOffset,
		OnComplete:    m.OnComplete,
		stopCh:        make(chan struct{}),
	}
	if initialOffset > 0 {
		s.currentOffset.Store(int64(initialOffset * 1000))
	}

	m.sessions[callID] = s
	if streamType == "live" {
		m.byChan[channelID] = s
	}
	m.mu.Unlock()

	answer := BuildAnswerSDP(channelID, channelID, m.localIP, s.SSRC, recv)
	s.wg.Add(1)
	go s.loop()
	log.Printf("[media] start %s ch=%s callID=%s -> %s:%d ssrc=%s tcp=%v",
		streamType, channelID, callID, s.remoteIP, s.remotePort, s.SSRC, s.isTCP)
	if m.OnStart != nil {
		m.OnStart(channelID, callID)
	}
	return answer, nil
}

func (m *SessionManager) GetSession(callID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[callID]
}

func (m *SessionManager) StopByCallID(callID string) {
	m.mu.Lock()
	s, ok := m.sessions[callID]
	m.mu.Unlock()
	if ok {
		s.Stop()
	}
}

// SessionInfo 供 UI/状态查询。
type SessionInfo struct {
	ChannelID     string  `json:"channelId"`
	CallID        string  `json:"callId"`
	SSRC          string  `json:"ssrc"`
	StreamType    string  `json:"streamType"`
	RemoteIP      string  `json:"remoteIp"`
	RemotePort    int     `json:"remotePort"`
	TCP           bool    `json:"tcp"`
	SourceReady   bool    `json:"sourceReady"`
	PacketsSent   uint64  `json:"packetsSent"`
	BytesSent     uint64  `json:"bytesSent"`
	DurationSec   int64   `json:"durationSec"`
	BitrateKbps   float64 `json:"bitrateKbps"`
	Scale         float64 `json:"scale"`
	Paused        bool    `json:"paused"`
	CurrentOffset float64 `json:"currentOffset"`
}

func (m *SessionManager) ListSessions() []SessionInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]SessionInfo, 0, len(m.sessions))
	for _, s := range m.sessions {
		dur := time.Since(s.startTime).Seconds()
		durSec := int64(dur)
		var kbps float64
		bytes := s.bytesSent.Load()
		if dur > 0 {
			kbps = float64(bytes*8) / dur / 1000.0
		}
		streamType := s.StreamType
		if streamType == "" {
			streamType = "live"
		}
		out = append(out, SessionInfo{
			ChannelID:     s.ChannelID,
			CallID:        s.CallID,
			SSRC:          s.SSRC,
			StreamType:    streamType,
			RemoteIP:      s.remoteIP,
			RemotePort:    s.remotePort,
			TCP:           s.isTCP,
			SourceReady:   s.source != nil,
			PacketsSent:   s.packetsSent.Load(),
			BytesSent:     bytes,
			DurationSec:   durSec,
			BitrateKbps:   kbps,
			Scale:         s.GetScale(),
			Paused:        s.IsPaused(),
			CurrentOffset: s.CurrentOffset(),
		})
	}
	return out
}

func (m *SessionManager) StopByChannel(ch string) {
	m.mu.Lock()
	s, ok := m.byChan[ch]
	m.mu.Unlock()
	if ok {
		s.Stop()
	}
}

func (m *SessionManager) StopAll() {
	m.mu.Lock()
	all := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		all = append(all, s)
	}
	m.mu.Unlock()
	for _, s := range all {
		s.Stop()
	}
}

func (m *SessionManager) remove(s *Session) {
	m.mu.Lock()
	delete(m.sessions, s.CallID)
	if cur, ok := m.byChan[s.ChannelID]; ok && cur == s {
		delete(m.byChan, s.ChannelID)
	}
	m.mu.Unlock()
	if m.OnStop != nil {
		m.OnStop(s.ChannelID, s.CallID)
	}
}

func (s *Session) dialTCP() error {
	addr := net.JoinHostPort(s.remoteIP, strconv.Itoa(s.remotePort))
	var lastErr error
	for i := 0; i < 5; i++ {
		select {
		case <-s.stopCh:
			return fmt.Errorf("session stopped")
		default:
		}
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			s.tcpConn = conn
			return nil
		}
		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}
	return lastErr
}

func (s *Session) Stop() {
	if s.stopped.Swap(true) {
		return
	}
	close(s.stopCh)
	if s.tcpConn != nil {
		_ = s.tcpConn.Close()
	}
}

func (s *Session) loop() {
	defer s.wg.Done()
	defer func() {
		if s.source != nil {
			_ = s.source.Close()
		}
		if s.tcpConn != nil {
			_ = s.tcpConn.Close()
		}
	}()

	var wallClock uint64
	if s.initialOffset > 0 {
		wallClock = uint64(s.initialOffset * 90000)
	}
	var lastFrameTime time.Time

	// 若为 TCP 推流，异步自适应连接对端媒体服务器
	if s.isTCP {
		if err := s.dialTCP(); err != nil {
			log.Printf("[media] dial TCP media server failed %s:%d: %v", s.remoteIP, s.remotePort, err)
			return
		}
		log.Printf("[media] TCP media connection established to %s:%d", s.remoteIP, s.remotePort)
	}

	// 先加载/抽流（可能耗时），期间不阻塞 SIP 200 OK
	if s.factory != nil {
		select {
		case <-s.stopCh:
			return
		default:
		}
		log.Printf("[media] loading source ch=%s ...", s.ChannelID)
		src, err := s.factory()
		if err != nil {
			log.Printf("[media] load source failed ch=%s: %v", s.ChannelID, err)
			return
		}
		s.source = src
		log.Printf("[media] source ready ch=%s", s.ChannelID)
		if s.initialOffset > 0 {
			act, err := s.source.Seek(s.initialOffset, s.fps)
			if err == nil {
				wallClock = uint64(act * 90000)
				s.currentOffset.Store(int64(act * 1000))
				log.Printf("[media] session %s ch=%s initial seek to %.2fs (actual=%.2fs wallClock=%d)",
					s.CallID, s.ChannelID, s.initialOffset, act, wallClock)
			}
		}
	}

	var udp *net.UDPConn
	if !s.isTCP {
		c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(s.localIP)})
		if err != nil {
			c, err = net.ListenUDP("udp", &net.UDPAddr{})
			if err != nil {
				log.Printf("[media] listen udp failed: %v", err)
				return
			}
		}
		_ = c.SetWriteBuffer(4 * 1024 * 1024)
		udp = c
		defer udp.Close()
	}

	raddr := &net.UDPAddr{IP: net.ParseIP(s.remoteIP), Port: s.remotePort}

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		// 处理挂起的 Seek 请求
		s.seekMu.Lock()
		if s.seekPending {
			targetSec := s.seekTarget
			doneCh := s.seekDone
			s.seekPending = false
			s.seekDone = nil
			s.seekMu.Unlock()

			actualSec := targetSec
			if s.source != nil {
				act, err := s.source.Seek(targetSec, s.fps)
				if err == nil {
					actualSec = act
				} else {
					log.Printf("[media] source seek failed ch=%s: %v", s.ChannelID, err)
				}
			}
			wallClock = uint64(actualSec * 90000)
			s.currentOffset.Store(int64(actualSec * 1000))
			lastFrameTime = time.Time{} // 立即发帧，无需等待前一帧的间隔
			log.Printf("[media] session %s ch=%s seek applied target=%.2fs actual=%.2fs wallClock=%d",
				s.CallID, s.ChannelID, targetSec, actualSec, wallClock)
			if doneCh != nil {
				select {
				case doneCh <- actualSec:
				default:
				}
			}
		} else {
			s.seekMu.Unlock()
		}

		if s.paused.Load() {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		scale := s.GetScale()
		fps := float64(s.fps) * scale
		if fps <= 0 {
			fps = 25.0
		}
		interval := time.Duration(float64(time.Second) / fps)
		if interval <= 0 {
			interval = 10 * time.Millisecond
		}

		if !lastFrameTime.IsZero() {
			elapsed := time.Since(lastFrameTime)
			if elapsed < interval {
				wait := interval - elapsed
				select {
				case <-s.stopCh:
					return
				case <-time.After(wait):
				}
			}
		}
		lastFrameTime = time.Now()

		if s.source == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		frame, err := s.source.Next()
		if err != nil {
			log.Printf("[media] source error ch=%s: %v", s.ChannelID, err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if len(frame) == 0 {
			continue
		}
		// 90kHz：PES 与 RTP 时间戳同步
		rtpTS := uint32(wallClock)
		ps := PackVideoPES(frame, wallClock)
		pkts := RTPPacketizePS(ps, s.ssrcNum, &s.seq, rtpTS, 96, s.payload)
		wallClock += uint64(90000 / s.fps)
		s.currentOffset.Add(int64(1000 / s.fps))
		for i, pkt := range pkts {
			s.packetsSent.Add(1)
			s.bytesSent.Add(uint64(len(pkt)))
			if s.isTCP {
				if s.tcpConn == nil {
					return
				}
				// GB28181 TCP：WVP 常用 RFC4571 风格 2 字节大端长度 + RTP
				lb := []byte{byte(len(pkt) >> 8), byte(len(pkt))}
				if _, err := s.tcpConn.Write(append(lb, pkt...)); err != nil {
					log.Printf("[media] tcp write error: %v", err)
					return
				}
			} else {
				if _, err := udp.WriteToUDP(pkt, raddr); err != nil {
					log.Printf("[media] udp write error: %v", err)
					return
				}
				// 每发送 16 个 UDP 包（约 20KB）微休眠 50 微秒，平滑突发峰值，杜绝下半部分宏块丢包
				if (i+1)%16 == 0 {
					time.Sleep(50 * time.Microsecond)
				}
			}
		}
	}
}
