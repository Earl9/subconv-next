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

func TestRenderGoldenFullFormatTopLevelOrder(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "balanced"
	got := mustRender(t, nodes, opts)
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "full-format-top-level-order.yaml"), got)

	text := string(got)
	orderedNeedles := []string{
		"mixed-port:",
		"allow-lan:",
		"mode:",
		"log-level:",
		"ipv6:",
		"unified-delay:",
		"tcp-concurrent:",
		"find-process-mode:",
		"global-client-fingerprint:",
		"dns:",
		"profile:",
		"sniffer:",
		"geodata-mode:",
		"geo-auto-update:",
		"geodata-loader:",
		"geo-update-interval:",
		"geox-url:",
		"proxies:",
		"proxy-groups:",
		"rule-providers:",
		"rules:",
	}
	last := -1
	for _, needle := range orderedNeedles {
		pos := strings.Index(text, needle)
		if pos == -1 {
			t.Fatalf("missing top-level key %q:\n%s", needle, text)
		}
		if pos < last {
			t.Fatalf("top-level order broken at %q:\n%s", needle, text)
		}
		last = pos
	}
}

func TestRenderGoldenDNSSnifferGeodata(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "dns-sniffer-geodata.yaml"), got)
}

func TestRenderGoldenAnyTLSPreserveFingerprint(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "anytls.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "anytls-preserve-fingerprint.yaml"), got)
	if !strings.Contains(string(got), "client-fingerprint: chrome") {
		t.Fatalf("expected anytls output to preserve client-fingerprint:\n%s", string(got))
	}
}

func TestRenderGoldenVLESSXHTTPReality(t *testing.T) {
	nodes := mustParseContent(t, []byte("vless://uuid-1@example.com:443?type=xhttp&security=reality&sni=example.com&fp=chrome&pbk=pub&sid=abcd&spx=%2Fspider&path=%2Fdemo&mode=auto&no-grpc-header=false#xhttp"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "vless-xhttp-reality.yaml"), got)

	text := string(got)
	needles := []string{
		"encryption: none",
		"network: xhttp",
		"xhttp-opts:",
		"spider-x: /spider",
	}
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			t.Fatalf("missing %q in rendered output:\n%s", needle, text)
		}
	}
	if strings.Contains(text, "_spider-x") {
		t.Fatalf("rendered output must not contain _spider-x:\n%s", text)
	}
}

func TestRenderGoldenRuleProvidersBalanced(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "balanced"
	got := mustRender(t, nodes, opts)
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "rule-providers-balanced.yaml"), got)

	text := string(got)
	for _, needle := range []string{
		"category-ai-chat-!cn:",
		"google:",
		"github:",
		"telegram:",
		"RULE-SET,category-ai-chat-!cn,🤖 AI 服务",
		"RULE-SET,google,🔍 谷歌服务",
		"RULE-SET,github,🐱 代码托管",
		"RULE-SET,telegram,📲 电报消息",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("balanced output missing %q:\n%s", needle, text)
		}
	}
	if strings.Count(text, ".mrs") <= 20 {
		t.Fatalf("expected more than 20 provider references in balanced output:\n%s", text)
	}
}

func TestRenderGoldenUIKeyExpansion(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"streaming", "gaming", "finance", "social"}
	got := mustRender(t, nodes, opts)
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "ui-key-expansion.yaml"), got)

	text := string(got)
	for _, needle := range []string{
		"🎬 奈飞",
		"🏰 迪士尼+",
		"📺 欧美流媒体",
		"🎌 亚洲流媒体",
		"🎮 Steam",
		"🖥️ PC 游戏",
		"🎯 主机游戏",
		"💳 支付平台",
		"₿ 加密货币",
		"🐦 推特/X",
		"📘 Meta 系",
		"🎙️ Discord",
		"💬 其他社交",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("ui key expansion missing %q:\n%s", needle, text)
		}
	}
}

