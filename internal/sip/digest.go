package sip

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ParseWWWAuthenticate 解析 WWW-Authenticate 头。
type AuthChallenge struct {
	Realm     string
	Nonce     string
	Algorithm string // MD5 / MD5-sess / SHA-256
	QOP       string // auth / auth-int
	Opaque    string
	Stale     string
}

func ParseWWWAuthenticate(h string) (*AuthChallenge, error) {
	c := &AuthChallenge{Algorithm: "MD5"}
	h = strings.TrimSpace(h)
	// 可能是 Digest realm="x", nonce="y", ...
	if i := strings.Index(strings.ToLower(h), "digest"); i >= 0 {
		h = h[i+len("digest"):]
	}
	for _, part := range splitAuthParams(h) {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(kv[0]))
		v := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		switch k {
		case "realm":
			c.Realm = v
		case "nonce":
			c.Nonce = v
		case "algorithm":
			c.Algorithm = v
		case "qop":
			c.QOP = v
		case "opaque":
			c.Opaque = v
		case "stale":
			c.Stale = v
		}
	}
	if c.Realm == "" || c.Nonce == "" {
		return nil, fmt.Errorf("incomplete digest challenge: %q", h)
	}
	return c, nil
}

func splitAuthParams(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '"':
			inQuote = !inQuote
			cur.WriteByte(ch)
		case ',':
			if inQuote {
				cur.WriteByte(ch)
			} else {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// BuildAuthorization 根据 challenge 构造 Authorization 头。
func BuildAuthorization(username, password, method, uri string, c *AuthChallenge, nc int, cnonce string) string {
	algo := strings.ToUpper(c.Algorithm)
	if algo == "" {
		algo = "MD5"
	}
	// HA1
	var ha1 string
	switch algo {
	case "SHA-256":
		ha1 = hashHex(username+":"+c.Realm+":"+password, "sha256")
	case "MD5-SESS":
		ha1 = hashHex(username+":"+c.Realm+":"+password, "md5")
		ha1 = hashHex(ha1+":"+c.Nonce+":"+cnonce, "md5")
	default: // MD5
		algo = "MD5"
		ha1 = hashHex(username+":"+c.Realm+":"+password, "md5")
	}
	ha2 := hashHex(method+":"+uri, hashAlgo(algo))
	ncStr := fmt.Sprintf("%08x", nc)

	var response string
	qop := ""
	if c.QOP != "" {
		// 取 auth
		for _, p := range strings.Split(c.QOP, ",") {
			p = strings.TrimSpace(p)
			if p == "auth" || p == "auth-int" {
				qop = p
				break
			}
		}
	}
	if qop == "auth" || qop == "auth-int" {
		response = hashHex(ha1+":"+c.Nonce+":"+ncStr+":"+cnonce+":"+qop+":"+ha2, hashAlgo(algo))
	} else {
		response = hashHex(ha1+":"+c.Nonce+":"+ha2, hashAlgo(algo))
	}

	var b strings.Builder
	fmt.Fprintf(&b, `Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s", algorithm=%s`,
		username, c.Realm, c.Nonce, uri, response, algo)
	if qop != "" {
		fmt.Fprintf(&b, ", qop=%s, nc=%s, cnonce=\"%s\"", qop, ncStr, cnonce)
	}
	if c.Opaque != "" {
		fmt.Fprintf(&b, `, opaque="%s"`, c.Opaque)
	}
	return b.String()
}

func hashAlgo(algo string) string {
	if strings.HasPrefix(strings.ToUpper(algo), "SHA-256") {
		return "sha256"
	}
	return "md5"
}

func hashHex(data, algo string) string {
	switch algo {
	case "sha256":
		sum := sha256.Sum256([]byte(data))
		return hex.EncodeToString(sum[:])
	default:
		sum := md5.Sum([]byte(data))
		return hex.EncodeToString(sum[:])
	}
}

func RandomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
