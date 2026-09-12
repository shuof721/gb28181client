package media

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// TalkSessionInfo 供 UI 与监控展示
type TalkSessionInfo struct {
	ChannelID   string  `json:"channelId"`
	CallID      string  `json:"callId"`
	SSRC        string  `json:"ssrc"`
	StreamType  string  `json:"streamType"` // "talk" 或 "broadcast"
	RemoteIP    string  `json:"remoteIp"`
	RemotePort  int     `json:"remotePort"`
	LocalPort   int     `json:"localPort"`
	TCP         bool    `json:"tcp"`
	AudioCodec  string  `json:"audioCodec"`
	UplinkMode  string  `json:"uplinkMode"` // "synthetic" 或 "mic"
	RxPackets   uint64  `json:"rxPackets"`
	RxBytes     uint64  `json:"rxBytes"`
	TxPackets   uint64  `json:"txPackets"`
	TxBytes     uint64  `json:"txBytes"`
	RxVolume    float64 `json:"rxVolume"` // 0 ~ 100
	TxVolume    float64 `json:"txVolume"` // 0 ~ 100
	DurationSec int64   `json:"durationSec"`
}

// TalkSession 单路语音对讲/广播会话
type TalkSession struct {
	ChannelID  string
	CallID     string
	SSRC       string
	ssrcNum    uint32
	StreamType string
	RemoteIP   string
	RemotePort int
	LocalIP    string
	LocalPort  int
	IsTCP      bool
	Payload    uint8 // 8: PCMA

	udpConn *net.UDPConn
	tcpMu   sync.Mutex
	tcpConn net.Conn
	raddr   *net.UDPAddr

	rxPackets atomic.Uint64
	rxBytes   atomic.Uint64
	txPackets atomic.Uint64
	txBytes   atomic.Uint64
	rxVolBits atomic.Uint64
	txVolBits atomic.Uint64
	startTime time.Time

	uplinkMode atomic.Pointer[string] // "synthetic" 或 "mic"

	// 监听者分发 (WebSocket 连接通过此 channel 接收来自平台的下行 PCM16 数据)
	listenersMu sync.RWMutex
	listeners   map[chan []byte]struct{}

	// 来自前端麦克风的上行音频缓冲 (PCM16 字节流, 8000Hz 16bit 单声道)
	micInCh chan []int16

	stopCh  chan struct{}
	stopped atomic.Bool
	wg      sync.WaitGroup

	OnStop func(channelID, callID string)
}

func (s *TalkSession) setRxVol(v float64) {
	s.rxVolBits.Store(math.Float64bits(v))
}

func (s *TalkSession) GetRxVol() float64 {
	return math.Float64frombits(s.rxVolBits.Load())
}

func (s *TalkSession) setTxVol(v float64) {
	s.txVolBits.Store(math.Float64bits(v))
}

func (s *TalkSession) GetTxVol() float64 {
	return math.Float64frombits(s.txVolBits.Load())
}

func (s *TalkSession) UplinkMode() string {
	if p := s.uplinkMode.Load(); p != nil {
		return *p
	}
	return "synthetic"
}

func (s *TalkSession) SetUplinkMode(mode string) {
	if mode != "mic" {
		mode = "synthetic"
	}
	s.uplinkMode.Store(&mode)
}

// RegisterListener 注册下行音频监听者 (向其发送平台下行 PCM16 裸数据，每帧 320 字节)
func (s *TalkSession) RegisterListener(ch chan []byte) {
	s.listenersMu.Lock()
	s.listeners[ch] = struct{}{}
	s.listenersMu.Unlock()
}

// UnregisterListener 注销下行音频监听者
func (s *TalkSession) UnregisterListener(ch chan []byte) {
	s.listenersMu.Lock()
	delete(s.listeners, ch)
	s.listenersMu.Unlock()
}

// PushMicPCM 写入来自前端麦克风的 PCM16 采样
func (s *TalkSession) PushMicPCM(samples []int16) {
	if s.stopped.Load() {
		return
	}
	select {
	case s.micInCh <- samples:
	default:
		// 缓冲满时丢弃最旧帧，防止延迟累积
	}
}

