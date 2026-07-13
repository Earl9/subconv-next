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
		"proxies:",
		"proxy-groups:",
		"rule-providers:",
		"rules:",
	}
	last := -1
	for _, needle := range orderedNeedles {
		pos := topLevelKeyIndex(text, needle)
		if pos == -1 {
			t.Fatalf("missing top-level key %q:\n%s", needle, text)
		}
		if pos < last {
			t.Fatalf("top-level order broken at %q:\n%s", needle, text)
		}
		last = pos
	}
}

func topLevelKeyIndex(text, key string) int {
	if strings.HasPrefix(text, key) {
		return 0
	}
	pos := strings.Index(text, "\n"+key)
	if pos == -1 {
		return -1
	}
	return pos + 1
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

func TestRenderVLESSXUDPPacketEncoding(t *testing.T) {
	nodes := mustParseContent(t, []byte("vless://uuid-1@example.com:443?type=tcp&security=tls&sni=example.com&packet_encoding=XUDP&allowInsecure=1#xudp"))
	got := mustRender(t, nodes, standardRenderOptions())
	text := string(got)
	if !strings.Contains(text, "packet-encoding: xudp") {
		t.Fatalf("rendered output missing packet-encoding xudp:\n%s", text)
	}
	if !strings.Contains(text, "skip-cert-verify: true") {
		t.Fatalf("rendered output missing vless skip-cert-verify:\n%s", text)
	}
	if strings.Contains(text, "packet_encoding") {
		t.Fatalf("rendered output must not contain URI packet_encoding alias:\n%s", text)
	}
}

func TestRenderIgnoresGeodataOptions(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.GeodataMode = true
	opts.GeoAutoUpdate = true
	opts.GeodataLoader = "standard"
	opts.GeoUpdateInterval = 24
	opts.GeoxURL = &model.GeoxURLConfig{
		GeoIP:   "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geoip.dat",
		GeoSite: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geosite.dat",
		MMDB:    "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/country.mmdb",
		ASN:     "https://github.com/xishang0128/geoip/releases/download/latest/GeoLite2-ASN.mmdb",
	}

	got := mustRender(t, nodes, opts)
	text := string(got)
	for _, needle := range []string{
		"geodata-mode:",
		"geo-auto-update:",
		"geodata-loader:",
		"geo-update-interval:",
		"geox-url:",
	} {
		if strings.Contains(text, needle) {
			t.Fatalf("rendered output must not contain %q:\n%s", needle, text)
		}
	}
}

func TestRenderForcesProfileAndSnifferDefaults(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.Profile = &model.ProfileConfig{
		StoreSelected: true,
		StoreFakeIP:   false,
	}
	opts.Sniffer = &model.SnifferConfig{
		Enable:      true,
		ParsePureIP: true,
	}

	got := mustRender(t, nodes, opts)
	text := string(got)
	for _, needle := range []string{
		"store-fake-ip: false",
		"parse-pure-ip: false",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("rendered output missing forced value %q:\n%s", needle, text)
		}
	}
	for _, needle := range []string{
		"store-fake-ip: true",
		"parse-pure-ip: true",
	} {
		if strings.Contains(text, needle) {
			t.Fatalf("rendered output must not contain stale value %q:\n%s", needle, text)
		}
	}
}

func TestRenderPreservesMihomoProxyFields(t *testing.T) {
	nodes := mustParseContent(t, []byte(`
proxies:
  - name: "hysteria yaml"
    type: hysteria
    server: hysteria.example.com
    port: 443
    auth-str: hy-secret
    protocol: udp
    obfs: salamander
    obfs-param: obfs-secret
    up: 100
    down: 100
    sni: hysteria.example.com
    skip-cert-verify: true
    custom-hy-flag: keep-v1
  - name: "hy2 yaml"
    type: hysteria2
    server: hkt03ddns.poke-mon.xyz
    port: 20000
    ports: 20000-50000
    mport: 20000-50000
    udp: true
    password: hy2-secret
    sni: www.bing.com
    skip-cert-verify: false
    custom-hy2-flag: keep-me
`))
	got := mustRender(t, nodes, standardRenderOptions())
	text := string(got)
	for _, needle := range []string{
		"type: hysteria",
		"auth-str: hy-secret",
		"protocol: udp",
		"obfs: salamander",
		"obfs-param: obfs-secret",
		"up: 100",
		"down: 100",
		"custom-hy-flag: keep-v1",
		"ports: 20000-50000",
		"mport: 20000-50000",
		"skip-cert-verify: false",
		"custom-hy2-flag: keep-me",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("rendered output missing preserved field %q:\n%s", needle, text)
		}
	}
}

