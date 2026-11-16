package config

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"subconv-next/internal/model"
)

func TestAdditionalRuleValidationRequiresCompleteRulesAndPreservesLine(t *testing.T) {
	for _, rules := range [][]string{
		nil,
		{"# a comment", "", "DOMAIN-SUFFIX,example.com,DIRECT", "IP-CIDR,192.0.2.0/24,🚀 节点选择,no-resolve"},
		{"RULE-SET,example-provider,🚀 节点选择"},
	} {
		if err := ValidateAdditionalRules(rules); err != nil {
			t.Fatalf("valid direct rules rejected: %v", err)
		}
	}
	for _, malformed := range []string{"DOMAIN-SUFFIX,example.com", "IP-CIDR,192.0.2.0/99,DIRECT", "DST-PORT,99999,REJECT", "MATCH,DIRECT"} {
		var detail *CustomRuleValidationError
		err := ValidateAdditionalRules([]string{"# comment\n", malformed})
		if !errors.As(err, &detail) || detail.Field != "additional_rules" || detail.Line != 3 {
			t.Fatalf("invalid direct rule %q: error=%v want additional_rules line 3", malformed, err)
		}
	}
}

func TestEditingJSONKeepsLegacyAdditionalRulesButStrictValidationRejectsThem(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.AdditionalRules = []string{"DOMAIN-SUFFIX,example.com"}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJSONBytes(data); err == nil {
		t.Fatal("strict load accepted a direct rule with no target")
	}
	loaded, err := LoadJSONBytesForEditing(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Render.AdditionalRules, cfg.Render.AdditionalRules) {
		t.Fatal("editing lost malformed direct rules")
	}
	if err := Validate(loaded); err == nil {
		t.Fatal("malformed editing rules can still be saved")
	}
	loaded.Render.AdditionalRules[0] = "DOMAIN-SUFFIX,example.com,DIRECT"
	if err := Validate(loaded); err != nil {
		t.Fatalf("corrected rule cannot be saved: %v", err)
	}
}
