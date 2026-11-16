package config

import (
	"encoding/json"
	"reflect"
	"testing"

	"subconv-next/internal/model"
)

func TestEditingJSONKeepsLegacyInvalidCustomRulesWithoutWeakeningStrictLoad(t *testing.T) {
	cfg := model.DefaultConfig()
	payload := []string{"DOMAIN-SUFFIX,example.com", "DOMAIN-SUFFIX,"}
	cfg.Render.CustomRules = []model.CustomRule{{Key: "legacy", Label: "Legacy", Enabled: true, SourceType: "inline", Behavior: "classical", Payload: payload}}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJSONBytes(data); err == nil {
		t.Fatal("strict config load accepted an invalid legacy rule")
	}
	loaded, err := LoadJSONBytesForEditing(data)
	if err != nil {
		t.Fatalf("legacy configuration cannot be edited: %v", err)
	}
	if !reflect.DeepEqual(loaded.Render.CustomRules[0].Payload, payload) {
		t.Fatalf("legacy payload changed: %#v", loaded.Render.CustomRules[0].Payload)
	}
	if err := Validate(loaded); err == nil {
		t.Fatal("editing load weakened save validation")
	}
	cfg.Service.ListenPort = -1
	data, err = json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJSONBytesForEditing(data); err == nil {
		t.Fatal("editing config skipped validation outside custom rules")
	}
}
