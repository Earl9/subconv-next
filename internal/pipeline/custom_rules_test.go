package pipeline

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
)

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
