package renderer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"subconv-next/internal/model"
)

const probeURL = "https://www.gstatic.com/generate_204"

var defaultDNSServers = []string{
	"https://1.1.1.1/dns-query",
	"https://8.8.8.8/dns-query",
}

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
	MixedPort               int                           `yaml:"mixed-port"`
	AllowLAN                bool                          `yaml:"allow-lan"`
	Mode                    string                        `yaml:"mode"`
	LogLevel                string                        `yaml:"log-level"`
	IPv6                    bool                          `yaml:"ipv6"`
	UnifiedDelay            bool                          `yaml:"unified-delay,omitempty"`
	TCPConcurrent           bool                          `yaml:"tcp-concurrent,omitempty"`
	FindProcessMode         string                        `yaml:"find-process-mode,omitempty"`
	GlobalClientFingerprint string                        `yaml:"global-client-fingerprint,omitempty"`
	DNS                     *mihomoDNS                    `yaml:"dns,omitempty"`
	Profile                 *mihomoProfile                `yaml:"profile,omitempty"`
	Sniffer                 *mihomoSniffer                `yaml:"sniffer,omitempty"`
	Proxies                 []mihomoProxy                 `yaml:"proxies"`
	ProxyGroups             []mihomoProxyGroup            `yaml:"proxy-groups"`
	RuleProviders           map[string]mihomoRuleProvider `yaml:"rule-providers,omitempty"`
	Rules                   []string                      `yaml:"rules"`
}

type mihomoDNS struct {
	Enable            bool                     `yaml:"enable"`
	Listen            string                   `yaml:"listen,omitempty"`
	UseSystemHosts    *bool                    `yaml:"use-system-hosts,omitempty"`
	IPv6              bool                     `yaml:"ipv6"`
	EnhancedMode      string                   `yaml:"enhanced-mode,omitempty"`
	FakeIPRange       string                   `yaml:"fake-ip-range,omitempty"`
	DefaultNameserver []string                 `yaml:"default-nameserver,omitempty"`
	Nameserver        []string                 `yaml:"nameserver,omitempty"`
	Fallback          []string                 `yaml:"fallback,omitempty"`
	FallbackFilter    *mihomoDNSFallbackFilter `yaml:"fallback-filter,omitempty"`
	FakeIPFilter      []string                 `yaml:"fake-ip-filter,omitempty"`
	NameserverPolicy  map[string][]string      `yaml:"nameserver-policy,omitempty"`
}

type mihomoDNSFallbackFilter struct {
	GeoIP  bool     `yaml:"geoip"`
	IPCIDR []string `yaml:"ipcidr,omitempty"`
	Domain []string `yaml:"domain,omitempty"`
}

type mihomoProfile struct {
	StoreSelected bool `yaml:"store-selected,omitempty"`
	StoreFakeIP   bool `yaml:"store-fake-ip,omitempty"`
}

type mihomoSniffer struct {
	Enable      bool                       `yaml:"enable"`
	ParsePureIP bool                       `yaml:"parse-pure-ip,omitempty"`
	Sniff       map[string]mihomoSniffRule `yaml:"sniff,omitempty"`
}

