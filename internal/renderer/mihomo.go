package renderer

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	"subconv-next/internal/model"
	"subconv-next/internal/templatecatalog"
)

const probeURL = "https://www.gstatic.com/generate_204"

var regionGroups = []struct {
	Tag  string
	Name string
}{
	{Tag: "HK", Name: "🇭🇰 香港"},
	{Tag: "JP", Name: "🇯🇵 日本"},
	{Tag: "US", Name: "🇺🇸 美国"},
	{Tag: "SG", Name: "🇸🇬 新加坡"},
	{Tag: "TW", Name: "🇹🇼 台湾"},
	{Tag: "GB", Name: "🇬🇧 英国"},
	{Tag: "DE", Name: "🇩🇪 德国"},
	{Tag: "NL", Name: "🇳🇱 荷兰"},
	{Tag: "RU", Name: "🇷🇺 俄罗斯"},
	{Tag: "KR", Name: "🇰🇷 韩国"},
	{Tag: "FR", Name: "🇫🇷 法国"},
	{Tag: "CA", Name: "🇨🇦 加拿大"},
	{Tag: "AU", Name: "🇦🇺 澳大利亚"},
}

const (
	groupNodeSelect = "🚀 节点选择"
	groupAutoSelect = "⚡ 自动选择"
	groupFinal      = "🐟 漏网之鱼"
)

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
	GeodataMode             bool                          `yaml:"geodata-mode,omitempty"`
	GeoAutoUpdate           bool                          `yaml:"geo-auto-update,omitempty"`
	GeodataLoader           string                        `yaml:"geodata-loader,omitempty"`
	GeoUpdateInterval       int                           `yaml:"geo-update-interval,omitempty"`
	GeoxURL                 *mihomoGeoxURL                `yaml:"geox-url,omitempty"`
	DNS                     *mihomoDNS                    `yaml:"dns,omitempty"`
	Profile                 *mihomoProfile                `yaml:"profile,omitempty"`
	Sniffer                 *mihomoSniffer                `yaml:"sniffer,omitempty"`
	Proxies                 []mihomoProxy                 `yaml:"proxies"`
	ProxyGroups             []mihomoProxyGroup            `yaml:"proxy-groups"`
	RuleProviders           map[string]mihomoRuleProvider `yaml:"rule-providers,omitempty"`
	RuleProviderOrder       []string                      `yaml:"-"`
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

