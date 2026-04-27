package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"subconv-next/internal/model"
)

type uciSection struct {
	Type    string
	Name    string
	Options map[string][]string
}

func LoadUCIBytes(data []byte) (model.Config, error) {
	sections, err := parseUCISections(data)
	if err != nil {
		return model.Config{}, err
	}

	cfg := model.DefaultConfig()
	for _, section := range sections {
		switch section.Type {
		case "service":
			if err := applyServiceSection(&cfg.Service, section); err != nil {
				return model.Config{}, err
			}
		case "subscription":
			sub := model.DefaultSubscriptionConfig()
			if err := applySubscriptionSection(&sub, section); err != nil {
				return model.Config{}, err
			}
			cfg.Subscriptions = append(cfg.Subscriptions, sub)
		case "inline":
			inline := model.DefaultInlineConfig()
			if err := applyInlineSection(&inline, section); err != nil {
				return model.Config{}, err
			}
			cfg.Inline = append(cfg.Inline, inline)
		case "render":
			if err := applyRenderSection(&cfg.Render, section); err != nil {
				return model.Config{}, err
			}
		}
	}

	cfg = normalizeConfig(cfg)

	if err := validateConfig(cfg); err != nil {
		return model.Config{}, err
	}

	return cfg, nil
}

func parseUCISections(data []byte) ([]uciSection, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var sections []uciSection
	var current *uciSection
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		keyword, rest := splitWord(line)
		switch keyword {
		case "config":
			section, err := parseConfigLine(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			sections = append(sections, section)
			current = &sections[len(sections)-1]
		case "option":
			if current == nil {
				return nil, fmt.Errorf("line %d: option without config section", lineNo)
			}
			key, value, err := parseOptionLine(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			current.Options[key] = []string{value}
		case "list":
			if current == nil {
				return nil, fmt.Errorf("line %d: list without config section", lineNo)
			}
			key, value, err := parseOptionLine(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			current.Options[key] = append(current.Options[key], value)
		default:
			return nil, fmt.Errorf("line %d: unsupported directive %q", lineNo, keyword)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan UCI config: %w", err)
	}

	return sections, nil
}

func parseConfigLine(input string) (uciSection, error) {
	sectionType, rest, err := parseValue(input)
	if err != nil {
		return uciSection{}, fmt.Errorf("parse config type: %w", err)
	}

	name := ""
	if strings.TrimSpace(rest) != "" {
		name, rest, err = parseValue(rest)
		if err != nil {
			return uciSection{}, fmt.Errorf("parse config name: %w", err)
		}
	}
	if strings.TrimSpace(rest) != "" {
		return uciSection{}, fmt.Errorf("unexpected trailing content in config line")
	}

	return uciSection{
		Type:    sectionType,
		Name:    name,
		Options: make(map[string][]string),
	}, nil
}

func parseOptionLine(input string) (string, string, error) {
	key, rest, err := parseValue(input)
	if err != nil {
		return "", "", fmt.Errorf("parse option key: %w", err)
	}

	value, rest, err := parseValue(rest)
	if err != nil {
		return "", "", fmt.Errorf("parse option value: %w", err)
	}
	if strings.TrimSpace(rest) != "" {
		return "", "", fmt.Errorf("unexpected trailing content in option line")
	}

	return key, value, nil
}

func parseValue(input string) (string, string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", "", fmt.Errorf("missing value")
	}

	switch trimmed[0] {
	case '\'', '"':
		quote := trimmed[0]
		for i := 1; i < len(trimmed); i++ {
			if trimmed[i] == quote {
				return trimmed[1:i], strings.TrimSpace(trimmed[i+1:]), nil
			}
		}
		return "", "", fmt.Errorf("unterminated quoted value")
	default:
		end := len(trimmed)
		for i, r := range trimmed {
			if unicode.IsSpace(r) {
				end = i
				break
			}
		}
		return trimmed[:end], strings.TrimSpace(trimmed[end:]), nil
	}
}

func splitWord(input string) (string, string) {
	input = strings.TrimSpace(input)
	for i, r := range input {
		if unicode.IsSpace(r) {
			return input[:i], strings.TrimSpace(input[i+1:])
		}
	}
	return input, ""
}

func applyServiceSection(dst *model.ServiceConfig, section uciSection) error {
	if err := setBool(section, "enabled", &dst.Enabled); err != nil {
		return fmt.Errorf("service.enabled: %w", err)
	}
	setString(section, "listen_addr", &dst.ListenAddr)
	if err := setInt(section, "listen_port", &dst.ListenPort); err != nil {
		return fmt.Errorf("service.listen_port: %w", err)
	}
	setString(section, "log_level", &dst.LogLevel)
	setString(section, "template", &dst.Template)
	setString(section, "output_path", &dst.OutputPath)
	setString(section, "cache_dir", &dst.CacheDir)
	setString(section, "state_path", &dst.StatePath)
	if err := setInt(section, "refresh_interval", &dst.RefreshInterval); err != nil {
		return fmt.Errorf("service.refresh_interval: %w", err)
	}
	if err := setInt(section, "max_subscription_bytes", &dst.MaxSubscriptionBytes); err != nil {
		return fmt.Errorf("service.max_subscription_bytes: %w", err)
	}
	if err := setInt(section, "fetch_timeout_seconds", &dst.FetchTimeoutSeconds); err != nil {
		return fmt.Errorf("service.fetch_timeout_seconds: %w", err)
	}
	if err := setBool(section, "allow_lan", &dst.AllowLAN); err != nil {
		return fmt.Errorf("service.allow_lan: %w", err)
	}
	return nil
}

func applySubscriptionSection(dst *model.SubscriptionConfig, section uciSection) error {
	setString(section, "name", &dst.Name)
	if err := setBool(section, "enabled", &dst.Enabled); err != nil {
		return fmt.Errorf("subscription.enabled: %w", err)
	}
	setString(section, "url", &dst.URL)
	setString(section, "user_agent", &dst.UserAgent)
	if err := setBool(section, "insecure_skip_verify", &dst.InsecureSkipVerify); err != nil {
		return fmt.Errorf("subscription.insecure_skip_verify: %w", err)
	}
	return nil
}

func applyInlineSection(dst *model.InlineConfig, section uciSection) error {
	setString(section, "name", &dst.Name)
	if err := setBool(section, "enabled", &dst.Enabled); err != nil {
		return fmt.Errorf("inline.enabled: %w", err)
	}
	setString(section, "content", &dst.Content)
	return nil
}

func applyRenderSection(dst *model.RenderConfig, section uciSection) error {
	if err := setInt(section, "mixed_port", &dst.MixedPort); err != nil {
		return fmt.Errorf("render.mixed_port: %w", err)
	}
	if err := setBool(section, "allow_lan", &dst.AllowLAN); err != nil {
		return fmt.Errorf("render.allow_lan: %w", err)
	}
	setString(section, "mode", &dst.Mode)
	setString(section, "log_level", &dst.LogLevel)
	if err := setBool(section, "ipv6", &dst.IPv6); err != nil {
		return fmt.Errorf("render.ipv6: %w", err)
	}
	if err := setBool(section, "dns_enabled", &dst.DNSEnabled); err != nil {
		return fmt.Errorf("render.dns_enabled: %w", err)
	}
	setString(section, "enhanced_mode", &dst.EnhancedMode)
	return nil
}

func setString(section uciSection, key string, dst *string) {
	if value, ok := sectionValue(section, key); ok {
		*dst = value
	}
}

func setInt(section uciSection, key string, dst *int) error {
	value, ok := sectionValue(section, key)
	if !ok {
		return nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid integer %q", value)
	}
	*dst = parsed
	return nil
}

func setBool(section uciSection, key string, dst *bool) error {
	value, ok := sectionValue(section, key)
	if !ok {
		return nil
	}

	parsed, err := parseBool(value)
	if err != nil {
		return err
	}
	*dst = parsed
	return nil
}

func sectionValue(section uciSection, key string) (string, bool) {
	values, ok := section.Options[key]
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[len(values)-1], true
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q", value)
	}
}
