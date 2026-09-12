package media

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func TestPackVideoPES(t *testing.T) {
	es := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0x88, 0x84, 0x00}
	ps := PackVideoPES(es, 90000)
	if !bytes.HasPrefix(ps, []byte{0x00, 0x00, 0x01, 0xBA}) {
		t.Fatalf("no pack header: %x", ps[:8])
	}
	if !bytes.Contains(ps, []byte{0x00, 0x00, 0x01, 0xE0}) {
		t.Fatalf("no PES video start")
	}
	if !bytes.Contains(ps, es) {
		t.Fatalf("ES not embedded")
	}
}

func TestRTPPacketize(t *testing.T) {
	ps := bytes.Repeat([]byte{0xAB}, 3000)
	seq := uint16(0)
	pkts := RTPPacketizePS(ps, 0x01000001, &seq, 100, 96, 1400)
	if len(pkts) != 3 {
		t.Fatalf("pkts=%d", len(pkts))
	}
	// 最后一包 marker=1
	if pkts[len(pkts)-1][1]&0x80 == 0 {
		t.Fatal("last packet missing marker")
	}
	if seq != 3 {
		t.Fatalf("seq=%d", seq)
	}
	// 还原负载
	var out []byte
	for _, p := range pkts {
		out = append(out, p[12:]...)
	}
	if !bytes.Equal(out, ps) {
		t.Fatal("payload mismatch")
	}
}

func TestParseSDP(t *testing.T) {
	s := "v=0\r\n" +
		"o=34020000002000000001 0 0 IN IP4 192.168.1.10\r\n" +
		"s=Play\r\n" +
		"c=IN IP4 192.168.1.10\r\n" +
		"t=0 0\r\n" +
		"m=video 30000 RTP/AVP 96 97 98\r\n" +
		"a=recvonly\r\n" +
		"a=rtpmap:96 PS/90000\r\n" +
		"y=0100000054123456789\r\n"
	info, err := ParseSDP(s)
	if err != nil {
		t.Fatal(err)
	}
	if info.IP != "192.168.1.10" || info.VideoPort != 30000 || info.SSRC != "0100000054123456789" {
		t.Fatalf("%+v", info)
	}
}

func TestSyntheticSource(t *testing.T) {
	src, err := NewSyntheticSource(320, 240, 50)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := src.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(frame, []byte{0x00, 0x00, 0x00, 0x01}) {
		t.Fatalf("no annexb start: %x", frame[:8])
	}
}

func TestPlaybackSDP(t *testing.T) {
	sdp := "v=0\r\n" +
		"o=34020000002000000001 0 0 IN IP4 192.168.1.10\r\n" +
		"s=Playback\r\n" +
		"u=34020000001320000001:3\r\n" +
		"c=IN IP4 192.168.1.10\r\n" +
		"t=1726041600 1726048800\r\n" +
		"m=video 30002 RTP/AVP 96\r\n" +
		"a=recvonly\r\n" +
		"a=rtpmap:96 PS/90000\r\n" +
		"y=0100000054\r\n"

	info, err := ParseSDP(sdp)
	if err != nil {
		t.Fatal(err)
	}
	if info.SessionName != "Playback" {
		t.Fatalf("expected Playback, got %s", info.SessionName)
	}
	if info.StartTime != "1726041600" || info.EndTime != "1726048800" {
		t.Fatalf("unexpected t line: %s %s", info.StartTime, info.EndTime)
	}

	ans := BuildAnswerSDP("34020000002000000001", "34020000001320000001", "192.168.1.50", "0100000054", info)
	if !bytes.Contains([]byte(ans), []byte("s=Playback")) {
		t.Fatal("answer SDP missing s=Playback")
	}
	if !bytes.Contains([]byte(ans), []byte("u=34020000001320000001:3")) {
		t.Fatal("answer SDP missing u=...:3")
	}
	if !bytes.Contains([]byte(ans), []byte("t=1726041600 1726048800")) {
		t.Fatal("answer SDP missing t=...")
	}
}

func TestSessionControl(t *testing.T) {
	sess := &Session{
		scale: 1.0,
	}
	sess.SetScale(2.0)
	if sess.GetScale() != 2.0 {
		t.Fatalf("expected scale 2.0, got %.2f", sess.GetScale())
	}

	sess.Pause()
	if !sess.IsPaused() {
		t.Fatal("expected paused")
	}
	sess.Resume()
	if sess.IsPaused() {
		t.Fatal("expected not paused")
	}

	sess.Seek(15.5)
	if sess.CurrentOffset() != 15.5 {
		t.Fatalf("expected offset 15.5, got %.2f", sess.CurrentOffset())
	}
}

