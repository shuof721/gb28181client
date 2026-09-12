package media

import (
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestTalkManagerAndSession(t *testing.T) {
	// 创建本地 UDP 接收端模拟 WVP/ZLM 接收设备音频
	platformRecv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer platformRecv.Close()
	pAddr := platformRecv.LocalAddr().(*net.UDPAddr)

	tm := NewTalkManager("127.0.0.1")
	sdp := &SDPInfo{
		IP:          "127.0.0.1",
		AudioPort:   pAddr.Port,
		SessionName: "Talk",
		SSRC:        "0200000001",
	}

	callID := "test-talk-callid"
	channelID := "34020000001320000001"
	sess, err := tm.StartTalkSession(channelID, callID, sdp, "talk")
	if err != nil {
		t.Fatalf("start talk session error: %v", err)
	}
	defer sess.Stop()

	// 1. 验证设备上行音频：平台端能收到来自设备端的 PCMA RTP 包
	buf := make([]byte, 1024)
	_ = platformRecv.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, _, err := platformRecv.ReadFrom(buf)
	if err != nil {
		t.Fatalf("failed to receive uplink audio RTP packet: %v", err)
	}
	if n < 12+160 {
		t.Fatalf("uplink packet size %d too small, expected >= 172", n)
	}
	payload, pt, _, _, err := RTPDepacketizeAudio(buf[:n])
	if err != nil {
		t.Fatalf("failed to depacketize uplink rtp: %v", err)
	}
	if pt != 8 {
		t.Errorf("expected pt 8, got %d", pt)
	}
	if len(payload) != 160 {
		t.Errorf("expected 160 bytes payload, got %d", len(payload))
	}

	// 2. 验证设备下行音频接收与监听者分发
	pcmCh := make(chan []byte, 10)
	sess.RegisterListener(pcmCh)
	defer sess.UnregisterListener(pcmCh)

	// 模拟平台向设备端 localPort 发送一个 RTP 音频包
	devAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: sess.LocalPort}
	sinePcm := NewToneGenerator(440, 8000).NextFrame(160, 0.7)
	testPayload := PCM16ToALaw(sinePcm)
	testSeq := uint16(1)
	testPkt := RTPPacketizeAudio(testPayload, 8, 0x12345678, &testSeq, 0)

	if _, err := platformRecv.WriteToUDP(testPkt, devAddr); err != nil {
		t.Fatalf("failed to send downlink rtp packet: %v", err)
	}

	select {
	case receivedPCM := <-pcmCh:
		if len(receivedPCM) != 320 {
			t.Errorf("expected 320 bytes PCM16, got %d", len(receivedPCM))
		}
		// 验证电平计算有跳动
		time.Sleep(20 * time.Millisecond)
		if sess.GetRxVol() <= 0 {
			t.Errorf("expected non-zero rx volume, got %f", sess.GetRxVol())
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for downlink pcm listener")
	}

	// 3. 验证麦克风推流模式切换
	sess.SetUplinkMode("mic")
	if sess.UplinkMode() != "mic" {
		t.Errorf("expected mic mode, got %s", sess.UplinkMode())
	}
	micSamples := make([]int16, 160)
	for i := range micSamples {
		micSamples[i] = 10000
	}
	sess.PushMicPCM(micSamples)

	// 4. 验证会话列表
	list := tm.ListSessions()
	if len(list) != 1 {
		t.Fatalf("expected 1 session in list, got %d", len(list))
	}
	if list[0].CallID != callID {
		t.Errorf("expected callID %s, got %s", callID, list[0].CallID)
	}

	// 5. 验证停止并清除
	tm.StopByCallID(callID)
	if tm.GetSession(callID) != nil {
		t.Error("expected session to be stopped and cleaned")
	}
	if len(tm.ListSessions()) != 0 {
		t.Error("expected empty session list after stop")
	}
}

func TestTalkManagerTCP(t *testing.T) {
	// 模拟 WVP / ZLM 开启 TCP RTP 监听端口
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	tcpPort := ln.Addr().(*net.TCPAddr).Port

	tm := NewTalkManager("127.0.0.1")
	sdp := &SDPInfo{
		IP:          "127.0.0.1",
		AudioPort:   tcpPort,
		SessionName: "Talk",
		SSRC:        "0200006530",
		IsTCP:       true,
		TCPMode:     "passive",
	}

	callID := "test-tcp-talk-callid"
	channelID := "34020000001320000001"

	// 异步接收设备连接
	connCh := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			connCh <- conn
		}
	}()

	sess, err := tm.StartTalkSession(channelID, callID, sdp, "talk")
	if err != nil {
		t.Fatalf("start TCP talk session error: %v", err)
	}
	defer sess.Stop()

	var platformConn net.Conn
	select {
	case platformConn = <-connCh:
		defer platformConn.Close()
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for device to establish TCP connection")
	}

	// 1. 验证设备上行音频：通过 TCP (RFC 4571 2 字节大端前缀) 接收设备端音频包
	_ = platformConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var lenBuf [2]byte
	if _, err := io.ReadFull(platformConn, lenBuf[:]); err != nil {
		t.Fatalf("failed to read RFC 4571 packet length: %v", err)
	}
	pktLen := int(binary.BigEndian.Uint16(lenBuf[:]))
	if pktLen < 12+160 {
		t.Fatalf("TCP uplink packet size %d too small, expected >= 172", pktLen)
	}
	pktBuf := make([]byte, pktLen)
	if _, err := io.ReadFull(platformConn, pktBuf); err != nil {
		t.Fatalf("failed to read TCP packet body: %v", err)
	}

	payload, pt, _, _, err := RTPDepacketizeAudio(pktBuf)
	if err != nil {
		t.Fatalf("failed to depacketize TCP audio RTP: %v", err)
	}
	if pt != 8 {
		t.Errorf("expected pt 8, got %d", pt)
	}
	if len(payload) != 160 {
		t.Errorf("expected 160 bytes payload, got %d", len(payload))
	}

	// 2. 验证下行音频 TCP 接收：模拟平台向设备端通过同一 TCP 链路下发音频
	pcmCh := make(chan []byte, 10)
	sess.RegisterListener(pcmCh)
	defer sess.UnregisterListener(pcmCh)

	sinePcm := NewToneGenerator(440, 8000).NextFrame(160, 0.7)
	testPayload := PCM16ToALaw(sinePcm)
	testSeq := uint16(10)
	testPkt := RTPPacketizeAudio(testPayload, 8, ParseSSRC("0200006530"), &testSeq, 0)
	downlinkFrame := append([]byte{byte(len(testPkt) >> 8), byte(len(testPkt))}, testPkt...)

	if _, err := platformConn.Write(downlinkFrame); err != nil {
		t.Fatalf("failed to send downlink TCP frame: %v", err)
	}

	select {
	case receivedPCM := <-pcmCh:
		if len(receivedPCM) != 320 {
			t.Errorf("expected 320 bytes PCM16, got %d", len(receivedPCM))
		}
		time.Sleep(20 * time.Millisecond)
		if sess.GetRxVol() <= 0 {
			t.Errorf("expected non-zero rx volume, got %f", sess.GetRxVol())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for TCP downlink pcm")
	}
}