func TestRenderMieru(t *testing.T) {
	nodes := mustParseContent(t, []byte("mieru://user:secret@example.com?port-range=2090-2099&transport=udp&udp=1&multiplexing=MULTIPLEXING_HIGH&handshake-mode=HANDSHAKE_STANDARD#mieru-node"))
	got := mustRender(t, nodes, standardRenderOptions())
	text := string(got)
	for _, needle := range []string{
		"type: mieru",
		"server: example.com",
		"port-range: 2090-2099",
		"transport: UDP",
		"udp: true",
		"username: user",
		"password: secret",
		"multiplexing: MULTIPLEXING_HIGH",
		"handshake-mode: HANDSHAKE_STANDARD",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("mieru output missing %q:\n%s", needle, text)
		}
	}
}

func TestRenderHTTPAndSOCKS5(t *testing.T) {
	nodes := mustParseContent(t, []byte(strings.Join([]string{
		"https://user:secret@example.com:8443?sni=proxy.example.com&skip-cert-verify=1#https-node",
		"socks5://sock:s3cr3t@socks.example.net:1080?udp=0#socks-node",
	}, "\n")))
	got := mustRender(t, nodes, standardRenderOptions())
	text := string(got)
	for _, needle := range []string{
		"type: http",
		"server: example.com",
		"port: 8443",
		"username: user",
		"password: secret",
		"tls: true",
		"sni: proxy.example.com",
		"skip-cert-verify: true",
		"type: socks5",
		"server: socks.example.net",
		"udp: false",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("http/socks output missing %q:\n%s", needle, text)
		}
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

func TestMicrosoftCloudDriveHasSeparateRuleGroup(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"microsoft", "onedrive"}
	got := mustRender(t, nodes, opts)

	text := string(got)
	for _, needle := range []string{
		"- {name: ☁️ 微软云盘, type: select",
		"- {name: Ⓜ️ 微软服务, type: select",
		"onedrive: {type: http",
		"microsoft: {type: http",
		"RULE-SET,onedrive,☁️ 微软云盘",
		"RULE-SET,microsoft,Ⓜ️ 微软服务",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("microsoft split output missing %q:\n%s", needle, text)
		}
	}
	if strings.Contains(text, "RULE-SET,onedrive,Ⓜ️ 微软服务") {
		t.Fatalf("onedrive must not target microsoft service group:\n%s", text)
	}
	if strings.Index(text, "RULE-SET,onedrive,☁️ 微软云盘") > strings.Index(text, "RULE-SET,microsoft,Ⓜ️ 微软服务") {
		t.Fatalf("onedrive rule must be emitted before microsoft rule:\n%s", text)
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
		"respect-rules: false",
		"nameserver: [",
		"223.5.5.5",
		"fallback: [",
		"https://dns.alidns.com/dns-query",
		"fake-ip-filter: [",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("compact dns output missing %q:\n%s", needle, text)
		}
	}
}

func TestRenderCustomDNS(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	opts := standardRenderOptions()
	opts.CustomDNS = true
	opts.DNS = &model.DNSConfig{
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
			"+.example.com": {
				"https://dns.example/dns-query",
			},
		},
	}

	got := string(mustRender(t, nodes, opts))
	for _, needle := range []string{
		"listen: 0.0.0.0:1053",
		"use-hosts: false",
		"use-system-hosts: true",
		"respect-rules: true",
		"enhanced-mode: redir-host",
		"https://dns.google/dns-query",
		"proxy-server-nameserver:",
		"direct-nameserver-follow-policy: true",
		"https://dns.cloudflare.com/dns-query",
		"dns.example",
		"fake-ip-filter:",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("custom dns output missing %q:\n%s", needle, got)
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
	assertRulesUseDoubleQuotes(t, text)
}

