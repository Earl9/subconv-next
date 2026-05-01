package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

func newTestServer(t *testing.T, cfg model.Config) (*Server, model.Config) {
	t.Helper()
	cfg = testConfigWithDataDir(t, cfg)
	return NewServer("0.1.0-test", cfg), cfg
}

func testConfigWithDataDir(t *testing.T, cfg model.Config) model.Config {
	t.Helper()
	cfg = config.Normalize(cfg)
	dataDir := configuredTestDataDir(cfg)
	if dataDir == "" {
		dataDir = t.TempDir()
	}
	if shouldUseTestPath(cfg.Service.StatePath, model.DefaultStatePath) {
		cfg.Service.StatePath = filepath.Join(dataDir, "state.json")
	}
	if shouldUseTestPath(cfg.Service.CacheDir, model.DefaultCacheDir) {
		cfg.Service.CacheDir = filepath.Join(dataDir, "cache")
	}
	if shouldUseTestPath(cfg.Service.OutputPath, model.DefaultOutputPath) {
		cfg.Service.OutputPath = filepath.Join(dataDir, "mihomo.yaml")
	}
	return cfg
}

func configuredTestDataDir(cfg model.Config) string {
	for _, candidate := range []struct {
		value        string
		defaultValue string
	}{
		{cfg.Service.StatePath, model.DefaultStatePath},
		{cfg.Service.OutputPath, model.DefaultOutputPath},
		{cfg.Service.CacheDir, model.DefaultCacheDir},
	} {
		value := strings.TrimSpace(candidate.value)
		if value == "" || value == candidate.defaultValue || !filepath.IsAbs(value) {
			continue
		}
		if filepath.Base(value) == "cache" {
			return filepath.Dir(value)
		}
		return filepath.Dir(value)
	}
	return ""
}

func shouldUseTestPath(value, defaultValue string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == defaultValue
}

func createWorkspaceForTest(t *testing.T, server *Server, cfg model.Config) string {
	t.Helper()
	ref := createWorkspaceRefForTest(t, server, cfg)
	return ref.ID
}

func createWorkspaceRefForTest(t *testing.T, server *Server, cfg model.Config) workspaceRef {
	t.Helper()
	ref, err := server.createWorkspace()
	if err != nil {
		t.Fatalf("createWorkspace() error = %v", err)
	}
	cfg = config.Normalize(cfg)
	cfg.Service.OutputPath = ref.OutputPath
	cfg.Service.StatePath = ref.StatePath
	cfg.Service.CacheDir = ref.CacheDir
	if err := config.WriteJSON(ref.ConfigPath, cfg); err != nil {
		t.Fatalf("WriteJSON(workspace config) error = %v", err)
	}
	server.setWorkspaceStatus(ref.Hash, model.RuntimeStatus{
		StartedAt:                time.Now().UTC(),
		Running:                  true,
		RefreshInterval:          cfg.Service.RefreshInterval,
		UpstreamSourceCount:      enabledSubscriptionCount(cfg.Subscriptions),
		EnabledSubscriptionCount: enabledSubscriptionCount(cfg.Subscriptions),
		OutputPath:               cfg.Service.OutputPath,
	})
	return ref
}

func withWorkspace(path, workspaceID string) string {
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + "workspace=" + workspaceID
}

func decodeRefreshResponse(t *testing.T, rec *httptest.ResponseRecorder) refreshResponse {
	t.Helper()
	var body refreshResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	return body
}

func publishedPathFromURL(t *testing.T, rawURL string) string {
	t.Helper()
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		t.Fatalf("parse published url: %v", err)
	}
	return parsed.RequestURI()
}

func publishedTokenFromURL(t *testing.T, rawURL string) string {
	t.Helper()
	path := strings.Trim(publishedPathFromURL(t, rawURL), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[0] != "s" || parts[2] != "mihomo.yaml" {
		t.Fatalf("published url path = %q, want /s/{token}/mihomo.yaml", path)
	}
	return parts[1]
}

func publishedDirCount(t *testing.T, server *Server) int {
	t.Helper()
	entries, err := os.ReadDir(server.publishedRootDir())
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("ReadDir(publishedRootDir) error = %v", err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}
	return count
}

func TestHandleHealthz(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body healthzResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !body.OK {
		t.Fatalf("body.OK = false, want true")
	}
	if body.Version != "0.1.0-test" {
		t.Fatalf("body.Version = %q, want %q", body.Version, "0.1.0-test")
	}
	wantDataDir := filepath.Dir(cfg.Service.StatePath)
	if body.DataDir != wantDataDir {
		t.Fatalf("body.DataDir = %q, want %q", body.DataDir, wantDataDir)
	}
	if body.UptimeSeconds < 0 {
		t.Fatalf("body.UptimeSeconds = %d, want >= 0", body.UptimeSeconds)
	}
}

func TestRandomPublishedIdentifiersAreStrongAndURLSafe(t *testing.T) {
	seenTokens := map[string]struct{}{}
	seenPublishIDs := map[string]struct{}{}
	for i := 0; i < 128; i++ {
		token, err := randomSubscriptionToken()
		if err != nil {
			t.Fatalf("randomSubscriptionToken() error = %v", err)
		}
		if len(token) < 32 {
			t.Fatalf("token length = %d, want at least 32", len(token))
		}
		if strings.Contains(token, "=") || strings.ContainsAny(token, "/+") {
			t.Fatalf("token = %q, want raw URL-safe base64", token)
		}
		if _, ok := seenTokens[token]; ok {
			t.Fatalf("duplicate subscription token generated: %q", token)
		}
		seenTokens[token] = struct{}{}

		publishID, err := randomPublishID()
		if err != nil {
			t.Fatalf("randomPublishID() error = %v", err)
		}
		if !strings.HasPrefix(publishID, "p_") || len(publishID) < 20 {
			t.Fatalf("publish id = %q, want p_ prefix and random suffix", publishID)
		}
		if strings.Contains(publishID, "=") || strings.ContainsAny(publishID, "/+") {
			t.Fatalf("publish id = %q, want raw URL-safe base64", publishID)
		}
		if _, ok := seenPublishIDs[publishID]; ok {
			t.Fatalf("duplicate publish id generated: %q", publishID)
		}
		seenPublishIDs[publishID] = struct{}{}
	}
}

func TestHandleStatus(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Subscriptions = []model.SubscriptionConfig{
		{Name: "enabled", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent},
		{Name: "disabled", Enabled: false, URL: "https://example.com/b", UserAgent: model.DefaultUserAgent},
	}

	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceID := ref.ID
	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/status", workspaceID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body statusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !body.Running {
		t.Fatalf("body.Running = false, want true")
	}
	if body.EnabledSubscriptionCount != 1 {
		t.Fatalf("body.EnabledSubscriptionCount = %d, want 1", body.EnabledSubscriptionCount)
	}
	if body.OutputPath != ref.OutputPath {
		t.Fatalf("body.OutputPath = %q, want %q", body.OutputPath, ref.OutputPath)
	}
	if body.RefreshInterval != cfg.Service.RefreshInterval {
		t.Fatalf("body.RefreshInterval = %d, want %d", body.RefreshInterval, cfg.Service.RefreshInterval)
	}
}

func TestHandleSubscriptionMeta(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-1", Name: "主力机场", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent},
		{ID: "source-2", Name: "备用机场", Enabled: true, URL: "https://example.com/b", UserAgent: model.DefaultUserAgent},
	}

	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceID := ref.ID
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	if err := pipeline.SaveNodeState(workspaceCfg, model.NodeState{
		SubscriptionMeta: map[string]model.SubscriptionMeta{
			"source-1": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-1",
				SourceName: "主力机场",
				Download:   30 * 1024 * 1024 * 1024,
				Total:      200 * 1024 * 1024 * 1024,
				Expire:     1779235200,
				FromHeader: true,
			}),
			"source-2": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-2",
				SourceName: "备用机场",
				Download:   10 * 1024 * 1024 * 1024,
				Total:      100 * 1024 * 1024 * 1024,
				Expire:     1780185600,
				FromHeader: true,
			}),
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/subscription-meta", workspaceID), nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body subscriptionMetaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.OK {
		t.Fatalf("body.OK = false, want true")
	}
	if body.Aggregate.Total != 300*1024*1024*1024 || body.Aggregate.Used != 40*1024*1024*1024 {
		t.Fatalf("body.Aggregate = %#v, want aggregated total/used", body.Aggregate)
	}
	if body.Aggregate.Expire != 1779235200 || body.Aggregate.ExpireSourceName != "主力机场" {
		t.Fatalf("body.Aggregate = %#v, want earliest expire from source-1", body.Aggregate)
	}
	if len(body.Sources) != 2 {
		t.Fatalf("len(body.Sources) = %d, want 2", len(body.Sources))
	}
}