func TestRenderNoDuplicatePaths(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "full"
	cfg := mustRenderConfig(t, nodes, opts)

	seen := map[string]string{}
	for name, provider := range cfg.RuleProviders {
		if existing, ok := seen[provider.Path]; ok {
			t.Fatalf("duplicate provider path %q for %q and %q", provider.Path, existing, name)
		}
		seen[provider.Path] = name
	}
}

func TestValidateNoMissingReference(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "full"
	cfg := mustRenderConfig(t, nodes, opts)
	warnings := ValidateMihomoConfig(cfg)
	for _, warning := range warnings {
		switch warning.Code {
		case "missing_proxy_reference", "missing_rule_provider_reference", "missing_rule_target", "duplicate_proxy_name", "duplicate_group_name", "duplicate_rule_provider_path":
			t.Fatalf("unexpected validation warning: %+v", warning)
		}
	}
}

func TestRenderMatchLast(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "full"
	got := mustRender(t, nodes, opts)
	lines := strings.Split(strings.TrimSpace(string(got)), "\n")
	if !strings.Contains(lines[len(lines)-1], "MATCH,🐟 漏网之鱼") {
		t.Fatalf("MATCH must be last line in rules section:\n%s", string(got))
	}
}

func TestRenderCompactProxies(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "anytls.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "compact-proxies.yaml"), got)
	if !strings.Contains(string(got), "- {name:") {
		t.Fatalf("proxies should use compact flow mapping:\n%s", string(got))
	}
}

func TestRenderCompactRuleProviders(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "compact-rule-providers.yaml"), got)
	if !strings.Contains(string(got), "category-ai-chat-!cn: {type: http") {
		t.Fatalf("rule-providers should use compact flow mapping:\n%s", string(got))
	}
}

func TestRenderCompactDNS(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "compact-dns.yaml"), got)
	text := string(got)
	for _, needle := range []string{
		"default-nameserver: [",
		"nameserver: [",
		"fallback: [",
		"fallback-filter: {geoip: true",
		"fake-ip-filter: [",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("compact dns output missing %q:\n%s", needle, text)
		}
	}
}

func TestRenderCompactProxyGroups(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "compact-proxy-groups.yaml"), got)
	if !strings.Contains(string(got), `- {name: "🚀 节点选择"`) {
		t.Fatalf("proxy-groups should use compact flow mapping:\n%s", string(got))
	}
}

func TestRenderRulesLinePerItem(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "rules-line-per-item.yaml"), got)
	text := string(got)
	if !strings.Contains(text, "\nrules:\n  - ") || !strings.Contains(text, "MATCH,🐟 漏网之鱼") {
		t.Fatalf("rules should remain one line per item:\n%s", text)
	}
}

func TestRenderTemplateModeAppliesPreset(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.TemplateRuleMode = "template"
	opts.ExternalConfig.TemplateKey = "cm_online_multi_country_cf"
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"private"}

	got := mustRender(t, nodes, opts)
	text := string(got)
	for _, needle := range []string{
		"☁️ 云服务",
		"RULE-SET,cloudflare,☁️ 云服务",
		"RULE-SET,google,🔍 谷歌服务",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("template mode output missing %q:\n%s", needle, text)
		}
	}
}

func TestRenderTemplateModeFallbackToLiteTemplate(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.Template = "lite"
	opts.TemplateRuleMode = "template"
	opts.ExternalConfig.TemplateKey = "none"
	opts.RuleMode = "full"

	got := mustRender(t, nodes, opts)
	text := string(got)
	if strings.Contains(text, "rule-providers:") {
		t.Fatalf("lite template fallback should not emit rule-providers:\n%s", text)
	}
	if strings.Contains(text, "🤖 AI 服务") || strings.Contains(text, "🔍 谷歌服务") {
		t.Fatalf("lite template fallback should not emit business groups:\n%s", text)
	}
	if !strings.Contains(text, "MATCH,🐟 漏网之鱼") {
		t.Fatalf("lite template fallback must retain final MATCH:\n%s", text)
	}
}

