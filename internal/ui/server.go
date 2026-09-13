package ui

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/device"
	"github.com/local/gb28181-device/internal/gb28181"
	"github.com/local/gb28181-device/internal/media"
)

const maxUploadBytes = 512 << 20 // 512MB

type Server struct {
	mgr *device.Manager
	mux *http.ServeMux
}

func New(mgr *device.Manager) *Server {
	s := &Server{mgr: mgr, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(device.AssetsDir))))

	// 媒体公共库
	s.mux.HandleFunc("/api/videos", s.handleVideos)
	s.mux.HandleFunc("/api/videos/upload", s.handleUpload)
	s.mux.HandleFunc("/api/videos/delete", s.handleDeleteVideo)

	// 多设备管理总调度
	s.mux.HandleFunc("/api/devices", s.handleDeviceDispatch)
	s.mux.HandleFunc("/api/devices/", s.handleDeviceDispatch)

	// 向下兼容单设备状态回退
	s.mux.HandleFunc("/api/status", s.handleLegacyStatus)
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("[ui] console at http://%s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// handleDeviceDispatch 统一分发 /api/devices/* 路由
func (s *Server) handleDeviceDispatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/devices")
	path = strings.Trim(path, "/")
	if path == "" {
		if r.Method == http.MethodGet {
			s.handleListDevices(w, r)
			return
		}
		if r.Method == http.MethodPost {
			s.handleCreateDevice(w, r)
			return
		}
		writeErr(w, 405, "Method not allowed")
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		switch parts[0] {
		case "next-port":
			s.handleNextPort(w, r)
			return
		case "start-all":
			s.handleStartAll(w, r)
			return
		case "stop-all":
			s.handleStopAll(w, r)
			return
		default:
			// /api/devices/{id}
			id := parts[0]
			switch r.Method {
			case http.MethodGet:
				s.handleGetDevice(w, r, id)
			case http.MethodPut, http.MethodPost:
				s.handleUpdateDevice(w, r, id)
			case http.MethodDelete:
				s.handleDeleteDevice(w, r, id)
			default:
				writeErr(w, 405, "Method not allowed")
			}
			return
		}
	}

	id := parts[0]
	action := parts[1]
	switch action {
	case "start":
		s.handleDeviceStart(w, r, id)
	case "stop":
		s.handleDeviceStop(w, r, id)
	case "restart":
		s.handleDeviceRestart(w, r, id)
	case "register":
		s.handleDeviceRegister(w, r, id)
	case "unregister":
		s.handleDeviceUnregister(w, r, id)
	case "keepalive":
		s.handleDeviceKeepalive(w, r, id)
	case "alarm":
		if len(parts) >= 3 && parts[2] == "auto" {
			s.handleDeviceAlarmAuto(w, r, id)
		} else {
			s.handleDeviceAlarm(w, r, id)
		}
	case "alarms":
		if len(parts) >= 3 && parts[2] == "clear" {
			s.handleDeviceAlarmsClear(w, r, id)
		} else {
			s.handleDeviceAlarms(w, r, id)
		}
	case "guard":
		s.handleDeviceGuard(w, r, id)
	case "session", "sessions":
		if len(parts) >= 3 && parts[2] == "stop" {
			s.handleDeviceStopSession(w, r, id)
		} else {
			s.handleDeviceSessions(w, r, id)
		}
	case "talk":
		if len(parts) >= 3 {
			switch parts[2] {
			case "ws":
				s.handleDeviceTalkWS(w, r, id)
			case "mode":
				s.handleDeviceTalkMode(w, r, id)
			case "stop":
				s.handleDeviceTalkStop(w, r, id)
			default:
				writeErr(w, 404, "Not found")
			}
		} else {
			s.handleDeviceTalkSessions(w, r, id)
		}
	case "logs":
		s.handleDeviceLogs(w, r, id)
	case "channels":
		if len(parts) >= 3 {
			switch parts[2] {
			case "add":
				s.handleChannelAdd(w, r, id)
			case "update":
				s.handleChannelUpdate(w, r, id)
			case "remove":
				s.handleChannelRemove(w, r, id)
			case "bind":
				s.handleChannelBind(w, r, id)
			case "audio":
				s.handleChannelAudio(w, r, id)
			case "status":
				s.handleChannelStatus(w, r, id)
			default:
				writeErr(w, 404, "Not found")
			}
		} else {
			writeErr(w, 404, "Not found")
		}
	case "catalog":
		if len(parts) >= 3 && parts[2] == "notify" {
			s.handleDeviceCatalogNotify(w, r, id)
		} else {
			writeErr(w, 404, "Not found")
		}
	case "gps":
		if len(parts) >= 3 {
			switch parts[2] {
			case "config":
				s.handleDeviceGPSConfig(w, r, id)
			case "report":
				s.handleDeviceGPSReport(w, r, id)
			case "sync":
				s.handleDeviceGPSSync(w, r, id)
			default:
				writeErr(w, 404, "Not found")
			}
		} else {
			s.handleDeviceGPSStatus(w, r, id)
		}
	case "subscriptions":
		s.handleDeviceSubscriptions(w, r, id)
	case "media":
		if len(parts) >= 3 && parts[2] == "mode" {
			s.handleMediaMode(w, r, id)
		} else {
			writeErr(w, 404, "Not found")
		}
	case "records":
		s.handleDeviceRecords(w, r, id)
	case "ptz":
		if len(parts) >= 3 {
			switch parts[2] {
			case "control":
				s.handleDevicePTZControl(w, r, id)
			case "preset":
				if len(parts) >= 4 && parts[3] == "call" {
					s.handleDevicePTZCallPreset(w, r, id)
				} else {
					s.handleDevicePTZPreset(w, r, id)
				}
			default:
				writeErr(w, 404, "Not found")
			}
		} else {
			s.handleDevicePTZStatus(w, r, id)
		}
	case "control":
		if len(parts) >= 3 {
			switch parts[2] {
			case "events":
				s.handleDeviceControlEvents(w, r, id)
			case "iframe":
				s.handleDeviceControlIFrame(w, r, id)
			case "reboot":
				s.handleDeviceControlReboot(w, r, id)
			case "record":
				s.handleDeviceControlRecord(w, r, id)
			case "time":
				s.handleDeviceControlTime(w, r, id)
			case "home-position":
				s.handleDeviceControlHomePosition(w, r, id)
			default:
				writeErr(w, 404, "Not found")
			}
		} else {
			writeErr(w, 404, "Not found")
		}
	default:
		writeErr(w, 404, "Unknown device action")
	}
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	summaries := s.mgr.ListSummaries()
	writeJSON(w, 200, map[string]any{"devices": summaries})
}

