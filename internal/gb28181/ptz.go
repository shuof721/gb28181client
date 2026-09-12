package gb28181

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// PTZAction 云台控制动作类型。
type PTZAction string

const (
	PTZActionStop         PTZAction = "stop"
	PTZActionMove         PTZAction = "move"
	PTZActionPresetSet    PTZAction = "preset-set"
	PTZActionPresetCall   PTZAction = "preset-call"
	PTZActionPresetDelete PTZAction = "preset-delete"
	PTZActionCruiseAdd    PTZAction = "cruise-add"
	PTZActionCruiseDel    PTZAction = "cruise-del"
	PTZActionCruiseSpeed  PTZAction = "cruise-speed"
	PTZActionCruiseDwell  PTZAction = "cruise-dwell"
	PTZActionCruiseStart  PTZAction = "cruise-start"
	PTZActionCruiseStop   PTZAction = "cruise-stop"
	PTZActionUnknown      PTZAction = "unknown"
)

// PTZCommand 解析后的 PTZ 指令。
type PTZCommand struct {
	Action      PTZAction `json:"action"`
	PanDir      int       `json:"panDir"`      // -1: 左, 0: 停止, 1: 右
	TiltDir     int       `json:"tiltDir"`     // -1: 下, 0: 停止, 1: 上
	ZoomDir     int       `json:"zoomDir"`     // -1: 缩小, 0: 停止, 1: 放大
	FocusDir    int       `json:"focusDir"`    // -1: 近, 0: 停止, 1: 远
	IrisDir     int       `json:"irisDir"`     // -1: 关, 0: 停止, 1: 开
	PanSpeed    int       `json:"panSpeed"`    // 0 ~ 255
	TiltSpeed   int       `json:"tiltSpeed"`   // 0 ~ 255
	ZoomSpeed   int       `json:"zoomSpeed"`   // 0 ~ 15
	FocusSpeed  int       `json:"focusSpeed"`  // 0 ~ 255
	IrisSpeed   int       `json:"irisSpeed"`   // 0 ~ 255
	PresetID    int       `json:"presetId"`    // 1 ~ 255
	CruiseID    int       `json:"cruiseId"`    // 1 ~ 255
	Address     int       `json:"address"`     // 设备逻辑地址
	Description string    `json:"description"` // 人类可读说明
	RawHex      string    `json:"rawHex"`
}

