package device

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Subscriber 表示一个与平台建立的 SIP 订阅对话 (RFC 3265 / RFC 6665)
type Subscriber struct {
	ID           string    `json:"id"`           // 通常即 CallID
	Event        string    `json:"event"`        // 规范事件名: Catalog, presence, MobilePosition
	CallID       string    `json:"callId"`
	FromHeader   string    `json:"fromHeader"`   // 平台的 From (NOTIFY 时用作 To)
	ToHeader     string    `json:"toHeader"`     // 设备的 To (带本地 tag, NOTIFY 时用作 From)
	ContactURI   string    `json:"contactUri"`   // 平台的 Contact URI (NOTIFY 的 Request-URI)
	PlatformID   string    `json:"platformId"`   // 平台 SIP 国标编码
	SubscribedAt time.Time `json:"subscribedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	ExpiresSec   int       `json:"expiresSec"`
	Interval     int       `json:"interval"` // 针对位置上报的推荐周期 (秒)

	cseq int32
}

// NextCSeq 获取此订阅对话内的递增 CSeq
func (s *Subscriber) NextCSeq() int {
	return int(atomic.AddInt32(&s.cseq, 1))
}

// IsExpired 检查该订阅是否已自然过期
func (s *Subscriber) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// RemainingSec 返回剩余订阅有效期 (秒)
func (s *Subscriber) RemainingSec() int {
	rem := int(time.Until(s.ExpiresAt).Seconds())
	if rem < 0 {
		return 0
	}
	return rem
}

// SubscriptionManager 管理所有外部平台发起的 SIP 订阅关系
type SubscriptionManager struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber // key: CallID
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscribers: make(map[string]*Subscriber),
	}
}

// AddOrUpdate 记录或刷新订阅对话
func (sm *SubscriptionManager) AddOrUpdate(sub *Subscriber) {
	if sub == nil || sub.CallID == "" {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 清理过期项
	now := time.Now()
	for k, v := range sm.subscribers {
		if now.After(v.ExpiresAt) {
			delete(sm.subscribers, k)
		}
	}

	sm.subscribers[sub.CallID] = sub
}

// Remove 移除指定 Call-ID 的订阅
func (sm *SubscriptionManager) Remove(callID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.subscribers, callID)
}

// Get 获取指定 Call-ID 的订阅
func (sm *SubscriptionManager) Get(callID string) *Subscriber {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.subscribers[callID]
}

// ListByEvent 列出指定事件 (Catalog / presence / MobilePosition) 的有效活跃订阅者
func (sm *SubscriptionManager) ListByEvent(eventName string) []*Subscriber {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var list []*Subscriber
	for k, sub := range sm.subscribers {
		if now.After(sub.ExpiresAt) {
			delete(sm.subscribers, k)
			continue
		}
		if eventMatches(sub.Event, eventName) {
			list = append(list, sub)
		}
	}
	return list
}

// ListAll 列出全部有效活跃订阅者
func (sm *SubscriptionManager) ListAll() []*Subscriber {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var list []*Subscriber
	for k, sub := range sm.subscribers {
		if now.After(sub.ExpiresAt) {
			delete(sm.subscribers, k)
			continue
		}
		list = append(list, sub)
	}
	return list
}

func eventMatches(subEv, targetEv string) bool {
	s1 := strings.ToLower(strings.TrimSpace(subEv))
	s2 := strings.ToLower(strings.TrimSpace(targetEv))
	if s1 == s2 {
		return true
	}
	// 针对移动位置的兼容性匹配：presence 与 mobileposition 等效
	isLoc1 := s1 == "presence" || s1 == "mobileposition"
	isLoc2 := s2 == "presence" || s2 == "mobileposition"
	return isLoc1 && isLoc2
}
