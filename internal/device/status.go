package device

import (
	"sync"
	"time"
)

// LogBuffer 最近日志环形缓冲，供 Web UI 展示。
type LogBuffer struct {
	mu   sync.Mutex
	buf  []string
	max  int
}

func NewLogBuffer(max int) *LogBuffer {
	if max <= 0 {
		max = 400
	}
	return &LogBuffer{max: max, buf: make([]string, 0, max)}
}

func (b *LogBuffer) Append(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, line)
	if len(b.buf) > b.max {
		b.buf = b.buf[len(b.buf)-b.max:]
	}
}

func (b *LogBuffer) Tail(n int) []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n <= 0 || n > len(b.buf) {
		n = len(b.buf)
	}
	out := make([]string, n)
	copy(out, b.buf[len(b.buf)-n:])
	return out
}

// Status 对外状态快照。
type Status struct {
	Registered bool      `json:"registered"`
	DeviceID   string    `json:"deviceId"`
	DeviceName string    `json:"deviceName"`
	Server     string    `json:"server"`
	Local      string    `json:"local"`
	Transport  string    `json:"transport"`
	MediaMode  string    `json:"mediaMode"`
	MediaSource string   `json:"mediaSource"`
	Channels   []ChannelStatus `json:"channels"`
	Sessions   []any     `json:"sessions"`
	UptimeSec  int64     `json:"uptimeSec"`
	StartedAt  time.Time `json:"startedAt"`
}

type ChannelStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	MP4    string `json:"mp4"`
	Source string `json:"source"`
}