func (s *TalkSession) Info() TalkSessionInfo {
	dur := int64(0)
	if !s.startTime.IsZero() {
		dur = int64(time.Since(s.startTime).Seconds())
	}
	return TalkSessionInfo{
		ChannelID:   s.ChannelID,
		CallID:      s.CallID,
		SSRC:        s.SSRC,
		StreamType:  s.StreamType,
		RemoteIP:    s.RemoteIP,
		RemotePort:  s.RemotePort,
		LocalPort:   s.LocalPort,
		TCP:         s.IsTCP,
		AudioCodec:  "PCMA",
		UplinkMode:  s.UplinkMode(),
		RxPackets:   s.rxPackets.Load(),
		RxBytes:     s.rxBytes.Load(),
		TxPackets:   s.txPackets.Load(),
		TxBytes:     s.txBytes.Load(),
		RxVolume:    s.GetRxVol(),
		TxVolume:    s.GetTxVol(),
		DurationSec: dur,
	}
}

func (s *TalkSession) Stop() {
	if s.stopped.Swap(true) {
		return
	}
	close(s.stopCh)
	if s.udpConn != nil {
		_ = s.udpConn.Close()
	}
	s.tcpMu.Lock()
	if s.tcpConn != nil {
		_ = s.tcpConn.Close()
		s.tcpConn = nil
	}
	s.tcpMu.Unlock()
	s.listenersMu.Lock()
	for ch := range s.listeners {
		close(ch)
	}
	s.listeners = map[chan []byte]struct{}{}
	s.listenersMu.Unlock()

	if s.OnStop != nil {
		s.OnStop(s.ChannelID, s.CallID)
	}
}

func (s *TalkSession) getTCPConn() net.Conn {
	s.tcpMu.Lock()
	defer s.tcpMu.Unlock()
	return s.tcpConn
}

func (s *TalkSession) setTCPConn(c net.Conn) {
	s.tcpMu.Lock()
	s.tcpConn = c
	s.tcpMu.Unlock()
}

// dialTCP 连接对端 TCP 媒体端口（带重试机制）
func (s *TalkSession) dialTCP() error {
	addr := net.JoinHostPort(s.RemoteIP, strconv.Itoa(s.RemotePort))
	var lastErr error
	for i := 0; i < 5; i++ {
		select {
		case <-s.stopCh:
			return fmt.Errorf("session stopped")
		default:
		}
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			s.setTCPConn(conn)
			log.Printf("[talk] TCP audio media connection established to %s", addr)
			return nil
		}
		lastErr = err
		log.Printf("[talk] dial TCP %s attempt %d/5 failed: %v", addr, i+1, err)
		time.Sleep(150 * time.Millisecond)
	}
	return lastErr
}

// handleRxPacket 处理解包单个下行 RTP 音频数据包并分发
func (s *TalkSession) handleRxPacket(pkt []byte) {
	if len(pkt) < 12 {
		return
	}
	payload, pt, _, _, err := RTPDepacketizeAudio(pkt)
	if err != nil || len(payload) == 0 {
		return
	}
	// 忽略 RTCP 控制报文 (Payload Type 200-204)
	if pt >= 200 && pt <= 204 {
		return
	}

	s.rxPackets.Add(1)
	s.rxBytes.Add(uint64(len(pkt)))

	// 解码 PCMA / PCMU -> PCM16
	var pcm []int16
	if pt == 0 {
		pcm = ULawToPCM16(payload)
	} else {
		// pt == 8 或默认按 PCMA
		pcm = ALawToPCM16(payload)
	}

	// 计算音量
	vol := CalculateRMSLevel(pcm)
	s.setRxVol(vol)

	// 广播给 WebSocket 前端 (转为小端 PCM16 字节切片)
	pcmBytes := make([]byte, len(pcm)*2)
	for i, v := range pcm {
		binary.LittleEndian.PutUint16(pcmBytes[i*2:i*2+2], uint16(v))
	}

	s.listenersMu.RLock()
	for ch := range s.listeners {
		select {
		case ch <- pcmBytes:
		default:
		}
	}
	s.listenersMu.RUnlock()
}

