package sip

import (
	"strings"
	"testing"
)

func TestParseRequest(t *testing.T) {
	raw := "REGISTER sip:34020000002000000001@192.168.1.10:5060 SIP/2.0\r\n" +
		"Via: SIP/2.0/UDP 192.168.1.20:5070;rport;branch=z9hG4bKabc\r\n" +
		"From: <sip:34020000001180000001@192.168.1.10>;tag=123\r\n" +
		"To: <sip:34020000001180000001@192.168.1.10>\r\n" +
		"Call-ID: xyz@192.168.1.20\r\n" +
		"CSeq: 1 REGISTER\r\n" +
		"Contact: <sip:34020000001180000001@192.168.1.20:5070>\r\n" +
		"Expires: 3600\r\n" +
		"Content-Length: 0\r\n\r\n"

	m, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !m.IsRequest || m.Method != "REGISTER" {
		t.Fatalf("want REGISTER request, got %+v", m)
	}
	if m.Branch() != "z9hG4bKabc" {
		t.Fatalf("branch=%q", m.Branch())
	}
	if m.FromTag() != "123" {
		t.Fatalf("from tag=%q", m.FromTag())
	}
	if m.FromUser() != "34020000001180000001" {
		t.Fatalf("from user=%q", m.FromUser())
	}
}

func TestParseResponseWithBody(t *testing.T) {
	body := "v=0\r\n"
	raw := "SIP/2.0 200 OK\r\n" +
		"Via: SIP/2.0/UDP 1.2.3.4:5060;branch=z9hG4bK1\r\n" +
		"From: <sip:a@b>;tag=x\r\n" +
		"To: <sip:c@d>;tag=y\r\n" +
		"Call-ID: 1\r\n" +
		"CSeq: 1 INVITE\r\n" +
		"Content-Type: application/sdp\r\n" +
		"Content-Length: 7\r\n\r\n" + body
	m, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if m.StatusCode != 200 {
		t.Fatalf("status=%d", m.StatusCode)
	}
	if string(m.Body) != body {
		t.Fatalf("body=%q", m.Body)
	}
	if m.ToTag() != "y" {
		t.Fatalf("to tag=%q", m.ToTag())
	}
}

func TestBuildAuthorization(t *testing.T) {
	c := &AuthChallenge{
		Realm:     "3402000000",
		Nonce:     "abcdef",
		Algorithm: "MD5",
	}
	auth := BuildAuthorization("device", "pass", "REGISTER", "sip:device@host", c, 1, "cnonce1")
	if !strings.Contains(auth, `username="device"`) {
		t.Fatalf("auth=%s", auth)
	}
	if !strings.Contains(auth, "response=") {
		t.Fatalf("missing response: %s", auth)
	}
}

func TestBytesRoundTrip(t *testing.T) {
	req := NewRequest("MESSAGE", "sip:x@y")
	req.SetHeader("Via", "SIP/2.0/UDP 1.1.1.1:1;branch=z9hG4bK1")
	req.SetHeader("From", "<sip:a@b>;tag=1")
	req.SetHeader("To", "<sip:c@d>")
	req.SetHeader("Call-ID", "cid")
	req.SetHeader("CSeq", "2 MESSAGE")
	req.SetHeader("Content-Type", "Application/MANSCDP+xml")
	req.Body = []byte("<Notify></Notify>")
	data := req.Bytes()
	if !strings.Contains(string(data), "Content-Length: 17") {
		t.Fatalf("data=\n%s", data)
	}
	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if m.Method != "MESSAGE" || len(m.Body) != 17 {
		t.Fatalf("method=%s bodyLen=%d", m.Method, len(m.Body))
	}
}
