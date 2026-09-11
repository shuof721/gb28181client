// 简易 GB28181 平台模拟器：用于本地联调设备端，不必起 WVP。
// 功能：接收 REGISTER（Digest 可选）、MESSAGE；主动下发 Catalog/DeviceInfo；
// 可选对指定通道发送 INVITE 点播并接收 SDP 应答。
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/local/gb28181-device/internal/sip"
)

type mock struct {
	ua         *sip.UA
	devID      string
	pass       string
	auth       bool
	realm      string
	sn         int32
	inviteCh   string
	inviteIP   string
	invitePort int
	// 注册后设备联系地址
	devContact string
}

func main() {
	listenIP := flag.String("listen-ip", "127.0.0.1", "监听 IP")
	listenPort := flag.Int("listen-port", 5060, "监听端口")
	devID := flag.String("device-id", "34020000001180000001", "设备国标编号")
	pass := flag.String("password", "12345678", "设备密码")
	auth := flag.Bool("auth", false, "是否要求 Digest 鉴权")
	realm := flag.String("realm", "3402000000", "Digest realm")
	inviteCh := flag.String("invite-channel", "", "注册成功后向该通道发 INVITE 点播（可空）")
	inviteIP := flag.String("invite-ip", "127.0.0.1", "INVITE 媒体接收 IP")
	invitePort := flag.Int("invite-port", 30000, "INVITE 媒体接收端口")
	flag.Parse()

	m := &mock{
		devID:      *devID,
		pass:       *pass,
		auth:       *auth,
		realm:      *realm,
		inviteCh:   *inviteCh,
		inviteIP:   *inviteIP,
		invitePort: *invitePort,
	}
	m.ua = sip.NewUA(*listenIP, *listenPort, *listenIP, *listenPort, "udp", "platform", "")
	if err := m.ua.Start(); err != nil {
		log.Fatal(err)
	}
	m.ua.Handle("REGISTER", m.onRegister)
	m.ua.Handle("MESSAGE", m.onMessage)
	m.ua.Handle("INVITE", m.onInvite)
	m.ua.Handle("BYE", m.onBye)
	m.ua.Handle("ACK", func(msg *sip.Message, src net.Addr) {})
	m.ua.Handle("OPTIONS", func(msg *sip.Message, src net.Addr) {
		_ = m.ua.Reply(msg, src, 200, "OK", nil, "")
	})

	log.Printf("[platform] listening udp %s:%d", *listenIP, *listenPort)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	m.ua.Close()
}

func (m *mock) nextSN() string {
	return strconv.Itoa(int(atomic.AddInt32(&m.sn, 1)))
}

func (m *mock) onRegister(msg *sip.Message, src net.Addr) {
	authz := msg.GetHeader("Authorization")
	if m.auth && authz == "" {
		resp := sip.NewResponse(401, "Unauthorized")
		copyBasic(resp, msg)
		resp.SetHeader("WWW-Authenticate",
			fmt.Sprintf(`Digest realm="%s", nonce="%s", algorithm=MD5`, m.realm, sip.RandomToken(8)))
		resp.SetHeader("Content-Length", "0")
		_ = m.ua.SendRaw(src.String(), resp)
		log.Printf("[platform] REGISTER challenge from %s", msg.FromUser())
		return
	}

	exp := msg.GetHeader("Expires")
	if exp == "" {
		exp = "3600"
	}
	resp := sip.NewResponse(200, "OK")
	copyBasic(resp, msg)
	if c := msg.GetHeader("Contact"); c != "" {
		resp.SetHeader("Contact", c)
	}
	resp.SetHeader("Expires", exp)
	resp.SetHeader("Date", time.Now().Format(time.RFC1123))
	resp.SetHeader("Content-Length", "0")
	_ = m.ua.SendRaw(src.String(), resp)

	if exp == "0" {
		log.Printf("[platform] UNREGISTER %s", msg.FromUser())
		return
	}

	m.devContact = src.String()
	log.Printf("[platform] REGISTER OK from=%s src=%s expires=%s", msg.FromUser(), src, exp)

	go func() {
		time.Sleep(200 * time.Millisecond)
		dst := m.devContact
		m.sendCatalogQuery(dst, m.devID)
		m.sendDeviceInfoQuery(dst, m.devID)
		if m.inviteCh != "" {
			time.Sleep(500 * time.Millisecond)
			m.sendInvite(dst, m.inviteCh)
		}
	}()
}

func (m *mock) onMessage(msg *sip.Message, src net.Addr) {
	_ = m.ua.Reply(msg, src, 200, "OK", nil, "")
	body := string(msg.Body)
	if body == "" {
		return
	}
	log.Printf("[platform] MESSAGE from=%s ctype=%s body=\n%s",
		msg.FromUser(), msg.GetHeader("Content-Type"), body)
}

func (m *mock) onInvite(msg *sip.Message, src net.Addr) {
	// 正常流程中平台是 UAC 发 INVITE；此处若收到说明角色搞反了
	log.Printf("[platform] unexpected INVITE callID=%s", msg.CallID())
	_ = m.ua.Reply(msg, src, 405, "Method Not Allowed", nil, "")
}

