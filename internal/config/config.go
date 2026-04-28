package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"subconv-next/internal/model"
)

func Load(path string) (model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return LoadJSONBytes(data)
	default:
		return LoadUCIBytes(data)
	}
}

func normalizeConfig(cfg model.Config) model.Config {
	cfg.Service.SubscriptionToken = strings.TrimSpace(cfg.Service.SubscriptionToken)
	if cfg.Subscriptions == nil {
		cfg.Subscriptions = []model.SubscriptionConfig{}
	}
	if cfg.Inline == nil {
		cfg.Inline = []model.InlineConfig{}
	}
	if cfg.Render.AdditionalRules == nil {
		cfg.Render.AdditionalRules = []string{}
	}
	if cfg.Render.EnabledRules == nil {
		cfg.Render.EnabledRules = []string{}
	}
	if cfg.Render.CustomRules == nil {
		cfg.Render.CustomRules = []model.CustomRule{}
	}
	if strings.TrimSpace(cfg.Render.ExternalConfig.TemplateKey) == "" {
		cfg.Render.ExternalConfig.TemplateKey = "none"
	}
	if strings.TrimSpace(cfg.Render.ExternalConfig.TemplateLabel) == "" {
		cfg.Render.ExternalConfig.TemplateLabel = "不选择，由接口提供方提供"
	}
	if strings.TrimSpace(cfg.Render.SourceMode) == "" {
		switch strings.TrimSpace(cfg.Render.TemplateRuleMode) {
		case "template", "rules":
			cfg.Render.SourceMode = cfg.Render.TemplateRuleMode
		default:
			cfg.Render.SourceMode = "rules"
		}
	}
	if strings.TrimSpace(cfg.Render.TemplateRuleMode) == "" {
		cfg.Render.TemplateRuleMode = "rules"
	}
	if strings.TrimSpace(cfg.Render.GeodataLoader) == "" {
		cfg.Render.GeodataLoader = "standard"
	}
	if cfg.Render.GeoUpdateInterval == 0 {
		cfg.Render.GeoUpdateInterval = 24
	}
	if cfg.Render.GeoxURL == nil {
		defaults := model.DefaultRenderConfig()
		cfg.Render.GeoxURL = defaults.GeoxURL
	}
	if cfg.Render.RuleProviders == nil {
		cfg.Render.RuleProviders = []model.RuleProviderConfig{}
	}
	if cfg.Render.CustomProxyGroups == nil {
		cfg.Render.CustomProxyGroups = []model.CustomProxyGroupConfig{}
	}
	if !cfg.Render.SourcePrefix {
		// keep explicit false as-is
	} else {
		cfg.Render.SourcePrefix = true
	}
	if strings.TrimSpace(cfg.Render.SourcePrefixFormat) == "" {
		cfg.Render.SourcePrefixFormat = "[{source}] {name}"
	}
	if strings.TrimSpace(cfg.Render.SourcePrefixSeparator) == "" {
		cfg.Render.SourcePrefixSeparator = " "
	}
	if strings.TrimSpace(cfg.Render.DedupeScope) == "" {
		cfg.Render.DedupeScope = "global"
	}
	for i := range cfg.Render.CustomRules {
		rule := &cfg.Render.CustomRules[i]
		rule.Key = strings.TrimSpace(strings.ToLower(rule.Key))
		rule.Label = strings.TrimSpace(rule.Label)
		if strings.TrimSpace(rule.Icon) == "" {
			rule.Icon = strings.TrimSpace(rule.Emoji)
		}
		rule.Icon = strings.TrimSpace(rule.Icon)
		if !rule.Enabled {
			// preserve explicit false only when source was set later; default to true for legacy empty rules
			if rule.Key != "" || rule.Label != "" || rule.SourceType != "" {
				rule.Enabled = rule.Enabled
			} else {
				rule.Enabled = true
			}
		} else {
			rule.Enabled = true
		}
		if strings.TrimSpace(rule.TargetMode) == "" {
			rule.TargetMode = "new_group"
		}
		if strings.TrimSpace(rule.SourceType) == "" {
			if len(rule.Payload) == 0 && strings.TrimSpace(rule.URL) == "" && strings.TrimSpace(rule.Path) == "" {
				rule.SourceType = "group_only"
			} else {
				rule.SourceType = "inline"
			}
		}
		if strings.TrimSpace(rule.Behavior) == "" {
			rule.Behavior = "domain"
		}
		if strings.TrimSpace(rule.Format) == "" {
			rule.Format = "text"
		}
		if rule.Interval <= 0 {
			rule.Interval = 86400
		}
		if strings.TrimSpace(rule.InsertPosition) == "" {
			rule.InsertPosition = "before_match"
		}
		if strings.TrimSpace(rule.TargetGroup) == "" {
			rule.TargetGroup = defaultCustomRuleTargetGroup(*rule)
		}
	}

	seenSubNames := map[string]int{}
	for i := range cfg.Subscriptions {
		sub := &cfg.Subscriptions[i]
		if strings.TrimSpace(sub.Name) == "" {
			sub.Name = fmt.Sprintf("source-%d", i+1)
		}
		sub.Name = uniqueDisplayName(sub.Name, seenSubNames)
		if strings.TrimSpace(sub.ID) == "" {
			sub.ID = stableConfigID("sub", sub.Name, sub.URL, i)
		}
		if strings.TrimSpace(sub.UserAgent) == "" {
			sub.UserAgent = model.DefaultUserAgent + " Docker"
		}
	}

	seenInlineNames := map[string]int{}
	for i := range cfg.Inline {
		inline := &cfg.Inline[i]
		if strings.TrimSpace(inline.Name) == "" {
			inline.Name = fmt.Sprintf("manual-%d", i+1)
		}
		inline.Name = uniqueDisplayName(inline.Name, seenInlineNames)
		if strings.TrimSpace(inline.ID) == "" {
			inline.ID = stableConfigID("inline", inline.Name, inline.Content, i)
		}
	}
	return cfg
}

