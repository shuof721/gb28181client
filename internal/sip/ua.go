package sip

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

// UA 是 GB28181 设备侧 SIP User Agent。
// 职责：注册、鉴权、收发请求/响应、事务匹配。
type UA struct {
	LocalIP   string
	LocalPort int
	ServerIP  string
	ServerPort int
	Transport string // udp | tcp

	Username string // device id
	Password string
	UserAgent string

	udp *UDPTransport
	tcp *TCPTransport

	mu       sync.Mutex
	handlers map[string]func(*Message, net.Addr) // method -> handler for inbound requests

	// 出站事务：branch -> waiter
	txMu   sync.Mutex
	waiters map[string]chan *Message

	cseq int32
	callID string

	// 注册状态
	registered atomic.Bool

	OnRegisterSuccess func(expires int)
	OnRegisterFail    func(status int, reason string)
	OnUnregistered    func()
}

func NewUA(localIP string, localPort int, serverIP string, serverPort int, transport, username, password string) *UA {
	return &UA{
		LocalIP:    localIP,
		LocalPort:  localPort,
		ServerIP:   serverIP,
		ServerPort: serverPort,
		Transport:  transport,
		Username:   username,
		Password:   password,
		UserAgent:  "GB28181-SimDevice/1.0",
		handlers:   map[string]func(*Message, net.Addr){},
		waiters:    map[string]chan *Message{},
		callID:     RandomToken(8) + "@" + localIP,
	}
}

func (ua *UA) Start() error {
	var lastErr error
	u, err := ListenUDP(ua.LocalIP, ua.LocalPort)
	if err == nil {
		ua.udp = u
		go u.ReadLoop(ua.handleRaw)
	} else {
		lastErr = err
	}

	t, err := ListenTCP(ua.LocalIP, ua.LocalPort)
	if err == nil {
		ua.tcp = t
		t.SetHandler(ua.handleRaw)
		go t.AcceptLoop(ua.handleRaw)
	} else {
		if lastErr == nil {
			lastErr = err
		}
	}

	if ua.udp == nil && ua.tcp == nil {
		return fmt.Errorf("failed to listen on %s:%d (udp/tcp): %v", ua.LocalIP, ua.LocalPort, lastErr)
	}
	return nil
}

func (ua *UA) Close() {
	if ua.udp != nil {
		_ = ua.udp.Close()
	}
	if ua.tcp != nil {
		_ = ua.tcp.Close()
	}
}

func (ua *UA) IsRegistered() bool { return ua.registered.Load() }

func (ua *UA) Handle(method string, h func(*Message, net.Addr)) {
	ua.mu.Lock()
	ua.handlers[strings.ToUpper(method)] = h
	ua.mu.Unlock()
}

func (ua *UA) serverAddr() string {
	return JoinHostPort(ua.ServerIP, ua.ServerPort)
}

func (ua *UA) localHostPort() string {
	return JoinHostPort(ua.LocalIP, ua.LocalPort)
}

func (ua *UA) contactHeader() string {
	if strings.EqualFold(ua.Transport, "tcp") {
		return fmt.Sprintf("<sip:%s@%s;transport=tcp>", ua.Username, ua.localHostPort())
	}
	return fmt.Sprintf("<sip:%s@%s>", ua.Username, ua.localHostPort())
}

// ContactURI 导出规范化的 Contact URI
func (ua *UA) ContactURI() string {
	return ua.contactHeader()
}

func (ua *UA) nextCSeq() int {
	return int(atomic.AddInt32(&ua.cseq, 1))
}

// NextCSeq 导出给 mock 平台等外部构造请求使用。
func (ua *UA) NextCSeq() int {
	return ua.nextCSeq()
}

func (ua *UA) send(dst string, data []byte) error {
	if strings.EqualFold(ua.Transport, "tcp") && ua.tcp != nil {
		return ua.tcp.Send(dst, data)
	}
	if ua.udp != nil {
		return ua.udp.Send(dst, data)
	}
	if ua.tcp != nil {
		return ua.tcp.Send(dst, data)
	}
	return fmt.Errorf("no transport available")
}

