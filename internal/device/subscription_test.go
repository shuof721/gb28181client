package device

import (
	"testing"
	"time"
)

func TestSubscriptionManagerLifecycle(t *testing.T) {
	sm := NewSubscriptionManager()

	// 1. 添加 Catalog 订阅
	sub1 := &Subscriber{
		ID:           "sub-1",
		Event:        "Catalog",
		CallID:       "call-cat-1",
		FromHeader:   "<sip:34020000002000000001@127.0.0.1>;tag=plat123",
		ToHeader:     "<sip:34020000001180000001@127.0.0.1>;tag=dev123",
		ContactURI:   "sip:34020000002000000001@127.0.0.1:5060",
		PlatformID:   "34020000002000000001",
		SubscribedAt: time.Now(),
		ExpiresAt:    time.Now().Add(60 * time.Second),
		ExpiresSec:   60,
	}
	sm.AddOrUpdate(sub1)

	// 2. 添加 presence / MobilePosition 订阅
	sub2 := &Subscriber{
		ID:           "sub-2",
		Event:        "presence",
		CallID:       "call-gps-2",
		FromHeader:   "<sip:34020000002000000001@127.0.0.1>;tag=plat456",
		ToHeader:     "<sip:34020000001180000001@127.0.0.1>;tag=dev456",
		ContactURI:   "sip:34020000002000000001@127.0.0.1:5060",
		PlatformID:   "34020000002000000001",
		SubscribedAt: time.Now(),
		ExpiresAt:    time.Now().Add(60 * time.Second),
		ExpiresSec:   60,
		Interval:     5,
	}
	sm.AddOrUpdate(sub2)

	// 3. 按事件检索
	catSubs := sm.ListByEvent("Catalog")
	if len(catSubs) != 1 || catSubs[0].CallID != "call-cat-1" {
		t.Fatalf("expected 1 Catalog subscriber, got %d", len(catSubs))
	}

	gpsSubs := sm.ListByEvent("MobilePosition")
	if len(gpsSubs) != 1 || gpsSubs[0].CallID != "call-gps-2" {
		t.Fatalf("expected 1 MobilePosition/presence subscriber, got %d", len(gpsSubs))
	}

	all := sm.ListAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 total subscribers, got %d", len(all))
	}

	// 4. 剩余时间与 CSeq 递增
	if sub1.RemainingSec() <= 0 {
		t.Fatal("expected positive remaining seconds")
	}
	seq1 := sub1.NextCSeq()
	seq2 := sub1.NextCSeq()
	if seq2 != seq1+1 {
		t.Fatalf("expected seq increment, got %d and %d", seq1, seq2)
	}

	// 5. 移除订阅
	sm.Remove("call-cat-1")
	catSubsAfter := sm.ListByEvent("Catalog")
	if len(catSubsAfter) != 0 {
		t.Fatalf("expected 0 Catalog subscribers after remove, got %d", len(catSubsAfter))
	}
}