func (s *Server) handleNextPort(w http.ResponseWriter, r *http.Request) {
	port := s.mgr.NextAvailablePort()
	writeJSON(w, 200, map[string]int{"port": port})
}

func (s *Server) handleCreateDevice(w http.ResponseWriter, r *http.Request) {
	var profile config.DeviceProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		writeErr(w, 400, "Invalid JSON: "+err.Error())
		return
	}
	if profile.SIP.LocalPort <= 0 {
		profile.SIP.LocalPort = s.mgr.NextAvailablePort()
	}
	if err := s.mgr.AddDevice(&profile); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "id": profile.Device.ID, "port": profile.SIP.LocalPort})
}

func (s *Server) handleGetDevice(w http.ResponseWriter, r *http.Request, id string) {
	md, err := s.mgr.Get(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	resp := map[string]any{
		"profile": md.Profile,
		"running": md.Running,
		"error":   md.Err,
	}
	if md.Running && md.Dev != nil {
		resp["status"] = md.Dev.Status()
	}
	writeJSON(w, 200, resp)
}

func (s *Server) handleUpdateDevice(w http.ResponseWriter, r *http.Request, id string) {
	var profile config.DeviceProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		writeErr(w, 400, "Invalid JSON: "+err.Error())
		return
	}
	restart := r.URL.Query().Get("restart") == "true" || r.URL.Query().Get("restart") == "1"
	if err := s.mgr.UpdateProfile(id, &profile, restart); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "id": id})
}

func (s *Server) handleDeleteDevice(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.mgr.DeleteDevice(id); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "deleted", "id": id})
}

func (s *Server) handleDeviceStart(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	if err := s.mgr.Start(id); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "started", "id": id})
}

func (s *Server) handleDeviceStop(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	if err := s.mgr.Stop(id); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "stopped", "id": id})
}

