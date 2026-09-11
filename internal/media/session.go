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
			info.StartTime = line[2:]
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
	if info.IP == "" || info.VideoPort == 0 {
		return nil, fmt.Errorf("invalid sdp: missing c= or m=video")
	}
	return info, nil
}

// BuildAnswerSDP 构造设备 200 OK 中的 SDP（sendonly）。
func BuildAnswerSDP(deviceID, channelID, localIP string, ssrc string, recv *SDPInfo) string {
	// o=设备ID 0 0 IN IP4 localIP
	// s=Play
	// c=IN IP4 localIP
	// t=0 0
	// m=video PORT RTP/AVP 96
	// a=sendonly
	// a=rtpmap:96 PS/90000
	// y=ssrc
	port := 0 // 设备侧接收端口若不需要可填 0；WVP 用我们 sendonly，端口可任意
	// 许多实现填 0 或一个本地端口
	port = 0
	proto := "RTP/AVP"
	if recv.IsTCP {
		proto = "TCP/RTP/AVP"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "v=0\r\n")
	fmt.Fprintf(&b, "o=%s 0 0 IN IP4 %s\r\n", deviceID, localIP)
	fmt.Fprintf(&b, "s=%s\r\n", firstNonEmpty(recv.SessionName, "Play"))
	fmt.Fprintf(&b, "c=IN IP4 %s\r\n", localIP)
	fmt.Fprintf(&b, "t=0 0\r\n")
	fmt.Fprintf(&b, "m=video %d %s 96\r\n", port, proto)
	if recv.IsTCP {
		fmt.Fprintf(&b, "a=setup:passive\r\n")
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

	stopCh  chan struct{}
	stopped atomic.Bool
	wg      sync.WaitGroup
}

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*Session // key: callID
	byChan   map[string]*Session

	sourceFactory func(channelID string) (H264Source, error)
	fps           int
	payload       int
	localIP       string

	OnStart func(ch string, callID string)
	OnStop  func(ch string, callID string)
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

// StartLive 处理 INVITE 实时点播。
// 注意：这里只建会话并立刻返回 SDP，不在锁内/不在调用栈里抽视频，
// 否则 ffmpeg 抽流几十秒会导致 WVP 收流超时。
func (m *SessionManager) StartLive(channelID, callID string, recv *SDPInfo) (answerSDP string, err error) {
	m.mu.Lock()
	if old, ok := m.byChan[channelID]; ok {
		m.mu.Unlock()
		old.Stop()
		m.mu.Lock()
	}

	ssrcNum := ParseSSRC(recv.SSRC)
	if ssrcNum == 0 {
		ssrcNum = 0x01000000 | (uint32(time.Now().UnixNano()) & 0x00FFFFFF)
	}

	s := &Session{
		ChannelID:  channelID,
		CallID:     callID,
		SSRC:       FormatSSRC(ssrcNum),
		ssrcNum:    ssrcNum,
		remoteIP:   recv.IP,
		remotePort: recv.VideoPort,
		isTCP:      recv.IsTCP,
		factory:    func() (H264Source, error) { return m.sourceFactory(channelID) },
		fps:        m.fps,
		payload:    m.payload,
		localIP:    m.localIP,
		stopCh:     make(chan struct{}),
	}

	if recv.IsTCP {
		if err := s.dialTCP(); err != nil {
			m.mu.Unlock()
			return "", err
		}
	}

	m.sessions[callID] = s
	m.byChan[channelID] = s
	m.mu.Unlock()

	answer := BuildAnswerSDP(channelID, channelID, m.localIP, s.SSRC, recv)
	s.wg.Add(1)
	go s.loop()
	log.Printf("[media] start live ch=%s callID=%s -> %s:%d ssrc=%s tcp=%v",
		channelID, callID, s.remoteIP, s.remotePort, s.SSRC, s.isTCP)
	if m.OnStart != nil {
		m.OnStart(channelID, callID)
	}
	return answer, nil
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
	ChannelID  string `json:"channelId"`
	CallID     string `json:"callId"`
	SSRC       string `json:"ssrc"`
	RemoteIP   string `json:"remoteIp"`
	RemotePort int    `json:"remotePort"`
	TCP        bool   `json:"tcp"`
	SourceReady bool  `json:"sourceReady"`
}

func (m *SessionManager) ListSessions() []SessionInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]SessionInfo, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, SessionInfo{
			ChannelID:   s.ChannelID,
			CallID:      s.CallID,
			SSRC:        s.SSRC,
			RemoteIP:    s.remoteIP,
			RemotePort:  s.remotePort,
			TCP:         s.isTCP,
			SourceReady: s.source != nil,
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
	// GB28181 TCP 被动：设备连接平台 media 端口
	addr := net.JoinHostPort(s.remoteIP, strconv.Itoa(s.remotePort))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	s.tcpConn = conn
	return nil
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
		udp = c
		defer udp.Close()
	}

	raddr := &net.UDPAddr{IP: net.ParseIP(s.remoteIP), Port: s.remotePort}
	var wallClock uint64
	interval := time.Second / time.Duration(s.fps)
	if interval <= 0 {
		interval = 40 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if s.source == nil {
				continue
			}
			frame, err := s.source.Next()
			if err != nil {
				log.Printf("[media] source error ch=%s: %v", s.ChannelID, err)
				// synthetic/file 循环源一般不 error；file EOF 已内部循环
				time.Sleep(200 * time.Millisecond)
				continue
			}
			if len(frame) == 0 {
				continue
			}
			// 90kHz
			wallClock += uint64(90000 / s.fps)
			rtpTS := uint32(wallClock)
			ps := PackVideoPES(frame, wallClock)
			pkts := RTPPacketizePS(ps, s.ssrcNum, &s.seq, rtpTS, 96, s.payload)
			for _, pkt := range pkts {
				if s.isTCP {
					if s.tcpConn == nil {
						return
					}
					// GB28181 TCP：WVP 常用 RFC4571 风格 2 字节大端长度 + RTP
					// （另有 0x24 interleaved 模式，当前实现长度前缀）
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
				}
			}
		}
	}
}
