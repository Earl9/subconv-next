package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"subconv-next/internal/model"
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
	if got := rec.Header().Get("Content-Type"); got != "application/yaml" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/yaml")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("type: ss")) {
		t.Fatalf("body = %q, want ss yaml", rec.Body.String())
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
	if !bytes.Contains(rec.Body.Bytes(), []byte("SubConv Next")) {
		t.Fatalf("body = %q, want embedded UI", rec.Body.String())
	}
}
