package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	NextRefreshAt            string   `json:"next_refresh_at"`
	RefreshInterval          int      `json:"refresh_interval"`
	Refreshing               bool     `json:"refreshing"`
	YAMLExists               bool     `json:"yaml_exists"`
	YAMLUpdatedAt            string   `json:"yaml_updated_at"`
	UpstreamSourceCount      int      `json:"upstream_source_count"`
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

type subscriptionMetaSourceResponse struct {
	SourceID     string  `json:"source_id,omitempty"`
	SourceName   string  `json:"source_name,omitempty"`
	Upload       int64   `json:"upload,omitempty"`
	Download     int64   `json:"download,omitempty"`
	Total        int64   `json:"total,omitempty"`
	Used         int64   `json:"used,omitempty"`
	Remaining    int64   `json:"remaining,omitempty"`
	UsedRatio    float64 `json:"used_ratio,omitempty"`
	Expire       int64   `json:"expire,omitempty"`
	FromHeader   bool    `json:"from_header,omitempty"`
	FromInfoNode bool    `json:"from_info_node,omitempty"`
	FetchedAt    string  `json:"fetched_at,omitempty"`
}

type subscriptionMetaResponse struct {
	OK        bool                             `json:"ok"`
	Aggregate model.AggregatedSubscriptionMeta `json:"aggregate"`
	Sources   []subscriptionMetaSourceResponse `json:"sources"`
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
		NextRefreshAt:            formatTime(status.NextRefreshAt),
		RefreshInterval:          status.RefreshInterval,
		Refreshing:               status.Refreshing,
		YAMLExists:               status.YAMLExists,
		YAMLUpdatedAt:            formatTime(status.YAMLUpdatedAt),
		UpstreamSourceCount:      status.UpstreamSourceCount,
		NodeCount:                status.NodeCount,
		NodeNames:                status.NodeNames,
		EnabledSubscriptionCount: status.EnabledSubscriptionCount,
		OutputPath:               status.OutputPath,
		LastError:                maskSensitiveText(status.LastError),
	})
}

