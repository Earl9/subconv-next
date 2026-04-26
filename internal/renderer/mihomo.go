package renderer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"subconv-next/internal/model"
)

const probeURL = "https://www.gstatic.com/generate_204"

var regionGroups = []struct {
	Tag  string
	Name string
}{
	{Tag: "HK", Name: "香港"},
	{Tag: "JP", Name: "日本"},
	{Tag: "US", Name: "美国"},
	{Tag: "SG", Name: "新加坡"},
	{Tag: "TW", Name: "台湾"},
	{Tag: "KR", Name: "韩国"},
	{Tag: "DE", Name: "德国"},
	{Tag: "GB", Name: "英国"},
	{Tag: "FR", Name: "法国"},
	{Tag: "CA", Name: "加拿大"},
	{Tag: "AU", Name: "澳大利亚"},
}

type mihomoConfig struct {
	MixedPort   int                `yaml:"mixed-port"`
	AllowLAN    bool               `yaml:"allow-lan"`
	Mode        string             `yaml:"mode"`
	LogLevel    string             `yaml:"log-level"`
	IPv6        bool               `yaml:"ipv6"`
	DNS         *mihomoDNS         `yaml:"dns,omitempty"`
	Proxies     []mihomoProxy      `yaml:"proxies"`
	ProxyGroups []mihomoProxyGroup `yaml:"proxy-groups"`
	Rules       []string           `yaml:"rules"`
}

type mihomoDNS struct {
	Enable       bool     `yaml:"enable"`
	IPv6         bool     `yaml:"ipv6"`
	EnhancedMode string   `yaml:"enhanced-mode,omitempty"`
	Nameserver   []string `yaml:"nameserver,omitempty"`
}

type mihomoProxy struct {
	Name                     string                 `yaml:"name"`
	Type                     string                 `yaml:"type"`
	Server                   string                 `yaml:"server,omitempty"`
	Port                     int                    `yaml:"port,omitempty"`
	Cipher                   string                 `yaml:"cipher,omitempty"`
	Password                 string                 `yaml:"password,omitempty"`
	Username                 string                 `yaml:"username,omitempty"`
	UUID                     string                 `yaml:"uuid,omitempty"`
	AlterID                  interface{}            `yaml:"alterId,omitempty"`
	Network                  string                 `yaml:"network,omitempty"`
	TLS                      bool                   `yaml:"tls,omitempty"`
	UDP                      *bool                  `yaml:"udp,omitempty"`
	ServerName               string                 `yaml:"servername,omitempty"`
	SNI                      string                 `yaml:"sni,omitempty"`
	ClientFingerprint        string                 `yaml:"client-fingerprint,omitempty"`
	SkipCertVerify           *bool                  `yaml:"skip-cert-verify,omitempty"`
	ALPN                     []string               `yaml:"alpn,omitempty"`
	Flow                     string                 `yaml:"flow,omitempty"`
	RealityOpts              *mihomoRealityOpts     `yaml:"reality-opts,omitempty"`
	WSOpts                   *mihomoWSOpts          `yaml:"ws-opts,omitempty"`
	GrpcOpts                 *mihomoGRPCOpts        `yaml:"grpc-opts,omitempty"`
	Obfs                     string                 `yaml:"obfs,omitempty"`
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
	Peers                    []mihomoWGPeer         `yaml:"peers,omitempty"`
	AmneziaWGOption          map[string]interface{} `yaml:"amnezia-wg-option,omitempty"`
}

type mihomoRealityOpts struct {
	PublicKey string `yaml:"public-key,omitempty"`
	ShortID   string `yaml:"short-id,omitempty"`
	SpiderX   string `yaml:"spider-x,omitempty"`
}

