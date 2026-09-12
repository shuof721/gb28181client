package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SIP     SIPConfig     `yaml:"sip" json:"sip"`
	Device  DeviceConfig  `yaml:"device" json:"device"`
	Media   MediaConfig   `yaml:"media" json:"media"`
	Record  RecordConfig  `yaml:"record" json:"record"`
	Logging LoggingConfig `yaml:"logging" json:"logging"`
	UI      UIConfig      `yaml:"ui" json:"ui"`
}

// RecordConfig 虚拟录像库参数配置。
type RecordConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`             // 是否开启虚拟录像响应
	Mode         string `yaml:"mode" json:"mode"`                   // continuous (全天连续) | work_hours (工作时段) | alarm (报警)
	SliceMinutes int    `yaml:"slice_minutes" json:"slice_minutes"` // 切片时长 (分钟，默认 60)
	RetainDays   int    `yaml:"retain_days" json:"retain_days"`     // 录像保留天数 (默认 7 天)
	RecordType   string `yaml:"record_type" json:"record_type"`     // 默认录像类型: time | alarm | all
	FileSizeMB   int    `yaml:"file_size_mb" json:"file_size_mb"`   // 模拟单切片大小 (MB，默认 100)
}

// DeviceProfile 单个模拟设备的持久化配置（JSON 存储）。
type DeviceProfile struct {
	Enabled bool         `yaml:"enabled" json:"enabled"`
	SIP     SIPConfig    `yaml:"sip" json:"sip"`
	Device  DeviceConfig `yaml:"device" json:"device"`
	Media   MediaConfig  `yaml:"media" json:"media"`
	Record  RecordConfig `yaml:"record" json:"record"`
}

func (p *DeviceProfile) ToConfig() *Config {
	cfg := &Config{
		SIP:    p.SIP,
		Device: p.Device,
		Media:  p.Media,
		Record: p.Record,
	}
	cfg.applyDefaults()
	return cfg
}

func (p *DeviceProfile) ApplyDefaults() {
	cfg := p.ToConfig()
	p.SIP = cfg.SIP
	p.Device = cfg.Device
	p.Media = cfg.Media
	p.Record = cfg.Record
}

func (p *DeviceProfile) Validate() error {
	cfg := p.ToConfig()
	return cfg.Validate()
}

type UIConfig struct {
	// 轻量 Web 控制台，空或 enabled=false 则不启动
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Listen  string `yaml:"listen" json:"listen"` // 如 "127.0.0.1:8080"
}

type SIPConfig struct {
	// 平台侧
	ServerID   string `yaml:"server_id" json:"server_id"`
	ServerIP   string `yaml:"server_ip" json:"server_ip"`
	ServerPort int    `yaml:"server_port" json:"server_port"`
	// 本地监听
	LocalIP   string `yaml:"local_ip" json:"local_ip"`
	LocalPort int    `yaml:"local_port" json:"local_port"`
	// 传输：udp / tcp
	Transport string `yaml:"transport" json:"transport"`
	// 字符编码：GB2312 / UTF-8
	Charset   string `yaml:"charset" json:"charset"`
	// 注册
	Username string `yaml:"username" json:"username"` // 通常等于 DeviceID
	Password string `yaml:"password" json:"password"`
	// 注册有效期（秒），过期前会自动刷新
	Expires int `yaml:"expires" json:"expires"`
	// 心跳间隔（秒）
	KeepaliveInterval int `yaml:"keepalive_interval" json:"keepalive_interval"`
	// 心跳超时次数，超过后重新注册
	KeepaliveTimeoutCount int `yaml:"keepalive_timeout_count" json:"keepalive_timeout_count"`
}

type DeviceConfig struct {
	// 20 位国标设备编码
	ID           string          `yaml:"id" json:"id"`
	Domain       string          `yaml:"domain" json:"domain"` // 通常为 ID 前 10 位
	Name         string          `yaml:"name" json:"name"`
	Manufacturer string          `yaml:"manufacturer" json:"manufacturer"`
	Model        string          `yaml:"model" json:"model"`
	Firmware     string          `yaml:"firmware" json:"firmware"`
	// NVR 下挂通道
	Channels     []ChannelConfig `yaml:"channels" json:"channels"`
}

type ChannelConfig struct {
	ID           string `yaml:"id" json:"id"`
	Name         string `yaml:"name" json:"name"`
	Manufacturer string `yaml:"manufacturer" json:"manufacturer"`
	Model        string `yaml:"model" json:"model"`
	Address      string `yaml:"address" json:"address"`
	// ON / OFF
	Status       string `yaml:"status" json:"status"`
	Parental     int    `yaml:"parental" json:"parental"`
	ParentID     string `yaml:"parent_id" json:"parent_id"`
	SafetyWay    int    `yaml:"safety_way" json:"safety_way"`
	RegisterWay  int    `yaml:"register_way" json:"register_way"`
	Secrecy      int    `yaml:"secrecy" json:"secrecy"`
	CivilCode    string     `yaml:"civil_code" json:"civil_code"`
	PTZType      int        `yaml:"ptz_type,omitempty" json:"ptz_type,omitempty"` // 1: 球机, 2: 半球, 3: 固定枪机, 4: 遥控枪机
	PTZ          *PTZConfig `yaml:"ptz,omitempty" json:"ptz,omitempty"`
}

// PTZPreset 预置位配置。
type PTZPreset struct {
	ID   int     `yaml:"id" json:"id"`
	Name string  `yaml:"name" json:"name"`
	Pan  float64 `yaml:"pan" json:"pan"`   // 水平角度 0.0 ~ 360.0°
	Tilt float64 `yaml:"tilt" json:"tilt"` // 垂直角度 -90.0 ~ +90.0°
	Zoom float64 `yaml:"zoom" json:"zoom"` // 变倍倍率 1.0 ~ 30.0x
}

// PTZConfig 云台配置。
type PTZConfig struct {
	Enabled bool        `yaml:"enabled" json:"enabled"`
	Pan     float64     `yaml:"pan" json:"pan"`
	Tilt    float64     `yaml:"tilt" json:"tilt"`
	Zoom    float64     `yaml:"zoom" json:"zoom"`
	Presets []PTZPreset `yaml:"presets" json:"presets"`
}

// DefaultPTZConfig 返回默认初始预置位配置。
func DefaultPTZConfig() *PTZConfig {
	return &PTZConfig{
		Enabled: true,
		Pan:     0.0,
		Tilt:    0.0,
		Zoom:    1.0,
		Presets: []PTZPreset{
			{ID: 1, Name: "大门全景", Pan: 0.0, Tilt: 0.0, Zoom: 1.0},
			{ID: 2, Name: "主干道", Pan: 90.0, Tilt: -10.0, Zoom: 2.5},
			{ID: 3, Name: "周界巡查", Pan: 220.0, Tilt: 15.0, Zoom: 1.5},
		},
	}
}

type MediaConfig struct {
	// shared:     所有通道共用下面的 source/mp4_file/h264_file
	// per_channel: 按 channels[通道ID] 配置各自视频；未配置的通道回退到全局
	Mode string `yaml:"mode" json:"mode"`
	// synthetic | file | mp4
	Source string `yaml:"source" json:"source"`
	// file 源时的 Annex-B H.264 文件路径
	H264File string `yaml:"h264_file" json:"h264_file"`
	// mp4 源时的本地 MP4 路径（点播时抽 H.264 循环推流）
	MP4File string `yaml:"mp4_file" json:"mp4_file"`
	// per_channel 模式：key = 20 位通道国标编号
	Channels map[string]ChannelMediaConfig `yaml:"channels" json:"channels"`
	// 宽高（synthetic 源）
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
	// 帧率
	FPS int `yaml:"fps" json:"fps"`
	// RTP 包最大负载（不含 RTP 头）
	RTPPayloadMax int `yaml:"rtp_payload_max" json:"rtp_payload_max"`
	// 本地媒体 IP（发送 RTP 的源地址），默认同 sip.local_ip
	LocalIP string `yaml:"local_ip" json:"local_ip"`
}

// ChannelMediaConfig 单个通道的媒体源。
type ChannelMediaConfig struct {
	Source   string `yaml:"source" json:"source"` // synthetic | file | mp4，空则继承全局
	H264File string `yaml:"h264_file" json:"h264_file"`
	MP4File  string `yaml:"mp4_file" json:"mp4_file"`
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
			ServerID:              "34020000002000000001",
			ServerIP:              "127.0.0.1",
			ServerPort:            5060,
			LocalIP:               "127.0.0.1",
			LocalPort:             5070,
			Transport:             "udp",
			Charset:               "GB2312",
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
		Record: RecordConfig{
			Enabled:      true,
			Mode:         "continuous",
			SliceMinutes: 60,
			RetainDays:   7,
			RecordType:   "time",
			FileSizeMB:   100,
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
	if c.SIP.Charset == "" {
		c.SIP.Charset = "GB2312"
	}
	if c.SIP.ServerID == "" && len(c.SIP.Username) >= 10 {
		c.SIP.ServerID = c.SIP.Username[:10] + "2000000001"
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
	if c.Record.Mode == "" {
		c.Record.Mode = "continuous"
	}
	if c.Record.SliceMinutes <= 0 {
		c.Record.SliceMinutes = 60
	}
	if c.Record.RetainDays <= 0 {
		c.Record.RetainDays = 7
	}
	if c.Record.RecordType == "" {
		c.Record.RecordType = "time"
	}
	if c.Record.FileSizeMB <= 0 {
		c.Record.FileSizeMB = 100
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