func assertRulesUseDoubleQuotes(t *testing.T, rendered string) {
	t.Helper()
	rulesIndex := strings.Index(rendered, "\nrules:\n")
	if rulesIndex < 0 {
		t.Fatalf("rendered output has no rules section:\n%s", rendered)
	}
	for _, line := range strings.Split(rendered[rulesIndex+len("\nrules:\n"):], "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, `  - "`) || !strings.HasSuffix(line, `"`) {
			t.Fatalf("rule line is not consistently double quoted: %q", line)
		}
	}
}

func TestRenderRulesAlwaysUseDoubleQuotes(t *testing.T) {
	rendered, err := renderConfig(mihomoConfig{Rules: []string{
		"RULE-SET,adobe,REJECT",
		"DOMAIN-KEYWORD,tracker,DIRECT",
		"RULE-SET,microsoft,Ⓜ️ 微软服务",
		"RULE-SET,a,😶‍🌫️ 18+",
	}})
	if err != nil {
		t.Fatalf("renderConfig() error = %v", err)
	}

	text := string(rendered)
	for _, rule := range []string{
		"RULE-SET,adobe,REJECT",
		"DOMAIN-KEYWORD,tracker,DIRECT",
		"RULE-SET,microsoft,Ⓜ️ 微软服务",
		"RULE-SET,a,😶‍🌫️ 18+",
	} {
		if !strings.Contains(text, `  - "`+rule+`"`) {
			t.Fatalf("rendered output does not double quote %q:\n%s", rule, text)
		}
	}
	assertRulesUseDoubleQuotes(t, text)
}

func TestRenderAdditionalRulesNormalizesPastedList(t *testing.T) {
	nodes := []model.NodeIR{testSSNode("Test Node", "example.com")}
	opts := standardRenderOptions()
	opts.AdditionalRules = []string{
		"- DOMAIN-KEYWORD,tracker,PROXY",
		"##- IP-CIDR,0.0.0.0/0,PROXY",
		"DOMAIN-SUFFIX,39.la,DIRECT # direct site",
		"",
		"# Adobe block",
		"- RULE-SET,Adobe,REJECT",
	}
	opts.RuleProviders = []model.RuleProviderConfig{
		{Name: "Adobe", Type: "http", Behavior: "classical", URL: "https://example.com/adobe.yaml"},
	}

	cfg := mustRenderConfig(t, nodes, opts)
	for _, want := range []string{
		"DOMAIN-KEYWORD,tracker,🚀 节点选择",
		"DOMAIN-SUFFIX,39.la,DIRECT",
		"RULE-SET,Adobe,REJECT",
	} {
		if !containsString(cfg.Rules, want) {
			t.Fatalf("rules = %#v, want %q", cfg.Rules, want)
		}
	}
	for _, rule := range cfg.Rules {
		if strings.Contains(rule, "0.0.0.0/0") {
			t.Fatalf("disabled rule leaked into output: %#v", cfg.Rules)
		}
	}
	if got := cfg.Rules[len(cfg.Rules)-1]; !strings.HasPrefix(got, "MATCH,") {
		t.Fatalf("last rule = %q, want MATCH", got)
	}
}

