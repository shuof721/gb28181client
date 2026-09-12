package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/device"
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
		s.handleDeviceAlarm(w, r, id)
	case "session":
		if len(parts) >= 3 && parts[2] == "stop" {
			s.handleDeviceStopSession(w, r, id)
		} else {
			writeErr(w, 404, "Not found")
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
			default:
				writeErr(w, 404, "Not found")
			}
		} else {
			writeErr(w, 404, "Not found")
		}
	case "media":
		if len(parts) >= 3 && parts[2] == "mode" {
			s.handleMediaMode(w, r, id)
		} else {
			writeErr(w, 404, "Not found")
		}
	case "records":
		s.handleDeviceRecords(w, r, id)
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
		ChannelID   string `json:"channelId"`
		AlarmMethod string `json:"alarmMethod"`
		Priority    string `json:"priority"`
		Description string `json:"description"`
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
			req.AlarmMethod = "2" // 默认移动侦测
		}
	}
	if req.Priority == "" {
		req.Priority = r.URL.Query().Get("priority")
		if req.Priority == "" {
			req.Priority = "4"
		}
	}
	if req.Description == "" {
		req.Description = r.URL.Query().Get("desc")
		if req.Description == "" {
			req.Description = "Web 控制台模拟报警"
		}
	}
	if err := dev.SendAlarmAdvanced(req.ChannelID, req.AlarmMethod, req.Priority, req.Description); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "alarm sent", "device": id, "channel": req.ChannelID})
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
