package parser

import (
	"strings"

	"subconv-next/internal/model"
)

func parseAnyTLS(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolAnyTLS, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	node.TLS.Enabled = true
	node.UDP = model.Bool(true)
	if u.User != nil {
		if password := u.User.Username(); password != "" {
			node.Auth.Password = password
		} else if fallback, ok := u.User.Password(); ok {
			node.Auth.Password = fallback
		}
	}

	q := u.Query()
	node.TLS.SNI = firstQuery(q, "sni", "servername")
	node.TLS.ALPN = parseCSV(firstQuery(q, "alpn"))
	node.TLS.Insecure = parseBoolString(firstQuery(q, "insecure", "skip-cert-verify", "allowInsecure"))
	node.TLS.ClientFingerprint = firstQuery(q, "client-fingerprint", "fp")

	if echConfig := firstQuery(q, "ech", "ech-config"); echConfig != "" {
		node.TLS.ECH = &model.ECHOptions{
			Enabled: true,
			Config:  echConfig,
		}
	}

	setRaw(&node, "idleSessionCheckInterval", firstQuery(q, "idle-session-check-interval"))
	setRaw(&node, "idleSessionTimeout", firstQuery(q, "idle-session-timeout"))
	setRaw(&node, "minIdleSession", firstQuery(q, "min-idle-session"))

	if security := strings.ToLower(firstQuery(q, "security")); security == "reality" ||
		firstQuery(q, "pbk", "publicKey", "public-key") != "" ||
		firstQuery(q, "sid", "shortId", "short-id") != "" {
		node.Warnings = append(node.Warnings, "reality parameters ignored for anytls")
	}

	if raw := unknownQueryParams(q, "sni", "servername", "alpn", "insecure", "skip-cert-verify", "allowInsecure", "client-fingerprint", "fp", "idle-session-check-interval", "idle-session-timeout", "min-idle-session", "ech", "ech-config", "security", "pbk", "publicKey", "public-key", "sid", "shortId", "short-id"); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}
