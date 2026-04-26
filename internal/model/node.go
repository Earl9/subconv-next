package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type Protocol string

const (
	ProtocolSS        Protocol = "ss"
	ProtocolVMess     Protocol = "vmess"
	ProtocolVLESS     Protocol = "vless"
	ProtocolTrojan    Protocol = "trojan"
	ProtocolHysteria2 Protocol = "hysteria2"
	ProtocolTUIC      Protocol = "tuic"
	ProtocolAnyTLS    Protocol = "anytls"
	ProtocolWireGuard Protocol = "wireguard"
	ProtocolHTTP      Protocol = "http"
	ProtocolSOCKS5    Protocol = "socks5"
)

type NodeIR struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      Protocol               `json:"type"`
	Server    string                 `json:"server"`
	Port      int                    `json:"port,omitempty"`
	Auth      Auth                   `json:"auth,omitempty"`
	TLS       TLSOptions             `json:"tls,omitempty"`
	Transport TransportOptions       `json:"transport,omitempty"`
	WireGuard *WireGuardOptions      `json:"wireguard,omitempty"`
	UDP       *bool                  `json:"udp,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Source    SourceInfo             `json:"source,omitempty"`
	Raw       map[string]interface{} `json:"raw,omitempty"`
	Warnings  []string               `json:"warnings,omitempty"`
}

type Auth struct {
	UUID         string `json:"uuid,omitempty"`
	Password     string `json:"password,omitempty"`
	Username     string `json:"username,omitempty"`
	Token        string `json:"token,omitempty"`
	PrivateKey   string `json:"private_key,omitempty"`
	PublicKey    string `json:"public_key,omitempty"`
	PreSharedKey string `json:"pre_shared_key,omitempty"`
}

type TLSOptions struct {
	Enabled           bool            `json:"enabled,omitempty"`
	SNI               string          `json:"sni,omitempty"`
	ALPN              []string        `json:"alpn,omitempty"`
	Insecure          bool            `json:"insecure,omitempty"`
	Fingerprint       string          `json:"fingerprint,omitempty"`
	ClientFingerprint string          `json:"client_fingerprint,omitempty"`
	Reality           *RealityOptions `json:"reality,omitempty"`
	ECH               *ECHOptions     `json:"ech,omitempty"`
}

type RealityOptions struct {
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
	SpiderX   string `json:"spider_x,omitempty"`
}

type ECHOptions struct {
	Enabled bool   `json:"enabled,omitempty"`
	Config  string `json:"config,omitempty"`
}

type TransportOptions struct {
	Network     string            `json:"network,omitempty"`
	Path        string            `json:"path,omitempty"`
	Host        string            `json:"host,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type WireGuardOptions struct {
	IP                   string                 `json:"ip,omitempty"`
	IPv6                 string                 `json:"ipv6,omitempty"`
	AllowedIPs           []string               `json:"allowed_ips,omitempty"`
	Reserved             []int                  `json:"reserved,omitempty"`
	ReservedString       string                 `json:"reserved_string,omitempty"`
	MTU                  int                    `json:"mtu,omitempty"`
	PersistentKeepalive  int                    `json:"persistent_keepalive,omitempty"`
	RemoteDNSResolve     bool                   `json:"remote_dns_resolve,omitempty"`
	DNS                  []string               `json:"dns,omitempty"`
	Peers                []WGPeer               `json:"peers,omitempty"`
	AmneziaWG            map[string]interface{} `json:"amnezia_wg,omitempty"`
}

type WGPeer struct {
	Server       string   `json:"server,omitempty"`
	Port         int      `json:"port,omitempty"`
	PublicKey    string   `json:"public_key,omitempty"`
	PreSharedKey string   `json:"pre_shared_key,omitempty"`
	AllowedIPs   []string `json:"allowed_ips,omitempty"`
	Reserved     []int    `json:"reserved,omitempty"`
}

type SourceInfo struct {
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}

func Bool(v bool) *bool {
	return &v
}

func NormalizeNodes(nodes []NodeIR) []NodeIR {
	normalized := make([]NodeIR, 0, len(nodes))
	for _, node := range nodes {
		normalized = append(normalized, NormalizeNode(node))
	}
	return DedupeNodes(normalized)
}

func NormalizeNode(node NodeIR) NodeIR {
	node.Type = Protocol(strings.ToLower(strings.TrimSpace(string(node.Type))))
	node.Name = sanitizeText(node.Name)
	node.Server = strings.ToLower(strings.TrimSpace(node.Server))
	node.Source.Name = sanitizeText(node.Source.Name)
	node.Source.Kind = sanitizeText(node.Source.Kind)

	if node.Port < 0 {
		node.Port = 0
	}

	if node.Name == "" {
		node.Name = defaultNodeName(node)
	}

	node.TLS.SNI = sanitizeText(node.TLS.SNI)
	node.TLS.Fingerprint = sanitizeText(node.TLS.Fingerprint)
	node.TLS.ClientFingerprint = sanitizeText(node.TLS.ClientFingerprint)
	node.TLS.ALPN = cleanStringSlice(node.TLS.ALPN)
	node.Tags = mergeUniqueStrings(node.Tags, inferRegionTags(node.Name))
	node.Warnings = cleanStringSlice(node.Warnings)

	if node.ID == "" {
		node.ID = StableNodeID(node)
	}

	return node
}

