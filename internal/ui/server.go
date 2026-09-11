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
	dev *device.Device
	cfg *config.Config
	mux *http.ServeMux
}

func New(dev *device.Device, cfg *config.Config) *Server {
	s := &Server{dev: dev, cfg: cfg, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/register", s.handleRegister)
	s.mux.HandleFunc("/api/keepalive", s.handleKeepalive)
	s.mux.HandleFunc("/api/alarm", s.handleAlarm)
	s.mux.HandleFunc("/api/session/stop", s.handleStopSession)
	s.mux.HandleFunc("/api/videos", s.handleVideos)
	s.mux.HandleFunc("/api/videos/upload", s.handleUpload)
	s.mux.HandleFunc("/api/channels/add", s.handleAddChannel)
	s.mux.HandleFunc("/api/channels/remove", s.handleRemoveChannel)
	s.mux.HandleFunc("/api/channels/bind", s.handleBindChannel)
	s.mux.HandleFunc("/api/media/mode", s.handleMediaMode)
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

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.dev.Status())
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if n <= 0 {
		n = 150
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	lines := s.dev.Logs(n * 3)
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

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	if err := s.dev.RegisterNow(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "registered"})
}

func (s *Server) handleKeepalive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	if err := s.dev.KeepaliveNow(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "keepalive sent"})
}

func (s *Server) handleAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	ch := r.URL.Query().Get("channel")
	if ch == "" {
		st := s.dev.Status()
		if len(st.Channels) == 0 {
			writeErr(w, 400, "no channels")
			return
		}
		ch = st.Channels[0].ID
	}
	if err := s.dev.SendAlarm(ch, "UI trigger"); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "alarm sent", "channel": ch})
}

func (s *Server) handleStopSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	id := r.URL.Query().Get("callId")
	if id == "" {
		writeErr(w, 400, "callId required")
		return
	}
	s.dev.StopSession(id)
	writeJSON(w, 200, map[string]string{"ok": "stopped"})
}

func (s *Server) handleVideos(w http.ResponseWriter, r *http.Request) {
	list, err := s.dev.ListVideos()
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

func (s *Server) handleAddChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req device.AddChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.dev.AddChannel(req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "added", "id": req.ID})
}

func (s *Server) handleRemoveChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, 400, "id required")
		return
	}
	if err := s.dev.RemoveChannel(id); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "removed"})
}

func (s *Server) handleBindChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "POST only")
		return
	}
	var req device.BindChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	if err := s.dev.BindChannelVideo(req); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "bound"})
}

func (s *Server) handleMediaMode(w http.ResponseWriter, r *http.Request) {
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
	if err := s.dev.SetMediaMode(body.Mode); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ok": body.Mode})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}
