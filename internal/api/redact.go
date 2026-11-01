package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"subconv-next/internal/model"
)

func RedactURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return maskSensitiveText(raw)
	}
	if parsed.User != nil {
		parsed.User = url.User(maskedSecretValue)
	}
	if parsed.RawQuery != "" {
		values := parsed.Query()
		for key := range values {
			values.Set(key, "***")
		}
		parsed.RawQuery = values.Encode()
	}
	return parsed.String()
}

func RedactSecret(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return "***"
}

// GET /api/config intentionally masks subscription query credentials. Saving
// that response must retain the original URL for the same subscription, rather
// than replacing a real token with the UI's placeholder. Never transfer a token
// to a different URL or source identity.
func restoreRedactedSubscriptionURLs(next *model.Config, previous model.Config) error {
	previousByID := make(map[string]model.SubscriptionConfig, len(previous.Subscriptions))
	for _, source := range previous.Subscriptions {
		previousByID[source.ID] = source
	}
	for i := range next.Subscriptions {
		source := &next.Subscriptions[i]
		old, exists := previousByID[source.ID]
		if exists && source.URL == RedactURL(old.URL) {
			source.URL = old.URL
		}
		parsed, err := url.Parse(source.URL)
		if err == nil {
			for _, values := range parsed.Query() {
				for _, value := range values {
					if value == "***" || value == maskedSecretValue {
						return fmt.Errorf("订阅地址含脱敏凭据，无法恢复原地址，请重新填写完整订阅地址")
					}
				}
			}
		}
		if exists && source.SourceLogo == RedactURL(old.SourceLogo) {
			source.SourceLogo = old.SourceLogo
		}
	}
	return nil
}

func RedactConfig(cfg model.Config) model.Config {
	cfg.Service.AccessToken = RedactSecret(cfg.Service.AccessToken)
	cfg.Service.SubscriptionToken = RedactSecret(cfg.Service.SubscriptionToken)
	for i := range cfg.Subscriptions {
		cfg.Subscriptions[i].URL = RedactURL(cfg.Subscriptions[i].URL)
		cfg.Subscriptions[i].SourceLogo = RedactURL(cfg.Subscriptions[i].SourceLogo)
	}
	for i := range cfg.Render.CustomRules {
		cfg.Render.CustomRules[i].URL = RedactURL(cfg.Render.CustomRules[i].URL)
	}
	for i := range cfg.Render.RuleProviders {
		cfg.Render.RuleProviders[i].URL = RedactURL(cfg.Render.RuleProviders[i].URL)
		cfg.Render.RuleProviders[i].Headers = redactHeaderList(cfg.Render.RuleProviders[i].Headers)
	}
	for i := range cfg.Render.CustomProxyGroups {
		cfg.Render.CustomProxyGroups[i].URL = RedactURL(cfg.Render.CustomProxyGroups[i].URL)
	}
	cfg.Render.ExternalConfig.CustomURL = RedactURL(cfg.Render.ExternalConfig.CustomURL)
	if cfg.Render.GeoxURL != nil {
		geox := *cfg.Render.GeoxURL
		geox.GeoIP = RedactURL(geox.GeoIP)
		geox.GeoSite = RedactURL(geox.GeoSite)
		geox.MMDB = RedactURL(geox.MMDB)
		geox.ASN = RedactURL(geox.ASN)
		cfg.Render.GeoxURL = &geox
	}
	if cfg.Render.DNS != nil {
		dns := model.CloneDNSConfig(cfg.Render.DNS)
		dns.DefaultNameserver = redactURLList(dns.DefaultNameserver)
		dns.Nameserver = redactURLList(dns.Nameserver)
		dns.ProxyNameserver = redactURLList(dns.ProxyNameserver)
		dns.DirectNameserver = redactURLList(dns.DirectNameserver)
		dns.Fallback = redactURLList(dns.Fallback)
		for key, values := range dns.NameserverPolicy {
			dns.NameserverPolicy[key] = redactURLList(values)
		}
		cfg.Render.DNS = dns
	}
	return cfg
}

func redactHeaderList(headers map[string][]string) map[string][]string {
	if headers == nil {
		return nil
	}
	redacted := make(map[string][]string, len(headers))
	for key, values := range headers {
		redacted[key] = append([]string(nil), values...)
		if isSensitiveField(key) {
			for i := range redacted[key] {
				redacted[key][i] = maskedSecretValue
			}
		}
	}
	return redacted
}

func redactURLList(values []string) []string {
	redacted := append([]string(nil), values...)
	for i := range redacted {
		redacted[i] = RedactURL(redacted[i])
	}
	return redacted
}

// redactConfigForResponse removes secrets from every configuration response
// and hides server-owned filesystem details in anonymous public mode. The
// latter are restored authoritatively by applyWorkspaceConfigPolicy before a
// config is written, so the web UI can safely fall back to its defaults.
func redactConfigForResponse(cfg model.Config, public bool) model.Config {
	cfg = RedactConfig(cfg)
	if public {
		cfg.Service.OutputPath = ""
		cfg.Service.CacheDir = ""
		cfg.Service.StatePath = ""
		cfg.Service.ListenAddr = ""
		cfg.Service.ListenPort = 0
	}
	return cfg
}

func publicResponsePath(public bool, path string) string {
	if public {
		return ""
	}
	return path
}

func RedactNode(node model.NodeIR) model.NodeIR {
	node.Auth = maskAuthSecrets(node.Auth)
	node.Transport = maskTransportSecrets(node.Transport)
	node.WireGuard = maskWireGuardSecrets(node.WireGuard)
	node.Raw = maskSensitiveMap(node.Raw)
	return node
}

func RedactLogLine(line string) string {
	return maskSensitiveText(line)
}

func maskTransportSecrets(transport model.TransportOptions) model.TransportOptions {
	if len(transport.Headers) == 0 {
		return transport
	}
	headers := make(map[string]string, len(transport.Headers))
	for key, value := range transport.Headers {
		if isSensitiveField(key) {
			headers[key] = maskedSecretValue
			continue
		}
		headers[key] = value
	}
	transport.Headers = headers
	return transport
}

func maskWireGuardSecrets(wg *model.WireGuardOptions) *model.WireGuardOptions {
	if wg == nil {
		return nil
	}
	data, _ := json.Marshal(wg)
	var cloned model.WireGuardOptions
	_ = json.Unmarshal(data, &cloned)
	for i := range cloned.Peers {
		if strings.TrimSpace(cloned.Peers[i].PreSharedKey) != "" {
			cloned.Peers[i].PreSharedKey = maskedSecretValue
		}
	}
	return &cloned
}