func StableNodeID(node NodeIR) string {
	payload := strings.Join([]string{
		string(node.Type),
		strings.ToLower(strings.TrimSpace(node.Server)),
		fmt.Sprintf("%d", node.Port),
		sanitizeText(node.Name),
		authFingerprint(node.Auth),
	}, "|")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func DedupeNodes(nodes []NodeIR) []NodeIR {
	seen := make(map[string]int, len(nodes))
	out := make([]NodeIR, 0, len(nodes))

	for _, node := range nodes {
		key := dedupeKey(node)
		if idx, ok := seen[key]; ok {
			existing := out[idx]
			existing.Tags = mergeUniqueStrings(existing.Tags, node.Tags)
			existing.Warnings = mergeUniqueStrings(existing.Warnings, node.Warnings)
			existing.Warnings = mergeUniqueStrings(existing.Warnings, []string{duplicateWarning(node.Source)})
			out[idx] = existing
			continue
		}

		seen[key] = len(out)
		out = append(out, node)
	}

	return out
}

func defaultNodeName(node NodeIR) string {
	parts := []string{string(node.Type)}
	if node.Server != "" {
		parts = append(parts, node.Server)
	}
	if node.Port > 0 {
		parts = append(parts, fmt.Sprintf("%d", node.Port))
	}
	if len(parts) == 0 {
		return "node"
	}
	return strings.Join(parts, "-")
}

func authFingerprint(auth Auth) string {
	values := []string{
		strings.TrimSpace(auth.UUID),
		strings.TrimSpace(auth.Password),
		strings.TrimSpace(auth.Username),
		strings.TrimSpace(auth.Token),
		strings.TrimSpace(auth.PrivateKey),
		strings.TrimSpace(auth.PublicKey),
		strings.TrimSpace(auth.PreSharedKey),
	}
	return strings.Join(values, "|")
}

func dedupeKey(node NodeIR) string {
	transport := strings.Join([]string{
		node.Transport.Network,
		node.Transport.Path,
		node.Transport.Host,
		node.Transport.ServiceName,
	}, "|")

	authKey := firstNonEmpty(
		node.Auth.UUID,
		node.Auth.Password,
		node.Auth.PrivateKey,
		node.Auth.Token,
		node.Auth.Username,
	)

	return strings.Join([]string{
		string(node.Type),
		strings.ToLower(strings.TrimSpace(node.Server)),
		fmt.Sprintf("%d", node.Port),
		authKey,
		transport,
		node.TLS.SNI,
	}, "|")
}

func duplicateWarning(source SourceInfo) string {
	if source.Name != "" {
		return fmt.Sprintf("duplicate skipped from source %s", source.Name)
	}
	return "duplicate skipped"
}

func inferRegionTags(name string) []string {
	matchers := []struct {
		tag      string
		patterns []string
	}{
		{tag: "HK", patterns: []string{"香港", "HK", "HONG KONG", "港"}},
		{tag: "JP", patterns: []string{"日本", "JP", "JAPAN", "日"}},
		{tag: "US", patterns: []string{"美国", "US", "UNITED STATES", "美"}},
		{tag: "SG", patterns: []string{"新加坡", "SG", "SINGAPORE", "狮城"}},
		{tag: "TW", patterns: []string{"台湾", "TW", "TAIWAN", "台"}},
		{tag: "KR", patterns: []string{"韩国", "KR", "KOREA", "韩"}},
		{tag: "DE", patterns: []string{"德国", "DE", "GERMANY"}},
		{tag: "GB", patterns: []string{"英国", "GB", "UK", "UNITED KINGDOM"}},
		{tag: "FR", patterns: []string{"法国", "FR", "FRANCE"}},
		{tag: "CA", patterns: []string{"加拿大", "CA", "CANADA"}},
		{tag: "AU", patterns: []string{"澳大利亚", "AU", "AUSTRALIA"}},
	}

	upperName := strings.ToUpper(name)
	var tags []string
	for _, matcher := range matchers {
		for _, pattern := range matcher.patterns {
			if strings.Contains(upperName, strings.ToUpper(pattern)) {
				tags = append(tags, matcher.tag)
				break
			}
		}
	}

	return cleanStringSlice(tags)
}

func mergeUniqueStrings(a, b []string) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	var merged []string
	for _, value := range append(append([]string{}, a...), b...) {
		value = sanitizeText(value)
		if value == "" {
			continue
		}
		if _, ok := set[value]; ok {
			continue
		}
		set[value] = struct{}{}
		merged = append(merged, value)
	}
	sort.Strings(merged)
	return merged
}

func cleanStringSlice(values []string) []string {
	return mergeUniqueStrings(nil, values)
}

func sanitizeText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(' ')
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
