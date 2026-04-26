package parser

import "subconv-next/internal/model"

func parseHysteria2(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolHysteria2, source)
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
	node.TLS.SNI = firstQuery(q, "sni")
	node.TLS.Insecure = parseBoolString(firstQuery(q, "insecure", "skip-cert-verify"))
	node.TLS.ALPN = parseCSV(firstQuery(q, "alpn"))
	setRaw(&node, "obfs", firstQuery(q, "obfs"))
	setRaw(&node, "obfsPassword", firstQuery(q, "obfs-password", "obfsPassword"))
	setRaw(&node, "ports", firstQuery(q, "ports"))
	setRaw(&node, "hopInterval", firstQuery(q, "hop-interval"))
	setRaw(&node, "up", firstQuery(q, "up"))
	setRaw(&node, "down", firstQuery(q, "down"))
	if raw := unknownQueryParams(q, "sni", "insecure", "skip-cert-verify", "alpn", "obfs", "obfs-password", "obfsPassword", "ports", "hop-interval", "up", "down"); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}
