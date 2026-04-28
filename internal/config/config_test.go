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
			RefreshOnRequest:     true,
			StaleIfError:         true,
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
			MixedPort:         7890,
			AllowLAN:          false,
			Mode:              "rule",
			LogLevel:          "info",
			IPv6:              false,
			DNSEnabled:        true,
			EnhancedMode:      "fake-ip",
			Emoji:             true,
			ShowNodeType:      true,
			IncludeInfoNode:   true,
			UDP:               true,
			FilterIllegal:     true,
			GeodataMode:       true,
			GeoAutoUpdate:     true,
			GeodataLoader:     "standard",
			GeoUpdateInterval: 24,
			OutputFilename:    "mihomo.yaml",
			ExternalConfig: model.ExternalConfig{
				TemplateKey:   "none",
				TemplateLabel: "不选择，由接口提供方提供",
				CustomURL:     "",
			},
			GeoxURL: &model.GeoxURLConfig{
				GeoIP:   "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geoip.dat",
				GeoSite: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geosite.dat",
				MMDB:    "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/country.mmdb",
				ASN:     "https://github.com/xishang0128/geoip/releases/download/latest/GeoLite2-ASN.mmdb",
			},
			TemplateRuleMode: "rules",
			RuleMode:         "custom",
			EnabledRules:     []string{},
			CustomRules:      []model.CustomRule{},
			DNS: &model.DNSConfig{
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
					"https://000000.dns.nextdns.io/dns-query#h3=true",
					"https://dns.alidns.com/dns-query",
					"https://doh.pub/dns-query",
					"https://public.dns.iij.jp/dns-query",
					"https://101.101.101.101/dns-query",
					"https://208.67.220.220/dns-query",
					"tls://8.8.4.4",
					"tls://1.0.0.1:853",
					"https://cloudflare-dns.com/dns-query",
					"https://dns.google/dns-query",
				},
			},
			Profile: &model.ProfileConfig{
				StoreSelected: true,
				StoreFakeIP:   false,
			},
			Sniffer: &model.SnifferConfig{
				Enable:      true,
				ParsePureIP: true,
				HTTP: &model.SniffHTTP{
					Ports:               []string{"80", "8080-8880"},
					OverrideDestination: true,
				},
				QUIC: &model.SniffProtocol{
					Ports: []string{"443", "8443"},
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
	if jsonCfg.Render.GeodataMode != true || jsonCfg.Render.GeoAutoUpdate != true || jsonCfg.Render.GeodataLoader != "standard" || jsonCfg.Render.GeoUpdateInterval != 24 {
		t.Fatalf("geodata defaults mismatch: %+v", jsonCfg.Render)
	}
	if !jsonCfg.Render.SourcePrefix || jsonCfg.Render.SourcePrefixFormat != "[{source}] {name}" || jsonCfg.Render.DedupeScope != "global" {
		t.Fatalf("source prefix defaults mismatch: %+v", jsonCfg.Render)
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
}
