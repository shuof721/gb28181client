package sip

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Handler func(msg *Message, src net.Addr)

// Transport 抽象 UDP/TCP 收发。
type Transport interface {
	LocalAddr() net.Addr
	Send(dst string, data []byte) error
	Close() error
}

type UDPTransport struct {
	conn *net.UDPConn
}

func ListenUDP(ip string, port int) (*UDPTransport, error) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &UDPTransport{conn: conn}, nil
}

func (t *UDPTransport) LocalAddr() net.Addr { return t.conn.LocalAddr() }

func (t *UDPTransport) Send(dst string, data []byte) error {
	raddr, err := net.ResolveUDPAddr("udp", dst)
	if err != nil {
		return err
	}
	_, err = t.conn.WriteToUDP(data, raddr)
	return err
}

func (t *UDPTransport) Close() error { return t.conn.Close() }

func (t *UDPTransport) ReadLoop(h Handler) {
	buf := make([]byte, 64*1024)
	for {
		n, src, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		msg, err := Parse(data)
		if err != nil {
			continue
		}
		go h(msg, src)
	}
}

type TCPTransport struct {
	ln        net.Listener
	mu        sync.Mutex
	conns     map[string]net.Conn
	localIP   string
	localPort int
	handler   Handler
}

func ListenTCP(ip string, port int) (*TCPTransport, error) {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPTransport{
		ln:        ln,
		conns:     map[string]net.Conn{},
		localIP:   ip,
		localPort: port,
	}, nil
}

func (t *TCPTransport) SetHandler(h Handler) {
	t.mu.Lock()
	t.handler = h
	t.mu.Unlock()
}

func (t *TCPTransport) getHandler() Handler {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.handler
}

func (t *TCPTransport) LocalAddr() net.Addr { return t.ln.Addr() }

func (t *TCPTransport) Send(dst string, data []byte) error {
	t.mu.Lock()
	conn, ok := t.conns[dst]
	h := t.handler
	t.mu.Unlock()

	if ok {
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if _, err := conn.Write(data); err == nil {
			return nil
		}
		t.drop(dst)
	}

	conn, err := net.DialTimeout("tcp", dst, 5*time.Second)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.conns[dst] = conn
	t.mu.Unlock()
	if h != nil {
		go t.readLoop(conn, dst, h)
	}
	_, err = conn.Write(data)
	return err
}

func (t *TCPTransport) drop(key string) {
	t.mu.Lock()
	if c, ok := t.conns[key]; ok {
		_ = c.Close()
		delete(t.conns, key)
	}
	t.mu.Unlock()
}

func (t *TCPTransport) Close() error {
	t.mu.Lock()
	for k, c := range t.conns {
		_ = c.Close()
		delete(t.conns, k)
	}
	t.mu.Unlock()
	return t.ln.Close()
}

func (t *TCPTransport) AcceptLoop(h Handler) {
	t.SetHandler(h)
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			return
		}
		key := conn.RemoteAddr().String()
		t.mu.Lock()
		t.conns[key] = conn
		t.mu.Unlock()
		go t.readLoop(conn, key, h)
	}
}

func (t *TCPTransport) readLoop(conn net.Conn, key string, h Handler) {
	defer func() {
		_ = conn.Close()
		t.drop(key)
	}()
	r := bufio.NewReader(conn)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		msg, err := readTCPMessage(r)
		if err != nil {
			return
		}
		go h(msg, conn.RemoteAddr())
	}
}

// readTCPMessage 处理 Content-Length 分帧。
func readTCPMessage(r *bufio.Reader) (*Message, error) {
	var head []byte

	// 1. 跳过起始行之前的所有空行（CRLF 心跳或残余换行，RFC 3261 7.5 规范要求）
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			continue
		}
		head = append(head, line...)
		break
	}

	// 2. 读取剩余的 Headers 直到空行
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		head = append(head, line...)
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	cl := 0
	for _, line := range strings.Split(strings.ReplaceAll(string(head), "\r\n", "\n"), "\n") {
		if line == "" {
			continue
		}
		if i := strings.Index(line, ":"); i > 0 {
			if strings.EqualFold(strings.TrimSpace(line[:i]), "Content-Length") {
				v, _ := strconv.Atoi(strings.TrimSpace(line[i+1:]))
				cl = v
			}
		}
	}
	body := make([]byte, cl)
	if cl > 0 {
		if _, err := io.ReadFull(r, body); err != nil {
			return nil, err
		}
	}
	return Parse(append(head, body...))
}

func JoinHostPort(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func SplitHostPort(s string) (string, int, error) {
	h, p, err := net.SplitHostPort(s)
	if err != nil {
		return "", 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, err
	}
	return h, port, nil
}