type mihomoGeoxURL struct {
	GeoIP   string `yaml:"geoip,omitempty"`
	GeoSite string `yaml:"geosite,omitempty"`
	MMDB    string `yaml:"mmdb,omitempty"`
	ASN     string `yaml:"asn,omitempty"`
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
	RealityOpts              *mihomoRealityOpts     `yaml:"reality-opts,omitempty"`
	WSOpts                   *mihomoWSOpts          `yaml:"ws-opts,omitempty"`
	GrpcOpts                 *mihomoGRPCOpts        `yaml:"grpc-opts,omitempty"`
	H2Opts                   *mihomoH2Opts          `yaml:"h2-opts,omitempty"`
	XHTTPOpts                *mihomoXHTTPOpts       `yaml:"xhttp-opts,omitempty"`
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

type mihomoH2Opts struct {
	Host []string `yaml:"host,omitempty"`
	Path string   `yaml:"path,omitempty"`
}

type mihomoXHTTPOpts struct {
	Path         string `yaml:"path,omitempty"`
	Mode         string `yaml:"mode,omitempty"`
	NoGRPCHeader *bool  `yaml:"no-grpc-header,omitempty"`
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
	Lazy     *bool    `yaml:"lazy,omitempty"`
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

type compiledCustomRule struct {
	Rule           model.CustomRule
	TargetGroup    string
	CreateGroup    bool
	Provider       *mihomoRuleProvider
	RuleLine       string
	InsertPosition string
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
	opts.Emoji = cfg.Render.Emoji
	opts.ShowNodeType = cfg.Render.ShowNodeType
	opts.IncludeInfoNode = cfg.Render.IncludeInfoNode
	opts.SkipTLSVerify = cfg.Render.SkipTLSVerify
	opts.UDP = cfg.Render.UDP
	opts.NodeList = cfg.Render.NodeList
	opts.SortNodes = cfg.Render.SortNodes
	opts.FilterIllegal = cfg.Render.FilterIllegal
	opts.InsertURL = cfg.Render.InsertURL
	opts.GroupProxyMode = cfg.Render.GroupProxyMode
	opts.SourcePrefix = cfg.Render.SourcePrefix
	opts.SourcePrefixFormat = cfg.Render.SourcePrefixFormat
	opts.SourcePrefixSeparator = cfg.Render.SourcePrefixSeparator
	opts.NameOptions = cfg.Render.NameOptions
	opts.DedupeScope = cfg.Render.DedupeScope
	opts.GeodataMode = cfg.Render.GeodataMode
	opts.GeoAutoUpdate = cfg.Render.GeoAutoUpdate
	opts.GeodataLoader = cfg.Render.GeodataLoader
	opts.GeoUpdateInterval = cfg.Render.GeoUpdateInterval
	opts.GeoxURL = cfg.Render.GeoxURL
	opts.IncludeKeywords = cfg.Render.IncludeKeywords
	opts.ExcludeKeywords = cfg.Render.ExcludeKeywords
	opts.OutputFilename = cfg.Render.OutputFilename
	opts.TemplateRuleMode = firstNonEmptyString(cfg.Render.SourceMode, cfg.Render.TemplateRuleMode)
	opts.ExternalConfig = cfg.Render.ExternalConfig
	opts.RuleMode = cfg.Render.RuleMode
	opts.EnabledRules = append([]string(nil), cfg.Render.EnabledRules...)
	opts.CustomRules = append([]model.CustomRule(nil), cfg.Render.CustomRules...)
	opts.UnifiedDelay = cfg.Render.UnifiedDelay
	opts.TCPConcurrent = cfg.Render.TCPConcurrent
	opts.FindProcessMode = cfg.Render.FindProcessMode
	opts.GlobalClientFingerprint = strings.TrimSpace(cfg.Render.GlobalClientFingerprint)
	opts.DNS = cfg.Render.DNS
	opts.Profile = cfg.Render.Profile
	opts.Sniffer = cfg.Render.Sniffer
	opts.FinalPolicy = cfg.Render.FinalPolicy
	opts.AdditionalRules = append([]string(nil), cfg.Render.AdditionalRules...)
	opts.RuleProviders = append([]model.RuleProviderConfig(nil), cfg.Render.RuleProviders...)
	opts.CustomProxyGroups = append([]model.CustomProxyGroupConfig(nil), cfg.Render.CustomProxyGroups...)
	opts.SubscriptionInfo = model.NormalizeSubscriptionInfoConfig(cfg.Render.SubscriptionInfo)

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
	if strings.TrimSpace(opts.GroupProxyMode) == "" {
		opts.GroupProxyMode = defaults.GroupProxyMode
	}
	if strings.TrimSpace(opts.TemplateRuleMode) == "" {
		opts.TemplateRuleMode = defaults.TemplateRuleMode
	}
	if strings.TrimSpace(opts.ExternalConfig.TemplateKey) == "" {
		opts.ExternalConfig.TemplateKey = defaults.ExternalConfig.TemplateKey
	}
	if strings.TrimSpace(opts.ExternalConfig.TemplateLabel) == "" {
		opts.ExternalConfig.TemplateLabel = defaults.ExternalConfig.TemplateLabel
	}
	opts.NameOptions.KeepRawName = true
	if !opts.SourcePrefix {
		opts.NameOptions.SourcePrefixMode = "none"
	} else if strings.TrimSpace(opts.NameOptions.SourcePrefixMode) == "" {
		opts.NameOptions.SourcePrefixMode = defaults.NameOptions.SourcePrefixMode
	}
	if strings.TrimSpace(opts.NameOptions.SourcePrefixSeparator) == "" {
		opts.NameOptions.SourcePrefixSeparator = defaults.NameOptions.SourcePrefixSeparator
	}
	if strings.TrimSpace(opts.NameOptions.DedupeSuffixStyle) == "" {
		opts.NameOptions.DedupeSuffixStyle = defaults.NameOptions.DedupeSuffixStyle
	}
	opts.SubscriptionInfo = model.NormalizeSubscriptionInfoConfig(opts.SubscriptionInfo)

	return applyTemplateMode(opts)
}

func applyTemplateMode(opts model.RenderOptions) model.RenderOptions {
	if !strings.EqualFold(strings.TrimSpace(opts.TemplateRuleMode), "template") {
		return opts
	}

	preset := templatecatalog.Resolve(opts.ExternalConfig.TemplateKey, opts.Template)
	if strings.TrimSpace(preset.Template) != "" {
		opts.Template = preset.Template
	}
	if strings.TrimSpace(preset.RuleMode) != "" {
		opts.RuleMode = preset.RuleMode
	}
	opts.EnabledRules = append([]string(nil), preset.EnabledRules...)
	opts.CustomRules = nil
	if strings.TrimSpace(preset.GroupProxyMode) != "" {
		opts.GroupProxyMode = preset.GroupProxyMode
	}
	return opts
}

func RenderMihomo(nodes []model.NodeIR, opts model.RenderOptions) ([]byte, error) {
	opts = NormalizeRenderOptions(opts)
	nodes = normalizeRenderNodesPreserveNames(nodes)
	nodes = ensureUniqueNames(nodes, opts.NameOptions.DedupeSuffixStyle)

	proxies, err := buildProxies(nodes)
	if err != nil {
		return nil, err
	}
	resolvedEnabledRules := resolveEnabledRules(opts.RuleMode, opts.EnabledRules)

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
		GeodataMode:             opts.GeodataMode,
		GeoAutoUpdate:           opts.GeoAutoUpdate,
		GeodataLoader:           strings.TrimSpace(opts.GeodataLoader),
		GeoUpdateInterval:       opts.GeoUpdateInterval,
		Proxies:                 proxies,
		RuleProviders:           buildRuleProviders(opts.RuleProviders, resolvedEnabledRules, opts.CustomRules, opts.CustomProxyGroups),
		RuleProviderOrder:       orderedProviderNames(resolvedEnabledRules),
		Rules:                   buildRules(opts.Template, opts.FinalPolicy, resolvedEnabledRules, opts.AdditionalRules, opts.RuleProviders, opts.CustomRules, opts.CustomProxyGroups),
	}
	proxyGroups, err := buildProxyGroups(nodes, opts.Template, opts.CustomProxyGroups, resolvedEnabledRules, opts.CustomRules, opts.GroupProxyMode)
	if err != nil {
		return nil, err
	}
	cfg.ProxyGroups = proxyGroups
	cfg.GeoxURL = buildGeoxURLConfig(opts.GeoxURL)
	cfg.DNS = buildDNSConfig(opts)
	cfg.Profile = buildProfileConfig(opts.Profile)
	cfg.Sniffer = buildSnifferConfig(opts.Sniffer)
	if err := failOnCriticalWarnings(ValidateMihomoConfig(cfg)); err != nil {
		return nil, err
	}

	return renderConfig(cfg), nil
}

func buildDNSConfig(opts model.RenderOptions) *mihomoDNS {
	if !opts.DNSEnabled {
		return nil
	}

	defaultRender := model.DefaultRenderConfig()
	base := defaultRender.DNS
	if base == nil {
		return nil
	}

	cfg := *base
	if opts.DNS != nil {
		cfg = *opts.DNS
	}

	dns := &mihomoDNS{
		Enable:            cfg.Enable,
		Listen:            strings.TrimSpace(cfg.Listen),
		UseSystemHosts:    model.Bool(cfg.UseSystemHosts),
		IPv6:              opts.IPv6,
		EnhancedMode:      firstNonEmptyString(strings.TrimSpace(cfg.EnhancedMode), strings.TrimSpace(opts.EnhancedMode)),
		FakeIPRange:       strings.TrimSpace(cfg.FakeIPRange),
		DefaultNameserver: cloneStrings(cfg.DefaultNameserver),
		Nameserver:        cloneStrings(cfg.Nameserver),
		Fallback:          cloneStrings(cfg.Fallback),
		FakeIPFilter:      cloneStrings(cfg.FakeIPFilter),
		NameserverPolicy:  cloneStringSliceMap(cfg.NameserverPolicy),
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
		cfg = model.DefaultRenderConfig().Profile
	}
	if cfg == nil {
		return nil
	}
	return &mihomoProfile{
		StoreSelected: cfg.StoreSelected,
		StoreFakeIP:   cfg.StoreFakeIP,
	}
}

func buildGeoxURLConfig(cfg *model.GeoxURLConfig) *mihomoGeoxURL {
	if cfg == nil {
		cfg = model.DefaultRenderConfig().GeoxURL
	}
	if cfg == nil {
		return nil
	}
	return &mihomoGeoxURL{
		GeoIP:   strings.TrimSpace(cfg.GeoIP),
		GeoSite: strings.TrimSpace(cfg.GeoSite),
		MMDB:    strings.TrimSpace(cfg.MMDB),
		ASN:     strings.TrimSpace(cfg.ASN),
	}
}

func buildSnifferConfig(cfg *model.SnifferConfig) *mihomoSniffer {
	if cfg == nil {
		cfg = model.DefaultRenderConfig().Sniffer
	}
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
	case model.ProtocolSSR:
		proxy.Cipher = rawString(node.Raw, "method")
		proxy.Password = node.Auth.Password
		proxy.Protocol = rawString(node.Raw, "protocol")
		proxy.ProtocolParam = rawString(node.Raw, "protocolParam")
		proxy.Obfs = rawString(node.Raw, "obfs")
		proxy.ObfsParam = rawString(node.Raw, "obfsParam")
	case model.ProtocolVMess:
		proxy.UUID = node.Auth.UUID
		proxy.Cipher = firstNonEmptyString(rawString(node.Raw, "cipher"), rawString(node.Raw, "scy"), "auto")
		proxy.Network = node.Transport.Network
		proxy.TLS = node.TLS.Enabled
		proxy.ServerName = node.TLS.SNI
		proxy.ClientFingerprint = node.TLS.ClientFingerprint
		proxy.AlterID = rawScalar(node.Raw, "alterId")
		applyTransport(&proxy, node.Transport)
	case model.ProtocolVLESS:
		proxy.UUID = node.Auth.UUID
		proxy.Encryption = firstNonEmptyString(rawString(node.Raw, "encryption"), "none")
		proxy.Network = node.Transport.Network
		proxy.TLS = node.TLS.Enabled
		proxy.ServerName = node.TLS.SNI
		proxy.ClientFingerprint = node.TLS.ClientFingerprint
		proxy.PacketEncoding = rawString(node.Raw, "packetEncoding")
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
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
		applyTransport(&proxy, node.Transport)
	case model.ProtocolHysteria2:
		proxy.Password = node.Auth.Password
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
		proxy.Obfs = rawString(node.Raw, "obfs")
		proxy.ObfsPassword = rawString(node.Raw, "obfsPassword")
	case model.ProtocolTUIC:
		proxy.UUID = node.Auth.UUID
		proxy.Password = node.Auth.Password
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
		proxy.CongestionController = rawString(node.Raw, "congestionController")
		proxy.UDPRelayMode = rawString(node.Raw, "udpRelayMode")
		proxy.ReduceRTT = rawBoolPointer(node.Raw, "reduceRTT")
	case model.ProtocolAnyTLS:
		proxy.Password = node.Auth.Password
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.ClientFingerprint = node.TLS.ClientFingerprint
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
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
	case "h2":
		opts := &mihomoH2Opts{
			Path: transport.Path,
		}
		if len(transport.H2Hosts) > 0 {
			opts.Host = append([]string(nil), transport.H2Hosts...)
		} else if transport.Host != "" {
			opts.Host = []string{transport.Host}
		}
		if opts.Path != "" || len(opts.Host) > 0 {
			proxy.H2Opts = opts
		}
	case "xhttp":
		proxy.XHTTPOpts = &mihomoXHTTPOpts{
			Path:         transport.Path,
			Mode:         firstNonEmptyString(transport.Mode, "auto"),
			NoGRPCHeader: firstNonNilBool(transport.NoGRPCHeader, model.Bool(false)),
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

func buildProxyGroups(nodes []model.NodeIR, template string, customGroups []model.CustomProxyGroupConfig, enabledRules []string, customRules []model.CustomRule, groupProxyMode string) ([]mihomoProxyGroup, error) {
	groupProxyMode = strings.ToLower(strings.TrimSpace(groupProxyMode))
	if groupProxyMode == "" {
		groupProxyMode = "compact"
	}
	allNames := nodeNames(nodes)
	regularNames := nodeNames(filterOutInfoNodes(nodes))
	if len(allNames) == 0 {
		return []mihomoProxyGroup{
			{
				Name:    groupNodeSelect,
				Type:    "select",
				Proxies: []string{"DIRECT"},
			},
		}, nil
	}

	groupNames := map[string]struct{}{}
	var groups []mihomoProxyGroup

	addGroup := func(group mihomoProxyGroup) error {
		group.Proxies = uniqueOrdered(group.Proxies)
		if len(group.Proxies) == 0 {
			return nil
		}
		if _, exists := groupNames[group.Name]; exists {
			return fmt.Errorf("duplicate proxy-group name %q", group.Name)
		}
		groups = append(groups, group)
		groupNames[group.Name] = struct{}{}
		return nil
	}

	regionNameOrder := regionProxyNames(nodes)
	var regionGroupNames []string
	for _, region := range regionGroups {
		names := regionNameOrder[region.Tag]
		if len(names) == 0 {
			continue
		}
		regionGroupNames = append(regionGroupNames, region.Name)
	}

	if err := addGroup(mihomoProxyGroup{
		Name:    groupNodeSelect,
		Type:    "select",
		Proxies: append([]string{groupAutoSelect, "DIRECT", "REJECT"}, append(append([]string{}, regionGroupNames...), allNames...)...),
	}); err != nil {
		return nil, err
	}
	if err := addGroup(mihomoProxyGroup{
		Name:     groupAutoSelect,
		Type:     "url-test",
		Proxies:  regularNames,
		URL:      probeURL,
		Interval: 300,
		Lazy:     model.Bool(false),
	}); err != nil {
		return nil, err
	}

	for _, region := range regionGroups {
		names := regionNameOrder[region.Tag]
		if len(names) == 0 {
			continue
		}
		if err := addGroup(mihomoProxyGroup{
			Name:    region.Name,
			Type:    "select",
			Proxies: append([]string{groupAutoSelect, "DIRECT"}, names...),
		}); err != nil {
			return nil, err
		}
	}

	resolvedCustomRules := resolveCustomRules(customRules, enabledRules, regionGroupNames, customGroups)
	if len(enabledRules) > 0 || len(resolvedCustomRules) > 0 {
		for _, category := range orderedRuleCategories(enabledRules) {
			if err := addGroup(mihomoProxyGroup{
				Name:    category.GroupName,
				Type:    "select",
				Proxies: serviceGroupProxiesForCategory(category.Key, groupProxyMode, regionGroupNames, allNames),
			}); err != nil {
				return nil, err
			}
		}

		customGroupProxies := serviceGroupProxiesForCategory("", groupProxyMode, regionGroupNames, allNames)
		for _, rule := range resolvedCustomRules {
			if !rule.CreateGroup {
				continue
			}
			if err := addGroup(mihomoProxyGroup{Name: rule.TargetGroup, Type: "select", Proxies: customGroupProxies}); err != nil {
				return nil, err
			}
		}
	}

	if err := addGroup(mihomoProxyGroup{
		Name:    groupFinal,
		Type:    "select",
		Proxies: serviceGroupProxiesForCategory("", groupProxyMode, regionGroupNames, allNames),
	}); err != nil {
		return nil, err
	}

	for _, group := range customGroups {
		if !group.Enabled {
			continue
		}
		if err := addGroup(mihomoProxyGroup{
			Name:     group.Name,
			Type:     strings.ToLower(strings.TrimSpace(group.Type)),
			Proxies:  append([]string(nil), group.Members...),
			URL:      strings.TrimSpace(group.URL),
			Interval: group.Interval,
		}); err != nil {
			return nil, err
		}
	}

	return groups, nil
}

func filterOutInfoNodes(nodes []model.NodeIR) []model.NodeIR {
	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		if rendererInfoNode(node) {
			continue
		}
		out = append(out, node)
	}
	return out
}

func rendererInfoNode(node model.NodeIR) bool {
	if value, ok := node.Raw["_infoNode"].(bool); ok && value {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(node.Name))
	for _, pattern := range []string{"剩余流量", "已用流量", "总流量", "到期", "过期", "有效期", "套餐", "官网"} {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	for _, word := range []string{"expire", "traffic", "used", "remaining", "total"} {
		if containsWord(lower, word) {
			return true
		}
	}
	return false
}

func containsWord(value, word string) bool {
	for index := 0; ; {
		pos := strings.Index(value[index:], word)
		if pos == -1 {
			return false
		}
		pos += index
		beforeOK := pos == 0 || !rendererAlphaNum(value[pos-1])
		afterIndex := pos + len(word)
		afterOK := afterIndex >= len(value) || !rendererAlphaNum(value[afterIndex])
		if beforeOK && afterOK {
			return true
		}
		index = pos + len(word)
	}
}

func rendererAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '_'
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

func serviceGroupProxiesForCategory(categoryKey, groupProxyMode string, regionGroupNames, allNames []string) []string {
	switch categoryKey {
	case "adblock":
		return []string{"REJECT", "DIRECT", groupNodeSelect}
	case "private", "domestic":
		return []string{"DIRECT", "REJECT", groupNodeSelect, groupAutoSelect}
	default:
		proxies := []string{groupNodeSelect, groupAutoSelect, "DIRECT", "REJECT"}
		if groupProxyMode == "regional" || groupProxyMode == "full" {
			proxies = append(proxies, regionGroupNames...)
		}
		if groupProxyMode == "full" {
			proxies = append(proxies, allNames...)
		}
		return proxies
	}
}

func resolveCustomRules(rules []model.CustomRule, enabledRules []string, regionGroupNames []string, customGroups []model.CustomProxyGroupConfig) []compiledCustomRule {
	reserved := map[string]struct{}{
		groupNodeSelect: {},
		groupAutoSelect: {},
		groupFinal:      {},
	}
	for _, region := range regionGroupNames {
		reserved[region] = struct{}{}
	}
	for _, category := range orderedRuleCategories(enabledRules) {
		reserved[category.GroupName] = struct{}{}
	}
	for _, group := range customGroups {
		if strings.TrimSpace(group.Name) != "" {
			reserved[group.Name] = struct{}{}
		}
	}

	var out []compiledCustomRule
	for _, rule := range rules {
		if !rule.Enabled || strings.TrimSpace(rule.Key) == "" || strings.TrimSpace(rule.Label) == "" {
			continue
		}
		targetGroup := strings.TrimSpace(rule.TargetGroup)
		createGroup := false
		switch strings.ToLower(strings.TrimSpace(rule.TargetMode)) {
		case "direct":
			targetGroup = "DIRECT"
		case "reject":
			targetGroup = "REJECT"
		case "existing_group":
			if targetGroup == "" {
				targetGroup = groupNodeSelect
			}
		default:
			createGroup = true
			if targetGroup == "" {
				targetGroup = customRuleDisplayName(rule)
			}
			targetGroup = uniqueGroupName(targetGroup, reserved)
			reserved[targetGroup] = struct{}{}
		}

		compiled := compiledCustomRule{
			Rule:           rule,
			TargetGroup:    targetGroup,
			CreateGroup:    createGroup,
			InsertPosition: firstNonEmptyString(strings.TrimSpace(rule.InsertPosition), "before_match"),
		}

		if !strings.EqualFold(strings.TrimSpace(rule.SourceType), "group_only") {
			provider := buildCustomRuleProvider(rule)
			compiled.Provider = &provider
			compiled.RuleLine = renderRuleSet(RuleSpec{
				Provider:    rule.Key,
				TargetGroup: targetGroup,
				NoResolve:   rule.NoResolve,
			})
		}

		out = append(out, compiled)
	}
	return out
}

func customRulesForPosition(rules []compiledCustomRule, position string) []string {
	var out []string
	for _, rule := range rules {
		if rule.RuleLine == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(rule.InsertPosition), position) {
			out = append(out, rule.RuleLine)
		}
	}
	return out
}

func buildCustomRuleProvider(rule model.CustomRule) mihomoRuleProvider {
	sourceType := strings.ToLower(strings.TrimSpace(rule.SourceType))
	switch sourceType {
	case "http":
		return mihomoRuleProvider{
			Type:     "http",
			Behavior: strings.ToLower(strings.TrimSpace(rule.Behavior)),
			URL:      strings.TrimSpace(rule.URL),
			Path:     firstNonEmptyString(strings.TrimSpace(rule.Path), customRuleProviderPath(rule)),
			Interval: rule.Interval,
			Format:   strings.ToLower(strings.TrimSpace(rule.Format)),
		}
	case "file":
		return mihomoRuleProvider{
			Type:     "file",
			Behavior: strings.ToLower(strings.TrimSpace(rule.Behavior)),
			Path:     strings.TrimSpace(rule.Path),
			Format:   strings.ToLower(strings.TrimSpace(rule.Format)),
		}
	default:
		return mihomoRuleProvider{
			Type:     "inline",
			Behavior: strings.ToLower(strings.TrimSpace(rule.Behavior)),
			Format:   strings.ToLower(strings.TrimSpace(rule.Format)),
			Payload:  append([]string(nil), rule.Payload...),
		}
	}
}

func customRuleProviderPath(rule model.CustomRule) string {
	format := strings.ToLower(strings.TrimSpace(rule.Format))
	switch format {
	case "mrs":
		return "./ruleset/" + rule.Key + ".mrs"
	case "text":
		return "./ruleset/" + rule.Key + ".txt"
	default:
		return "./ruleset/" + rule.Key + ".yaml"
	}
}

func customRuleDisplayName(rule model.CustomRule) string {
	if strings.TrimSpace(rule.Icon) != "" {
		return strings.TrimSpace(rule.Icon + " " + rule.Label)
	}
	return strings.TrimSpace(rule.Label)
}

func uniqueGroupName(name string, reserved map[string]struct{}) string {
	candidate := strings.TrimSpace(name)
	if candidate == "" {
		candidate = "自定义规则"
	}
	if _, ok := reserved[candidate]; !ok {
		return candidate
	}
	for i := 2; ; i++ {
		next := fmt.Sprintf("%s %d", candidate, i)
		if _, ok := reserved[next]; !ok {
			return next
		}
	}
}

func buildRules(template string, finalPolicy string, enabledRules []string, additionalRules []string, providers []model.RuleProviderConfig, customRules []model.CustomRule, customGroups []model.CustomProxyGroupConfig) []string {
	finalPolicy = strings.TrimSpace(finalPolicy)
	if finalPolicy == "" {
		finalPolicy = groupFinal
	}

	var baseRules []string
	enabled := normalizeEnabledRuleKeys(enabledRules)
	resolvedCustomRules := resolveCustomRules(customRules, enabled, nil, customGroups)
	for _, category := range orderedRuleCategories(enabled) {
		if category.Key == "domestic" {
			baseRules = append(baseRules, customRulesForPosition(resolvedCustomRules, "before_domestic")...)
		}
		if category.Key == "non_cn" {
			baseRules = append(baseRules, customRulesForPosition(resolvedCustomRules, "before_non_cn")...)
		}
		if category.Key == "domestic" {
			for _, spec := range category.Rules {
				baseRules = append(baseRules, renderRuleSet(spec))
			}
			continue
		}
		for _, spec := range category.Rules {
			baseRules = append(baseRules, renderRuleSet(spec))
		}
		if category.Key == "adblock" {
			baseRules = append(baseRules, customRulesForPosition(resolvedCustomRules, "after_adblock")...)
		}
	}
	if containsString(enabled, "domestic") {
		baseRules = append(baseRules, renderRuleSet(RuleSpec{
			Provider:    "cn",
			TargetGroup: "🔒 国内服务",
		}))
	}

	baseRules = append(baseRules, customRulesForPosition(resolvedCustomRules, "before_match")...)
	baseRules = append(baseRules, fmt.Sprintf("MATCH,%s", normalizeFinalPolicyName(finalPolicy)))
	return mergeRules(baseRules, additionalRules, providerRules(providers))
}

func renderRuleSet(spec RuleSpec) string {
	rule := fmt.Sprintf("RULE-SET,%s,%s", spec.Provider, spec.TargetGroup)
	if spec.NoResolve {
		rule += ",no-resolve"
	}
	return rule
}

func normalizeFinalPolicyName(value string) string {
	switch strings.TrimSpace(value) {
	case "", "漏网之鱼":
		return groupFinal
	case "节点选择":
		return groupNodeSelect
	default:
		return value
	}
}

func mergeRules(baseRules, additionalRules, providerRules []string) []string {
	base := uniqueOrdered(baseRules)
	providerExtras := normalizeRules(providerRules)
	extras := normalizeRules(additionalRules)
	matchRule := ""
	if len(base) > 0 && strings.HasPrefix(strings.TrimSpace(base[len(base)-1]), "MATCH,") {
		matchRule = base[len(base)-1]
		base = base[:len(base)-1]
	}
	merged := append([]string{}, base...)
	merged = append(merged, providerExtras...)
	merged = append(merged, extras...)
	if matchRule != "" {
		merged = append(merged, matchRule)
	}
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

func buildRuleProviders(providers []model.RuleProviderConfig, enabledRules []string, customRules []model.CustomRule, customGroups []model.CustomProxyGroupConfig) map[string]mihomoRuleProvider {
	if len(providers) == 0 && len(enabledRules) == 0 && len(customRules) == 0 {
		return nil
	}

	out := make(map[string]mihomoRuleProvider)
	for _, category := range orderedRuleCategories(enabledRules) {
		for _, provider := range category.Providers {
			out[provider.Name] = mihomoRuleProvider{
				Type:     provider.Type,
				URL:      provider.URL,
				Path:     provider.Path,
				Interval: provider.Interval,
				Behavior: provider.Behavior,
				Format:   provider.Format,
			}
		}
	}
	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}

		entry := mihomoRuleProvider{
			Type:     strings.ToLower(strings.TrimSpace(provider.Type)),
			URL:      strings.TrimSpace(provider.URL),
			Path:     providerPath(provider),
			Interval: provider.Interval,
			Behavior: strings.ToLower(strings.TrimSpace(provider.Behavior)),
			Format:   strings.ToLower(strings.TrimSpace(provider.Format)),
		}
		if strings.TrimSpace(provider.Proxy) != "" {
			entry.Proxy = strings.TrimSpace(provider.Proxy)
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
	for _, rule := range resolveCustomRules(customRules, enabledRules, nil, customGroups) {
		if rule.Provider == nil {
			continue
		}
		out[rule.Rule.Key] = *rule.Provider
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

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonNilBool(values ...*bool) *bool {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
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

func normalizeRenderNodesPreserveNames(nodes []model.NodeIR) []model.NodeIR {
	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		name := node.Name
		node = model.NormalizeNode(node)
		if name != "" {
			node.Name = name
		}
		out = append(out, node)
	}
	return out
}

func ensureUniqueNames(nodes []model.NodeIR, suffixStyle string) []model.NodeIR {
	return model.EnsureUniqueProxyNames(nodes, suffixStyle)
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

func containsString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}

var yamlUnicodeEscapePattern = regexp.MustCompile(`\\U[0-9A-Fa-f]{8}|\\u[0-9A-Fa-f]{4}`)

func unescapeYAMLUnicodeEscapes(value string) string {
	return yamlUnicodeEscapePattern.ReplaceAllStringFunc(value, func(match string) string {
		decoded, err := strconv.Unquote(`"` + match + `"`)
		if err != nil {
			return match
		}
		return decoded
	})
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
	case model.ProtocolSS, model.ProtocolSSR, model.ProtocolVMess, model.ProtocolVLESS, model.ProtocolTrojan, model.ProtocolHysteria2, model.ProtocolTUIC, model.ProtocolAnyTLS, model.ProtocolWireGuard:
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
	sections := make([]struct {
		key   string
		value *yaml.Node
	}, 0, 21)

	addSection := func(key string, value *yaml.Node) {
		if value == nil {
			return
		}
		sections = append(sections, struct {
			key   string
			value *yaml.Node
		}{key: key, value: value})
	}

	addSection("mixed-port", scalarNode(cfg.MixedPort))
	addSection("allow-lan", scalarNode(cfg.AllowLAN))
	addSection("mode", scalarNode(cfg.Mode))
	addSection("log-level", scalarNode(cfg.LogLevel))
	addSection("ipv6", scalarNode(cfg.IPv6))
	if cfg.UnifiedDelay {
		addSection("unified-delay", scalarNode(cfg.UnifiedDelay))
	}
	if cfg.TCPConcurrent {
		addSection("tcp-concurrent", scalarNode(cfg.TCPConcurrent))
	}
	if cfg.FindProcessMode != "" {
		addSection("find-process-mode", scalarNode(cfg.FindProcessMode))
	}
	if cfg.GlobalClientFingerprint != "" {
		addSection("global-client-fingerprint", scalarNode(cfg.GlobalClientFingerprint))
	}
	if cfg.DNS != nil {
		addSection("dns", buildDNSNode(*cfg.DNS))
	}
	if cfg.Profile != nil {
		addSection("profile", buildProfileNode(*cfg.Profile))
	}
	if cfg.Sniffer != nil {
		addSection("sniffer", buildSnifferNode(*cfg.Sniffer))
	}
	if cfg.GeodataMode {
		addSection("geodata-mode", scalarNode(cfg.GeodataMode))
	}
	if cfg.GeoAutoUpdate {
		addSection("geo-auto-update", scalarNode(cfg.GeoAutoUpdate))
	}
	if cfg.GeodataLoader != "" {
		addSection("geodata-loader", scalarNode(cfg.GeodataLoader))
	}
	if cfg.GeoUpdateInterval > 0 {
		addSection("geo-update-interval", scalarNode(cfg.GeoUpdateInterval))
	}
	if cfg.GeoxURL != nil {
		addSection("geox-url", buildGeoxURLNode(*cfg.GeoxURL))
	}
	addSection("proxies", buildProxiesNode(cfg.Proxies))
	addSection("proxy-groups", buildProxyGroupsNode(cfg.ProxyGroups))
	if len(cfg.RuleProviders) > 0 {
		addSection("rule-providers", buildRuleProvidersNode(cfg.RuleProviders, cfg.RuleProviderOrder))
	}
	addSection("rules", buildRulesNode(cfg.Rules))

	var out bytes.Buffer
	for index, section := range sections {
		if index > 0 {
			out.WriteString("\n\n")
		}
		doc := mappingNode()
		appendMap(doc, section.key, section.value)

		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(doc); err != nil {
			panic(err)
		}
		_ = enc.Close()
		out.Write(bytes.TrimSuffix(buf.Bytes(), []byte("\n")))
	}
	return []byte(unescapeYAMLUnicodeEscapes(out.String()))
}

func buildDNSNode(dns mihomoDNS) *yaml.Node {
	node := mappingNode()
	appendMap(node, "enable", scalarNode(dns.Enable))
	if dns.Listen != "" {
		appendMap(node, "listen", scalarNode(dns.Listen))
	}
	if dns.UseSystemHosts != nil {
		appendMap(node, "use-system-hosts", scalarNode(*dns.UseSystemHosts))
	}
	appendMap(node, "ipv6", scalarNode(dns.IPv6))
	if dns.EnhancedMode != "" {
		appendMap(node, "enhanced-mode", scalarNode(dns.EnhancedMode))
	}
	if dns.FakeIPRange != "" {
		appendMap(node, "fake-ip-range", scalarNode(dns.FakeIPRange))
	}
	if len(dns.DefaultNameserver) > 0 {
		appendMap(node, "default-nameserver", flowStringSeqNode(dns.DefaultNameserver))
	}
	if len(dns.Nameserver) > 0 {
		appendMap(node, "nameserver", flowStringSeqNode(dns.Nameserver))
	}
	if len(dns.Fallback) > 0 {
		appendMap(node, "fallback", flowStringSeqNode(dns.Fallback))
	}
	if dns.FallbackFilter != nil {
		filter := flowMappingNode()
		appendMap(filter, "geoip", scalarNode(dns.FallbackFilter.GeoIP))
		if len(dns.FallbackFilter.IPCIDR) > 0 {
			appendMap(filter, "ipcidr", flowStringSeqNode(dns.FallbackFilter.IPCIDR))
		}
		if len(dns.FallbackFilter.Domain) > 0 {
			appendMap(filter, "domain", flowStringSeqNode(dns.FallbackFilter.Domain))
		}
		appendMap(node, "fallback-filter", filter)
	}
	if len(dns.FakeIPFilter) > 0 {
		appendMap(node, "fake-ip-filter", flowStringSeqNode(dns.FakeIPFilter))
	}
	if len(dns.NameserverPolicy) > 0 {
		policy := mappingNode()
		keys := make([]string, 0, len(dns.NameserverPolicy))
		for key := range dns.NameserverPolicy {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendMap(policy, key, stringSeqNode(dns.NameserverPolicy[key]))
		}
		appendMap(node, "nameserver-policy", policy)
	}
	return node
}

func buildGeoxURLNode(cfg mihomoGeoxURL) *yaml.Node {
	node := mappingNode()
	if cfg.GeoIP != "" {
		appendMap(node, "geoip", scalarNode(cfg.GeoIP))
	}
	if cfg.GeoSite != "" {
		appendMap(node, "geosite", scalarNode(cfg.GeoSite))
	}
	if cfg.MMDB != "" {
		appendMap(node, "mmdb", scalarNode(cfg.MMDB))
	}
	if cfg.ASN != "" {
		appendMap(node, "asn", scalarNode(cfg.ASN))
	}
	return node
}

func buildProfileNode(profile mihomoProfile) *yaml.Node {
	node := mappingNode()
	appendMap(node, "store-selected", scalarNode(profile.StoreSelected))
	appendMap(node, "store-fake-ip", scalarNode(profile.StoreFakeIP))
	return node
}

func buildSnifferNode(sniffer mihomoSniffer) *yaml.Node {
	node := mappingNode()
	appendMap(node, "enable", scalarNode(sniffer.Enable))
	appendMap(node, "parse-pure-ip", scalarNode(sniffer.ParsePureIP))
	if len(sniffer.Sniff) == 0 {
		return node
	}
	sniff := mappingNode()
	for _, key := range []string{"HTTP", "QUIC", "TLS"} {
		rule, ok := sniffer.Sniff[key]
		if !ok {
			continue
		}
		ruleNode := mappingNode()
		if len(rule.Ports) > 0 {
			appendMap(ruleNode, "ports", stringSeqNode(rule.Ports))
		}
		if key == "HTTP" {
			appendMap(ruleNode, "override-destination", scalarNode(rule.OverrideDestination))
		}
		appendMap(sniff, key, ruleNode)
	}
	appendMap(node, "sniff", sniff)
	return node
}

func buildProxiesNode(proxies []mihomoProxy) *yaml.Node {
	seq := sequenceNode()
	for _, proxy := range proxies {
		seq.Content = append(seq.Content, buildProxyNode(proxy))
	}
	return seq
}

func buildProxyNode(proxy mihomoProxy) *yaml.Node {
	node := flowMappingNode()
	appendMap(node, "name", scalarNode(proxy.Name))
	appendMap(node, "type", scalarNode(proxy.Type))
	if proxy.Server != "" {
		appendMap(node, "server", scalarNode(proxy.Server))
	}
	if proxy.Port != 0 {
		appendMap(node, "port", scalarNode(proxy.Port))
	}
	if proxy.Encryption != "" {
		appendMap(node, "encryption", scalarNode(proxy.Encryption))
	}
	if proxy.Cipher != "" {
		appendMap(node, "cipher", scalarNode(proxy.Cipher))
	}
	if proxy.Password != "" {
		appendMap(node, "password", scalarNode(proxy.Password))
	}
	if proxy.Username != "" {
		appendMap(node, "username", scalarNode(proxy.Username))
	}
	if proxy.UUID != "" {
		appendMap(node, "uuid", scalarNode(proxy.UUID))
	}
	if proxy.AlterID != nil {
		appendMap(node, "alterId", nodeFromAny(proxy.AlterID))
	}
	if proxy.Network != "" {
		appendMap(node, "network", scalarNode(proxy.Network))
	}
	if proxy.TLS {
		appendMap(node, "tls", scalarNode(proxy.TLS))
	}
	if proxy.UDP != nil {
		appendMap(node, "udp", scalarNode(*proxy.UDP))
	}
	if proxy.ServerName != "" {
		appendMap(node, "servername", scalarNode(proxy.ServerName))
	}
	if proxy.SNI != "" {
		appendMap(node, "sni", scalarNode(proxy.SNI))
	}
	if proxy.ClientFingerprint != "" {
		appendMap(node, "client-fingerprint", scalarNode(proxy.ClientFingerprint))
	}
	if proxy.PacketEncoding != "" {
		appendMap(node, "packet-encoding", scalarNode(proxy.PacketEncoding))
	}
	if proxy.SkipCertVerify != nil {
		appendMap(node, "skip-cert-verify", scalarNode(*proxy.SkipCertVerify))
	}
	if len(proxy.ALPN) > 0 {
		appendMap(node, "alpn", flowStringSeqNode(proxy.ALPN))
	}
	if proxy.Flow != "" {
		appendMap(node, "flow", scalarNode(proxy.Flow))
	}
	if proxy.RealityOpts != nil {
		reality := flowMappingNode()
		if proxy.RealityOpts.PublicKey != "" {
			appendMap(reality, "public-key", scalarNode(proxy.RealityOpts.PublicKey))
		}
		if proxy.RealityOpts.ShortID != "" {
			appendMap(reality, "short-id", scalarNode(proxy.RealityOpts.ShortID))
		}
		if proxy.RealityOpts.SpiderX != "" {
			appendMap(reality, "spider-x", scalarNode(proxy.RealityOpts.SpiderX))
		}
		appendMap(node, "reality-opts", reality)
	}
	if proxy.WSOpts != nil {
		ws := flowMappingNode()
		if proxy.WSOpts.Path != "" {
			appendMap(ws, "path", scalarNode(proxy.WSOpts.Path))
		}
		if len(proxy.WSOpts.Headers) > 0 {
			headers := flowMappingNode()
			for _, key := range sortedKeysString(proxy.WSOpts.Headers) {
				appendMap(headers, key, scalarNode(proxy.WSOpts.Headers[key]))
			}
			appendMap(ws, "headers", headers)
		}
		appendMap(node, "ws-opts", ws)
	}
	if proxy.GrpcOpts != nil && proxy.GrpcOpts.GrpcServiceName != "" {
		grpc := flowMappingNode()
		appendMap(grpc, "grpc-service-name", scalarNode(proxy.GrpcOpts.GrpcServiceName))
		appendMap(node, "grpc-opts", grpc)
	}
	if proxy.H2Opts != nil {
		h2 := flowMappingNode()
		if len(proxy.H2Opts.Host) > 0 {
			appendMap(h2, "host", flowStringSeqNode(proxy.H2Opts.Host))
		}
		if proxy.H2Opts.Path != "" {
			appendMap(h2, "path", scalarNode(proxy.H2Opts.Path))
		}
		appendMap(node, "h2-opts", h2)
	}
	if proxy.XHTTPOpts != nil {
		xhttp := flowMappingNode()
		if proxy.XHTTPOpts.Path != "" {
			appendMap(xhttp, "path", scalarNode(proxy.XHTTPOpts.Path))
		}
		if proxy.XHTTPOpts.Mode != "" {
			appendMap(xhttp, "mode", scalarNode(proxy.XHTTPOpts.Mode))
		}
		if proxy.XHTTPOpts.NoGRPCHeader != nil {
			appendMap(xhttp, "no-grpc-header", scalarNode(*proxy.XHTTPOpts.NoGRPCHeader))
		}
		appendMap(node, "xhttp-opts", xhttp)
	}
	if proxy.Obfs != "" {
		appendMap(node, "obfs", scalarNode(proxy.Obfs))
	}
	if proxy.ObfsPassword != "" {
		appendMap(node, "obfs-password", scalarNode(proxy.ObfsPassword))
	}
	if proxy.CongestionController != "" {
		appendMap(node, "congestion-controller", scalarNode(proxy.CongestionController))
	}
	if proxy.UDPRelayMode != "" {
		appendMap(node, "udp-relay-mode", scalarNode(proxy.UDPRelayMode))
	}
	if proxy.ReduceRTT != nil {
		appendMap(node, "reduce-rtt", scalarNode(*proxy.ReduceRTT))
	}
	if proxy.IdleSessionCheckInterval != nil {
		appendMap(node, "idle-session-check-interval", nodeFromAny(proxy.IdleSessionCheckInterval))
	}
	if proxy.IdleSessionTimeout != nil {
		appendMap(node, "idle-session-timeout", nodeFromAny(proxy.IdleSessionTimeout))
	}
	if proxy.MinIdleSession != nil {
		appendMap(node, "min-idle-session", nodeFromAny(proxy.MinIdleSession))
	}
	if proxy.IP != "" {
		appendMap(node, "ip", scalarNode(proxy.IP))
	}
	if proxy.IPv6 != "" {
		appendMap(node, "ipv6", scalarNode(proxy.IPv6))
	}
	if proxy.PrivateKey != "" {
		appendMap(node, "private-key", scalarNode(proxy.PrivateKey))
	}
	if proxy.PublicKey != "" {
		appendMap(node, "public-key", scalarNode(proxy.PublicKey))
	}
	if len(proxy.AllowedIPs) > 0 {
		appendMap(node, "allowed-ips", flowStringSeqNode(proxy.AllowedIPs))
	}
	if proxy.PreSharedKey != "" {
		appendMap(node, "pre-shared-key", scalarNode(proxy.PreSharedKey))
	}
	if proxy.Reserved != nil {
		appendMap(node, "reserved", nodeFromAny(proxy.Reserved))
	}
	if proxy.PersistentKeepalive != 0 {
		appendMap(node, "persistent-keepalive", scalarNode(proxy.PersistentKeepalive))
	}
	if proxy.MTU != 0 {
		appendMap(node, "mtu", scalarNode(proxy.MTU))
	}
	if proxy.RemoteDNSResolve != nil {
		appendMap(node, "remote-dns-resolve", scalarNode(*proxy.RemoteDNSResolve))
	}
	if len(proxy.DNS) > 0 {
		appendMap(node, "dns", flowStringSeqNode(proxy.DNS))
	}
	if len(proxy.Peers) > 0 {
		peers := flowSequenceNode()
		for _, peer := range proxy.Peers {
			peerNode := flowMappingNode()
			if peer.Server != "" {
				appendMap(peerNode, "server", scalarNode(peer.Server))
			}
			if peer.Port != 0 {
				appendMap(peerNode, "port", scalarNode(peer.Port))
			}
			if peer.PublicKey != "" {
				appendMap(peerNode, "public-key", scalarNode(peer.PublicKey))
			}
			if peer.PreSharedKey != "" {
				appendMap(peerNode, "pre-shared-key", scalarNode(peer.PreSharedKey))
			}
			if len(peer.AllowedIPs) > 0 {
				appendMap(peerNode, "allowed-ips", flowStringSeqNode(peer.AllowedIPs))
			}
			if peer.Reserved != nil {
				appendMap(peerNode, "reserved", nodeFromAny(peer.Reserved))
			}
			peers.Content = append(peers.Content, peerNode)
		}
		appendMap(node, "peers", peers)
	}
	if len(proxy.AmneziaWGOption) > 0 {
		appendMap(node, "amnezia-wg-option", nodeFromAny(proxy.AmneziaWGOption))
	}
	return node
}

func buildProxyGroupsNode(groups []mihomoProxyGroup) *yaml.Node {
	seq := sequenceNode()
	for _, group := range groups {
		node := flowMappingNode()
		appendMap(node, "name", scalarNode(group.Name))
		appendMap(node, "type", scalarNode(group.Type))
		appendMap(node, "proxies", flowStringSeqNode(group.Proxies))
		if group.URL != "" {
			appendMap(node, "url", scalarNode(group.URL))
		}
		if group.Interval != 0 {
			appendMap(node, "interval", scalarNode(group.Interval))
		}
		if group.Lazy != nil {
			appendMap(node, "lazy", scalarNode(*group.Lazy))
		}
		seq.Content = append(seq.Content, node)
	}
	return seq
}

func buildRuleProvidersNode(providers map[string]mihomoRuleProvider, preferredOrder []string) *yaml.Node {
	node := mappingNode()
	for _, key := range orderedProviderRenderKeys(providers, preferredOrder) {
		provider := providers[key]
		providerNode := flowMappingNode()
		appendMap(providerNode, "type", scalarNode(provider.Type))
		if provider.Behavior != "" {
			appendMap(providerNode, "behavior", scalarNode(provider.Behavior))
		}
		if provider.URL != "" {
			appendMap(providerNode, "url", scalarNode(provider.URL))
		}
		if provider.Path != "" {
			appendMap(providerNode, "path", scalarNode(provider.Path))
		}
		if provider.Interval != 0 {
			appendMap(providerNode, "interval", scalarNode(provider.Interval))
		}
		if provider.Format != "" {
			appendMap(providerNode, "format", scalarNode(provider.Format))
		}
		if provider.Proxy != "" {
			appendMap(providerNode, "proxy", scalarNode(provider.Proxy))
		}
		if provider.SizeLimit > 0 {
			appendMap(providerNode, "size-limit", scalarNode(provider.SizeLimit))
		}
		if len(provider.Header) > 0 {
			headers := flowMappingNode()
			headerKeys := make([]string, 0, len(provider.Header))
			for headerKey := range provider.Header {
				headerKeys = append(headerKeys, headerKey)
			}
			sort.Strings(headerKeys)
			for _, headerKey := range headerKeys {
				appendMap(headers, headerKey, flowStringSeqNode(provider.Header[headerKey]))
			}
			appendMap(providerNode, "header", headers)
		}
		if len(provider.Payload) > 0 {
			appendMap(providerNode, "payload", flowStringSeqNode(provider.Payload))
		}
		appendMap(node, key, providerNode)
	}
	return node
}

func buildRulesNode(rules []string) *yaml.Node {
	return stringSeqNode(rules)
}

func mappingNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

func sequenceNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
}

func flowMappingNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Style: yaml.FlowStyle}
}

func flowSequenceNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
}

func appendMap(node *yaml.Node, key string, value *yaml.Node) {
	if node == nil || value == nil {
		return
	}
	node.Content = append(node.Content, scalarStringNode(key), value)
}

func scalarStringNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func scalarNode(value interface{}) *yaml.Node {
	switch typed := value.(type) {
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: typed}
	case bool:
		if typed {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(typed)}
	case int64:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(typed, 10)}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprint(value)}
	}
}

