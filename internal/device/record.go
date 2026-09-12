package device

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/gb28181"
)

// ParseFlexibleTime 解析国标平台可能发送的各种时间戳格式。
func ParseFlexibleTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// 1. 14 位 YYYYMMDDHHmmss 或 8 位 YYYYMMDD 紧凑格式
	if len(s) == 14 {
		if t, err := time.ParseInLocation("20060102150405", s, time.Local); err == nil {
			return t, nil
		}
	} else if len(s) == 8 {
		if t, err := time.ParseInLocation("20060102", s, time.Local); err == nil {
			return t, nil
		}
	}

	// 2. 常见日期时间格式（含分隔符）
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}

	// 3. 尝试纯数字 Unix 时间戳（秒级 10 位或毫秒级 13 位）
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
		if len(s) == 13 { // 毫秒
			return time.UnixMilli(n), nil
		}
		if len(s) == 10 { // 秒
			return time.Unix(n, 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized time format: %q", s)
}

// GenerateRecordItems 根据设备录像排程配置动态生成匹配的录像切片列表。
func GenerateRecordItems(
	cfg config.RecordConfig,
	channelID string,
	channelName string,
	nvrID string,
	startStr string,
	endStr string,
	queryType string,
) []gb28181.RecordItem {
	if !cfg.Enabled {
		return []gb28181.RecordItem{}
	}

	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if t, err := ParseFlexibleTime(startStr); err == nil && !t.IsZero() {
		startTime = t
	}
	if t, err := ParseFlexibleTime(endStr); err == nil && !t.IsZero() {
		endTime = t
	}

	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}

	// 1. 保留天数过滤（默认最近 RetainDays 天，更早的录像已被循环覆盖）
	retainDays := cfg.RetainDays
	if retainDays <= 0 {
		retainDays = 7
	}
	earliest := now.AddDate(0, 0, -retainDays)
	earliestDay := time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 0, 0, 0, 0, earliest.Location())

	if endTime.Before(earliestDay) {
		return []gb28181.RecordItem{}
	}
	if startTime.Before(earliestDay) {
		startTime = earliestDay
	}

	// 2. 不能查询未来的时间
	if startTime.After(now) {
		return []gb28181.RecordItem{}
	}
	if endTime.After(now) {
		endTime = now
	}

	sliceMinutes := cfg.SliceMinutes
	if sliceMinutes <= 0 {
		sliceMinutes = 60
	}
	sliceDur := time.Duration(sliceMinutes) * time.Minute

	// 单切片默认大小 (字节)
	baseFileSizeMB := cfg.FileSizeMB
	if baseFileSizeMB <= 0 {
		baseFileSizeMB = 100
	}
	baseBytes := int64(baseFileSizeMB) * 1024 * 1024

	queryType = strings.ToLower(strings.TrimSpace(queryType))
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = "continuous"
	}

	var items []gb28181.RecordItem
	maxItems := 500 // 安全上限，防止过长跨度耗尽资源

	cur := startTime
	for cur.Before(endTime) && len(items) < maxItems {
		segEnd := cur.Add(sliceDur)
		if segEnd.After(endTime) {
			segEnd = endTime
		}

		itemType := "time"
		includeSlice := true

		switch mode {
		case "work_hours":
			// 工作时段模式：仅 08:00 ~ 18:00 生成录像
			h := cur.Hour()
			if h < 8 || h >= 18 {
				includeSlice = false
			}
		case "alarm":
			// 报警模式：所有切片均为 alarm 类型
			itemType = "alarm"
		default:
			// continuous 连续全天模式：可偶尔在特定时刻标记一段报警录像（如逢 4 小时一段报警）
			if cur.Hour()%4 == 0 && cur.Minute() < 30 {
				itemType = "alarm"
			}
		}

		// 检查查询类型过滤
		if includeSlice {
			if queryType == "alarm" && itemType != "alarm" {
				includeSlice = false
			} else if queryType == "time" && itemType != "time" {
				includeSlice = false
			}
		}

		if includeSlice {
			// 按时长折算文件大小
			sizeBytes := baseBytes
			if segEnd.Sub(cur) < sliceDur {
				sizeBytes = baseBytes * int64(segEnd.Sub(cur)) / int64(sliceDur)
				if sizeBytes <= 0 {
					sizeBytes = 1024 * 1024
				}
			}

			filePath := fmt.Sprintf("/record/%s/%s/%s_%s.mp4",
				channelID,
				cur.Format("20060102"),
				cur.Format("150405"),
				segEnd.Format("150405"),
			)

			items = append(items, gb28181.RecordItem{
				DeviceID:   channelID,
				Name:       channelName,
				FilePath:   filePath,
				Address:    "LocalDisk",
				StartTime:  cur.Format("2006-01-02T15:04:05"),
				EndTime:    segEnd.Format("2006-01-02T15:04:05"),
				Secrecy:    0,
				Type:       itemType,
				RecorderID: nvrID,
				FileSize:   strconv.FormatInt(sizeBytes, 10),
			})
		}

		cur = segEnd
	}

	return items
}
