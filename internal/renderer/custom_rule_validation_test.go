package renderer

import (
	"reflect"
	"strings"
	"testing"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
)

func TestRenderPromotesAdvancedClassicalRulesToClassicalProvider(t *testing.T) {
	for _, payload := range []string{"IN-PORT,7890", "DSCP,10", "PROCESS-NAME-WILDCARD,firefox*", "REMATCH-NAME,relay"} {
		t.Run(payload, func(t *testing.T) {
			opts := standardRenderOptions()
			opts.RuleMode = "custom"
			opts.CustomRules = []model.CustomRule{{Key: "advanced", Label: "Advanced", Enabled: true, TargetMode: "direct", SourceType: "inline", Behavior: "domain", Format: "text", Payload: []string{payload}}}
			cfg := mustRenderConfig(t, []model.NodeIR{testSSNode("Test", "example.com")}, opts)
			if provider := cfg.RuleProviders["advanced"]; provider.Behavior != "classical" {
				t.Fatalf("provider behavior=%q, want classical for %q", provider.Behavior, payload)
			}
		})
	}
}

func TestValidatedPastedRulesRenderAsSeparateCanonicalEntries(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.RuleMode = "custom"
	cfg.Render.AdditionalRules = []string{"# direct rules\n- 'DOMAIN,first.example,PROXY' # note\r\nIP-CIDR,192.0.2.0/24,PROXY,SRC,NO-RESOLVE"}
	cfg.Render.CustomRules = []model.CustomRule{{Key: "pasted", Label: "Pasted", Enabled: true, TargetMode: "direct", SourceType: "inline", Behavior: "classical", Payload: []string{"# provider rules\n- 'DOMAIN-SUFFIX,second.example,DIRECT' # note\r\nIP-CIDR,198.51.100.0/24,PROXY,src,no-resolve"}}}
	cfg = config.Normalize(cfg)
	if err := config.Validate(cfg); err != nil {
		t.Fatal(err)
	}
	opts := standardRenderOptions()
	opts.RuleMode = cfg.Render.RuleMode
	opts.AdditionalRules = cfg.Render.AdditionalRules
	opts.CustomRules = cfg.Render.CustomRules
	got := mustRenderConfig(t, []model.NodeIR{testSSNode("Test", "example.com")}, opts)
	for _, want := range []string{"DOMAIN,first.example,🚀 节点选择", "IP-CIDR,192.0.2.0/24,🚀 节点选择,src,no-resolve"} {
		if !containsString(got.Rules, want) {
			t.Fatalf("rendered rules do not contain %q: %q", want, got.Rules)
		}
	}
	want := []string{"DOMAIN-SUFFIX,second.example", "IP-CIDR,198.51.100.0/24,src,no-resolve"}
	if !reflect.DeepEqual(got.RuleProviders["pasted"].Payload, want) {
		t.Fatalf("payload = %q, want %q", got.RuleProviders["pasted"].Payload, want)
	}
	for _, rule := range got.Rules {
		if strings.ContainsAny(rule, "\r\n") {
			t.Fatalf("multiple rules combined in one YAML scalar: %q", rule)
		}
	}
}

func TestProviderRegexKeepsLiteralPolicySuffix(t *testing.T) {
	opts := standardRenderOptions()
	opts.CustomRules = []model.CustomRule{{Key: "regex", Label: "Regex", Enabled: true, TargetMode: "direct", SourceType: "inline", Behavior: "classical", Payload: []string{`DOMAIN-REGEX,^prefix,DIRECT`, `PROCESS-NAME-REGEX,^program,REJECT`}}}
	got := mustRenderConfig(t, []model.NodeIR{testSSNode("Test", "example.com")}, opts)
	if !reflect.DeepEqual(got.RuleProviders["regex"].Payload, opts.CustomRules[0].Payload) {
		t.Fatalf("provider regex was altered: %q", got.RuleProviders["regex"].Payload)
	}
}
