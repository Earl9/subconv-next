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