func (ua *UA) handleRaw(msg *Message, src net.Addr) {
	if !msg.IsRequest {
		ua.dispatchResponse(msg)
		return
	}
	// 平台对我们的请求
	ua.mu.Lock()
	h := ua.handlers[msg.Method]
	ua.mu.Unlock()
	if h != nil {
		h(msg, src)
		return
	}
	// 默认 405
	resp := ua.buildResponse(msg, 405, "Method Not Allowed")
	_ = ua.send(src.String(), resp.Bytes())
}

func (ua *UA) dispatchResponse(msg *Message) {
	branch := msg.Branch()
	ua.txMu.Lock()
	ch := ua.waiters[branch]
	ua.txMu.Unlock()
	if ch != nil {
		select {
		case ch <- msg:
		default:
		}
	}
}

// Request 发送请求并等待最终响应（简化：忽略 1xx，收到 >=200 返回）。
// 对 401/407 自动做 Digest 重试。
func (ua *UA) Request(method, requestURI string, extra func(*Message), body []byte, contentType string) (*Message, error) {
	resp, err := ua.requestOnce(method, requestURI, extra, body, contentType, "")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 || resp.StatusCode == 407 {
		hdr := "WWW-Authenticate"
		if resp.StatusCode == 407 {
			hdr = "Proxy-Authenticate"
		}
		ch, perr := ParseWWWAuthenticate(resp.GetHeader(hdr))
		if perr != nil {
			return resp, perr
		}
		uri := requestURI
		if strings.HasPrefix(uri, "sip:") {
			uri = "sip:" + strings.TrimPrefix(uri, "sip:")
		}
		auth := BuildAuthorization(ua.Username, ua.Password, method, uri, ch, 1, RandomToken(6))
		resp2, err2 := ua.requestOnce(method, requestURI, extra, body, contentType, auth)
		return resp2, err2
	}
	return resp, nil
}

func (ua *UA) requestOnce(method, requestURI string, extra func(*Message), body []byte, contentType, authorization string) (*Message, error) {
	req := NewRequest(method, requestURI)
	via := fmt.Sprintf("SIP/2.0/%s %s;rport", strings.ToUpper(ua.Transport), ua.localHostPort())
	branch := "z9hG4bK" + RandomToken(8)
	via += ";branch=" + branch
	req.SetHeader("Via", via)
	req.SetHeader("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", ua.Username, ua.ServerIP, RandomToken(6)))
	req.SetHeader("To", fmt.Sprintf("<sip:%s@%s>", ua.Username, ua.ServerIP))
	req.SetHeader("Call-ID", ua.callID)
	req.SetHeader("CSeq", fmt.Sprintf("%d %s", ua.nextCSeq(), method))
	req.SetHeader("Max-Forwards", "70")
	req.SetHeader("User-Agent", ua.UserAgent)
	req.SetHeader("Contact", ua.contactHeader())
	if contentType != "" {
		req.SetHeader("Content-Type", contentType)
	}
	if authorization != "" {
		req.SetHeader("Authorization", authorization)
	}
	if body != nil {
		req.Body = body
	}
	if extra != nil {
		extra(req)
	}

	ch := make(chan *Message, 4)
	ua.txMu.Lock()
	ua.waiters[branch] = ch
	ua.txMu.Unlock()
	defer func() {
		ua.txMu.Lock()
		delete(ua.waiters, branch)
		ua.txMu.Unlock()
	}()

	dst := ua.serverAddr()
	if err := ua.send(dst, req.Bytes()); err != nil {
		return nil, err
	}

	deadline := time.After(8 * time.Second)
	for {
		select {
		case m := <-ch:
			if m.StatusCode < 200 {
				continue // 等最终响应
			}
			return m, nil
		case <-deadline:
			return nil, fmt.Errorf("sip %s timeout", method)
		}
	}
}

func (ua *UA) buildResponse(req *Message, code int, reason string) *Message {
	resp := NewResponse(code, reason)
	// Via 原样带回
	if v := req.GetHeader("Via"); v != "" {
		resp.SetHeader("Via", v)
	}
	// From/To 交换
	from := req.GetHeader("From")
	to := req.GetHeader("To")
	if to != "" && !strings.Contains(strings.ToLower(to), "tag=") {
		to = to + ";tag=" + RandomToken(6)
	}
	resp.SetHeader("From", from)
	resp.SetHeader("To", to)
	resp.SetHeader("Call-ID", req.CallID())
	if cseq := req.GetHeader("CSeq"); cseq != "" {
		resp.SetHeader("CSeq", cseq)
	}
	resp.SetHeader("User-Agent", ua.UserAgent)
	resp.SetHeader("Content-Length", strconv.Itoa(len(resp.Body)))
	return resp
}

