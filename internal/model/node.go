package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Protocol string

const (
	ProtocolSS        Protocol = "ss"
	ProtocolSSR       Protocol = "ssr"
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
	Sources   []SourceInfo           `json:"sources,omitempty"`
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
	Network      string            `json:"network,omitempty"`
	Path         string            `json:"path,omitempty"`
	Host         string            `json:"host,omitempty"`
	ServiceName  string            `json:"service_name,omitempty"`
	Mode         string            `json:"mode,omitempty"`
	NoGRPCHeader *bool             `json:"no_grpc_header,omitempty"`
	H2Hosts      []string          `json:"h2_hosts,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

type WireGuardOptions struct {
	IP                  string                 `json:"ip,omitempty"`
	IPv6                string                 `json:"ipv6,omitempty"`
	AllowedIPs          []string               `json:"allowed_ips,omitempty"`
	Reserved            []int                  `json:"reserved,omitempty"`
	ReservedString      string                 `json:"reserved_string,omitempty"`
	MTU                 int                    `json:"mtu,omitempty"`
	PersistentKeepalive int                    `json:"persistent_keepalive,omitempty"`
	RemoteDNSResolve    bool                   `json:"remote_dns_resolve,omitempty"`
	DNS                 []string               `json:"dns,omitempty"`
	Peers               []WGPeer               `json:"peers,omitempty"`
	AmneziaWG           map[string]interface{} `json:"amnezia_wg,omitempty"`
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
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Emoji   string `json:"emoji,omitempty"`
	Kind    string `json:"kind,omitempty"`
	URLHash string `json:"url_hash,omitempty"`
}

func Bool(v bool) *bool {
	return &v
}

func NormalizeNodes(nodes []NodeIR) []NodeIR {
	return NormalizeNodesWithScope(nodes, "global")
}

func NormalizeNodesNoDedupe(nodes []NodeIR) []NodeIR {
	normalized := make([]NodeIR, 0, len(nodes))
	for _, node := range nodes {
		normalized = append(normalized, NormalizeNode(node))
	}
	return normalized
}

func NormalizeNodesWithScope(nodes []NodeIR, scope string) []NodeIR {
	normalized := NormalizeNodesNoDedupe(nodes)
	return DedupeNodesByScope(normalized, scope)
}

func NormalizeNode(node NodeIR) NodeIR {
	node.Type = Protocol(strings.ToLower(strings.TrimSpace(string(node.Type))))
	node.Server = strings.ToLower(strings.TrimSpace(node.Server))
	node.Source.ID = sanitizeText(node.Source.ID)
	node.Source.Name = sanitizeText(node.Source.Name)
	node.Source.Emoji = strings.TrimSpace(node.Source.Emoji)
	node.Source.Kind = sanitizeText(node.Source.Kind)
	node.Source.URLHash = sanitizeText(node.Source.URLHash)
	node.Sources = normalizeSourceList(node.Source, node.Sources)

	if node.Port < 0 {
		node.Port = 0
	}

	if strings.TrimSpace(node.Name) == "" {
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
	return StableNodeIDWithScope(node, "global", 0)
}

func StableNodeIDWithScope(node NodeIR, scope string, originalIndex int) string {
	authDigest := sha256Hex(authFingerprint(node.Auth))
	transportDigest := sha256Hex(transportFingerprint(node))
	parts := []string{
		string(node.Type),
		strings.ToLower(strings.TrimSpace(node.Server)),
		fmt.Sprintf("%d", node.Port),
		authDigest,
		transportDigest,
	}
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "per_source":
		parts = append(parts, strings.ToLower(strings.TrimSpace(node.Source.ID)))
	case "none":
		parts = append(parts, strings.ToLower(strings.TrimSpace(node.Source.ID)), fmt.Sprintf("%d", originalIndex))
	default:
	}
	payload := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func transportFingerprint(node NodeIR) string {
	parts := []string{
		strings.TrimSpace(node.Transport.Network),
		strings.TrimSpace(node.Transport.Path),
		strings.TrimSpace(node.Transport.Host),
		strings.TrimSpace(node.Transport.ServiceName),
		strings.TrimSpace(node.Transport.Mode),
		strings.TrimSpace(node.TLS.SNI),
		strings.TrimSpace(node.TLS.ClientFingerprint),
		strings.TrimSpace(node.TLS.Fingerprint),
	}

	if node.Transport.NoGRPCHeader != nil {
		parts = append(parts, strconv.FormatBool(*node.Transport.NoGRPCHeader))
	}
	parts = append(parts, append([]string(nil), node.Transport.H2Hosts...)...)

	if node.TLS.Reality != nil {
		parts = append(parts,
			strings.TrimSpace(node.TLS.Reality.PublicKey),
			strings.TrimSpace(node.TLS.Reality.ShortID),
			strings.TrimSpace(node.TLS.Reality.SpiderX),
		)
	}
	if node.TLS.ECH != nil {
		parts = append(parts, strconv.FormatBool(node.TLS.ECH.Enabled), strings.TrimSpace(node.TLS.ECH.Config))
	}
	if node.WireGuard != nil {
		parts = append(parts,
			strings.TrimSpace(node.WireGuard.IP),
			strings.TrimSpace(node.WireGuard.IPv6),
			strconv.Itoa(node.WireGuard.MTU),
			strconv.Itoa(node.WireGuard.PersistentKeepalive),
		)
		parts = append(parts, append([]string(nil), node.WireGuard.AllowedIPs...)...)
	}
	return strings.Join(parts, "|")
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func DedupeNodes(nodes []NodeIR) []NodeIR {
	return DedupeNodesByScope(nodes, "global")
}

func DedupeNodesByScope(nodes []NodeIR, scope string) []NodeIR {
	seen := make(map[string]int, len(nodes))
	out := make([]NodeIR, 0, len(nodes))

	for index, node := range nodes {
		node.ID = StableNodeIDWithScope(node, scope, index)
		if strings.EqualFold(strings.TrimSpace(node.Source.Kind), "custom") && !strings.HasPrefix(node.ID, "custom-") {
			node.ID = "custom-" + node.ID
		}
		key := dedupeKey(node, scope, index)
		if idx, ok := seen[key]; ok {
			if strings.EqualFold(strings.TrimSpace(scope), "none") {
				seen[key] = len(out)
				out = append(out, node)
				continue
			}
			existing := out[idx]
			existing.Tags = mergeUniqueStrings(existing.Tags, node.Tags)
			existing.Warnings = mergeUniqueStrings(existing.Warnings, node.Warnings)
			existing.Warnings = mergeUniqueStrings(existing.Warnings, []string{duplicateWarning(node.Source)})
			existing.Sources = mergeSources(existing.Sources, append([]SourceInfo{existing.Source}, node.Source)...)
			out[idx] = existing
			continue
		}

		seen[key] = len(out)
		out = append(out, node)
	}

	return out
}

func normalizeSourceList(primary SourceInfo, sources []SourceInfo) []SourceInfo {
	var out []SourceInfo
	if primary.ID != "" || primary.Name != "" || primary.Emoji != "" || primary.Kind != "" || primary.URLHash != "" {
		out = append(out, normalizeSource(primary))
	}
	for _, source := range sources {
		out = append(out, normalizeSource(source))
	}
	return mergeSources(nil, out...)
}

func normalizeSource(source SourceInfo) SourceInfo {
	source.ID = sanitizeText(source.ID)
	source.Name = sanitizeText(source.Name)
	source.Emoji = strings.TrimSpace(source.Emoji)
	source.Kind = sanitizeText(source.Kind)
	source.URLHash = sanitizeText(source.URLHash)
	return source
}

func mergeSources(existing []SourceInfo, more ...SourceInfo) []SourceInfo {
	seen := make(map[string]struct{}, len(existing)+len(more))
	var out []SourceInfo
	combined := append(append([]SourceInfo{}, existing...), more...)
	for _, source := range combined {
		source = normalizeSource(source)
		key := strings.Join([]string{source.ID, source.Name, source.Emoji, source.Kind, source.URLHash}, "|")
		if key == "||||" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, source)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func MergeSourcesForView(node NodeIR) []SourceInfo {
	return mergeSources(node.Sources, node.Source)
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

func dedupeKey(node NodeIR, scope string, originalIndex int) string {
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
		sourceDedupeKey(node.Source, scope, originalIndex),
	}, "|")
}

func sourceDedupeKey(source SourceInfo, scope string, originalIndex int) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "per_source":
		return strings.ToLower(strings.TrimSpace(source.ID))
	case "none":
		return strings.ToLower(strings.TrimSpace(source.ID)) + "|" + fmt.Sprintf("%d", originalIndex)
	default:
		return ""
	}
}

func duplicateWarning(source SourceInfo) string {
	if source.Name != "" {
		return fmt.Sprintf("duplicate skipped from source %s", source.Name)
	}
	return "duplicate skipped"
}

func DefaultNameOptions() NameOptions {
	return NameOptions{
		KeepRawName:           true,
		SourcePrefixMode:      "emoji_name",
		SourcePrefixSeparator: "｜",
		DedupeSuffixStyle:     "#n",
	}
}

func EffectiveNameOptions(render RenderConfig) NameOptions {
	options := render.NameOptions
	defaults := DefaultNameOptions()
	options.KeepRawName = true
	if !render.SourcePrefix {
		options.SourcePrefixMode = "none"
	} else if strings.TrimSpace(options.SourcePrefixMode) == "" {
		options.SourcePrefixMode = defaults.SourcePrefixMode
	}
	options.SourcePrefixMode = normalizeSourcePrefixMode(options.SourcePrefixMode)
	if strings.TrimSpace(options.SourcePrefixSeparator) == "" {
		options.SourcePrefixSeparator = defaults.SourcePrefixSeparator
	}
	if strings.TrimSpace(options.DedupeSuffixStyle) == "" {
		options.DedupeSuffixStyle = defaults.DedupeSuffixStyle
	}
	return options
}

func normalizeSourcePrefixMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "none", "emoji", "name", "emoji_name":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return DefaultNameOptions().SourcePrefixMode
	}
}

func BuildSourcePrefix(source SourceInfo, options NameOptions) string {
	mode := normalizeSourcePrefixMode(options.SourcePrefixMode)
	emoji := strings.TrimSpace(source.Emoji)
	name := strings.TrimSpace(source.Name)

	switch mode {
	case "none":
		return ""
	case "emoji":
		return emoji
	case "name":
		return name
	case "emoji_name":
		if emoji != "" && name != "" {
			return emoji + " " + name
		}
		if emoji != "" {
			return emoji
		}
		return name
	default:
		return ""
	}
}

func BuildYamlNodeName(rawName string, source SourceInfo, options NameOptions) string {
	defaults := DefaultNameOptions()
	if strings.TrimSpace(options.SourcePrefixMode) == "" {
		options.SourcePrefixMode = defaults.SourcePrefixMode
	}
	options.SourcePrefixMode = normalizeSourcePrefixMode(options.SourcePrefixMode)
	if strings.TrimSpace(options.SourcePrefixSeparator) == "" {
		options.SourcePrefixSeparator = defaults.SourcePrefixSeparator
	}

	name := rawName

	prefix := BuildSourcePrefix(source, options)
	if prefix == "" {
		return name
	}
	if name == "" {
		return prefix
	}
	if options.SourcePrefixMode == "emoji" {
		return prefix + " " + name
	}
	return prefix + options.SourcePrefixSeparator + name
}

func EnsureUniqueProxyNames(nodes []NodeIR, suffixStyle string) []NodeIR {
	out := append([]NodeIR(nil), nodes...)
	counts := make(map[string]int, len(out))
	for i := range out {
		base := out[i].Name
		if strings.TrimSpace(base) == "" {
			base = fmt.Sprintf("%s-%d", out[i].Type, i+1)
		}
		counts[base]++
		if counts[base] == 1 {
			out[i].Name = base
			continue
		}
		out[i].Name = appendDedupeSuffix(base, counts[base], suffixStyle)
	}
	return out
}

func appendDedupeSuffix(base string, index int, suffixStyle string) string {
	style := suffixStyle
	if style == "" {
		style = DefaultNameOptions().DedupeSuffixStyle
	}
	suffix := strings.ReplaceAll(style, "n", strconv.Itoa(index))
	if !strings.Contains(style, "n") {
		suffix = fmt.Sprintf("%s%d", style, index)
	}
	if suffix == "" {
		suffix = fmt.Sprintf("#%d", index)
	}
	if strings.HasPrefix(suffix, " ") {
		return base + suffix
	}
	return base + " " + suffix
}

func inferRegionTags(name string) []string {
	matchers := []struct {
		tag               string
		substringPatterns []string
		tokenPatterns     []string
	}{
		{tag: "HK", substringPatterns: []string{"香港", "HONG KONG", "港"}, tokenPatterns: []string{"HK"}},
		{tag: "JP", substringPatterns: []string{"日本", "JAPAN", "日"}, tokenPatterns: []string{"JP"}},
		{tag: "US", substringPatterns: []string{"美国", "UNITED STATES", "美"}, tokenPatterns: []string{"US"}},
		{tag: "SG", substringPatterns: []string{"新加坡", "SINGAPORE", "狮城"}, tokenPatterns: []string{"SG"}},
		{tag: "TW", substringPatterns: []string{"台湾", "TAIWAN", "台"}, tokenPatterns: []string{"TW"}},
		{tag: "KR", substringPatterns: []string{"韩国", "KOREA", "韩"}, tokenPatterns: []string{"KR"}},
		{tag: "DE", substringPatterns: []string{"德国", "GERMANY"}, tokenPatterns: []string{"DE"}},
		{tag: "GB", substringPatterns: []string{"英国", "UNITED KINGDOM"}, tokenPatterns: []string{"GB", "UK"}},
		{tag: "NL", substringPatterns: []string{"荷兰", "NETHERLANDS"}, tokenPatterns: []string{"NL"}},
		{tag: "RU", substringPatterns: []string{"俄罗斯", "RUSSIA"}, tokenPatterns: []string{"RU"}},
		{tag: "FR", substringPatterns: []string{"法国", "FRANCE"}, tokenPatterns: []string{"FR"}},
		{tag: "CA", substringPatterns: []string{"加拿大", "CANADA"}, tokenPatterns: []string{"CA"}},
		{tag: "AU", substringPatterns: []string{"澳大利亚", "AUSTRALIA"}, tokenPatterns: []string{"AU"}},
	}

	upperName := strings.ToUpper(name)
	tokens := tokenizeName(upperName)
	var tags []string
	for _, matcher := range matchers {
		for _, pattern := range matcher.substringPatterns {
			if strings.Contains(upperName, strings.ToUpper(pattern)) {
				tags = append(tags, matcher.tag)
				break
			}
		}
		if len(tags) > 0 && tags[len(tags)-1] == matcher.tag {
			continue
		}
		for _, pattern := range matcher.tokenPatterns {
			if tokenSetContains(tokens, pattern) {
				tags = append(tags, matcher.tag)
				break
			}
		}
	}

	return cleanStringSlice(tags)
}

func tokenizeName(value string) []string {
	var (
		tokens []string
		buf    strings.Builder
	)

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		tokens = append(tokens, buf.String())
		buf.Reset()
	}

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return tokens
}

func tokenSetContains(tokens []string, target string) bool {
	target = strings.ToUpper(strings.TrimSpace(target))
	for _, token := range tokens {
		if token == target {
			return true
		}
	}
	return false
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
