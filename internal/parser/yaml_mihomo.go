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
	Port                     yamlInt                `yaml:"port,omitempty"`
	PortRange                string                 `yaml:"port-range,omitempty"`
	Transport                string                 `yaml:"transport,omitempty"`
	Encryption               string                 `yaml:"encryption,omitempty"`
	Cipher                   string                 `yaml:"cipher,omitempty"`
	Password                 string                 `yaml:"password,omitempty"`
	Protocol                 string                 `yaml:"protocol,omitempty"`
	ProtocolParam            string                 `yaml:"protocol-param,omitempty"`
	Username                 string                 `yaml:"username,omitempty"`
	UUID                     string                 `yaml:"uuid,omitempty"`
	AlterID                  interface{}            `yaml:"alterId,omitempty"`
	Network                  string                 `yaml:"network,omitempty"`
	TLS                      yamlBool               `yaml:"tls,omitempty"`
	UDP                      *yamlBool              `yaml:"udp,omitempty"`
	ServerName               string                 `yaml:"servername,omitempty"`
	SNI                      string                 `yaml:"sni,omitempty"`
	ClientFingerprint        string                 `yaml:"client-fingerprint,omitempty"`
	PacketEncoding           string                 `yaml:"packet-encoding,omitempty"`
	SkipCertVerify           *yamlBool              `yaml:"skip-cert-verify,omitempty"`
	ALPN                     yamlStringList         `yaml:"alpn,omitempty"`
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
	ReduceRTT                *yamlBool              `yaml:"reduce-rtt,omitempty"`
	IdleSessionCheckInterval interface{}            `yaml:"idle-session-check-interval,omitempty"`
	IdleSessionTimeout       interface{}            `yaml:"idle-session-timeout,omitempty"`
	MinIdleSession           interface{}            `yaml:"min-idle-session,omitempty"`
	Multiplexing             string                 `yaml:"multiplexing,omitempty"`
	HandshakeMode            string                 `yaml:"handshake-mode,omitempty"`
	TrafficPattern           string                 `yaml:"traffic-pattern,omitempty"`
	IP                       string                 `yaml:"ip,omitempty"`
	IPv6                     string                 `yaml:"ipv6,omitempty"`
	PrivateKey               string                 `yaml:"private-key,omitempty"`
	PublicKey                string                 `yaml:"public-key,omitempty"`
	AllowedIPs               yamlStringList         `yaml:"allowed-ips,omitempty"`
	PreSharedKey             string                 `yaml:"pre-shared-key,omitempty"`
	Reserved                 interface{}            `yaml:"reserved,omitempty"`
	PersistentKeepalive      yamlInt                `yaml:"persistent-keepalive,omitempty"`
	MTU                      yamlInt                `yaml:"mtu,omitempty"`
	RemoteDNSResolve         *yamlBool              `yaml:"remote-dns-resolve,omitempty"`
	DNS                      yamlStringList         `yaml:"dns,omitempty"`
	Peers                    []mihomoYAMLWGPeer     `yaml:"peers,omitempty"`
	AmneziaWGOption          map[string]interface{} `yaml:"amnezia-wg-option,omitempty"`
	OriginalFields           map[string]interface{} `yaml:"-"`
}

const mihomoProxyFieldsRawKey = "_mihomoProxyFields"

type yamlInt int

func (n *yamlInt) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	var raw interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if parsed, ok := normalizeInt(raw); ok {
		*n = yamlInt(parsed)
	}
	return nil
}

type yamlBool bool

func (b *yamlBool) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	var raw interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	switch typed := raw.(type) {
	case bool:
		*b = yamlBool(typed)
	case string:
		*b = yamlBool(parseBoolString(typed))
	}
	return nil
}

type yamlStringList []string

func (list *yamlStringList) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	switch value.Kind {
	case yaml.SequenceNode:
		out := make([]string, 0, len(value.Content))
		for _, item := range value.Content {
			text := strings.TrimSpace(fmt.Sprint(yamlNodeToInterface(item)))
			if text != "" {
				out = append(out, text)
			}
		}
		*list = out
	case yaml.ScalarNode:
		*list = yamlStringList(parseCSV(value.Value))
	}
	return nil
}

type yamlStringMap map[string]string

func (m *yamlStringMap) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	switch value.Kind {
	case yaml.MappingNode:
		out := make(map[string]string, len(value.Content)/2)
		for index := 0; index+1 < len(value.Content); index += 2 {
			key := strings.TrimSpace(value.Content[index].Value)
			if key == "" {
				continue
			}
			out[key] = strings.TrimSpace(fmt.Sprint(yamlNodeToInterface(value.Content[index+1])))
		}
		*m = out
	case yaml.ScalarNode:
		if header := strings.TrimSpace(value.Value); header != "" {
			*m = yamlStringMap{"Host": header}
		}
	}
	return nil
}

