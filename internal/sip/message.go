package sip

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Message 表示一条 SIP 报文（请求或响应）。
type Message struct {
	IsRequest  bool
	Method     string
	RequestURI string
	StatusCode int
	Reason     string

	Headers map[string][]string
	Body    []byte

	// 原始起始行，便于调试
	StartLine string
}

func NewRequest(method, requestURI string) *Message {
	return &Message{
		IsRequest:  true,
		Method:     method,
		RequestURI: requestURI,
		Headers:    map[string][]string{},
	}
}

func NewResponse(status int, reason string) *Message {
	return &Message{
		IsRequest:  false,
		StatusCode: status,
		Reason:     reason,
		Headers:    map[string][]string{},
	}
}

func (m *Message) SetHeader(name, value string) {
	m.Headers[strings.ToLower(name)] = []string{value}
}

func (m *Message) AddHeader(name, value string) {
	k := strings.ToLower(name)
	m.Headers[k] = append(m.Headers[k], value)
}

func (m *Message) GetHeader(name string) string {
	vs := m.Headers[strings.ToLower(name)]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (m *Message) GetHeaders(name string) []string {
	return m.Headers[strategies(name)]
}

func strategies(name string) string {
	return strings.ToLower(name)
}

func (m *Message) Clone() *Message {
	c := &Message{
		IsRequest:  m.IsRequest,
		Method:     m.Method,
		RequestURI: m.RequestURI,
		StatusCode: m.StatusCode,
		Reason:     m.Reason,
		StartLine:  m.StartLine,
		Headers:    make(map[string][]string, len(m.Headers)),
		Body:       append([]byte(nil), m.Body...),
	}
	for k, vs := range m.Headers {
		c.Headers[k] = append([]string(nil), vs...)
	}
	return c
}

// CSeq 返回 (seq, method)
func (m *Message) CSeq() (int, string, error) {
	raw := m.GetHeader("CSeq")
	if raw == "" {
		return 0, "", fmt.Errorf("missing CSeq")
	}
	parts := strings.Fields(raw)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("bad CSeq: %q", raw)
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("bad CSeq number: %w", err)
	}
	return n, strings.ToUpper(parts[1]), nil
}

func (m *Message) CallID() string { return m.GetHeader("Call-ID") }

func (m *Message) Branch() string {
	via := m.GetHeader("Via")
	// Via: SIP/2.0/UDP host:port;branch=z9hG4bKxxx;rport
	for _, part := range strings.Split(via, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "branch=") {
			return strings.TrimSpace(part[len("branch="):])
		}
	}
	return ""
}

func (m *Message) FromTag() string {
	return headerParam(m.GetHeader("From"), "tag")
}

func (m *Message) ToTag() string {
	return headerParam(m.GetHeader("To"), "tag")
}

func (m *Message) FromUser() string {
	return uriUser(m.GetHeader("From"))
}

func (m *Message) ToUser() string {
	return uriUser(m.GetHeader("To"))
}

func headerParam(header, key string) string {
	// 可能形如: "Name" <sip:user@host>;tag=xxx
	idx := strings.Index(strings.ToLower(header), strings.ToLower(key)+"=")
	if idx < 0 {
		return ""
	}
	rest := header[idx+len(key)+1:]
	// 截断到下一个 ;
	if i := strings.Index(rest, ";"); i >= 0 {
		rest = rest[:i]
	}
	return strings.Trim(rest, " \t\"")
}

func uriUser(header string) string {
	// 提取 sip:user@host 中的 user
	lt := strings.Index(header, "<")
	gt := strings.Index(header, ">")
	uri := header
	if lt >= 0 && gt > lt {
		uri = header[lt+1 : gt]
	}
	uri = strings.TrimSpace(uri)
	uri = strings.TrimPrefix(uri, "sip:")
	uri = strings.TrimPrefix(uri, "sips:")
	if i := strings.Index(uri, "@"); i >= 0 {
		return uri[:i]
	}
	if i := strings.IndexAny(uri, ";>"); i >= 0 {
		return uri[:i]
	}
	return uri
}