// ParsePTZCmd 解析 GB/T 28181 8字节 PTZCmd 十六进制指令。
// 格式通常如 "A50F01021F0000D0" 或带空格 "A5 0F 01 02 1F 00 00 D0"。
func ParsePTZCmd(hexStr string) (*PTZCommand, error) {
	s := strings.TrimSpace(strings.ReplaceAll(hexStr, " ", ""))
	if len(s) < 16 {
		return nil, fmt.Errorf("invalid ptz cmd length: expected at least 16 hex chars, got %d", len(s))
	}
	b, err := hex.DecodeString(s[:16])
	if err != nil {
		return nil, fmt.Errorf("invalid hex in ptz cmd: %w", err)
	}

	if b[0] != 0xA5 {
		return nil, fmt.Errorf("invalid ptz header: expected 0xA5, got 0x%02X", b[0])
	}

	var cmdByte byte
	var panSpeed, tiltSpeed, zoomSpeed, focusSpeed, irisSpeed, presetID, cruiseID, address int

	// GB/T 28181-2016 附录 A.3.1 表 A.1 规范格式：
	// Byte 0: 0xA5 (首字节)
	// Byte 1: 0x0F (组合码，高4位版本信息 0000，低4位校验位 1111)
	// Byte 2: 地址码低8位 (设备逻辑地址)
	// Byte 3: 指令码 (命令字节)
	// Byte 4: 水平控制速度 (0~255) / 聚焦速度 (0~255) / 巡航组号 (1~255)
	// Byte 5: 垂直控制速度 (0~255) / 光圈速度 (0~255) / 巡航速度/停留时间
	// Byte 6: 变倍控制速度 (高4位 0~15) / 预置位号 (1~255)
	// Byte 7: 校验码 (前7字节之和模256)
	isStandardGB := true
	if (b[2] >= 0x81 && b[2] <= 0x89) || (b[2] > 0x0F && b[3] == 0x00 && b[4] == 0x00 && b[5] == 0x00) {
		// 容错：个别非标准工具将控制码直接置于 Byte 2
		isStandardGB = false
	}

	if isStandardGB {
		address = int(b[2])
		cmdByte = b[3]
		panSpeed = int(b[4])
		focusSpeed = int(b[4])
		cruiseID = int(b[4])
		tiltSpeed = int(b[5])
		irisSpeed = int(b[5])
		zoomSpeed = int((b[6] >> 4) & 0x0F)
		presetID = int(b[6])
	} else {
		address = 1
		cmdByte = b[2]
		panSpeed = int(b[3])
		focusSpeed = int(b[3])
		cruiseID = int(b[3])
		tiltSpeed = int(b[4])
		irisSpeed = int(b[4])
		zoomSpeed = int((b[5] >> 4) & 0x0F)
		presetID = int(b[6])
	}

	cmd := &PTZCommand{
		RawHex:     strings.ToUpper(hex.EncodeToString(b)),
		Address:    address,
		PanSpeed:   panSpeed,
		TiltSpeed:  tiltSpeed,
		ZoomSpeed:  zoomSpeed,
		FocusSpeed: focusSpeed,
		IrisSpeed:  irisSpeed,
		PresetID:   presetID,
		CruiseID:   cruiseID,
	}

	// 1. 预置位与巡航控制指令 (指令码高位为 0x80)
	if cmdByte >= 0x80 {
		switch cmdByte {
		case 0x81:
			cmd.Action = PTZActionPresetSet
			cmd.Description = fmt.Sprintf("设置预置位 %d", presetID)
			return cmd, nil
		case 0x82:
			cmd.Action = PTZActionPresetCall
			cmd.Description = fmt.Sprintf("调用预置位 %d", presetID)
			return cmd, nil
		case 0x83:
			cmd.Action = PTZActionPresetDelete
			cmd.Description = fmt.Sprintf("删除预置位 %d", presetID)
			return cmd, nil
		case 0x84:
			cmd.Action = PTZActionCruiseAdd
			cmd.Description = fmt.Sprintf("巡航组 %d 添加预置位 %d", cmd.CruiseID, presetID)
			return cmd, nil
		case 0x85:
			cmd.Action = PTZActionCruiseDel
			cmd.Description = fmt.Sprintf("巡航组 %d 删除预置位 %d", cmd.CruiseID, presetID)
			return cmd, nil
		case 0x86:
			cmd.Action = PTZActionCruiseSpeed
			cmd.Description = fmt.Sprintf("设置巡航组 %d 速度 %d", cmd.CruiseID, tiltSpeed)
			return cmd, nil
		case 0x87:
			cmd.Action = PTZActionCruiseDwell
			cmd.Description = fmt.Sprintf("设置巡航组 %d 停留时间 %d 秒", cmd.CruiseID, tiltSpeed)
			return cmd, nil
		case 0x88:
			cmd.Action = PTZActionCruiseStart
			if cmd.CruiseID == 0 {
				cmd.CruiseID = presetID
			}
			cmd.Description = fmt.Sprintf("开始巡航组 %d", cmd.CruiseID)
			return cmd, nil
		case 0x89:
			cmd.Action = PTZActionCruiseStop
			if cmd.CruiseID == 0 {
				cmd.CruiseID = presetID
			}
			cmd.Description = fmt.Sprintf("停止巡航组 %d", cmd.CruiseID)
			return cmd, nil
		default:
			cmd.Action = PTZActionUnknown
			cmd.Description = fmt.Sprintf("未知扩展控制指令 0x%02X", cmdByte)
			return cmd, nil
		}
	}

	// 2. 停止指令 (0x00)
	if cmdByte == 0x00 {
		cmd.Action = PTZActionStop
		cmd.Description = "停止转动"
		return cmd, nil
	}

	// 3. 镜头变倍及云台运动 / 光圈聚焦辅助控制 - 严格遵循 GB/T 28181-2016 附录 A.3.1 表 A.1 规范
	cmd.Action = PTZActionMove
	var parts []string

	if (cmdByte & 0xC0) == 0x40 {
		// 辅助控制：聚焦与光圈扩展控制 (Bit 7=0, Bit 6=1，即 0x40 系列指令)
		// Byte 5 为聚焦控制速度，Byte 6 为光圈控制速度
		if cmdByte&0x02 != 0 {
			cmd.FocusDir = -1
			parts = append(parts, "聚焦近")
		} else if cmdByte&0x01 != 0 {
			cmd.FocusDir = 1
			parts = append(parts, "聚焦远")
		}
		if cmdByte&0x04 != 0 {
			cmd.IrisDir = -1
			parts = append(parts, "光圈关")
		} else if cmdByte&0x08 != 0 {
			cmd.IrisDir = 1
			parts = append(parts, "光圈开")
		}
	} else if (cmdByte & 0xC0) == 0x00 {
		// 云台移动与镜头变倍 (Bit 7=0, Bit 6=0)
		// 水平转动 (Bit 0: 右 0x01, Bit 1: 左 0x02)
		if cmdByte&0x01 != 0 {
			cmd.PanDir = 1
			parts = append(parts, "右")
		} else if cmdByte&0x02 != 0 {
			cmd.PanDir = -1
			parts = append(parts, "左")
		}

		// 垂直转动 (Bit 3: 上 0x08, Bit 2: 下 0x04)
		if cmdByte&0x08 != 0 {
			cmd.TiltDir = 1
			parts = append(parts, "上")
		} else if cmdByte&0x04 != 0 {
			cmd.TiltDir = -1
			parts = append(parts, "下")
		}

		// 镜头变倍 (Bit 4: 变倍加/放大/拉近 0x10, Bit 5: 变倍减/缩小/拉远 0x20)
		if cmdByte&0x10 != 0 {
			cmd.ZoomDir = 1
			parts = append(parts, "变倍放大")
		} else if cmdByte&0x20 != 0 {
			cmd.ZoomDir = -1
			parts = append(parts, "变倍缩小")
		}

		// 若指令包含变倍但平台未指定速度，赋予平滑默认速度
		if cmd.ZoomDir != 0 && cmd.ZoomSpeed <= 0 {
			cmd.ZoomSpeed = 4
		}
	}

	if len(parts) == 0 {
		cmd.Action = PTZActionUnknown
		cmd.Description = fmt.Sprintf("未知运动指令 0x%02X", cmdByte)
	} else {
		cmd.Description = strings.Join(parts, "")
		if cmd.PanDir != 0 && cmd.PanSpeed > 0 {
			cmd.Description += fmt.Sprintf(" 水平速度:%d", cmd.PanSpeed)
		}
		if cmd.TiltDir != 0 && cmd.TiltSpeed > 0 {
			cmd.Description += fmt.Sprintf(" 垂直速度:%d", cmd.TiltSpeed)
		}
		if cmd.ZoomDir != 0 && cmd.ZoomSpeed > 0 {
			cmd.Description += fmt.Sprintf(" 变倍速度:%d", cmd.ZoomSpeed)
		}
		if cmd.FocusDir != 0 && cmd.FocusSpeed > 0 {
			cmd.Description += fmt.Sprintf(" 聚焦速度:%d", cmd.FocusSpeed)
		}
		if cmd.IrisDir != 0 && cmd.IrisSpeed > 0 {
			cmd.Description += fmt.Sprintf(" 光圈速度:%d", cmd.IrisSpeed)
		}
	}

	return cmd, nil
}

