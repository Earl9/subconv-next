package renderer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/parser"
)

func TestRenderGoldenLiteBasic(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:     "lite",
		MixedPort:    7890,
		Mode:         "rule",
		LogLevel:     "info",
		DNSEnabled:   true,
		EnhancedMode: "fake-ip",
	})
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "lite-basic.yaml"), got)
}

func TestRenderGoldenStandardVLESSReality(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "vless-reality.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-vless-reality.yaml"), got)
}

func TestRenderGoldenStandardHy2TUIC(t *testing.T) {
	nodes := append(
		mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "hy2.txt")),
		mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "tuic-v5.txt"))...,
	)
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-hy2-tuic.yaml"), got)
}

func TestRenderGoldenStandardAnyTLS(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "anytls.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-anytls.yaml"), got)
}

func TestRenderGoldenStandardWireGuard(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "wireguard-uri.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-wireguard.yaml"), got)
}

func TestRenderGoldenDedupeRenamed(t *testing.T) {
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{
			Name:   "dup",
			Type:   model.ProtocolTrojan,
			Server: "one.example.com",
			Port:   443,
			Auth:   model.Auth{Password: "a"},
			TLS:    model.TLSOptions{Enabled: true, SNI: "one.example.com"},
			UDP:    model.Bool(true),
		}),
		model.NormalizeNode(model.NodeIR{
			Name:   "dup",
			Type:   model.ProtocolTrojan,
			Server: "two.example.com",
			Port:   443,
			Auth:   model.Auth{Password: "b"},
			TLS:    model.TLSOptions{Enabled: true, SNI: "two.example.com"},
			UDP:    model.Bool(true),
		}),
	}

	got := mustRender(t, nodes, model.RenderOptions{
		Template:   "lite",
		MixedPort:  7890,
		Mode:       "rule",
		LogLevel:   "info",
		DNSEnabled: false,
	})
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "dedupe-renamed.yaml"), got)
}

func TestRenderAdditionalRulesBeforeMatch(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:        "lite",
		MixedPort:       7890,
		Mode:            "rule",
		LogLevel:        "info",
		DNSEnabled:      false,
		AdditionalRules: []string{"DOMAIN-SUFFIX,example.com,节点选择", "DOMAIN-KEYWORD,stream,节点选择"},
	})

	text := string(got)
	matchPos := strings.Index(text, "MATCH,节点选择")
	examplePos := strings.Index(text, "DOMAIN-SUFFIX,example.com,节点选择")
	streamPos := strings.Index(text, "DOMAIN-KEYWORD,stream,节点选择")
	if examplePos == -1 || streamPos == -1 {
		t.Fatalf("rendered yaml missing additional rules:\n%s", text)
	}
	if examplePos > matchPos || streamPos > matchPos {
		t.Fatalf("additional rules should appear before MATCH rule:\n%s", text)
	}
}

func TestRenderRuleProvidersIncluded(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:   "lite",
		MixedPort:  7890,
		Mode:       "rule",
		LogLevel:   "info",
		DNSEnabled: false,
		RuleProviders: []model.RuleProviderConfig{
			{
				Name:      "apple-rules",
				Type:      "http",
				URL:       "https://example.com/apple.yaml",
				Behavior:  "classical",
				Format:    "yaml",
				Interval:  86400,
				Proxy:     "DIRECT",
				Policy:    "Apple",
				Enabled:   true,
				NoResolve: true,
			},
		},
	})

	text := string(got)
	if !strings.Contains(text, "rule-providers:") {
		t.Fatalf("rendered yaml missing rule-providers section:\n%s", text)
	}
	if !strings.Contains(text, "apple-rules:") || !strings.Contains(text, "RULE-SET,apple-rules,Apple,no-resolve") {
		t.Fatalf("rendered yaml missing provider mapping:\n%s", text)
	}
}

func TestRenderCustomProxyGroupsIncluded(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:   "lite",
		MixedPort:  7890,
		Mode:       "rule",
		LogLevel:   "info",
		DNSEnabled: false,
		CustomProxyGroups: []model.CustomProxyGroupConfig{
			{
				Name:    "测试分组",
				Type:    "select",
				Members: []string{"ss-node", "DIRECT"},
				Enabled: true,
			},
		},
	})

	text := string(got)
	if !strings.Contains(text, "- name: 测试分组") || !strings.Contains(text, "- ss-node") {
		t.Fatalf("rendered yaml missing custom proxy group:\n%s", text)
	}
}

