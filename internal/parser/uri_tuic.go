package parser

import "subconv-next/internal/model"

func parseTUIC(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolTUIC, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	node.TLS.Enabled = true
	node.UDP = model.Bool(true)
	if u.User != nil {
		node.Auth.UUID = u.User.Username()
		if password, ok := u.User.Password(); ok {
			node.Auth.Password = password
		}
	}

	q := u.Query()
	node.TLS.SNI = firstQuery(q, "sni", "servername")
	node.TLS.ALPN = parseCSV(firstQuery(q, "alpn"))
	node.TLS.Insecure = parseBoolString(firstQuery(q, "allow_insecure", "allow-insecure", "insecure"))
	if node.TLS.Insecure {
		setRaw(&node, "skipCertVerify", true)
	}
	setRaw(&node, "congestionController", firstQuery(q, "congestion_control", "congestion-controller"))
	setRaw(&node, "udpRelayMode", firstQuery(q, "udp_relay_mode", "udp-relay-mode"))
	if reduceRTT := firstQuery(q, "reduce_rtt", "reduce-rtt"); reduceRTT != "" {
		setRaw(&node, "reduceRTT", parseBoolString(reduceRTT))
	}
	if raw := unknownQueryParams(q, "sni", "servername", "alpn", "allow_insecure", "allow-insecure", "insecure", "congestion_control", "congestion-controller", "udp_relay_mode", "udp-relay-mode", "reduce_rtt", "reduce-rtt"); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}
