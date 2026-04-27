package api

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
	"subconv-next/internal/parser"
	"subconv-next/internal/pipeline"
	"subconv-next/internal/renderer"
)

type healthzResponse struct {
	OK            bool   `json:"ok"`
	Version       string `json:"version"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type statusResponse struct {
	Running                  bool     `json:"running"`
	LastRefreshAt            string   `json:"last_refresh_at"`
	LastSuccessAt            string   `json:"last_success_at"`
	NodeCount                int      `json:"node_count"`
	NodeNames                []string `json:"node_names,omitempty"`
	EnabledSubscriptionCount int      `json:"enabled_subscription_count"`
	OutputPath               string   `json:"output_path"`
	LastError                string   `json:"last_error"`
}

type parseRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type parseResponse struct {
	OK       bool                `json:"ok"`
	Nodes    []model.NodeIR      `json:"nodes"`
	Warnings []string            `json:"warnings,omitempty"`
	Errors   []parser.ParseError `json:"errors,omitempty"`
}

type generateRequest struct {
	Nodes    []model.NodeIR `json:"nodes"`
	Template string         `json:"template"`
}

type generateResponse struct {
	OK   bool   `json:"ok"`
	YAML string `json:"yaml"`
}

type configResponse struct {
	OK     bool         `json:"ok"`
	Config model.Config `json:"config"`
}

type nodeSummary struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Server     string   `json:"server,omitempty"`
	Port       int      `json:"port,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	SourceName string   `json:"source_name,omitempty"`
	SourceKind string   `json:"source_kind,omitempty"`
}

type nodesResponse struct {
	OK       bool                `json:"ok"`
	Nodes    []nodeSummary       `json:"nodes"`
	Warnings []string            `json:"warnings,omitempty"`
	Errors   []parser.ParseError `json:"errors,omitempty"`
}

type refreshResponse struct {
	OK         bool   `json:"ok"`
	NodeCount  int    `json:"node_count"`
	OutputPath string `json:"output_path"`
}

