package pipeline

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
)

func TestRenderCachedConfigAppliesProspectiveDeletionWithoutNetworkOrWrites(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/rules" {
			_, _ = w.Write([]byte("DOMAIN-SUFFIX,retained.example\n"))
			return
		}
		w.Header().Set("Subscription-Userinfo", "upload=10; download=20; total=1000")
		_, _ = w.Write([]byte("trojan://test@delete.example:443#delete-me\ntrojan://test@keep.example:443#keep"))
	}))
	defer upstream.Close()
	cfg := cachedRenderTestConfig(t)
	cfg.Service.AllowLAN = true
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", Enabled: true, URL: upstream.URL + "/nodes", AllowLAN: true}}
	cfg.Render.CustomRules = []model.CustomRule{{Key: "retained", Label: "Retained", Enabled: true, SourceType: "http", Format: "text", Behavior: "classical", TargetMode: "direct", URL: upstream.URL + "/rules", Interval: 1}}
	if err := SaveNodeState(cfg, model.DefaultNodeState()); err != nil {
		t.Fatal(err)
	}
	initial, err := RenderConfig(cfg)
	if err != nil || initial.NodeCount != 2 || requests.Load() != 2 {
		t.Fatalf("initial render: nodes=%d, requests=%d, err=%v", initial.NodeCount, requests.Load(), err)
	}

	// An expired rule snapshot must still be used without checking upstream.
	metas, err := filepath.Glob(filepath.Join(cfg.Service.CacheDir, "custom-rules", "*.meta.json"))
	if err != nil || len(metas) != 1 {
		t.Fatalf("custom rule cache metadata = %v, err=%v", metas, err)
	}
	data, err := os.ReadFile(metas[0])
	if err != nil {
		t.Fatal(err)
	}
	var meta fetcher.CacheMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatal(err)
	}
	meta.FetchedAt = time.Now().Add(-24 * time.Hour)
	data, err = json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metas[0], data, 0o600); err != nil {
		t.Fatal(err)
	}

	next := model.DefaultNodeState()
	next.DeletedNodes = []string{initial.Nodes[0].ID}
	result, err := RenderCachedConfig(cfg, next)
	if err != nil || result.NodeCount != 1 || result.Nodes[0].Server != "keep.example" {
		t.Fatalf("cached deletion: nodes=%d, err=%v", result.NodeCount, err)
	}
	if requests.Load() != 2 || !strings.Contains(string(result.YAML), "retained.example") || strings.Contains(string(result.YAML), "delete.example") {
		t.Fatal("cached rendering refetched upstream or lost cached rules/deletion")
	}
	if result.SubscriptionMeta["source"].Total != 1000 || result.Audit.FinalCount != 1 {
		t.Fatal("cached rendering lost subscription metadata or the final audit")
	}
	persisted, err := LoadNodeState(cfg)
	if err != nil || len(persisted.DeletedNodes) != 0 {
		t.Fatalf("prospective state was saved: deleted=%v, err=%v", persisted.DeletedNodes, err)
	}
	if _, err := os.Stat(cfg.Service.OutputPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rendering wrote output: %v", err)
	}

	next.DeletedNodes = []string{initial.Nodes[0].ID, initial.Nodes[1].ID}
	empty, err := RenderCachedConfig(cfg, next)
	if !errors.Is(err, ErrNoNodes) || empty.Audit.FinalCount != 0 || empty.Audit.ExcludedCount != 2 || len(empty.State.DeletedNodes) != 2 || empty.SubscriptionMeta["source"].Total != 1000 {
		t.Fatalf("empty cached result lost state/audit/metadata: result=%+v, err=%v", empty, err)
	}
	if requests.Load() != 2 {
		t.Fatal("deleting the last node refetched upstream")
	}
}

func TestRenderCachedConfigRejectsMissingSourceCacheWithoutFetching(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("trojan://test@remote.example:443#remote"))
	}))
	defer upstream.Close()
	cfg := cachedRenderTestConfig(t)
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://test@keep.example:443#keep"}}
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "missing", Name: "missing", Enabled: true, URL: upstream.URL, AllowLAN: true}}
	result, err := RenderCachedConfig(cfg, model.DefaultNodeState())
	if err == nil || len(result.YAML) != 0 || len(result.Errors) != 1 || result.Errors[0].Kind != "FETCH_FAILED" || requests.Load() != 0 {
		t.Fatalf("missing source produced partial output or fetched upstream: result=%+v, requests=%d, err=%v", result, requests.Load(), err)
	}
	preview := CollectPreviewNodes(cfg)
	if len(preview.Nodes) != 1 || len(preview.Errors) != 0 || len(preview.Warnings) != 1 || requests.Load() != 0 {
		t.Fatal("the editor's offline cache-miss behavior changed")
	}
	cfg.Subscriptions[0].Enabled = false
	result, err = RenderCachedConfig(cfg, model.DefaultNodeState())
	if err != nil || result.NodeCount != 1 || requests.Load() != 0 {
		t.Fatalf("a disabled source required a cache: nodes=%d, requests=%d, err=%v", result.NodeCount, requests.Load(), err)
	}
}

func TestRenderCachedConfigRejectsMissingRuleCacheWithoutFetching(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("example.com\n"))
	}))
	defer upstream.Close()
	cfg := cachedRenderTestConfig(t)
	cfg.Service.AllowLAN = true
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://test@keep.example:443#keep"}}
	cfg.Render.CustomRules = []model.CustomRule{{Key: "missing", Label: "Missing", Enabled: true, SourceType: "http", Format: "text", TargetMode: "direct", URL: upstream.URL}}
	result, err := RenderCachedConfig(cfg, model.DefaultNodeState())
	if !errors.Is(err, fetcher.ErrCacheMiss) || len(result.YAML) != 0 || requests.Load() != 0 {
		t.Fatalf("missing rule cache: requests=%d, err=%v", requests.Load(), err)
	}
}

func cachedRenderTestConfig(t *testing.T) model.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.CacheDir = filepath.Join(dir, "cache")
	cfg.Service.StatePath = filepath.Join(dir, "state.json")
	cfg.Service.OutputPath = filepath.Join(dir, "current.yaml")
	return cfg
}