func TestCustomRulesPrecedeBuiltInRules(t *testing.T) {
	nodes := []model.NodeIR{testSSNode("Test Node", "example.com")}
	opts := standardRenderOptions()
	opts.RuleMode = "custom"
	opts.EnabledRules = []string{"adblock", "domestic", "non_cn", "priority_one", "priority_two"}
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "priority_one",
			Label:          "Priority One",
			Enabled:        true,
			TargetMode:     "direct",
			SourceType:     "inline",
			Behavior:       "domain",
			Format:         "text",
			Payload:        []string{"priority-one.example"},
			InsertPosition: "before_match",
		},
		{
			Key:            "priority_two",
			Label:          "Priority Two",
			Enabled:        true,
			TargetMode:     "reject",
			SourceType:     "inline",
			Behavior:       "domain",
			Format:         "text",
			Payload:        []string{"priority-two.example"},
			InsertPosition: "after_adblock",
		},
	}
	opts.AdditionalRules = []string{"DOMAIN-SUFFIX,additional.example,DIRECT"}
	opts.RuleProviders = []model.RuleProviderConfig{
		{
			Name:     "legacy_custom",
			Type:     "http",
			Behavior: "domain",
			URL:      "https://example.com/legacy-custom.yaml",
			Policy:   "DIRECT",
			Enabled:  true,
		},
	}

	cfg := mustRenderConfig(t, nodes, opts)
	wantPrefix := []string{
		"RULE-SET,priority_one,DIRECT",
		"RULE-SET,priority_two,REJECT",
		"DOMAIN-SUFFIX,additional.example,DIRECT",
		"RULE-SET,legacy_custom,DIRECT",
	}
	if len(cfg.Rules) < len(wantPrefix)+1 {
		t.Fatalf("rules = %#v, want priority rules and built-in rules", cfg.Rules)
	}
	for index, want := range wantPrefix {
		if cfg.Rules[index] != want {
			t.Fatalf("rules[%d] = %q, want %q; rules = %#v", index, cfg.Rules[index], want, cfg.Rules)
		}
	}
	if got := cfg.Rules[len(wantPrefix)]; got != "RULE-SET,category-ads-all,🛑 广告拦截" {
		t.Fatalf("first built-in rule = %q, want adblock after custom rules", got)
	}
	if got := cfg.Rules[len(cfg.Rules)-1]; !strings.HasPrefix(got, "MATCH,") {
		t.Fatalf("last rule = %q, want MATCH", got)
	}
}

func TestRenderAdditionalRuleSetBuiltinTargets(t *testing.T) {
	nodes := []model.NodeIR{testSSNode("Test Node", "example.com")}
	opts := standardRenderOptions()
	opts.AdditionalRules = []string{
		"RULE-SET,adobe,DIRECT",
		"RULE-SET,adobe,REJECT",
	}
	opts.RuleProviders = []model.RuleProviderConfig{
		{Name: "adobe", Type: "http", Behavior: "classical", URL: "https://example.com/adobe.yaml", Policy: "DIRECT", Enabled: true},
	}

	got := string(mustRender(t, nodes, opts))
	for _, want := range []string{
		"RULE-SET,adobe,DIRECT",
		"RULE-SET,adobe,REJECT",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered output missing %q:\n%s", want, got)
		}
	}
}

func TestRenderCustomRuleExistingProxyAlias(t *testing.T) {
	nodes := []model.NodeIR{testSSNode("Test Node", "example.com")}
	opts := standardRenderOptions()
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "one",
			Label:          "One",
			Enabled:        true,
			TargetMode:     "existing_group",
			TargetGroup:    "PROXY",
			SourceType:     "inline",
			Behavior:       "domain",
			Format:         "text",
			Payload:        []string{"DOMAIN-SUFFIX,javdb.com,PROXY"},
			InsertPosition: "before_match",
		},
	}
	opts.EnabledRules = append(opts.EnabledRules, "one")

	cfg := mustRenderConfig(t, nodes, opts)
	if !containsString(cfg.Rules, "RULE-SET,one,One") {
		t.Fatalf("rules = %#v, want custom rule target group", cfg.Rules)
	}
	if containsString(cfg.Rules, "RULE-SET,one,PROXY") || containsString(cfg.Rules, "RULE-SET,one,🚀 节点选择") {
		t.Fatalf("rules = %#v, raw PROXY/main target should not remain", cfg.Rules)
	}
	if group := findProxyGroup(t, cfg.ProxyGroups, "One"); group.Name != "One" || len(group.Proxies) == 0 {
		t.Fatalf("proxy-groups = %#v, want custom rule group One", cfg.ProxyGroups)
	}
	provider, ok := cfg.RuleProviders["one"]
	if !ok {
		t.Fatalf("missing custom rule provider: %#v", cfg.RuleProviders)
	}
	if !containsString(provider.Payload, "DOMAIN-SUFFIX,javdb.com") || containsString(provider.Payload, "DOMAIN-SUFFIX,javdb.com,PROXY") {
		t.Fatalf("provider payload = %#v, want policy target stripped", provider.Payload)
	}
	if provider.Behavior != "classical" {
		t.Fatalf("provider behavior = %q, want classical for full rule payload", provider.Behavior)
	}
	if provider.Format != "text" {
		t.Fatalf("provider format = %q, want text", provider.Format)
	}
}

