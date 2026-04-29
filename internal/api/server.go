package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	staticui "subconv-next/internal/api/static"
	"subconv-next/internal/model"
)

type Server struct {
	version string

	mu              sync.RWMutex
	config          model.Config
	configPath      string
	status          model.RuntimeStatus
	logLines        []string
	workspaceStatus map[string]model.RuntimeStatus
	workspaceLogs   map[string][]string
	logWriteMu      sync.Mutex

	refreshMu      sync.Mutex
	refreshRunning bool
	refreshDone    chan struct{}

	siteLogoMu    sync.RWMutex
	siteLogoCache map[string]siteLogoCacheEntry

	refreshBeforeRun func()
	refreshAfterRun  func()
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
			UpstreamSourceCount:      enabledSubscriptionCount(cfg.Subscriptions),
			RefreshInterval:          effectiveRefreshInterval(cfg),
			NextRefreshAt:            nextRefreshTime(now, cfg),
			OutputPath:               cfg.Service.OutputPath,
			YAMLExists:               yamlFileExists(cfg.Service.OutputPath),
			YAMLUpdatedAt:            yamlFileUpdatedAt(cfg.Service.OutputPath),
		},
		siteLogoCache:   map[string]siteLogoCacheEntry{},
		workspaceStatus: map[string]model.RuntimeStatus{},
		workspaceLogs:   map[string][]string{},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/workspaces", s.handleWorkspaces)
	mux.HandleFunc("/api/workspaces/", s.handleWorkspaceSubroutes)
	mux.HandleFunc("/api/published", s.handlePublished)
	mux.HandleFunc("/api/published/", s.handlePublishedSubroutes)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/site-logo", s.handleSiteLogo)
	mux.HandleFunc("/api/subscription-meta", s.handleSubscriptionMeta)
	mux.HandleFunc("/api/audit", s.handleAudit)
	mux.HandleFunc("/api/preview-yaml", s.handlePreviewYAML)
	mux.HandleFunc("/api/validate-output", s.handleValidateOutput)
	mux.HandleFunc("/api/nodes", s.handleNodes)
	mux.HandleFunc("/api/nodes/", s.handleNodeSubroutes)
	mux.HandleFunc("/api/parse", s.handleParse)
	mux.HandleFunc("/api/generate", s.handleGenerate)
	mux.HandleFunc("/api/refresh", s.handleRefresh)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/s/", s.handlePublishedSubscriptionYAML)
	mux.HandleFunc("/sub/mihomo.yaml", s.handleDeprecatedSubscriptionYAML)
	mux.HandleFunc("/favicon.svg", serveEmbeddedAsset("favicon.svg", "image/svg+xml"))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
			return
		}
		http.Redirect(w, r, "/favicon.svg", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/style.css", serveEmbeddedAsset("style.css", "text/css; charset=utf-8"))
	mux.HandleFunc("/app.js", serveEmbeddedAsset("app.js", "application/javascript; charset=utf-8"))
	mux.HandleFunc("/", serveIndex)
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
	s.status.UpstreamSourceCount = enabledSubscriptionCount(cfg.Subscriptions)
	s.status.RefreshInterval = effectiveRefreshInterval(cfg)
	s.status.OutputPath = cfg.Service.OutputPath
	s.status.NextRefreshAt = nextRefreshTime(time.Now().UTC(), cfg)
	s.status.YAMLExists = yamlFileExists(cfg.Service.OutputPath)
	s.status.YAMLUpdatedAt = yamlFileUpdatedAt(cfg.Service.OutputPath)
}

func (s *Server) appendLog(message string) {
	if message == "" {
		return
	}
	line := fmt.Sprintf("%s %s", time.Now().UTC().Format(time.RFC3339), maskSensitiveText(message))
	s.mu.Lock()
	s.logLines = append(s.logLines, line)
	if len(s.logLines) > 500 {
		s.logLines = append([]string(nil), s.logLines[len(s.logLines)-500:]...)
	}
	s.mu.Unlock()
	s.writeLogLine(line)
}

func (s *Server) snapshotLogs(tail int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tail <= 0 || tail >= len(s.logLines) {
		return maskLogLines(s.logLines)
	}
	return maskLogLines(s.logLines[len(s.logLines)-tail:])
}

func (s *Server) snapshotWorkspaceLogs(workspaceHash string, tail int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lines := s.workspaceLogs[workspaceHash]
	if tail <= 0 || tail >= len(lines) {
		return maskLogLines(lines)
	}
	return maskLogLines(lines[len(lines)-tail:])
}

func (s *Server) appendWorkspaceLog(workspaceHash, message string) {
	if strings.TrimSpace(workspaceHash) == "" || message == "" {
		return
	}
	line := fmt.Sprintf("%s %s", time.Now().UTC().Format(time.RFC3339), maskSensitiveText(message))
	s.mu.Lock()
	lines := append(s.workspaceLogs[workspaceHash], line)
	if len(lines) > 500 {
		lines = append([]string(nil), lines[len(lines)-500:]...)
	}
	s.workspaceLogs[workspaceHash] = lines
	s.mu.Unlock()
	s.writeLogLine("workspace=" + shortNodeID(workspaceHash) + " " + line)
}

func (s *Server) snapshotWorkspaceStatus(workspaceHash string) model.RuntimeStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workspaceStatus[workspaceHash]
}

