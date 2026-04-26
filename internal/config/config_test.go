package config

import (
	"path/filepath"
	"reflect"
	"testing"

	"subconv-next/internal/model"
)

func TestLoadJSONAndUCIParity(t *testing.T) {
	jsonPath := filepath.Join("..", "..", "testdata", "config", "basic.json")
	uciPath := filepath.Join("..", "..", "testdata", "config", "basic.uci")

	jsonCfg, err := Load(jsonPath)
	if err != nil {
		t.Fatalf("Load(%q) error = %v", jsonPath, err)
	}

	uciCfg, err := Load(uciPath)
	if err != nil {
		t.Fatalf("Load(%q) error = %v", uciPath, err)
	}

	if !reflect.DeepEqual(jsonCfg, uciCfg) {
		t.Fatalf("JSON and UCI configs differ\njson=%+v\nuci=%+v", jsonCfg, uciCfg)
	}

	want := model.Config{
		Service: model.ServiceConfig{
			Enabled:              true,
			ListenAddr:           "127.0.0.1",
			ListenPort:           9876,
			LogLevel:             "info",
			Template:             "standard",
			OutputPath:           "/data/mihomo.yaml",
			CacheDir:             "/data/cache",
			StatePath:            "/data/state.json",
			RefreshInterval:      1800,
			MaxSubscriptionBytes: 5242880,
			FetchTimeoutSeconds:  15,
			AllowLAN:             false,
		},
		Subscriptions: []model.SubscriptionConfig{
			{
				Name:               "example",
				Enabled:            true,
				URL:                "https://example.com/sub",
				UserAgent:          model.DefaultUserAgent,
				InsecureSkipVerify: false,
			},
		},
		Inline: []model.InlineConfig{
			{
				Name:    "manual",
				Enabled: true,
				Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#manual",
			},
		},
		Render: model.RenderConfig{
			MixedPort:    7890,
			AllowLAN:     false,
			Mode:         "rule",
			LogLevel:     "info",
			IPv6:         false,
			DNSEnabled:   true,
			EnhancedMode: "fake-ip",
		},
	}

	if !reflect.DeepEqual(jsonCfg, want) {
		t.Fatalf("parsed config mismatch\n got=%+v\nwant=%+v", jsonCfg, want)
	}
}

func TestParseUCISectionsSupportsCommentsQuotesAndLists(t *testing.T) {
	data := []byte(`
# comment
config render "mihomo"
	option mode rule
	list names 'alpha'
	list names "beta"
`)

	sections, err := parseUCISections(data)
	if err != nil {
		t.Fatalf("parseUCISections() error = %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("section count = %d, want 1", len(sections))
	}

	section := sections[0]
	if section.Type != "render" {
		t.Fatalf("section.Type = %q, want %q", section.Type, "render")
	}
	if section.Name != "mihomo" {
		t.Fatalf("section.Name = %q, want %q", section.Name, "mihomo")
	}
	if got := section.Options["mode"]; !reflect.DeepEqual(got, []string{"rule"}) {
		t.Fatalf("mode values = %#v, want %#v", got, []string{"rule"})
	}
	if got := section.Options["names"]; !reflect.DeepEqual(got, []string{"alpha", "beta"}) {
		t.Fatalf("names values = %#v, want %#v", got, []string{"alpha", "beta"})
	}
}

func TestLoadJSONBytesAppliesEntryDefaults(t *testing.T) {
	data := []byte(`{
  "subscriptions": [
    {
      "name": "example",
      "url": "https://example.com/sub"
    }
  ],
  "inline": [
    {
      "name": "manual",
      "content": "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#manual"
    }
  ]
}`)

	cfg, err := LoadJSONBytes(data)
	if err != nil {
		t.Fatalf("LoadJSONBytes() error = %v", err)
	}

	if len(cfg.Subscriptions) != 1 {
		t.Fatalf("len(Subscriptions) = %d, want 1", len(cfg.Subscriptions))
	}
	if !cfg.Subscriptions[0].Enabled {
		t.Fatalf("Subscriptions[0].Enabled = false, want true")
	}
	if cfg.Subscriptions[0].UserAgent != model.DefaultUserAgent {
		t.Fatalf("Subscriptions[0].UserAgent = %q, want %q", cfg.Subscriptions[0].UserAgent, model.DefaultUserAgent)
	}

	if len(cfg.Inline) != 1 {
		t.Fatalf("len(Inline) = %d, want 1", len(cfg.Inline))
	}
	if !cfg.Inline[0].Enabled {
		t.Fatalf("Inline[0].Enabled = false, want true")
	}
}
