package renderer

import "testing"

func TestValidateMihomoConfigWarnings(t *testing.T) {
	cfg := mihomoConfig{
		Proxies: []mihomoProxy{
			{Name: "any-1", Type: "anytls"},
			{Name: "xhttp-1", Type: "vless", Network: "xhttp"},
		},
		ProxyGroups: []mihomoProxyGroup{
			{Name: "节点选择", Type: "select", Proxies: []string{"missing-proxy"}},
			{Name: "节点选择", Type: "select", Proxies: []string{"DIRECT"}},
		},
		Rules: []string{
			"RULE-SET,google,谷歌服务",
		},
	}

	warnings := ValidateMihomoConfig(cfg)
	got := map[string]bool{}
	for _, warning := range warnings {
		got[warning.Code] = true
	}

	wantCodes := []string{
		"anytls_missing_client_fingerprint",
		"vless_xhttp_missing_opts",
		"duplicate_group_name",
		"missing_proxy_reference",
		"rules_without_rule_providers",
		"missing_rule_target",
	}
	for _, code := range wantCodes {
		if !got[code] {
			t.Fatalf("missing warning code %q in %#v", code, warnings)
		}
	}
}

func TestFailOnCriticalWarnings(t *testing.T) {
	err := failOnCriticalWarnings([]ValidationWarning{
		{Code: "missing_proxy_reference", Message: "broken"},
		{Code: "anytls_missing_client_fingerprint", Message: "warn"},
	})
	if err == nil {
		t.Fatalf("failOnCriticalWarnings() error = nil, want non-nil")
	}
}

func TestValidateMihomoConfigRegionGroupRoleWarnings(t *testing.T) {
	cfg := mihomoConfig{
		Proxies: []mihomoProxy{
			{Name: "HK Node", Type: "ss"},
			{Name: "JP Node", Type: "ss"},
		},
		ProxyGroups: []mihomoProxyGroup{
			{Name: groupNodeSelect, Type: "select", Proxies: []string{groupAutoSelect, "🇭🇰 香港"}},
			{Name: groupAutoSelect, Type: "url-test", Proxies: []string{"HK Node", "🇭🇰 香港"}},
			{Name: "⚡ 香港自动", Type: "url-test", Proxies: []string{"HK Node", groupAutoSelect}},
			{Name: "🇭🇰 香港", Type: "select", Proxies: []string{groupAutoSelect, groupNodeSelect, "⚡ 日本自动", "HK Node"}},
			{Name: "⚡ 日本自动", Type: "url-test", Proxies: []string{"JP Node"}},
			{Name: "🇯🇵 日本", Type: "select", Proxies: []string{"DIRECT", "JP Node"}},
		},
		Rules: []string{"MATCH," + groupNodeSelect},
	}

	warnings := ValidateMihomoConfig(cfg)
	got := map[string]bool{}
	for _, warning := range warnings {
		got[warning.Code] = true
	}

	for _, code := range []string{
		"invalid_global_auto_reference",
		"invalid_region_auto_reference",
		"invalid_region_group_reference",
	} {
		if !got[code] {
			t.Fatalf("missing warning code %q in %#v", code, warnings)
		}
	}
}