func (s *Server) setWorkspaceStatus(workspaceHash string, status model.RuntimeStatus) {
	if strings.TrimSpace(workspaceHash) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workspaceStatus[workspaceHash] = status
}

func (s *Server) setRefreshSuccess(nodeCount int, nodeNames []string, outputPath string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.Running = true
	s.status.LastRefreshAt = at
	s.status.LastSuccessAt = at
	s.status.NextRefreshAt = nextRefreshTime(at, s.config)
	s.status.NodeCount = nodeCount
	s.status.NodeNames = append([]string(nil), nodeNames...)
	s.status.OutputPath = outputPath
	s.status.YAMLExists = true
	s.status.YAMLUpdatedAt = at
	s.status.LastError = ""
	s.status.RefreshStage = ""
}

func (s *Server) setRefreshFailure(message string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.Running = true
	s.status.LastRefreshAt = at
	s.status.NextRefreshAt = nextRefreshTime(at, s.config)
	s.status.YAMLExists = yamlFileExists(s.config.Service.OutputPath)
	s.status.YAMLUpdatedAt = yamlFileUpdatedAt(s.config.Service.OutputPath)
	s.status.LastError = message
	s.status.RefreshStage = ""
}

func (s *Server) setRefreshing(refreshing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Refreshing = refreshing
	if !refreshing {
		s.status.RefreshStage = ""
	}
}

func (s *Server) setRefreshStage(stage string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.RefreshStage = strings.TrimSpace(stage)
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

const (
	maxAppLogBytes = 5 * 1024 * 1024
	maxAppLogFiles = 3
)

func (s *Server) writeLogLine(line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	s.logWriteMu.Lock()
	defer s.logWriteMu.Unlock()

	dir := s.logsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "app.log")
	if err := rotateLogFile(path, int64(len(line)+1)); err != nil {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(line + "\n")
}

func rotateLogFile(path string, incomingBytes int64) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size()+incomingBytes <= maxAppLogBytes {
		return nil
	}
	oldest := fmt.Sprintf("%s.%d", path, maxAppLogFiles)
	_ = os.Remove(oldest)
	for index := maxAppLogFiles - 1; index >= 1; index-- {
		current := fmt.Sprintf("%s.%d", path, index)
		next := fmt.Sprintf("%s.%d", path, index+1)
		if _, err := os.Stat(current); err == nil {
			_ = os.Rename(current, next)
		}
	}
	return os.Rename(path, path+".1")
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

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
		return
	}
	serveEmbeddedAsset("index.html", "text/html; charset=utf-8")(w, r)
}

func serveEmbeddedAsset(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
			return
		}
		data, err := staticui.Assets.ReadFile(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(data)
	}
}

func maskLogLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, maskSensitiveText(line))
	}
	return out
}

var (
	publishedPathPattern = regexp.MustCompile(`/s/[^/\s]+/mihomo\.yaml`)
	schemeSecretPattern  = regexp.MustCompile(`(?i)\b(ss|trojan|anytls|tuic|vless|vmess|wireguard|socks5|http)://([^@/\s]+)@`)
	uuidPattern          = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	secretPairPattern    = regexp.MustCompile(`(?i)\b(password|uuid|token|private[-_ ]?key|pre[-_ ]?shared[-_ ]?key|authorization|cookie)\s*[:=]\s*[^,\s"']+`)
)

func maskSensitiveText(value string) string {
	keys := []string{"token", "sig", "key", "auth", "password", "uuid", "private-key", "private_key", "pre-shared-key", "presharedkey", "access_token", "authorization", "cookie"}
	masked := value
	for _, key := range keys {
		masked = maskQueryValue(masked, key)
	}
	masked = publishedPathPattern.ReplaceAllString(masked, "/s/<redacted>/mihomo.yaml")
	masked = schemeSecretPattern.ReplaceAllString(masked, `$1://***@`)
	masked = uuidPattern.ReplaceAllString(masked, "***")
	masked = secretPairPattern.ReplaceAllStringFunc(masked, maskSecretPair)
	masked = maskHeaderValue(masked, "authorization")
	masked = maskHeaderValue(masked, "cookie")
	return masked
}

func maskSecretPair(input string) string {
	for index, r := range input {
		if r == ':' || r == '=' {
			return input[:index+1] + "***"
		}
	}
	return input
}

func maskQueryValue(input, key string) string {
	pattern := strings.ToLower(key) + "="
	lower := strings.ToLower(input)
	searchFrom := 0
	for {
		index := strings.Index(lower[searchFrom:], pattern)
		if index == -1 {
			return input
		}
		index += searchFrom
		start := index + len(pattern)
		end := start
		for end < len(input) {
			switch input[end] {
			case '&', ' ', '\n', '\r', '\t', '"', '\'':
				goto done
			}
			end++
		}
	done:
		input = input[:start] + "***" + input[end:]
		lower = strings.ToLower(input)
		searchFrom = start + 3
	}
}

func maskHeaderValue(input, key string) string {
	lower := strings.ToLower(input)
	pattern := strings.ToLower(key) + ":"
	index := strings.Index(lower, pattern)
	if index == -1 {
		return input
	}
	start := index + len(pattern)
	end := start
	for end < len(input) {
		switch input[end] {
		case '\n', '\r':
			goto done
		}
		end++
	}
done:
	return input[:start] + " ***" + input[end:]
}