func (s *Server) handleDeviceRestart(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	if err := s.mgr.Restart(id); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "restarted", "id": id})
}

func (s *Server) handleStartAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	errs := s.mgr.StartAll()
	if len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		writeJSON(w, 207, map[string]any{"ok": false, "errors": msgs})
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "all started"})
}

func (s *Server) handleStopAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	s.mgr.StopAll()
	writeJSON(w, 200, map[string]string{"ok": "all stopped"})
}

func (s *Server) handleDeviceRegister(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := dev.RegisterNow(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "registered", "id": id})
}

func (s *Server) handleDeviceUnregister(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := dev.UnregisterNow(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "unregistered", "id": id})
}

func (s *Server) handleDeviceKeepalive(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := dev.KeepaliveNow(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "keepalive sent", "id": id})
}

func (s *Server) handleDeviceAlarm(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}

	var req struct {
		ChannelID   string  `json:"channelId"`
		AlarmMethod string  `json:"alarmMethod"`
		AlarmType   string  `json:"alarmType"`
		Priority    string  `json:"priority"`
		Description string  `json:"description"`
		Longitude   float64 `json:"longitude"`
		Latitude    float64 `json:"latitude"`
		Force       bool    `json:"force"`
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.ChannelID == "" {
		req.ChannelID = r.URL.Query().Get("channel")
	}
	if req.ChannelID == "" {
		st := dev.Status()
		if len(st.Channels) > 0 {
			req.ChannelID = st.Channels[0].ID
		} else {
			req.ChannelID = id
		}
	}
	if req.AlarmMethod == "" {
		req.AlarmMethod = r.URL.Query().Get("method")
		if req.AlarmMethod == "" {
			req.AlarmMethod = "5" // 默认：视频报警
		}
	}
	if req.AlarmType == "" {
		req.AlarmType = r.URL.Query().Get("type")
		if req.AlarmType == "" {
			req.AlarmType = "2" // 默认：移动侦测
		}
	}
	if req.Priority == "" {
		req.Priority = r.URL.Query().Get("priority")
		if req.Priority == "" {
			req.Priority = "3"
		}
	}
	if req.Description == "" {
		req.Description = r.URL.Query().Get("desc")
	}
	if !req.Force && (r.URL.Query().Get("force") == "true" || r.URL.Query().Get("force") == "1") {
		req.Force = true
	}

	rec, err := dev.SendAlarmEvent(req.ChannelID, req.AlarmMethod, req.AlarmType, req.Priority, req.Description, req.Longitude, req.Latitude, req.Force)
	if err != nil {
		if rec != nil && rec.Status == "suppressed" {
			writeJSON(w, 200, map[string]any{"ok": false, "suppressed": true, "error": err.Error(), "record": rec})
			return
		}
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "record": rec})
}

func (s *Server) handleDeviceAlarms(w http.ResponseWriter, r *http.Request, id string) {
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	am := dev.AlarmManager()
	if am == nil {
		writeJSON(w, 200, map[string]any{"records": []any{}, "guardStatus": "ResetGuard", "dutyStatus": "OFFDUTY", "autoAlarm": false})
		return
	}
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	writeJSON(w, 200, map[string]any{
		"guardStatus": am.GetGuard(""),
		"dutyStatus":  am.DutyStatus(id),
		"autoAlarm":   am.IsAutoAlarmRunning(),
		"records":     am.ListRecords(limit),
	})
}

