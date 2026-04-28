package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

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
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
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
	if body.OutputPath != cfg.Service.OutputPath {
		t.Fatalf("body.OutputPath = %q, want %q", body.OutputPath, cfg.Service.OutputPath)
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

	server := NewServer("0.1.0-test", cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/subscription-meta", nil)
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

	req := httptest.NewRequest(http.MethodGet, "/api/nodes", nil)
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

	listReq := httptest.NewRequest(http.MethodGet, "/api/nodes", nil)
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

	detailReq := httptest.NewRequest(http.MethodGet, "/api/nodes/"+nodeID, nil)
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
	overrideReq := httptest.NewRequest(http.MethodPut, "/api/nodes/"+nodeID+"/override", overrideBody)
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

	resetReq := httptest.NewRequest(http.MethodPost, "/api/nodes/"+nodeID+"/reset", nil)
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

	listReq := httptest.NewRequest(http.MethodGet, "/api/nodes", nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)

	var listBody nodeListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	nodeID := listBody.Nodes[0].ID

	disableReq := httptest.NewRequest(http.MethodPost, "/api/nodes/disable", bytes.NewBufferString(`{"ids":["`+nodeID+`"]}`))
	disableRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(disableRec, disableReq)
	if disableRec.Code != http.StatusOK {
		t.Fatalf("disable status code = %d, want %d; body=%s", disableRec.Code, http.StatusOK, disableRec.Body.String())
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusBadRequest {
		t.Fatalf("refresh status code = %d, want %d when all nodes disabled", refreshRec.Code, http.StatusBadRequest)
	}

	customReq := httptest.NewRequest(http.MethodPost, "/api/nodes/custom", bytes.NewBufferString(`{"content":"vless://uuid-1@example.com:443?sni=example.com#custom-node","content_type":"uri"}`))
	customRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(customRec, customReq)
	if customRec.Code != http.StatusOK {
		t.Fatalf("custom status code = %d, want %d; body=%s", customRec.Code, http.StatusOK, customRec.Body.String())
	}

	enableReq := httptest.NewRequest(http.MethodPost, "/api/nodes/enable", bytes.NewBufferString(`{"ids":["`+nodeID+`"]}`))
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
	data, err := os.ReadFile(cfg.Service.OutputPath)
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

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
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

func TestHandleSubscriptionYAML(t *testing.T) {
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/yaml; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", got, "text/yaml; charset=utf-8")
	}
	if got := rec.Header().Get("Content-Disposition"); got != `inline; filename="mihomo.yaml"` {
		t.Fatalf("Content-Disposition = %q, want %q", got, `inline; filename="mihomo.yaml"`)
	}
	if got := rec.Header().Get("X-SubConv-Refresh-Status"); got != "fresh" {
		t.Fatalf("X-SubConv-Refresh-Status = %q, want %q", got, "fresh")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("type: ss")) {
		t.Fatalf("body = %q, want ss yaml", rec.Body.String())
	}
}

func TestHandleSubscriptionYAMLDownload(t *testing.T) {
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="mihomo.yaml"` {
		t.Fatalf("Content-Disposition = %q, want %q", got, `attachment; filename="mihomo.yaml"`)
	}
}

func TestHandleSubscriptionYAMLRequiresTokenWhenConfigured(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.SubscriptionToken = "secret-token"
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
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/sub/mihomo.yaml?token=secret-token", nil)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestHandleSubscriptionYAMLUserinfoAggregateHeader(t *testing.T) {
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Subscription-Userinfo"); got != "upload=0; download=42949672960; total=322122547200; expire=1779235200" {
		t.Fatalf("Subscription-Userinfo = %q, want aggregated header", got)
	}
}

func TestHandleSubscriptionYAMLOmitsUserinfoWhenMergeStrategyNone(t *testing.T) {
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
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

	req := httptest.NewRequest(http.MethodGet, "/api/logs?tail=1", nil)
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

func TestHandleSubscriptionYAMLCacheFresh(t *testing.T) {
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if !bytes.Equal(rec.Body.Bytes(), oldYAML) {
		t.Fatalf("body = %q, want cached yaml", rec.Body.Bytes())
	}
	if got := rec.Header().Get("X-SubConv-Refresh-Status"); got != "cached" {
		t.Fatalf("X-SubConv-Refresh-Status = %q, want %q", got, "cached")
	}
}

func TestHandleSubscriptionYAMLCacheExpired(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Service.RefreshInterval = 1
	cfg.Service.RefreshOnRequest = true
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczNAZXhhbXBsZS5jb206NDQz#expired"},
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("expired")) {
		t.Fatalf("body = %q, want refreshed yaml", rec.Body.Bytes())
	}
	if got := rec.Header().Get("X-SubConv-Refresh-Status"); got != "fresh" {
		t.Fatalf("X-SubConv-Refresh-Status = %q, want %q", got, "fresh")
	}
}

func TestHandleSubscriptionYAMLStaleIfError(t *testing.T) {
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !bytes.Equal(rec.Body.Bytes(), oldYAML) {
		t.Fatalf("body = %q, want stale yaml", rec.Body.Bytes())
	}
	if got := rec.Header().Get("X-SubConv-Refresh-Status"); got != "stale" {
		t.Fatalf("X-SubConv-Refresh-Status = %q, want %q", got, "stale")
	}
	if got := rec.Header().Get("X-SubConv-Warning"); got == "" {
		t.Fatalf("X-SubConv-Warning is empty, want stale warning")
	}
}

func TestHandleSubscriptionYAMLNoStaleAvailable(t *testing.T) {
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

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d; body=%s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}
}

func TestRefreshLock(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{Name: "manual", Enabled: true, Content: "ss://YWVzLTI1Ni1nY206cGFzczRAZXhhbXBsZS5jb206NDQz#lock"},
	}
	server := NewServer("0.1.0-test", cfg)

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
		req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
	}()

	<-started

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
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

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	refreshRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("initial refresh status code = %d", refreshRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/nodes", nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	var listBody nodeListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	nodeID := listBody.Nodes[0].ID

	overrideBody := bytes.NewBufferString(`{"enabled":true,"name":"renamed-node"}`)
	overrideReq := httptest.NewRequest(http.MethodPut, "/api/nodes/"+nodeID+"/override", overrideBody)
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

	data, err := os.ReadFile(cfg.Service.OutputPath)
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

func TestHandleConfigGetAndPut(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := model.DefaultConfig()
	server := NewServer("0.1.0-test", cfg)
	server.SetConfigPath(configPath)

	getReq := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status code = %d, want %d", getRec.Code, http.StatusOK)
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
	putReq := httptest.NewRequest(http.MethodPut, "/api/config", putBody)
	putRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT status code = %d, want %d; body=%s", putRec.Code, http.StatusOK, putRec.Body.String())
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	if !bytes.Contains(data, []byte(`"name": "manual"`)) {
		t.Fatalf("written config = %q, want inline source", string(data))
	}
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
