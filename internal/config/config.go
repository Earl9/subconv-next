package config

import (
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
	if cfg.Subscriptions == nil {
		cfg.Subscriptions = []model.SubscriptionConfig{}
	}
	if cfg.Inline == nil {
		cfg.Inline = []model.InlineConfig{}
	}
	if cfg.Render.AdditionalRules == nil {
		cfg.Render.AdditionalRules = []string{}
	}
	if cfg.Render.RuleProviders == nil {
		cfg.Render.RuleProviders = []model.RuleProviderConfig{}
	}
	if cfg.Render.CustomProxyGroups == nil {
		cfg.Render.CustomProxyGroups = []model.CustomProxyGroupConfig{}
	}
	return cfg
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
		if strings.TrimSpace(inline.Name) == "" {
			return fmt.Errorf("inline[%d].name is required", i)
		}
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
