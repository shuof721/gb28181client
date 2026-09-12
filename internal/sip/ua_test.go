package sip

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestUAMessageCallIDAndAddressing(t *testing.T) {
	// 模拟服务端 UA
	srv := NewUA("127.0.0.1", 59310, "127.0.0.1", 59311, "udp", "34020000002000000001", "")
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	recCalls := make(chan *Message, 5)
	srv.Handle("MESSAGE", func(m *Message, src net.Addr) {
		_ = srv.Reply(m, src, 200, "OK", nil, "")
		recCalls <- m
	})

	// 客户端设备 UA
	cli := NewUA("127.0.0.1", 59311, "127.0.0.1", 59310, "udp", "34020000001180000001", "")
	cli.SetServerID("34020000002000000001")
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	// 发送两条不同的 MESSAGE
	_, err := cli.SendMessage([]byte("msg1"), "text/plain")
	if err != nil {
		t.Fatalf("msg1 failed: %v", err)
	}

	_, err = cli.SendMessage([]byte("msg2"), "text/plain")
	if err != nil {
		t.Fatalf("msg2 failed: %v", err)
	}

	select {
	case m1 := <-recCalls:
		select {
		case m2 := <-recCalls:
			if m1.CallID() == m2.CallID() {
				t.Fatalf("Call-ID must be unique between independent MESSAGE requests: got %s vs %s", m1.CallID(), m2.CallID())
			}
			if !strings.Contains(m1.GetHeader("To"), "34020000002000000001") {
				t.Fatalf("To header should address ServerID: got %s", m1.GetHeader("To"))
			}
			if !strings.Contains(m1.RequestURI, "34020000002000000001") {
				t.Fatalf("RequestURI should address ServerID: got %s", m1.RequestURI)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for m2")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for m1")
	}
}
