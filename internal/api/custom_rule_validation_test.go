package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"subconv-next/internal/model"
)

func TestCustomRuleValidationReportsOriginalLineAndDoesNotMutateOrFetch(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1) }))
	defer upstream.Close()
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@keep.example:443#keep"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	pub, err := server.loadPublishedByID(published.PublishID)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{ref.ConfigPath, ref.StatePath, pub.CurrentPath, pub.MetaPath}
	before := make(map[string][]byte)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before[path] = data
	}
	base := model.CustomRule{Key: "test", Label: "Test", SourceType: "inline", Behavior: "classical", Format: "text", TargetMode: "direct", Payload: []string{"DOMAIN-SUFFIX,example.com"}}
	text := "# first line\n\nDOMAIN-SUFFIX,example.comDOMAIN,example.org\nIP-CIDR,10.0.0.0/99"
	payload := customRuleValidationRequest{Rule: base, PayloadText: &text}
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", payload, http.StatusBadRequest)
	var response customRuleValidationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.OK || response.Error == nil || response.Error.Field != "payload" || response.Error.Line != 4 || !strings.Contains(response.Error.Message, "第 4 行") {
		t.Fatalf("validation did not locate original line: %s", data)
	}
	text = "DOMAIN-SUFFIX,example.com\nIP-CIDR,10.0.0.0/8,no-resolve"
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{Rule: base, PayloadText: &text}, http.StatusOK)
	base.SourceType, base.URL, base.Interval = "http", upstream.URL+"/rules?token=private", 3600
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{Rule: base}, http.StatusOK)
	if requests.Load() != 0 {
		t.Fatal("draft validation fetched remote content")
	}
	for _, path := range paths {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before[path], after) {
			t.Fatalf("validation changed %s: %v", path, err)
		}
	}
}

func TestCustomRuleValidationRejectsMalformedFieldsAndRequiresWorkspace(t *testing.T) {
	server, cfg := newTestServer(t, model.DefaultConfig())
	ref := createWorkspaceRefForTest(t, server, cfg)
	rule := model.CustomRule{Key: "test", Label: "Test", SourceType: "http", URL: "https://", Interval: 3600}
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{Rule: rule}, http.StatusBadRequest)
	var response customRuleValidationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.Error == nil || response.Error.Field != "url" {
		t.Fatalf("missing URL error: %s", data)
	}
	rule.URL, rule.Interval = "https://example.com/rules", -1
	data = nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{Rule: rule}, http.StatusBadRequest)
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.Error == nil || response.Error.Field != "interval" {
		t.Fatalf("negative interval was normalized away: %s", data)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/custom-rules/validate", strings.NewReader(`{"rule":{}}`))
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "WORKSPACE_REQUIRED") {
		t.Fatalf("missing workspace accepted: %d %s", rec.Code, rec.Body.String())
	}
}

func TestConfigSaveRejectsInvalidCustomRuleWithoutChangingStoredConfig(t *testing.T) {
	server, cfg := newTestServer(t, model.DefaultConfig())
	ref := createWorkspaceRefForTest(t, server, cfg)
	before, err := os.ReadFile(ref.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Render.CustomRules = []model.CustomRule{{Key: "bad", Label: "Bad", Enabled: true, SourceType: "inline", Behavior: "classical", Payload: []string{"DST-PORT,90000"}}}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", cfg, http.StatusBadRequest)
	after, err := os.ReadFile(ref.ConfigPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("invalid rule overwrote saved configuration")
	}
}

func TestCustomRuleValidationAppliesPublicLimitsAndRejectsInvalidRequests(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.PublicConverter = true
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	rule := model.CustomRule{Key: "test", Label: "Test", SourceType: "inline", Payload: []string{"example.com"}}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{Rule: rule}, http.StatusOK)
	text := strings.Repeat("example.com\n", publicMaxRulePayloadItems+1)
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{Rule: rule, PayloadText: &text}, http.StatusBadRequest)
	var response customRuleValidationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.Error == nil || response.Error.Field != "payload" {
		t.Fatalf("public payload limit bypassed: %s", data)
	}
	for _, entry := range []struct {
		method, body string
		status       int
	}{
		{http.MethodGet, "", http.StatusMethodNotAllowed},
		{http.MethodPost, `{"unknown":true}`, http.StatusBadRequest},
		{http.MethodPost, `{"rule":`, http.StatusBadRequest},
	} {
		req := httptest.NewRequest(entry.method, withWorkspace("/api/custom-rules/validate", ref.ID), strings.NewReader(entry.body))
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		if rec.Code != entry.status {
			t.Fatalf("invalid request status=%d want=%d", rec.Code, entry.status)
		}
	}
}

func TestLegacyInvalidCustomRulesCanBeReadAndCorrectedButCannotBeSavedUnchanged(t *testing.T) {
	server, cfg := newTestServer(t, model.DefaultConfig())
	ref := createWorkspaceRefForTest(t, server, cfg)
	stored, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	stored.Render.CustomRules = []model.CustomRule{{Key: "legacy", Label: "Legacy", Enabled: true, SourceType: "inline", Behavior: "classical", Format: "text", Payload: []string{"DOMAIN-SUFFIX,example.com", "DOMAIN-SUFFIX,"}}}
	legacy, err := json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ref.ConfigPath, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/config", nil, http.StatusOK)
	var response configResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	payload := response.Config.Render.CustomRules[0].Payload
	if len(payload) != 2 || payload[1] != "DOMAIN-SUFFIX," {
		t.Fatal("reading the old config dropped the malformed rule instead of keeping it editable")
	}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusBadRequest)
	unchanged, err := os.ReadFile(ref.ConfigPath)
	if err != nil || !bytes.Equal(legacy, unchanged) {
		t.Fatal("rejected legacy rule changed stored config")
	}
	response.Config.Render.CustomRules[0].Payload[1] = "DOMAIN-SUFFIX,example.org"
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusOK)
	corrected, err := server.loadWorkspaceConfig(ref)
	if err != nil || corrected.Render.CustomRules[0].Payload[1] != "DOMAIN-SUFFIX,example.org" {
		t.Fatalf("corrected legacy rule was not saved: %v", err)
	}
}