func TestSyntheticSourceSeek(t *testing.T) {
	src, err := NewSyntheticSource(320, 240, 25)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := src.Seek(10.0, 25)
	if err != nil {
		t.Fatal(err)
	}
	if actual != 10.0 {
		t.Fatalf("expected 10.0, got %.2f", actual)
	}
	if src.frameIdx != 250 {
		t.Fatalf("expected frameIdx 250, got %d", src.frameIdx)
	}
}

func TestFileSourceSeek(t *testing.T) {
	sps := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1E}
	pps := []byte{0x00, 0x00, 0x00, 0x01, 0x68, 0xCE, 0x38, 0x80}
	idrAU := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0x88, 0x84, 0x00}
	pAU := []byte{0x00, 0x00, 0x00, 0x01, 0x61, 0x9A, 0x00, 0x00}

	fs := &FileSource{
		sps:        sps,
		pps:        pps,
		aus:        [][]byte{idrAU, pAU, pAU, idrAU, pAU},
		idrIndices: []int{0, 3},
	}

	// 5 帧，25 fps，时长 0.2 秒
	// seek 到 0.12 秒（对应第 3 帧，即第 2 个 IDR 帧）
	act, err := fs.Seek(0.12, 25)
	if err != nil {
		t.Fatal(err)
	}
	if fs.pos != 3 {
		t.Fatalf("expected pos 3, got %d", fs.pos)
	}
	if !fs.forceI {
		t.Fatal("expected forceI to be true after seek")
	}
	_ = act

	// Next() 取帧，应该强制包含 SPS 和 PPS
	nextFrame, err := fs.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(nextFrame, sps) || !bytes.Contains(nextFrame, pps) {
		t.Fatal("next frame after seek must contain SPS and PPS")
	}
	if fs.forceI {
		t.Fatal("forceI should be reset to false after Next()")
	}
}

func TestSessionSeekRTPStream(t *testing.T) {
	// 启动本地 UDP 接收端模拟流媒体服务（ZLMediaKit）
	udpRecv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer udpRecv.Close()
	localAddr := udpRecv.LocalAddr().(*net.UDPAddr)

	mgr := NewSessionManager("127.0.0.1", 25, 1400, func(channelID string) (H264Source, error) {
		return NewSyntheticSource(320, 240, 25)
	})

	sdp := &SDPInfo{
		IP:        "127.0.0.1",
		VideoPort: localAddr.Port,
		SSRC:      "0100000001",
	}

	callID := "test-seek-callid"
	_, err = mgr.StartPlayback("34020000001320000001", callID, sdp, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.StopByCallID(callID)

	sess := mgr.GetSession(callID)
	if sess == nil {
		t.Fatal("session not found")
	}

	buf := make([]byte, 2048)
	// 读取第一个包，验证初始时间戳在较小区间
	_ = udpRecv.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := udpRecv.ReadFrom(buf)
	if err != nil {
		t.Fatalf("failed to read initial RTP packet: %v", err)
	}
	initTS := binary.BigEndian.Uint32(buf[4:8])
	if initTS > 90000*5 {
		t.Fatalf("initial timestamp unexpectedly large: %d", initTS)
	}

	// 触发 Seek 到 10 秒
	actual := sess.Seek(10.0)
	if actual != 10.0 {
		t.Fatalf("expected seek actual 10.0, got %.2f", actual)
	}

	// 读取后续收到的 RTP 包，验证时间戳跳转到 10.0 * 90000 (约 900000) 附近
	gotSeekTS := false
	for i := 0; i < 150; i++ {
		_ = udpRecv.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err = udpRecv.ReadFrom(buf)
		if err != nil {
			t.Logf("read error at i=%d: %v", i, err)
			break
		}
		if n < 12 {
			continue
		}
		ts := binary.BigEndian.Uint32(buf[4:8])
		// 900000 附近（允许后续发出的帧有小幅 3600 递增）
		if ts >= 900000 && ts < 900000+90000*2 {
			gotSeekTS = true
			break
		}
	}

	if !gotSeekTS {
		t.Fatalf("expected RTP timestamp to jump to >= 900000 after seek to 10s, last ts=%d", binary.BigEndian.Uint32(buf[4:8]))
	}
}

