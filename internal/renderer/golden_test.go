package renderer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestRenderVMessDefaultsCipherAuto(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "vmess.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	if !strings.Contains(string(got), "cipher: auto") {
		t.Fatalf("vmess output should include cipher: auto:\n%s", string(got))
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

func TestRenderTemplateModeIgnoresRulesModeCustomRules(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.TemplateRuleMode = "template"
	opts.ExternalConfig.TemplateKey = "cm_online_multi_country_cf"
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"private"}
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "draft_only_rule",
			Label:          "草稿规则",
			Enabled:        true,
			TargetMode:     "direct",
			SourceType:     "inline",
			Behavior:       "domain",
			Format:         "text",
			Payload:        []string{"example.com"},
			InsertPosition: "before_match",
		},
	}

	got := mustRender(t, nodes, opts)
	text := string(got)
	if strings.Contains(text, "draft_only_rule") || strings.Contains(text, "草稿规则") {
		t.Fatalf("template mode must not include rules-mode custom rules:\n%s", text)
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

func TestNoRegionGroupsGenerated(t *testing.T) {
	nodes := model.NormalizeNodes([]model.NodeIR{
		testSSNode("JP Node", "jp.example.com"),
		testSSNode("US Node", "us.example.com"),
		testSSNode("HK Node", "hk.example.com"),
		testSSNode("SG Node", "sg.example.com"),
	})
	opts := standardRenderOptions()
	opts.EnabledRules = []string{"ai"}

	cfg := mustRenderConfig(t, nodes, opts)
	regionGroupNames := []string{"🇭🇰 香港", "🇯🇵 日本", "🇺🇸 美国", "🇸🇬 新加坡", "🇹🇼 台湾", "🇬🇧 英国", "🇩🇪 德国", "🇳🇱 荷兰", "🇷🇺 俄罗斯", "🇰🇷 韩国"}
	for _, name := range regionGroupNames {
		assertNoProxyGroup(t, cfg.ProxyGroups, name)
	}

	main := findProxyGroup(t, cfg.ProxyGroups, groupNodeSelect)
	for _, name := range []string{groupAutoSelect, "DIRECT", "REJECT", "JP Node", "US Node", "HK Node", "SG Node"} {
		assertContainsProxy(t, main.Proxies, name)
	}
	for _, name := range regionGroupNames {
		assertNotContainsProxy(t, main.Proxies, name)
	}

	globalAuto := findProxyGroup(t, cfg.ProxyGroups, groupAutoSelect)
	assertEqualProxies(t, globalAuto.Proxies, []string{"JP Node", "US Node", "HK Node", "SG Node"})

	ai := findProxyGroup(t, cfg.ProxyGroups, "🤖 AI 服务")
	for _, name := range []string{groupNodeSelect, groupAutoSelect, "DIRECT", "REJECT"} {
		assertContainsProxy(t, ai.Proxies, name)
	}
}

func TestRuleGroupsIncludeRealNodesInFullMode(t *testing.T) {
	nodes := model.NormalizeNodes([]model.NodeIR{
		testSSNode("JP Node", "jp.example.com"),
		testSSNode("US Node", "us.example.com"),
	})
	opts := standardRenderOptions()
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"ai", "youtube", "netflix"}
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{RuleGroupNodeMode: "full"})

	cfg := mustRenderConfig(t, nodes, opts)
	for _, groupName := range []string{"🤖 AI 服务", "📹 油管视频", "🎬 奈飞", groupFinal} {
		group := findProxyGroup(t, cfg.ProxyGroups, groupName)
		for _, name := range []string{groupNodeSelect, groupAutoSelect, "DIRECT", "REJECT", "JP Node", "US Node"} {
			assertContainsProxy(t, group.Proxies, name)
		}
	}
}

func TestRuleGroupsStayCompactInCompactMode(t *testing.T) {
	nodes := model.NormalizeNodes([]model.NodeIR{
		testSSNode("JP Node", "jp.example.com"),
		testSSNode("US Node", "us.example.com"),
	})
	opts := standardRenderOptions()
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"ai"}
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{RuleGroupNodeMode: "compact"})

	cfg := mustRenderConfig(t, nodes, opts)
	ai := findProxyGroup(t, cfg.ProxyGroups, "🤖 AI 服务")
	assertEqualProxies(t, ai.Proxies, []string{groupNodeSelect, groupAutoSelect, "DIRECT", "REJECT"})
	assertNotContainsProxy(t, ai.Proxies, "JP Node")
	assertNotContainsProxy(t, ai.Proxies, "US Node")
}

func TestSpecialRuleGroupsRemainCompactInFullMode(t *testing.T) {
	nodes := model.NormalizeNodes([]model.NodeIR{
		testSSNode("JP Node", "jp.example.com"),
		testSSNode("US Node", "us.example.com"),
	})
	opts := standardRenderOptions()
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"adblock", "private", "domestic"}
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{RuleGroupNodeMode: "full"})

	cfg := mustRenderConfig(t, nodes, opts)
	for _, groupName := range []string{"🛑 广告拦截", "🏠 私有网络", "🔒 国内服务"} {
		group := findProxyGroup(t, cfg.ProxyGroups, groupName)
		assertNotContainsProxy(t, group.Proxies, "JP Node")
		assertNotContainsProxy(t, group.Proxies, "US Node")
	}
	assertEqualProxies(t, findProxyGroup(t, cfg.ProxyGroups, "🛑 广告拦截").Proxies, []string{"REJECT", "DIRECT", groupNodeSelect})
	assertEqualProxies(t, findProxyGroup(t, cfg.ProxyGroups, "🏠 私有网络").Proxies, []string{"DIRECT", "REJECT", groupNodeSelect, groupAutoSelect})
	assertEqualProxies(t, findProxyGroup(t, cfg.ProxyGroups, "🔒 国内服务").Proxies, []string{"DIRECT", "REJECT", groupNodeSelect, groupAutoSelect})
}