type mihomoSniffRule struct {
	Ports               []string `yaml:"ports,omitempty"`
	OverrideDestination bool     `yaml:"override-destination,omitempty"`
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

type mihomoRuleProvider struct {
	Type      string              `yaml:"type"`
	URL       string              `yaml:"url,omitempty"`
	Path      string              `yaml:"path,omitempty"`
	Interval  int                 `yaml:"interval,omitempty"`
	Proxy     string              `yaml:"proxy,omitempty"`
	Behavior  string              `yaml:"behavior"`
	Format    string              `yaml:"format,omitempty"`
	SizeLimit int64               `yaml:"size-limit,omitempty"`
	Header    map[string][]string `yaml:"header,omitempty"`
	Payload   []string            `yaml:"payload,omitempty"`
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
	opts.UnifiedDelay = cfg.Render.UnifiedDelay
	opts.TCPConcurrent = cfg.Render.TCPConcurrent
	opts.FindProcessMode = cfg.Render.FindProcessMode
	opts.GlobalClientFingerprint = cfg.Render.GlobalClientFingerprint
	opts.DNS = cfg.Render.DNS
	opts.Profile = cfg.Render.Profile
	opts.Sniffer = cfg.Render.Sniffer
	opts.FinalPolicy = cfg.Render.FinalPolicy
	opts.AdditionalRules = append([]string(nil), cfg.Render.AdditionalRules...)
	opts.RuleProviders = append([]model.RuleProviderConfig(nil), cfg.Render.RuleProviders...)
	opts.CustomProxyGroups = append([]model.CustomProxyGroupConfig(nil), cfg.Render.CustomProxyGroups...)

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
		MixedPort:               opts.MixedPort,
		AllowLAN:                opts.AllowLAN,
		Mode:                    opts.Mode,
		LogLevel:                opts.LogLevel,
		IPv6:                    opts.IPv6,
		UnifiedDelay:            opts.UnifiedDelay,
		TCPConcurrent:           opts.TCPConcurrent,
		FindProcessMode:         opts.FindProcessMode,
		GlobalClientFingerprint: opts.GlobalClientFingerprint,
		Proxies:                 proxies,
		ProxyGroups:             buildProxyGroups(nodes, opts.Template, opts.CustomProxyGroups),
		RuleProviders:           buildRuleProviders(opts.RuleProviders),
		Rules:                   buildRules(opts.Template, opts.FinalPolicy, opts.AdditionalRules, opts.RuleProviders),
	}
	cfg.DNS = buildDNSConfig(opts)
	cfg.Profile = buildProfileConfig(opts.Profile)
	cfg.Sniffer = buildSnifferConfig(opts.Sniffer)

	return renderConfig(cfg), nil
}

func buildDNSConfig(opts model.RenderOptions) *mihomoDNS {
	if opts.DNS == nil {
		if !opts.DNSEnabled {
			return nil
		}
		return &mihomoDNS{
			Enable:       true,
			IPv6:         opts.IPv6,
			EnhancedMode: strings.TrimSpace(opts.EnhancedMode),
			Nameserver:   append([]string(nil), defaultDNSServers...),
		}
	}

	cfg := opts.DNS
	dns := &mihomoDNS{
		Enable:            cfg.Enable,
		Listen:            strings.TrimSpace(cfg.Listen),
		UseSystemHosts:    model.Bool(cfg.UseSystemHosts),
		IPv6:              opts.IPv6,
		EnhancedMode:      strings.TrimSpace(cfg.EnhancedMode),
		FakeIPRange:       strings.TrimSpace(cfg.FakeIPRange),
		DefaultNameserver: cloneStrings(cfg.DefaultNameserver),
		Nameserver:        cloneStrings(cfg.Nameserver),
		Fallback:          cloneStrings(cfg.Fallback),
		FakeIPFilter:      cloneStrings(cfg.FakeIPFilter),
		NameserverPolicy:  cloneStringSliceMap(cfg.NameserverPolicy),
	}

	if dns.EnhancedMode == "" {
		dns.EnhancedMode = strings.TrimSpace(opts.EnhancedMode)
	}
	if dns.Enable && len(dns.Nameserver) == 0 {
		dns.Nameserver = append([]string(nil), defaultDNSServers...)
	}
	if cfg.FallbackFilter != nil {
		dns.FallbackFilter = &mihomoDNSFallbackFilter{
			GeoIP:  cfg.FallbackFilter.GeoIP,
			IPCIDR: cloneStrings(cfg.FallbackFilter.IPCIDR),
			Domain: cloneStrings(cfg.FallbackFilter.Domain),
		}
	}

	return dns
}

func buildProfileConfig(cfg *model.ProfileConfig) *mihomoProfile {
	if cfg == nil {
		return nil
	}
	return &mihomoProfile{
		StoreSelected: cfg.StoreSelected,
		StoreFakeIP:   cfg.StoreFakeIP,
	}
}

