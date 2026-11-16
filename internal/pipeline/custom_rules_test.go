package pipeline

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
	"subconv-next/internal/rulevalidation"
)

func TestRemoteCustomRuleInvalidRefreshPreservesOfflineSnapshot(t *testing.T) {
	var body atomic.Value
	body.Store("example.com\nexample.net\n")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(body.Load().(string)))
	}))
	defer server.Close()
	cfg := model.DefaultConfig()
	cfg.Service.CacheDir, cfg.Service.AllowLAN = t.TempDir(), true
	cfg.Render.CustomRules = []model.CustomRule{{Key: "remote", Label: "Remote", Enabled: true, SourceType: "http", URL: server.URL, Behavior: "domain", Format: "text"}}
	first, _, err := snapshotRemoteCustomRules(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body.Store("<!doctype html><html>upstream failure</html>")
	second, warnings, err := snapshotRemoteCustomRules(cfg)
	if err != nil || !reflect.DeepEqual(second.Render.CustomRules[0].Payload, first.Render.CustomRules[0].Payload) || !strings.Contains(strings.Join(warnings, " "), "cached snapshot") {
		t.Fatalf("invalid remote update lost valid rules: warnings=%v error=%v", warnings, err)
	}
	before := calls.Load()
	offline, _, err := snapshotRemoteCustomRulesForRender(cfg, true)
	if err != nil || !reflect.DeepEqual(offline.Render.CustomRules[0].Payload, first.Render.CustomRules[0].Payload) || calls.Load() != before {
		t.Fatalf("offline render lost the valid snapshot after upstream error: %v", err)
	}
}

type customRuleFetcherStub struct {
	fetched  fetcher.FetchedSubscription
	warnings []string
	err      error
	calls    int
}

func (stub *customRuleFetcherStub) Fetch(context.Context, fetcher.Source) (fetcher.FetchedSubscription, []string, error) {
	stub.calls++
	return stub.fetched, stub.warnings, stub.err
}

func TestSnapshotRemoteCustomRulesEmbedsYAMLPayload(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{
		Key:        "adobe",
		Label:      "Adobe",
		Enabled:    true,
		SourceType: "http",
		Behavior:   "classical",
		Format:     "yaml",
		URL:        "https://example.com/adobe.yaml",
		Payload:    nil,
	}}
	stub := &customRuleFetcherStub{fetched: fetcher.FetchedSubscription{
		Content: []byte("payload:\n  - DOMAIN-SUFFIX,adobe.com\n  - DOMAIN-SUFFIX,adobe.io\n  - DOMAIN-SUFFIX,adobe.com\n"),
	}}

	got, warnings, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
	if err != nil {
		t.Fatalf("snapshotRemoteCustomRulesWithFetcher() error = %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("Fetch() calls = %d, want 1", stub.calls)
	}
	if cfg.Render.CustomRules[0].SourceType != "http" {
		t.Fatalf("input config was mutated: %+v", cfg.Render.CustomRules[0])
	}
	rule := got.Render.CustomRules[0]
	if rule.SourceType != "inline" || rule.Format != "text" || rule.URL != "" {
		t.Fatalf("resolved rule = %+v, want inline snapshot", rule)
	}
	wantPayload := []string{"DOMAIN-SUFFIX,adobe.com", "DOMAIN-SUFFIX,adobe.io"}
	if !reflect.DeepEqual(rule.Payload, wantPayload) {
		t.Fatalf("Payload = %#v, want %#v", rule.Payload, wantPayload)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "embedded 2 entries") {
		t.Fatalf("warnings = %#v, want snapshot entry count", warnings)
	}
}

func TestSnapshotRemoteCustomRulesUsesCachedFetchResult(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{
		Key:        "cached",
		Label:      "Cached",
		Enabled:    true,
		SourceType: "http",
		Behavior:   "domain",
		Format:     "text",
		URL:        "https://example.com/cached.txt",
	}}
	stub := &customRuleFetcherStub{
		fetched: fetcher.FetchedSubscription{
			Content:   []byte("example.com\nexample.net\n"),
			FromCache: true,
		},
		warnings: []string{"using cached remote rule"},
	}

	got, warnings, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
	if err != nil {
		t.Fatalf("snapshotRemoteCustomRulesWithFetcher() error = %v", err)
	}
	if !reflect.DeepEqual(got.Render.CustomRules[0].Payload, []string{"example.com", "example.net"}) {
		t.Fatalf("Payload = %#v", got.Render.CustomRules[0].Payload)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "using cached remote rule") || !strings.Contains(joined, "cached snapshot") {
		t.Fatalf("warnings = %#v, want cache diagnostics", warnings)
	}
}

func TestSnapshotRemoteCustomRulesFailsClosedWithoutSnapshot(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{
		Key:        "required",
		Label:      "Required",
		Enabled:    true,
		SourceType: "http",
		Format:     "yaml",
		URL:        "https://example.com/required.yaml",
	}}
	stub := &customRuleFetcherStub{err: errors.New("network unavailable")}

	_, _, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("error = %v, want fetch failure", err)
	}
}