func Normalize(cfg model.Config) model.Config {
	return normalizeConfig(cfg)
}

func Validate(cfg model.Config) error {
	if strings.TrimSpace(cfg.Service.ListenAddr) == "" {
		return fmt.Errorf("service.listen_addr must not be empty")
	}
	if cfg.Service.ListenPort < 1 || cfg.Service.ListenPort > 65535 {
		return fmt.Errorf("service.listen_port must be between 1 and 65535")
	}
	if cfg.Service.RefreshInterval < 0 {
		return fmt.Errorf("service.refresh_interval must be >= 0")
	}
	if cfg.Service.MaxSubscriptionBytes < 1 {
		return fmt.Errorf("service.max_subscription_bytes must be >= 1")
	}
	if cfg.Service.FetchTimeoutSeconds < 1 {
		return fmt.Errorf("service.fetch_timeout_seconds must be >= 1")
	}
	if !isAllowedTemplate(cfg.Service.Template) {
		return fmt.Errorf("service.template must be one of lite, standard, full")
	}
	if !filepath.IsAbs(cfg.Service.OutputPath) {
		return fmt.Errorf("service.output_path must be an absolute path")
	}
	if !filepath.IsAbs(cfg.Service.CacheDir) {
		return fmt.Errorf("service.cache_dir must be an absolute path")
	}

	for i, sub := range cfg.Subscriptions {
		if strings.TrimSpace(sub.ID) == "" {
			return fmt.Errorf("subscriptions[%d].id is required", i)
		}
		if strings.TrimSpace(sub.Name) == "" {
			return fmt.Errorf("subscriptions[%d].name is required", i)
		}
		if strings.TrimSpace(sub.URL) == "" {
			return fmt.Errorf("subscriptions[%d].url is required", i)
		}
		if err := validateSubscriptionURL(sub.URL); err != nil {
			return fmt.Errorf("subscriptions[%d].url: %w", i, err)
		}
		if strings.TrimSpace(sub.UserAgent) == "" {
			return fmt.Errorf("subscriptions[%d].user_agent must not be empty", i)
		}
	}

	for i, inline := range cfg.Inline {
		if strings.TrimSpace(inline.ID) == "" {
			return fmt.Errorf("inline[%d].id is required", i)
		}
		if strings.TrimSpace(inline.Name) == "" {
			return fmt.Errorf("inline[%d].name is required", i)
		}
	}

	if !isAllowedDedupeScope(cfg.Render.DedupeScope) {
		return fmt.Errorf("render.dedupe_scope must be one of global, per_source, none")
	}
	if err := validateCustomRules(cfg.Render.CustomRules); err != nil {
		return err
	}

	seenProviders := make(map[string]struct{}, len(cfg.Render.RuleProviders))
	for i, provider := range cfg.Render.RuleProviders {
		if strings.TrimSpace(provider.Name) == "" {
			return fmt.Errorf("render.rule_providers[%d].name is required", i)
		}
		if _, ok := seenProviders[provider.Name]; ok {
			return fmt.Errorf("render.rule_providers[%d].name must be unique", i)
		}
		seenProviders[provider.Name] = struct{}{}
		if !isAllowedProviderType(provider.Type) {
			return fmt.Errorf("render.rule_providers[%d].type must be one of http, file, inline", i)
		}
		if !isAllowedProviderBehavior(provider.Behavior) {
			return fmt.Errorf("render.rule_providers[%d].behavior must be one of domain, ipcidr, classical", i)
		}
		if provider.Format != "" && !isAllowedProviderFormat(provider.Format) {
			return fmt.Errorf("render.rule_providers[%d].format must be one of yaml, text, mrs", i)
		}
		if strings.TrimSpace(provider.Policy) == "" {
			return fmt.Errorf("render.rule_providers[%d].policy is required", i)
		}
		if strings.EqualFold(provider.Type, "http") {
			if err := validateSubscriptionURL(provider.URL); err != nil {
				return fmt.Errorf("render.rule_providers[%d].url: %w", i, err)
			}
		}
		if strings.EqualFold(provider.Type, "file") && strings.TrimSpace(provider.Path) == "" {
			return fmt.Errorf("render.rule_providers[%d].path is required for file providers", i)
		}
		if strings.EqualFold(provider.Type, "inline") && len(provider.Payload) == 0 {
			return fmt.Errorf("render.rule_providers[%d].payload is required for inline providers", i)
		}
	}

	seenGroups := make(map[string]struct{}, len(cfg.Render.CustomProxyGroups))
	for i, group := range cfg.Render.CustomProxyGroups {
		if strings.TrimSpace(group.Name) == "" {
			return fmt.Errorf("render.custom_proxy_groups[%d].name is required", i)
		}
		if _, ok := seenGroups[group.Name]; ok {
			return fmt.Errorf("render.custom_proxy_groups[%d].name must be unique", i)
		}
		seenGroups[group.Name] = struct{}{}
		if !isAllowedProxyGroupType(group.Type) {
			return fmt.Errorf("render.custom_proxy_groups[%d].type must be one of select, url-test, fallback, relay", i)
		}
		if len(group.Members) == 0 {
			return fmt.Errorf("render.custom_proxy_groups[%d].members is required", i)
		}
		if (strings.EqualFold(group.Type, "url-test") || strings.EqualFold(group.Type, "fallback")) && strings.TrimSpace(group.URL) == "" {
			return fmt.Errorf("render.custom_proxy_groups[%d].url is required for url-test and fallback groups", i)
		}
	}

	return nil
}

