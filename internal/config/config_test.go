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
			Enabled:                         true,
			ListenAddr:                      "127.0.0.1",
			ListenPort:                      9876,
			LogLevel:                        "info",
			Template:                        "standard",
			OutputPath:                      "/data/mihomo.yaml",
			CacheDir:                        "/data/cache",
			StatePath:                       "/data/state.json",
			RefreshInterval:                 1800,
			RefreshOnRequest:                true,
			StaleIfError:                    true,
			StrictMode:                      true,
			WorkspaceTTLSeconds:             86400,
			WorkspaceCleanupIntervalSeconds: 3600,
			WorkspaceCleanupInterval:        3600,
			PublicBaseURL:                   "",
			MaxSubscriptionBytes:            5242880,
			FetchTimeoutSeconds:             15,
			AllowLAN:                        false,
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
			MixedPort:       model.DefaultMixedPort,
			AllowLAN:        true,
			Mode:            "rule",
			LogLevel:        "info",
			IPv6:            false,
			DNSEnabled:      true,
			EnhancedMode:    "fake-ip",
			Emoji:           false,
			ShowNodeType:    false,
			IncludeInfoNode: false,
			ShowInfoNodes:   false,
			UDP:             true,
			FilterIllegal:   true,
			OutputFilename:  "mihomo.yaml",
			ExternalConfig: model.ExternalConfig{
				TemplateKey:   "none",
				TemplateLabel: "不选择，由接口提供方提供",
				CustomURL:     "",
			},
			TemplateRuleMode: "rules",
			RuleMode:         "custom",
			EnabledRules:     []string{},
			CustomRules:      []model.CustomRule{},
			DNS: &model.DNSConfig{
				Enable:         true,
				Listen:         "127.0.0.1:5335",
				UseHosts:       true,
				UseSystemHosts: false,
				RespectRules:   false,
				EnhancedMode:   "fake-ip",
				FakeIPRange:    "198.18.0.1/16",
				DefaultNameserver: []string{
					"223.5.5.5",
					"119.29.29.29",
				},
				Nameserver: []string{
					"223.5.5.5",
					"119.29.29.29",
				},
				Fallback: []string{
					"https://dns.alidns.com/dns-query",
				},
				FakeIPFilter: []string{
					"*.lan",
					"*.local",
					"localhost",
					"*.msftconnecttest.com",
					"*.msftncsi.com",
					"*.nintendo.net",
					"*.playstation.net",
					"*.xboxlive.com",
					"stun.*",
					"time.*",
				},
			},
			Profile: &model.ProfileConfig{
				StoreSelected: true,
				StoreFakeIP:   false,
			},
			Sniffer: &model.SnifferConfig{
				Enable:      true,
				ParsePureIP: false,
				HTTP: &model.SniffHTTP{
					Ports:               []string{"80", "8080-8880"},
					OverrideDestination: true,
				},
				TLS: &model.SniffProtocol{
					Ports: []string{"443", "8443"},
				},
			},
			AdditionalRules:   []string{},
			RuleProviders:     []model.RuleProviderConfig{},
			CustomProxyGroups: []model.CustomProxyGroupConfig{},
		},
	}

	if !reflect.DeepEqual(jsonCfg.Service, want.Service) {
		t.Fatalf("service mismatch\n got=%+v\nwant=%+v", jsonCfg.Service, want.Service)
	}
	if len(jsonCfg.Subscriptions) != 1 {
		t.Fatalf("subscription count = %d, want 1", len(jsonCfg.Subscriptions))
	}
	if jsonCfg.Subscriptions[0].ID == "" || jsonCfg.Subscriptions[0].Name != want.Subscriptions[0].Name || jsonCfg.Subscriptions[0].URL != want.Subscriptions[0].URL || jsonCfg.Subscriptions[0].UserAgent != want.Subscriptions[0].UserAgent {
		t.Fatalf("subscriptions mismatch\n got=%+v\nwant=%+v", jsonCfg.Subscriptions, want.Subscriptions)
	}
	if len(jsonCfg.Inline) != 1 {
		t.Fatalf("inline count = %d, want 1", len(jsonCfg.Inline))
	}
	if jsonCfg.Inline[0].ID == "" || jsonCfg.Inline[0].Name != want.Inline[0].Name || jsonCfg.Inline[0].Content != want.Inline[0].Content {
		t.Fatalf("inline mismatch\n got=%+v\nwant=%+v", jsonCfg.Inline, want.Inline)
	}
	if jsonCfg.Render.GeodataMode || jsonCfg.Render.GeoAutoUpdate || jsonCfg.Render.GeodataLoader != "" || jsonCfg.Render.GeoUpdateInterval != 0 || jsonCfg.Render.GeoxURL != nil {
		t.Fatalf("geodata defaults should be disabled: %+v", jsonCfg.Render)
	}
	if !jsonCfg.Render.SourcePrefix || jsonCfg.Render.SourcePrefixFormat != "{emoji} {name}" || jsonCfg.Render.DedupeScope != "global" {
		t.Fatalf("source prefix defaults mismatch: %+v", jsonCfg.Render)
	}
	if jsonCfg.Render.NameOptions.SourcePrefixMode != "emoji_name" || jsonCfg.Render.NameOptions.SourcePrefixSeparator != "｜" {
		t.Fatalf("name option defaults mismatch: %+v", jsonCfg.Render.NameOptions)
	}
	if jsonCfg.Render.DNS == nil || jsonCfg.Render.DNS.Listen != "127.0.0.1:5335" {
		t.Fatalf("dns defaults missing: %+v", jsonCfg.Render.DNS)
	}
	if jsonCfg.Render.Profile == nil || !jsonCfg.Render.Profile.StoreSelected {
		t.Fatalf("profile defaults missing: %+v", jsonCfg.Render.Profile)
	}
	if jsonCfg.Render.Sniffer == nil || !jsonCfg.Render.Sniffer.Enable {
		t.Fatalf("sniffer defaults missing: %+v", jsonCfg.Render.Sniffer)
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
	if cfg.Subscriptions[0].ID == "" {
		t.Fatalf("Subscriptions[0].ID is empty")
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
	if cfg.Inline[0].ID == "" {
		t.Fatalf("Inline[0].ID is empty")
	}
}

func TestLoadJSONBytesDeduplicatesSubscriptionIDs(t *testing.T) {
	data := []byte(`{
  "subscriptions": [
    {
      "id": "source-1",
      "name": "main",
      "url": "https://example.com/a"
    },
    {
      "id": "source-1",
      "name": "backup",
      "url": "https://example.com/b"
    }
  ]
}`)

	cfg, err := LoadJSONBytes(data)
	if err != nil {
		t.Fatalf("LoadJSONBytes() error = %v", err)
	}
	if len(cfg.Subscriptions) != 2 {
		t.Fatalf("len(Subscriptions) = %d, want 2", len(cfg.Subscriptions))
	}
	if cfg.Subscriptions[0].ID != "source-1" {
		t.Fatalf("Subscriptions[0].ID = %q, want source-1", cfg.Subscriptions[0].ID)
	}
	if cfg.Subscriptions[1].ID == "" || cfg.Subscriptions[1].ID == cfg.Subscriptions[0].ID {
		t.Fatalf("subscription IDs were not deduplicated: %#v", cfg.Subscriptions)
	}
}

func TestNormalizeMigratesLegacyComplexDefaultDNS(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.DNS = &model.DNSConfig{
		Enable:         true,
		Listen:         "127.0.0.1:5335",
		UseSystemHosts: false,
		EnhancedMode:   "fake-ip",
		FakeIPRange:    "198.18.0.1/16",
		DefaultNameserver: []string{
			"180.76.76.76",
			"182.254.118.118",
			"8.8.8.8",
			"180.184.2.2",
		},
		Nameserver: []string{
			"180.76.76.76",
			"119.29.29.29",
			"180.184.1.1",
			"223.5.5.5",
			"8.8.8.8",
			"https://223.6.6.6/dns-query#h3=true",
			"https://dns.alidns.com/dns-query",
			"https://cloudflare-dns.com/dns-query",
			"https://doh.pub/dns-query",
		},
		Fallback: []string{
			"https://dns.google/dns-query",
		},
		FallbackFilter: &model.DNSFallbackFilter{GeoIP: true},
	}

	got := Normalize(cfg)
	if got.Render.DNS == nil {
		t.Fatalf("Render.DNS = nil")
	}
	if !reflect.DeepEqual(got.Render.DNS.DefaultNameserver, []string{"223.5.5.5", "119.29.29.29"}) {
		t.Fatalf("DefaultNameserver = %#v, want compact defaults", got.Render.DNS.DefaultNameserver)
	}
	if !reflect.DeepEqual(got.Render.DNS.Nameserver, []string{"223.5.5.5", "119.29.29.29"}) {
		t.Fatalf("Nameserver = %#v, want updated defaults", got.Render.DNS.Nameserver)
	}
	if !reflect.DeepEqual(got.Render.DNS.Fallback, []string{"https://dns.alidns.com/dns-query"}) {
		t.Fatalf("Fallback = %#v, want updated defaults", got.Render.DNS.Fallback)
	}
	if got.Render.DNS.FallbackFilter != nil {
		t.Fatalf("fallback filter must be disabled after migration: %+v", got.Render.DNS.FallbackFilter)
	}
	if got.Render.DNS.NameserverPolicy != nil {
		t.Fatalf("nameserver-policy must be disabled after migration: %+v", got.Render.DNS.NameserverPolicy)
	}
}

func TestNormalizePreservesCustomDNS(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.CustomDNS = true
	cfg.Render.DNS = &model.DNSConfig{
		Enable:         true,
		Listen:         "0.0.0.0:1053",
		UseHosts:       false,
		UseSystemHosts: true,
		RespectRules:   true,
		EnhancedMode:   "redir-host",
		FakeIPRange:    "198.18.0.1/16",
		DefaultNameserver: []string{
			"1.1.1.1",
		},
		Nameserver: []string{
			"https://dns.google/dns-query",
		},
		ProxyNameserver: []string{
			"223.5.5.5",
		},
		DirectNameserver: []string{
			"https://doh.pub/dns-query",
		},
		DirectFollowPolicy: true,
		Fallback: []string{
			"https://dns.cloudflare.com/dns-query",
		},
		FakeIPFilter: []string{
			"*.lan",
		},
		NameserverPolicy: map[string][]string{
			"geosite:cn,private": {
				"https://dns.alidns.com/dns-query",
			},
		},
	}

	got := Normalize(cfg)
	if !got.Render.CustomDNS {
		t.Fatalf("CustomDNS = false, want true")
	}
	if got.Render.DNS == nil {
		t.Fatalf("Render.DNS = nil")
	}
	if got.Render.DNS.Listen != "0.0.0.0:1053" {
		t.Fatalf("Listen = %q, want custom value", got.Render.DNS.Listen)
	}
	if got.Render.DNS.UseHosts {
		t.Fatalf("UseHosts = true, want false")
	}
	if !got.Render.DNS.UseSystemHosts || !got.Render.DNS.RespectRules {
		t.Fatalf("custom DNS booleans were not preserved: %+v", got.Render.DNS)
	}
	if !reflect.DeepEqual(got.Render.DNS.Nameserver, []string{"https://dns.google/dns-query"}) {
		t.Fatalf("Nameserver = %#v, want custom resolver", got.Render.DNS.Nameserver)
	}
	if !reflect.DeepEqual(got.Render.DNS.Fallback, []string{"https://dns.cloudflare.com/dns-query"}) {
		t.Fatalf("Fallback = %#v, want custom fallback", got.Render.DNS.Fallback)
	}
	if got.Render.DNS.NameserverPolicy["geosite:cn,private"][0] != "https://dns.alidns.com/dns-query" {
		t.Fatalf("NameserverPolicy = %#v, want custom policy", got.Render.DNS.NameserverPolicy)
	}
}

func TestValidateSubscriptionNameAndURL(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg = Normalize(cfg)
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-1", Name: "", Enabled: true, URL: "https://example.com/sub", UserAgent: model.DefaultUserAgent},
	}
	if err := Validate(cfg); err == nil {
		t.Fatalf("Validate() error = nil, want name required")
	}

	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-1", Name: "main", Enabled: true, URL: "ftp://example.com/sub", UserAgent: model.DefaultUserAgent},
	}
	if err := Validate(cfg); err == nil {
		t.Fatalf("Validate() error = nil, want http/https restriction")
	}

	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-1", Name: "main", Enabled: true, URL: "https://example.com/a", UserAgent: model.DefaultUserAgent},
		{ID: "sub-1", Name: "backup", Enabled: true, URL: "https://example.com/b", UserAgent: model.DefaultUserAgent},
	}
	if err := Validate(cfg); err == nil {
		t.Fatalf("Validate() error = nil, want duplicate subscription id restriction")
	}
}
