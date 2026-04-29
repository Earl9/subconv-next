package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"subconv-next/internal/model"
)

type mihomoSubscription struct {
	Proxies []mihomoYAMLProxy `yaml:"proxies"`
}

type mihomoYAMLProxy struct {
	Name                     string                 `yaml:"name"`
	Type                     string                 `yaml:"type"`
	Server                   string                 `yaml:"server,omitempty"`
	Port                     int                    `yaml:"port,omitempty"`
	Encryption               string                 `yaml:"encryption,omitempty"`
	Cipher                   string                 `yaml:"cipher,omitempty"`
	Password                 string                 `yaml:"password,omitempty"`
	Protocol                 string                 `yaml:"protocol,omitempty"`
	ProtocolParam            string                 `yaml:"protocol-param,omitempty"`
	Username                 string                 `yaml:"username,omitempty"`
	UUID                     string                 `yaml:"uuid,omitempty"`
	AlterID                  interface{}            `yaml:"alterId,omitempty"`
	Network                  string                 `yaml:"network,omitempty"`
	TLS                      bool                   `yaml:"tls,omitempty"`
	UDP                      *bool                  `yaml:"udp,omitempty"`
	ServerName               string                 `yaml:"servername,omitempty"`
	SNI                      string                 `yaml:"sni,omitempty"`
	ClientFingerprint        string                 `yaml:"client-fingerprint,omitempty"`
	PacketEncoding           string                 `yaml:"packet-encoding,omitempty"`
	SkipCertVerify           *bool                  `yaml:"skip-cert-verify,omitempty"`
	ALPN                     []string               `yaml:"alpn,omitempty"`
	Flow                     string                 `yaml:"flow,omitempty"`
	RealityOpts              *mihomoYAMLRealityOpts `yaml:"reality-opts,omitempty"`
	WSOpts                   *mihomoYAMLWSOpts      `yaml:"ws-opts,omitempty"`
	GrpcOpts                 *mihomoYAMLGRPCOpts    `yaml:"grpc-opts,omitempty"`
	H2Opts                   *mihomoYAMLH2Opts      `yaml:"h2-opts,omitempty"`
	XHTTPOpts                *mihomoYAMLXHTTPOpts   `yaml:"xhttp-opts,omitempty"`
	Obfs                     string                 `yaml:"obfs,omitempty"`
	ObfsParam                string                 `yaml:"obfs-param,omitempty"`
	ObfsPassword             string                 `yaml:"obfs-password,omitempty"`
	CongestionController     string                 `yaml:"congestion-controller,omitempty"`
	UDPRelayMode             string                 `yaml:"udp-relay-mode,omitempty"`
	ReduceRTT                *bool                  `yaml:"reduce-rtt,omitempty"`
	IdleSessionCheckInterval interface{}            `yaml:"idle-session-check-interval,omitempty"`
	IdleSessionTimeout       interface{}            `yaml:"idle-session-timeout,omitempty"`
	MinIdleSession           interface{}            `yaml:"min-idle-session,omitempty"`
	IP                       string                 `yaml:"ip,omitempty"`
	IPv6                     string                 `yaml:"ipv6,omitempty"`
	PrivateKey               string                 `yaml:"private-key,omitempty"`
	PublicKey                string                 `yaml:"public-key,omitempty"`
	AllowedIPs               []string               `yaml:"allowed-ips,omitempty"`
	PreSharedKey             string                 `yaml:"pre-shared-key,omitempty"`
	Reserved                 interface{}            `yaml:"reserved,omitempty"`
	PersistentKeepalive      int                    `yaml:"persistent-keepalive,omitempty"`
	MTU                      int                    `yaml:"mtu,omitempty"`
	RemoteDNSResolve         *bool                  `yaml:"remote-dns-resolve,omitempty"`
	DNS                      []string               `yaml:"dns,omitempty"`
	Peers                    []mihomoYAMLWGPeer     `yaml:"peers,omitempty"`
	AmneziaWGOption          map[string]interface{} `yaml:"amnezia-wg-option,omitempty"`
}

type mihomoYAMLRealityOpts struct {
	PublicKey string `yaml:"public-key,omitempty"`
	ShortID   string `yaml:"short-id,omitempty"`
	SpiderX   string `yaml:"spider-x,omitempty"`
}