func TestHandleParse(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	body := bytes.NewBufferString(`{"content":"vless://uuid-1@example.com:443?sni=example.com#demo","content_type":"auto"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/parse", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var response parseResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK {
		t.Fatalf("response.OK = false, want true")
	}
	if len(response.Nodes) != 1 || response.Nodes[0].Type != model.ProtocolVLESS {
		t.Fatalf("response.Nodes = %#v, want one vless node", response.Nodes)
	}
}

func TestHandleNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)
	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes", workspaceID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var response nodeListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK {
		t.Fatalf("response.OK = false, want true")
	}
	if len(response.Nodes) != 1 {
		t.Fatalf("len(response.Nodes) = %d, want 1", len(response.Nodes))
	}
	if response.Nodes[0].Source.Name != "manual" || response.Nodes[0].Source.Kind != "inline" {
		t.Fatalf("response.Nodes[0] = %#v, want inline source metadata", response.Nodes[0])
	}
	if response.Summary.Total != 1 || response.Summary.Enabled != 1 {
		t.Fatalf("response.Summary = %#v, want one enabled node", response.Summary)
	}
}

func TestNodeOverrideMaskingAndReset(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "trojan://super-secret@example.com:443?sni=example.com#trojan-node",
		},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceID := ref.ID
	listReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes", workspaceID), nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)

	var listBody nodeListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Nodes) != 1 {
		t.Fatalf("len(listBody.Nodes) = %d, want 1", len(listBody.Nodes))
	}
	nodeID := listBody.Nodes[0].ID

	detailReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes/"+nodeID, workspaceID), nil)
	detailRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(detailRec, detailReq)

	var detailBody nodeDetailResponse
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailBody); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}
	if detailBody.Node.Auth.Password != maskedSecretValue {
		t.Fatalf("detailBody.Node.Auth.Password = %q, want masked secret", detailBody.Node.Auth.Password)
	}

	overrideBody := bytes.NewBufferString(`{
	  "enabled": true,
	  "name": "renamed-node",
	  "region": "JP",
	  "tags": ["jp"],
	  "server": "example.com",
	  "port": 443,
	  "udp": true,
	  "tls": {"enabled": true, "sni": "example.com", "client_fingerprint": "chrome"},
	  "auth": {"password": "********"},
	  "transport": {},
	  "raw": {}
	}`)
	overrideReq := httptest.NewRequest(http.MethodPut, withWorkspace("/api/nodes/"+nodeID+"/override", workspaceID), overrideBody)
	overrideRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(overrideRec, overrideReq)
	if overrideRec.Code != http.StatusOK {
		t.Fatalf("override status code = %d, want %d; body=%s", overrideRec.Code, http.StatusOK, overrideRec.Body.String())
	}

	listRec = httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response after override: %v", err)
	}
	if listBody.Nodes[0].Name != "renamed-node" || !listBody.Nodes[0].Modified {
		t.Fatalf("listBody.Nodes[0] = %#v, want renamed modified node", listBody.Nodes[0])
	}

	resetReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/nodes/"+nodeID+"/reset", workspaceID), nil)
	resetRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(resetRec, resetReq)
	if resetRec.Code != http.StatusOK {
		t.Fatalf("reset status code = %d, want %d; body=%s", resetRec.Code, http.StatusOK, resetRec.Body.String())
	}

	listRec = httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response after reset: %v", err)
	}
	if listBody.Nodes[0].Name != "trojan-node" || listBody.Nodes[0].Modified {
		t.Fatalf("listBody.Nodes[0] after reset = %#v, want original unmodified node", listBody.Nodes[0])
	}
}

func TestNodeDetailMasksWireGuardPeerSecrets(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "wireguard://private-key@example.com:51820?public-key=server-key&pre-shared-key=peer-secret&ip=172.16.0.2/32#wg-node",
		},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	listReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes", workspaceID), nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	var listBody nodeListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Nodes) != 1 {
		t.Fatalf("len(listBody.Nodes) = %d, want 1", len(listBody.Nodes))
	}

	detailReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes/"+listBody.Nodes[0].ID, workspaceID), nil)
	detailRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("detail status code = %d; body=%s", detailRec.Code, detailRec.Body.String())
	}
	if strings.Contains(detailRec.Body.String(), "private-key") || strings.Contains(detailRec.Body.String(), "peer-secret") {
		t.Fatalf("node detail leaked wireguard secret: %s", detailRec.Body.String())
	}
}

func TestDisableNodeAndAddCustomNode(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceID := ref.ID
	listReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes", workspaceID), nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)

	var listBody nodeListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	nodeID := listBody.Nodes[0].ID

	disableReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/nodes/disable", workspaceID), bytes.NewBufferString(`{"ids":["`+nodeID+`"]}`))
	disableRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(disableRec, disableReq)
	if disableRec.Code != http.StatusOK {
		t.Fatalf("disable status code = %d, want %d; body=%s", disableRec.Code, http.StatusOK, disableRec.Body.String())
	}

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusBadRequest {
		t.Fatalf("refresh status code = %d, want %d when all nodes disabled", refreshRec.Code, http.StatusBadRequest)
	}

	customReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/nodes/custom", workspaceID), bytes.NewBufferString(`{"content":"vless://uuid-1@example.com:443?sni=example.com#custom-node","content_type":"uri"}`))
	customRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(customRec, customReq)
	if customRec.Code != http.StatusOK {
		t.Fatalf("custom status code = %d, want %d; body=%s", customRec.Code, http.StatusOK, customRec.Body.String())
	}

	enableReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/nodes/enable", workspaceID), bytes.NewBufferString(`{"ids":["`+nodeID+`"]}`))
	enableRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(enableRec, enableReq)
	if enableRec.Code != http.StatusOK {
		t.Fatalf("enable status code = %d, want %d; body=%s", enableRec.Code, http.StatusOK, enableRec.Body.String())
	}

	refreshRec = httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d, want %d; body=%s", refreshRec.Code, http.StatusOK, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	data, err := os.ReadFile(refreshBody.OutputPath)
	if err != nil {
		t.Fatalf("ReadFile(output) error = %v", err)
	}
	if !bytes.Contains(data, []byte("custom-node")) {
		t.Fatalf("rendered yaml = %q, want custom node included", string(data))
	}
}

func TestHandleGenerate(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	body := bytes.NewBufferString(`{
  "template":"lite",
  "nodes":[
    {
      "name":"ss-node",
      "type":"ss",
      "server":"example.com",
      "port":443,
      "auth":{"password":"pass"},
      "raw":{"method":"aes-256-gcm"},
      "udp":true
    }
  ]
}`)
	req := httptest.NewRequest(http.MethodPost, "/api/generate", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var response generateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK {
		t.Fatalf("response.OK = false, want true")
	}
	if !bytes.Contains([]byte(response.YAML), []byte("type: ss")) {
		t.Fatalf("response.YAML = %q, want ss proxy", response.YAML)
	}
}

func TestHandleRefresh(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)
	req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var response refreshResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || response.NodeCount != 1 {
		t.Fatalf("response = %#v, want ok with one node", response)
	}
}

func TestFixedSubscriptionPathDisabled(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestFixedSubscriptionPathDownloadDisabled(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml?download=1", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); got != "" {
		t.Fatalf("Content-Disposition = %q, want empty", got)
	}
}

func TestFixedSubscriptionPathIgnoresLegacyToken(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.AccessToken = "secret-token"
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#ss-node",
		},
	}
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml?token=secret-token", nil)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestHandleAudit(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	if err := pipeline.SaveNodeState(cfg, model.NodeState{
		LastAudit: model.AuditReport{
			RawCount:      10,
			FinalCount:    4,
			ExcludedCount: 6,
			ExcludedNodes: []model.ExcludedNode{{Name: "剩余流量：100GB", Reason: "info_node", Source: model.SourceInfo{Name: "SecOne"}}},
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	if err := pipeline.SaveNodeState(workspaceCfg, model.NodeState{
		LastAudit: model.AuditReport{
			RawCount:      10,
			FinalCount:    4,
			ExcludedCount: 6,
			ExcludedNodes: []model.ExcludedNode{{Name: "剩余流量：100GB", Reason: "info_node", Source: model.SourceInfo{Name: "SecOne"}}},
		},
	}); err != nil {
		t.Fatalf("SaveNodeState(workspace) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/audit", ref.ID), nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["raw_count"].(float64) != 10 || body["final_count"].(float64) != 4 {
		t.Fatalf("body = %#v, want audit counts", body)
	}
}

func TestFixedSubscriptionPathDoesNotExposeUserinfoHeader(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	cfg.Service.OutputPath = filepath.Join(dir, "mihomo.yaml")
	cfg.Service.RefreshOnRequest = false
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-1", Name: "主力机场", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent},
		{ID: "source-2", Name: "备用机场", Enabled: true, URL: "https://example.com/b", UserAgent: model.DefaultUserAgent},
	}
	if err := os.WriteFile(cfg.Service.OutputPath, []byte("mixed-port: 7897\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := pipeline.SaveNodeState(cfg, model.NodeState{
		SubscriptionMeta: map[string]model.SubscriptionMeta{
			"source-1": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-1",
				SourceName: "主力机场",
				Upload:     0,
				Download:   30 * 1024 * 1024 * 1024,
				Total:      200 * 1024 * 1024 * 1024,
				Expire:     1779235200,
				FromHeader: true,
			}),
			"source-2": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-2",
				SourceName: "备用机场",
				Upload:     0,
				Download:   10 * 1024 * 1024 * 1024,
				Total:      100 * 1024 * 1024 * 1024,
				Expire:     1780185600,
				FromHeader: true,
			}),
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	server, cfg := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if got := rec.Header().Get("Subscription-Userinfo"); got != "" {
		t.Fatalf("Subscription-Userinfo = %q, want empty", got)
	}
}

func TestFixedSubscriptionPathOmittedWithMergeStrategyNone(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	cfg.Service.OutputPath = filepath.Join(dir, "mihomo.yaml")
	cfg.Service.RefreshOnRequest = false
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-1", Name: "主力机场", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent},
	}
	cfg.Render.SubscriptionInfo = &model.SubscriptionInfoConfig{
		Enabled:        true,
		ExposeHeader:   true,
		ShowPerSource:  true,
		MergeStrategy:  "none",
		ExpireStrategy: "earliest",
	}
	if err := os.WriteFile(cfg.Service.OutputPath, []byte("mixed-port: 7897\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := pipeline.SaveNodeState(cfg, model.NodeState{
		SubscriptionMeta: map[string]model.SubscriptionMeta{
			"source-1": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-1",
				SourceName: "主力机场",
				Download:   30 * 1024 * 1024 * 1024,
				Total:      200 * 1024 * 1024 * 1024,
				Expire:     1779235200,
				FromHeader: true,
			}),
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	server, cfg := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if got := rec.Header().Get("Subscription-Userinfo"); got != "" {
		t.Fatalf("Subscription-Userinfo = %q, want empty when merge_strategy=none", got)
	}
}

func TestHandleLogs(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)
	server.appendLog("first")
	server.appendLog("second")
	ref := createWorkspaceRefForTest(t, server, cfg)
	server.appendWorkspaceLog(ref.Hash, "first")
	server.appendWorkspaceLog(ref.Hash, "second")
	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/logs?tail=1", ref.ID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var response logsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK {
		t.Fatalf("response.OK = false, want true")
	}
	if len(response.Lines) != 1 || !bytes.Contains([]byte(response.Lines[0]), []byte("second")) {
		t.Fatalf("response.Lines = %#v, want last log line", response.Lines)
	}
}

func TestFixedSubscriptionPathDoesNotServeCachedYAML(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.RefreshInterval = 3600
	cfg.Service.RefreshOnRequest = true
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczJAZXhhbXBsZS5jb206NDQz#fresh"},
	}
	server, cfg := newTestServer(t, cfg)

	oldYAML := []byte("old: yaml\n")
	if err := os.WriteFile(cfg.Service.OutputPath, oldYAML, 0o644); err != nil {
		t.Fatalf("WriteFile(output) error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if bytes.Equal(rec.Body.Bytes(), oldYAML) {
		t.Fatalf("fixed subscription path served cached yaml")
	}
}

func TestFixedSubscriptionPathDoesNotRefreshExpiredYAML(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.RefreshInterval = 1
	cfg.Service.RefreshOnRequest = true
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczNAZXhhbXBsZS5jb206NDQz#fresh-node"},
	}
	server, cfg := newTestServer(t, cfg)

	if err := os.WriteFile(cfg.Service.OutputPath, []byte("old: yaml\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(output) error = %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(cfg.Service.OutputPath, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes(output) error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("fresh-node")) {
		t.Fatalf("fixed subscription path refreshed yaml: %q", rec.Body.Bytes())
	}
}

func TestFixedSubscriptionPathDoesNotServeStaleYAML(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.RefreshInterval = 1
	cfg.Service.RefreshOnRequest = true
	cfg.Service.StaleIfError = true
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-1", Name: "broken", Enabled: true, URL: "https://127.0.0.1/broken", UserAgent: model.DefaultUserAgent},
	}
	server, cfg := newTestServer(t, cfg)

	oldYAML := []byte("old: stale\n")
	if err := os.WriteFile(cfg.Service.OutputPath, oldYAML, 0o644); err != nil {
		t.Fatalf("WriteFile(output) error = %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(cfg.Service.OutputPath, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes(output) error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if bytes.Equal(rec.Body.Bytes(), oldYAML) {
		t.Fatalf("fixed subscription path served stale yaml")
	}
}

func TestFixedSubscriptionPathNoStaleAvailable(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.RefreshInterval = 1
	cfg.Service.RefreshOnRequest = true
	cfg.Service.StaleIfError = true
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-1", Name: "broken", Enabled: true, URL: "https://127.0.0.1/broken", UserAgent: model.DefaultUserAgent},
	}
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestStrictModePreventsWritingBadYAML(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	cfg.Service.OutputPath = filepath.Join(dir, "mihomo.yaml")
	cfg.Service.StrictMode = true
	cfg.Render.IncludeInfoNode = true
	cfg.Render.ShowInfoNodes = true
	cfg.Render.ShowNodeType = false
	cfg.Render.Emoji = false
	cfg.Render.SourcePrefix = false
	cfg.Render.CustomProxyGroups = []model.CustomProxyGroupConfig{
		{
			Name:     "bad-fallback",
			Type:     "fallback",
			Members:  []string{"剩余流量：100GB"},
			URL:      "https://www.gstatic.com/generate_204",
			Interval: 300,
			Enabled:  true,
		},
	}
	if err := pipeline.SaveNodeState(cfg, model.NodeState{
		CustomNodes: []model.NodeIR{
			model.NormalizeNode(model.NodeIR{
				Name:   "剩余流量：100GB",
				Type:   model.ProtocolSS,
				Server: "info.example.com",
				Port:   443,
				Auth:   model.Auth{Password: "p"},
				Raw:    map[string]interface{}{"method": "aes-256-gcm", "_infoNode": true},
				Source: model.SourceInfo{Name: "manual", Kind: "custom"},
			}),
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}
	oldYAML := []byte("old: yaml\n")
	if err := os.WriteFile(cfg.Service.OutputPath, oldYAML, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	if err := pipeline.SaveNodeState(workspaceCfg, model.NodeState{
		CustomNodes: []model.NodeIR{
			model.NormalizeNode(model.NodeIR{
				Name:   "剩余流量：100GB",
				Type:   model.ProtocolSS,
				Server: "info.example.com",
				Port:   443,
				Auth:   model.Auth{Password: "p"},
				Raw:    map[string]interface{}{"method": "aes-256-gcm", "_infoNode": true},
				Source: model.SourceInfo{Name: "manual", Kind: "custom"},
			}),
		},
	}); err != nil {
		t.Fatalf("SaveNodeState(workspace) error = %v", err)
	}
	published, _, err := server.ensureWorkspacePublishedRef(&ref)
	if err != nil {
		t.Fatalf("ensureWorkspacePublishedRef() error = %v", err)
	}
	if err := os.WriteFile(published.CurrentPath, oldYAML, 0o644); err != nil {
		t.Fatalf("WriteFile(published output) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", ref.ID), nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	got, err := os.ReadFile(published.CurrentPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(got, oldYAML) {
		t.Fatalf("output yaml was overwritten:\n%s", string(got))
	}
}

func TestRefreshLock(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczRAZXhhbXBsZS5jb206NDQz#lock"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	started := make(chan struct{})
	release := make(chan struct{})
	server.refreshBeforeRun = func() {
		select {
		case <-started:
		default:
			close(started)
		}
		<-release
	}

	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
	}()

	<-started

	req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "REFRESH_IN_PROGRESS") {
		t.Fatalf("body = %s, want REFRESH_IN_PROGRESS", rec.Body.String())
	}

	close(release)
	<-firstDone
}

func TestOverridesSurviveRefresh(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczVAZXhhbXBsZS5jb206NDQz#original"},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceID := ref.ID
	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("initial refresh status code = %d", refreshRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes", workspaceID), nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	var listBody nodeListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	nodeID := listBody.Nodes[0].ID

	overrideBody := bytes.NewBufferString(`{"enabled":true,"name":"renamed-node"}`)
	overrideReq := httptest.NewRequest(http.MethodPut, withWorkspace("/api/nodes/"+nodeID+"/override", workspaceID), overrideBody)
	overrideRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(overrideRec, overrideReq)
	if overrideRec.Code != http.StatusOK {
		t.Fatalf("override status code = %d; body=%s", overrideRec.Code, overrideRec.Body.String())
	}

	refreshRec = httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("second refresh status code = %d", refreshRec.Code)
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	data, err := os.ReadFile(refreshBody.OutputPath)
	if err != nil {
		t.Fatalf("ReadFile(output) error = %v", err)
	}
	if !bytes.Contains(data, []byte("renamed-node")) {
		t.Fatalf("rendered yaml = %q, want override name to survive refresh", string(data))
	}
}

func TestMaskSensitiveTextMasksSubscriptionToken(t *testing.T) {
	value := "fetch subscription 主力机场 https://example.com/sub?token=abc123&x=1"
	masked := maskSensitiveText(value)
	if strings.Contains(masked, "abc123") {
		t.Fatalf("maskSensitiveText() leaked token: %q", masked)
	}
	if !strings.Contains(masked, "token=***") {
		t.Fatalf("maskSensitiveText() = %q, want masked token", masked)
	}
}

func TestMaskSensitiveTextMasksKeyValueSecrets(t *testing.T) {
	value := "password: p1 uuid=00000000-0000-0000-0000-000000000000 private-key=client-key Authorization: Bearer abc Cookie: sid=def"
	masked := maskSensitiveText(value)
	for _, secret := range []string{"p1", "00000000-0000-0000-0000-000000000000", "client-key", "Bearer abc", "sid=def"} {
		if strings.Contains(masked, secret) {
			t.Fatalf("maskSensitiveText() leaked %q in %q", secret, masked)
		}
	}
}

func TestMaskSensitiveTextMasksPublishedURLAndNodeSecrets(t *testing.T) {
	value := strings.Join([]string{
		"subscription http://127.0.0.1:9876/s/SDCWuCZAorYdAi-87nrSgNu9oGMqILT2/mihomo.yaml",
		"upstream https://example.com/sub?password=p1&uuid=00000000-0000-0000-0000-000000000000&private-key=client-key",
		"node ss://secret@example.com:443#demo",
	}, " ")
	masked := maskSensitiveText(value)
	for _, secret := range []string{
		"SDCWuCZAorYdAi-87nrSgNu9oGMqILT2",
		"p1",
		"00000000-0000-0000-0000-000000000000",
		"client-key",
		"secret@example.com",
	} {
		if strings.Contains(masked, secret) {
			t.Fatalf("maskSensitiveText() leaked %q in %q", secret, masked)
		}
	}
	if !strings.Contains(masked, "/s/<redacted>/mihomo.yaml") || !strings.Contains(masked, "ss://***@") {
		t.Fatalf("maskSensitiveText() = %q, want redacted published URL and node URI", masked)
	}
}

func TestHandleConfigGetAndPut(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.AccessToken = "config-token"
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-1", Name: "secret-source", Enabled: true, URL: "https://example.com/sub?token=abc123&user=bob", UserAgent: model.DefaultUserAgent},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)

	getReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/config", ref.ID), nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status code = %d, want %d", getRec.Code, http.StatusOK)
	}
	if strings.Contains(getRec.Body.String(), "abc123") || strings.Contains(getRec.Body.String(), "config-token") {
		t.Fatalf("GET /api/config leaked secret: %s", getRec.Body.String())
	}
	if !strings.Contains(getRec.Body.String(), "token=%2A%2A%2A") {
		t.Fatalf("GET /api/config did not redact subscription URL query: %s", getRec.Body.String())
	}

	putBody := bytes.NewBufferString(`{
  "service":{
    "enabled":true,
    "listen_addr":"0.0.0.0",
    "listen_port":9876,
    "log_level":"info",
    "template":"standard",
    "output_path":"/data/mihomo.yaml",
    "cache_dir":"/data/cache",
    "state_path":"/data/state.json",
    "refresh_interval":3600,
    "max_subscription_bytes":5242880,
    "fetch_timeout_seconds":15,
    "allow_lan":false
  },
  "subscriptions":[],
  "inline":[
    {"name":"manual","enabled":true,"content":"ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node"}
  ],
  "render":{
    "mixed_port":7890,
    "allow_lan":false,
    "mode":"rule",
    "log_level":"info",
    "ipv6":false,
    "dns_enabled":true,
    "enhanced_mode":"fake-ip"
  }
}`)
	putReq := httptest.NewRequest(http.MethodPut, withWorkspace("/api/config", ref.ID), putBody)
	putRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT status code = %d, want %d; body=%s", putRec.Code, http.StatusOK, putRec.Body.String())
	}

	data, err := os.ReadFile(ref.ConfigPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	if !bytes.Contains(data, []byte(`"name": "manual"`)) {
		t.Fatalf("written config = %q, want inline source", string(data))
	}
}

func TestConfigRequiresWorkspace(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestWorkspaceCreateAndDelete(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var createBody workspaceResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &createBody); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createBody.WorkspaceID == "" || len(createBody.WorkspaceID) < 20 {
		t.Fatalf("workspace_id = %q, want sufficiently random id", createBody.WorkspaceID)
	}

	getReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/config", createBody.WorkspaceID), nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("workspace config status = %d, want %d; body=%s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/workspaces/"+createBody.WorkspaceID, nil)
	deleteRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status code = %d, want %d; body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
	}

	getRec = httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want %d; body=%s", getRec.Code, http.StatusNotFound, getRec.Body.String())
	}
}

func TestWorkspaceRandom(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)
	refA := createWorkspaceRefForTest(t, server, cfg)
	refB := createWorkspaceRefForTest(t, server, cfg)
	if refA.ID == refB.ID {
		t.Fatalf("workspace ids should differ: %q", refA.ID)
	}
	if len(refA.ID) < 20 || len(refB.ID) < 20 {
		t.Fatalf("workspace ids too short: %q %q", refA.ID, refB.ID)
	}
}

func TestRootDoesNotLoadPreviousConfig(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "SecOne") || strings.Contains(body, "https://example.com/sub") {
		t.Fatalf("root html leaked previous config: %s", body)
	}
}

func TestWorkspaceIsolation(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	refA := createWorkspaceRefForTest(t, server, cfg)
	cfgA, err := server.loadWorkspaceConfig(refA)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig(A) error = %v", err)
	}
	cfgA.Subscriptions = []model.SubscriptionConfig{{ID: "source-1", Name: "SecOne", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent}}
	if err := config.WriteJSON(refA.ConfigPath, cfgA); err != nil {
		t.Fatalf("WriteJSON(A) error = %v", err)
	}

	refB := createWorkspaceRefForTest(t, server, cfg)
	cfgB, err := server.loadWorkspaceConfig(refB)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig(B) error = %v", err)
	}
	cfgB.Subscriptions = []model.SubscriptionConfig{{ID: "source-1", Name: "Other", Enabled: true, URL: "https://example.com/b", UserAgent: model.DefaultUserAgent}}
	if err := config.WriteJSON(refB.ConfigPath, cfgB); err != nil {
		t.Fatalf("WriteJSON(B) error = %v", err)
	}

	reqA := httptest.NewRequest(http.MethodGet, withWorkspace("/api/config", refA.ID), nil)
	recA := httptest.NewRecorder()
	server.Handler().ServeHTTP(recA, reqA)
	reqB := httptest.NewRequest(http.MethodGet, withWorkspace("/api/config", refB.ID), nil)
	recB := httptest.NewRecorder()
	server.Handler().ServeHTTP(recB, reqB)

	if !strings.Contains(recA.Body.String(), "SecOne") || strings.Contains(recA.Body.String(), "Other") {
		t.Fatalf("workspace A body invalid: %s", recA.Body.String())
	}
	if !strings.Contains(recB.Body.String(), "Other") || strings.Contains(recB.Body.String(), "SecOne") {
		t.Fatalf("workspace B body invalid: %s", recB.Body.String())
	}
}

func TestSubscriptionTokenCannotAccessConfig(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#ss-node"},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", ref.ID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	var body refreshResponse
	if err := json.Unmarshal(refreshRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	parts := strings.Split(body.SubscriptionURL, "/s/")
	if len(parts) != 2 {
		t.Fatalf("subscription url = %q, want /s/{token}/mihomo.yaml", body.SubscriptionURL)
	}
	token := strings.TrimSuffix(parts[1], "/mihomo.yaml")

	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/config", token), nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPublishedURLUsesConfiguredPublicBaseURL(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Service.PublicBaseURL = "https://subconv.example.com/base"
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#public-base-url"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshReq.Host = "internal.local:9876"
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	if !strings.HasPrefix(refreshBody.SubscriptionURL, "https://subconv.example.com/base/s/") {
		t.Fatalf("SubscriptionURL = %q, want configured public base URL", refreshBody.SubscriptionURL)
	}

	statusReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/published", workspaceID), nil)
	statusReq.Host = "internal.local:9876"
	statusRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status code = %d; body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusBody publishedStatusResponse
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode published status: %v", err)
	}
	if statusBody.SubscriptionURL != refreshBody.SubscriptionURL {
		t.Fatalf("published status URL = %q, want %q", statusBody.SubscriptionURL, refreshBody.SubscriptionURL)
	}
}

func TestPublishedSubscriptionStillWorksAfterWorkspaceDeleted(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#ss-node"},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", ref.ID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	var body refreshResponse
	if err := json.Unmarshal(refreshRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	subPath := strings.TrimPrefix(body.SubscriptionURL, "http://"+refreshReq.Host)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/workspaces/"+ref.ID, nil)
	deleteRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status code = %d; body=%s", deleteRec.Code, deleteRec.Body.String())
	}

	subReq := httptest.NewRequest(http.MethodGet, subPath, nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("subscription status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if !bytes.Contains(subRec.Body.Bytes(), []byte("ss-node")) {
		t.Fatalf("subscription body = %q, want published yaml", subRec.Body.String())
	}
}

func TestPublishedSubscriptionPersistsAfterServerRestart(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#persisted-link"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)

	restarted := NewServer("0.1.0-test", cfg)
	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	restarted.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("subscription after restart status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if !bytes.Contains(subRec.Body.Bytes(), []byte("persisted-link")) {
		t.Fatalf("subscription body = %q, want persisted published yaml", subRec.Body.String())
	}
}

func TestRefreshOverwritesCurrentYAMLAndKeepsSameToken(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#stable"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	var first refreshResponse
	for index := 0; index < 10; index++ {
		req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("refresh #%d status code = %d; body=%s", index+1, rec.Code, rec.Body.String())
		}
		body := decodeRefreshResponse(t, rec)
		if index == 0 {
			first = body
			continue
		}
		if body.SubscriptionURL != first.SubscriptionURL {
			t.Fatalf("subscription url changed on refresh #%d: %q != %q", index+1, body.SubscriptionURL, first.SubscriptionURL)
		}
		if body.PublishID != first.PublishID {
			t.Fatalf("publish id changed on refresh #%d: %q != %q", index+1, body.PublishID, first.PublishID)
		}
	}

	if publishedDirCount(t, server) != 1 {
		t.Fatalf("published dir count = %d, want 1", publishedDirCount(t, server))
	}
	entries, err := os.ReadDir(filepath.Join(server.publishedRootDir(), first.PublishID))
	if err != nil {
		t.Fatalf("ReadDir(published) error = %v", err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 2 || !containsString(names, "current.yaml") || !containsString(names, "meta.json") {
		t.Fatalf("published entries = %#v, want only current.yaml and meta.json", names)
	}
}

func TestRefreshPostWriteValidationFailureDoesNotPublishCurrentYAML(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Render.SourcePrefix = false
	cfg.Render.NameOptions.SourcePrefixMode = "none"
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#post-write-invalid"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	server.refreshAfterWrite = func(path string) {
		_ = os.WriteFile(path, []byte(`
proxies:
  - {name: "post-write-invalid", type: ss, server: example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "post-write-invalid", "missing-node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["post-write-invalid"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🚀 节点选择
`), 0o644)
		server.refreshAfterWrite = nil
	}

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusBadRequest {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	if !strings.Contains(refreshRec.Body.String(), "missing-node") {
		t.Fatalf("refresh body = %s, want validation reason", refreshRec.Body.String())
	}

	ref, err := server.loadWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("loadWorkspace() error = %v", err)
	}
	if ref.Meta.PublishID != "" {
		t.Fatalf("workspace PublishID = %q, want release of failed published ref", ref.Meta.PublishID)
	}
	if count := publishedDirCount(t, server); count != 0 {
		t.Fatalf("published dir count = %d, want 0 after failed publish", count)
	}
	if matches, err := filepath.Glob(filepath.Join(server.publishedRootDir(), "*", "*current.yaml*")); err != nil {
		t.Fatalf("Glob(current.yaml) error = %v", err)
	} else if len(matches) != 0 {
		t.Fatalf("published yaml files = %v, want none after failed validation", matches)
	}
}

func TestRotateTokenInvalidatesOldLink(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#rotate"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	oldURL := refreshBody.SubscriptionURL

	rotateReq := httptest.NewRequest(http.MethodPost, "/api/published/"+refreshBody.PublishID+"/rotate-token", nil)
	rotateRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rotateRec, rotateReq)
	if rotateRec.Code != http.StatusOK {
		t.Fatalf("rotate status code = %d; body=%s", rotateRec.Code, rotateRec.Body.String())
	}
	var rotateBody publishedStatusResponse
	if err := json.Unmarshal(rotateRec.Body.Bytes(), &rotateBody); err != nil {
		t.Fatalf("decode rotate response: %v", err)
	}
	if rotateBody.URL == oldURL {
		t.Fatalf("rotateBody.URL = %q, want new url", rotateBody.URL)
	}
	if publishedDirCount(t, server) != 1 {
		t.Fatalf("published dir count = %d, want 1", publishedDirCount(t, server))
	}

	oldSubReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, oldURL), nil)
	oldSubRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(oldSubRec, oldSubReq)
	if oldSubRec.Code != http.StatusNotFound {
		t.Fatalf("old subscription status code = %d, want %d; body=%s", oldSubRec.Code, http.StatusNotFound, oldSubRec.Body.String())
	}

	newSubReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, rotateBody.URL), nil)
	newSubRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(newSubRec, newSubReq)
	if newSubRec.Code != http.StatusOK {
		t.Fatalf("new subscription status code = %d, want %d; body=%s", newSubRec.Code, http.StatusOK, newSubRec.Body.String())
	}
}

func TestPublishedSubscriptionReturnsUserinfoHeaders(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#userinfo"},
	}
	cfg.Render.SubscriptionInfo = &model.SubscriptionInfoConfig{
		Enabled:        true,
		ExposeHeader:   false,
		ShowPerSource:  true,
		MergeStrategy:  "none",
		ExpireStrategy: "earliest",
	}
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-a", Name: "A", Enabled: false, URL: "https://example.com/a"},
		{ID: "source-b", Name: "B", Enabled: false, URL: "https://example.net/b"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	workspaceRef, err := server.loadWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("loadWorkspace() error = %v", err)
	}
	workspaceCfg, err := server.loadWorkspaceConfig(workspaceRef)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	published, err := server.loadPublishedByID(refreshBody.PublishID)
	if err != nil {
		t.Fatalf("loadPublishedByID() error = %v", err)
	}
	published.Meta.SubscriptionInfo, published.Meta.SourceUserinfo = buildPublishedSubscriptionUserinfo(workspaceCfg, map[string]model.SubscriptionMeta{
		"source-a": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
			SourceID:   "source-a",
			SourceName: "A",
			Upload:     10,
			Download:   20,
			Total:      100,
			Expire:     2000,
			FromHeader: true,
		}),
		"source-b": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
			SourceID:   "source-b",
			SourceName: "B",
			Upload:     30,
			Download:   40,
			Total:      300,
			Expire:     1000,
			FromHeader: true,
		}),
	}, time.Now().UTC())
	if err := server.savePublishedMeta(published); err != nil {
		t.Fatalf("savePublishedMeta() error = %v", err)
	}

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("subscription status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if got := subRec.Header().Get("Content-Type"); got != "text/yaml; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/yaml", got)
	}
	if got := subRec.Header().Get("Content-Disposition"); got != `attachment; filename="mihomo.yaml"` {
		t.Fatalf("Content-Disposition = %q, want attachment filename", got)
	}
	viewReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL)+"?view=1", nil)
	viewRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(viewRec, viewReq)
	if viewRec.Code != http.StatusOK {
		t.Fatalf("view subscription status code = %d; body=%s", viewRec.Code, viewRec.Body.String())
	}
	if got := viewRec.Header().Get("Content-Disposition"); got != `inline; filename="mihomo.yaml"` {
		t.Fatalf("view Content-Disposition = %q, want inline filename", got)
	}
	if got := subRec.Header().Get("Profile-Update-Interval"); got != "24" {
		t.Fatalf("Profile-Update-Interval = %q, want 24", got)
	}
	userinfo := subRec.Header().Get("Subscription-Userinfo")
	if userinfo != "upload=40; download=60; total=400; expire=1000" {
		t.Fatalf("Subscription-Userinfo = %q, want aggregated bytes", userinfo)
	}
	if regexp.MustCompile(`(?i)(gb|mb|kb|tb|年|月|日|-)`).MatchString(userinfo) {
		t.Fatalf("Subscription-Userinfo = %q, want byte numbers only", userinfo)
	}

	statusReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/published", workspaceID), nil)
	statusRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("published status code = %d; body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusBody publishedStatusResponse
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode published status: %v", err)
	}
	if !statusBody.HasSubscriptionUserinfo || statusBody.SubscriptionInfoHeader != userinfo {
		t.Fatalf("published status = %#v, want subscription userinfo status", statusBody)
	}

	metaBytes, err := os.ReadFile(filepath.Join(server.publishedRootDir(), refreshBody.PublishID, "meta.json"))
	if err != nil {
		t.Fatalf("ReadFile(meta.json) error = %v", err)
	}
	if bytes.Contains(metaBytes, []byte("/s/")) {
		t.Fatalf("published meta leaked full subscription URL: %s", string(metaBytes))
	}
	if !bytes.Contains(metaBytes, []byte(`"subscription_userinfo"`)) || !bytes.Contains(metaBytes, []byte(`"source_userinfo"`)) {
		t.Fatalf("published meta = %s, want subscription info persisted", string(metaBytes))
	}

	restarted := NewServer("0.1.0-test", cfg)
	headReq := httptest.NewRequest(http.MethodHead, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	headRec := httptest.NewRecorder()
	restarted.Handler().ServeHTTP(headRec, headReq)
	if headRec.Code != http.StatusOK {
		t.Fatalf("HEAD subscription status code = %d; body=%s", headRec.Code, headRec.Body.String())
	}
	if got := headRec.Header().Get("Subscription-Userinfo"); got != userinfo {
		t.Fatalf("HEAD Subscription-Userinfo = %q, want %q after restart", got, userinfo)
	}
	if got := headRec.Header().Get("Profile-Update-Interval"); got != "24" {
		t.Fatalf("HEAD Profile-Update-Interval = %q, want 24", got)
	}
	if got := headRec.Header().Get("Content-Disposition"); got != `attachment; filename="mihomo.yaml"` {
		t.Fatalf("HEAD Content-Disposition = %q, want attachment filename", got)
	}
	if headRec.Body.Len() != 0 {
		t.Fatalf("HEAD body length = %d, want 0", headRec.Body.Len())
	}

	logs := restarted.snapshotLogs(20)
	logText := strings.Join(logs, "\n")
	if !strings.Contains(logText, "serve subscription publish="+refreshBody.PublishID) ||
		!strings.Contains(logText, "userinfo=present") ||
		!strings.Contains(logText, `header="upload=40; download=60; total=400; expire=1000"`) {
		t.Fatalf("logs = %q, want published serve userinfo log", logText)
	}
	if strings.Contains(logText, publishedTokenFromURL(t, refreshBody.SubscriptionURL)) {
		t.Fatalf("logs leaked full token: %q", logText)
	}
}

func TestPublishedSubscriptionOmitsUserinfoHeaderWhenUnavailable(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#no-info"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("subscription status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if got := subRec.Header().Get("Subscription-Userinfo"); got != "" {
		t.Fatalf("Subscription-Userinfo = %q, want empty without upstream info", got)
	}
	if got := subRec.Header().Get("Profile-Update-Interval"); got != "24" {
		t.Fatalf("Profile-Update-Interval = %q, want 24", got)
	}
	if got := subRec.Header().Get("Content-Disposition"); got != `attachment; filename="mihomo.yaml"` {
		t.Fatalf("Content-Disposition = %q, want attachment filename", got)
	}
}

func TestPublishedSubscriptionRestoresMissingUserinfoFromWorkspaceState(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#restore-userinfo"},
	}
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-a", Name: "A", Enabled: false, URL: "https://example.com/a"},
		{ID: "source-b", Name: "B", Enabled: false, URL: "https://example.net/b"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	ref, err := server.loadWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("loadWorkspace() error = %v", err)
	}
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	if err := pipeline.SaveNodeState(workspaceCfg, model.NodeState{
		SubscriptionMeta: map[string]model.SubscriptionMeta{
			"source-a": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-a",
				SourceName: "A",
				Upload:     11,
				Download:   22,
				Total:      111,
				Expire:     4000,
				FromHeader: true,
			}),
			"source-b": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
				SourceID:   "source-b",
				SourceName: "B",
				Upload:     33,
				Download:   44,
				Total:      333,
				Expire:     3000,
				FromHeader: true,
			}),
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}
	published, err := server.loadPublishedByID(refreshBody.PublishID)
	if err != nil {
		t.Fatalf("loadPublishedByID() error = %v", err)
	}
	published.Meta.SubscriptionInfo = nil
	published.Meta.SourceUserinfo = nil
	if err := server.savePublishedMeta(published); err != nil {
		t.Fatalf("savePublishedMeta() error = %v", err)
	}

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("subscription status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if got := subRec.Header().Get("Subscription-Userinfo"); got != "upload=44; download=66; total=444; expire=3000" {
		t.Fatalf("Subscription-Userinfo = %q, want restored aggregate", got)
	}

	metaBytes, err := os.ReadFile(filepath.Join(server.publishedRootDir(), refreshBody.PublishID, "meta.json"))
	if err != nil {
		t.Fatalf("ReadFile(meta.json) error = %v", err)
	}
	if !bytes.Contains(metaBytes, []byte(`"subscription_userinfo"`)) || !bytes.Contains(metaBytes, []byte(`"total": 444`)) {
		t.Fatalf("published meta = %s, want restored subscription_userinfo persisted", string(metaBytes))
	}
}

func TestPublishedSubscriptionDoesNotSynthesizeUserinfoFromInfoNode(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#info-node"},
	}
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-a", Name: "A", Enabled: false, URL: "https://example.com/a"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)
	workspaceRef, err := server.loadWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("loadWorkspace() error = %v", err)
	}
	workspaceCfg, err := server.loadWorkspaceConfig(workspaceRef)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	published, err := server.loadPublishedByID(refreshBody.PublishID)
	if err != nil {
		t.Fatalf("loadPublishedByID() error = %v", err)
	}
	published.Meta.SubscriptionInfo, published.Meta.SourceUserinfo = buildPublishedSubscriptionUserinfo(workspaceCfg, map[string]model.SubscriptionMeta{
		"source-a": model.NormalizeSubscriptionMeta(model.SubscriptionMeta{
			SourceID:     "source-a",
			SourceName:   "A",
			Download:     20,
			Total:        100,
			Expire:       2000,
			FromInfoNode: true,
		}),
	}, time.Now().UTC())
	if err := server.savePublishedMeta(published); err != nil {
		t.Fatalf("savePublishedMeta() error = %v", err)
	}

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("subscription status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if got := subRec.Header().Get("Subscription-Userinfo"); got != "" {
		t.Fatalf("Subscription-Userinfo = %q, want empty without upstream header", got)
	}
}

func TestGetPublishedByIDAndBindWorkspace(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#draft-restore"},
	}
	server, cfg := newTestServer(t, cfg)
	originalWorkspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", originalWorkspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)

	getReq := httptest.NewRequest(http.MethodGet, "/api/published/"+refreshBody.PublishID, nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get published status code = %d; body=%s", getRec.Code, getRec.Body.String())
	}
	var getBody publishedStatusResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode get published response: %v", err)
	}
	if getBody.PublishID != refreshBody.PublishID {
		t.Fatalf("PublishID = %q, want %q", getBody.PublishID, refreshBody.PublishID)
	}
	if getBody.SubscriptionURL == "" || getBody.SubscriptionURL != getBody.URL {
		t.Fatalf("SubscriptionURL = %q URL = %q, want matching non-empty urls", getBody.SubscriptionURL, getBody.URL)
	}
	if getBody.Status != "active" {
		t.Fatalf("Status = %q, want active", getBody.Status)
	}
	token := publishedTokenFromURL(t, getBody.SubscriptionURL)
	if getBody.TokenHint == "" || strings.Contains(getBody.TokenHint, token) {
		t.Fatalf("TokenHint = %q, token = %q, want hint without full token", getBody.TokenHint, token)
	}
	if strings.Contains(getRec.Body.String(), `"token"`) {
		t.Fatalf("published status leaked raw token field: %s", getRec.Body.String())
	}

	restoredWorkspaceID := createWorkspaceForTest(t, server, cfg)
	bindBody := bytes.NewBufferString(`{"publish_id":"` + refreshBody.PublishID + `"}`)
	bindReq := httptest.NewRequest(http.MethodPost, "/api/workspaces/"+restoredWorkspaceID+"/bind-publish", bindBody)
	bindReq.Header.Set("Content-Type", "application/json")
	bindRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(bindRec, bindReq)
	if bindRec.Code != http.StatusOK {
		t.Fatalf("bind published status code = %d; body=%s", bindRec.Code, bindRec.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/published", restoredWorkspaceID), nil)
	statusRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("published status code = %d; body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusBody publishedStatusResponse
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode published status response: %v", err)
	}
	if statusBody.URL != refreshBody.SubscriptionURL {
		t.Fatalf("bound workspace URL = %q, want original %q", statusBody.URL, refreshBody.SubscriptionURL)
	}

	secondRefreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", restoredWorkspaceID), nil)
	secondRefreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(secondRefreshRec, secondRefreshReq)
	if secondRefreshRec.Code != http.StatusOK {
		t.Fatalf("second refresh status code = %d; body=%s", secondRefreshRec.Code, secondRefreshRec.Body.String())
	}
	secondRefreshBody := decodeRefreshResponse(t, secondRefreshRec)
	if secondRefreshBody.PublishID != refreshBody.PublishID {
		t.Fatalf("second refresh PublishID = %q, want %q", secondRefreshBody.PublishID, refreshBody.PublishID)
	}
	if secondRefreshBody.SubscriptionURL != refreshBody.SubscriptionURL {
		t.Fatalf("second refresh URL = %q, want %q", secondRefreshBody.SubscriptionURL, refreshBody.SubscriptionURL)
	}
}

func TestRestoreDraftBindsPublishAndRefreshKeepsSameLink(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#restore-draft"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceA := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceA), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh A status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	first := decodeRefreshResponse(t, refreshRec)

	workspaceB := createWorkspaceForTest(t, server, cfg)
	restoreReqBody, err := json.Marshal(restoreDraftRequest{
		Config: cfg,
		PublishRef: restoreDraftPublishRefRequest{
			PublishID: first.PublishID,
		},
		NodeState: &model.NodeState{
			DisabledNodes: []string{"node-disabled-in-draft"},
		},
	})
	if err != nil {
		t.Fatalf("marshal restore draft request: %v", err)
	}
	restoreReq := httptest.NewRequest(http.MethodPost, "/api/workspaces/"+workspaceB+"/restore-draft", bytes.NewReader(restoreReqBody))
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(restoreRec, restoreReq)
	if restoreRec.Code != http.StatusOK {
		t.Fatalf("restore draft status code = %d; body=%s", restoreRec.Code, restoreRec.Body.String())
	}
	var restoreBody restoreDraftResponse
	if err := json.Unmarshal(restoreRec.Body.Bytes(), &restoreBody); err != nil {
		t.Fatalf("decode restore draft response: %v", err)
	}
	if !restoreBody.Publish.Exists || restoreBody.Publish.PublishID != first.PublishID {
		t.Fatalf("restore publish = %#v, want existing %q", restoreBody.Publish, first.PublishID)
	}
	if restoreBody.Publish.SubscriptionURL != first.SubscriptionURL {
		t.Fatalf("restore URL = %q, want %q", restoreBody.Publish.SubscriptionURL, first.SubscriptionURL)
	}
	refB, err := server.loadWorkspace(workspaceB)
	if err != nil {
		t.Fatalf("loadWorkspace(B) error = %v", err)
	}
	if refB.Meta.PublishID != first.PublishID {
		t.Fatalf("workspace B PublishID = %q, want %q", refB.Meta.PublishID, first.PublishID)
	}

	if _, err := os.Stat(filepath.Join(server.publishedRootDir(), first.PublishID, "current.yaml")); err != nil {
		t.Fatalf("stat current.yaml before refresh: %v", err)
	}

	secondReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceB), nil)
	secondRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("refresh B status code = %d; body=%s", secondRec.Code, secondRec.Body.String())
	}
	second := decodeRefreshResponse(t, secondRec)
	if second.PublishID != first.PublishID {
		t.Fatalf("refresh B PublishID = %q, want %q", second.PublishID, first.PublishID)
	}
	if second.SubscriptionURL != first.SubscriptionURL {
		t.Fatalf("refresh B URL = %q, want %q", second.SubscriptionURL, first.SubscriptionURL)
	}
	if publishedDirCount(t, server) != 1 {
		t.Fatalf("published dir count = %d, want 1", publishedDirCount(t, server))
	}
	if _, err := os.Stat(filepath.Join(server.publishedRootDir(), first.PublishID, "current.yaml")); err != nil {
		t.Fatalf("stat current.yaml after refresh: %v", err)
	}
}

func TestRestoreDraftMissingPublishClearsBinding(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#missing-restore"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceA := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceA), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	first := decodeRefreshResponse(t, refreshRec)
	if err := server.deletePublished(first.PublishID); err != nil {
		t.Fatalf("deletePublished() error = %v", err)
	}

	workspaceB := createWorkspaceForTest(t, server, cfg)
	restoreReqBody, err := json.Marshal(restoreDraftRequest{
		Config: cfg,
		PublishRef: restoreDraftPublishRefRequest{
			PublishID: first.PublishID,
		},
	})
	if err != nil {
		t.Fatalf("marshal restore draft request: %v", err)
	}
	restoreReq := httptest.NewRequest(http.MethodPost, "/api/workspaces/"+workspaceB+"/restore-draft", bytes.NewReader(restoreReqBody))
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(restoreRec, restoreReq)
	if restoreRec.Code != http.StatusOK {
		t.Fatalf("restore draft status code = %d; body=%s", restoreRec.Code, restoreRec.Body.String())
	}
	var restoreBody restoreDraftResponse
	if err := json.Unmarshal(restoreRec.Body.Bytes(), &restoreBody); err != nil {
		t.Fatalf("decode restore draft response: %v", err)
	}
	if restoreBody.Publish.Exists || restoreBody.Publish.Reason != "PUBLISHED_NOT_FOUND" {
		t.Fatalf("restore publish = %#v, want missing published", restoreBody.Publish)
	}
	refB, err := server.loadWorkspace(workspaceB)
	if err != nil {
		t.Fatalf("loadWorkspace(B) error = %v", err)
	}
	if refB.Meta.PublishID != "" {
		t.Fatalf("workspace B PublishID = %q, want empty", refB.Meta.PublishID)
	}

	secondReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceB), nil)
	secondRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("refresh B status code = %d; body=%s", secondRec.Code, secondRec.Body.String())
	}
	second := decodeRefreshResponse(t, secondRec)
	if second.PublishID == "" || second.PublishID == first.PublishID {
		t.Fatalf("new PublishID = %q, old = %q", second.PublishID, first.PublishID)
	}
}

func TestGetPublishedByIDNotFound(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/published/p_missing", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "PUBLISHED_NOT_FOUND") {
		t.Fatalf("body = %s, want PUBLISHED_NOT_FOUND", rec.Body.String())
	}
}

func TestWorkspaceCleanupRemovesWorkspaceButKeepsPublished(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#cleanup"},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", ref.ID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status code = %d; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshBody := decodeRefreshResponse(t, refreshRec)

	ref, err := server.loadWorkspaceByHash(ref.Hash)
	if err != nil {
		t.Fatalf("loadWorkspaceByHash() error = %v", err)
	}
	ref.Meta.LastAccessAt = time.Now().UTC().Add(-48 * time.Hour)
	if err := server.saveWorkspaceMeta(ref); err != nil {
		t.Fatalf("saveWorkspaceMeta() error = %v", err)
	}
	if err := server.cleanupExpiredWorkspaces(); err != nil {
		t.Fatalf("cleanupExpiredWorkspaces() error = %v", err)
	}
	if _, err := os.Stat(ref.Dir); !os.IsNotExist(err) {
		t.Fatalf("workspace dir still exists after cleanup: err=%v", err)
	}

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("published subscription status code = %d, want %d; body=%s", subRec.Code, http.StatusOK, subRec.Body.String())
	}
}

func TestDeletePublishedInvalidatesLink(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#delete-published"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	refreshBody := decodeRefreshResponse(t, refreshRec)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/published/"+refreshBody.PublishID, nil)
	deleteRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete published status code = %d; body=%s", deleteRec.Code, deleteRec.Body.String())
	}

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusNotFound {
		t.Fatalf("subscription status after delete = %d, want %d; body=%s", subRec.Code, http.StatusNotFound, subRec.Body.String())
	}

	ref, err := server.loadWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("loadWorkspace() error = %v", err)
	}
	if ref.Meta.PublishID != "" {
		t.Fatalf("workspace PublishID = %q, want cleared after publish delete", ref.Meta.PublishID)
	}
	statusReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/published", workspaceID), nil)
	statusRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("published status code = %d; body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusBody publishedStatusResponse
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode published status: %v", err)
	}
	if statusBody.PublishID != "" || statusBody.SubscriptionURL != "" {
		t.Fatalf("published status after delete = %#v, want empty", statusBody)
	}
}

func TestPublishedAccessCountAndNoStoreHeaders(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#access"},
	}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	refreshReq := httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", workspaceID), nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	refreshBody := decodeRefreshResponse(t, refreshRec)

	subReq := httptest.NewRequest(http.MethodGet, publishedPathFromURL(t, refreshBody.SubscriptionURL), nil)
	subRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusOK {
		t.Fatalf("published access status code = %d; body=%s", subRec.Code, subRec.Body.String())
	}
	if got := subRec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want %q", got, "no-store")
	}
	if got := subRec.Header().Get("X-Robots-Tag"); got != "noindex, nofollow, noarchive" {
		t.Fatalf("X-Robots-Tag = %q", got)
	}
	if got := subRec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}

	statusReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/published", workspaceID), nil)
	statusRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("published status code = %d; body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusBody publishedStatusResponse
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode published status: %v", err)
	}
	if statusBody.AccessCount < 1 {
		t.Fatalf("AccessCount = %d, want >= 1", statusBody.AccessCount)
	}
	if statusBody.LastAccessAt == "" {
		t.Fatalf("LastAccessAt is empty")
	}
}

func TestNodeStateDraftAPI(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)

	putState := model.NodeState{
		NodeOverrides: map[string]model.NodeOverride{
			"node-1": {
				Enabled: true,
				Name:    "renamed-node",
				Fields:  model.NodeOverrideFields{Server: "example.org", Port: 8443},
			},
		},
		DisabledNodes: []string{"node-2"},
		DeletedNodes:  []string{"node-3"},
		CustomNodes: []model.NodeIR{
			{
				ID:     "custom-node-1",
				Name:   "custom-node",
				Type:   model.ProtocolSS,
				Server: "127.0.0.1",
				Port:   8388,
				Auth:   model.Auth{Password: "secret"},
				Source: model.SourceInfo{Kind: "custom", Name: "manual"},
			},
		},
	}
	body, err := json.Marshal(nodeStateRequest{State: putState})
	if err != nil {
		t.Fatalf("marshal node state request: %v", err)
	}
	putReq := httptest.NewRequest(http.MethodPut, withWorkspace("/api/nodes/state", workspaceID), bytes.NewReader(body))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("put node state status code = %d; body=%s", putRec.Code, putRec.Body.String())
	}

	ref, err := server.loadWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("loadWorkspace() error = %v", err)
	}
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	loaded, err := pipeline.LoadNodeState(workspaceCfg)
	if err != nil {
		t.Fatalf("LoadNodeState() error = %v", err)
	}
	loaded.SubscriptionMeta = map[string]model.SubscriptionMeta{
		"source-1": {SourceID: "source-1", SourceName: "hidden-meta", Total: 1024},
	}
	loaded.LastAudit = model.AuditReport{RawCount: 99, FinalCount: 88}
	if err := pipeline.SaveNodeState(workspaceCfg, loaded); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes/state", workspaceID), nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get node state status code = %d; body=%s", getRec.Code, getRec.Body.String())
	}
	var getBody nodeStateResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode node state response: %v", err)
	}
	if got := getBody.State.NodeOverrides["node-1"].Name; got != "renamed-node" {
		t.Fatalf("override name = %q, want renamed-node", got)
	}
	if !containsString(getBody.State.DisabledNodes, "node-2") || !containsString(getBody.State.DeletedNodes, "node-3") {
		t.Fatalf("state disabled/deleted = %#v/%#v", getBody.State.DisabledNodes, getBody.State.DeletedNodes)
	}
	if len(getBody.State.CustomNodes) != 1 {
		t.Fatalf("len(CustomNodes) = %d, want 1", len(getBody.State.CustomNodes))
	}
	if len(getBody.State.SubscriptionMeta) != 0 {
		t.Fatalf("SubscriptionMeta = %#v, want omitted from draft state", getBody.State.SubscriptionMeta)
	}
	if getBody.State.LastAudit.RawCount != 0 || getBody.State.LastAudit.FinalCount != 0 {
		t.Fatalf("LastAudit = %#v, want zero draft audit", getBody.State.LastAudit)
	}
}

func TestNoTokenInLogs(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	server, cfg := newTestServer(t, cfg)
	token := "abcd1234efgh5678"
	server.appendLog("published url http://127.0.0.1:9876/s/" + token + "/mihomo.yaml")

	data, err := os.ReadFile(filepath.Join(server.logsDir(), "app.log"))
	if err != nil {
		t.Fatalf("ReadFile(app.log) error = %v", err)
	}
	content := string(data)
	if strings.Contains(content, token) {
		t.Fatalf("log leaked token: %s", content)
	}
	if !strings.Contains(content, "/s/<redacted>/mihomo.yaml") {
		t.Fatalf("log content = %q, want redacted published path", content)
	}
}

func TestAppLogRotation(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	server, cfg := newTestServer(t, cfg)

	logPath := filepath.Join(server.logsDir(), "app.log")
	if err := os.MkdirAll(server.logsDir(), 0o755); err != nil {
		t.Fatalf("MkdirAll(logsDir) error = %v", err)
	}
	if err := os.WriteFile(logPath, bytes.Repeat([]byte("a"), maxAppLogBytes), 0o644); err != nil {
		t.Fatalf("WriteFile(app.log) error = %v", err)
	}

	server.appendLog("rotate-me")

	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatalf("Stat(app.log.1) error = %v", err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(app.log) error = %v", err)
	}
	if !bytes.Contains(data, []byte("rotate-me")) {
		t.Fatalf("app.log = %q, want rotated fresh log line", string(data))
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestRootServesHTML(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("SubConv Next 订阅转换")) {
		t.Fatalf("body = %q, want embedded UI", rec.Body.String())
	}
}

func TestStyleCSSServed(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", got, "text/css; charset=utf-8")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(".app-shell")) {
		t.Fatalf("body = %q, want stylesheet content", rec.Body.String())
	}
}
