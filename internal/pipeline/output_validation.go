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

func ValidateOutputNoLeak(yamlBytes []byte, finalNodes FinalNodeSet, auditReport model.AuditReport, renderOpts model.RenderOptions) error {
	return ValidateFinalConfig(yamlBytes, finalNodes, auditReport, renderOpts)
}

func ValidateFinalNodeSet(finalNodes FinalNodeSet, auditReport model.AuditReport) error {
	finalNameSet := map[string]struct{}{}
	for _, node := range finalNodes.Nodes {
		name := strings.TrimSpace(node.Name)
		if name == "" {
			return fmt.Errorf("final node missing name")
		}
		if !isRenderableNode(node) {
			return fmt.Errorf("invalid node %q leaked into final nodes", name)
		}
		finalNameSet[name] = struct{}{}
	}

	for _, excluded := range auditReport.ExcludedNodes {
		name := strings.TrimSpace(excluded.Name)
		if name == "" || excluded.Reason == "duplicate" {
			continue
		}
		if _, ok := finalNameSet[name]; ok {
			return fmt.Errorf("excluded node %q leaked into final nodes: %s", name, excluded.Reason)
		}
	}

	return nil
}

func ValidateFinalConfig(yamlBytes []byte, finalNodes FinalNodeSet, auditReport model.AuditReport, renderOpts model.RenderOptions) error {
	if err := ValidateFinalNodeSet(finalNodes, auditReport); err != nil {
		return err
	}

	var out outputConfig
	if err := yaml.Unmarshal(yamlBytes, &out); err != nil {
		return fmt.Errorf("parse rendered yaml: %w", err)
	}

	allowedNames := map[string]struct{}{}
	realNames := map[string]struct{}{}
	infoNames := map[string]struct{}{}
	for _, node := range finalNodes.Nodes {
		allowedNames[node.Name] = struct{}{}
		if isInfoNode(node) {
			infoNames[node.Name] = struct{}{}
		} else {
			realNames[node.Name] = struct{}{}
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

	var mainSelect *outputProxyGroup
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
		groupName := strings.TrimSpace(group.Name)
		if groupName == "🚀 节点选择" {
			current := group
			mainSelect = &current
		}
		if isRegionProxyGroupName(group.Name) {
			return fmt.Errorf("region proxy-group %q must not be generated", group.Name)
		}
		if isKnownRegionAutoProxyGroupName(group.Name) {
			return fmt.Errorf("region auto proxy-group %q must not be generated", group.Name)
		}
		if strings.TrimSpace(group.Name) == "⚡ 自动选择" {
			for _, proxyName := range group.Proxies {
				proxyName = strings.TrimSpace(proxyName)
				if proxyName == "" {
					continue
				}
				if _, ok := realNames[proxyName]; !ok {
					return fmt.Errorf("auto proxy-group %q must only reference real node %q", group.Name, proxyName)
				}
			}
		}
		if shouldRequireRealNodeInPolicyGroup(group.Name, renderOpts) && !containsAnyRealNode(group.Proxies, realNames) {
			return fmt.Errorf("proxy-group %q must include at least one real node in full mode", group.Name)
		}
		if isSpecialCompactPolicyGroup(group.Name) {
			if err := validateSpecialCompactPolicyGroup(group, realNames); err != nil {
				return err
			}
		}
		for _, proxyName := range group.Proxies {
			proxyName = strings.TrimSpace(proxyName)
			if proxyName == "" {
				continue
			}
			switch proxyName {
			case "DIRECT", "REJECT":
				continue
			}
			if reason, leaked := excludedNames[proxyName]; leaked {
				return fmt.Errorf("excluded node %q leaked into proxy-group %q: %s", proxyName, group.Name, reason)
			}
			if _, ok := groupNames[proxyName]; ok {
				continue
			}
			if _, ok := allowedNames[proxyName]; !ok {
				return fmt.Errorf("proxy-group %q references unknown node/group %q", group.Name, proxyName)
			}
			if _, info := infoNames[proxyName]; info && (groupType == "url-test" || groupType == "fallback" || groupType == "load-balance") {
				return fmt.Errorf("info node %q leaked into %s group %q", proxyName, groupType, group.Name)
			}
		}
	}

	if err := validateMainSelectGroup(mainSelect, realNames); err != nil {
		return err
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
		case "", "DIRECT", "REJECT":
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

func validateMainSelectGroup(group *outputProxyGroup, realNames map[string]struct{}) error {
	if group == nil {
		return fmt.Errorf("main proxy-group %q must be generated", "🚀 节点选择")
	}
	for _, required := range []string{"⚡ 自动选择", "DIRECT", "REJECT"} {
		if !containsProxyRef(group.Proxies, required) {
			return fmt.Errorf("main proxy-group %q must contain %q", group.Name, required)
		}
	}
	if !containsAnyRealNode(group.Proxies, realNames) {
		return fmt.Errorf("main proxy-group %q must include at least one real node", group.Name)
	}
	return nil
}

func isSpecialCompactPolicyGroup(groupName string) bool {
	switch strings.TrimSpace(groupName) {
	case "🛑 广告拦截", "🏠 私有网络", "🔒 国内服务":
		return true
	default:
		return false
	}
}

func validateSpecialCompactPolicyGroup(group outputProxyGroup, realNames map[string]struct{}) error {
	if containsAnyRealNode(group.Proxies, realNames) {
		return fmt.Errorf("special proxy-group %q must stay compact and must not include real nodes", group.Name)
	}
	switch strings.TrimSpace(group.Name) {
	case "🛑 广告拦截":
		return requireExactProxyRefs(group, []string{"REJECT", "DIRECT", "🚀 节点选择"})
	case "🏠 私有网络", "🔒 国内服务":
		return requireExactProxyRefs(group, []string{"DIRECT", "REJECT", "🚀 节点选择", "⚡ 自动选择"})
	default:
		return nil
	}
}

func requireExactProxyRefs(group outputProxyGroup, want []string) error {
	if len(group.Proxies) != len(want) {
		return fmt.Errorf("special proxy-group %q must stay compact: got %v, want %v", group.Name, group.Proxies, want)
	}
	for i := range want {
		if strings.TrimSpace(group.Proxies[i]) != want[i] {
			return fmt.Errorf("special proxy-group %q must stay compact: got %v, want %v", group.Name, group.Proxies, want)
		}
	}
	return nil
}

func containsProxyRef(proxies []string, want string) bool {
	for _, proxyName := range proxies {
		if strings.TrimSpace(proxyName) == want {
			return true
		}
	}
	return false
}

func shouldRequireRealNodeInPolicyGroup(groupName string, opts model.RenderOptions) bool {
	groupOpt := model.NormalizeGroupOptions(opts.GroupOptions)
	if groupOpt.RuleGroupNodeMode != "full" {
		return false
	}
	groupName = strings.TrimSpace(groupName)
	switch groupName {
	case "", "🛑 广告拦截", "🏠 私有网络", "🔒 国内服务", "🚀 节点选择", "⚡ 自动选择":
		return false
	case "🤖 AI 服务", "📹 油管视频", "📚 教育学术", "☁️ 云服务", "🔍 谷歌服务", "📲 电报消息", "🐱 代码托管", "Ⓜ️ 微软服务", "🍏 苹果服务", "🐦 推特/X", "📘 Meta 系", "🎙️ Discord", "💬 其他社交", "🎬 奈飞", "🏰 迪士尼+", "📺 欧美流媒体", "🎌 亚洲流媒体", "🎮 Steam", "🖥️ PC 游戏", "🎯 主机游戏", "🛠️ 开发工具", "💾 网盘存储", "💳 支付平台", "₿ 加密货币", "📰 新闻资讯", "🛒 海淘购物", "🌍 非中国", "🐟 漏网之鱼":
		return true
	default:
		return false
	}
}

func containsAnyRealNode(proxies []string, realNames map[string]struct{}) bool {
	for _, proxyName := range proxies {
		if _, ok := realNames[strings.TrimSpace(proxyName)]; ok {
			return true
		}
	}
	return false
}

func isRegionProxyGroupName(name string) bool {
	name = strings.TrimSpace(name)
	regionPatterns := []struct {
		flag  string
		label string
	}{
		{flag: "🇭🇰", label: "香港"},
		{flag: "🇯🇵", label: "日本"},
		{flag: "🇺🇸", label: "美国"},
		{flag: "🇸🇬", label: "新加坡"},
		{flag: "🇹🇼", label: "台湾"},
		{flag: "🇬🇧", label: "英国"},
		{flag: "🇩🇪", label: "德国"},
		{flag: "🇳🇱", label: "荷兰"},
		{flag: "🇷🇺", label: "俄罗斯"},
		{flag: "🇰🇷", label: "韩国"},
		{flag: "🇫🇷", label: "法国"},
		{flag: "🇨🇦", label: "加拿大"},
		{flag: "🇦🇺", label: "澳大利亚"},
	}
	for _, pattern := range regionPatterns {
		if strings.HasPrefix(name, pattern.flag) && strings.Contains(name, pattern.label) {
			return true
		}
	}
	return false
}

func isKnownRegionAutoProxyGroupName(name string) bool {
	switch strings.TrimSpace(name) {
	case "⚡ 香港自动", "⚡ 日本自动", "⚡ 美国自动", "⚡ 新加坡自动", "⚡ 台湾自动", "⚡ 英国自动", "⚡ 德国自动", "⚡ 荷兰自动", "⚡ 俄罗斯自动", "⚡ 韩国自动", "⚡ 法国自动", "⚡ 加拿大自动", "⚡ 澳大利亚自动":
		return true
	default:
		return false
	}
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