// rxLoop 下行接收循环 (平台 -> 模拟端)
func (s *TalkSession) rxLoop() {
	defer s.wg.Done()
	buf := make([]byte, 2048)

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		if s.IsTCP {
			conn := s.getTCPConn()
			if conn == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// RFC 4571: 2 字节 RTP 长度大端前缀
			var lenBuf [2]byte
			_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					// 读超时衰减音量
					s.setRxVol(s.GetRxVol() * 0.7)
					continue
				}
				select {
				case <-s.stopCh:
					return
				default:
					time.Sleep(50 * time.Millisecond)
					continue
				}
			}

			pktLen := int(binary.BigEndian.Uint16(lenBuf[:]))
			if pktLen < 12 || pktLen > 2048 {
				continue
			}
			pktBuf := make([]byte, pktLen)
			if _, err := io.ReadFull(conn, pktBuf); err != nil {
				continue
			}
			s.handleRxPacket(pktBuf)
			continue
		}

		if s.udpConn == nil {
			return
		}

		_ = s.udpConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := s.udpConn.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				// 读取超时，衰减 Rx 音量
				s.setRxVol(s.GetRxVol() * 0.7)
				continue
			}
			select {
			case <-s.stopCh:
				return
			default:
				time.Sleep(50 * time.Millisecond)
				continue
			}
		}

		s.handleRxPacket(buf[:n])
	}
}

// txLoop 上行推流循环 (模拟端 -> 平台)
func (s *TalkSession) txLoop() {
	defer s.wg.Done()

	// 20ms 一帧 (8000Hz * 0.02s = 160 samples)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	beepGen := NewBeepGenerator()
	seq := uint16(1)
	ts := uint32(0)

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
		}

		mode := s.UplinkMode()
		var pcm []int16

		if mode == "mic" {
			// 从麦克风输入队列读取
			select {
			case samples := <-s.micInCh:
				pcm = samples
			default:
				// 无麦克风输入时发送静音帧维持心跳与推流
				pcm = make([]int16, 160)
			}
		} else {
			// 内置对讲模拟提示音
			pcm = beepGen.NextFrame()
		}

		if len(pcm) > 160 {
			pcm = pcm[:160]
		} else if len(pcm) < 160 {
			padded := make([]int16, 160)
			copy(padded, pcm)
			pcm = padded
		}

		// 计算并记录上行音量
		vol := CalculateRMSLevel(pcm)
		s.setTxVol(vol)

		// 编码为 PCMA (160 samples -> 160 bytes)
		payload := PCM16ToALaw(pcm)
		pkt := RTPPacketizeAudio(payload, s.Payload, s.ssrcNum, &seq, ts)
		ts += 160

		if s.IsTCP {
			conn := s.getTCPConn()
			if conn != nil {
				lb := []byte{byte(len(pkt) >> 8), byte(len(pkt))}
				if _, err := conn.Write(append(lb, pkt...)); err == nil {
					s.txPackets.Add(1)
					s.txBytes.Add(uint64(len(pkt)))
				} else {
					select {
					case <-s.stopCh:
						return
					default:
					}
				}
			}
		} else {
			if s.udpConn != nil && s.raddr != nil {
				if _, err := s.udpConn.WriteToUDP(pkt, s.raddr); err == nil {
					s.txPackets.Add(1)
					s.txBytes.Add(uint64(len(pkt)))
				}
			}
		}
	}
}

// TalkManager 管理全部对讲会话
type TalkManager struct {
	mu       sync.Mutex
	sessions map[string]*TalkSession // key: callID
	byChan   map[string]*TalkSession // key: channelID
	localIP  string

	OnStart func(channelID, callID string)
	OnStop  func(channelID, callID string)
}

func NewTalkManager(localIP string) *TalkManager {
	return &TalkManager{
		sessions: make(map[string]*TalkSession),
		byChan:   make(map[string]*TalkSession),
		localIP:  localIP,
	}
}