func (s *Server) handleDeviceAlarmsClear(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	am := dev.AlarmManager()
	if am != nil {
		am.ClearRecords()
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleDeviceGuard(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	var req struct {
		ChannelID string `json:"channelId"`
		Guard     bool   `json:"guard"`
		Status    string `json:"status"`
		GuardCmd  string `json:"guardCmd"`
		AlarmCmd  string `json:"alarmCmd"`
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	cmd := req.GuardCmd
	if cmd == "" {
		cmd = req.Status
	}
	if cmd == "" {
		cmd = r.URL.Query().Get("guardCmd")
	}
	if cmd == "" {
		cmd = r.URL.Query().Get("status")
	}

	isGuard := req.Guard
	if cmd != "" {
		if strings.EqualFold(cmd, "SetGuard") {
			isGuard = true
		} else if strings.EqualFold(cmd, "ResetGuard") {
			isGuard = false
		}
	}
	if req.ChannelID == "" {
		req.ChannelID = r.URL.Query().Get("channel")
	}

	st := dev.Status()
	allChs := make([]string, 0, len(st.Channels))
	for _, ch := range st.Channels {
		allChs = append(allChs, ch.ID)
	}

	am := dev.AlarmManager()
	if am != nil {
		if strings.EqualFold(req.AlarmCmd, "ResetAlarm") || strings.EqualFold(cmd, "ResetAlarm") {
			if req.ChannelID == "" || req.ChannelID == st.DeviceID {
				am.ResetAlarm("", allChs)
				am.ResetAlarm(st.DeviceID, allChs)
			} else {
				am.ResetAlarm(req.ChannelID)
			}
		} else {
			if req.ChannelID == "" || req.ChannelID == st.DeviceID {
				am.SetGuard("", isGuard, allChs)
				am.SetGuard(st.DeviceID, isGuard, allChs)
			} else {
				am.SetGuard(req.ChannelID, isGuard)
			}
		}
	}
	writeJSON(w, 200, map[string]any{
		"ok":          true,
		"channelId":   req.ChannelID,
		"guardStatus": am.GetGuard(req.ChannelID),
		"dutyStatus":  am.DutyStatus(req.ChannelID),
	})
}

func (s *Server) handleDeviceAlarmAuto(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	var req struct {
		Enabled     bool     `json:"enabled"`
		Interval    int      `json:"interval"`    // 秒
		IntervalSec int      `json:"intervalSec"` // 秒
		Channels    []string `json:"channels"`
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	am := dev.AlarmManager()
	if am == nil {
		writeErr(w, 500, "alarm manager not available")
		return
	}
	if req.Enabled {
		sec := req.IntervalSec
		if sec <= 0 {
			sec = req.Interval
		}
		if sec <= 0 {
			sec = 15
		}
		interval := time.Duration(sec) * time.Second
		if len(req.Channels) == 0 {
			st := dev.Status()
			for _, ch := range st.Channels {
				req.Channels = append(req.Channels, ch.ID)
			}
		}
		am.StartAutoAlarm(interval, req.Channels, func(ch string) {
			_, _ = dev.SendAlarmEvent(ch, "5", "2", "3", "", 0, 0)
		})
	} else {
		am.StopAutoAlarm()
	}
	writeJSON(w, 200, map[string]any{
		"ok":      true,
		"running": am.IsAutoAlarmRunning(),
	})
}

func (s *Server) handleDeviceStopSession(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	callID := r.URL.Query().Get("callId")
	if callID == "" {
		writeErr(w, 400, "callId required")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	dev.StopSession(callID)
	writeJSON(w, 200, map[string]string{"ok": "stopped"})
}

func (s *Server) handleDeviceSessions(w http.ResponseWriter, r *http.Request, id string) {
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	mm := dev.MediaManager()
	if mm == nil {
		writeJSON(w, 200, map[string]any{"sessions": []any{}})
		return
	}
	writeJSON(w, 200, map[string]any{"sessions": mm.ListSessions()})
}

func (s *Server) handleDeviceTalkSessions(w http.ResponseWriter, r *http.Request, id string) {
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"sessions": dev.TalkManager().ListSessions()})
}

func (s *Server) handleDeviceTalkStop(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	callID := r.URL.Query().Get("callId")
	if callID != "" {
		dev.StopTalk(callID)
	} else {
		dev.StopAllTalk()
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleDeviceTalkMode(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}

	var req struct {
		CallID string `json:"callId"`
		Mode   string `json:"mode"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Mode == "" {
		req.Mode = r.URL.Query().Get("mode")
	}
	if req.CallID == "" {
		req.CallID = r.URL.Query().Get("callId")
	}

	tm := dev.TalkManager()
	var sess *media.TalkSession
	if req.CallID != "" {
		sess = tm.GetSession(req.CallID)
	} else {
		sess = tm.GetActiveSession()
	}
	if sess == nil {
		writeErr(w, 404, "no active talk session")
		return
	}
	sess.SetUplinkMode(req.Mode)
	writeJSON(w, 200, map[string]any{"ok": true, "mode": sess.UplinkMode()})
}

func (s *Server) handleDeviceTalkWS(w http.ResponseWriter, r *http.Request, id string) {
	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	tm := dev.TalkManager()
	callID := r.URL.Query().Get("callId")
	var sess *media.TalkSession
	if callID != "" {
		sess = tm.GetSession(callID)
	} else {
		sess = tm.GetActiveSession()
	}
	if sess == nil {
		writeErr(w, 404, "no active talk session")
		return
	}

	ws, err := UpgradeWS(w, r)
	if err != nil {
		log.Printf("[talk-ws] upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	pcmCh := make(chan []byte, 40)
	sess.RegisterListener(pcmCh)
	defer sess.UnregisterListener(pcmCh)

	// 协程：将接收自平台的 PCM 音频推给浏览器
	go func() {
		for pcm := range pcmCh {
			if err := ws.WriteBinary(pcm); err != nil {
				return
			}
		}
	}()

	// 主循环：读取浏览器麦克风发来的 PCM 音频
	for {
		payload, isBinary, err := ws.ReadFrame()
		if err != nil {
			return
		}
		if isBinary && len(payload) > 0 {
			numSamples := len(payload) / 2
			samples := make([]int16, numSamples)
			for i := 0; i < numSamples; i++ {
				samples[i] = int16(binary.LittleEndian.Uint16(payload[i*2 : i*2+2]))
			}
			sess.PushMicPCM(samples)
		}
	}
}

func (s *Server) handleDeviceLogs(w http.ResponseWriter, r *http.Request, id string) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if n <= 0 {
		n = 150
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	dev, err := s.mgr.GetDevice(id)
	if err != nil {
		// 设备未运行，返回空日志
		writeJSON(w, 200, map[string]any{"lines": []string{fmt.Sprintf("[%s] device is currently stopped", id)}})
		return
	}

	lines := dev.Logs(n * 3)
	if q != "" {
		filtered := lines[:0]
		for _, ln := range lines {
			if strings.Contains(strings.ToLower(ln), q) {
				filtered = append(filtered, ln)
			}
		}
		lines = filtered
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	writeJSON(w, 200, map[string]any{"lines": lines})
}

func (s *Server) handleDeviceRecords(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "GET only")
		return
	}
	md, err := s.mgr.Get(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}

	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		if len(md.Profile.Device.Channels) > 0 {
			channelID = md.Profile.Device.Channels[0].ID
		} else {
			channelID = md.Profile.Device.ID
		}
	}
	channelName := channelID
	for _, ch := range md.Profile.Device.Channels {
		if ch.ID == channelID {
			channelName = ch.Name
			break
		}
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	queryType := r.URL.Query().Get("type")
	if queryType == "" {
		queryType = "all"
	}

	records := device.GenerateRecordItems(
		md.Profile.Record,
		channelID,
		channelName,
		md.Profile.Device.ID,
		startStr,
		endStr,
		queryType,
	)

	writeJSON(w, 200, map[string]any{
		"records": records,
		"total":   len(records),
		"config":  md.Profile.Record,
	})
}

func (s *Server) handleDevicePTZStatus(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "GET only")
		return
	}
	md, err := s.mgr.Get(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	if !md.Running || md.Dev == nil {
		writeErr(w, 400, "设备未运行")
		return
	}

	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		if len(md.Profile.Device.Channels) > 0 {
			channelID = md.Profile.Device.Channels[0].ID
		} else {
			channelID = md.Profile.Device.ID
		}
	}

	ptz := md.Dev.GetChannelPTZ(channelID)
	writeJSON(w, 200, ptz.Status())
}

func (s *Server) handleDevicePTZControl(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	md, err := s.mgr.Get(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	if !md.Running || md.Dev == nil {
		writeErr(w, 400, "设备未运行")
		return
	}

	var req struct {
		ChannelID string   `json:"channelId"`
		Action    string   `json:"action"`
		PanSpeed  int      `json:"panSpeed"`
		TiltSpeed int      `json:"tiltSpeed"`
		ZoomSpeed int      `json:"zoomSpeed"`
		RawHex    string   `json:"rawHex"`
		Pan       *float64 `json:"pan"`
		Tilt      *float64 `json:"tilt"`
		Zoom      *float64 `json:"zoom"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "无效请求体: "+err.Error())
		return
	}

	channelID := req.ChannelID
	if channelID == "" {
		if len(md.Profile.Device.Channels) > 0 {
			channelID = md.Profile.Device.Channels[0].ID
		} else {
			channelID = md.Profile.Device.ID
		}
	}

	ptz := md.Dev.GetChannelPTZ(channelID)
	if req.RawHex != "" {
		ptzCmd, err := gb28181.ParsePTZCmd(req.RawHex)
		if err != nil {
			writeErr(w, 400, "无效 PTZ 十六进制指令: "+err.Error())
			return
		}
		st, err := ptz.ExecuteCommand(ptzCmd)
		if err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, st)
		return
	}

	if req.Action == "set_pose" && req.Pan != nil {
		tilt := 0.0
		zoom := 1.0
		if req.Tilt != nil {
			tilt = *req.Tilt
		}
		if req.Zoom != nil {
			zoom = *req.Zoom
		}
		st := ptz.SetPose(*req.Pan, tilt, zoom)
		writeJSON(w, 200, st)
		return
	}

	st := ptz.ManualControl(req.Action, req.PanSpeed, req.TiltSpeed, req.ZoomSpeed)
	writeJSON(w, 200, st)
}

func (s *Server) handleDevicePTZPreset(w http.ResponseWriter, r *http.Request, id string) {
	md, err := s.mgr.Get(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	if !md.Running || md.Dev == nil {
		writeErr(w, 400, "设备未运行")
		return
	}

	switch r.Method {
	case http.MethodGet:
		channelID := r.URL.Query().Get("channel")
		if channelID == "" && len(md.Profile.Device.Channels) > 0 {
			channelID = md.Profile.Device.Channels[0].ID
		}
		ptz := md.Dev.GetChannelPTZ(channelID)
		writeJSON(w, 200, ptz.Status())

	case http.MethodPost:
		var req struct {
			ChannelID  string  `json:"channelId"`
			PresetID   int     `json:"presetId"`
			Name       string  `json:"name"`
			Pan        float64 `json:"pan"`
			Tilt       float64 `json:"tilt"`
			Zoom       float64 `json:"zoom"`
			UseCurrent bool    `json:"useCurrent"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, 400, "无效请求体: "+err.Error())
			return
		}
		channelID := req.ChannelID
		if channelID == "" && len(md.Profile.Device.Channels) > 0 {
			channelID = md.Profile.Device.Channels[0].ID
		}
		ptz := md.Dev.GetChannelPTZ(channelID)
		st, err := ptz.SetPreset(req.PresetID, req.Name, req.Pan, req.Tilt, req.Zoom, req.UseCurrent)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, st)

	case http.MethodDelete:
		channelID := r.URL.Query().Get("channel")
		if channelID == "" && len(md.Profile.Device.Channels) > 0 {
			channelID = md.Profile.Device.Channels[0].ID
		}
		presetIDStr := r.URL.Query().Get("presetId")
		presetID, _ := strconv.Atoi(presetIDStr)
		if presetID <= 0 {
			var req struct {
				ChannelID string `json:"channelId"`
				PresetID  int    `json:"presetId"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.ChannelID != "" {
				channelID = req.ChannelID
			}
			presetID = req.PresetID
		}
		if presetID <= 0 {
			writeErr(w, 400, "缺少有效 presetId 参数")
			return
		}
		ptz := md.Dev.GetChannelPTZ(channelID)
		st, err := ptz.DeletePreset(presetID)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, st)

	default:
		writeErr(w, 405, "Method not allowed")
	}
}

func (s *Server) handleDevicePTZCallPreset(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	md, err := s.mgr.Get(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	if !md.Running || md.Dev == nil {
		writeErr(w, 400, "设备未运行")
		return
	}

	var req struct {
		ChannelID string `json:"channelId"`
		PresetID  int    `json:"presetId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "无效请求体: "+err.Error())
		return
	}
	channelID := req.ChannelID
	if channelID == "" && len(md.Profile.Device.Channels) > 0 {
		channelID = md.Profile.Device.Channels[0].ID
	}
	ptz := md.Dev.GetChannelPTZ(channelID)
	st, err := ptz.CallPreset(req.PresetID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, st)
}

func (s *Server) handleChannelAdd(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req device.AddChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.AddChannel(id, req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "added", "id": req.ID})
}

func (s *Server) handleChannelUpdate(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req device.UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.UpdateChannel(id, req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "updated", "id": req.ID})
}

func (s *Server) handleChannelRemove(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	chID := r.URL.Query().Get("id")
	if chID == "" {
		writeErr(w, 400, "id required")
		return
	}
	if err := s.mgr.RemoveChannel(id, chID); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "removed"})
}

func (s *Server) handleChannelBind(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req device.BindChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.BindChannelVideo(id, req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "bound"})
}

func (s *Server) handleChannelAudio(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req struct {
		ChannelID    string `json:"channelId"`
		AudioEnabled *bool  `json:"audioEnabled"`
		AudioSource  string `json:"audioSource"`
		AudioFile    string `json:"audioFile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	bindReq := device.BindChannelRequest{
		ChannelID:    req.ChannelID,
		AudioEnabled: req.AudioEnabled,
		AudioSource:  req.AudioSource,
		AudioFile:    req.AudioFile,
	}
	if err := s.mgr.BindChannelVideo(id, bindReq); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "channelId": req.ChannelID})
}

func (s *Server) handleMediaMode(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.SetMediaMode(id, body.Mode); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": body.Mode})
}

func (s *Server) handleChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req struct {
		ChannelID string `json:"channelId"`
		Status    string `json:"status"` // ON | OFF
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.SetChannelStatus(id, req.ChannelID, req.Status); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "status_updated", "channelId": req.ChannelID, "status": req.Status})
}

