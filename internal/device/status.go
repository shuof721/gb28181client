package device

import (
	"sync"
	"time"

	"github.com/local/gb28181-device/internal/media"
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
	Registered   bool                    `json:"registered"`
	DeviceID     string                  `json:"deviceId"`
	DeviceName   string                  `json:"deviceName"`
	Server       string                  `json:"server"`
	Local        string                  `json:"local"`
	Transport    string                  `json:"transport"`
	MediaMode    string                  `json:"mediaMode"`
	MediaSource  string                  `json:"mediaSource"`
	GuardStatus  string                  `json:"guardStatus"` // "SetGuard" (已布防) | "ResetGuard" (已撤防)
	DutyStatus   string                  `json:"dutyStatus"`  // "ONDUTY" (在防) | "OFFDUTY" (撤防) | "ALARM" (报警中)
	AutoAlarm    bool                    `json:"autoAlarm"`
	Alarms       []AlarmEventRecord      `json:"alarms"`
	Channels     []ChannelStatus         `json:"channels"`
	Sessions     []any                   `json:"sessions"`
	TalkSessions []media.TalkSessionInfo `json:"talkSessions"`
	UptimeSec    int64                   `json:"uptimeSec"`
	StartedAt    time.Time               `json:"startedAt"`
}

type ChannelStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	GuardStatus string `json:"guardStatus,omitempty"` // 该通道布防状态 SetGuard | ResetGuard
	DutyStatus  string `json:"dutyStatus,omitempty"`  // 国标防区状态 ONDUTY | OFFDUTY | ALARM
	IsAlarming  bool   `json:"isAlarming"`            // 是否处于报警激活中
	MP4         string `json:"mp4"`
	H264        string `json:"h264"`
	Source      string `json:"source"`
}