func (proxy *mihomoYAMLProxy) UnmarshalYAML(value *yaml.Node) error {
	type rawProxy mihomoYAMLProxy
	var decoded rawProxy
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*proxy = mihomoYAMLProxy(decoded)
	proxy.OriginalFields = yamlMappingNodeToMap(value)
	return nil
}

type mihomoYAMLRealityOpts struct {
	PublicKey string `yaml:"public-key,omitempty"`
	ShortID   string `yaml:"short-id,omitempty"`
	SpiderX   string `yaml:"spider-x,omitempty"`
}

type mihomoYAMLWSOpts struct {
	Path    string        `yaml:"path,omitempty"`
	Headers yamlStringMap `yaml:"headers,omitempty"`
}

type mihomoYAMLGRPCOpts struct {
	GrpcServiceName string `yaml:"grpc-service-name,omitempty"`
}

type mihomoYAMLH2Opts struct {
	Host yamlStringList `yaml:"host,omitempty"`
	Path string         `yaml:"path,omitempty"`
}

type mihomoYAMLXHTTPOpts struct {
	Path         string    `yaml:"path,omitempty"`
	Mode         string    `yaml:"mode,omitempty"`
	NoGRPCHeader *yamlBool `yaml:"no-grpc-header,omitempty"`
}

type mihomoYAMLWGPeer struct {
	Server       string         `yaml:"server,omitempty"`
	Port         yamlInt        `yaml:"port,omitempty"`
	PublicKey    string         `yaml:"public-key,omitempty"`
	PreSharedKey string         `yaml:"pre-shared-key,omitempty"`
	AllowedIPs   yamlStringList `yaml:"allowed-ips,omitempty"`
	Reserved     interface{}    `yaml:"reserved,omitempty"`
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
	node.Port = int(proxy.Port)
	if len(proxy.OriginalFields) > 0 {
		setRaw(&node, mihomoProxyFieldsRawKey, proxy.OriginalFields)
	}
	if proxy.UDP != nil {
		v := bool(*proxy.UDP)
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
		node.TLS.Enabled = bool(proxy.TLS)
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.ServerName), strings.TrimSpace(proxy.SNI))
		node.TLS.ClientFingerprint = strings.TrimSpace(proxy.ClientFingerprint)
		setRaw(&node, "cipher", firstNonEmpty(strings.TrimSpace(proxy.Cipher), strings.TrimSpace(proxy.Encryption), "auto"))
		setRaw(&node, "alterId", proxy.AlterID)
		applyMihomoTransport(&node, proxy)
	case model.ProtocolVLESS:
		node.Auth.UUID = strings.TrimSpace(proxy.UUID)
		node.TLS.Enabled = bool(proxy.TLS) || proxy.RealityOpts != nil || strings.TrimSpace(proxy.SNI) != "" || strings.TrimSpace(proxy.ServerName) != ""
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.ServerName), strings.TrimSpace(proxy.SNI))
		node.TLS.ClientFingerprint = strings.TrimSpace(proxy.ClientFingerprint)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
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
		node.TLS.ALPN = append([]string(nil), []string(proxy.ALPN)...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
		applyMihomoTransport(&node, proxy)
	case model.ProtocolHysteria:
		node.Auth.Password = firstNonEmpty(
			strings.TrimSpace(proxy.Password),
			mihomoOriginalFieldString(proxy.OriginalFields, "auth-str", "auth_str", "auth"),
		)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), []string(proxy.ALPN)...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
		setRaw(&node, "obfs", strings.TrimSpace(proxy.Obfs))
		setRaw(&node, "obfsParam", strings.TrimSpace(proxy.ObfsParam))
	case model.ProtocolHysteria2:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), []string(proxy.ALPN)...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
		setRaw(&node, "obfs", strings.TrimSpace(proxy.Obfs))
		setRaw(&node, "obfsPassword", strings.TrimSpace(proxy.ObfsPassword))
	case model.ProtocolTUIC:
		node.Auth.UUID = strings.TrimSpace(proxy.UUID)
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), []string(proxy.ALPN)...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
		setRaw(&node, "congestionController", strings.TrimSpace(proxy.CongestionController))
		setRaw(&node, "udpRelayMode", strings.TrimSpace(proxy.UDPRelayMode))
		if proxy.ReduceRTT != nil {
			setRaw(&node, "reduceRTT", bool(*proxy.ReduceRTT))
		}
	case model.ProtocolAnyTLS:
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = true
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), []string(proxy.ALPN)...)
		node.TLS.ClientFingerprint = strings.TrimSpace(proxy.ClientFingerprint)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
		setRaw(&node, "idleSessionCheckInterval", proxy.IdleSessionCheckInterval)
		setRaw(&node, "idleSessionTimeout", proxy.IdleSessionTimeout)
		setRaw(&node, "minIdleSession", proxy.MinIdleSession)
	case model.ProtocolMieru:
		node.Auth.Username = strings.TrimSpace(proxy.Username)
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		setRaw(&node, "portRange", strings.TrimSpace(proxy.PortRange))
		setRaw(&node, "transport", normalizeMieruTransport(proxy.Transport))
		setRaw(&node, "multiplexing", strings.TrimSpace(proxy.Multiplexing))
		setRaw(&node, "handshakeMode", strings.TrimSpace(proxy.HandshakeMode))
		setRaw(&node, "trafficPattern", strings.TrimSpace(proxy.TrafficPattern))
	case model.ProtocolHTTP, model.ProtocolSOCKS5:
		node.Auth.Username = strings.TrimSpace(proxy.Username)
		node.Auth.Password = strings.TrimSpace(proxy.Password)
		node.TLS.Enabled = bool(proxy.TLS)
		node.TLS.SNI = firstNonEmpty(strings.TrimSpace(proxy.SNI), strings.TrimSpace(proxy.ServerName))
		node.TLS.ALPN = append([]string(nil), []string(proxy.ALPN)...)
		if proxy.SkipCertVerify != nil {
			node.TLS.Insecure = bool(*proxy.SkipCertVerify)
		}
	case model.ProtocolWireGuard:
		node.Auth.PrivateKey = strings.TrimSpace(proxy.PrivateKey)
		node.Auth.PublicKey = strings.TrimSpace(proxy.PublicKey)
		node.Auth.PreSharedKey = strings.TrimSpace(proxy.PreSharedKey)
		node.WireGuard = &model.WireGuardOptions{
			IP:                  strings.TrimSpace(proxy.IP),
			IPv6:                strings.TrimSpace(proxy.IPv6),
			AllowedIPs:          append([]string(nil), []string(proxy.AllowedIPs)...),
			MTU:                 int(proxy.MTU),
			PersistentKeepalive: int(proxy.PersistentKeepalive),
			DNS:                 append([]string(nil), []string(proxy.DNS)...),
			AmneziaWG:           cloneAnyMap(proxy.AmneziaWGOption),
		}
		if proxy.RemoteDNSResolve != nil {
			node.WireGuard.RemoteDNSResolve = bool(*proxy.RemoteDNSResolve)
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
			node.Transport.H2Hosts = append([]string(nil), []string(proxy.H2Opts.Host)...)
			if len(proxy.H2Opts.Host) > 0 {
				node.Transport.Host = strings.TrimSpace(proxy.H2Opts.Host[0])
			}
		}
	case "xhttp":
		if proxy.XHTTPOpts != nil {
			node.Transport.Path = strings.TrimSpace(proxy.XHTTPOpts.Path)
			node.Transport.Mode = strings.TrimSpace(proxy.XHTTPOpts.Mode)
			if proxy.XHTTPOpts.NoGRPCHeader != nil {
				v := bool(*proxy.XHTTPOpts.NoGRPCHeader)
				node.Transport.NoGRPCHeader = &v
				setRaw(node, "noGrpcHeader", v)
			}
		}
	}
}

