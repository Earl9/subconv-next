package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"subconv-next/internal/model"
)

func ParseWireGuardConfig(content []byte, source model.SourceInfo) (model.NodeIR, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	currentSection := ""
	interfaceValues := make(map[string][]string)
	var peerValues []map[string][]string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			if currentSection == "peer" {
				peerValues = append(peerValues, make(map[string][]string))
			}
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return model.NodeIR{}, fmt.Errorf("invalid wireguard line %q", line)
		}

		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		switch currentSection {
		case "interface":
			interfaceValues[key] = append(interfaceValues[key], value)
		case "peer":
			if len(peerValues) == 0 {
				peerValues = append(peerValues, make(map[string][]string))
			}
			peerValues[len(peerValues)-1][key] = append(peerValues[len(peerValues)-1][key], value)
		}
	}

	if err := scanner.Err(); err != nil {
		return model.NodeIR{}, fmt.Errorf("scan wireguard config: %w", err)
	}
	if len(interfaceValues) == 0 || len(peerValues) == 0 {
		return model.NodeIR{}, fmt.Errorf("missing wireguard interface or peer section")
	}

	node := newBaseNode(model.ProtocolWireGuard, source)
	node.UDP = model.Bool(true)
	node.WireGuard = &model.WireGuardOptions{}
	node.Auth.PrivateKey = firstSliceValue(interfaceValues["privatekey"])

	for _, address := range parseCSV(firstSliceValue(interfaceValues["address"])) {
		if strings.Contains(address, ":") {
			if node.WireGuard.IPv6 == "" {
				node.WireGuard.IPv6 = address
			}
			continue
		}
		if node.WireGuard.IP == "" {
			node.WireGuard.IP = address
		}
	}
	node.WireGuard.DNS = parseCSV(firstSliceValue(interfaceValues["dns"]))
	if mtu, ok := parseIntString(firstSliceValue(interfaceValues["mtu"])); ok {
		node.WireGuard.MTU = mtu
	}

	for _, peerValue := range peerValues {
		peer := model.WGPeer{
			PublicKey:    firstSliceValue(peerValue["publickey"]),
			PreSharedKey: firstSliceValue(peerValue["presharedkey"]),
			AllowedIPs:   parseCSV(firstSliceValue(peerValue["allowedips"])),
		}
		if len(peer.AllowedIPs) == 0 {
			peer.AllowedIPs = []string{"0.0.0.0/0"}
		}

		endpoint := firstSliceValue(peerValue["endpoint"])
		if endpoint != "" {
			host, port, err := splitEndpoint(endpoint)
			if err != nil {
				return model.NodeIR{}, err
			}
			peer.Server = host
			peer.Port = port
		}

		if keepalive, ok := parseIntString(firstSliceValue(peerValue["persistentkeepalive"])); ok {
			node.WireGuard.PersistentKeepalive = keepalive
		}

		node.WireGuard.Peers = append(node.WireGuard.Peers, peer)
	}

	firstPeer := node.WireGuard.Peers[0]
	node.Server = firstPeer.Server
	node.Port = firstPeer.Port
	node.Auth.PublicKey = firstPeer.PublicKey
	node.Auth.PreSharedKey = firstPeer.PreSharedKey
	node.WireGuard.AllowedIPs = append([]string(nil), firstPeer.AllowedIPs...)

	if len(node.WireGuard.Peers) == 1 {
		node.WireGuard.Peers = nil
	}

	return node, nil
}

func firstSliceValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