func TestRenderRulesModeIgnoresTemplatePreset(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.TemplateRuleMode = "rules"
	opts.ExternalConfig.TemplateKey = "cm_online_full_cf"
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"private"}

	got := mustRender(t, nodes, opts)
	text := string(got)
	if !strings.Contains(text, "🏠 私有网络") {
		t.Fatalf("rules mode should keep selected rules:\n%s", text)
	}
	if strings.Contains(text, "☁️ 云服务") {
		t.Fatalf("rules mode must not be overridden by template preset:\n%s", text)
	}
}

func TestRenderDuplicateGroupNameFails(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "balanced"
	opts.CustomProxyGroups = []model.CustomProxyGroupConfig{
		{Name: "🤖 AI 服务", Type: "select", Members: []string{"ss-node", "DIRECT"}, Enabled: true},
	}
	_, err := RenderMihomo(nodes, opts)
	if err == nil || !strings.Contains(err.Error(), "duplicate proxy-group name") {
		t.Fatalf("RenderMihomo() error = %v, want duplicate proxy-group name", err)
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

func mustParseContent(t *testing.T, content []byte) []model.NodeIR {
	t.Helper()
	result := parser.ParseContent(content, model.SourceInfo{Name: "inline", Kind: "test"})
	if len(result.Errors) != 0 {
		t.Fatalf("ParseContent() errors = %#v", result.Errors)
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

func mustRenderConfig(t *testing.T, nodes []model.NodeIR, opts model.RenderOptions) mihomoConfig {
	t.Helper()
	opts = NormalizeRenderOptions(opts)
	nodes = model.NormalizeNodes(nodes)
	nodes = ensureUniqueNames(nodes)
	proxies, err := buildProxies(nodes)
	if err != nil {
		t.Fatalf("buildProxies() error = %v", err)
	}
	enabled := resolveEnabledRules(opts.RuleMode, opts.EnabledRules)
	proxyGroups, err := buildProxyGroups(nodes, opts.Template, opts.CustomProxyGroups, enabled, opts.CustomRules, opts.GroupProxyMode)
	if err != nil {
		t.Fatalf("buildProxyGroups() error = %v", err)
	}
	return mihomoConfig{
		MixedPort:               opts.MixedPort,
		AllowLAN:                opts.AllowLAN,
		Mode:                    opts.Mode,
		LogLevel:                opts.LogLevel,
		IPv6:                    opts.IPv6,
		UnifiedDelay:            opts.UnifiedDelay,
		TCPConcurrent:           opts.TCPConcurrent,
		FindProcessMode:         opts.FindProcessMode,
		GlobalClientFingerprint: opts.GlobalClientFingerprint,
		DNS:                     buildDNSConfig(opts),
		Profile:                 buildProfileConfig(opts.Profile),
		Sniffer:                 buildSnifferConfig(opts.Sniffer),
		GeodataMode:             opts.GeodataMode,
		GeoAutoUpdate:           opts.GeoAutoUpdate,
		GeodataLoader:           opts.GeodataLoader,
		GeoUpdateInterval:       opts.GeoUpdateInterval,
		GeoxURL:                 buildGeoxURLConfig(opts.GeoxURL),
		Proxies:                 proxies,
		ProxyGroups:             proxyGroups,
		RuleProviders:           buildRuleProviders(opts.RuleProviders, enabled, opts.CustomRules, opts.CustomProxyGroups),
		RuleProviderOrder:       orderedProviderNames(enabled),
		Rules:                   buildRules(opts.Template, opts.FinalPolicy, enabled, opts.AdditionalRules, opts.RuleProviders, opts.CustomRules, opts.CustomProxyGroups),
	}
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
	opts := model.DefaultRenderOptions()
	opts.Template = "standard"
	opts.RuleMode = "balanced"
	opts.GroupProxyMode = "compact"
	return opts
}
