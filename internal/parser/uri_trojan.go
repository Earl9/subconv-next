package parser

import "subconv-next/internal/model"

func parseTrojan(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolTrojan, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	node.TLS.Enabled = true
	node.UDP = model.Bool(true)
	if u.User != nil {
		node.Auth.Password = u.User.Username()
	}

	q := u.Query()
	node.TLS.SNI = firstQuery(q, "sni", "servername")
	node.TLS.ALPN = parseCSV(firstQuery(q, "alpn"))
	node.TLS.Insecure = parseBoolString(firstQuery(q, "allowInsecure", "allow-insecure", "insecure", "skip-cert-verify"))
	node.Transport.Network = firstQuery(q, "type", "network")
	node.Transport.Host = firstQuery(q, "host")
	node.Transport.Path = firstQuery(q, "path")
	if raw := unknownQueryParams(q, "sni", "servername", "alpn", "allowInsecure", "allow-insecure", "insecure", "skip-cert-verify", "type", "network", "host", "path"); raw != nil {
		node.Raw = raw
	}

	return node, nil
}
