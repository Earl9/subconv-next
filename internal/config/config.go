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