func mihomoOriginalFieldString(fields map[string]interface{}, keys ...string) string {
	if len(fields) == 0 {
		return ""
	}
	for _, key := range keys {
		value, ok := fields[key]
		if !ok || value == nil {
			continue
		}
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" {
			return text
		}
	}
	return ""
}

func parseWGPeers(peers []mihomoYAMLWGPeer) []model.WGPeer {
	out := make([]model.WGPeer, 0, len(peers))
	for _, peer := range peers {
		out = append(out, model.WGPeer{
			Server:       strings.TrimSpace(peer.Server),
			Port:         int(peer.Port),
			PublicKey:    strings.TrimSpace(peer.PublicKey),
			PreSharedKey: strings.TrimSpace(peer.PreSharedKey),
			AllowedIPs:   append([]string(nil), []string(peer.AllowedIPs)...),
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
	case string:
		return parseIntString(typed)
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

func yamlMappingNodeToMap(node *yaml.Node) map[string]interface{} {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	out := make(map[string]interface{}, len(node.Content)/2)
	for index := 0; index+1 < len(node.Content); index += 2 {
		keyNode := node.Content[index]
		valueNode := node.Content[index+1]
		key := strings.TrimSpace(keyNode.Value)
		if key == "" {
			continue
		}
		out[key] = yamlNodeToInterface(valueNode)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func yamlNodeToInterface(node *yaml.Node) interface{} {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.AliasNode:
		return yamlNodeToInterface(node.Alias)
	case yaml.MappingNode:
		return yamlMappingNodeToMap(node)
	case yaml.SequenceNode:
		out := make([]interface{}, 0, len(node.Content))
		for _, item := range node.Content {
			out = append(out, yamlNodeToInterface(item))
		}
		return out
	case yaml.ScalarNode:
		var out interface{}
		if err := node.Decode(&out); err == nil {
			return out
		}
		return node.Value
	default:
		var out interface{}
		if err := node.Decode(&out); err == nil {
			return out
		}
		return node.Value
	}
}