// StartTalkSession 开启一路对讲/广播
func (tm *TalkManager) StartTalkSession(channelID, callID string, recv *SDPInfo, streamType string) (*TalkSession, error) {
	tm.mu.Lock()
	if old, ok := tm.byChan[channelID]; ok {
		tm.mu.Unlock()
		old.Stop()
		tm.mu.Lock()
	}

	remotePort := recv.AudioPort
	if remotePort <= 0 {
		remotePort = recv.VideoPort
	}
	if remotePort <= 0 {
		tm.mu.Unlock()
		return nil, fmt.Errorf("missing audio port in sdp")
	}

	ssrcNum := ParseSSRC(recv.SSRC)
	if ssrcNum == 0 {
		ssrcNum = 0x02000000 | (uint32(time.Now().UnixNano()) & 0x00FFFFFF)
	}

	// 监听本地可用 UDP 端口
	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(tm.localIP), Port: 0})
	if err != nil {
		c, err = net.ListenUDP("udp", &net.UDPAddr{Port: 0})
		if err != nil {
			tm.mu.Unlock()
			return nil, fmt.Errorf("listen local audio udp failed: %w", err)
		}
	}
	localPort := c.LocalAddr().(*net.UDPAddr).Port

	payload := uint8(8)
	if recv.AudioPayload == 0 || recv.AudioPayload == 8 {
		payload = 8
	} else if recv.AudioPayload > 0 {
		payload = uint8(recv.AudioPayload)
	}

	defaultMode := "synthetic"
	s := &TalkSession{
		ChannelID:  channelID,
		CallID:     callID,
		SSRC:       FormatSSRC(ssrcNum),
		ssrcNum:    ssrcNum,
		StreamType: streamType,
		RemoteIP:   recv.IP,
		RemotePort: remotePort,
		LocalIP:    tm.localIP,
		LocalPort:  localPort,
		IsTCP:      recv.IsTCP,
		Payload:    payload,
		udpConn:    c,
		raddr:      &net.UDPAddr{IP: net.ParseIP(recv.IP), Port: remotePort},
		startTime:  time.Now(),
		listeners:  make(map[chan []byte]struct{}),
		micInCh:    make(chan []int16, 20),
		stopCh:     make(chan struct{}),
	}
	s.uplinkMode.Store(&defaultMode)
	s.OnStop = func(ch, cid string) {
		tm.Remove(cid)
	}

	if recv.IsTCP {
		// 快速尝试建立连接，若首次未完备则转入后台重试，确保不阻塞 SIP 响应
		if err := s.dialTCP(); err != nil {
			log.Printf("[talk] initial dial TCP failed: %v, retrying in background", err)
			go func() {
				if err := s.dialTCP(); err != nil {
					log.Printf("[talk] background dial TCP failed: %v", err)
					s.Stop()
				}
			}()
		}
	}

	tm.sessions[callID] = s
	tm.byChan[channelID] = s
	tm.mu.Unlock()

	s.wg.Add(2)
	go s.rxLoop()
	go s.txLoop()

	log.Printf("[talk] started %s ch=%s callID=%s localPort=%d -> %s:%d ssrc=%s tcp=%v",
		streamType, channelID, callID, localPort, recv.IP, remotePort, s.SSRC, recv.IsTCP)

	if tm.OnStart != nil {
		tm.OnStart(channelID, callID)
	}
	return s, nil
}

func (tm *TalkManager) GetSession(callID string) *TalkSession {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.sessions[callID]
}

func (tm *TalkManager) GetActiveSession() *TalkSession {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, s := range tm.sessions {
		return s
	}
	return nil
}

func (tm *TalkManager) StopByCallID(callID string) {
	tm.mu.Lock()
	s, ok := tm.sessions[callID]
	tm.mu.Unlock()
	if ok {
		s.Stop()
	}
}

func (tm *TalkManager) StopAll() {
	tm.mu.Lock()
	all := make([]*TalkSession, 0, len(tm.sessions))
	for _, s := range tm.sessions {
		all = append(all, s)
	}
	tm.mu.Unlock()
	for _, s := range all {
		s.Stop()
	}
}

func (tm *TalkManager) Remove(callID string) {
	tm.mu.Lock()
	s, ok := tm.sessions[callID]
	if !ok {
		tm.mu.Unlock()
		return
	}
	delete(tm.sessions, callID)
	if cur, ok2 := tm.byChan[s.ChannelID]; ok2 && cur == s {
		delete(tm.byChan, s.ChannelID)
	}
	tm.mu.Unlock()

	if tm.OnStop != nil {
		tm.OnStop(s.ChannelID, callID)
	}
}

func (tm *TalkManager) ListSessions() []TalkSessionInfo {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	out := make([]TalkSessionInfo, 0, len(tm.sessions))
	for _, s := range tm.sessions {
		out = append(out, s.Info())
	}
	return out
}