func validateConfig(cfg model.Config) error {
	return Validate(cfg)
}

func stableConfigID(prefix, name, raw string, index int) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.ToLower(strings.TrimSpace(prefix)),
		strings.TrimSpace(name),
		strings.TrimSpace(raw),
		fmt.Sprintf("%d", index),
	}, "|")))
	return prefix + "-" + hex.EncodeToString(sum[:])[:12]
}

func uniqueDisplayName(raw string, seen map[string]int) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		name = "source"
	}
	seen[name]++
	if seen[name] == 1 {
		return name
	}
	return fmt.Sprintf("%s %d", name, seen[name])
}

func isAllowedTemplate(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "lite", "standard", "full":
		return true
	default:
		return false
	}
}

func isAllowedProviderType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http", "file", "inline":
		return true
	default:
		return false
	}
}

func isAllowedProviderBehavior(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "domain", "ipcidr", "classical":
		return true
	default:
		return false
	}
}

func isAllowedProviderFormat(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yaml", "text", "mrs":
		return true
	default:
		return false
	}
}

func isAllowedProxyGroupType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "select", "url-test", "fallback", "relay":
		return true
	default:
		return false
	}
}

func isAllowedDedupeScope(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "global", "per_source", "none":
		return true
	default:
		return false
	}
}

