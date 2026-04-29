package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"subconv-next/internal/model"
	"subconv-next/internal/parser"
	"subconv-next/internal/pipeline"
)

const maskedSecretValue = "********"

type nodeListSummary struct {
	Total    int `json:"total"`
	Enabled  int `json:"enabled"`
	Disabled int `json:"disabled"`
	Modified int `json:"modified"`
	Warnings int `json:"warnings"`
}

type nodeSourcePayload struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Kind    string `json:"kind,omitempty"`
	URLHash string `json:"url_hash,omitempty"`
}

type nodeTLSPayload struct {
	Enabled           bool   `json:"enabled"`
	SNI               string `json:"sni,omitempty"`
	ClientFingerprint string `json:"client_fingerprint,omitempty"`
	Insecure          bool   `json:"insecure,omitempty"`
}

type nodeListItem struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	Server           string            `json:"server,omitempty"`
	Port             int               `json:"port,omitempty"`
	Region           string            `json:"region,omitempty"`
	Enabled          bool              `json:"enabled"`
	Modified         bool              `json:"modified"`
	Source           nodeSourcePayload `json:"source"`
	UDP              bool              `json:"udp"`
	TLS              nodeTLSPayload    `json:"tls"`
	TransportNetwork string            `json:"transport_network,omitempty"`
	HasReality       bool              `json:"has_reality,omitempty"`
	HasIPv6          bool              `json:"has_ipv6,omitempty"`
	SourceCount      int               `json:"source_count,omitempty"`
	Warnings         []string          `json:"warnings,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
}

type nodeListResponse struct {
	OK       bool                `json:"ok"`
	Summary  nodeListSummary     `json:"summary"`
	Nodes    []nodeListItem      `json:"nodes"`
	Warnings []string            `json:"warnings,omitempty"`
	Errors   []parser.ParseError `json:"errors,omitempty"`
}

type nodeDetail struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Type         string                  `json:"type"`
	Server       string                  `json:"server,omitempty"`
	Port         int                     `json:"port,omitempty"`
	Region       string                  `json:"region,omitempty"`
	Enabled      bool                    `json:"enabled"`
	Modified     bool                    `json:"modified"`
	Tags         []string                `json:"tags,omitempty"`
	Source       nodeSourcePayload       `json:"source"`
	UDP          *bool                   `json:"udp,omitempty"`
	TLS          model.TLSOptions        `json:"tls,omitempty"`
	Auth         model.Auth              `json:"auth,omitempty"`
	Transport    model.TransportOptions  `json:"transport,omitempty"`
	WireGuard    *model.WireGuardOptions `json:"wireguard,omitempty"`
	Raw          map[string]interface{}  `json:"raw,omitempty"`
	Warnings     []string                `json:"warnings,omitempty"`
	Sources      []nodeSourcePayload     `json:"sources,omitempty"`
	SecretMasked bool                    `json:"secret_masked"`
	StableShort  string                  `json:"stable_short"`
}

type nodeDetailResponse struct {
	OK   bool       `json:"ok"`
	Node nodeDetail `json:"node"`
}

type nodeOverrideRequest struct {
	Enabled   bool                    `json:"enabled"`
	Name      string                  `json:"name"`
	Region    string                  `json:"region"`
	Tags      []string                `json:"tags"`
	Server    string                  `json:"server"`
	Port      int                     `json:"port"`
	UDP       *bool                   `json:"udp"`
	TLS       *model.TLSOptions       `json:"tls"`
	Auth      *model.Auth             `json:"auth"`
	Transport *model.TransportOptions `json:"transport"`
	WireGuard *model.WireGuardOptions `json:"wireguard"`
	Raw       map[string]interface{}  `json:"raw"`
}

type bulkRenameRequest struct {
	Scope       string   `json:"scope"`
	IDs         []string `json:"ids"`
	Mode        string   `json:"mode"`
	Prefix      string   `json:"prefix"`
	Suffix      string   `json:"suffix"`
	Pattern     string   `json:"pattern"`
	Replacement string   `json:"replacement"`
	Q           string   `json:"q,omitempty"`
	Type        string   `json:"type,omitempty"`
	Region      string   `json:"region,omitempty"`
	Status      string   `json:"status,omitempty"`
	Source      string   `json:"source,omitempty"`
}

type bulkRenamePreviewItem struct {
	ID  string `json:"id"`
	Old string `json:"old"`
	New string `json:"new"`
}

type bulkRenameResponse struct {
	OK      bool                    `json:"ok"`
	Changed int                     `json:"changed"`
	Preview []bulkRenamePreviewItem `json:"preview"`
}

type toggleNodesRequest struct {
	IDs []string `json:"ids"`
}

type customNodeRequest struct {
	Content     string       `json:"content"`
	ContentType string       `json:"content_type"`
	Node        model.NodeIR `json:"node"`
}

type validateNodesResponse struct {
	OK       bool                             `json:"ok"`
	Warnings []pipeline.NodeValidationWarning `json:"warnings"`
}

type genericOKResponse struct {
	OK bool `json:"ok"`
}

func (s *Server) handleNodeSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	path = strings.Trim(path, "/")
	if path == "" {
		s.handleNodes(w, r)
		return
	}

	switch {
	case path == "bulk-rename":
		s.handleBulkRename(w, r)
	case path == "delete":
		s.handleDeleteNodes(w, r)
	case path == "disable":
		s.handleToggleNodes(w, r, false)
	case path == "enable":
		s.handleToggleNodes(w, r, true)
	case path == "custom":
		s.handleCustomNodes(w, r)
	case strings.HasPrefix(path, "custom/"):
		s.handleDeleteCustomNode(w, r, strings.TrimPrefix(path, "custom/"))
	case path == "validate":
		s.handleValidateNodes(w, r)
	case path == "overrides/clear":
		s.handleClearOverrides(w, r)
	case strings.HasSuffix(path, "/override"):
		s.handleSaveNodeOverride(w, r, strings.TrimSuffix(path, "/override"))
	case strings.HasSuffix(path, "/reset"):
		s.handleResetNodeOverride(w, r, strings.TrimSuffix(path, "/reset"))
	default:
		s.handleNodeDetail(w, r, path)
	}
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}
	collected := pipeline.CollectNodesWithState(cfg, state, true, false)
	disabled := model.DisabledNodeSet(state.DisabledNodes)

	filtered := filterNodeList(collected.Nodes, state, disabled, r)
	paged := filtered
	if !wantsAllNodes(r) {
		page, pageSize := parsePage(r)
		paged = paginateNodes(filtered, page, pageSize)
	}

	writeJSON(w, http.StatusOK, nodeListResponse{
		OK:       true,
		Summary:  summarizeNodes(filtered, state, disabled),
		Nodes:    buildNodeListItems(paged, state, disabled),
		Warnings: collected.Warnings,
		Errors:   collected.Errors,
	})
	_ = ref
}

func wantsAllNodes(r *http.Request) bool {
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get("all"))) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func (s *Server) handleNodeDetail(w http.ResponseWriter, r *http.Request, nodeID string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	node, state, cfg, ref, err := s.lookupNode(r, nodeID)
	if err != nil {
		handleWorkspaceOrNodeError(w, err)
		return
	}

	detail := buildNodeDetail(node, state)
	writeJSON(w, http.StatusOK, nodeDetailResponse{
		OK:   true,
		Node: detail,
	})
	_, _ = cfg, ref
}

func (s *Server) handleSaveNodeOverride(w http.ResponseWriter, r *http.Request, nodeID string) {
	if r.Method != http.MethodPut {
		methodNotAllowed(w, http.MethodPut)
		return
	}

	node, state, cfg, ref, err := s.lookupNode(r, nodeID)
	if err != nil {
		handleWorkspaceOrNodeError(w, err)
		return
	}

	var req nodeOverrideRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Region = strings.ToUpper(strings.TrimSpace(req.Region))
	if req.Name == "" {
		req.Name = node.Name
	}
	if strings.TrimSpace(req.Server) == "" {
		req.Server = node.Server
	}
	if req.Port == 0 {
		req.Port = node.Port
	}
	if req.UDP == nil {
		req.UDP = node.UDP
	}
	req.Auth = preserveMaskedAuth(node.Auth, req.Auth)
	req.Raw = preserveMaskedRaw(node.Raw, req.Raw)
	if req.TLS == nil {
		tlsCopy := node.TLS
		req.TLS = &tlsCopy
	}
	if req.Transport == nil {
		transportCopy := node.Transport
		req.Transport = &transportCopy
	}
	if req.WireGuard == nil && node.WireGuard != nil {
		wgCopy := *node.WireGuard
		req.WireGuard = &wgCopy
	}
	if len(req.Tags) == 0 {
		req.Tags = append([]string(nil), node.Tags...)
	}

	override := model.NodeOverride{
		Enabled: req.Enabled,
		Name:    req.Name,
		Region:  req.Region,
		Tags:    append([]string(nil), req.Tags...),
		Fields: model.NodeOverrideFields{
			Server:    req.Server,
			Port:      req.Port,
			UDP:       req.UDP,
			TLS:       req.TLS,
			Auth:      req.Auth,
			Transport: req.Transport,
			WireGuard: req.WireGuard,
			Raw:       req.Raw,
		},
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if state.NodeOverrides == nil {
		state.NodeOverrides = map[string]model.NodeOverride{}
	}
	state.NodeOverrides[nodeID] = override
	if req.Enabled {
		state.DisabledNodes = removeIDs(state.DisabledNodes, []string{nodeID})
	} else {
		state.DisabledNodes = addIDs(state.DisabledNodes, []string{nodeID})
	}

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, "node override updated: "+shortNodeID(nodeID))
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleResetNodeOverride(w http.ResponseWriter, r *http.Request, nodeID string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}
	delete(state.NodeOverrides, nodeID)
	state.DisabledNodes = removeIDs(state.DisabledNodes, []string{nodeID})

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, "node override reset: "+shortNodeID(nodeID))
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleToggleNodes(w http.ResponseWriter, r *http.Request, enabled bool) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req toggleNodesRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}

	if enabled {
		state.DisabledNodes = removeIDs(state.DisabledNodes, req.IDs)
		state.DeletedNodes = removeIDs(state.DeletedNodes, req.IDs)
	} else {
		state.DisabledNodes = addIDs(state.DisabledNodes, req.IDs)
	}

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	if enabled {
		s.appendWorkspaceLog(ref.Hash, fmt.Sprintf("nodes enabled: %d", len(req.IDs)))
	} else {
		s.appendWorkspaceLog(ref.Hash, fmt.Sprintf("nodes disabled: %d", len(req.IDs)))
	}
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleDeleteNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req toggleNodesRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}
	state.DeletedNodes = addIDs(state.DeletedNodes, req.IDs)
	state.DisabledNodes = removeIDs(state.DisabledNodes, req.IDs)

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, fmt.Sprintf("nodes deleted: %d", len(req.IDs)))
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleBulkRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req bulkRenameRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}
	collected := pipeline.CollectNodesWithState(cfg, state, true, false)
	disabled := model.DisabledNodeSet(state.DisabledNodes)
	targets := resolveBulkRenameTargets(collected.Nodes, state, disabled, req)

	var changed int
	var preview []bulkRenamePreviewItem
	if state.NodeOverrides == nil {
		state.NodeOverrides = map[string]model.NodeOverride{}
	}

	for _, node := range targets {
		newName, err := buildRenamedNodeName(node, req)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_RENAME", err.Error())
			return
		}
		if newName == "" || newName == node.Name {
			continue
		}

		override := state.NodeOverrides[node.ID]
		override.Enabled = true
		override.Name = newName
		if len(override.Tags) == 0 {
			override.Tags = append([]string(nil), node.Tags...)
		}
		if strings.TrimSpace(override.Region) == "" {
			override.Region = model.NodeRegionCode(node)
		}
		override.Fields.Server = firstNonEmptyString(override.Fields.Server, node.Server)
		override.Fields.Port = firstNonZero(override.Fields.Port, node.Port)
		if override.Fields.UDP == nil {
			override.Fields.UDP = node.UDP
		}
		if override.Fields.TLS == nil {
			tlsCopy := node.TLS
			override.Fields.TLS = &tlsCopy
		}
		if override.Fields.Auth == nil {
			authCopy := node.Auth
			override.Fields.Auth = &authCopy
		}
		if override.Fields.Transport == nil {
			transportCopy := node.Transport
			override.Fields.Transport = &transportCopy
		}
		if override.Fields.WireGuard == nil && node.WireGuard != nil {
			wgCopy := *node.WireGuard
			override.Fields.WireGuard = &wgCopy
		}
		if override.Fields.Raw == nil && node.Raw != nil {
			override.Fields.Raw = cloneRaw(node.Raw)
		}
		override.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		state.NodeOverrides[node.ID] = override
		changed++
		if len(preview) < 10 {
			preview = append(preview, bulkRenamePreviewItem{ID: node.ID, Old: node.Name, New: newName})
		}
	}

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, fmt.Sprintf("bulk rename applied: changed=%d", changed))
	writeJSON(w, http.StatusOK, bulkRenameResponse{
		OK:      true,
		Changed: changed,
		Preview: preview,
	})
}

func (s *Server) handleCustomNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req customNodeRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}

	var newNodes []model.NodeIR
	if strings.TrimSpace(req.Content) != "" {
		parsed := parser.ParseContent([]byte(req.Content), model.SourceInfo{Name: "manual", Kind: "custom"})
		if len(parsed.Nodes) == 0 {
			writeAPIError(w, http.StatusBadRequest, "PARSE_FAILED", "未解析到可用节点")
			return
		}
		newNodes = parsed.Nodes
	} else {
		req.Node.Source = model.SourceInfo{Name: "manual", Kind: "custom"}
		newNodes = []model.NodeIR{req.Node}
	}

	for _, node := range newNodes {
		node.Source = model.SourceInfo{Name: "manual", Kind: "custom"}
		node = model.NormalizeNode(node)
		if !strings.HasPrefix(node.ID, "custom-") {
			node.ID = "custom-" + node.ID
		}
		state.CustomNodes = append(state.CustomNodes, node)
	}

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, fmt.Sprintf("custom nodes added: %d", len(newNodes)))
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleDeleteCustomNode(w http.ResponseWriter, r *http.Request, nodeID string) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w, http.MethodDelete)
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}

	filtered := state.CustomNodes[:0]
	for _, node := range state.CustomNodes {
		if node.ID == nodeID {
			continue
		}
		filtered = append(filtered, node)
	}
	state.CustomNodes = filtered
	delete(state.NodeOverrides, nodeID)
	state.DisabledNodes = removeIDs(state.DisabledNodes, []string{nodeID})
	state.DeletedNodes = removeIDs(state.DeletedNodes, []string{nodeID})

	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, "custom node deleted: "+shortNodeID(nodeID))
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) handleValidateNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	cfg, state, _, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}
	collected := pipeline.CollectNodesWithState(cfg, state, true, false)
	writeJSON(w, http.StatusOK, validateNodesResponse{
		OK:       true,
		Warnings: collected.ValidationWarnings,
	})
}

func (s *Server) handleClearOverrides(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		handleWorkspaceOrStateError(w, err)
		return
	}
	state.NodeOverrides = map[string]model.NodeOverride{}
	state.DisabledNodes = []string{}
	state.DeletedNodes = []string{}
	if err := pipeline.SaveNodeState(cfg, state); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "STATE_SAVE_FAILED", err.Error())
		return
	}

	s.appendWorkspaceLog(ref.Hash, "node overrides cleared")
	writeJSON(w, http.StatusOK, genericOKResponse{OK: true})
}

func (s *Server) lookupNode(r *http.Request, nodeID string) (model.NodeIR, model.NodeState, model.Config, workspaceRef, error) {
	cfg, state, ref, err := s.workspaceConfigAndState(r)
	if err != nil {
		return model.NodeIR{}, model.NodeState{}, model.Config{}, workspaceRef{}, err
	}
	collected := pipeline.CollectNodesWithState(cfg, state, true, false)
	for _, node := range collected.Nodes {
		if node.ID == nodeID {
			return node, state, cfg, ref, nil
		}
	}
	return model.NodeIR{}, state, cfg, ref, fmt.Errorf("node %s not found", shortNodeID(nodeID))
}

func (s *Server) workspaceConfigAndState(r *http.Request) (model.Config, model.NodeState, workspaceRef, error) {
	ref, err := s.requireWorkspace(r)
	if err != nil {
		return model.Config{}, model.NodeState{}, workspaceRef{}, err
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		return model.Config{}, model.NodeState{}, workspaceRef{}, err
	}
	state, err := pipeline.LoadNodeState(cfg)
	if err != nil {
		return model.Config{}, model.NodeState{}, workspaceRef{}, err
	}
	return cfg, state, ref, nil
}

func handleWorkspaceOrStateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errWorkspaceRequired), errors.Is(err, errWorkspaceNotFound):
		handleWorkspaceError(w, err)
	default:
		writeAPIError(w, http.StatusInternalServerError, "STATE_LOAD_FAILED", err.Error())
	}
}

func handleWorkspaceOrNodeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errWorkspaceRequired), errors.Is(err, errWorkspaceNotFound):
		handleWorkspaceError(w, err)
	default:
		writeAPIError(w, http.StatusNotFound, "NODE_NOT_FOUND", err.Error())
	}
}

func filterNodeList(nodes []model.NodeIR, state model.NodeState, disabled map[string]struct{}, r *http.Request) []model.NodeIR {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	typeFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	regionFilter := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("region")))
	statusFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	sourceFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))

	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		if q != "" {
			haystack := strings.ToLower(strings.Join([]string{
				node.Name,
				node.Server,
				string(node.Type),
				model.NodeRegionLabel(model.NodeRegionCode(node)),
				model.NodeRegionCode(node),
				strings.Join(node.Tags, " "),
			}, " "))
			if !strings.Contains(haystack, q) {
				continue
			}
		}
		if typeFilter != "" && typeFilter != "all" && strings.ToLower(string(node.Type)) != typeFilter {
			continue
		}
		if regionFilter != "" && regionFilter != "ALL" && model.NodeRegionCode(node) != regionFilter {
			continue
		}
		if sourceFilter != "" && !strings.Contains(strings.ToLower(node.Source.Name), sourceFilter) {
			continue
		}
		if !matchesStatusFilter(node, state, disabled, statusFilter) {
			continue
		}
		out = append(out, node)
	}
	return out
}

func matchesStatusFilter(node model.NodeIR, state model.NodeState, disabled map[string]struct{}, status string) bool {
	switch status {
	case "", "all", "全部节点":
		return true
	case "enabled", "已启用":
		_, ok := disabled[node.ID]
		return !ok
	case "disabled", "已禁用":
		_, ok := disabled[node.ID]
		return ok
	case "modified", "已修改":
		_, ok := state.NodeOverrides[node.ID]
		return ok
	case "warning", "有警告":
		return len(node.Warnings) > 0
	default:
		return true
	}
}

func buildNodeListItems(nodes []model.NodeIR, state model.NodeState, disabled map[string]struct{}) []nodeListItem {
	items := make([]nodeListItem, 0, len(nodes))
	for _, node := range nodes {
		_, isDisabled := disabled[node.ID]
		_, isModified := state.NodeOverrides[node.ID]
		items = append(items, nodeListItem{
			ID:       node.ID,
			Name:     node.Name,
			Type:     string(node.Type),
			Server:   node.Server,
			Port:     node.Port,
			Region:   model.NodeRegionCode(node),
			Enabled:  !isDisabled,
			Modified: isModified,
			Source: nodeSourcePayload{
				ID:      node.Source.ID,
				Name:    node.Source.Name,
				Kind:    node.Source.Kind,
				URLHash: node.Source.URLHash,
			},
			UDP: derefBool(node.UDP),
			TLS: nodeTLSPayload{
				Enabled:           node.TLS.Enabled,
				SNI:               node.TLS.SNI,
				ClientFingerprint: node.TLS.ClientFingerprint,
				Insecure:          node.TLS.Insecure,
			},
			TransportNetwork: strings.TrimSpace(node.Transport.Network),
			HasReality:       node.TLS.Reality != nil,
			HasIPv6:          strings.Contains(node.Server, ":") || (node.WireGuard != nil && strings.TrimSpace(node.WireGuard.IPv6) != ""),
			SourceCount:      sourceCountFallback(node),
			Warnings:         append([]string(nil), node.Warnings...),
			Tags:             append([]string(nil), node.Tags...),
		})
	}
	return items
}

func summarizeNodes(nodes []model.NodeIR, state model.NodeState, disabled map[string]struct{}) nodeListSummary {
	summary := nodeListSummary{Total: len(nodes)}
	for _, node := range nodes {
		if _, ok := disabled[node.ID]; ok {
			summary.Disabled++
		} else {
			summary.Enabled++
		}
		if _, ok := state.NodeOverrides[node.ID]; ok {
			summary.Modified++
		}
		if len(node.Warnings) > 0 {
			summary.Warnings++
		}
	}
	return summary
}

func buildNodeDetail(node model.NodeIR, state model.NodeState) nodeDetail {
	disabled := model.DisabledNodeSet(state.DisabledNodes)
	_, isDisabled := disabled[node.ID]
	_, isModified := state.NodeOverrides[node.ID]
	return nodeDetail{
		ID:       node.ID,
		Name:     node.Name,
		Type:     string(node.Type),
		Server:   node.Server,
		Port:     node.Port,
		Region:   model.NodeRegionCode(node),
		Enabled:  !isDisabled,
		Modified: isModified,
		Tags:     append([]string(nil), node.Tags...),
		Source: nodeSourcePayload{
			ID:      node.Source.ID,
			Name:    node.Source.Name,
			Kind:    node.Source.Kind,
			URLHash: node.Source.URLHash,
		},
		UDP:          node.UDP,
		TLS:          node.TLS,
		Auth:         maskAuthSecrets(node.Auth),
		Transport:    maskTransportSecrets(node.Transport),
		WireGuard:    maskWireGuardSecrets(node.WireGuard),
		Raw:          maskSensitiveMap(node.Raw),
		Warnings:     append([]string(nil), node.Warnings...),
		Sources:      buildSourcePayloads(node),
		SecretMasked: true,
		StableShort:  shortNodeID(node.ID),
	}
}

func buildSourcePayloads(node model.NodeIR) []nodeSourcePayload {
	sources := model.MergeSourcesForView(node)
	out := make([]nodeSourcePayload, 0, len(sources))
	for _, source := range sources {
		out = append(out, nodeSourcePayload{
			ID:      source.ID,
			Name:    source.Name,
			Kind:    source.Kind,
			URLHash: source.URLHash,
		})
	}
	return out
}

func sourceCountFallback(node model.NodeIR) int {
	sources := model.MergeSourcesForView(node)
	if len(sources) == 0 {
		return 1
	}
	return len(sources)
}

func parsePage(r *http.Request) (int, int) {
	page := 1
	pageSize := 200
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("page_size")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			pageSize = n
		}
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return page, pageSize
}

func paginateNodes(nodes []model.NodeIR, page, pageSize int) []model.NodeIR {
	if pageSize <= 0 || len(nodes) <= pageSize {
		return nodes
	}
	start := (page - 1) * pageSize
	if start >= len(nodes) {
		return []model.NodeIR{}
	}
	end := start + pageSize
	if end > len(nodes) {
		end = len(nodes)
	}
	return nodes[start:end]
}

func preserveMaskedAuth(current model.Auth, incoming *model.Auth) *model.Auth {
	if incoming == nil {
		authCopy := current
		return &authCopy
	}
	if incoming.UUID == maskedSecretValue {
		incoming.UUID = current.UUID
	}
	if incoming.Password == maskedSecretValue {
		incoming.Password = current.Password
	}
	if incoming.Token == maskedSecretValue {
		incoming.Token = current.Token
	}
	if incoming.PrivateKey == maskedSecretValue {
		incoming.PrivateKey = current.PrivateKey
	}
	if incoming.PreSharedKey == maskedSecretValue {
		incoming.PreSharedKey = current.PreSharedKey
	}
	return incoming
}

func preserveMaskedRaw(current, incoming map[string]interface{}) map[string]interface{} {
	if incoming == nil {
		return cloneRaw(current)
	}

	out := cloneRaw(incoming)
	for key, value := range out {
		if masked, ok := value.(string); ok && masked == maskedSecretValue {
			if existing, ok := current[key]; ok {
				out[key] = existing
			}
		}
		if child, ok := value.(map[string]interface{}); ok {
			existing, _ := current[key].(map[string]interface{})
			out[key] = preserveMaskedRaw(existing, child)
		}
	}
	return out
}

func maskAuthSecrets(auth model.Auth) model.Auth {
	if strings.TrimSpace(auth.UUID) != "" {
		auth.UUID = maskedSecretValue
	}
	if strings.TrimSpace(auth.Password) != "" {
		auth.Password = maskedSecretValue
	}
	if strings.TrimSpace(auth.Token) != "" {
		auth.Token = maskedSecretValue
	}
	if strings.TrimSpace(auth.PrivateKey) != "" {
		auth.PrivateKey = maskedSecretValue
	}
	if strings.TrimSpace(auth.PreSharedKey) != "" {
		auth.PreSharedKey = maskedSecretValue
	}
	return auth
}

func maskSensitiveMap(raw map[string]interface{}) map[string]interface{} {
	if raw == nil {
		return nil
	}
	out := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		switch typed := value.(type) {
		case string:
			if isSensitiveField(key) && strings.TrimSpace(typed) != "" {
				out[key] = maskedSecretValue
			} else {
				out[key] = typed
			}
		case map[string]interface{}:
			out[key] = maskSensitiveMap(typed)
		default:
			out[key] = value
		}
	}
	return out
}

func cloneRaw(raw map[string]interface{}) map[string]interface{} {
	if raw == nil {
		return nil
	}
	data, _ := json.Marshal(raw)
	var cloned map[string]interface{}
	_ = json.Unmarshal(data, &cloned)
	return cloned
}

func addIDs(existing, ids []string) []string {
	set := model.DisabledNodeSet(existing)
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		set[id] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func removeIDs(existing, ids []string) []string {
	set := model.DisabledNodeSet(existing)
	for _, id := range ids {
		delete(set, strings.TrimSpace(id))
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func resolveBulkRenameTargets(nodes []model.NodeIR, state model.NodeState, disabled map[string]struct{}, req bulkRenameRequest) []model.NodeIR {
	switch strings.ToLower(strings.TrimSpace(req.Scope)) {
	case "selected":
		set := model.DisabledNodeSet(req.IDs)
		var out []model.NodeIR
		for _, node := range nodes {
			if _, ok := set[node.ID]; ok {
				out = append(out, node)
			}
		}
		return out
	case "current_filtered":
		if len(req.IDs) > 0 {
			set := model.DisabledNodeSet(req.IDs)
			var out []model.NodeIR
			for _, node := range nodes {
				if _, ok := set[node.ID]; ok {
					out = append(out, node)
				}
			}
			return out
		}
		urlValues := make(urlValues)
		urlValues.set("q", req.Q)
		urlValues.set("type", req.Type)
		urlValues.set("region", req.Region)
		urlValues.set("status", req.Status)
		urlValues.set("source", req.Source)
		fakeReq, _ := http.NewRequest(http.MethodGet, "/api/nodes?"+urlValues.encode(), nil)
		return filterNodeList(nodes, state, disabled, fakeReq)
	default:
		return nodes
	}
}

func buildRenamedNodeName(node model.NodeIR, req bulkRenameRequest) (string, error) {
	switch strings.ToLower(strings.TrimSpace(req.Mode)) {
	case "add_prefix":
		return strings.TrimSpace(req.Prefix + node.Name), nil
	case "add_suffix":
		return strings.TrimSpace(node.Name + req.Suffix), nil
	case "regex_replace":
		re, err := regexp.Compile(req.Pattern)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(re.ReplaceAllString(node.Name, req.Replacement)), nil
	case "protocol_prefix":
		return strings.TrimSpace("[" + strings.ToUpper(string(node.Type)) + "] " + node.Name), nil
	case "region_emoji":
		emoji := model.NodeRegionEmoji(model.NodeRegionCode(node))
		if emoji == "" || strings.HasPrefix(node.Name, emoji+" ") {
			return node.Name, nil
		}
		return strings.TrimSpace(emoji + " " + node.Name), nil
	case "remove_info_text":
		re := regexp.MustCompile(`(?i)(剩余流量[:：]?\s*[^ ]+|到期时间[:：]?\s*[^ ]+|官网[:：]?\s*[^ ]+|套餐[:：]?\s*[^ ]+)`)
		cleaned := strings.TrimSpace(re.ReplaceAllString(node.Name, ""))
		cleaned = strings.Join(strings.Fields(cleaned), " ")
		return cleaned, nil
	default:
		return "", fmt.Errorf("unsupported rename mode %q", req.Mode)
	}
}

func isSensitiveField(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{"password", "uuid", "private", "private-key", "pre_shared", "pre-shared", "preshared", "token", "secret", "authorization", "cookie"} {
		if strings.Contains(key, token) {
			return true
		}
	}
	return false
}

func shortNodeID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func derefBool(value *bool) bool {
	return value != nil && *value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

type urlValues map[string]string

func (v urlValues) set(key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	v[key] = value
}

func (v urlValues) encode() string {
	if len(v) == 0 {
		return ""
	}
	keys := make([]string, 0, len(v))
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+v[key])
	}
	return strings.Join(parts, "&")
}
