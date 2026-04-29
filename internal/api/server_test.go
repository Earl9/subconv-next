package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

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
	server := NewServer("0.1.0-test", cfg)

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
	if body.UptimeSeconds < 0 {
		t.Fatalf("body.UptimeSeconds = %d, want >= 0", body.UptimeSeconds)
	}
}

func TestHandleStatus(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Subscriptions = []model.SubscriptionConfig{
		{Name: "enabled", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent},
		{Name: "disabled", Enabled: false, URL: "https://example.com/b", UserAgent: model.DefaultUserAgent},
	}

	server := NewServer("0.1.0-test", cfg)
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

	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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

	server := NewServer("0.1.0-test", cfg)
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

	server := NewServer("0.1.0-test", cfg)
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

	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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

	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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

func TestHandleConfigGetAndPut(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.AccessToken = "config-token"
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "source-1", Name: "secret-source", Enabled: true, URL: "https://example.com/sub?token=abc123&user=bob", UserAgent: model.DefaultUserAgent},
	}
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestWorkspaceCreateAndDelete(t *testing.T) {
	cfg := model.DefaultConfig()
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)
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

func TestPublishedSubscriptionStillWorksAfterWorkspaceDeleted(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#ss-node"},
	}
	server := NewServer("0.1.0-test", cfg)
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

func TestRefreshOverwritesCurrentYAMLAndKeepsSameToken(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#stable"},
	}
	server := NewServer("0.1.0-test", cfg)
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

func TestRotateTokenInvalidatesOldLink(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#rotate"},
	}
	server := NewServer("0.1.0-test", cfg)
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

func TestWorkspaceCleanupRemovesWorkspaceButKeepsPublished(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#cleanup"},
	}
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)
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
}

func TestPublishedAccessCountAndNoStoreHeaders(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczdAZXhhbXBsZS5jb206NDQz#access"},
	}
	server := NewServer("0.1.0-test", cfg)
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

func TestNoTokenInLogs(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	server := NewServer("0.1.0-test", cfg)
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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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
	server := NewServer("0.1.0-test", cfg)

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
