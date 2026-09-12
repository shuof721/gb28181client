package device

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/local/gb28181-device/internal/config"
)

func TestDeviceVoiceTalkInviteAndBye(t *testing.T) {
	// 启动模拟 SIP 平台接收 UDP 端口
	platformConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer platformConn.Close()
	platPort := platformConn.LocalAddr().(*net.UDPAddr).Port

	// 启动模拟设备
	devPort := 15870
	cfg := &config.Config{
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: platPort,
			LocalIP:    "127.0.0.1",
			LocalPort:  devPort,
			Transport:  "udp",
			Username:   "34020000001180000005",
			Password:   "12345678",
			Expires:    3600,
		},
		Device: config.DeviceConfig{
			ID:   "34020000001180000005",
			Name: "Test-Talk-Device",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000005", Name: "Talk-IPC"},
			},
		},
		Media: config.MediaConfig{
			LocalIP:        "127.0.0.1",
			FPS:            25,
			RTPPayloadMax:  1400,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatalf("start device: %v", err)
	}
	defer dev.Stop()

	// 模拟平台发起语音对讲 INVITE
	audioPort := 32000
	sdp := fmt.Sprintf(
		"v=0\r\n"+
			"o=34020000002000000001 0 0 IN IP4 127.0.0.1\r\n"+
			"s=Talk\r\n"+
			"c=IN IP4 127.0.0.1\r\n"+
			"t=0 0\r\n"+
			"m=audio %d RTP/AVP 8\r\n"+
			"a=sendrecv\r\n"+
			"a=rtpmap:8 PCMA/8000\r\n"+
			"y=0200000001\r\n", audioPort)

	callID := "talk-callid-9999"
	inviteReq := fmt.Sprintf(
		"INVITE sip:34020000001320000005@127.0.0.1:%d SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP 127.0.0.1:%d;branch=z9hG4bKtalktest\r\n"+
			"From: <sip:34020000002000000001@127.0.0.1:%d>;tag=fromtalk\r\n"+
			"To: <sip:34020000001320000005@127.0.0.1:%d>\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: 1 INVITE\r\n"+
			"Contact: <sip:34020000002000000001@127.0.0.1:%d>\r\n"+
			"Content-Type: Application/SDP\r\n"+
			"Subject: 34020000001320000005:0200000001,34020000001180000005:0\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		devPort, platPort, platPort, devPort, callID, platPort, len(sdp), sdp)

	devAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: devPort}
	if _, err := platformConn.WriteToUDP([]byte(inviteReq), devAddr); err != nil {
		t.Fatalf("send invite: %v", err)
	}

	// 平台接收设备 200 OK (可能先收到 REGISTER，循环读取直到收到包含 m=audio 的 200 OK)
	buf := make([]byte, 4096)
	var respStr string
	for i := 0; i < 5; i++ {
		_ = platformConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := platformConn.ReadFrom(buf)
		if err != nil {
			t.Fatalf("receive invite 200 ok: %v", err)
		}
		s := string(buf[:n])
		if strings.Contains(s, "SIP/2.0 200 OK") && strings.Contains(s, "m=audio") {
			respStr = s
			break
		}
	}
	if respStr == "" {
		t.Fatalf("expected 200 OK with audio SDP, got none")
	}
	if !strings.Contains(respStr, "m=audio") || !strings.Contains(respStr, "PCMA/8000") {
		t.Fatalf("expected audio SDP in 200 OK, got: %s", respStr)
	}

	// 验证状态中包含 TalkSession
	time.Sleep(50 * time.Millisecond)
	st := dev.Status()
	if len(st.TalkSessions) != 1 {
		t.Fatalf("expected 1 talk session, got %d", len(st.TalkSessions))
	}
	if st.TalkSessions[0].CallID != callID {
		t.Errorf("expected callID %s, got %s", callID, st.TalkSessions[0].CallID)
	}

	// 模拟平台发送 BYE
	byeReq := fmt.Sprintf(
		"BYE sip:34020000001320000005@127.0.0.1:%d SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP 127.0.0.1:%d;branch=z9hG4bKbyetest\r\n"+
			"From: <sip:34020000002000000001@127.0.0.1:%d>;tag=fromtalk\r\n"+
			"To: <sip:34020000001320000005@127.0.0.1:%d>;tag=totalk\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: 2 BYE\r\n"+
			"Content-Length: 0\r\n\r\n",
		devPort, platPort, platPort, devPort, callID)

	if _, err := platformConn.WriteToUDP([]byte(byeReq), devAddr); err != nil {
		t.Fatalf("send bye: %v", err)
	}

	// 接收 200 OK
	var byeRespStr string
	for i := 0; i < 5; i++ {
		_ = platformConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := platformConn.ReadFrom(buf)
		if err != nil {
			t.Fatalf("receive bye 200 ok: %v", err)
		}
		s := string(buf[:n])
		if strings.Contains(s, "SIP/2.0 200 OK") && strings.Contains(s, callID) {
			byeRespStr = s
			break
		}
	}
	if byeRespStr == "" {
		t.Fatalf("expected 200 OK for BYE")
	}

	// 验证 TalkSession 已经清除
	time.Sleep(50 * time.Millisecond)
	st = dev.Status()
	if len(st.TalkSessions) != 0 {
		t.Errorf("expected 0 talk sessions after BYE, got %d", len(st.TalkSessions))
	}
}
