package api

import (
	"errors"
	"fmt"
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
	DataDir       string `json:"data_dir"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type statusResponse struct {
	Running                  bool     `json:"running"`
	LastRefreshAt            string   `json:"last_refresh_at"`
	LastSuccessAt            string   `json:"last_success_at"`
	NextRefreshAt            string   `json:"next_refresh_at"`
	RefreshInterval          int      `json:"refresh_interval"`
	Refreshing               bool     `json:"refreshing"`
	RefreshStage             string   `json:"refresh_stage,omitempty"`
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

type workspaceResponse struct {
	OK          bool   `json:"ok"`
	WorkspaceID string `json:"workspace_id"`
	ExpiresAt   string `json:"expires_at"`
}

type publishedStatusResponse struct {
	OK                      bool   `json:"ok"`
	PublishID               string `json:"publish_id,omitempty"`
	URL                     string `json:"url,omitempty"`
	SubscriptionURL         string `json:"subscription_url,omitempty"`
	TokenHint               string `json:"token_hint,omitempty"`
	CreatedAt               string `json:"created_at,omitempty"`
	UpdatedAt               string `json:"updated_at,omitempty"`
	LastAccessAt            string `json:"last_access_at,omitempty"`
	AccessCount             int    `json:"access_count"`
	Status                  string `json:"status,omitempty"`
	HasSubscriptionUserinfo bool   `json:"has_subscription_userinfo,omitempty"`
	SubscriptionInfoHeader  string `json:"subscription_info_header,omitempty"`
}

type bindPublishedRequest struct {
	PublishID string `json:"publish_id"`
}

type restoreDraftPublishRefRequest struct {
	PublishID string `json:"publish_id"`
}

type restoreDraftRequest struct {
	Config     model.Config                  `json:"config"`
	PublishRef restoreDraftPublishRefRequest `json:"publish_ref,omitempty"`
	NodeState  *model.NodeState              `json:"node_state,omitempty"`
}

type restoreDraftPublishResponse struct {
	Exists                  bool   `json:"exists"`
	Reason                  string `json:"reason,omitempty"`
	PublishID               string `json:"publish_id,omitempty"`
	URL                     string `json:"url,omitempty"`
	SubscriptionURL         string `json:"subscription_url,omitempty"`
	TokenHint               string `json:"token_hint,omitempty"`
	CreatedAt               string `json:"created_at,omitempty"`
	UpdatedAt               string `json:"updated_at,omitempty"`
	LastAccessAt            string `json:"last_access_at,omitempty"`
	AccessCount             int    `json:"access_count"`
	Status                  string `json:"status,omitempty"`
	HasSubscriptionUserinfo bool   `json:"has_subscription_userinfo,omitempty"`
	SubscriptionInfoHeader  string `json:"subscription_info_header,omitempty"`
}

type restoreDraftResponse struct {
	OK          bool                        `json:"ok"`
	WorkspaceID string                      `json:"workspace_id"`
	Publish     restoreDraftPublishResponse `json:"publish"`
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

type auditResponse struct {
	OK            bool                 `json:"ok"`
	RawCount      int                  `json:"raw_count"`
	FinalCount    int                  `json:"final_count"`
	ExcludedCount int                  `json:"excluded_count"`
	ExcludedNodes []model.ExcludedNode `json:"excluded_nodes,omitempty"`
	Warnings      []model.AuditWarning `json:"warnings,omitempty"`
}

type refreshResponse struct {
	OK              bool   `json:"ok"`
	NodeCount       int    `json:"node_count"`
	OutputPath      string `json:"output_path"`
	PublishID       string `json:"publish_id,omitempty"`
	TokenHint       string `json:"token_hint,omitempty"`
	SubscriptionURL string `json:"subscription_url,omitempty"`
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
		DataDir:       s.baseDataDir(),
		UptimeSeconds: s.uptimeSeconds(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
		return
	}
	status := s.snapshotWorkspaceStatus(ref.Hash)
	if status.StartedAt.IsZero() {
		status = model.RuntimeStatus{
			StartedAt:                ref.Meta.CreatedAt,
			Running:                  true,
			RefreshInterval:          effectiveRefreshInterval(cfg),
			NextRefreshAt:            nextRefreshTime(time.Now().UTC(), cfg),
			YAMLExists:               yamlFileExists(cfg.Service.OutputPath),
			YAMLUpdatedAt:            yamlFileUpdatedAt(cfg.Service.OutputPath),
			UpstreamSourceCount:      enabledSubscriptionCount(cfg.Subscriptions),
			EnabledSubscriptionCount: enabledSubscriptionCount(cfg.Subscriptions),
			OutputPath:               cfg.Service.OutputPath,
		}
	}
	writeJSON(w, http.StatusOK, statusResponse{
		Running:                  status.Running,
		LastRefreshAt:            formatTime(status.LastRefreshAt),
		LastSuccessAt:            formatTime(status.LastSuccessAt),
		NextRefreshAt:            formatTime(status.NextRefreshAt),
		RefreshInterval:          status.RefreshInterval,
		Refreshing:               status.Refreshing,
		RefreshStage:             status.RefreshStage,
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

	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
		return
	}
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

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
		return
	}
	state, err := pipeline.LoadNodeState(cfg)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_LOAD_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, auditResponse{
		OK:            true,
		RawCount:      state.LastAudit.RawCount,
		FinalCount:    state.LastAudit.FinalCount,
		ExcludedCount: state.LastAudit.ExcludedCount,
		ExcludedNodes: state.LastAudit.ExcludedNodes,
		Warnings:      state.LastAudit.Warnings,
	})
}

func (s *Server) handleValidateOutput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
		return
	}
	state, err := pipeline.LoadNodeState(cfg)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_LOAD_FAILED", err.Error())
		return
	}
	collected := pipeline.CollectNodes(cfg)
	finalSet, audit, err := pipeline.BuildFinalNodes(cfg, state, collected.Nodes)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "BUILD_FINAL_NODES_FAILED", err.Error())
		return
	}
	data, err := os.ReadFile(cfg.Service.OutputPath)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "YAML_READ_FAILED", err.Error())
		return
	}
	if err := pipeline.ValidateFinalConfig(data, finalSet, audit, renderer.OptionsFromConfig(cfg)); err != nil {
		writeAPIError(w, http.StatusBadRequest, "OUTPUT_LEAK_DETECTED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handlePreviewYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
		return
	}
	data, err := os.ReadFile(cfg.Service.OutputPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "YAML_READ_FAILED", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
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
		ref, err := s.requireWorkspace(r)
		if err != nil {
			handleWorkspaceError(w, err)
			return
		}
		cfg, err := s.loadWorkspaceConfig(ref)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, configResponse{
			OK:     true,
			Config: RedactConfig(cfg),
		})
	case http.MethodPut:
		ref, err := s.requireWorkspace(r)
		if err != nil {
			handleWorkspaceError(w, err)
			return
		}
		var req model.Config
		if err := decodeJSONBody(r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		req = config.Normalize(req)
		req.Service.OutputPath = ref.OutputPath
		req.Service.StatePath = ref.StatePath
		req.Service.CacheDir = ref.CacheDir
		if err := config.Validate(req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
			return
		}
		if err := config.WriteJSON(ref.ConfigPath, req); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "CONFIG_WRITE_FAILED", err.Error())
			return
		}
		s.appendLog("workspace config updated: " + ref.Hash)
		writeJSON(w, http.StatusOK, configResponse{
			OK:     true,
			Config: RedactConfig(req),
		})
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (s *Server) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	ref, err := s.createWorkspace()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "WORKSPACE_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, workspaceResponse{
		OK:          true,
		WorkspaceID: ref.ID,
		ExpiresAt:   formatTime(s.workspaceExpiresAt(ref.Meta)),
	})
}

func (s *Server) handleWorkspaceSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/workspaces/"), "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	id := strings.TrimSpace(parts[0])
	if id == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 2 && parts[1] == "bind-publish" {
		s.handleBindWorkspacePublished(w, r, id)
		return
	}
	if len(parts) == 2 && parts[1] == "restore-draft" {
		s.handleRestoreWorkspaceDraft(w, r, id)
		return
	}
	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		methodNotAllowed(w, http.MethodDelete)
		return
	}
	if err := s.deleteWorkspace(id); err != nil {
		handleWorkspaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleBindWorkspacePublished(w http.ResponseWriter, r *http.Request, workspaceID string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var req bindPublishedRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	publishID := strings.TrimSpace(req.PublishID)
	if publishID == "" {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "publish_id is required")
		return
	}
	published, err := s.loadPublishedByID(publishID)
	if err != nil || !publishedRestorable(published) {
		writeAPIError(w, http.StatusNotFound, "PUBLISHED_NOT_FOUND", "published subscription not found")
		return
	}
	ref, err := s.loadWorkspace(workspaceID)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	ref.Meta.PublishID = publishID
	ref.Meta.LegacyPublishedToken = ""
	ref.Meta.LegacyPublishedAt = time.Time{}
	if err := s.saveWorkspaceMeta(ref); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "WORKSPACE_BIND_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleRestoreWorkspaceDraft(w http.ResponseWriter, r *http.Request, workspaceID string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var req restoreDraftRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	ref, err := s.loadWorkspace(workspaceID)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	cfg := config.Normalize(req.Config)
	cfg.Service.OutputPath = ref.OutputPath
	cfg.Service.StatePath = ref.StatePath
	cfg.Service.CacheDir = ref.CacheDir
	if err := config.Validate(cfg); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_CONFIG", err.Error())
		return
	}
	if err := config.WriteJSON(ref.ConfigPath, cfg); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_WRITE_FAILED", err.Error())
		return
	}
	if req.NodeState != nil {
		state := model.NormalizeNodeState(*req.NodeState)
		state.SubscriptionMeta = nil
		state.LastAudit = model.AuditReport{}
		if err := pipeline.SaveNodeState(cfg, state); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
			return
		}
	}

	publishResp := restoreDraftPublishResponse{Exists: false}
	publishID := strings.TrimSpace(req.PublishRef.PublishID)
	if publishID != "" {
		published, err := s.loadPublishedByID(publishID)
		if err == nil && publishedRestorable(published) {
			ref.Meta.PublishID = published.ID
			ref.Meta.LegacyPublishedToken = ""
			ref.Meta.LegacyPublishedAt = time.Time{}
			publishResp = restoreDraftPublishedStatus(r, s.publishedStatus(r, published))
		} else {
			ref.Meta.PublishID = ""
			ref.Meta.LegacyPublishedToken = ""
			ref.Meta.LegacyPublishedAt = time.Time{}
			publishResp.Reason = "PUBLISHED_NOT_FOUND"
		}
	}
	if err := s.saveWorkspaceMeta(ref); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "WORKSPACE_RESTORE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, restoreDraftResponse{
		OK:          true,
		WorkspaceID: ref.ID,
		Publish:     publishResp,
	})
}

func (s *Server) handlePublished(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	if strings.TrimSpace(ref.Meta.PublishID) == "" {
		writeJSON(w, http.StatusOK, publishedStatusResponse{OK: true})
		return
	}
	published, err := s.loadPublishedByID(ref.Meta.PublishID)
	if err != nil || !publishedRestorable(published) {
		writeJSON(w, http.StatusOK, publishedStatusResponse{OK: true})
		return
	}
	writeJSON(w, http.StatusOK, s.publishedStatus(r, published))
}

func (s *Server) handlePublishedSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/published/"), "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	switch {
	case strings.HasSuffix(path, "/rotate-token"):
		publishID := strings.TrimSuffix(path, "/rotate-token")
		s.handleRotatePublishedToken(w, r, strings.Trim(publishID, "/"))
	case !strings.Contains(path, "/"):
		if r.Method == http.MethodGet {
			s.handleGetPublishedByID(w, r, path)
			return
		}
		s.handleDeletePublished(w, r, path)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleGetPublishedByID(w http.ResponseWriter, r *http.Request, publishID string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	published, err := s.loadPublishedByID(publishID)
	if err != nil || !publishedRestorable(published) {
		writeAPIError(w, http.StatusNotFound, "PUBLISHED_NOT_FOUND", "published subscription not found")
		return
	}
	writeJSON(w, http.StatusOK, s.publishedStatus(r, published))
}

func (s *Server) handleRotatePublishedToken(w http.ResponseWriter, r *http.Request, publishID string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	published, err := s.rotatePublishedToken(publishID)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "PUBLISHED_NOT_FOUND", "published subscription not found")
		return
	}
	writeJSON(w, http.StatusOK, s.publishedStatus(r, published))
}

func (s *Server) handleDeletePublished(w http.ResponseWriter, r *http.Request, publishID string) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w, http.MethodDelete)
		return
	}
	if err := s.deletePublished(publishID); err != nil {
		writeAPIError(w, http.StatusNotFound, "PUBLISHED_NOT_FOUND", "published subscription not found")
		return
	}
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) publishedStatus(r *http.Request, published publishedRef) publishedStatusResponse {
	url := publishedURL(s.publicOrigin(r), published.Meta.Token)
	userinfoHeader := publishedSubscriptionUserinfoHeader(published.Meta.SubscriptionInfo)
	return publishedStatusResponse{
		OK:                      true,
		PublishID:               published.ID,
		URL:                     url,
		SubscriptionURL:         url,
		TokenHint:               published.Meta.TokenHint,
		CreatedAt:               formatTime(published.Meta.CreatedAt),
		UpdatedAt:               formatTime(published.Meta.UpdatedAt),
		LastAccessAt:            formatTime(published.Meta.LastAccessAt),
		AccessCount:             published.Meta.AccessCount,
		Status:                  "active",
		HasSubscriptionUserinfo: userinfoHeader != "",
		SubscriptionInfoHeader:  userinfoHeader,
	}
}

func restoreDraftPublishedStatus(_ *http.Request, status publishedStatusResponse) restoreDraftPublishResponse {
	return restoreDraftPublishResponse{
		Exists:                  true,
		PublishID:               status.PublishID,
		URL:                     status.URL,
		SubscriptionURL:         status.SubscriptionURL,
		TokenHint:               status.TokenHint,
		CreatedAt:               status.CreatedAt,
		UpdatedAt:               status.UpdatedAt,
		LastAccessAt:            status.LastAccessAt,
		AccessCount:             status.AccessCount,
		Status:                  status.Status,
		HasSubscriptionUserinfo: status.HasSubscriptionUserinfo,
		SubscriptionInfoHeader:  status.SubscriptionInfoHeader,
	}
}

func publishedRestorable(published publishedRef) bool {
	if published.Meta.Revoked || strings.TrimSpace(published.Meta.Token) == "" {
		return false
	}
	if _, err := os.Stat(published.CurrentPath); err != nil {
		return false
	}
	return true
}

func handleWorkspaceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errWorkspaceRequired):
		writeAPIError(w, http.StatusBadRequest, "WORKSPACE_REQUIRED", "workspace is required")
	case errors.Is(err, errWorkspaceNotFound):
		writeAPIError(w, http.StatusNotFound, "WORKSPACE_NOT_FOUND", "workspace not found or expired")
	default:
		writeAPIError(w, http.StatusInternalServerError, "WORKSPACE_ERROR", err.Error())
	}
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
		return
	}
	published, created, err := s.ensureWorkspacePublishedRef(&ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "PUBLISH_PREPARE_FAILED", err.Error())
		return
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error())
		return
	}
	cfg.Service.OutputPath = published.CurrentPath
	outcome, err := s.startRefreshForConfig(cfg, ref.Hash, "manual refresh")
	if errors.Is(err, ErrRefreshInProgress) {
		writeAPIError(w, http.StatusConflict, "REFRESH_IN_PROGRESS", "refresh is already running")
		return
	}
	if err != nil {
		s.releaseWorkspacePublishedRef(&ref, published, created)
		message := err.Error()
		if errors.Is(err, pipeline.ErrNoNodes) {
			message = "no nodes available for rendering"
		}
		writeAPIError(w, http.StatusBadRequest, "REFRESH_FAILED", message)
		return
	}
	if err := s.finalizePublishedRefresh(&ref, &published, cfg, outcome.Result); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "PUBLISH_FAILED", err.Error())
		return
	}
	subscriptionURL := publishedURL(s.publicOrigin(r), published.Meta.Token)
	writeJSON(w, http.StatusOK, refreshResponse{
		OK:              true,
		NodeCount:       outcome.Result.NodeCount,
		OutputPath:      published.CurrentPath,
		PublishID:       published.ID,
		TokenHint:       published.Meta.TokenHint,
		SubscriptionURL: subscriptionURL,
	})
}

func (s *Server) handlePublishedSubscriptionYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/s/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "mihomo.yaml" {
		http.NotFound(w, r)
		return
	}
	data, published, err := s.loadPublishedYAML(parts[0])
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "subscription not found")
		return
	}
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	if wantsInlineSubscriptionYAML(r) {
		w.Header().Set("Content-Disposition", `inline; filename="mihomo.yaml"`)
	} else {
		w.Header().Set("Content-Disposition", `attachment; filename="mihomo.yaml"`)
	}
	w.Header().Set("Profile-Update-Interval", "24")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	userinfo := publishedSubscriptionUserinfoHeader(published.Meta.SubscriptionInfo)
	if userinfo != "" {
		w.Header().Set("Subscription-Userinfo", userinfo)
	}
	s.logPublishedSubscriptionServe(published, userinfo)
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(data)
}

func wantsInlineSubscriptionYAML(r *http.Request) bool {
	query := r.URL.Query()
	return query.Get("view") == "1" || query.Get("inline") == "1"
}

func (s *Server) handleDeprecatedSubscriptionYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	writeAPIError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "fixed subscription path is disabled; use /s/{token}/mihomo.yaml")
}

func (s *Server) handleSubscriptionYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cfg := s.snapshotConfig()
	if !subscriptionTokenAllowed(cfg, r) {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid subscription token")
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
	token := strings.TrimSpace(cfg.Service.AccessToken)
	if token == "" {
		token = strings.TrimSpace(cfg.Service.SubscriptionToken)
	}
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
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	w.Header().Set("X-Content-Type-Options", "nosniff")
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

func publishedSubscriptionUserinfoHeader(info *publishedSubscriptionUserinfo) string {
	if info == nil || !info.HeaderEnabled || info.Total <= 0 || info.Sources <= 0 {
		return ""
	}
	return model.FormatSubscriptionUserinfoHeader(model.AggregatedSubscriptionMeta{
		Upload:   info.Upload,
		Download: info.Download,
		Total:    info.Total,
		Expire:   info.Expire,
	})
}

func (s *Server) logPublishedSubscriptionServe(published publishedRef, userinfo string) {
	state := "missing"
	extra := ""
	if userinfo != "" {
		state = "present"
		extra = fmt.Sprintf(" header=%q", userinfo)
	}
	s.appendLog(fmt.Sprintf("serve subscription publish=%s token_hint=%s meta=ok userinfo=%s%s", published.ID, published.Meta.TokenHint, state, extra))
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

func normalizePublicOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "127.0.0.1:9876"
	}
	return scheme + "://" + host
}

func (s *Server) publicOrigin(r *http.Request) string {
	if publicBaseURL := strings.TrimRight(strings.TrimSpace(s.snapshotConfig().Service.PublicBaseURL), "/"); publicBaseURL != "" {
		return publicBaseURL
	}
	return normalizePublicOrigin(r)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	ref, err := s.requireWorkspace(r)
	if err != nil {
		handleWorkspaceError(w, err)
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
		Lines: s.snapshotWorkspaceLogs(ref.Hash, tail),
	})
}

func nodeNames(nodes []model.NodeIR) []string {
	out := make([]string, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, node.Name)
	}
	return out
}