func TestParseSDPTCP(t *testing.T) {
	wvpSDP := `v=0
o=34020000002000000001 0 0 IN IP4 192.168.42.198
s=Talk
c=IN IP4 192.168.42.198
t=0 0
m=audio 50000 TCP/RTP/AVP 8
a=setup:passive
a=connection:new
a=sendrecv
a=rtpmap:8 PCMA/8000
y=0200006530
`
	info, err := ParseSDP(wvpSDP)
	if err != nil {
		t.Fatalf("ParseSDP failed: %v", err)
	}
	if !info.IsTCP {
		t.Errorf("expected IsTCP to be true for TCP/RTP/AVP audio")
	}
	if info.AudioPort != 50000 {
		t.Errorf("expected AudioPort 50000, got %d", info.AudioPort)
	}
	if info.AudioPayload != 8 {
		t.Errorf("expected AudioPayload 8, got %d", info.AudioPayload)
	}
	if info.TCPMode != "passive" {
		t.Errorf("expected TCPMode passive, got %s", info.TCPMode)
	}
	if info.Direction != "sendrecv" {
		t.Errorf("expected Direction sendrecv, got %s", info.Direction)
	}

	ans := BuildTalkAnswerSDP("34020000001180000001", "34020000001320000001", "192.168.42.198", 51085, "0200006530", info)
	if !strings.Contains(ans, "TCP/RTP/AVP") {
		t.Errorf("expected answer to contain TCP/RTP/AVP, got:\n%s", ans)
	}
	if !strings.Contains(ans, "a=setup:active") {
		t.Errorf("expected answer to contain a=setup:active, got:\n%s", ans)
	}
	if !strings.Contains(ans, "a=sendrecv") {
		t.Errorf("expected answer to contain a=sendrecv, got:\n%s", ans)
	}
}