type mihomoWSOpts struct {
	Path    string            `yaml:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type mihomoGRPCOpts struct {
	GrpcServiceName string `yaml:"grpc-service-name,omitempty"`
}

type mihomoWGPeer struct {
	Server       string      `yaml:"server,omitempty"`
	Port         int         `yaml:"port,omitempty"`
	PublicKey    string      `yaml:"public-key,omitempty"`
	PreSharedKey string      `yaml:"pre-shared-key,omitempty"`
	AllowedIPs   []string    `yaml:"allowed-ips,omitempty"`
	Reserved     interface{} `yaml:"reserved,omitempty"`
}

type mihomoProxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

func OptionsFromConfig(cfg model.Config) model.RenderOptions {
	opts := model.DefaultRenderOptions()
	opts.Template = cfg.Service.Template
	opts.MixedPort = cfg.Render.MixedPort
	opts.AllowLAN = cfg.Render.AllowLAN
	opts.Mode = cfg.Render.Mode
	opts.LogLevel = cfg.Render.LogLevel
	opts.IPv6 = cfg.Render.IPv6
	opts.DNSEnabled = cfg.Render.DNSEnabled
	opts.EnhancedMode = cfg.Render.EnhancedMode

	return NormalizeRenderOptions(opts)
}

func NormalizeRenderOptions(opts model.RenderOptions) model.RenderOptions {
	defaults := model.DefaultRenderOptions()

	if strings.TrimSpace(opts.Template) == "" {
		opts.Template = defaults.Template
	}
	if opts.MixedPort == 0 {
		opts.MixedPort = defaults.MixedPort
	}
	if strings.TrimSpace(opts.Mode) == "" {
		opts.Mode = defaults.Mode
	}
	if strings.TrimSpace(opts.LogLevel) == "" {
		opts.LogLevel = defaults.LogLevel
	}
	if strings.TrimSpace(opts.EnhancedMode) == "" {
		opts.EnhancedMode = defaults.EnhancedMode
	}

	return opts
}

func RenderMihomo(nodes []model.NodeIR, opts model.RenderOptions) ([]byte, error) {
	opts = NormalizeRenderOptions(opts)
	nodes = model.NormalizeNodes(nodes)
	nodes = ensureUniqueNames(nodes)

	proxies, err := buildProxies(nodes)
	if err != nil {
		return nil, err
	}

	cfg := mihomoConfig{
		MixedPort:   opts.MixedPort,
		AllowLAN:    opts.AllowLAN,
		Mode:        opts.Mode,
		LogLevel:    opts.LogLevel,
		IPv6:        opts.IPv6,
		Proxies:     proxies,
		ProxyGroups: buildProxyGroups(nodes, opts.Template),
		Rules:       buildRules(opts.Template),
	}
	if opts.DNSEnabled {
		cfg.DNS = &mihomoDNS{
			Enable:       true,
			IPv6:         opts.IPv6,
			EnhancedMode: opts.EnhancedMode,
			Nameserver: []string{
				"https://1.1.1.1/dns-query",
				"https://8.8.8.8/dns-query",
			},
		}
	}

	return renderConfig(cfg), nil
}

func buildProxies(nodes []model.NodeIR) ([]mihomoProxy, error) {
	proxies := make([]mihomoProxy, 0, len(nodes))
	for _, node := range nodes {
		proxy, err := buildProxy(node)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, proxy)
	}
	return proxies, nil
}

func buildProxy(node model.NodeIR) (mihomoProxy, error) {
	proxy := mihomoProxy{
		Name:   node.Name,
		Type:   string(node.Type),
		Server: node.Server,
		Port:   node.Port,
		UDP:    udpOrDefault(node),
	}

	switch node.Type {
	case model.ProtocolSS:
		proxy.Cipher = rawString(node.Raw, "method")
		proxy.Password = node.Auth.Password
	case model.ProtocolVMess:
		proxy.UUID = node.Auth.UUID
		proxy.Network = node.Transport.Network
		proxy.TLS = node.TLS.Enabled
		proxy.ServerName = node.TLS.SNI
		proxy.ClientFingerprint = node.TLS.ClientFingerprint
		proxy.AlterID = rawScalar(node.Raw, "alterId")
		applyTransport(&proxy, node.Transport)
	case model.ProtocolVLESS:
		proxy.UUID = node.Auth.UUID
		proxy.Network = node.Transport.Network
		proxy.TLS = node.TLS.Enabled
		proxy.ServerName = node.TLS.SNI
		proxy.ClientFingerprint = node.TLS.ClientFingerprint
		proxy.Flow = rawString(node.Raw, "flow")
		if node.TLS.Reality != nil {
			proxy.RealityOpts = &mihomoRealityOpts{
				PublicKey: node.TLS.Reality.PublicKey,
				ShortID:   node.TLS.Reality.ShortID,
				SpiderX:   node.TLS.Reality.SpiderX,
			}
		}
		applyTransport(&proxy, node.Transport)
	case model.ProtocolTrojan:
		proxy.Password = node.Auth.Password
		proxy.Network = node.Transport.Network
		proxy.TLS = true
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = model.Bool(node.TLS.Insecure)
		applyTransport(&proxy, node.Transport)
	case model.ProtocolHysteria2:
		proxy.Password = node.Auth.Password
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = model.Bool(node.TLS.Insecure)
		proxy.Obfs = rawString(node.Raw, "obfs")
		proxy.ObfsPassword = rawString(node.Raw, "obfsPassword")
	case model.ProtocolTUIC:
		proxy.UUID = node.Auth.UUID
		proxy.Password = node.Auth.Password
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = model.Bool(node.TLS.Insecure)
		proxy.CongestionController = rawString(node.Raw, "congestionController")
		proxy.UDPRelayMode = rawString(node.Raw, "udpRelayMode")
		proxy.ReduceRTT = rawBoolPointer(node.Raw, "reduceRTT")
	case model.ProtocolAnyTLS:
		proxy.Password = node.Auth.Password
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.ClientFingerprint = node.TLS.ClientFingerprint
		proxy.SkipCertVerify = model.Bool(node.TLS.Insecure)
		proxy.IdleSessionCheckInterval = rawScalar(node.Raw, "idleSessionCheckInterval")
		proxy.IdleSessionTimeout = rawScalar(node.Raw, "idleSessionTimeout")
		proxy.MinIdleSession = rawScalar(node.Raw, "minIdleSession")
	case model.ProtocolWireGuard:
		if node.WireGuard == nil {
			return mihomoProxy{}, fmt.Errorf("wireguard node %q missing wireguard settings", node.Name)
		}
		proxy.IP = node.WireGuard.IP
		proxy.IPv6 = node.WireGuard.IPv6
		proxy.PrivateKey = node.Auth.PrivateKey
		proxy.PublicKey = node.Auth.PublicKey
		proxy.AllowedIPs = allowedIPsOrDefault(node.WireGuard.AllowedIPs)
		proxy.PreSharedKey = node.Auth.PreSharedKey
		proxy.Reserved = reservedValue(node.WireGuard)
		proxy.PersistentKeepalive = node.WireGuard.PersistentKeepalive
		proxy.MTU = node.WireGuard.MTU
		proxy.RemoteDNSResolve = boolPointerIfTrue(node.WireGuard.RemoteDNSResolve)
		proxy.DNS = append([]string(nil), node.WireGuard.DNS...)
		if len(node.WireGuard.Peers) > 0 {
			proxy.Server = ""
			proxy.Port = 0
			proxy.PublicKey = ""
			proxy.PreSharedKey = ""
			proxy.AllowedIPs = nil
			proxy.Peers = buildWGPeers(node.WireGuard.Peers)
		}
		if len(node.WireGuard.AmneziaWG) > 0 {
			proxy.AmneziaWGOption = cloneStringMap(node.WireGuard.AmneziaWG)
		}
	default:
		return mihomoProxy{}, fmt.Errorf("unsupported renderer protocol %q", node.Type)
	}

	return proxy, nil
}

func applyTransport(proxy *mihomoProxy, transport model.TransportOptions) {
	proxy.Network = transport.Network

	switch transport.Network {
	case "ws":
		wsOpts := &mihomoWSOpts{
			Path: transport.Path,
		}
		if transport.Host != "" {
			wsOpts.Headers = map[string]string{"Host": transport.Host}
		}
		if wsOpts.Path != "" || len(wsOpts.Headers) > 0 {
			proxy.WSOpts = wsOpts
		}
	case "grpc":
		if transport.ServiceName != "" {
			proxy.GrpcOpts = &mihomoGRPCOpts{
				GrpcServiceName: transport.ServiceName,
			}
		}
	}
}

func buildWGPeers(peers []model.WGPeer) []mihomoWGPeer {
	out := make([]mihomoWGPeer, 0, len(peers))
	for _, peer := range peers {
		out = append(out, mihomoWGPeer{
			Server:       peer.Server,
			Port:         peer.Port,
			PublicKey:    peer.PublicKey,
			PreSharedKey: peer.PreSharedKey,
			AllowedIPs:   allowedIPsOrDefault(peer.AllowedIPs),
			Reserved:     reservedSlice(peer.Reserved),
		})
	}
	return out
}

func buildProxyGroups(nodes []model.NodeIR, template string) []mihomoProxyGroup {
	template = strings.ToLower(strings.TrimSpace(template))
	allNames := nodeNames(nodes)
	if len(allNames) == 0 {
		return []mihomoProxyGroup{
			{
				Name:    "节点选择",
				Type:    "select",
				Proxies: []string{"DIRECT"},
			},
		}
	}

	groupNames := map[string]struct{}{}
	var groups []mihomoProxyGroup

	addGroup := func(group mihomoProxyGroup) {
		group.Proxies = uniqueOrdered(group.Proxies)
		if len(group.Proxies) == 0 {
			return
		}
		groups = append(groups, group)
		groupNames[group.Name] = struct{}{}
	}

	hasAuto := true
	hasFallback := template != "lite"
	if hasAuto {
		addGroup(mihomoProxyGroup{
			Name:     "自动选择",
			Type:     "url-test",
			Proxies:  allNames,
			URL:      probeURL,
			Interval: 300,
		})
	}
	if hasFallback {
		addGroup(mihomoProxyGroup{
			Name:     "故障转移",
			Type:     "fallback",
			Proxies:  allNames,
			URL:      probeURL,
			Interval: 300,
		})
	}

	regionNameOrder := regionProxyNames(nodes)
	for _, region := range regionGroups {
		names := regionNameOrder[region.Tag]
		if len(names) == 0 {
			continue
		}
		regionProxies := []string{"自动选择", "DIRECT"}
		if hasFallback {
			regionProxies = []string{"自动选择", "故障转移", "DIRECT"}
		}
		addGroup(mihomoProxyGroup{
			Name:    region.Name,
			Type:    "select",
			Proxies: append(regionProxies, names...),
		})
	}

	selectProxies := []string{}
	if hasAuto {
		selectProxies = append(selectProxies, "自动选择")
	}
	if hasFallback {
		selectProxies = append(selectProxies, "故障转移")
	}
	selectProxies = append(selectProxies, "DIRECT")
	if template == "lite" {
		selectProxies = append(selectProxies, "REJECT")
	}
	for _, region := range regionGroups {
		if _, ok := groupNames[region.Name]; ok {
			selectProxies = append(selectProxies, region.Name)
		}
	}
	selectProxies = append(selectProxies, allNames...)

	addGroup(mihomoProxyGroup{
		Name:    "节点选择",
		Type:    "select",
		Proxies: selectProxies,
	})

	if template != "lite" {
		serviceGroupProxies := []string{"节点选择", "自动选择", "故障转移", "DIRECT"}
		for _, region := range regionGroups {
			if _, ok := groupNames[region.Name]; ok {
				serviceGroupProxies = append(serviceGroupProxies, region.Name)
			}
		}

		addGroup(mihomoProxyGroup{Name: "AI", Type: "select", Proxies: serviceGroupProxies})
		addGroup(mihomoProxyGroup{Name: "流媒体", Type: "select", Proxies: serviceGroupProxies})
		addGroup(mihomoProxyGroup{Name: "Telegram", Type: "select", Proxies: serviceGroupProxies})
		addGroup(mihomoProxyGroup{Name: "GitHub", Type: "select", Proxies: serviceGroupProxies})
		addGroup(mihomoProxyGroup{Name: "Microsoft", Type: "select", Proxies: serviceGroupProxies})
		addGroup(mihomoProxyGroup{Name: "Apple", Type: "select", Proxies: serviceGroupProxies})
		addGroup(mihomoProxyGroup{Name: "国内直连", Type: "select", Proxies: []string{"DIRECT", "节点选择"}})
		addGroup(mihomoProxyGroup{Name: "漏网之鱼", Type: "select", Proxies: []string{"节点选择", "自动选择", "DIRECT"}})
	}

	return groups
}

func regionProxyNames(nodes []model.NodeIR) map[string][]string {
	regionNames := make(map[string][]string)
	for _, node := range nodes {
		for _, tag := range node.Tags {
			regionNames[tag] = append(regionNames[tag], node.Name)
		}
	}

	for key, names := range regionNames {
		regionNames[key] = uniqueOrdered(names)
	}
	return regionNames
}

func buildRules(template string) []string {
	template = strings.ToLower(strings.TrimSpace(template))
	if template == "lite" {
		return []string{
			"GEOSITE,private,DIRECT",
			"GEOIP,private,DIRECT,no-resolve",
			"GEOSITE,cn,DIRECT",
			"GEOIP,CN,DIRECT",
			"MATCH,节点选择",
		}
	}

	return []string{
		"GEOSITE,telegram,Telegram",
		"GEOSITE,github,GitHub",
		"GEOSITE,microsoft,Microsoft",
		"GEOSITE,apple,Apple",
		"GEOSITE,openai,AI",
		"GEOSITE,netflix,流媒体",
		"GEOSITE,private,DIRECT",
		"GEOIP,private,DIRECT,no-resolve",
		"GEOSITE,cn,国内直连",
		"GEOIP,CN,国内直连",
		"MATCH,漏网之鱼",
	}
}

func ensureUniqueNames(nodes []model.NodeIR) []model.NodeIR {
	out := make([]model.NodeIR, len(nodes))
	copy(out, nodes)

	counts := make(map[string]int, len(nodes))
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
		out[i].Name = fmt.Sprintf("%s %d", base, counts[base])
	}
	return out
}

func nodeNames(nodes []model.NodeIR) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}

func uniqueOrdered(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func rawString(raw map[string]interface{}, key string) string {
	if len(raw) == 0 {
		return ""
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func rawScalar(raw map[string]interface{}, key string) interface{} {
	value := rawString(raw, key)
	if value == "" {
		return nil
	}
	if n, err := strconv.Atoi(value); err == nil {
		return n
	}
	if b, ok := parseBool(value); ok {
		return b
	}
	return value
}

func rawBoolPointer(raw map[string]interface{}, key string) *bool {
	if len(raw) == 0 {
		return nil
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case bool:
		return model.Bool(typed)
	case string:
		if parsed, ok := parseBool(typed); ok {
			return model.Bool(parsed)
		}
	}
	return nil
}

func parseBool(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func udpOrDefault(node model.NodeIR) *bool {
	if node.UDP != nil {
		return node.UDP
	}
	switch node.Type {
	case model.ProtocolSS, model.ProtocolVMess, model.ProtocolVLESS, model.ProtocolTrojan, model.ProtocolHysteria2, model.ProtocolTUIC, model.ProtocolAnyTLS, model.ProtocolWireGuard:
		return model.Bool(true)
	default:
		return nil
	}
}

func boolPointerIfTrue(v bool) *bool {
	if !v {
		return nil
	}
	return model.Bool(true)
}

func allowedIPsOrDefault(values []string) []string {
	if len(values) == 0 {
		return []string{"0.0.0.0/0"}
	}
	return append([]string(nil), values...)
}

func reservedValue(wg *model.WireGuardOptions) interface{} {
	if wg == nil {
		return nil
	}
	if len(wg.Reserved) > 0 {
		return reservedSlice(wg.Reserved)
	}
	if strings.TrimSpace(wg.ReservedString) != "" {
		return wg.ReservedString
	}
	return nil
}

func reservedSlice(values []int) interface{} {
	if len(values) == 0 {
		return nil
	}
	out := append([]int(nil), values...)
	return out
}

func cloneStringMap(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	cloned := make(map[string]interface{}, len(values))
	for _, key := range keys {
		cloned[key] = values[key]
	}
	return cloned
}

func renderConfig(cfg mihomoConfig) []byte {
	var w yamlWriter

	w.scalar(0, "mixed-port", cfg.MixedPort)
	w.scalar(0, "allow-lan", cfg.AllowLAN)
	w.scalar(0, "mode", cfg.Mode)
	w.scalar(0, "log-level", cfg.LogLevel)
	w.scalar(0, "ipv6", cfg.IPv6)
	if cfg.DNS != nil {
		w.line(0, "dns:")
		renderDNS(&w, 4, *cfg.DNS)
	}

	w.line(0, "proxies:")
	for _, proxy := range cfg.Proxies {
		renderProxy(&w, 4, proxy)
	}

	w.line(0, "proxy-groups:")
	for _, group := range cfg.ProxyGroups {
		renderProxyGroup(&w, 4, group)
	}

	w.line(0, "rules:")
	for _, rule := range cfg.Rules {
		w.listScalar(4, rule)
	}

	return []byte(strings.TrimRight(w.String(), "\n"))
}

func renderDNS(w *yamlWriter, indent int, dns mihomoDNS) {
	w.scalar(indent, "enable", dns.Enable)
	w.scalar(indent, "ipv6", dns.IPv6)
	if dns.EnhancedMode != "" {
		w.scalar(indent, "enhanced-mode", dns.EnhancedMode)
	}
	if len(dns.Nameserver) > 0 {
		w.line(indent, "nameserver:")
		for _, item := range dns.Nameserver {
			w.listScalar(indent+4, item)
		}
	}
}

func renderProxy(w *yamlWriter, indent int, proxy mihomoProxy) {
	w.itemScalar(indent, "name", proxy.Name)
	w.scalar(indent+2, "type", proxy.Type)
	if proxy.Server != "" {
		w.scalar(indent+2, "server", proxy.Server)
	}
	if proxy.Port != 0 {
		w.scalar(indent+2, "port", proxy.Port)
	}
	if proxy.Cipher != "" {
		w.scalar(indent+2, "cipher", proxy.Cipher)
	}
	if proxy.Password != "" {
		w.scalar(indent+2, "password", proxy.Password)
	}
	if proxy.Username != "" {
		w.scalar(indent+2, "username", proxy.Username)
	}
	if proxy.UUID != "" {
		w.scalar(indent+2, "uuid", proxy.UUID)
	}
	if proxy.AlterID != nil {
		w.scalar(indent+2, "alterId", proxy.AlterID)
	}
	if proxy.Network != "" {
		w.scalar(indent+2, "network", proxy.Network)
	}
	if proxy.TLS {
		w.scalar(indent+2, "tls", proxy.TLS)
	}
	if proxy.UDP != nil {
		w.scalar(indent+2, "udp", *proxy.UDP)
	}
	if proxy.ServerName != "" {
		w.scalar(indent+2, "servername", proxy.ServerName)
	}
	if proxy.SNI != "" {
		w.scalar(indent+2, "sni", proxy.SNI)
	}
	if proxy.ClientFingerprint != "" {
		w.scalar(indent+2, "client-fingerprint", proxy.ClientFingerprint)
	}
	if proxy.SkipCertVerify != nil {
		w.scalar(indent+2, "skip-cert-verify", *proxy.SkipCertVerify)
	}
	if len(proxy.ALPN) > 0 {
		w.list(indent+2, "alpn", proxy.ALPN)
	}
	if proxy.Flow != "" {
		w.scalar(indent+2, "flow", proxy.Flow)
	}
	if proxy.RealityOpts != nil {
		w.line(indent+2, "reality-opts:")
		if proxy.RealityOpts.PublicKey != "" {
			w.scalar(indent+4, "public-key", proxy.RealityOpts.PublicKey)
		}
		if proxy.RealityOpts.ShortID != "" {
			w.scalar(indent+4, "short-id", proxy.RealityOpts.ShortID)
		}
		if proxy.RealityOpts.SpiderX != "" {
			w.scalar(indent+4, "spider-x", proxy.RealityOpts.SpiderX)
		}
	}
	if proxy.WSOpts != nil {
		w.line(indent+2, "ws-opts:")
		if proxy.WSOpts.Path != "" {
			w.scalar(indent+4, "path", proxy.WSOpts.Path)
		}
		if len(proxy.WSOpts.Headers) > 0 {
			w.line(indent+4, "headers:")
			keys := sortedKeysString(proxy.WSOpts.Headers)
			for _, key := range keys {
				w.scalar(indent+6, key, proxy.WSOpts.Headers[key])
			}
		}
	}
	if proxy.GrpcOpts != nil && proxy.GrpcOpts.GrpcServiceName != "" {
		w.line(indent+2, "grpc-opts:")
		w.scalar(indent+4, "grpc-service-name", proxy.GrpcOpts.GrpcServiceName)
	}
	if proxy.Obfs != "" {
		w.scalar(indent+2, "obfs", proxy.Obfs)
	}
	if proxy.ObfsPassword != "" {
		w.scalar(indent+2, "obfs-password", proxy.ObfsPassword)
	}
	if proxy.CongestionController != "" {
		w.scalar(indent+2, "congestion-controller", proxy.CongestionController)
	}
	if proxy.UDPRelayMode != "" {
		w.scalar(indent+2, "udp-relay-mode", proxy.UDPRelayMode)
	}
	if proxy.ReduceRTT != nil {
		w.scalar(indent+2, "reduce-rtt", *proxy.ReduceRTT)
	}
	if proxy.IdleSessionCheckInterval != nil {
		w.scalar(indent+2, "idle-session-check-interval", proxy.IdleSessionCheckInterval)
	}
	if proxy.IdleSessionTimeout != nil {
		w.scalar(indent+2, "idle-session-timeout", proxy.IdleSessionTimeout)
	}
	if proxy.MinIdleSession != nil {
		w.scalar(indent+2, "min-idle-session", proxy.MinIdleSession)
	}
	if proxy.IP != "" {
		w.scalar(indent+2, "ip", proxy.IP)
	}
	if proxy.IPv6 != "" {
		w.scalar(indent+2, "ipv6", proxy.IPv6)
	}
	if proxy.PrivateKey != "" {
		w.scalar(indent+2, "private-key", proxy.PrivateKey)
	}
	if proxy.PublicKey != "" {
		w.scalar(indent+2, "public-key", proxy.PublicKey)
	}
	if len(proxy.AllowedIPs) > 0 {
		w.list(indent+2, "allowed-ips", proxy.AllowedIPs)
	}
	if proxy.PreSharedKey != "" {
		w.scalar(indent+2, "pre-shared-key", proxy.PreSharedKey)
	}
	if proxy.Reserved != nil {
		renderInterfaceField(w, indent+2, "reserved", proxy.Reserved)
	}
	if proxy.PersistentKeepalive != 0 {
		w.scalar(indent+2, "persistent-keepalive", proxy.PersistentKeepalive)
	}
	if proxy.MTU != 0 {
		w.scalar(indent+2, "mtu", proxy.MTU)
	}
	if proxy.RemoteDNSResolve != nil {
		w.scalar(indent+2, "remote-dns-resolve", *proxy.RemoteDNSResolve)
	}
	if len(proxy.DNS) > 0 {
		w.list(indent+2, "dns", proxy.DNS)
	}
	if len(proxy.Peers) > 0 {
		w.line(indent+2, "peers:")
		for _, peer := range proxy.Peers {
			w.itemScalar(indent+4, "server", peer.Server)
			if peer.Port != 0 {
				w.scalar(indent+6, "port", peer.Port)
			}
			if peer.PublicKey != "" {
				w.scalar(indent+6, "public-key", peer.PublicKey)
			}
			if peer.PreSharedKey != "" {
				w.scalar(indent+6, "pre-shared-key", peer.PreSharedKey)
			}
			if len(peer.AllowedIPs) > 0 {
				w.list(indent+6, "allowed-ips", peer.AllowedIPs)
			}
			if peer.Reserved != nil {
				renderInterfaceField(w, indent+6, "reserved", peer.Reserved)
			}
		}
	}
	if len(proxy.AmneziaWGOption) > 0 {
		w.line(indent+2, "amnezia-wg-option:")
		renderMapStringAny(w, indent+4, proxy.AmneziaWGOption)
	}
}

func renderProxyGroup(w *yamlWriter, indent int, group mihomoProxyGroup) {
	w.itemScalar(indent, "name", group.Name)
	w.scalar(indent+2, "type", group.Type)
	w.list(indent+2, "proxies", group.Proxies)
	if group.URL != "" {
		w.scalar(indent+2, "url", group.URL)
	}
	if group.Interval != 0 {
		w.scalar(indent+2, "interval", group.Interval)
	}
}

func renderInterfaceField(w *yamlWriter, indent int, key string, value interface{}) {
	switch typed := value.(type) {
	case []string:
		w.list(indent, key, typed)
	case []int:
		w.line(indent, key+":")
		for _, item := range typed {
			w.listScalar(indent+2, item)
		}
	case map[string]interface{}:
		w.line(indent, key+":")
		renderMapStringAny(w, indent+2, typed)
	default:
		w.scalar(indent, key, typed)
	}
}

func renderMapStringAny(w *yamlWriter, indent int, values map[string]interface{}) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := values[key]
		switch typed := value.(type) {
		case []string:
			w.list(indent, key, typed)
		case []int:
			w.line(indent, key+":")
			for _, item := range typed {
				w.listScalar(indent+2, item)
			}
		case map[string]interface{}:
			w.line(indent, key+":")
			renderMapStringAny(w, indent+2, typed)
		default:
			w.scalar(indent, key, typed)
		}
	}
}

func sortedKeysString(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type yamlWriter struct {
	builder strings.Builder
}

func (w *yamlWriter) line(indent int, value string) {
	w.builder.WriteString(strings.Repeat(" ", indent))
	w.builder.WriteString(value)
	w.builder.WriteByte('\n')
}

func (w *yamlWriter) scalar(indent int, key string, value interface{}) {
	w.line(indent, fmt.Sprintf("%s: %s", key, yamlScalar(value)))
}

func (w *yamlWriter) itemScalar(indent int, key string, value interface{}) {
	w.line(indent, fmt.Sprintf("- %s: %s", key, yamlScalar(value)))
}

func (w *yamlWriter) list(indent int, key string, values []string) {
	w.line(indent, key+":")
	for _, value := range values {
		w.listScalar(indent+2, value)
	}
}

func (w *yamlWriter) listScalar(indent int, value interface{}) {
	w.line(indent, fmt.Sprintf("- %s", yamlScalar(value)))
}

func (w *yamlWriter) String() string {
	return w.builder.String()
}

func yamlScalar(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return yamlString(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case fmt.Stringer:
		return yamlString(typed.String())
	default:
		return yamlString(fmt.Sprint(value))
	}
}

func yamlString(value string) string {
	if value == "" {
		return `""`
	}
	if needsQuotedYAML(value) {
		escaped := strings.ReplaceAll(value, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return value
}

func needsQuotedYAML(value string) bool {
	if value == "" {
		return true
	}
	if strings.TrimSpace(value) != value {
		return true
	}
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") || strings.Contains(value, "\t") {
		return true
	}
	if strings.Contains(value, ": ") || strings.Contains(value, " #") {
		return true
	}
	switch value[0] {
	case '-', '?', '!', '*', '&', '{', '}', '[', ']', '#', '|', '>', '@', '`', '"', '\'':
		return true
	}
	lower := strings.ToLower(value)
	switch lower {
	case "true", "false", "null", "~", "yes", "no", "on", "off":
		return true
	}
	return false
}