func (s *Server) handleSubscriptionMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cfg := s.snapshotConfig()
	state, err := pipeline.LoadNodeState(cfg)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_LOAD_FAILED", err.Error())
		return
	}

	sources := pipeline.BuildSubscriptionMetaSources(cfg, state.SubscriptionMeta)
	responseSources := make([]subscriptionMetaSourceResponse, 0, len(sources))
	sourceMap := make(map[string]model.SubscriptionMeta, len(sources))
	for _, meta := range sources {
		meta = model.NormalizeSubscriptionMeta(meta)
		sourceMap[meta.SourceID] = meta
		responseSources = append(responseSources, subscriptionMetaSourceResponse{
			SourceID:     meta.SourceID,
			SourceName:   meta.SourceName,
			Upload:       meta.Upload,
			Download:     meta.Download,
			Total:        meta.Total,
			Used:         meta.Used,
			Remaining:    meta.Remaining,
			UsedRatio:    meta.UsedRatio,
			Expire:       meta.Expire,
			FromHeader:   meta.FromHeader,
			FromInfoNode: meta.FromInfoNode,
			FetchedAt:    meta.FetchedAt,
		})
	}

	writeJSON(w, http.StatusOK, subscriptionMetaResponse{
		OK:        true,
		Aggregate: pipeline.AggregateSubscriptionMetaForConfig(cfg, sourceMap),
		Sources:   responseSources,
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
		req = config.Normalize(req)
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

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	outcome, err := s.startRefresh("manual refresh")
	if errors.Is(err, ErrRefreshInProgress) {
		writeAPIError(w, http.StatusConflict, "REFRESH_IN_PROGRESS", "refresh is already running")
		return
	}
	if err != nil {
		message := err.Error()
		if errors.Is(err, pipeline.ErrNoNodes) {
			message = "no nodes available for rendering"
		}
		writeAPIError(w, http.StatusBadRequest, "REFRESH_FAILED", message)
		return
	}
	writeJSON(w, http.StatusOK, refreshResponse{
		OK:         true,
		NodeCount:  outcome.Result.NodeCount,
		OutputPath: outcome.Result.OutputPath,
	})
}

func (s *Server) handleSubscriptionYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cfg := s.snapshotConfig()
	if !subscriptionTokenAllowed(cfg, r) {
		writeAPIError(w, http.StatusForbidden, "INVALID_SUBSCRIPTION_TOKEN", "invalid subscription token")
		return
	}
	if rawURL := strings.TrimSpace(r.URL.Query().Get("url")); rawURL != "" {
		overrideCfg := cfg
		overrideCfg.Subscriptions = []model.SubscriptionConfig{
			{
				Name:      "quick-import",
				Enabled:   true,
				URL:       rawURL,
				UserAgent: model.DefaultUserAgent,
			},
		}
		overrideCfg.Inline = []model.InlineConfig{}

		result, err := pipeline.RenderConfig(overrideCfg)
		if err != nil {
			message := err.Error()
			if errors.Is(err, pipeline.ErrNoNodes) {
				message = "no nodes available for rendering"
			}
			writeAPIError(w, http.StatusBadRequest, "SUBSCRIPTION_UNAVAILABLE", message)
			return
		}

		writeSubscriptionYAML(w, r, overrideCfg, result.YAML, result.AggregateMeta, "fresh", "", result.NodeCount, time.Now().UTC())
		return
	}

	status := s.snapshotStatus()
	existing, readErr := os.ReadFile(cfg.Service.OutputPath)
	hasExisting := readErr == nil && len(existing) > 0
	expired := cacheExpired(cfg.Service.OutputPath, effectiveRefreshInterval(cfg))
	cachedAggregate := loadSubscriptionAggregate(cfg)

	if hasExisting && (!cfg.Service.RefreshOnRequest || !expired) {
		writeSubscriptionYAML(w, r, cfg, existing, cachedAggregate, "cached", "", status.NodeCount, status.YAMLUpdatedAt)
		return
	}

	outcome, err := s.startRefresh("subscription refresh")
	if errors.Is(err, ErrRefreshInProgress) {
		if hasExisting {
			writeSubscriptionYAML(w, r, cfg, existing, cachedAggregate, "cached", "", status.NodeCount, status.YAMLUpdatedAt)
			return
		}
		if s.waitForRefresh(time.Duration(cfg.Service.FetchTimeoutSeconds) * time.Second) {
			latest, latestErr := os.ReadFile(cfg.Service.OutputPath)
			if latestErr == nil && len(latest) > 0 {
				latestStatus := s.snapshotStatus()
				writeSubscriptionYAML(w, r, cfg, latest, loadSubscriptionAggregate(cfg), "fresh", "", latestStatus.NodeCount, latestStatus.YAMLUpdatedAt)
				return
			}
		}
		writeAPIError(w, http.StatusServiceUnavailable, "SUBSCRIPTION_UNAVAILABLE", "refresh is already running")
		return
	}
	if err != nil {
		message := err.Error()
		if errors.Is(err, pipeline.ErrNoNodes) {
			message = "no nodes available for rendering"
		}
		if cfg.Service.StaleIfError && hasExisting {
			writeSubscriptionYAML(w, r, cfg, existing, cachedAggregate, "stale", "upstream refresh failed, served stale config", status.NodeCount, status.YAMLUpdatedAt)
			return
		}
		writeAPIError(w, http.StatusServiceUnavailable, "SUBSCRIPTION_UNAVAILABLE", message)
		return
	}

	writeSubscriptionYAML(w, r, cfg, outcome.Result.YAML, outcome.Result.AggregateMeta, "fresh", "", outcome.Result.NodeCount, outcome.At)
}

func subscriptionTokenAllowed(cfg model.Config, r *http.Request) bool {
	token := strings.TrimSpace(cfg.Service.SubscriptionToken)
	if token == "" {
		return true
	}
	return r.URL.Query().Get("token") == token
}

func loadSubscriptionAggregate(cfg model.Config) model.AggregatedSubscriptionMeta {
	state, err := pipeline.LoadNodeState(cfg)
	if err != nil {
		return model.AggregatedSubscriptionMeta{}
	}
	return pipeline.AggregateSubscriptionMetaForConfig(cfg, state.SubscriptionMeta)
}

func writeSubscriptionYAML(w http.ResponseWriter, r *http.Request, cfg model.Config, data []byte, aggregate model.AggregatedSubscriptionMeta, refreshStatus, warning string, nodeCount int, generatedAt time.Time) {
	filename := sanitizeOutputFilename(cfg.Render.OutputFilename)
	if filename == "" {
		filename = "mihomo.yaml"
	}
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if userinfo := pipeline.BuildSubscriptionMetaHeader(cfg, aggregate); userinfo != "" {
		w.Header().Set("Subscription-Userinfo", userinfo)
	}
	if !generatedAt.IsZero() {
		w.Header().Set("X-SubConv-Generated-At", generatedAt.UTC().Format(time.RFC3339))
	}
	if nodeCount > 0 {
		w.Header().Set("X-SubConv-Node-Count", strconv.Itoa(nodeCount))
	}
	if refreshStatus != "" {
		w.Header().Set("X-SubConv-Refresh-Status", refreshStatus)
	}
	if warning != "" {
		w.Header().Set("X-SubConv-Warning", warning)
	}
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		w.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func sanitizeOutputFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = filepath.Base(value)
	value = strings.ReplaceAll(value, `"`, "")
	value = strings.ReplaceAll(value, `'`, "")
	return value
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
