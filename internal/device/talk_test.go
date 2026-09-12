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

func TestDeviceBroadcastNotifyAndInvite(t *testing.T) {
	// 启动模拟 SIP 平台接收 UDP 端口
	platformConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer platformConn.Close()
	platPort := platformConn.LocalAddr().(*net.UDPAddr).Port

	// 启动模拟设备
	devPort := 15872
	cfg := &config.Config{
		SIP: config.SIPConfig{
			ServerIP:   "127.0.0.1",
			ServerPort: platPort,
			LocalIP:    "127.0.0.1",
			LocalPort:  devPort,
			Transport:  "udp",
			Username:   "34020000001180000006",
			Password:   "12345678",
			Expires:    3600,
		},
		Device: config.DeviceConfig{
			ID:     "34020000001180000006",
			Domain: "3402000000",
			Name:   "Test-Broadcast-Device",
			Channels: []config.ChannelConfig{
				{ID: "34020000001320000006", Name: "Broadcast-IPC"},
			},
		},
		Media: config.MediaConfig{
			LocalIP:       "127.0.0.1",
			FPS:           25,
			RTPPayloadMax: 1400,
		},
	}

	dev := New(cfg)
	if err := dev.Start(); err != nil {
		t.Fatalf("start device: %v", err)
	}
	defer dev.Stop()

	// 1. 模拟平台下发语音广播 Notify (MESSAGE)
	notifyXML := `<?xml version="1.0"?>
<Notify>
  <CmdType>Broadcast</CmdType>
  <SN>888123</SN>
  <SourceID>34020000002000000001</SourceID>
  <TargetID>34020000001320000006</TargetID>
</Notify>`

	notifyReq := fmt.Sprintf(
		"MESSAGE sip:34020000001180000006@127.0.0.1:%d SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP 127.0.0.1:%d;branch=z9hG4bKnotifymsg\r\n"+
			"From: <sip:34020000002000000001@127.0.0.1:%d>;tag=fromplat\r\n"+
			"To: <sip:34020000001180000006@127.0.0.1:%d>\r\n"+
			"Call-ID: notify-callid-123\r\n"+
			"CSeq: 1 MESSAGE\r\n"+
			"Content-Type: Application/MANSCDP+xml\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		devPort, platPort, platPort, devPort, len(notifyXML), notifyXML)

	devAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: devPort}
	if _, err := platformConn.WriteToUDP([]byte(notifyReq), devAddr); err != nil {
		t.Fatalf("send notify: %v", err)
	}

	// 2. 平台接收设备发出的 INVITE (中间可能收到对 MESSAGE 的 200 OK 或 MESSAGE Response)
	buf := make([]byte, 4096)
	var inviteStr string
	var inviteFrom, inviteTo, inviteCallID, inviteCSeq, inviteVia string
	for i := 0; i < 10; i++ {
		_ = platformConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := platformConn.ReadFrom(buf)
		if err != nil {
			t.Fatalf("read from device timeout: %v", err)
		}
		s := string(buf[:n])
		if strings.HasPrefix(s, "REGISTER ") || strings.HasPrefix(s, "MESSAGE ") {
			resp := "SIP/2.0 200 OK\r\n"
			for _, l := range strings.Split(s, "\r\n") {
				if strings.HasPrefix(l, "Via:") || strings.HasPrefix(l, "From:") || strings.HasPrefix(l, "To:") || strings.HasPrefix(l, "Call-ID:") || strings.HasPrefix(l, "CSeq:") {
					resp += l + "\r\n"
				}
			}
			resp += "Content-Length: 0\r\n\r\n"
			_, _ = platformConn.WriteToUDP([]byte(resp), devAddr)
			continue
		}
		if strings.HasPrefix(s, "INVITE ") && strings.Contains(s, "s=Broadcast") {
			inviteStr = s
			// 提取头部
			for _, line := range strings.Split(s, "\r\n") {
				if strings.HasPrefix(line, "Via:") {
					inviteVia = strings.TrimSpace(line[4:])
				} else if strings.HasPrefix(line, "From:") {
					inviteFrom = strings.TrimSpace(line[5:])
				} else if strings.HasPrefix(line, "To:") {
					inviteTo = strings.TrimSpace(line[3:])
				} else if strings.HasPrefix(line, "Call-ID:") {
					inviteCallID = strings.TrimSpace(line[8:])
				} else if strings.HasPrefix(line, "CSeq:") {
					inviteCSeq = strings.TrimSpace(line[5:])
				}
			}
			break
		}
	}

	if inviteStr == "" {
		t.Fatalf("expected device to send outgoing INVITE with s=Broadcast, got none")
	}
	if !strings.Contains(inviteStr, "m=audio") || !strings.Contains(inviteStr, "a=recvonly") {
		t.Errorf("expected audio recvonly in INVITE offer, got:\n%s", inviteStr)
	}

	// 3. 平台回复 200 OK
	platSDP := fmt.Sprintf(
		"v=0\r\n"+
			"o=34020000002000000001 0 0 IN IP4 127.0.0.1\r\n"+
			"s=Broadcast\r\n"+
			"c=IN IP4 127.0.0.1\r\n"+
			"t=0 0\r\n"+
			"m=audio 35000 RTP/AVP 8\r\n"+
			"a=sendonly\r\n"+
			"a=rtpmap:8 PCMA/8000\r\n"+
			"y=0200000002\r\n")

	toWithTag := inviteTo + ";tag=plattag123"
	ok200 := fmt.Sprintf(
		"SIP/2.0 200 OK\r\n"+
			"Via: %s\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %s\r\n"+
			"Contact: <sip:34020000002000000001@127.0.0.1:%d>\r\n"+
			"Content-Type: Application/SDP\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		inviteVia, inviteFrom, toWithTag, inviteCallID, inviteCSeq, platPort, len(platSDP), platSDP)

	if _, err := platformConn.WriteToUDP([]byte(ok200), devAddr); err != nil {
		t.Fatalf("send 200 ok: %v", err)
	}

	// 4. 平台接收设备的 ACK
	var ackStr string
	for i := 0; i < 5; i++ {
		_ = platformConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := platformConn.ReadFrom(buf)
		if err != nil {
			t.Fatalf("read ACK timeout: %v", err)
		}
		s := string(buf[:n])
		if strings.HasPrefix(s, "ACK ") && strings.Contains(s, inviteCallID) {
			ackStr = s
			break
		}
	}
	if ackStr == "" {
		t.Fatalf("expected ACK from device, got none")
	}

	// 5. 验证会话已建立为 broadcast
	time.Sleep(50 * time.Millisecond)
	st := dev.Status()
	if len(st.TalkSessions) != 1 {
		t.Fatalf("expected 1 talk session, got %d", len(st.TalkSessions))
	}
	sess := st.TalkSessions[0]
	if sess.StreamType != "broadcast" {
		t.Errorf("expected streamType broadcast, got %s", sess.StreamType)
	}
	if sess.RemotePort != 35000 {
		t.Errorf("expected RemotePort 35000, got %d", sess.RemotePort)
	}

	// 6. 设备端主动 StopTalk 挂断广播并向平台发送 BYE
	dev.StopTalk(sess.CallID)

	var byeStr string
	for i := 0; i < 5; i++ {
		_ = platformConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := platformConn.ReadFrom(buf)
		if err != nil {
			break
		}
		s := string(buf[:n])
		if strings.HasPrefix(s, "BYE ") && strings.Contains(s, sess.CallID) {
			byeStr = s
			break
		}
	}
	if byeStr == "" {
		t.Errorf("expected outgoing BYE from device, got none")
	}
}