func (m *Message) Bytes() []byte {
	var b bytes.Buffer
	if m.IsRequest {
		fmt.Fprintf(&b, "%s %s SIP/2.0\r\n", m.Method, m.RequestURI)
	} else {
		reason := m.Reason
		if reason == "" {
			reason = statusText(m.StatusCode)
		}
		fmt.Fprintf(&b, "SIP/2.0 %d %s\r\n", m.StatusCode, reason)
	}
	// 保证 Content-Length
	bodyLen := len(m.Body)
	if m.Headers == nil {
		m.Headers = map[string][]string{}
	}
	// 用稳定顺序写出常用头，再写其余
	written := map[string]bool{
		"content-length": true,
	}
	write := func(name string) {
		k := strings.ToLower(name)
		vs, ok := m.Headers[k]
		if !ok {
			return
		}
		for _, v := range vs {
			fmt.Fprintf(&b, "%s: %s\r\n", canonicalHeader(k), v)
		}
		written[k] = true
	}
	for _, name := range []string{
		"Via", "From", "To", "Call-ID", "CSeq", "Contact",
		"Max-Forwards", "Expires", "User-Agent", "Subject",
		"Content-Type", "Authorization", "WWW-Authenticate",
		"Date", "Allow", "Supported", "Session-Expires",
	} {
		write(name)
	}
	// 其余头
	keys := make([]string, 0, len(m.Headers))
	for k := range m.Headers {
		if !written[k] {
			keys = append(keys, k)
		}
	}
	// 简单排序保持稳定
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for _, k := range keys {
		for _, v := range m.Headers[k] {
			fmt.Fprintf(&b, "%s: %s\r\n", canonicalHeader(k), v)
		}
	}
	fmt.Fprintf(&b, "Content-Length: %d\r\n\r\n", bodyLen)
	b.Write(m.Body)
	return b.Bytes()
}

func canonicalHeader(lower string) string {
	switch lower {
	case "call-id":
		return "Call-ID"
	case "cseq":
		return "CSeq"
	case "www-authenticate":
		return "WWW-Authenticate"
	case "max-forwards":
		return "Max-Forwards"
	case "content-type":
		return "Content-Type"
	case "content-length":
		return "Content-Length"
	case "user-agent":
		return "User-Agent"
	}
	// 其余保持原样（SIP 头大小写不敏感）
	if lower == "via" {
		return "Via"
	}
	// Title-case 简单处理
	parts := strings.Split(lower, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "-")
}

func statusText(code int) string {
	switch code {
	case 100:
		return "Trying"
	case 180:
		return "Ringing"
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 408:
		return "Request Timeout"
	case 480:
		return "Temporarily Unavailable"
	case 481:
		return "Call/Transaction Does Not Exist"
	case 486:
		return "Busy Here"
	case 487:
		return "Request Terminated"
	case 488:
		return "Not Acceptable Here"
	case 500:
		return "Server Internal Error"
	case 503:
		return "Service Unavailable"
	case 603:
		return "Decline"
	default:
		return "OK"
	}
}

// Parse 解析完整 SIP 报文。
func Parse(data []byte) (*Message, error) {
	// 分离 header / body
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	sepLen := 4
	if idx < 0 {
		idx = bytes.Index(data, []byte("\n\n"))
		sepLen = 2
	}
	if idx < 0 {
		return nil, fmt.Errorf("incomplete SIP message")
	}
	head := data[:idx]
	body := data[idx+sepLen:]

	lines := splitLines(head)
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty SIP message")
	}
	start := lines[0]
	m := &Message{Headers: map[string][]string{}, StartLine: start, Body: append([]byte(nil), body...)}

	if strings.HasPrefix(start, "SIP/2.0") {
		m.IsRequest = false
		parts := strings.SplitN(start, " ", 3)
		if len(parts) < 2 {
			return nil, fmt.Errorf("bad status line: %q", start)
		}
		code, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("bad status code: %w", err)
		}
		m.StatusCode = code
		if len(parts) == 3 {
			m.Reason = strings.TrimSpace(parts[2])
		}
	} else {
		m.IsRequest = true
		parts := strings.Fields(start)
		if len(parts) < 3 {
			return nil, fmt.Errorf("bad request line: %q", start)
		}
		m.Method = strings.ToUpper(parts[0])
		m.RequestURI = parts[1]
	}

	// 折叠头简单处理
	var lastKey string
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t') && lastKey != "" {
			vs := m.Headers[lastKey]
			if len(vs) > 0 {
				vs[len(vs)-1] += " " + strings.TrimSpace(line)
			}
			continue
		}
		i := strings.Index(line, ":")
		if i <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:i]))
		val := strings.TrimSpace(line[i+1:])
		m.Headers[key] = append(m.Headers[key], val)
		lastKey = key
	}

	// Content-Length 校验（若存在则以头部为准截断）
	if cl := m.GetHeader("Content-Length"); cl != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(cl)); err == nil {
			if n >= 0 && n <= len(m.Body) {
				m.Body = m.Body[:n]
			}
		}
	}
	return m, nil
}

func splitLines(b []byte) []string {
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		out = append(out, l)
	}
	return out
}