type logsResponse struct {
	OK    bool     `json:"ok"`
	Lines []string `json:"lines"`
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, healthzResponse{
		OK:            true,
		Version:       s.version,
		UptimeSeconds: s.uptimeSeconds(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	status := s.snapshotStatus()
	writeJSON(w, http.StatusOK, statusResponse{
		Running:                  status.Running,
		LastRefreshAt:            formatTime(status.LastRefreshAt),
		LastSuccessAt:            formatTime(status.LastSuccessAt),
		NodeCount:                status.NodeCount,
		NodeNames:                status.NodeNames,
		EnabledSubscriptionCount: status.EnabledSubscriptionCount,
		OutputPath:               status.OutputPath,
		LastError:                status.LastError,
	})
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req parseRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	result := parser.ParseContent([]byte(req.Content), model.SourceInfo{
		Name: "api",
		Kind: "inline",
	})
	writeJSON(w, http.StatusOK, parseResponse{
		OK:       true,
		Nodes:    result.Nodes,
		Warnings: result.Warnings,
		Errors:   result.Errors,
	})
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req generateRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	opts := renderer.OptionsFromConfig(s.snapshotConfig())
	if req.Template != "" {
		opts.Template = req.Template
	}

	rendered, err := renderer.RenderMihomo(req.Nodes, opts)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "RENDER_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, generateResponse{
		OK:   true,
		YAML: string(rendered),
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, configResponse{
			OK:     true,
			Config: s.snapshotConfig(),
		})
	case http.MethodPut:
		var req model.Config
		if err := decodeJSONBody(r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		if err := config.Validate(req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}

		path := s.configFilePath()
		if path == "" {
			writeAPIError(w, http.StatusInternalServerError, "CONFIG_PATH_UNSET", "config path is not set")
			return
		}
		if err := config.WriteJSON(path, req); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "CONFIG_WRITE_FAILED", err.Error())
			return
		}

		s.updateConfig(req)
		s.appendLog("config updated: " + path)
		writeJSON(w, http.StatusOK, configResponse{
			OK:     true,
			Config: req,
		})
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	collected := pipeline.CollectPreviewNodes(s.snapshotConfig())
	nodes := make([]nodeSummary, 0, len(collected.Nodes))
	for _, node := range collected.Nodes {
		nodes = append(nodes, nodeSummary{
			ID:         node.ID,
			Name:       node.Name,
			Type:       string(node.Type),
			Server:     node.Server,
			Port:       node.Port,
			Tags:       append([]string(nil), node.Tags...),
			SourceName: node.Source.Name,
			SourceKind: node.Source.Kind,
		})
	}

	writeJSON(w, http.StatusOK, nodesResponse{
		OK:       true,
		Nodes:    nodes,
		Warnings: collected.Warnings,
		Errors:   collected.Errors,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	result, err := pipeline.RenderConfig(s.snapshotConfig())
	now := time.Now().UTC()
	if err != nil {
		message := err.Error()
		if errors.Is(err, pipeline.ErrNoNodes) {
			message = "no nodes available for rendering"
		}
		s.setRefreshFailure(message, now)
		s.appendLog("refresh failed: " + message)
		writeAPIError(w, http.StatusBadRequest, "REFRESH_FAILED", message)
		return
	}
	if err := pipeline.WriteRendered(result.OutputPath, result.YAML); err != nil {
		s.setRefreshFailure(err.Error(), now)
		s.appendLog("refresh write failed: " + err.Error())
		writeAPIError(w, http.StatusInternalServerError, "WRITE_FAILED", err.Error())
		return
	}

	s.setRefreshSuccess(result.NodeCount, nodeNames(result.Nodes), result.OutputPath, now)
	for _, warning := range result.Warnings {
		s.appendLog("refresh warning: " + warning)
	}
	for _, parseErr := range result.Errors {
		s.appendLog("refresh parse error [" + parseErr.Kind + "]: " + parseErr.Message)
	}
	s.appendLog("refresh succeeded: wrote " + result.OutputPath)
	writeJSON(w, http.StatusOK, refreshResponse{
		OK:         true,
		NodeCount:  result.NodeCount,
		OutputPath: result.OutputPath,
	})
}

func (s *Server) handleSubscriptionYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cfg := s.snapshotConfig()
	if existing, err := os.ReadFile(cfg.Service.OutputPath); err == nil && len(existing) > 0 {
		writeSubscriptionYAML(w, r, existing)
		return
	}

	result, err := pipeline.RenderConfig(cfg)
	now := time.Now().UTC()
	if err != nil {
		message := err.Error()
		if errors.Is(err, pipeline.ErrNoNodes) {
			message = "no nodes available for rendering"
		}
		s.setRefreshFailure(message, now)
		s.appendLog("subscription render failed: " + message)
		writeAPIError(w, http.StatusNotFound, "SUBSCRIPTION_UNAVAILABLE", message)
		return
	}
	if err := pipeline.WriteRendered(result.OutputPath, result.YAML); err != nil {
		s.setRefreshFailure(err.Error(), now)
		s.appendLog("subscription write failed: " + err.Error())
		writeAPIError(w, http.StatusInternalServerError, "WRITE_FAILED", err.Error())
		return
	}

	s.setRefreshSuccess(result.NodeCount, nodeNames(result.Nodes), result.OutputPath, now)
	for _, warning := range result.Warnings {
		s.appendLog("subscription warning: " + warning)
	}
	for _, parseErr := range result.Errors {
		s.appendLog("subscription parse error [" + parseErr.Kind + "]: " + parseErr.Message)
	}
	s.appendLog("subscription served: " + result.OutputPath)
	writeSubscriptionYAML(w, r, result.YAML)
}

func writeSubscriptionYAML(w http.ResponseWriter, r *http.Request, data []byte) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", `attachment; filename="mihomo.yaml"`)
	} else {
		w.Header().Set("Content-Disposition", `inline; filename="mihomo.yaml"`)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	tail := 200
	if rawTail := r.URL.Query().Get("tail"); rawTail != "" {
		if parsed, err := strconv.Atoi(rawTail); err == nil {
			tail = parsed
		}
	}
	if tail < 1 {
		tail = 1
	}
	if tail > 1000 {
		tail = 1000
	}

	writeJSON(w, http.StatusOK, logsResponse{
		OK:    true,
		Lines: s.snapshotLogs(tail),
	})
}

func nodeNames(nodes []model.NodeIR) []string {
	out := make([]string, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, node.Name)
	}
	return out
}
