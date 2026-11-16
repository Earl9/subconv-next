package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"subconv-next/internal/model"
)

func TestAdditionalRulesValidationReportsOriginalLineWithoutChangingSavedConfig(t *testing.T) {
	server, cfg := newTestServer(t, model.DefaultConfig())
	ref := createWorkspaceRefForTest(t, server, cfg)
	before, err := os.ReadFile(ref.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	text := "# comment\n\nDOMAIN-SUFFIX,example.com,DIRECT\nIP-CIDR,10.0.0.0/99,DIRECT"
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{AdditionalRulesText: &text}, http.StatusBadRequest)
	var response customRuleValidationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.OK || response.Error == nil || response.Error.Field != "additional_rules" || response.Error.Line != 4 {
		t.Fatalf("direct rule error lost line: %s", data)
	}
	text = "DOMAIN-SUFFIX,example.com,DIRECT\nIP-CIDR,10.0.0.0/8,🚀 节点选择,no-resolve"
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{AdditionalRulesText: &text}, http.StatusOK)
	text = ""
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{AdditionalRulesText: &text}, http.StatusOK)
	after, err := os.ReadFile(ref.ConfigPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("direct rule validation changed saved config")
	}
}

func TestLegacyAdditionalRulesCanBeEditedButInvalidSaveDoesNotOverwrite(t *testing.T) {
	server, cfg := newTestServer(t, model.DefaultConfig())
	ref := createWorkspaceRefForTest(t, server, cfg)
	stored, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	stored.Render.AdditionalRules = []string{"DOMAIN-SUFFIX,example.com"}
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
	if len(response.Config.Render.AdditionalRules) != 1 || response.Config.Render.AdditionalRules[0] != stored.Render.AdditionalRules[0] {
		t.Fatal("editing did not preserve incomplete direct rule")
	}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusBadRequest)
	after, err := os.ReadFile(ref.ConfigPath)
	if err != nil || !bytes.Equal(legacy, after) {
		t.Fatal("bad direct rules overwrote existing config")
	}
	response.Config.Render.AdditionalRules[0] += ",DIRECT"
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusOK)
}

func TestAdditionalRulesValidationHonorsPublicPayloadLimit(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.PublicConverter = true
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	text := strings.Repeat("DOMAIN-SUFFIX,example.com,DIRECT\n", publicMaxRulePayloadItems+1)
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/custom-rules/validate", customRuleValidationRequest{AdditionalRulesText: &text}, http.StatusBadRequest)
	var response customRuleValidationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.Error == nil || response.Error.Field != "additional_rules" {
		t.Fatalf("public direct-rule limit bypassed: %s", data)
	}
}
