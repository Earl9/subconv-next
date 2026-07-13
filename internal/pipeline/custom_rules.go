package pipeline

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
)

const maxRemoteCustomRuleBytes = 2 * 1024 * 1024

type customRuleFetcher interface {
	Fetch(context.Context, fetcher.Source) (fetcher.FetchedSubscription, []string, error)
}

func snapshotRemoteCustomRules(cfg model.Config) (model.Config, []string, error) {
	opts := fetcher.OptionsFromConfig(cfg)
	opts.CacheDir = filepath.Join(cfg.Service.CacheDir, "custom-rules")
	if opts.MaxBodyBytes <= 0 || opts.MaxBodyBytes > maxRemoteCustomRuleBytes {
		opts.MaxBodyBytes = maxRemoteCustomRuleBytes
	}
	return snapshotRemoteCustomRulesWithFetcher(cfg, fetcher.New(opts))
}

func snapshotRemoteCustomRulesWithFetcher(cfg model.Config, remote customRuleFetcher) (model.Config, []string, error) {
	resolved := cfg
	resolved.Render.CustomRules = append([]model.CustomRule(nil), cfg.Render.CustomRules...)

	var warnings []string
	for i := range resolved.Render.CustomRules {
		rule := &resolved.Render.CustomRules[i]
		if !rule.Enabled || !strings.EqualFold(strings.TrimSpace(rule.SourceType), "http") {
			continue
		}

		format := remoteCustomRuleFormat(*rule)
		if format == "mrs" {
			warnings = append(warnings, fmt.Sprintf("custom rule %s uses mrs and remains a runtime rule provider", rule.Key))
			continue
		}

		fetched, fetchWarnings, err := remote.Fetch(context.Background(), fetcher.Source{
			Name:      "custom rule " + strings.TrimSpace(rule.Key),
			URL:       strings.TrimSpace(rule.URL),
			UserAgent: model.DefaultUserAgent,
			Enabled:   true,
			CacheTTL:  time.Duration(rule.Interval) * time.Second,
		})
		warnings = append(warnings, fetchWarnings...)
		if err != nil {
			return cfg, warnings, fmt.Errorf("fetch remote custom rule %s: %w", rule.Key, err)
		}

		payload, err := parseRemoteCustomRulePayload(fetched.Content, format)
		if err != nil {
			return cfg, warnings, fmt.Errorf("parse remote custom rule %s: %w", rule.Key, err)
		}

		rule.SourceType = "inline"
		rule.Format = "text"
		rule.Payload = payload
		rule.URL = ""
		rule.Path = ""
		source := "remote snapshot"
		if fetched.FromCache {
			source = "cached snapshot"
		}
		warnings = append(warnings, fmt.Sprintf("custom rule %s embedded %d entries from %s", rule.Key, len(payload), source))
	}

	return resolved, warnings, nil
}

func remoteCustomRuleFormat(rule model.CustomRule) string {
	if format := strings.ToLower(strings.TrimSpace(rule.Format)); format != "" {
		return format
	}
	source := strings.ToLower(strings.TrimSpace(rule.Path + " " + rule.URL))
	switch {
	case strings.Contains(source, ".mrs"):
		return "mrs"
	case strings.Contains(source, ".yaml"), strings.Contains(source, ".yml"):
		return "yaml"
	default:
		return "text"
	}
}

func parseRemoteCustomRulePayload(content []byte, format string) ([]string, error) {
	var payload []string
	if strings.EqualFold(strings.TrimSpace(format), "yaml") {
		var document struct {
			Payload []string `yaml:"payload"`
		}
		if err := yaml.Unmarshal(content, &document); err != nil {
			return nil, fmt.Errorf("decode yaml: %w", err)
		}
		payload = document.Payload
	} else {
		payload = strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	}

	normalized := make([]string, 0, len(payload))
	seen := make(map[string]struct{}, len(payload))
	for _, entry := range payload {
		entry = normalizeRemoteCustomRuleLine(entry)
		if entry == "" {
			continue
		}
		if _, exists := seen[entry]; exists {
			continue
		}
		seen[entry] = struct{}{}
		normalized = append(normalized, entry)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("payload is empty")
	}
	return normalized, nil
}

func normalizeRemoteCustomRuleLine(value string) string {
	line := strings.TrimSpace(value)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
	line = strings.Trim(line, `"'`)
	if index := strings.Index(line, " #"); index >= 0 {
		line = strings.TrimSpace(line[:index])
	}
	return line
}
