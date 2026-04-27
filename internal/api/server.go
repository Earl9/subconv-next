package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"subconv-next/internal/model"
	staticui "subconv-next/static"
)

type Server struct {
	version string

	mu         sync.RWMutex
	config     model.Config
	configPath string
	status     model.RuntimeStatus
	logLines   []string
}

func NewServer(version string, cfg model.Config) *Server {
	now := time.Now().UTC()

	return &Server{
		version: version,
		config:  cfg,
		status: model.RuntimeStatus{
			StartedAt:                now,
			Running:                  true,
			EnabledSubscriptionCount: enabledSubscriptionCount(cfg.Subscriptions),
			OutputPath:               cfg.Service.OutputPath,
		},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/nodes", s.handleNodes)
	mux.HandleFunc("/api/parse", s.handleParse)
	mux.HandleFunc("/api/generate", s.handleGenerate)
	mux.HandleFunc("/api/refresh", s.handleRefresh)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/sub/mihomo.yaml", s.handleSubscriptionYAML)
	mux.Handle("/", http.FileServer(http.FS(staticui.Assets)))
	return mux
}

func ListenAddress(cfg model.Config) string {
	return net.JoinHostPort(cfg.Service.ListenAddr, strconv.Itoa(cfg.Service.ListenPort))
}

func (s *Server) snapshotStatus() model.RuntimeStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Server) snapshotConfig() model.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Server) configFilePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configPath
}

func (s *Server) SetConfigPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configPath = path
}

func (s *Server) updateConfig(cfg model.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = cfg
	s.status.EnabledSubscriptionCount = enabledSubscriptionCount(cfg.Subscriptions)
	s.status.OutputPath = cfg.Service.OutputPath
}

func (s *Server) appendLog(message string) {
	if message == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	line := fmt.Sprintf("%s %s", time.Now().UTC().Format(time.RFC3339), message)
	s.logLines = append(s.logLines, line)
	if len(s.logLines) > 500 {
		s.logLines = append([]string(nil), s.logLines[len(s.logLines)-500:]...)
	}
}

func (s *Server) snapshotLogs(tail int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tail <= 0 || tail >= len(s.logLines) {
		return append([]string(nil), s.logLines...)
	}
	return append([]string(nil), s.logLines[len(s.logLines)-tail:]...)
}

func (s *Server) setRefreshSuccess(nodeCount int, nodeNames []string, outputPath string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.Running = true
	s.status.LastRefreshAt = at
	s.status.LastSuccessAt = at
	s.status.NodeCount = nodeCount
	s.status.NodeNames = append([]string(nil), nodeNames...)
	s.status.OutputPath = outputPath
	s.status.LastError = ""
}

func (s *Server) setRefreshFailure(message string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.Running = true
	s.status.LastRefreshAt = at
	s.status.LastError = message
}

func (s *Server) uptimeSeconds() int64 {
	startedAt := s.snapshotStatus().StartedAt
	if startedAt.IsZero() {
		return 0
	}

	uptime := int64(time.Since(startedAt).Seconds())
	if uptime < 0 {
		return 0
	}
	return uptime
}

func enabledSubscriptionCount(subscriptions []model.SubscriptionConfig) int {
	count := 0
	for _, sub := range subscriptions {
		if sub.Enabled {
			count++
		}
	}
	return count
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode request body: trailing content")
	}

	return nil
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func writeAPIError(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, map[string]any{
		"ok": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