// EncodePTZCmd 构造符合 GB/T 28181 标准的 8 字节 PTZCmd 十六进制字符串。
func EncodePTZCmd(cmd *PTZCommand) string {
	b := make([]byte, 8)
	b[0] = 0xA5
	b[1] = 0x0F // 版本与低4位校验
	b[2] = 0x01 // 设备逻辑地址低8位，标准GB通常为 0x01

	switch cmd.Action {
	case PTZActionStop:
		b[3] = 0x00
	case PTZActionPresetSet:
		b[3] = 0x81
		b[6] = byte(cmd.PresetID & 0xFF)
	case PTZActionPresetCall:
		b[3] = 0x82
		b[6] = byte(cmd.PresetID & 0xFF)
	case PTZActionPresetDelete:
		b[3] = 0x83
		b[6] = byte(cmd.PresetID & 0xFF)
	case PTZActionCruiseAdd:
		b[3] = 0x84
		b[4] = byte(cmd.CruiseID & 0xFF)
		b[6] = byte(cmd.PresetID & 0xFF)
	case PTZActionCruiseDel:
		b[3] = 0x85
		b[4] = byte(cmd.CruiseID & 0xFF)
		b[6] = byte(cmd.PresetID & 0xFF)
	case PTZActionCruiseSpeed:
		b[3] = 0x86
		b[4] = byte(cmd.CruiseID & 0xFF)
		b[5] = byte(cmd.PanSpeed & 0xFF)
	case PTZActionCruiseDwell:
		b[3] = 0x87
		b[4] = byte(cmd.CruiseID & 0xFF)
		b[5] = byte(cmd.PanSpeed & 0xFF)
	case PTZActionCruiseStart:
		b[3] = 0x88
		b[4] = byte(cmd.CruiseID & 0xFF)
	case PTZActionCruiseStop:
		b[3] = 0x89
		b[4] = byte(cmd.CruiseID & 0xFF)
	case PTZActionMove:
		var b3 byte
		// 严格遵循 GB/T 28181-2016 附录 A.3.1 表 A.1:
		// 0x01右, 0x02左, 0x08上, 0x04下, 0x20变倍放大(Zoom In), 0x10变倍缩小(Zoom Out)
		if cmd.PanDir > 0 {
			b3 |= 0x01
		} else if cmd.PanDir < 0 {
			b3 |= 0x02
		}
		if cmd.TiltDir > 0 {
			b3 |= 0x08
		} else if cmd.TiltDir < 0 {
			b3 |= 0x04
		}
		if cmd.ZoomDir > 0 {
			b3 |= 0x10 // 变倍加 (放大/拉近)
		} else if cmd.ZoomDir < 0 {
			b3 |= 0x20 // 变倍减 (缩小/拉远)
		}
		b[3] = b3
		b[4] = byte(cmd.PanSpeed & 0xFF)
		b[5] = byte(cmd.TiltSpeed & 0xFF)
		b[6] = byte((cmd.ZoomSpeed & 0x0F) << 4)
	}

	// 校验和：前 7 字节之和 % 256
	var sum uint32
	for i := 0; i < 7; i++ {
		sum += uint32(b[i])
	}
	b[7] = byte(sum & 0xFF)

	return strings.ToUpper(hex.EncodeToString(b))
}
