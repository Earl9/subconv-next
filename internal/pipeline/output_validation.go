package pipeline

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"subconv-next/internal/model"
)

type outputConfig struct {
	Proxies       []map[string]interface{} `yaml:"proxies"`
	ProxyGroups   []outputProxyGroup       `yaml:"proxy-groups"`
	RuleProviders map[string]interface{}   `yaml:"rule-providers"`
	Rules         []string                 `yaml:"rules"`
}

type outputProxyGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Proxies []string `yaml:"proxies"`
}

func ValidateOutputNoLeak(yamlBytes []byte, finalNodes FinalNodeSet, auditReport model.AuditReport, _ model.RenderOptions) error {
	var out outputConfig
	if err := yaml.Unmarshal(yamlBytes, &out); err != nil {
		return fmt.Errorf("parse rendered yaml: %w", err)
	}

	allowedNames := map[string]struct{}{}
	infoNames := map[string]struct{}{}
	for _, node := range finalNodes.Nodes {
		allowedNames[node.Name] = struct{}{}
		if isInfoNode(node) {
			infoNames[node.Name] = struct{}{}
		}
	}

	excludedNames := map[string]string{}
	for _, excluded := range auditReport.ExcludedNodes {
		name := strings.TrimSpace(excluded.Name)
		if name == "" {
			continue
		}
		if excluded.Reason == "duplicate" {
			continue
		}
		excludedNames[name] = excluded.Reason
	}

	groupNames := map[string]struct{}{}
	for _, group := range out.ProxyGroups {
		groupNames[strings.TrimSpace(group.Name)] = struct{}{}
	}

	for _, proxy := range out.Proxies {
		name := strings.TrimSpace(fmt.Sprint(proxy["name"]))
		if name == "" {
			return fmt.Errorf("rendered proxy missing name")
		}
		if _, ok := allowedNames[name]; !ok {
			return fmt.Errorf("rendered proxy %q not found in final nodes", name)
		}
		if reason, leaked := excludedNames[name]; leaked {
			return fmt.Errorf("excluded node %q leaked into proxies: %s", name, reason)
		}
	}

	for _, group := range out.ProxyGroups {
		groupType := strings.ToLower(strings.TrimSpace(group.Type))
		for _, proxyName := range group.Proxies {
			proxyName = strings.TrimSpace(proxyName)
			if proxyName == "" {
				continue
			}
			switch proxyName {
			case "DIRECT", "REJECT", "REJECT-DROP":
				continue
			}
			if _, ok := groupNames[proxyName]; ok {
				continue
			}
			if _, ok := allowedNames[proxyName]; !ok {
				return fmt.Errorf("proxy-group %q references unknown node/group %q", group.Name, proxyName)
			}
			if reason, leaked := excludedNames[proxyName]; leaked {
				return fmt.Errorf("excluded node %q leaked into proxy-group %q: %s", proxyName, group.Name, reason)
			}
			if _, info := infoNames[proxyName]; info && (groupType == "url-test" || groupType == "fallback" || groupType == "load-balance") {
				return fmt.Errorf("info node %q leaked into %s group %q", proxyName, groupType, group.Name)
			}
		}
	}

	for providerName := range collectRuleSetProviders(out.Rules) {
		if providerName == "MATCH" {
			continue
		}
		if _, ok := out.RuleProviders[providerName]; !ok {
			return fmt.Errorf("rules reference missing provider %q", providerName)
		}
	}

	for targetGroup := range collectRuleTargets(out.Rules) {
		switch targetGroup {
		case "", "DIRECT", "REJECT", "REJECT-DROP":
			continue
		}
		if _, ok := groupNames[targetGroup]; !ok {
			return fmt.Errorf("rules reference missing group %q", targetGroup)
		}
	}

	if len(out.Rules) > 0 {
		lastIndex := len(out.Rules) - 1
		for index, rule := range out.Rules {
			parts := strings.Split(strings.TrimSpace(rule), ",")
			if len(parts) == 0 || !strings.EqualFold(strings.TrimSpace(parts[0]), "MATCH") {
				continue
			}
			if index != lastIndex {
				return fmt.Errorf("MATCH rule must be last")
			}
		}
		if !strings.HasPrefix(strings.TrimSpace(out.Rules[lastIndex]), "MATCH,") {
			return fmt.Errorf("MATCH rule must be last")
		}
	}

	return nil
}

func collectRuleSetProviders(rules []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, rule := range rules {
		parts := strings.Split(strings.TrimSpace(rule), ",")
		if len(parts) < 3 {
			continue
		}
		if strings.EqualFold(parts[0], "RULE-SET") {
			out[strings.TrimSpace(parts[1])] = struct{}{}
		}
	}
	return out
}

func collectRuleTargets(rules []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, rule := range rules {
		parts := strings.Split(strings.TrimSpace(rule), ",")
		if len(parts) < 2 {
			continue
		}
		if strings.EqualFold(parts[0], "MATCH") {
			out[strings.TrimSpace(parts[1])] = struct{}{}
			continue
		}
		if len(parts) >= 3 {
			out[strings.TrimSpace(parts[2])] = struct{}{}
		}
	}
	return out
}