type mihomoYAMLWSOpts struct {
	Path    string            `yaml:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type mihomoYAMLGRPCOpts struct {
	GrpcServiceName string `yaml:"grpc-service-name,omitempty"`
}

type mihomoYAMLH2Opts struct {
	Host []string `yaml:"host,omitempty"`
	Path string   `yaml:"path,omitempty"`
}

type mihomoYAMLXHTTPOpts struct {
	Path         string `yaml:"path,omitempty"`
	Mode         string `yaml:"mode,omitempty"`
	NoGRPCHeader *bool  `yaml:"no-grpc-header,omitempty"`
}

type mihomoYAMLWGPeer struct {
	Server       string      `yaml:"server,omitempty"`
	Port         int         `yaml:"port,omitempty"`
	PublicKey    string      `yaml:"public-key,omitempty"`
	PreSharedKey string      `yaml:"pre-shared-key,omitempty"`
	AllowedIPs   []string    `yaml:"allowed-ips,omitempty"`
	Reserved     interface{} `yaml:"reserved,omitempty"`
}

func parseMihomoYAML(content []byte, source model.SourceInfo) ParseResult {
	var sub mihomoSubscription
	if err := yaml.Unmarshal(content, &sub); err != nil {
		return ParseResult{
			Errors: []ParseError{{Kind: "INVALID_YAML", Message: err.Error()}},
		}
	}
	if len(sub.Proxies) == 0 {
		return ParseResult{
			Warnings: []string{"YAML subscription contains zero proxies"},
			Errors:   []ParseError{{Kind: "EMPTY_YAML_PROXIES", Message: "YAML subscription contains zero proxies"}},
		}
	}

	result := ParseResult{}
	for index, proxy := range sub.Proxies {
		node, err := parseMihomoProxy(proxy, source)
		if err != nil {
			result.Errors = append(result.Errors, ParseError{
				Line:    index + 1,
				Kind:    "INVALID_YAML_PROXY",
				Message: err.Error(),
			})
			continue
		}
		result.Nodes = append(result.Nodes, node)
	}
	result.Nodes = model.NormalizeNodesNoDedupe(result.Nodes)
	return result
}

func parseMihomoProxy(proxy mihomoYAMLProxy, source model.SourceInfo) (model.NodeIR, error) {
	node := newBaseNode(model.Protocol(strings.ToLower(strings.TrimSpace(proxy.Type))), source)
	node.Name = strings.TrimSpace(proxy.Name)
	node.Server = strings.TrimSpace(proxy.Server)
	node.Port = proxy.Port
	if proxy.UDP != nil {
		v := *proxy.UDP
		node.UDP = &v
	}

	switch node.Type {
	case model.ProtocolSS:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		setRaw(&node, "method", strings.TrimSpace(proxy.Cipher))
	case model.ProtocolSSR:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		setRaw(&node, "method", strings.TrimSpace(proxy.Cipher))
		setRaw(&node, "protocol", strings.TrimSpace(proxy.Protocol))
		setRaw(&node, "protocolParam", strings.TrimSpace(proxy.ProtocolParam))
		setRaw(&node, "obfs", strings.TrimSpace(proxy.Obfs))
		setRaw(&node, "obfsParam", strings.TrimSpace(proxy.ObfsParam))
	case model.ProtocolVMess:
		node.Auth.UUID = strings.TrimSpace(proxy.UUID)
		node.TLS.Enabled = proxy.TLS
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.ServerName), strings.TrimSpace(proxy.SNI))
		node.TLS.ClientFingerprint = strings.TrimSpace(proxy.ClientFingerprint)
		setRaw(&node, "cipher", firstNonEmpty(strings.TrimSpace(proxy.Cipher), strings.TrimSpace(proxy.Encryption), "auto"))
		setRaw(&node, "alterId", proxy.AlterID)
		applyMihomoTransport(&node, proxy)
	case model.ProtocolVLESS:
		node.Auth.UUID = strings.TrimSpace(proxy.UUID)
		node.TLS.Enabled = proxy.TLS || proxy.RealityOpts != nil || strings.TrimSpace(proxy.SNI) != "" || strings.TrimSpace(proxy.ServerName) != ""
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.ServerName), strings.TrimSpace(proxy.SNI))
		node.TLS.ClientFingerprint = strings.TrimSpace(proxy.ClientFingerprint)
		setRaw(&node, "encryption", firstNonEmpty(strings.TrimSpace(proxy.Encryption), "none"))
		setRaw(&node, "packetEncoding", strings.TrimSpace(proxy.PacketEncoding))
		setRaw(&node, "flow", strings.TrimSpace(proxy.Flow))
		if proxy.RealityOpts != nil {
			node.TLS.Reality = &model.RealityOptions{
				PublicKey: strings.TrimSpace(proxy.RealityOpts.PublicKey),
				ShortID:   strings.TrimSpace(proxy.RealityOpts.ShortID),
				SpiderX:   strings.TrimSpace(proxy.RealityOpts.SpiderX),
			}
		}
		applyMihomoTransport(&node, proxy)
	case model.ProtocolTrojan:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), proxy.ALPN...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = *proxy.SkipCertVerify
		}
		applyMihomoTransport(&node, proxy)
	case model.ProtocolHysteria2:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), proxy.ALPN...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = *proxy.SkipCertVerify
		}
		setRaw(&node, "obfs", strings.TrimSpace(proxy.Obfs))
		setRaw(&node, "obfsPassword", strings.TrimSpace(proxy.ObfsPassword))
	case model.ProtocolTUIC:
		node.Auth.UUID = strings.TrimSpace(proxy.UUID)
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), proxy.ALPN...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = *proxy.SkipCertVerify
		}
		setRaw(&node, "congestionController", strings.TrimSpace(proxy.CongestionController))
		setRaw(&node, "udpRelayMode", strings.TrimSpace(proxy.UDPRelayMode))
		if proxy.ReduceRTT != nil {
			setRaw(&node, "reduceRTT", *proxy.ReduceRTT)
		}
	case model.ProtocolAnyTLS:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), proxy.ALPN...)
		node.TLS.ClientFingerprint = strings.TrimSpace(proxy.ClientFingerprint)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = *proxy.SkipCertVerify
		}
		setRaw(&node, "idleSessionCheckInterval", proxy.IdleSessionCheckInterval)
		setRaw(&node, "idleSessionTimeout", proxy.IdleSessionTimeout)
		setRaw(&node, "minIdleSession", proxy.MinIdleSession)
	case model.ProtocolWireGuard:
		node.Auth.PrivateKey = strings.TrimSpace(proxy.PrivateKey)
		node.Auth.PublicKey = strings.TrimSpace(proxy.PublicKey)
		node.Auth.PreSharedKey = strings.TrimSpace(proxy.PreSharedKey)
		node.WireGuard = &model.WireGuardOptions{
			IP:                  strings.TrimSpace(proxy.IP),
			IPv6:                strings.TrimSpace(proxy.IPv6),
			AllowedIPs:          append([]string(nil), proxy.AllowedIPs...),
			MTU:                 proxy.MTU,
			PersistentKeepalive: proxy.PersistentKeepalive,
			DNS:                 append([]string(nil), proxy.DNS...),
			AmneziaWG:           cloneAnyMap(proxy.AmneziaWGOption),
		}
		if proxy.RemoteDNSResolve != nil {
			node.WireGuard.RemoteDNSResolve = *proxy.RemoteDNSResolve
		}
		node.WireGuard.Reserved = parseReserved(proxy.Reserved)
		if len(proxy.Peers) > 0 {
			node.Server = ""
			node.Port = 0
			node.Auth.PublicKey = ""
			node.Auth.PreSharedKey = ""
			node.WireGuard.AllowedIPs = nil
			node.WireGuard.Peers = parseWGPeers(proxy.Peers)
		}
	default:
		return model.NodeIR{}, fmt.Errorf("unsupported yaml proxy type %q", proxy.Type)
	}

	return node, nil
}

func applyMihomoTransport(node *model.NodeIR, proxy mihomoYAMLProxy) {
	node.Transport.Network = strings.TrimSpace(proxy.Network)
	switch node.Transport.Network {
	case "ws":
		if proxy.WSOpts != nil {
			node.Transport.Path = strings.TrimSpace(proxy.WSOpts.Path)
			node.Transport.Host = strings.TrimSpace(proxy.WSOpts.Headers["Host"])
		}
	case "grpc":
		if proxy.GrpcOpts != nil {
			node.Transport.ServiceName = strings.TrimSpace(proxy.GrpcOpts.GrpcServiceName)
		}
	case "h2":
		if proxy.H2Opts != nil {
			node.Transport.Path = strings.TrimSpace(proxy.H2Opts.Path)
			node.Transport.H2Hosts = append([]string(nil), proxy.H2Opts.Host...)
			if len(proxy.H2Opts.Host) > 0 {
				node.Transport.Host = strings.TrimSpace(proxy.H2Opts.Host[0])
			}
		}
	case "xhttp":
		if proxy.XHTTPOpts != nil {
			node.Transport.Path = strings.TrimSpace(proxy.XHTTPOpts.Path)
			node.Transport.Mode = strings.TrimSpace(proxy.XHTTPOpts.Mode)
			if proxy.XHTTPOpts.NoGRPCHeader != nil {
				v := *proxy.XHTTPOpts.NoGRPCHeader
				node.Transport.NoGRPCHeader = &v
				setRaw(node, "noGrpcHeader", v)
			}
		}
	}
}

func parseWGPeers(peers []mihomoYAMLWGPeer) []model.WGPeer {
	out := make([]model.WGPeer, 0, len(peers))
	for _, peer := range peers {
		out = append(out, model.WGPeer{
			Server:       strings.TrimSpace(peer.Server),
			Port:         peer.Port,
			PublicKey:    strings.TrimSpace(peer.PublicKey),
			PreSharedKey: strings.TrimSpace(peer.PreSharedKey),
			AllowedIPs:   append([]string(nil), peer.AllowedIPs...),
			Reserved:     parseReserved(peer.Reserved),
		})
	}
	return out
}

func parseReserved(value interface{}) []int {
	switch typed := value.(type) {
	case []interface{}:
		out := make([]int, 0, len(typed))
		for _, item := range typed {
			if n, ok := normalizeInt(item); ok {
				out = append(out, n)
			}
		}
		return out
	case []int:
		return append([]int(nil), typed...)
	case string:
		parts := parseCSV(typed)
		out := make([]int, 0, len(parts))
		for _, part := range parts {
			if n, ok := parseIntString(part); ok {
				out = append(out, n)
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeInt(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}

func cloneAnyMap(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