func TestGeneratedCustomRuleGroupsIncludeRegionGroups(t *testing.T) {
	nodes := model.NormalizeNodes([]model.NodeIR{
		testSSNode("JP Node", "jp.example.com"),
		testSSNode("US Node", "us.example.com"),
	})
	opts := standardRenderOptions()
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "media",
			Label:          "Media",
			Enabled:        true,
			TargetMode:     "new_group",
			SourceType:     "inline",
			Behavior:       "classical",
			Format:         "text",
			Payload:        []string{"DOMAIN-SUFFIX,example.com"},
			InsertPosition: "before_match",
		},
	}
	opts.EnabledRules = append(opts.EnabledRules, "media")
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{
		EnableRegionGroups: true,
		RuleGroupNodeMode:  "full",
	})

	cfg := mustRenderConfig(t, nodes, opts)
	japan := findProxyGroup(t, cfg.ProxyGroups, "🇯🇵 日本")
	us := findProxyGroup(t, cfg.ProxyGroups, "🇺🇸 美国")
	if japan.Type != "select" || us.Type != "select" {
		t.Fatalf("region groups = %#v %#v, want select", japan, us)
	}
	assertEqualProxies(t, japan.Proxies, []string{"⚡ 日本自动", "DIRECT", "JP Node"})
	assertEqualProxies(t, us.Proxies, []string{"⚡ 美国自动", "DIRECT", "US Node"})

	japanAuto := findProxyGroup(t, cfg.ProxyGroups, "⚡ 日本自动")
	usAuto := findProxyGroup(t, cfg.ProxyGroups, "⚡ 美国自动")
	if japanAuto.Type != "url-test" || usAuto.Type != "url-test" {
		t.Fatalf("region auto groups = %#v %#v, want url-test", japanAuto, usAuto)
	}
	assertEqualProxies(t, japanAuto.Proxies, []string{"JP Node"})
	assertEqualProxies(t, usAuto.Proxies, []string{"US Node"})

	custom := findProxyGroup(t, cfg.ProxyGroups, "Media")
	for _, want := range []string{"🇯🇵 日本", "🇺🇸 美国", "JP Node", "US Node"} {
		assertContainsProxy(t, custom.Proxies, want)
	}
	main := findProxyGroup(t, cfg.ProxyGroups, groupNodeSelect)
	for _, want := range []string{"🇯🇵 日本", "🇺🇸 美国"} {
		assertContainsProxy(t, main.Proxies, want)
	}
}

func TestRenderNewCustomGroupDoesNotReuseReservedTargetName(t *testing.T) {
	nodes := []model.NodeIR{testSSNode("Test Node", "example.com")}
	opts := standardRenderOptions()
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "download",
			Label:          "download",
			Icon:           "💾",
			Enabled:        true,
			TargetMode:     "new_group",
			TargetGroup:    "💾 download",
			SourceType:     "inline",
			Behavior:       "classical",
			Format:         "text",
			Payload:        []string{"DOMAIN-SUFFIX,example.com"},
			InsertPosition: "before_match",
		},
	}
	opts.EnabledRules = append(opts.EnabledRules, "download")

	cfg := mustRenderConfig(t, nodes, opts)
	if !containsString(cfg.Rules, "RULE-SET,download,💾 download") {
		t.Fatalf("rules = %#v, want download rule to target its own group", cfg.Rules)
	}
	if containsString(cfg.Rules, "RULE-SET,download,🚀 节点选择 2") {
		t.Fatalf("rules = %#v, stale reserved target name must not be reused", cfg.Rules)
	}
	findProxyGroup(t, cfg.ProxyGroups, "💾 download")
}