func (m *mock) onBye(msg *sip.Message, src net.Addr) {
	_ = m.ua.Reply(msg, src, 200, "OK", nil, "")
}

func (m *mock) sendCatalogQuery(dst, deviceID string) {
	body := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Query>
  <CmdType>Catalog</CmdType>
  <SN>%s</SN>
  <DeviceID>%s</DeviceID>
</Query>`, m.nextSN(), deviceID)
	m.sendMessage(dst, body, deviceID)
}

func (m *mock) sendDeviceInfoQuery(dst, deviceID string) {
	body := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Query>
  <CmdType>DeviceInfo</CmdType>
  <SN>%s</SN>
  <DeviceID>%s</DeviceID>
</Query>`, m.nextSN(), deviceID)
	m.sendMessage(dst, body, deviceID)
}

func (m *mock) sendMessage(dst, body, deviceID string) {
	req := sip.NewRequest("MESSAGE", fmt.Sprintf("sip:%s@%s", deviceID, dst))
	branch := "z9hG4bK" + sip.RandomToken(8)
	req.SetHeader("Via", fmt.Sprintf("SIP/2.0/UDP %s:%d;rport;branch=%s",
		m.ua.LocalIP, m.ua.LocalPort, branch))
	req.SetHeader("From", fmt.Sprintf("<sip:platform@%s>;tag=%s", m.ua.LocalIP, sip.RandomToken(4)))
	req.SetHeader("To", fmt.Sprintf("<sip:%s@%s>", deviceID, m.ua.LocalIP))
	req.SetHeader("Call-ID", sip.RandomToken(8)+"@platform")
	req.SetHeader("CSeq", fmt.Sprintf("%d MESSAGE", m.ua.NextCSeq()))
	req.SetHeader("Max-Forwards", "70")
	req.SetHeader("User-Agent", "GB28181-MockPlatform/1.0")
	req.SetHeader("Content-Type", "Application/MANSCDP+xml")
	req.Body = []byte(body)
	req.SetHeader("Content-Length", strconv.Itoa(len(req.Body)))
	_ = m.ua.SendRaw(dst, req)
	log.Printf("[platform] sent query -> %s", dst)
}

// sendInvite 向设备通道发起点播 INVITE。
func (m *mock) sendInvite(dst, channelID string) {
	ssrc := fmt.Sprintf("01%08d", time.Now().UnixNano()%1e8)
	sdp := fmt.Sprintf("v=0\r\n"+
		"o=%s 0 0 IN IP4 %s\r\n"+
		"s=Play\r\n"+
		"c=IN IP4 %s\r\n"+
		"t=0 0\r\n"+
		"m=video %d RTP/AVP 96 97 98\r\n"+
		"a=recvonly\r\n"+
		"a=rtpmap:96 PS/90000\r\n"+
		"a=rtpmap:97 MPEG4/90000\r\n"+
		"a=rtpmap:98 H264/90000\r\n"+
		"y=%s\r\n",
		m.devID, m.inviteIP, m.inviteIP, m.invitePort, ssrc)

	callID := sip.RandomToken(8) + "@platform"
	fromTag := sip.RandomToken(6)
	req := sip.NewRequest("INVITE", fmt.Sprintf("sip:%s@%s", channelID, dst))
	branch := "z9hG4bK" + sip.RandomToken(8)
	req.SetHeader("Via", fmt.Sprintf("SIP/2.0/UDP %s:%d;rport;branch=%s",
		m.ua.LocalIP, m.ua.LocalPort, branch))
	req.SetHeader("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", m.devID, m.ua.LocalIP, fromTag))
	req.SetHeader("To", fmt.Sprintf("<sip:%s@%s>", channelID, m.ua.LocalIP))
	req.SetHeader("Call-ID", callID)
	req.SetHeader("CSeq", fmt.Sprintf("%d INVITE", m.ua.NextCSeq()))
	req.SetHeader("Max-Forwards", "70")
	req.SetHeader("Contact", fmt.Sprintf("<sip:platform@%s:%d>", m.ua.LocalIP, m.ua.LocalPort))
	req.SetHeader("Subject", fmt.Sprintf("%s,%s,%s", channelID, m.devID, ssrc))
	req.SetHeader("Content-Type", "Application/SDP")
	req.Body = []byte(sdp)
	req.SetHeader("Content-Length", strconv.Itoa(len(req.Body)))
	_ = m.ua.SendRaw(dst, req)
	log.Printf("[platform] sent INVITE ch=%s ssrc=%s media=%s:%d -> %s",
		channelID, ssrc, m.inviteIP, m.invitePort, dst)
	log.Printf("[platform] waiting device 200 OK with SDP...")
}

func copyBasic(resp, req *sip.Message) {
	if v := req.GetHeader("Via"); v != "" {
		resp.SetHeader("Via", v)
	}
	from := req.GetHeader("From")
	to := req.GetHeader("To")
	if !strings.Contains(strings.ToLower(to), "tag=") {
		to = to + ";tag=" + sip.RandomToken(6)
	}
	resp.SetHeader("From", from)
	resp.SetHeader("To", to)
	resp.SetHeader("Call-ID", req.CallID())
	resp.SetHeader("CSeq", req.GetHeader("CSeq"))
}
