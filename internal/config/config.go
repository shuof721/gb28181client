package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SIP     SIPConfig     `yaml:"sip"`
	Device  DeviceConfig  `yaml:"device"`
	Media   MediaConfig   `yaml:"media"`
	Logging LoggingConfig `yaml:"logging"`
	UI      UIConfig      `yaml:"ui"`
}

type UIConfig struct {
	// 轻量 Web 控制台，空或 enabled=false 则不启动
	Enabled bool `yaml:"enabled"`
	Listen  string `yaml:"listen"` // 如 "127.0.0.1:8080"
}

type SIPConfig struct {
	// 平台侧
	ServerIP   string `yaml:"server_ip"`
	ServerPort int    `yaml:"server_port"`
	// 本地监听
	LocalIP   string `yaml:"local_ip"`
	LocalPort int    `yaml:"local_port"`
	// 传输：udp / tcp
	Transport string `yaml:"transport"`
	// 注册
	Username string `yaml:"username"` // 通常等于 DeviceID
	Password string `yaml:"password"`
	// 注册有效期（秒），过期前会自动刷新
	Expires int `yaml:"expires"`
	// 心跳间隔（秒）
	KeepaliveInterval int `yaml:"keepalive_interval"`
	// 心跳超时次数，超过后重新注册
	KeepaliveTimeoutCount int `yaml:"keepalive_timeout_count"`
}

type DeviceConfig struct {
	// 20 位国标设备编码
	ID           string `yaml:"id"`
	Domain       string `yaml:"domain"` // 通常为 ID 前 10 位
	Name         string `yaml:"name"`
	Manufacturer string `yaml:"manufacturer"`
	Model        string `yaml:"model"`
	Firmware     string `yaml:"firmware"`
	// NVR 下挂通道
	Channels []ChannelConfig `yaml:"channels"`
}

type ChannelConfig struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Manufacturer string `yaml:"manufacturer"`
	Model        string `yaml:"model"`
	Address      string `yaml:"address"`
	// ON / OFF
	Status     string `yaml:"status"`
	Parental   int    `yaml:"parental"`
	ParentID   string `yaml:"parent_id"`
	SafetyWay  int    `yaml:"safety_way"`
	RegisterWay int   `yaml:"register_way"`
	Secrecy    int    `yaml:"secrecy"`
	CivilCode  string `yaml:"civil_code"`
}

type MediaConfig struct {
	// shared:     所有通道共用下面的 source/mp4_file/h264_file
	// per_channel: 按 channels[通道ID] 配置各自视频；未配置的通道回退到全局
	Mode string `yaml:"mode"`
	// synthetic | file | mp4
	Source string `yaml:"source"`
	// file 源时的 Annex-B H.264 文件路径
	H264File string `yaml:"h264_file"`
	// mp4 源时的本地 MP4 路径（点播时抽 H.264 循环推流）
	MP4File string `yaml:"mp4_file"`
	// per_channel 模式：key = 20 位通道国标编号
	Channels map[string]ChannelMediaConfig `yaml:"channels"`
	// 宽高（synthetic 源）
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
	// 帧率
	FPS int `yaml:"fps"`
	// RTP 包最大负载（不含 RTP 头）
	RTPPayloadMax int `yaml:"rtp_payload_max"`
	// 本地媒体 IP（发送 RTP 的源地址），默认同 sip.local_ip
	LocalIP string `yaml:"local_ip"`
}

// ChannelMediaConfig 单个通道的媒体源。
type ChannelMediaConfig struct {
	Source  string `yaml:"source"`     // synthetic | file | mp4，空则继承全局
	H264File string `yaml:"h264_file"`
	MP4File  string `yaml:"mp4_file"`
}

// NormalizeGBID 把 WVP 等平台可能带来的 "通道ID:SSRC" 规范为 20 位国标编号。
func NormalizeGBID(id string) string {
	id = strings.TrimSpace(id)
	if i := strings.IndexAny(id, ":;, "); i >= 0 {
		id = id[:i]
	}
	if len(id) > 20 {
		id = id[:20]
	}
	return id
}

// OptionsFor 返回该通道应使用的媒体源参数。
func (m *MediaConfig) OptionsFor(channelID string) SourceOptionsView {
	channelID = NormalizeGBID(channelID)
	opts := SourceOptionsView{
		Kind:   m.Source,
		H264:   m.H264File,
		MP4:    m.MP4File,
		Width:  m.Width,
		Height: m.Height,
		FPS:    m.FPS,
	}
	if !strings.EqualFold(m.Mode, "per_channel") {
		return opts
	}
	ch, ok := m.Channels[channelID]
	if !ok {
		return opts
	}
	if ch.Source != "" {
		opts.Kind = ch.Source
	}
	if ch.H264File != "" {
		opts.H264 = ch.H264File
	}
	if ch.MP4File != "" {
		opts.MP4 = ch.MP4File
	}
	return opts
}