func (s *Server) handleDeviceCatalogNotify(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req struct {
		ChannelID string `json:"channelId"`
		Event     string `json:"event"` // ON, OFF, VLOST, DEFECT, ADD, DEL, UPDATE
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.SendCatalogNotify(id, req.ChannelID, req.Event); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "catalog_notify_sent", "channelId": req.ChannelID, "event": req.Event})
}

func (s *Server) handleDeviceGPSStatus(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "GET only")
		return
	}
	chID := r.URL.Query().Get("channelId")
	if chID != "" {
		st, cfg, err := s.mgr.GetGPSStatus(id, chID)
		if err != nil {
			writeErr(w, 404, err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{
			"gps":       st,
			"config":    cfg,
			"channelId": chID,
		})
		return
	}

	masterSt, masterCfg, chStatuses, chConfigs, err := s.mgr.GetAllGPS(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"gps":             masterSt,
		"config":          masterCfg,
		"channelStatuses": chStatuses,
		"channelConfigs":  chConfigs,
	})
}

func (s *Server) handleDeviceGPSConfig(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeErr(w, 405, "POST or PUT only")
		return
	}
	chID := r.URL.Query().Get("channelId")
	var req config.MobilePositionConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if chID == "" {
		chID = req.ChannelID
	}
	if err := s.mgr.UpdateGPSConfig(id, req, chID); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	st, cfg, _ := s.mgr.GetGPSStatus(id, chID)
	writeJSON(w, 200, map[string]any{
		"ok":        true,
		"gps":       st,
		"config":    cfg,
		"channelId": chID,
	})
}