func TestFinalYAMLSerializationSanitizesRegionGroups(t *testing.T) {
	cfg := mihomoConfig{
		Proxies: []mihomoProxy{
			{Name: "HK Node", Type: "ss", Server: "hk.example.com", Port: 443},
			{Name: "JP Node", Type: "ss", Server: "jp.example.com", Port: 443},
		},
		ProxyGroups: []mihomoProxyGroup{
			{Name: groupNodeSelect, Type: "select", Proxies: []string{groupAutoSelect, "DIRECT", "🇭🇰 香港", "🇯🇵 日本"}},
			{Name: groupAutoSelect, Type: "url-test", Proxies: []string{"HK Node", "🇭🇰 香港", "DIRECT"}},
			{Name: "⚡ 香港自动", Type: "url-test", Proxies: []string{"HK Node", groupAutoSelect}},
			{Name: "🇭🇰 香港", Type: "select", Proxies: []string{groupAutoSelect, groupNodeSelect, "🇯🇵 日本", "⚡ 日本自动", "⚡ 香港自动", "DIRECT", "HK Node"}},
			{Name: "⚡ 日本自动", Type: "url-test", Proxies: []string{"JP Node"}},
			{Name: "🇯🇵 日本", Type: "select", Proxies: []string{groupAutoSelect, "DIRECT", "JP Node"}},
			{Name: "🤖 AI 服务", Type: "select", Proxies: []string{groupNodeSelect, groupAutoSelect, "DIRECT", "REJECT"}},
		},
		Rules: []string{"MATCH," + groupNodeSelect},
	}

	yamlText := string(renderConfig(cfg))
	var out struct {
		ProxyGroups []mihomoProxyGroup `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal([]byte(yamlText), &out); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v\n%s", err, yamlText)
	}

	assertNoProxyGroup(t, out.ProxyGroups, "🇭🇰 香港")
	assertNoProxyGroup(t, out.ProxyGroups, "🇯🇵 日本")
	assertNoProxyGroup(t, out.ProxyGroups, "⚡ 香港自动")
	assertNoProxyGroup(t, out.ProxyGroups, "⚡ 日本自动")

	globalAuto := findProxyGroup(t, out.ProxyGroups, groupAutoSelect)
	assertEqualProxies(t, globalAuto.Proxies, []string{"HK Node"})

	main := findProxyGroup(t, out.ProxyGroups, groupNodeSelect)
	for _, forbidden := range []string{"🇭🇰 香港", "🇯🇵 日本", "⚡ 香港自动", "⚡ 日本自动"} {
		assertNotContainsProxy(t, main.Proxies, forbidden)
	}

	ai := findProxyGroup(t, out.ProxyGroups, "🤖 AI 服务")
	assertContainsProxy(t, ai.Proxies, groupNodeSelect)
	assertContainsProxy(t, ai.Proxies, groupAutoSelect)
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
	nodes = ensureUniqueNames(nodes, opts.NameOptions.DedupeSuffixStyle)
	proxies, err := buildProxies(nodes)
	if err != nil {
		t.Fatalf("buildProxies() error = %v", err)
	}
	enabled := resolveEnabledRules(opts.RuleMode, opts.EnabledRules)
	proxyGroups, err := buildProxyGroups(nodes, opts.Template, opts.CustomProxyGroups, enabled, opts.CustomRules, opts.GroupProxyMode, opts.GroupOptions)
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

func testSSNode(name, server string) model.NodeIR {
	return model.NodeIR{
		Name:   name,
		Type:   model.ProtocolSS,
		Server: server,
		Port:   443,
		Auth: model.Auth{
			Password: "password",
		},
		Raw: map[string]interface{}{
			"method": "aes-256-gcm",
		},
	}
}

func findProxyGroup(t *testing.T, groups []mihomoProxyGroup, name string) mihomoProxyGroup {
	t.Helper()
	for _, group := range groups {
		if group.Name == name {
			return group
		}
	}
	t.Fatalf("missing proxy-group %q in %#v", name, groups)
	return mihomoProxyGroup{}
}

func assertNoProxyGroup(t *testing.T, groups []mihomoProxyGroup, name string) {
	t.Helper()
	for _, group := range groups {
		if group.Name == name {
			t.Fatalf("proxy-groups must not contain %q: %#v", name, groups)
		}
	}
}

func assertContainsProxy(t *testing.T, proxies []string, want string) {
	t.Helper()
	if !containsString(proxies, want) {
		t.Fatalf("proxies %#v missing %q", proxies, want)
	}
}

func assertNotContainsProxy(t *testing.T, proxies []string, forbidden string) {
	t.Helper()
	if containsString(proxies, forbidden) {
		t.Fatalf("proxies %#v must not contain %q", proxies, forbidden)
	}
}

func assertEqualProxies(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("proxies = %#v, want %#v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("proxies = %#v, want %#v", got, want)
		}
	}
}
