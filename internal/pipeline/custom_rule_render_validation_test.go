package pipeline

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
)

func TestRenderingRejectsStoredInvalidCustomRuleBeforeFetchingOrWriting(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		name := "enabled"
		if !enabled {
			name = "disabled"
		}
		for _, cached := range []bool{false, true} {
			mode := "refresh"
			if cached {
				mode = "cached"
			}
			t.Run(name+"/"+mode, func(t *testing.T) {
				var requests atomic.Int32
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					_, _ = w.Write([]byte("trojan://fixture@remote.example:443#remote"))
				}))
				defer upstream.Close()
				cfg := cachedRenderTestConfig(t)
				cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "Source", Enabled: true, AllowLAN: true, URL: upstream.URL}}
				cfg.Render.CustomRules = []model.CustomRule{{
					Key: "stored_invalid", Label: "Stored invalid", Enabled: enabled,
					SourceType: "inline", Behavior: "classical", Format: "text",
					TargetMode: "direct", Payload: []string{"DOMAIN-SUFFIX,"},
				}}
				if err := SaveNodeState(cfg, model.DefaultNodeState()); err != nil {
					t.Fatal(err)
				}
				beforeState, err := os.ReadFile(cfg.Service.StatePath)
				if err != nil {
					t.Fatal(err)
				}
				beforeYAML := []byte("previous published configuration\n")
				if err := os.WriteFile(cfg.Service.OutputPath, beforeYAML, 0o600); err != nil {
					t.Fatal(err)
				}
				var result RenderResult
				if cached {
					prospective := model.DefaultNodeState()
					prospective.DeletedNodes = []string{"uncommitted-deletion"}
					result, err = RenderCachedConfig(cfg, prospective)
				} else {
					result, err = RenderConfig(cfg)
				}
				var validationErr *config.CustomRuleValidationError
				if !errors.As(err, &validationErr) || validationErr.Field != "payload" || validationErr.Line != 1 {
					t.Fatalf("error = %v, want payload validation before source collection", err)
				}
				if len(result.YAML) != 0 || result.NodeCount != 0 || requests.Load() != 0 {
					t.Fatalf("invalid rule produced output or fetched: nodes=%d bytes=%d requests=%d", result.NodeCount, len(result.YAML), requests.Load())
				}
				afterState, err := os.ReadFile(cfg.Service.StatePath)
				if err != nil || !bytes.Equal(beforeState, afterState) {
					t.Fatal("invalid stored rule changed persisted node state")
				}
				afterYAML, err := os.ReadFile(cfg.Service.OutputPath)
				if err != nil || !bytes.Equal(beforeYAML, afterYAML) {
					t.Fatal("invalid stored rule replaced the published configuration")
				}
				if _, err := os.Stat(cfg.Service.CacheDir); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("invalid stored rule touched source caches: %v", err)
				}
			})
		}
	}
}

func TestRenderingAcceptsValidExistingCustomRule(t *testing.T) {
	for _, cached := range []bool{false, true} {
		mode := "refresh"
		if cached {
			mode = "cached"
		}
		t.Run(mode, func(t *testing.T) {
			cfg := cachedRenderTestConfig(t)
			cfg.Inline = []model.InlineConfig{{ID: "inline", Name: "Inline", Enabled: true, Content: "trojan://fixture@keep.example:443#keep"}}
			// An older draft can omit scalar defaults and use domain behavior for
			// a classical entry. Rendering retains its existing normalization.
			cfg.Render.CustomRules = []model.CustomRule{{
				Key: "existing", Label: "Existing", Enabled: true,
				SourceType: "inline", Behavior: "domain", TargetMode: "direct",
				Payload: []string{"DOMAIN-SUFFIX,existing.example"},
			}}
			var result RenderResult
			var err error
			if cached {
				result, err = RenderCachedConfig(cfg, model.DefaultNodeState())
			} else {
				result, err = RenderConfig(cfg)
			}
			if err != nil || result.NodeCount != 1 || !strings.Contains(string(result.YAML), "DOMAIN-SUFFIX,existing.example") {
				t.Fatalf("valid existing rule rejected: nodes=%d err=%v", result.NodeCount, err)
			}
			if !strings.Contains(string(result.YAML), "behavior: classical") {
				t.Fatal("existing automatic classical behavior was lost")
			}
			if _, err := os.Stat(cfg.Service.StatePath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("rendering wrote node state: %v", err)
			}
			if _, err := os.Stat(cfg.Service.OutputPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("rendering wrote published YAML: %v", err)
			}
		})
	}
}

func TestRenderingRejectsInvalidDirectRuleBeforeFetchingOrWriting(t *testing.T) {
	for _, rule := range []string{"DOMAIN-SUFFIX,example.com", "DOMAIN-SUFFIX,,DIRECT", "IP-CIDR,192.0.2.0/99,DIRECT", "DOMIAN,example.com,DIRECT"} {
		for _, cached := range []bool{false, true} {
			mode := "refresh"
			if cached {
				mode = "cached"
			}
			t.Run(rule+"/"+mode, func(t *testing.T) {
				var requests atomic.Int32
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					_, _ = w.Write([]byte("trojan://fixture@remote.example:443#remote"))
				}))
				defer upstream.Close()
				cfg := cachedRenderTestConfig(t)
				cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "Source", Enabled: true, AllowLAN: true, URL: upstream.URL}}
				cfg.Render.AdditionalRules = []string{"# valid comment", rule}
				if err := SaveNodeState(cfg, model.DefaultNodeState()); err != nil {
					t.Fatal(err)
				}
				beforeState, err := os.ReadFile(cfg.Service.StatePath)
				if err != nil {
					t.Fatal(err)
				}
				beforeYAML := []byte("previous published configuration\n")
				if err := os.WriteFile(cfg.Service.OutputPath, beforeYAML, 0o600); err != nil {
					t.Fatal(err)
				}
				var result RenderResult
				if cached {
					result, err = RenderCachedConfig(cfg, model.DefaultNodeState())
				} else {
					result, err = RenderConfig(cfg)
				}
				var validationErr *config.CustomRuleValidationError
				if !errors.As(err, &validationErr) || validationErr.Field != "additional_rules" || validationErr.Line != 2 {
					t.Fatalf("error = %v, want direct-rule error at original line 2", err)
				}
				if requests.Load() != 0 || len(result.YAML) != 0 {
					t.Fatal("invalid direct rule fetched a source or produced YAML")
				}
				afterState, err := os.ReadFile(cfg.Service.StatePath)
				if err != nil || !bytes.Equal(beforeState, afterState) {
					t.Fatal("invalid direct rule changed node state")
				}
				afterYAML, err := os.ReadFile(cfg.Service.OutputPath)
				if err != nil || !bytes.Equal(beforeYAML, afterYAML) {
					t.Fatal("invalid direct rule replaced published YAML")
				}
			})
		}
	}
}