func TestRenderFinalPolicyOverride(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:    "lite",
		MixedPort:   7890,
		Mode:        "rule",
		LogLevel:    "info",
		DNSEnabled:  false,
		FinalPolicy: "DIRECT",
	})

	text := string(got)
	if !strings.Contains(text, "MATCH,DIRECT") {
		t.Fatalf("rendered yaml missing overridden final policy:\n%s", text)
	}
	if strings.Contains(text, "MATCH,节点选择") {
		t.Fatalf("rendered yaml should not contain default final policy after override:\n%s", text)
	}
}

func TestRenderAdvancedBaseDNSProfileSniffer(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:                "lite",
		MixedPort:               7897,
		AllowLAN:                true,
		Mode:                    "rule",
		LogLevel:                "info",
		IPv6:                    false,
		UnifiedDelay:            true,
		TCPConcurrent:           true,
		FindProcessMode:         "strict",
		GlobalClientFingerprint: "chrome",
		DNS: &model.DNSConfig{
			Enable:         true,
			Listen:         "127.0.0.1:5335",
			UseSystemHosts: false,
			EnhancedMode:   "fake-ip",
			FakeIPRange:    "198.18.0.1/16",
			DefaultNameserver: []string{
				"180.76.76.76",
				"8.8.8.8",
			},
			Nameserver: []string{
				"180.76.76.76",
				"https://dns.alidns.com/dns-query",
			},
			Fallback: []string{
				"https://dns.google/dns-query",
				"tls://1.0.0.1:853",
			},
			FallbackFilter: &model.DNSFallbackFilter{
				GeoIP:  true,
				IPCIDR: []string{"240.0.0.0/4"},
				Domain: []string{"+.google.com"},
			},
			FakeIPFilter: []string{"*.lan", "pool.ntp.org"},
			NameserverPolicy: map[string][]string{
				"geosite:cn": []string{"223.5.5.5", "119.29.29.29"},
			},
		},
		Profile: &model.ProfileConfig{
			StoreSelected: true,
			StoreFakeIP:   false,
		},
		Sniffer: &model.SnifferConfig{
			Enable:      true,
			ParsePureIP: true,
			TLS: &model.SniffProtocol{
				Ports: []string{"443", "8443"},
			},
			HTTP: &model.SniffHTTP{
				Ports:               []string{"80", "8080-8880"},
				OverrideDestination: true,
			},
			QUIC: &model.SniffProtocol{
				Ports: []string{"443", "8443"},
			},
		},
	})

	text := string(got)
	needles := []string{
		"mixed-port: 7897",
		"allow-lan: true",
		"unified-delay: true",
		"tcp-concurrent: true",
		"find-process-mode: strict",
		"global-client-fingerprint: chrome",
		"dns:",
		"listen: 127.0.0.1:5335",
		"use-system-hosts: false",
		"fake-ip-range: 198.18.0.1/16",
		"default-nameserver:",
		"fallback-filter:",
		"fake-ip-filter:",
		"nameserver-policy:",
		"geosite:cn",
		"profile:",
		"store-selected: true",
		"store-fake-ip: false",
		"sniffer:",
		"parse-pure-ip: true",
		"TLS:",
		"HTTP:",
		"override-destination: true",
		"QUIC:",
	}

	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			t.Fatalf("rendered yaml missing %q:\n%s", needle, text)
		}
	}
}

func mustParseFile(t *testing.T, path string) []model.NodeIR {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	result := parser.ParseContent(content, model.SourceInfo{Name: filepath.Base(path), Kind: "test"})
	if len(result.Errors) != 0 {
		t.Fatalf("ParseContent(%q) errors = %#v", path, result.Errors)
	}
	return result.Nodes
}

func mustRender(t *testing.T, nodes []model.NodeIR, opts model.RenderOptions) []byte {
	t.Helper()

	got, err := RenderMihomo(nodes, opts)
	if err != nil {
		t.Fatalf("RenderMihomo() error = %v", err)
	}
	return got
}

func assertGoldenFile(t *testing.T, path string, got []byte) {
	t.Helper()

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func standardRenderOptions() model.RenderOptions {
	return model.RenderOptions{
		Template:     "standard",
		MixedPort:    7890,
		Mode:         "rule",
		LogLevel:     "info",
		DNSEnabled:   true,
		EnhancedMode: "fake-ip",
	}
}