// SourceOptionsView 供 device 层转成 media.SourceOptions，避免 config 依赖 media。
type SourceOptionsView struct {
	Kind   string
	H264   string
	MP4    string
	Width  int
	Height int
	FPS    int
}

type LoggingConfig struct {
	Level string `yaml:"level"` // debug / info / warn / error
}

func Default() *Config {
	return &Config{
		SIP: SIPConfig{
			ServerIP:              "127.0.0.1",
			ServerPort:            5060,
			LocalIP:               "127.0.0.1",
			LocalPort:             5070,
			Transport:             "udp",
			Username:              "34020000001180000001",
			Password:              "12345678",
			Expires:               3600,
			KeepaliveInterval:     60,
			KeepaliveTimeoutCount: 3,
		},
		Device: DeviceConfig{
			ID:           "34020000001180000001",
			Domain:       "3402000000",
			Name:         "模拟NVR",
			Manufacturer: "GB28181-Sim",
			Model:        "SIM-NVR-100",
			Firmware:     "V1.0.0",
			Channels: []ChannelConfig{
				{
					ID:            "34020000001320000001",
					Name:          "通道1",
					Manufacturer:  "GB28181-Sim",
					Model:         "SIM-IPC-01",
					Address:       "Camera-01",
					Status:        "ON",
					Parental:      0,
					ParentID:      "34020000001180000001",
					RegisterWay:   1,
					CivilCode:     "340200",
				},
				{
					ID:            "34020000001320000002",
					Name:          "通道2",
					Manufacturer:  "GB28181-Sim",
					Model:         "SIM-IPC-02",
					Address:       "Camera-02",
					Status:        "ON",
					Parental:      0,
					ParentID:      "34020000001180000001",
					RegisterWay:   1,
					CivilCode:     "340200",
				},
			},
		},
		Media: MediaConfig{
			Source:        "synthetic",
			Width:         1280,
			Height:        720,
			FPS:           25,
			RTPPayloadMax: 1400,
		},
		Logging: LoggingConfig{Level: "info"},
		UI: UIConfig{
			Enabled: true,
			Listen:  "127.0.0.1:8080",
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.UI.Listen == "" {
		c.UI.Listen = "127.0.0.1:8080"
	}
	if c.SIP.Transport == "" {
		c.SIP.Transport = "udp"
	}
	if c.SIP.Expires <= 0 {
		c.SIP.Expires = 3600
	}
	if c.SIP.KeepaliveInterval <= 0 {
		c.SIP.KeepaliveInterval = 60
	}
	if c.SIP.KeepaliveTimeoutCount <= 0 {
		c.SIP.KeepaliveTimeoutCount = 3
	}
	if c.SIP.Username == "" {
		c.SIP.Username = c.Device.ID
	}
	if c.Device.Domain == "" && len(c.Device.ID) >= 10 {
		c.Device.Domain = c.Device.ID[:10]
	}
	if c.Media.Source == "" {
		c.Media.Source = "synthetic"
	}
	if c.Media.Mode == "" {
		c.Media.Mode = "shared"
	}
	if c.Media.Width <= 0 {
		c.Media.Width = 1280
	}
	if c.Media.Height <= 0 {
		c.Media.Height = 720
	}
	if c.Media.FPS <= 0 {
		c.Media.FPS = 25
	}
	if c.Media.RTPPayloadMax <= 0 {
		c.Media.RTPPayloadMax = 1400
	}
	if c.Media.LocalIP == "" {
		c.Media.LocalIP = c.SIP.LocalIP
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	for i := range c.Device.Channels {
		ch := &c.Device.Channels[i]
		if ch.Status == "" {
			ch.Status = "ON"
		}
		if ch.ParentID == "" {
			ch.ParentID = c.Device.ID
		}
		if ch.RegisterWay == 0 {
			ch.RegisterWay = 1
		}
		if ch.Manufacturer == "" {
			ch.Manufacturer = c.Device.Manufacturer
		}
		if ch.Model == "" {
			ch.Model = "SIM-IPC"
		}
	}
}

func (c *Config) Validate() error {
	if len(c.Device.ID) != 20 {
		return fmt.Errorf("device.id must be 20 digits, got %d", len(c.Device.ID))
	}
	if c.SIP.ServerIP == "" || c.SIP.LocalIP == "" {
		return fmt.Errorf("sip.server_ip and sip.local_ip are required")
	}
	if c.SIP.ServerPort <= 0 || c.SIP.LocalPort <= 0 {
		return fmt.Errorf("sip ports must be positive")
	}
	t := c.SIP.Transport
	if t != "udp" && t != "tcp" {
		return fmt.Errorf("sip.transport must be udp or tcp")
	}
	if len(c.Device.Channels) == 0 {
		return fmt.Errorf("device.channels must not be empty")
	}
	for _, ch := range c.Device.Channels {
		if len(ch.ID) != 20 {
			return fmt.Errorf("channel id must be 20 digits: %s", ch.ID)
		}
	}
	return nil
}