func (s *Server) handleDeviceGPSReport(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	chID := r.URL.Query().Get("channelId")
	st, err := s.mgr.ReportGPSNow(id, chID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":        true,
		"gps":       st,
		"channelId": chID,
	})
}

func (s *Server) handleDeviceGPSSync(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req struct {
		Config     config.MobilePositionConfig `json:"config"`
		FollowMode bool                        `json:"followMode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.mgr.SyncAllChannelsGPS(id, req.Config, req.FollowMode); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	masterSt, masterCfg, chStatuses, chConfigs, _ := s.mgr.GetAllGPS(id)
	writeJSON(w, 200, map[string]any{
		"ok":              true,
		"gps":             masterSt,
		"config":          masterCfg,
		"channelStatuses": chStatuses,
		"channelConfigs":  chConfigs,
	})
}

func (s *Server) handleDeviceSubscriptions(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "GET only")
		return
	}
	subs, err := s.mgr.ListSubscriptions(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	if subs == nil {
		subs = []*device.Subscriber{}
	}
	writeJSON(w, 200, map[string]any{
		"subscribers": subs,
	})
}

func (s *Server) handleVideos(w http.ResponseWriter, r *http.Request) {
	list, err := device.ListVideos()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []device.VideoItem{}
	}
	writeJSON(w, 200, map[string]any{"videos": list})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeErr(w, 400, "解析上传失败: "+err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, 400, "缺少 file 字段")
		return
	}
	defer file.Close()

	name := filepath.Base(header.Filename)
	name = strings.ReplaceAll(name, `\`, "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".mp4" && ext != ".h264" && ext != ".264" {
		writeErr(w, 400, "仅支持 .mp4 / .h264")
		return
	}
	if name == "" || name == "." || name == ".." {
		writeErr(w, 400, "非法文件名")
		return
	}

	if err := os.MkdirAll(device.AssetsDir, 0o755); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	dstPath := filepath.Join(device.AssetsDir, name)
	dst, err := os.Create(dstPath)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	defer dst.Close()
	n, err := io.Copy(dst, file)
	if err != nil {
		_ = os.Remove(dstPath)
		writeErr(w, 500, "写入失败: "+err.Error())
		return
	}
	log.Printf("[ui] uploaded %s (%d bytes)", dstPath, n)
	writeJSON(w, 200, map[string]any{
		"ok":   true,
		"name": name,
		"path": filepath.ToSlash(filepath.Join(device.AssetsDir, name)),
		"size": n,
	})
}

func (s *Server) handleDeleteVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		var req struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		name = req.Name
	}
	if name == "" {
		writeErr(w, 400, "缺少 name 参数")
		return
	}
	if err := device.DeleteVideo(name); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "deleted", "name": name})
}

func (s *Server) handleDeviceControlEvents(w http.ResponseWriter, r *http.Request, id string) {
	events, err := s.mgr.GetControlEvents(id)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":     true,
		"events": events,
	})
}

func (s *Server) handleDeviceControlIFrame(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	channelID := r.URL.Query().Get("channelId")
	if channelID == "" {
		var body struct {
			ChannelID string `json:"channelId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		channelID = body.ChannelID
	}
	applied, err := s.mgr.ForceIFrame(id, channelID)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":        true,
		"channelId": channelID,
		"applied":   applied,
	})
}

