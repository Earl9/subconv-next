package parser

import (
	"strings"

	"subconv-next/internal/model"
)

func parseWireGuardURI(raw string, source model.SourceInfo) (model.NodeIR, error) {
	u, err := parseStandardURL(raw)
	if err != nil {
		return model.NodeIR{}, err
	}

	host, port, err := hostPortFromURL(u)
	if err != nil {
		return model.NodeIR{}, err
	}

	node := newBaseNode(model.ProtocolWireGuard, source)
	node.Name = parseFragmentName(u)
	node.Server = host
	node.Port = port
	node.UDP = model.Bool(true)
	node.WireGuard = &model.WireGuardOptions{}

	if u.User != nil {
		node.Auth.PrivateKey = u.User.Username()
	}

	q := u.Query()
	node.Auth.PublicKey = firstQuery(q, "public-key", "peer-public-key")
	node.Auth.PreSharedKey = firstQuery(q, "pre-shared-key", "preshared-key", "psk")
	node.WireGuard.IP = firstQuery(q, "ip", "address")
	node.WireGuard.IPv6 = firstQuery(q, "ipv6")
	node.WireGuard.AllowedIPs = parseCSV(firstQuery(q, "allowed-ips", "allowedIPs"))
	if len(node.WireGuard.AllowedIPs) == 0 {
		node.WireGuard.AllowedIPs = []string{"0.0.0.0/0"}
	}

	reserved := parseCSV(firstQuery(q, "reserved"))
	if len(reserved) > 0 {
		var ints []int
		ok := true
		for _, value := range reserved {
			n, parsed := parseIntString(value)
			if !parsed {
				ok = false
				break
			}
			ints = append(ints, n)
		}
		if ok {
			node.WireGuard.Reserved = ints
		} else {
			node.WireGuard.ReservedString = strings.Join(reserved, ",")
		}
	}

	if mtu, ok := parseIntString(firstQuery(q, "mtu")); ok {
		node.WireGuard.MTU = mtu
	}
	if keepalive, ok := parseIntString(firstQuery(q, "persistent-keepalive")); ok {
		node.WireGuard.PersistentKeepalive = keepalive
	}
	node.WireGuard.RemoteDNSResolve = parseBoolString(firstQuery(q, "remote-dns-resolve"))
	node.WireGuard.DNS = parseCSV(firstQuery(q, "dns"))

	if raw := unknownQueryParams(q, "public-key", "peer-public-key", "pre-shared-key", "preshared-key", "psk", "ip", "address", "ipv6", "allowed-ips", "allowedIPs", "reserved", "mtu", "persistent-keepalive", "remote-dns-resolve", "dns"); raw != nil {
		for key, value := range raw {
			setRaw(&node, key, value)
		}
	}

	return node, nil
}
