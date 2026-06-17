package parser

import (
	"fmt"
	"strings"

	"subconv-next/internal/model"
)

func parseMieru(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolMieru, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	node.UDP = model.Bool(true)

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
	if node.Port == 0 {
		if parsedPort, ok := parseIntString(firstQuery(q, "port")); ok {
			node.Port = parsedPort
		}
	}

	transport := normalizeMieruTransport(firstQuery(q, "transport"))
	if transport == "" {
		transport = "TCP"
	}
	setRaw(&node, "transport", transport)
	setRaw(&node, "portRange", firstQuery(q, "port-range", "port_range"))
	setRaw(&node, "multiplexing", firstQuery(q, "multiplexing"))
	setRaw(&node, "handshakeMode", firstQuery(q, "handshake-mode", "handshake_mode"))
	setRaw(&node, "trafficPattern", firstQuery(q, "traffic-pattern", "traffic_pattern"))
	if hasQuery(q, "udp") {
		node.UDP = model.Bool(parseBoolString(firstQuery(q, "udp")))
	}

	if node.Port == 0 && rawStringForParser(node.Raw, "portRange") == "" {
		return model.NodeIR{}, fmt.Errorf("missing port or port-range")
	}

	if raw := unknownQueryParams(q, "username", "user", "password", "passwd", "pwd", "port", "port-range", "port_range", "transport", "udp", "multiplexing", "handshake-mode", "handshake_mode", "traffic-pattern", "traffic_pattern"); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}

func normalizeMieruTransport(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "tcp":
		return "TCP"
	case "udp":
		return "UDP"
	default:
		return strings.TrimSpace(value)
	}
}

func rawStringForParser(raw map[string]interface{}, key string) string {
	if len(raw) == 0 {
		return ""
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}
