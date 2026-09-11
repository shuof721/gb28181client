package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/device"
)

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
	st := s.dev.Status()
	writeJSON(w, 200, st)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if n <= 0 {
		n = 120
	}
	writeJSON(w, 200, map[string]any{"lines": s.dev.Logs(n)})
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
		ch = s.cfg.Device.Channels[0].ID
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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}