func buildSnifferConfig(cfg *model.SnifferConfig) *mihomoSniffer {
	if cfg == nil {
		return nil
	}

	sniff := map[string]mihomoSniffRule{}
	if cfg.TLS != nil && len(cfg.TLS.Ports) > 0 {
		sniff["TLS"] = mihomoSniffRule{
			Ports: cloneStrings(cfg.TLS.Ports),
		}
	}
	if cfg.HTTP != nil && len(cfg.HTTP.Ports) > 0 {
		sniff["HTTP"] = mihomoSniffRule{
			Ports:               cloneStrings(cfg.HTTP.Ports),
			OverrideDestination: cfg.HTTP.OverrideDestination,
		}
	}
	if cfg.QUIC != nil && len(cfg.QUIC.Ports) > 0 {
		sniff["QUIC"] = mihomoSniffRule{
			Ports: cloneStrings(cfg.QUIC.Ports),
		}
	}

	return &mihomoSniffer{
		Enable:      cfg.Enable,
		ParsePureIP: cfg.ParsePureIP,
		Sniff:       sniff,
	}
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

func buildProxyGroups(nodes []model.NodeIR, template string, customGroups []model.CustomProxyGroupConfig) []mihomoProxyGroup {
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

	for _, group := range customGroups {
		if !group.Enabled {
			continue
		}
		addGroup(mihomoProxyGroup{
			Name:     group.Name,
			Type:     strings.ToLower(strings.TrimSpace(group.Type)),
			Proxies:  append([]string(nil), group.Members...),
			URL:      strings.TrimSpace(group.URL),
			Interval: group.Interval,
		})
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

func buildRules(template string, finalPolicy string, additionalRules []string, providers []model.RuleProviderConfig) []string {
	template = strings.ToLower(strings.TrimSpace(template))
	finalPolicy = strings.TrimSpace(finalPolicy)
	if finalPolicy == "" {
		finalPolicy = defaultFinalPolicy(template)
	}
	baseRules := []string{}
	if template == "lite" {
		baseRules = []string{
			"GEOSITE,private,DIRECT",
			"GEOIP,private,DIRECT,no-resolve",
			"GEOSITE,cn,DIRECT",
			"GEOIP,CN,DIRECT",
			fmt.Sprintf("MATCH,%s", finalPolicy),
		}
		return mergeRules(baseRules, additionalRules, providerRules(providers))
	}

	baseRules = []string{
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
		fmt.Sprintf("MATCH,%s", finalPolicy),
	}
	return mergeRules(baseRules, additionalRules, providerRules(providers))
}

func defaultFinalPolicy(template string) string {
	if strings.EqualFold(strings.TrimSpace(template), "lite") {
		return "节点选择"
	}
	return "漏网之鱼"
}

func mergeRules(baseRules, additionalRules, providerRules []string) []string {
	base := uniqueOrdered(baseRules)
	extras := normalizeRules(additionalRules)
	providerExtras := normalizeRules(providerRules)
	if len(extras) == 0 && len(providerExtras) == 0 {
		return base
	}

	matchIndex := len(base)
	for i, rule := range base {
		if strings.HasPrefix(strings.TrimSpace(rule), "MATCH,") {
			matchIndex = i
			break
		}
	}

	head := append([]string(nil), base[:matchIndex]...)
	tail := append([]string(nil), base[matchIndex:]...)
	merged := append(head, providerExtras...)
	merged = append(merged, extras...)
	merged = append(merged, tail...)
	return uniqueOrdered(merged)
}

func normalizeRules(rules []string) []string {
	var normalized []string
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		normalized = append(normalized, rule)
	}
	return uniqueOrdered(normalized)
}

func buildRuleProviders(providers []model.RuleProviderConfig) map[string]mihomoRuleProvider {
	if len(providers) == 0 {
		return nil
	}

	out := make(map[string]mihomoRuleProvider)
	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}

		entry := mihomoRuleProvider{
			Type:     strings.ToLower(strings.TrimSpace(provider.Type)),
			URL:      strings.TrimSpace(provider.URL),
			Path:     providerPath(provider),
			Interval: provider.Interval,
			Proxy:    strings.TrimSpace(provider.Proxy),
			Behavior: strings.ToLower(strings.TrimSpace(provider.Behavior)),
			Format:   strings.ToLower(strings.TrimSpace(provider.Format)),
		}
		if provider.SizeLimit > 0 {
			entry.SizeLimit = provider.SizeLimit
		}
		if len(provider.Headers) > 0 {
			entry.Header = cloneHeaderMap(provider.Headers)
		}
		if len(provider.Payload) > 0 {
			entry.Payload = append([]string(nil), provider.Payload...)
		}
		out[provider.Name] = entry
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func providerRules(providers []model.RuleProviderConfig) []string {
	var rules []string
	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}
		rule := fmt.Sprintf("RULE-SET,%s,%s", provider.Name, provider.Policy)
		if provider.NoResolve {
			rule += ",no-resolve"
		}
		rules = append(rules, rule)
	}
	return rules
}