func TestRenderCustomHTTPRuleInfersClassicalYAMLShape(t *testing.T) {
	nodes := []model.NodeIR{testSSNode("Test Node", "example.com")}
	opts := standardRenderOptions()
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "adobe",
			Label:          "Adobe",
			Enabled:        true,
			TargetMode:     "reject",
			SourceType:     "http",
			Behavior:       "domain",
			Format:         "text",
			URL:            "https://cdn.jsdelivr.net/gh/blackmatrix7/ios_rule_script@master/rule/Clash/Adobe/Adobe.yaml",
			Interval:       86400,
			InsertPosition: "before_match",
		},
	}
	opts.RuleMode = "custom"
	opts.EnabledRules = append(opts.EnabledRules, "adobe")

	cfg := mustRenderConfig(t, nodes, opts)
	provider, ok := cfg.RuleProviders["adobe"]
	if !ok {
		t.Fatalf("missing adobe provider: %#v", cfg.RuleProviders)
	}
	if provider.Behavior != "classical" || provider.Format != "yaml" || provider.Path != "./ruleset/adobe.yaml" {
		t.Fatalf("adobe provider = %#v, want classical yaml", provider)
	}
	if !containsString(cfg.Rules, "RULE-SET,adobe,REJECT") {
		t.Fatalf("rules = %#v, want adobe reject rule", cfg.Rules)
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
	regionGroupNames := []string{"🇭🇰 香港", "🇯🇵 日本", "🇺🇸 美国", "🇸🇬 新加坡", "🇹🇼 台湾", "🇬🇧 英国", "🇩🇪 德国", "🇳🇱 荷兰", "🇷🇺 俄罗斯", "🇰🇷 韩国", "⚡ 香港自动", "⚡ 日本自动", "⚡ 美国自动", "⚡ 新加坡自动", "⚡ 台湾自动", "⚡ 英国自动", "⚡ 德国自动", "⚡ 荷兰自动", "⚡ 俄罗斯自动", "⚡ 韩国自动"}
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

func TestRegionGroupsRespectEnableSwitch(t *testing.T) {
	nodes := model.NormalizeNodes([]model.NodeIR{
		testSSNode("JP Node", "jp.example.com"),
		testSSNode("US Node", "us.example.com"),
	})
	opts := standardRenderOptions()
	opts.CustomRules = []model.CustomRule{
		{
			Key:            "media",
			Label:          "Media",
			Enabled:        true,
			TargetMode:     "new_group",
			SourceType:     "inline",
			Behavior:       "classical",
			Format:         "text",
			Payload:        []string{"DOMAIN-SUFFIX,example.com"},
			InsertPosition: "before_match",
		},
	}
	opts.EnabledRules = append(opts.EnabledRules, "media")
	opts.GroupProxyMode = "regional"
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{
		EnableRegionGroups: false,
		RuleGroupNodeMode:  "full",
	})

	cfg := mustRenderConfig(t, nodes, opts)
	for _, name := range []string{"🇯🇵 日本", "🇺🇸 美国", "⚡ 日本自动", "⚡ 美国自动"} {
		assertNoProxyGroup(t, cfg.ProxyGroups, name)
	}

	custom := findProxyGroup(t, cfg.ProxyGroups, "Media")
	for _, forbidden := range []string{"🇯🇵 日本", "🇺🇸 美国", "⚡ 日本自动", "⚡ 美国自动"} {
		assertNotContainsProxy(t, custom.Proxies, forbidden)
	}
	for _, want := range []string{"JP Node", "US Node"} {
		assertContainsProxy(t, custom.Proxies, want)
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

func TestFinalYAMLSerializationPreservesRegionGroups(t *testing.T) {
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

	rendered, err := renderConfig(cfg)
	if err != nil {
		t.Fatalf("renderConfig() error = %v", err)
	}
	yamlText := string(rendered)
	var out struct {
		ProxyGroups []mihomoProxyGroup `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal([]byte(yamlText), &out); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v\n%s", err, yamlText)
	}

	findProxyGroup(t, out.ProxyGroups, "🇭🇰 香港")
	findProxyGroup(t, out.ProxyGroups, "🇯🇵 日本")

	globalAuto := findProxyGroup(t, out.ProxyGroups, groupAutoSelect)
	assertEqualProxies(t, globalAuto.Proxies, []string{"HK Node"})
	hkAuto := findProxyGroup(t, out.ProxyGroups, "⚡ 香港自动")
	assertEqualProxies(t, hkAuto.Proxies, []string{"HK Node"})

	main := findProxyGroup(t, out.ProxyGroups, groupNodeSelect)
	for _, want := range []string{"🇭🇰 香港", "🇯🇵 日本"} {
		assertContainsProxy(t, main.Proxies, want)
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