func TestSnapshotRemoteCustomRulesLeavesMRSAtRuntime(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{
		Key:        "binary",
		Label:      "Binary",
		Enabled:    true,
		SourceType: "http",
		Behavior:   "domain",
		Format:     "mrs",
		URL:        "https://example.com/binary.mrs",
	}}
	stub := &customRuleFetcherStub{}

	got, warnings, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
	if err != nil {
		t.Fatalf("snapshotRemoteCustomRulesWithFetcher() error = %v", err)
	}
	if stub.calls != 0 {
		t.Fatalf("Fetch() calls = %d, want 0", stub.calls)
	}
	if got.Render.CustomRules[0].SourceType != "http" {
		t.Fatalf("MRS source type = %q, want http", got.Render.CustomRules[0].SourceType)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "remains a runtime rule provider") {
		t.Fatalf("warnings = %#v", warnings)
	}
}

func TestSnapshotRemoteCustomRulesRejectsInvalidContentBeforeEmbedding(t *testing.T) {
	for _, fixture := range []struct {
		name, behavior, format, content string
		line                            int
	}{
		{"html_instead_of_rules", "domain", "text", "<!DOCTYPE html>\n<html><body>Not found</body></html>", 1},
		{"url_instead_of_domain", "domain", "text", "example.com\nhttps://example.org/path", 2},
		{"bad_cidr", "ipcidr", "text", "192.0.2.0/24\n192.0.2.0/99", 2},
		{"missing_classical_operand", "classical", "text", "DOMAIN-SUFFIX,example.com\nDOMAIN-SUFFIX,", 2},
		{"invalid_yaml_port", "classical", "yaml", "payload:\n  - DOMAIN-SUFFIX,example.com\n  - DST-PORT,70000\n", 2},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Render.CustomRules = []model.CustomRule{{
				Key: "invalid_source", Label: "Invalid source", Enabled: true,
				SourceType: "http", Behavior: fixture.behavior, Format: fixture.format,
				URL: "https://example.com/rules", Interval: 3600,
			}}
			stub := &customRuleFetcherStub{fetched: fetcher.FetchedSubscription{Content: []byte(fixture.content)}}
			got, _, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
			var validationErr *rulevalidation.Error
			if err == nil || !strings.Contains(err.Error(), "invalid_source") || !errors.As(err, &validationErr) || validationErr.Line != fixture.line {
				t.Fatalf("error = %v, want source key and validation error at line %d", err, fixture.line)
			}
			if !reflect.DeepEqual(got, cfg) || !reflect.DeepEqual(got.Render.CustomRules[0].Payload, cfg.Render.CustomRules[0].Payload) {
				t.Fatal("invalid remote content changed the source configuration")
			}
			if stub.calls != 1 {
				t.Fatalf("Fetch() calls = %d, want one snapshot attempt", stub.calls)
			}
		})
	}
}

func TestSnapshotRemoteCustomRulesInfersClassicalBehaviorBeforeValidation(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{
		Key: "inferred", Label: "Inferred", Enabled: true,
		SourceType: "http", Behavior: "domain", Format: "text",
		URL: "https://example.com/rules.txt",
	}}
	stub := &customRuleFetcherStub{fetched: fetcher.FetchedSubscription{
		Content: []byte("DOMAIN-SUFFIX,example.com\nIP-CIDR,192.0.2.0/24,no-resolve\n"),
	}}
	got, _, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
	if err != nil {
		t.Fatalf("inferred classical snapshot: %v", err)
	}
	if got.Render.CustomRules[0].Behavior != "classical" || got.Render.CustomRules[0].SourceType != "inline" {
		t.Fatalf("snapshot behavior not preserved: %+v", got.Render.CustomRules[0])
	}
}

func TestSnapshotRemoteCustomRulesAcceptsDomainWildcardsAndIPv6(t *testing.T) {
	for _, fixture := range []struct{ behavior, content string }{
		{"domain", "#.comment\n+.example.com\n*.example.org\nexample.net"},
		{"ipcidr", "192.0.2.0/24\n2001:db8::/32"},
	} {
		t.Run(fixture.behavior, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Render.CustomRules = []model.CustomRule{{
				Key: "valid_source", Label: "Valid source", Enabled: true,
				SourceType: "http", Behavior: fixture.behavior, Format: "text",
				URL: "https://example.com/rules.txt",
			}}
			stub := &customRuleFetcherStub{fetched: fetcher.FetchedSubscription{Content: []byte(fixture.content), FromCache: true}}
			got, _, err := snapshotRemoteCustomRulesWithFetcher(cfg, stub)
			if err != nil || len(got.Render.CustomRules[0].Payload) == 0 {
				t.Fatalf("valid %s snapshot rejected: %v", fixture.behavior, err)
			}
		})
	}
}