func providerPath(provider model.RuleProviderConfig) string {
	path := strings.TrimSpace(provider.Path)
	if path != "" {
		return path
	}
	if strings.EqualFold(provider.Type, "http") {
		ext := strings.ToLower(strings.TrimSpace(provider.Format))
		if ext == "" {
			ext = "yaml"
		}
		return fmt.Sprintf("./rule-providers/%s.%s", sanitizeProviderName(provider.Name), ext)
	}
	return ""
}

func sanitizeProviderName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "provider"
	}
	return out
}

func cloneHeaderMap(values map[string][]string) map[string][]string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make(map[string][]string, len(values))
	for _, key := range keys {
		out[key] = append([]string(nil), values[key]...)
	}
	return out
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func cloneStringSliceMap(values map[string][]string) map[string][]string {
	if len(values) == 0 {
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make(map[string][]string, len(values))
	for _, key := range keys {
		out[key] = cloneStrings(values[key])
	}
	return out
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
	if cfg.UnifiedDelay {
		w.scalar(0, "unified-delay", cfg.UnifiedDelay)
	}
	if cfg.TCPConcurrent {
		w.scalar(0, "tcp-concurrent", cfg.TCPConcurrent)
	}
	if cfg.FindProcessMode != "" {
		w.scalar(0, "find-process-mode", cfg.FindProcessMode)
	}
	if cfg.GlobalClientFingerprint != "" {
		w.scalar(0, "global-client-fingerprint", cfg.GlobalClientFingerprint)
	}
	if cfg.DNS != nil {
		w.line(0, "dns:")
		renderDNS(&w, 4, *cfg.DNS)
	}
	if cfg.Profile != nil {
		w.line(0, "profile:")
		renderProfile(&w, 4, *cfg.Profile)
	}
	if cfg.Sniffer != nil {
		w.line(0, "sniffer:")
		renderSniffer(&w, 4, *cfg.Sniffer)
	}

	w.line(0, "proxies:")
	for _, proxy := range cfg.Proxies {
		renderProxy(&w, 4, proxy)
	}

	w.line(0, "proxy-groups:")
	for _, group := range cfg.ProxyGroups {
		renderProxyGroup(&w, 4, group)
	}

	if len(cfg.RuleProviders) > 0 {
		w.line(0, "rule-providers:")
		renderRuleProviders(&w, 4, cfg.RuleProviders)
	}

	w.line(0, "rules:")
	for _, rule := range cfg.Rules {
		w.listScalar(4, rule)
	}

	return []byte(strings.TrimRight(w.String(), "\n"))
}

func renderDNS(w *yamlWriter, indent int, dns mihomoDNS) {
	w.scalar(indent, "enable", dns.Enable)
	if dns.Listen != "" {
		w.scalar(indent, "listen", dns.Listen)
	}
	if dns.UseSystemHosts != nil {
		w.scalar(indent, "use-system-hosts", *dns.UseSystemHosts)
	}
	w.scalar(indent, "ipv6", dns.IPv6)
	if dns.EnhancedMode != "" {
		w.scalar(indent, "enhanced-mode", dns.EnhancedMode)
	}
	if dns.FakeIPRange != "" {
		w.scalar(indent, "fake-ip-range", dns.FakeIPRange)
	}
	if len(dns.DefaultNameserver) > 0 {
		w.line(indent, "default-nameserver:")
		for _, item := range dns.DefaultNameserver {
			w.listScalar(indent+4, item)
		}
	}
	if len(dns.Nameserver) > 0 {
		w.line(indent, "nameserver:")
		for _, item := range dns.Nameserver {
			w.listScalar(indent+4, item)
		}
	}
	if len(dns.Fallback) > 0 {
		w.line(indent, "fallback:")
		for _, item := range dns.Fallback {
			w.listScalar(indent+4, item)
		}
	}
	if dns.FallbackFilter != nil {
		w.line(indent, "fallback-filter:")
		w.scalar(indent+4, "geoip", dns.FallbackFilter.GeoIP)
		if len(dns.FallbackFilter.IPCIDR) > 0 {
			w.line(indent+4, "ipcidr:")
			for _, item := range dns.FallbackFilter.IPCIDR {
				w.listScalar(indent+8, item)
			}
		}
		if len(dns.FallbackFilter.Domain) > 0 {
			w.line(indent+4, "domain:")
			for _, item := range dns.FallbackFilter.Domain {
				w.listScalar(indent+8, item)
			}
		}
	}
	if len(dns.FakeIPFilter) > 0 {
		w.line(indent, "fake-ip-filter:")
		for _, item := range dns.FakeIPFilter {
			w.listScalar(indent+4, item)
		}
	}
	if len(dns.NameserverPolicy) > 0 {
		keys := make([]string, 0, len(dns.NameserverPolicy))
		for key := range dns.NameserverPolicy {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		w.line(indent, "nameserver-policy:")
		for _, key := range keys {
			w.line(indent+4, fmt.Sprintf("%s:", yamlString(key)))
			for _, item := range dns.NameserverPolicy[key] {
				w.listScalar(indent+8, item)
			}
		}
	}
}

func renderProfile(w *yamlWriter, indent int, profile mihomoProfile) {
	w.scalar(indent, "store-selected", profile.StoreSelected)
	w.scalar(indent, "store-fake-ip", profile.StoreFakeIP)
}

func renderSniffer(w *yamlWriter, indent int, sniffer mihomoSniffer) {
	w.scalar(indent, "enable", sniffer.Enable)
	w.scalar(indent, "parse-pure-ip", sniffer.ParsePureIP)
	if len(sniffer.Sniff) == 0 {
		return
	}

	w.line(indent, "sniff:")
	for _, key := range []string{"TLS", "HTTP", "QUIC"} {
		rule, ok := sniffer.Sniff[key]
		if !ok {
			continue
		}
		w.line(indent+4, key+":")
		if len(rule.Ports) > 0 {
			w.list(indent+8, "ports", rule.Ports)
		}
		if key == "HTTP" {
			w.scalar(indent+8, "override-destination", rule.OverrideDestination)
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

func renderRuleProviders(w *yamlWriter, indent int, providers map[string]mihomoRuleProvider) {
	keys := make([]string, 0, len(providers))
	for key := range providers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		provider := providers[key]
		w.line(indent, fmt.Sprintf("%s:", key))
		w.scalar(indent+2, "type", provider.Type)
		if provider.Behavior != "" {
			w.scalar(indent+2, "behavior", provider.Behavior)
		}
		if provider.Format != "" {
			w.scalar(indent+2, "format", provider.Format)
		}
		if provider.URL != "" {
			w.scalar(indent+2, "url", provider.URL)
		}
		if provider.Path != "" {
			w.scalar(indent+2, "path", provider.Path)
		}
		if provider.Interval != 0 {
			w.scalar(indent+2, "interval", provider.Interval)
		}
		if provider.Proxy != "" {
			w.scalar(indent+2, "proxy", provider.Proxy)
		}
		if provider.SizeLimit > 0 {
			w.scalar(indent+2, "size-limit", provider.SizeLimit)
		}
		if len(provider.Header) > 0 {
			w.line(indent+2, "header:")
			headerKeys := make([]string, 0, len(provider.Header))
			for headerKey := range provider.Header {
				headerKeys = append(headerKeys, headerKey)
			}
			sort.Strings(headerKeys)
			for _, headerKey := range headerKeys {
				w.list(indent+4, headerKey, provider.Header[headerKey])
			}
		}
		if len(provider.Payload) > 0 {
			w.list(indent+2, "payload", provider.Payload)
		}
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
