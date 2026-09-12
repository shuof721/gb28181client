package gb28181

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// MANSRTSPRequest 解析 GB/T 28181-2016 附录 F 规定的基于 RTSP 的媒体控制请求。
type MANSRTSPRequest struct {
	Method    string  // PLAY, PAUSE, TEARDOWN
	URI       string  // RTSP URL
	Version   string  // RTSP/1.0
	CSeq      string  // CSeq 序号
	Scale     float64 // 倍速: 0.25, 0.5, 1.0, 2.0, 4.0
	HasScale  bool
	RangeNPT  float64 // Range: npt=xxx 秒
	HasRange  bool
	RawRange  string
	PauseTime string
}

// ParseMANSRTSP 从 SIP INFO 消息体中解析 MANSRTSP 请求。
func ParseMANSRTSP(data []byte) (*MANSRTSPRequest, error) {
	req := &MANSRTSPRequest{Scale: 1.0}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	firstLine := true

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if firstLine {
			firstLine = false
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				req.Method = strings.ToUpper(parts[0])
			}
			if len(parts) >= 2 {
				req.URI = parts[1]
			}
			if len(parts) >= 3 {
				req.Version = parts[2]
			}
			continue
		}

		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		switch strings.ToLower(key) {
		case "cseq":
			req.CSeq = val
		case "scale":
			if s, err := strconv.ParseFloat(val, 64); err == nil {
				req.Scale = s
				req.HasScale = true
			}
		case "range":
			req.RawRange = val
			// Range: npt=120.000000- 或 npt=now- 或 npt=120-200
			if strings.HasPrefix(strings.ToLower(val), "npt=") {
				nptPart := val[4:]
				if dash := strings.Index(nptPart, "-"); dash >= 0 {
					nptPart = nptPart[:dash]
				}
				if nptPart != "now" && nptPart != "" {
					if sec, err := strconv.ParseFloat(strings.TrimSpace(nptPart), 64); err == nil {
						req.RangeNPT = sec
						req.HasRange = true
					}
				}
			}
		case "pausetime":
			req.PauseTime = val
		}
	}
	return req, scanner.Err()
}

// BuildMANSRTSPResponse 构造响应 MANSRTSP 的 200 OK 文本。
func BuildMANSRTSPResponse(cseq string, scale float64, rangeNPT float64, isPause bool) []byte {
	var b strings.Builder
	b.WriteString("RTSP/1.0 200 OK\r\n")
	if cseq != "" {
		fmt.Fprintf(&b, "CSeq: %s\r\n", cseq)
	}
	if !isPause {
		if scale <= 0 {
			scale = 1.0
		}
		fmt.Fprintf(&b, "Scale: %.6f\r\n", scale)
		fmt.Fprintf(&b, "Range: npt=%.6f-\r\n", rangeNPT)
	}
	b.WriteString("\r\n")
	return []byte(b.String())
}