// SendResponse 根据收到请求的来源协议原路回复响应。
func (ua *UA) SendResponse(src net.Addr, msg *Message) error {
	data := msg.Bytes()
	if _, ok := src.(*net.UDPAddr); ok && ua.udp != nil {
		return ua.udp.Send(src.String(), data)
	}
	if _, ok := src.(*net.TCPAddr); ok && ua.tcp != nil {
		return ua.tcp.Send(src.String(), data)
	}
	return ua.send(src.String(), data)
}

// Reply 向 src 回复响应。
func (ua *UA) Reply(req *Message, src net.Addr, code int, reason string, body []byte, contentType string) error {
	resp := ua.buildResponse(req, code, reason)
	if body != nil {
		resp.Body = body
		if contentType != "" {
			resp.SetHeader("Content-Type", contentType)
		}
	}
	resp.SetHeader("Content-Length", strconv.Itoa(len(resp.Body)))
	return ua.SendResponse(src, resp)
}

// ===== 注册流程 =====

func (ua *UA) Register(expires int) error {
	if expires <= 0 {
		expires = 3600
	}
	requestURI := fmt.Sprintf("sip:%s@%s:%d", ua.Username, ua.ServerIP, ua.ServerPort)
	resp, err := ua.Request("REGISTER", requestURI, func(m *Message) {
		m.SetHeader("Expires", strconv.Itoa(expires))
		m.SetHeader("Contact", ua.contactHeader())
	}, nil, "")
	if err != nil {
		ua.registered.Store(false)
		return err
	}
	if resp.StatusCode == 200 {
		ua.registered.Store(true)
		log.Printf("[sip] REGISTER OK expires=%d", expires)
		if ua.OnRegisterSuccess != nil {
			ua.OnRegisterSuccess(expires)
		}
		return nil
	}
	ua.registered.Store(false)
	if ua.OnRegisterFail != nil {
		ua.OnRegisterFail(resp.StatusCode, resp.Reason)
	}
	return fmt.Errorf("register failed: %d %s", resp.StatusCode, resp.Reason)
}

func (ua *UA) Unregister() error {
	requestURI := fmt.Sprintf("sip:%s@%s:%d", ua.Username, ua.ServerIP, ua.ServerPort)
	_, err := ua.Request("REGISTER", requestURI, func(m *Message) {
		m.SetHeader("Expires", "0")
		m.SetHeader("Contact", ua.contactHeader())
	}, nil, "")
	ua.registered.Store(false)
	if ua.OnUnregistered != nil {
		ua.OnUnregistered()
	}
	return err
}

// SendRaw 直接发送已构造好的响应消息（INVITE 200 OK 等）。
func (ua *UA) SendRaw(dst string, msg *Message) error {
	return ua.send(dst, msg.Bytes())
}

// SendMessage 向平台发送 MESSAGE（如 Keepalive、报警通知）。
func (ua *UA) SendMessage(body []byte, contentType string) (*Message, error) {
	requestURI := fmt.Sprintf("sip:%s@%s:%d", ua.Username, ua.ServerIP, ua.ServerPort)
	// Subject 头 GB28181 常用：deviceid,sn
	return ua.Request("MESSAGE", requestURI, func(m *Message) {
		// 保持
	}, body, contentType)
}

// SendMessageWithSubject 带 Subject 的 MESSAGE。
func (ua *UA) SendMessageWithSubject(body []byte, contentType, subject string) (*Message, error) {
	requestURI := fmt.Sprintf("sip:%s@%s:%d", ua.Username, ua.ServerIP, ua.ServerPort)
	return ua.Request("MESSAGE", requestURI, func(m *Message) {
		if subject != "" {
			m.SetHeader("Subject", subject)
		}
	}, body, contentType)
}
