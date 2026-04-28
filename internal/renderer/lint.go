package renderer

import (
	"fmt"
	"strings"
)

type ValidationWarning struct {
	Code    string
	Message string
}

func ValidateMihomoConfig(cfg mihomoConfig) []ValidationWarning {
	var warnings []ValidationWarning

	proxyNames := make(map[string]struct{}, len(cfg.Proxies))
	for _, proxy := range cfg.Proxies {
		if _, exists := proxyNames[proxy.Name]; exists {
			warnings = append(warnings, ValidationWarning{
				Code:    "duplicate_proxy_name",
				Message: fmt.Sprintf("duplicate proxy name %q", proxy.Name),
			})
		}
		proxyNames[proxy.Name] = struct{}{}

		if proxy.Type == "anytls" && proxy.RealityOpts != nil {
			warnings = append(warnings, ValidationWarning{
				Code:    "anytls_reality_forbidden",
				Message: fmt.Sprintf("anytls proxy %q must not contain reality-opts", proxy.Name),
			})
		}
		if proxy.Type == "anytls" && proxy.ClientFingerprint == "" {
			warnings = append(warnings, ValidationWarning{
				Code:    "anytls_missing_client_fingerprint",
				Message: fmt.Sprintf("anytls proxy %q missing client-fingerprint", proxy.Name),
			})
		}
		if proxy.Type == "vless" && strings.EqualFold(strings.TrimSpace(proxy.Network), "xhttp") && proxy.XHTTPOpts == nil {
			warnings = append(warnings, ValidationWarning{
				Code:    "vless_xhttp_missing_opts",
				Message: fmt.Sprintf("vless proxy %q uses xhttp without xhttp-opts", proxy.Name),
			})
		}
	}

	groupNames := make(map[string]struct{}, len(cfg.ProxyGroups))
	for _, group := range cfg.ProxyGroups {
		if _, exists := groupNames[group.Name]; exists {
			warnings = append(warnings, ValidationWarning{
				Code:    "duplicate_group_name",
				Message: fmt.Sprintf("duplicate proxy-group name %q", group.Name),
			})
		}
		groupNames[group.Name] = struct{}{}
	}

	providerNames := make(map[string]struct{}, len(cfg.RuleProviders))
	providerPaths := make(map[string]string, len(cfg.RuleProviders))
	for name, provider := range cfg.RuleProviders {
		if _, exists := providerNames[name]; exists {
			warnings = append(warnings, ValidationWarning{
				Code:    "duplicate_rule_provider_name",
				Message: fmt.Sprintf("duplicate rule-provider name %q", name),
			})
		}
		providerNames[name] = struct{}{}
		if provider.Path != "" {
			if existingName, exists := providerPaths[provider.Path]; exists && existingName != name {
				warnings = append(warnings, ValidationWarning{
					Code:    "duplicate_rule_provider_path",
					Message: fmt.Sprintf("rule-provider path %q duplicated by %q and %q", provider.Path, existingName, name),
				})
			}
			providerPaths[provider.Path] = name
		}
		if strings.EqualFold(strings.TrimSpace(provider.Format), "mrs") {
			behavior := strings.ToLower(strings.TrimSpace(provider.Behavior))
			if behavior != "domain" && behavior != "ipcidr" {
				warnings = append(warnings, ValidationWarning{
					Code:    "invalid_mrs_behavior",
					Message: fmt.Sprintf("rule-provider %q with mrs format must use behavior domain or ipcidr", name),
				})
			}
		}
	}

	for _, group := range cfg.ProxyGroups {
		for _, ref := range group.Proxies {
			if isBuiltinProxyReference(ref) {
				continue
			}
			if _, ok := proxyNames[ref]; ok {
				continue
			}
			if _, ok := groupNames[ref]; ok {
				continue
			}
			warnings = append(warnings, ValidationWarning{
				Code:    "missing_proxy_reference",
				Message: fmt.Sprintf("proxy-group %q references unknown proxy/group %q", group.Name, ref),
			})
		}
	}

	hasRuleSetRule := false
	matchIndex := -1
	for index, rule := range cfg.Rules {
		parts := splitCSVRule(rule)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "RULE-SET":
			hasRuleSetRule = true
			if len(parts) < 3 {
				warnings = append(warnings, ValidationWarning{
					Code:    "invalid_rule_set",
					Message: fmt.Sprintf("rule %q is missing provider or target group", rule),
				})
				continue
			}
			if _, ok := providerNames[parts[1]]; !ok {
				warnings = append(warnings, ValidationWarning{
					Code:    "missing_rule_provider_reference",
					Message: fmt.Sprintf("rule %q references missing rule-provider %q", rule, parts[1]),
				})
			}
			if _, ok := groupNames[parts[2]]; !ok {
				warnings = append(warnings, ValidationWarning{
					Code:    "missing_rule_target",
					Message: fmt.Sprintf("rule %q references unknown group %q", rule, parts[2]),
				})
			}
		case "MATCH":
			matchIndex = index
		default:
			if len(parts) >= 3 && !isBuiltinProxyReference(parts[2]) {
				if _, ok := groupNames[parts[2]]; !ok {
					warnings = append(warnings, ValidationWarning{
						Code:    "missing_rule_target",
						Message: fmt.Sprintf("rule %q references unknown group %q", rule, parts[2]),
					})
				}
			}
		}
	}

	if hasRuleSetRule && len(cfg.RuleProviders) == 0 {
		warnings = append(warnings, ValidationWarning{
			Code:    "rules_without_rule_providers",
			Message: "enabled rules generated RULE-SET entries but rule-providers are empty",
		})
	}
	if matchIndex == -1 {
		warnings = append(warnings, ValidationWarning{
			Code:    "missing_match_rule",
			Message: "rules must contain a final MATCH entry",
		})
	} else if matchIndex != len(cfg.Rules)-1 {
		warnings = append(warnings, ValidationWarning{
			Code:    "match_not_last",
			Message: "MATCH rule must be the last rule",
		})
	}

	return warnings
}

func failOnCriticalWarnings(warnings []ValidationWarning) error {
	var critical []string
	for _, warning := range warnings {
		switch warning.Code {
		case "duplicate_proxy_name",
			"duplicate_group_name",
			"duplicate_rule_provider_name",
			"duplicate_rule_provider_path",
			"missing_proxy_reference",
			"missing_rule_provider_reference",
			"missing_rule_target",
			"invalid_rule_set",
			"missing_match_rule",
			"match_not_last",
			"invalid_mrs_behavior":
			critical = append(critical, warning.Message)
		}
	}
	if len(critical) == 0 {
		return nil
	}
	return fmt.Errorf("renderer validation failed: %s", strings.Join(critical, "; "))
}

func splitCSVRule(rule string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(rule); i++ {
		if rule[i] == ',' {
			parts = append(parts, rule[start:i])
			start = i + 1
		}
	}
	parts = append(parts, rule[start:])
	return parts
}

func isBuiltinProxyReference(value string) bool {
	switch value {
	case "DIRECT", "REJECT", "REJECT-DROP", "PASS":
		return true
	default:
		return false
	}
}