func stringSeqNode(values []string) *yaml.Node {
	seq := sequenceNode()
	for _, value := range values {
		seq.Content = append(seq.Content, scalarNode(value))
	}
	return seq
}

func flowStringSeqNode(values []string) *yaml.Node {
	seq := flowSequenceNode()
	for _, value := range values {
		seq.Content = append(seq.Content, scalarNode(value))
	}
	return seq
}

func nodeFromAny(value interface{}) *yaml.Node {
	switch typed := value.(type) {
	case *yaml.Node:
		return typed
	case nil:
		return nil
	case string, bool, int, int64:
		return scalarNode(typed)
	case []string:
		return flowStringSeqNode(typed)
	case []int:
		seq := flowSequenceNode()
		for _, item := range typed {
			seq.Content = append(seq.Content, scalarNode(item))
		}
		return seq
	case []interface{}:
		seq := flowSequenceNode()
		for _, item := range typed {
			seq.Content = append(seq.Content, nodeFromAny(item))
		}
		return seq
	case map[string]interface{}:
		node := flowMappingNode()
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendMap(node, key, nodeFromAny(typed[key]))
		}
		return node
	default:
		return scalarNode(fmt.Sprint(value))
	}
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

func renderGeoxURL(w *yamlWriter, indent int, cfg mihomoGeoxURL) {
	if cfg.GeoIP != "" {
		w.scalar(indent, "geoip", cfg.GeoIP)
	}
	if cfg.GeoSite != "" {
		w.scalar(indent, "geosite", cfg.GeoSite)
	}
	if cfg.MMDB != "" {
		w.scalar(indent, "mmdb", cfg.MMDB)
	}
	if cfg.ASN != "" {
		w.scalar(indent, "asn", cfg.ASN)
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
	if proxy.Encryption != "" {
		w.scalar(indent+2, "encryption", proxy.Encryption)
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
	if proxy.PacketEncoding != "" {
		w.scalar(indent+2, "packet-encoding", proxy.PacketEncoding)
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
	if proxy.H2Opts != nil {
		w.line(indent+2, "h2-opts:")
		if len(proxy.H2Opts.Host) > 0 {
			w.list(indent+4, "host", proxy.H2Opts.Host)
		}
		if proxy.H2Opts.Path != "" {
			w.scalar(indent+4, "path", proxy.H2Opts.Path)
		}
	}
	if proxy.XHTTPOpts != nil {
		w.line(indent+2, "xhttp-opts:")
		if proxy.XHTTPOpts.Path != "" {
			w.scalar(indent+4, "path", proxy.XHTTPOpts.Path)
		}
		if proxy.XHTTPOpts.Mode != "" {
			w.scalar(indent+4, "mode", proxy.XHTTPOpts.Mode)
		}
		if proxy.XHTTPOpts.NoGRPCHeader != nil {
			w.scalar(indent+4, "no-grpc-header", *proxy.XHTTPOpts.NoGRPCHeader)
		}
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
	if group.Lazy != nil {
		w.scalar(indent+2, "lazy", *group.Lazy)
	}
}

func renderRuleProviders(w *yamlWriter, indent int, providers map[string]mihomoRuleProvider, preferredOrder []string) {
	keys := orderedProviderRenderKeys(providers, preferredOrder)
	for _, key := range keys {
		provider := providers[key]
		w.line(indent, fmt.Sprintf("%s:", key))
		w.scalar(indent+2, "type", provider.Type)
		if provider.Behavior != "" {
			w.scalar(indent+2, "behavior", provider.Behavior)
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
		if provider.Format != "" {
			w.scalar(indent+2, "format", provider.Format)
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

func orderedProviderRenderKeys(providers map[string]mihomoRuleProvider, preferredOrder []string) []string {
	seen := make(map[string]struct{}, len(providers))
	var keys []string
	for _, key := range preferredOrder {
		if _, ok := providers[key]; !ok {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	var rest []string
	for key := range providers {
		if _, exists := seen[key]; exists {
			continue
		}
		rest = append(rest, key)
	}
	sort.Strings(rest)
	return append(keys, rest...)
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

func (w *yamlWriter) blank() {
	if w.builder.Len() == 0 {
		return
	}
	if strings.HasSuffix(w.builder.String(), "\n\n") {
		return
	}
	w.builder.WriteByte('\n')
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
