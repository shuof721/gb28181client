package ui

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// WSConn 简易纯 Go RFC 6455 WebSocket 连接
type WSConn struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
	closed bool
}

// UpgradeWS 将 HTTP 连接升级为 WebSocket
func UpgradeWS(w http.ResponseWriter, r *http.Request) (*WSConn, error) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return nil, errors.New("not websocket upgrade request")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, errors.New("missing Sec-WebSocket-Key")
	}

	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("webserver doesn't support hijacking")
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijack failed: %w", err)
	}

	resp := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n", acceptKey)

	if _, err := conn.Write([]byte(resp)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &WSConn{
		conn:   conn,
		reader: bufrw.Reader,
	}, nil
}

func (ws *WSConn) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		return nil
	}
	ws.closed = true
	// 发送 Close 帧
	_, _ = ws.conn.Write([]byte{0x88, 0x00})
	return ws.conn.Close()
}

// WriteBinary 发送二进制数据帧 (Opcode 0x02, Server->Client 不做掩码)
func (ws *WSConn) WriteBinary(data []byte) error {
	return ws.writeFrame(0x02, data)
}

// WriteText 发送文本帧 (Opcode 0x01)
func (ws *WSConn) WriteText(text string) error {
	return ws.writeFrame(0x01, []byte(text))
}

func (ws *WSConn) writeFrame(opcode byte, data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		return errors.New("ws closed")
	}

	n := len(data)
	var header []byte
	if n < 126 {
		header = []byte{0x80 | opcode, byte(n)}
	} else if n <= 65535 {
		header = make([]byte, 4)
		header[0] = 0x80 | opcode
		header[1] = 126
		binary.BigEndian.PutUint16(header[2:4], uint16(n))
	} else {
		header = make([]byte, 10)
		header[0] = 0x80 | opcode
		header[1] = 127
		binary.BigEndian.PutUint64(header[2:10], uint64(n))
	}

	if _, err := ws.conn.Write(header); err != nil {
		return err
	}
	if n > 0 {
		_, err := ws.conn.Write(data)
		return err
	}
	return nil
}

// ReadFrame 读取一帧 WebSocket 数据 (返回数据及是否为二进制帧)
func (ws *WSConn) ReadFrame() (payload []byte, isBinary bool, err error) {
	for {
		b0, err := ws.reader.ReadByte()
		if err != nil {
			return nil, false, err
		}
		opcode := b0 & 0x0F

		b1, err := ws.reader.ReadByte()
		if err != nil {
			return nil, false, err
		}
		masked := (b1 & 0x80) != 0
		len7 := int(b1 & 0x7F)

		var length int64
		if len7 < 126 {
			length = int64(len7)
		} else if len7 == 126 {
			var l uint16
			if err := binary.Read(ws.reader, binary.BigEndian, &l); err != nil {
				return nil, false, err
			}
			length = int64(l)
		} else {
			var l uint64
			if err := binary.Read(ws.reader, binary.BigEndian, &l); err != nil {
				return nil, false, err
			}
			length = int64(l)
		}

		var maskKey [4]byte
		if masked {
			if _, err := io.ReadFull(ws.reader, maskKey[:]); err != nil {
				return nil, false, err
			}
		}

		data := make([]byte, length)
		if _, err := io.ReadFull(ws.reader, data); err != nil {
			return nil, false, err
		}

		if masked {
			for i := 0; i < len(data); i++ {
				data[i] ^= maskKey[i%4]
			}
		}

		switch opcode {
		case 0x08: // Close
			_ = ws.Close()
			return nil, false, io.EOF
		case 0x09: // Ping
			// 响应 Pong
			_ = ws.writeFrame(0x0A, data)
			continue
		case 0x0A: // Pong
			continue
		case 0x01: // Text
			return data, false, nil
		case 0x02: // Binary
			return data, true, nil
		default:
			// 忽略其他扩展帧
			continue
		}
	}
}