func (s *Server) handleDeviceControlReboot(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	if err := s.mgr.TriggerReboot(id); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":      true,
		"message": "远程重启流程已启动",
	})
}

func (s *Server) handleDeviceControlRecord(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var body struct {
		Recording bool `json:"recording"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "bad json: "+err.Error())
		return
	}
	if err := s.mgr.SetRecording(id, body.Recording); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":        true,
		"recording": body.Recording,
	})
}

func (s *Server) handleDeviceControlTime(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var body struct {
		Time  string `json:"time"`
		Reset bool   `json:"reset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "bad json: "+err.Error())
		return
	}
	if body.Reset {
		if err := s.mgr.ResetTime(id); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{
			"ok":         true,
			"reset":      true,
			"deviceTime": time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}
	t, err := s.mgr.SetTime(id, body.Time)
	if err != nil {
		writeErr(w, 400, "时间格式错误(支持 YYYY-MM-DD HH:mm:ss 或 HH:mm:ss): "+err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":         true,
		"deviceTime": t.Format("2006-01-02 15:04:05"),
	})
}

func (s *Server) handleDeviceControlHomePosition(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var body struct {
		ChannelID   string `json:"channelId"`
		Enabled     bool   `json:"enabled"`
		PresetIndex int    `json:"presetIndex"`
		ResetSec    int    `json:"resetSec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "bad json: "+err.Error())
		return
	}
	st, err := s.mgr.SetHomePosition(id, body.ChannelID, body.Enabled, body.PresetIndex, body.ResetSec)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"ok":  true,
		"ptz": st,
	})
}

func (s *Server) handleLegacyStatus(w http.ResponseWriter, r *http.Request) {
	summaries := s.mgr.ListSummaries()
	if len(summaries) == 0 {
		writeJSON(w, 200, map[string]any{"devices": []any{}})
		return
	}
	// 返回第一台设备的 status
	md, err := s.mgr.Get(summaries[0].ID)
	if err == nil && md.Running && md.Dev != nil {
		writeJSON(w, 200, md.Dev.Status())
		return
	}
	writeJSON(w, 200, map[string]any{"deviceId": summaries[0].ID, "running": false})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}
