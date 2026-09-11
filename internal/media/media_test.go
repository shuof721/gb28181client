package media

import (
	"bytes"
	"testing"
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
