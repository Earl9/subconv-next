package api

import (
	"encoding/json"
	"net/url"
	"strings"

	"subconv-next/internal/model"
)

func RedactURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return maskSensitiveText(raw)
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

func RedactConfig(cfg model.Config) model.Config {
	cfg.Service.AccessToken = RedactSecret(cfg.Service.AccessToken)
	cfg.Service.SubscriptionToken = RedactSecret(cfg.Service.SubscriptionToken)
	for i := range cfg.Subscriptions {
		cfg.Subscriptions[i].URL = RedactURL(cfg.Subscriptions[i].URL)
	}
	return cfg
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
