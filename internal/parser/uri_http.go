package parser

import (
	"strings"

	"subconv-next/internal/model"
)

func parseHTTPProxy(raw string, source model.SourceInfo) (model.NodeIR, error) {
	return parseHTTPLikeProxy(raw, source, model.ProtocolHTTP)
}

func parseHTTPSProxy(raw string, source model.SourceInfo) (model.NodeIR, error) {
	node, err := parseHTTPLikeProxy(raw, source, model.ProtocolHTTP)
	if err != nil {
		return model.NodeIR{}, err
	}
	node.TLS.Enabled = true
	return node, nil
}

func parseSOCKS5Proxy(raw string, source model.SourceInfo) (model.NodeIR, error) {
	return parseHTTPLikeProxy(raw, source, model.ProtocolSOCKS5)
}

func parseHTTPLikeProxy(raw string, source model.SourceInfo, protocol model.Protocol) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(protocol, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port

	q := u.Query()
	if u.User != nil {
		node.Auth.Username = strings.TrimSpace(u.User.Username())
		if password, ok := u.User.Password(); ok {
			node.Auth.Password = strings.TrimSpace(password)
		}
	}
	if node.Auth.Username == "" {
		node.Auth.Username = firstQuery(q, "username", "user")
	}
	if node.Auth.Password == "" {
		node.Auth.Password = firstQuery(q, "password", "passwd", "pwd")
	}

	if hasQuery(q, "tls", "security") {
		tlsValue := firstQuery(q, "tls", "security")
		node.TLS.Enabled = parseBoolString(tlsValue) || strings.EqualFold(strings.TrimSpace(tlsValue), "tls")
	}
	node.TLS.SNI = firstQuery(q, "sni", "servername")
	node.TLS.ALPN = parseCSV(firstQuery(q, "alpn"))
	node.TLS.Insecure = parseBoolString(firstQuery(q, "allowInsecure", "allow-insecure", "insecure", "skip-cert-verify"))
	if hasQuery(q, "udp") {
		node.UDP = model.Bool(parseBoolString(firstQuery(q, "udp")))
	}
	if node.TLS.Insecure {
		setRaw(&node, "skipCertVerify", true)
	}

	if raw := unknownQueryParams(q, "username", "user", "password", "passwd", "pwd", "tls", "security", "sni", "servername", "alpn", "allowInsecure", "allow-insecure", "insecure", "skip-cert-verify", "udp"); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}