func validateCustomRules(rules []model.CustomRule) error {
	seen := make(map[string]struct{}, len(rules))
	for i, rule := range rules {
		if strings.TrimSpace(rule.Key) == "" {
			return fmt.Errorf("render.custom_rules[%d].key is required", i)
		}
		if !isValidCustomRuleKey(rule.Key) {
			return fmt.Errorf("render.custom_rules[%d].key must match [a-z0-9_-]+", i)
		}
		if _, ok := seen[rule.Key]; ok {
			return fmt.Errorf("render.custom_rules[%d].key must be unique", i)
		}
		seen[rule.Key] = struct{}{}
		if strings.TrimSpace(rule.Label) == "" {
			return fmt.Errorf("render.custom_rules[%d].label is required", i)
		}
		if !isAllowedCustomRuleTargetMode(rule.TargetMode) {
			return fmt.Errorf("render.custom_rules[%d].target_mode must be one of new_group, existing_group, direct, reject", i)
		}
		if !isAllowedCustomRuleSourceType(rule.SourceType) {
			return fmt.Errorf("render.custom_rules[%d].source_type must be one of inline, http, file, group_only", i)
		}
		if !isAllowedProviderBehavior(rule.Behavior) {
			return fmt.Errorf("render.custom_rules[%d].behavior must be one of domain, ipcidr, classical", i)
		}
		if !isAllowedProviderFormat(rule.Format) {
			return fmt.Errorf("render.custom_rules[%d].format must be one of yaml, text, mrs", i)
		}
		if strings.EqualFold(rule.Format, "mrs") && strings.EqualFold(rule.Behavior, "classical") {
			return fmt.Errorf("render.custom_rules[%d]: format mrs cannot be used with classical behavior", i)
		}
		switch strings.ToLower(strings.TrimSpace(rule.SourceType)) {
		case "http":
			if err := validateSubscriptionURL(rule.URL); err != nil {
				return fmt.Errorf("render.custom_rules[%d].url: %w", i, err)
			}
		case "file":
			if strings.TrimSpace(rule.Path) == "" {
				return fmt.Errorf("render.custom_rules[%d].path is required", i)
			}
		case "inline":
			if len(rule.Payload) == 0 {
				return fmt.Errorf("render.custom_rules[%d].payload is required", i)
			}
		}
		if strings.EqualFold(rule.TargetMode, "existing_group") && strings.TrimSpace(rule.TargetGroup) == "" {
			return fmt.Errorf("render.custom_rules[%d].target_group is required for existing_group", i)
		}
		if !isAllowedCustomRuleInsertPosition(rule.InsertPosition) {
			return fmt.Errorf("render.custom_rules[%d].insert_position must be one of after_adblock, before_domestic, before_non_cn, before_match", i)
		}
	}
	return nil
}

func defaultCustomRuleTargetGroup(rule model.CustomRule) string {
	switch strings.ToLower(strings.TrimSpace(rule.TargetMode)) {
	case "direct":
		return "DIRECT"
	case "reject":
		return "REJECT"
	case "existing_group":
		return strings.TrimSpace(rule.TargetGroup)
	default:
		if strings.TrimSpace(rule.Icon) != "" {
			return strings.TrimSpace(rule.Icon + " " + rule.Label)
		}
		return strings.TrimSpace(rule.Label)
	}
}

func isAllowedCustomRuleTargetMode(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "new_group", "existing_group", "direct", "reject":
		return true
	default:
		return false
	}
}

func isAllowedCustomRuleSourceType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "inline", "http", "file", "group_only":
		return true
	default:
		return false
	}
}

func isAllowedCustomRuleInsertPosition(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "after_adblock", "before_domestic", "before_non_cn", "before_match":
		return true
	default:
		return false
	}
}

func isValidCustomRuleKey(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func validateSubscriptionURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fmt.Errorf("must not be empty")
	}
	switch {
	case strings.HasPrefix(strings.ToLower(value), "http://"), strings.HasPrefix(strings.ToLower(value), "https://"):
		return nil
	default:
		return fmt.Errorf("must use http or https scheme")
	}
}
