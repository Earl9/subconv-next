package config

import (
	"errors"
	"strings"
	"testing"

	"subconv-next/internal/model"
)

func TestValidateCustomRuleRejectsInvalidFieldsAndPayload(t *testing.T) {
	valid := model.CustomRule{Key: "test", Label: "Test", Enabled: true, TargetMode: "direct", SourceType: "inline", Behavior: "domain", Format: "text", Payload: []string{"example.com"}, InsertPosition: "before_match"}
	cases := []struct {
		name, field string
		mutate      func(*model.CustomRule)
	}{
		{"invalid key", "key", func(r *model.CustomRule) { r.Key = "bad key" }},
		{"policy punctuation", "label", func(r *model.CustomRule) { r.Label = "Bad,Group" }},
		{"missing group", "target_group", func(r *model.CustomRule) { r.TargetMode = "existing_group" }},
		{"invalid URL", "url", func(r *model.CustomRule) { r.SourceType, r.URL, r.Interval = "http", "https://", 3600 }},
		{"URL credentials", "url", func(r *model.CustomRule) {
			r.SourceType, r.URL, r.Interval = "http", "https://user:secret@example.com/rules", 3600
		}},
		{"invalid port", "url", func(r *model.CustomRule) {
			r.SourceType, r.URL, r.Interval = "http", "https://example.com:99999/rules", 3600
		}},
		{"negative interval", "interval", func(r *model.CustomRule) { r.SourceType, r.URL, r.Interval = "http", "https://example.com/rules", -1 }},
		{"inline mrs", "format", func(r *model.CustomRule) { r.Format = "mrs" }},
		{"invalid CIDR", "payload", func(r *model.CustomRule) { r.Behavior, r.Payload = "ipcidr", []string{"192.168.0.0/99"} }},
		{"missing value", "payload", func(r *model.CustomRule) { r.Behavior, r.Payload = "classical", []string{"DOMAIN-SUFFIX,"} }},
		{"unknown type", "payload", func(r *model.CustomRule) { r.Behavior, r.Payload = "classical", []string{"DOMAIN-SUFIX,example.com"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := valid
			tc.mutate(&rule)
			err := ValidateCustomRule(rule)
			var detail *CustomRuleValidationError
			if !errors.As(err, &detail) || detail.Field != tc.field {
				t.Fatalf("error = %v; want field %s", err, tc.field)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("validation exposed URL credentials")
			}
		})
	}
}

func TestCustomRuleValidationPreservesInputLinesAndClassicalInference(t *testing.T) {
	rule := model.CustomRule{Key: "test", Label: "Test", TargetMode: "direct", SourceType: "inline", Behavior: "domain", Format: "text", InsertPosition: "before_match", Payload: []string{"# comment", "", "DOMAIN-SUFFIX,example.comDOMAIN,example.org", "IP-CIDR,10.0.0.0/99"}}
	var detail *CustomRuleValidationError
	if err := ValidateCustomRule(rule); !errors.As(err, &detail) || detail.Field != "payload" || detail.Line != 4 {
		t.Fatalf("error = %v; want payload line 4", err)
	}
	rule.Payload[3] = "IP-CIDR,10.0.0.0/8,no-resolve"
	if err := ValidateCustomRule(rule); err != nil {
		t.Fatalf("valid inferred classical rule rejected: %v", err)
	}
}

func TestConfigValidationRejectsInvalidCustomRulesAfterNormalization(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{Key: "bad", Label: "Bad", Enabled: true, SourceType: "inline", Behavior: "classical", Payload: []string{"DST-PORT,70000"}}}
	if err := Validate(Normalize(cfg)); err == nil || !strings.Contains(err.Error(), "custom_rules[0]") {
		t.Fatalf("invalid custom payload accepted: %v", err)
	}
	cfg.Render.CustomRules[0] = model.CustomRule{Key: "remote", Label: "Remote", SourceType: "http", URL: "https://example.com/rules", Interval: -1}
	if err := Validate(Normalize(cfg)); err == nil || !strings.Contains(err.Error(), "interval") {
		t.Fatalf("negative custom interval was hidden by normalization: %v", err)
	}
	cfg.Render.CustomRules[0].Interval = 0
	if err := Validate(Normalize(cfg)); err != nil {
		t.Fatalf("omitted interval default no longer works: %v", err)
	}
}

func TestNormalizePromotesNewClassicalSyntaxBeforeRendering(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{Key: "modern", Label: "Modern", SourceType: "inline", Behavior: "domain", Payload: []string{"IN-PORT,7890", "DSCP,8", "PROCESS-NAME-WILDCARD,chrome*"}}}
	normalized := Normalize(cfg)
	if normalized.Render.CustomRules[0].Behavior != "classical" {
		t.Fatal("new classical rule types would be rendered as domain payload")
	}
	if err := Validate(normalized); err != nil {
		t.Fatalf("valid modern classical syntax was rejected: %v", err)
	}
}

func TestPastedMultilineCustomRulesDoNotHideInvalidFollowingLines(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomRules = []model.CustomRule{{Key: "pasted", Label: "Pasted", SourceType: "inline", Behavior: "classical", Payload: []string{"# comment\n- 'DOMAIN,example.com' # note\nIP-CIDR,192.0.2.0/99"}}}
	// Validate the original input first so the UI can highlight its physical line.
	rule := cfg.Render.CustomRules[0]
	rule.TargetMode, rule.Format, rule.InsertPosition = "direct", "text", "before_match"
	var detail *CustomRuleValidationError
	if err := ValidateCustomRule(rule); !errors.As(err, &detail) || detail.Line != 3 {
		t.Fatalf("error=%v, want payload line 3", err)
	}
	if err := Validate(Normalize(cfg)); err == nil {
		t.Fatal("normalization hid an invalid rule after a leading comment")
	}
}
