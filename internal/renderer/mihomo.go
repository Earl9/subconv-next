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

const mihomoProxyFieldsRawKey = "_mihomoProxyFields"

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
	Enable             bool                     `yaml:"enable"`
	Listen             string                   `yaml:"listen,omitempty"`
	UseHosts           *bool                    `yaml:"use-hosts,omitempty"`
	UseSystemHosts     *bool                    `yaml:"use-system-hosts,omitempty"`
	IPv6               bool                     `yaml:"ipv6"`
	RespectRules       *bool                    `yaml:"respect-rules,omitempty"`
	EnhancedMode       string                   `yaml:"enhanced-mode,omitempty"`
	FakeIPRange        string                   `yaml:"fake-ip-range,omitempty"`
	DefaultNameserver  []string                 `yaml:"default-nameserver,omitempty"`
	Nameserver         []string                 `yaml:"nameserver,omitempty"`
	ProxyNameserver    []string                 `yaml:"proxy-server-nameserver,omitempty"`
	DirectNameserver   []string                 `yaml:"direct-nameserver,omitempty"`
	DirectFollowPolicy bool                     `yaml:"direct-nameserver-follow-policy,omitempty"`
	Fallback           []string                 `yaml:"fallback,omitempty"`
	FallbackFilter     *mihomoDNSFallbackFilter `yaml:"fallback-filter,omitempty"`
	FakeIPFilter       []string                 `yaml:"fake-ip-filter,omitempty"`
	NameserverPolicy   map[string][]string      `yaml:"nameserver-policy,omitempty"`
}

type mihomoDNSFallbackFilter struct {
	GeoIP     bool     `yaml:"geoip"`
	GeoIPCode string   `yaml:"geoip-code,omitempty"`
	IPCIDR    []string `yaml:"ipcidr,omitempty"`
	Domain    []string `yaml:"domain,omitempty"`
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
	Multiplexing             string                 `yaml:"multiplexing,omitempty"`
	HandshakeMode            string                 `yaml:"handshake-mode,omitempty"`
	TrafficPattern           string                 `yaml:"traffic-pattern,omitempty"`
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
	Extra                    map[string]interface{} `yaml:"-"`
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
	Rule        model.CustomRule
	TargetGroup string
	CreateGroup bool
	Provider    *mihomoRuleProvider
	RuleLine    string
}

type resolvedRegionProxyGroup struct {
	Tag      string
	Name     string
	AutoName string
	Proxies  []string
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
	opts.CustomDNS = cfg.Render.CustomDNS
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
	opts.GroupOptions = cfg.Render.GroupOptions
	opts.SourcePrefix = cfg.Render.SourcePrefix
	opts.SourcePrefixFormat = cfg.Render.SourcePrefixFormat
	opts.SourcePrefixSeparator = cfg.Render.SourcePrefixSeparator
	opts.NameOptions = cfg.Render.NameOptions
	opts.DedupeScope = cfg.Render.DedupeScope
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
	opts.GroupOptions = model.NormalizeGroupOptions(opts.GroupOptions)
	if strings.TrimSpace(opts.TemplateRuleMode) == "" {
		opts.TemplateRuleMode = defaults.TemplateRuleMode
	}
	if strings.TrimSpace(opts.RuleMode) == "" {
		opts.RuleMode = defaults.RuleMode
	}
	if strings.TrimSpace(opts.ExternalConfig.TemplateKey) == "" {
		opts.ExternalConfig.TemplateKey = defaults.ExternalConfig.TemplateKey
	}
	if strings.TrimSpace(opts.ExternalConfig.TemplateLabel) == "" {
		opts.ExternalConfig.TemplateLabel = defaults.ExternalConfig.TemplateLabel
	}
	opts.GeodataMode = false
	opts.GeoAutoUpdate = false
	opts.GeodataLoader = ""
	opts.GeoUpdateInterval = 0
	opts.GeoxURL = nil
	if opts.Profile == nil {
		opts.Profile = defaults.Profile
	}
	if opts.Profile != nil {
		opts.Profile.StoreSelected = true
		opts.Profile.StoreFakeIP = false
	}
	if opts.Sniffer == nil {
		opts.Sniffer = defaults.Sniffer
	}
	if opts.Sniffer != nil {
		opts.Sniffer.Enable = true
		opts.Sniffer.ParsePureIP = false
		opts.Sniffer.QUIC = nil
	}
	opts.MixedPort = model.DefaultMixedPort
	opts.AllowLAN = true
	opts.Mode = model.DefaultMode
	opts.LogLevel = model.DefaultLogLevel
	opts.IPv6 = false
	opts.DNSEnabled = true
	opts.DNS = model.NormalizeDNSConfig(opts.DNS, opts.CustomDNS)
	opts.EnhancedMode = model.DefaultEnhancedMode
	opts.UnifiedDelay = true
	opts.TCPConcurrent = false
	opts.FindProcessMode = "strict"
	opts.GlobalClientFingerprint = "chrome"
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
	proxyGroups, err := buildProxyGroups(nodes, opts.Template, opts.CustomProxyGroups, resolvedEnabledRules, opts.CustomRules, opts.GroupProxyMode, opts.GroupOptions)
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

	return renderConfig(cfg)
}

func buildDNSConfig(opts model.RenderOptions) *mihomoDNS {
	if !opts.DNSEnabled {
		return nil
	}
	dns := model.NormalizeDNSConfig(opts.DNS, opts.CustomDNS)
	if dns == nil {
		return nil
	}

	var useSystemHosts *bool
	if opts.CustomDNS || dns.UseSystemHosts {
		useSystemHosts = boolPointer(dns.UseSystemHosts)
	}

	return &mihomoDNS{
		Enable:             dns.Enable,
		Listen:             dns.Listen,
		UseHosts:           boolPointer(dns.UseHosts),
		UseSystemHosts:     useSystemHosts,
		IPv6:               false,
		RespectRules:       boolPointer(dns.RespectRules),
		EnhancedMode:       dns.EnhancedMode,
		FakeIPRange:        dns.FakeIPRange,
		DefaultNameserver:  append([]string(nil), dns.DefaultNameserver...),
		Nameserver:         append([]string(nil), dns.Nameserver...),
		ProxyNameserver:    append([]string(nil), dns.ProxyNameserver...),
		DirectNameserver:   append([]string(nil), dns.DirectNameserver...),
		DirectFollowPolicy: dns.DirectFollowPolicy,
		Fallback:           append([]string(nil), dns.Fallback...),
		FallbackFilter:     buildDNSFallbackFilter(dns.FallbackFilter),
		FakeIPFilter:       append([]string(nil), dns.FakeIPFilter...),
		NameserverPolicy:   cloneDNSPolicy(dns.NameserverPolicy),
	}
}

func buildDNSFallbackFilter(filter *model.DNSFallbackFilter) *mihomoDNSFallbackFilter {
	if filter == nil {
		return nil
	}
	return &mihomoDNSFallbackFilter{
		GeoIP:     filter.GeoIP,
		GeoIPCode: filter.GeoIPCode,
		IPCIDR:    append([]string(nil), filter.IPCIDR...),
		Domain:    append([]string(nil), filter.Domain...),
	}
}

func cloneDNSPolicy(policy map[string][]string) map[string][]string {
	if policy == nil {
		return nil
	}
	cloned := make(map[string][]string, len(policy))
	for key, values := range policy {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
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
		Extra:  mihomoProxyExtraFields(node),
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
		proxy.PacketEncoding = normalizePacketEncoding(rawString(node.Raw, "packetEncoding"))
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
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
	case model.ProtocolHysteria:
		if rawMihomoProxyFieldString(node.Raw, "password") != "" {
			proxy.Password = node.Auth.Password
		} else if node.Auth.Password != "" && rawMihomoProxyFieldString(node.Raw, "auth-str", "auth_str", "auth") == "" {
			if proxy.Extra == nil {
				proxy.Extra = make(map[string]interface{})
			}
			proxy.Extra["auth-str"] = node.Auth.Password
		}
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
		proxy.Obfs = rawString(node.Raw, "obfs")
		proxy.ObfsParam = rawString(node.Raw, "obfsParam")
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
	case model.ProtocolMieru:
		proxy.Username = node.Auth.Username
		proxy.Password = node.Auth.Password
		proxy.PortRange = rawString(node.Raw, "portRange")
		proxy.Transport = normalizeMieruTransport(rawString(node.Raw, "transport"))
		if proxy.Transport == "" {
			proxy.Transport = "TCP"
		}
		proxy.Multiplexing = rawString(node.Raw, "multiplexing")
		proxy.HandshakeMode = rawString(node.Raw, "handshakeMode")
		proxy.TrafficPattern = rawString(node.Raw, "trafficPattern")
	case model.ProtocolHTTP, model.ProtocolSOCKS5:
		proxy.Username = node.Auth.Username
		proxy.Password = node.Auth.Password
		proxy.TLS = node.TLS.Enabled
		proxy.SNI = node.TLS.SNI
		proxy.ALPN = node.TLS.ALPN
		proxy.SkipCertVerify = boolPointerIfTrue(node.TLS.Insecure)
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

func buildProxyGroups(nodes []model.NodeIR, template string, customGroups []model.CustomProxyGroupConfig, enabledRules []string, customRules []model.CustomRule, groupProxyMode string, groupOpt model.GroupOptions) ([]mihomoProxyGroup, error) {
	groupProxyMode = strings.ToLower(strings.TrimSpace(groupProxyMode))
	if groupProxyMode == "" {
		groupProxyMode = "compact"
	}
	groupOpt = model.NormalizeGroupOptions(groupOpt)
	allNames := nodeNames(nodes)
	regularNodes := filterOutInfoNodes(nodes)
	regularNames := nodeNames(regularNodes)
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

	resolvedCustomRules := resolveCustomRules(customRules, enabledRules, nil, customGroups)
	regionProxyGroups := resolveRegionProxyGroups(regularNodes)
	if !shouldGenerateRegionProxyGroups(groupOpt) {
		regionProxyGroups = nil
	}
	regionNames := regionProxyGroupNames(regionProxyGroups)
	regionReservedNames := regionProxyGroupReservedNames(regionProxyGroups)
	if len(regionReservedNames) > 0 {
		resolvedCustomRules = resolveCustomRules(customRules, enabledRules, regionReservedNames, customGroups)
	}

	mainSelectProxies := []string{"DIRECT", "REJECT"}
	if len(regularNames) > 0 {
		mainSelectProxies = append([]string{groupAutoSelect}, mainSelectProxies...)
		mainSelectProxies = append(mainSelectProxies, regionNames...)
		mainSelectProxies = append(mainSelectProxies, regularNames...)
	}
	if err := addGroup(mihomoProxyGroup{
		Name:    groupNodeSelect,
		Type:    "select",
		Proxies: mainSelectProxies,
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
	for _, region := range regionProxyGroups {
		if err := addGroup(mihomoProxyGroup{
			Name:     region.AutoName,
			Type:     "url-test",
			Proxies:  region.Proxies,
			URL:      probeURL,
			Interval: 300,
			Lazy:     model.Bool(false),
		}); err != nil {
			return nil, err
		}
		regionSelectProxies := append([]string{region.AutoName, "DIRECT"}, region.Proxies...)
		if err := addGroup(mihomoProxyGroup{
			Name:    region.Name,
			Type:    "select",
			Proxies: regionSelectProxies,
		}); err != nil {
			return nil, err
		}
	}

	if len(enabledRules) > 0 || len(resolvedCustomRules) > 0 {
		for _, category := range orderedRuleCategories(enabledRules) {
			if err := addGroup(mihomoProxyGroup{
				Name:    category.GroupName,
				Type:    "select",
				Proxies: serviceGroupProxiesForCategory(category.Key, groupOpt, regularNames, regionNames),
			}); err != nil {
				return nil, err
			}
		}

		customGroupProxies := serviceGroupProxiesForCategory("", groupOpt, regularNames, regionNames)
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
		Proxies: serviceGroupProxiesForCategory("", groupOpt, regularNames, regionNames),
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

	return sanitizeGeneratedProxyGroups(groups, regularNames), nil
}

func shouldGenerateRegionProxyGroups(groupOpt model.GroupOptions) bool {
	return groupOpt.EnableRegionGroups
}

func resolveRegionProxyGroups(nodes []model.NodeIR) []resolvedRegionProxyGroup {
	byTag := make(map[string][]string, len(regionGroups))
	for _, node := range nodes {
		name := strings.TrimSpace(node.Name)
		if name == "" {
			continue
		}
		tag := model.NodeRegionCode(node)
		if tag == "" || tag == "OTHER" {
			continue
		}
		byTag[tag] = append(byTag[tag], name)
	}

	out := make([]resolvedRegionProxyGroup, 0, len(regionGroups))
	for _, region := range regionGroups {
		proxies := uniqueOrdered(byTag[region.Tag])
		if len(proxies) == 0 {
			continue
		}
		out = append(out, resolvedRegionProxyGroup{
			Tag:      region.Tag,
			Name:     region.Name,
			AutoName: regionAutoGroupName(region.Name),
			Proxies:  proxies,
		})
	}
	return out
}

func regionProxyGroupNames(groups []resolvedRegionProxyGroup) []string {
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		out = append(out, group.Name)
	}
	return out
}

func regionProxyGroupReservedNames(groups []resolvedRegionProxyGroup) []string {
	out := make([]string, 0, len(groups)*2)
	for _, group := range groups {
		out = append(out, group.Name)
		out = append(out, group.AutoName)
	}
	return out
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

func regionAutoGroupName(regionName string) string {
	label := strings.TrimSpace(regionName)
	fields := strings.Fields(label)
	if len(fields) > 1 {
		label = strings.Join(fields[1:], "")
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = "地区"
	}
	return "⚡ " + label + "自动"
}

func sanitizeGeneratedProxyGroups(groups []mihomoProxyGroup, regularNames []string) []mihomoProxyGroup {
	realNameSet := stringSet(regularNames)

	out := make([]mihomoProxyGroup, 0, len(groups))
	for _, group := range groups {
		if isAutoProxyGroup(group) {
			group.Proxies = filterProxyRefs(group.Proxies, func(ref string) bool {
				_, ok := realNameSet[ref]
				return ok
			})
		}
		out = append(out, group)
	}
	return out
}

func isAutoProxyGroup(group mihomoProxyGroup) bool {
	switch strings.ToLower(strings.TrimSpace(group.Type)) {
	case "url-test", "fallback", "load-balance":
		return true
	default:
		return false
	}
}

func filterProxyRefs(proxies []string, allow func(string) bool) []string {
	out := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		if proxy == "" || !allow(proxy) {
			continue
		}
		out = append(out, proxy)
	}
	return uniqueOrdered(out)
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func serviceGroupProxiesForCategory(categoryKey string, groupOpt model.GroupOptions, realNames []string, regionNames []string) []string {
	switch categoryKey {
	case "adblock":
		return []string{"REJECT", "DIRECT", groupNodeSelect}
	case "private", "domestic":
		proxies := []string{"DIRECT", "REJECT", groupNodeSelect}
		if len(realNames) > 0 {
			proxies = append(proxies, groupAutoSelect)
		}
		return proxies
	default:
		proxies := []string{groupNodeSelect}
		if len(realNames) > 0 {
			proxies = append(proxies, groupAutoSelect)
		}
		proxies = append(proxies, "DIRECT", "REJECT")
		proxies = append(proxies, regionNames...)
		if model.NormalizeGroupOptions(groupOpt).RuleGroupNodeMode == "full" {
			proxies = append(proxies, realNames...)
		}
		return uniqueOrdered(proxies)
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
			targetGroup = normalizeRulePolicyTarget(targetGroup)
			if targetGroup == groupNodeSelect {
				createGroup = true
				targetGroup = uniqueGroupName(customRuleDisplayName(rule), reserved)
				reserved[targetGroup] = struct{}{}
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
			Rule:        rule,
			TargetGroup: targetGroup,
			CreateGroup: createGroup,
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

func customRuleLines(rules []compiledCustomRule) []string {
	var out []string
	for _, rule := range rules {
		if rule.RuleLine == "" {
			continue
		}
		out = append(out, rule.RuleLine)
	}
	return out
}

func buildCustomRuleProvider(rule model.CustomRule) mihomoRuleProvider {
	sourceType := strings.ToLower(strings.TrimSpace(rule.SourceType))
	switch sourceType {
	case "http":
		behavior, format := normalizeRemoteRuleProviderShape(rule.Behavior, rule.Format, rule.URL, rule.Path)
		return mihomoRuleProvider{
			Type:     "http",
			Behavior: behavior,
			URL:      strings.TrimSpace(rule.URL),
			Path:     firstNonEmptyString(strings.TrimSpace(rule.Path), customRuleProviderPathForFormat(rule.Key, format)),
			Interval: rule.Interval,
			Format:   format,
		}
	case "file":
		return mihomoRuleProvider{
			Type:     "file",
			Behavior: strings.ToLower(strings.TrimSpace(rule.Behavior)),
			Path:     strings.TrimSpace(rule.Path),
			Format:   strings.ToLower(strings.TrimSpace(rule.Format)),
		}
	default:
		payload := normalizeRuleProviderPayload(rule.Payload)
		behavior, format := normalizeInlineRuleProviderShape(rule.Behavior, rule.Format, payload)
		return mihomoRuleProvider{
			Type:     "inline",
			Behavior: behavior,
			Format:   format,
			Payload:  payload,
		}
	}
}

func normalizeRemoteRuleProviderShape(behavior string, format string, rawURL string, path string) (string, string) {
	behavior = strings.ToLower(strings.TrimSpace(behavior))
	format = strings.ToLower(strings.TrimSpace(format))
	if behavior == "" {
		behavior = "domain"
	}
	if format == "" {
		format = inferRuleProviderFormatFromSource(rawURL, path)
	}
	if format == "" {
		format = "text"
	}
	if isClassicalYAMLRuleProviderSource(rawURL, path) {
		behavior = "classical"
		format = "yaml"
	}
	if behavior == "classical" && format == "mrs" {
		format = "yaml"
	}
	return behavior, format
}

func inferRuleProviderFormatFromSource(rawURL string, path string) string {
	source := strings.ToLower(strings.TrimSpace(firstNonEmptyString(path, rawURL)))
	switch {
	case strings.HasSuffix(source, ".mrs"):
		return "mrs"
	case strings.HasSuffix(source, ".yaml"), strings.HasSuffix(source, ".yml"):
		return "yaml"
	case strings.HasSuffix(source, ".txt"), strings.HasSuffix(source, ".text"):
		return "text"
	default:
		return ""
	}
}

func isClassicalYAMLRuleProviderSource(rawURL string, path string) bool {
	source := strings.ToLower(strings.TrimSpace(rawURL + " " + path))
	return strings.Contains(source, "/rule/clash/") || strings.Contains(source, "blackmatrix7/ios_rule_script")
}

func normalizeInlineRuleProviderShape(behavior string, format string, payload []string) (string, string) {
	behavior = strings.ToLower(strings.TrimSpace(behavior))
	format = strings.ToLower(strings.TrimSpace(format))
	if behavior == "" {
		behavior = "domain"
	}
	if format == "" {
		format = "text"
	}
	if hasClassicalRuleProviderPayload(payload) {
		behavior = "classical"
		if format == "mrs" {
			format = "text"
		}
	}
	return behavior, format
}

func hasClassicalRuleProviderPayload(payload []string) bool {
	for _, item := range payload {
		if isClassicalRuleProviderPayloadLine(item) {
			return true
		}
	}
	return false
}

func isClassicalRuleProviderPayloadLine(line string) bool {
	parts := strings.Split(strings.TrimSpace(line), ",")
	if len(parts) < 2 {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(parts[0])) {
	case "DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD", "DOMAIN-WILDCARD", "DOMAIN-REGEX",
		"GEOSITE", "IP-CIDR", "IP-CIDR6", "IP-SUFFIX", "IP-ASN", "GEOIP", "SRC-GEOIP",
		"SRC-IP-ASN", "SRC-IP-CIDR", "SRC-IP-SUFFIX", "DST-PORT", "SRC-PORT",
		"PROCESS-NAME", "PROCESS-PATH", "PROCESS-PATH-REGEX", "PROCESS-NAME-REGEX",
		"IN-TYPE", "IN-USER", "IN-NAME", "NETWORK", "UID", "SUB-RULE", "RULE-SET", "AND", "OR", "NOT":
		return true
	default:
		return false
	}
}

func normalizeRuleProviderPayload(payload []string) []string {
	var out []string
	for _, item := range payload {
		item = normalizeRuleProviderPayloadLine(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return uniqueOrdered(out)
}

func normalizeRuleProviderPayloadLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
	line = strings.Trim(line, `"'`)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	if idx := strings.Index(line, " #"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return line
	}
	targetIndex := len(parts) - 1
	if strings.EqualFold(strings.TrimSpace(parts[targetIndex]), "no-resolve") && len(parts) >= 4 {
		targetIndex--
	}
	if isRulePolicyTarget(parts[targetIndex]) {
		parts = append(parts[:targetIndex], parts[targetIndex+1:]...)
	}
	return strings.Join(parts, ",")
}

func customRuleProviderPathForFormat(key string, format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "mrs":
		return "./ruleset/" + key + ".mrs"
	case "text":
		return "./ruleset/" + key + ".txt"
	default:
		return "./ruleset/" + key + ".yaml"
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
			for _, spec := range category.Rules {
				baseRules = append(baseRules, renderRuleSet(spec))
			}
			continue
		}
		for _, spec := range category.Rules {
			baseRules = append(baseRules, renderRuleSet(spec))
		}
	}
	if containsString(enabled, "domestic") {
		baseRules = append(baseRules, renderRuleSet(RuleSpec{
			Provider:    "cn",
			TargetGroup: "🔒 国内服务",
		}))
	}

	baseRules = append(baseRules, fmt.Sprintf("MATCH,%s", normalizeFinalPolicyName(finalPolicy)))
	return mergeRules(baseRules, customRuleLines(resolvedCustomRules), additionalRules, providerRules(providers))
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

func mergeRules(baseRules, customRules, additionalRules, providerRules []string) []string {
	base := uniqueOrdered(baseRules)
	custom := normalizeRules(customRules)
	providerExtras := normalizeRules(providerRules)
	extras := normalizeRules(additionalRules)
	matchRule := ""
	if len(base) > 0 && strings.HasPrefix(strings.TrimSpace(base[len(base)-1]), "MATCH,") {
		matchRule = base[len(base)-1]
		base = base[:len(base)-1]
	}
	merged := append([]string{}, custom...)
	merged = append(merged, extras...)
	merged = append(merged, providerExtras...)
	merged = append(merged, base...)
	if matchRule != "" {
		merged = append(merged, matchRule)
	}
	return uniqueOrdered(merged)
}

func normalizeRules(rules []string) []string {
	var normalized []string
	for _, rule := range rules {
		rule = normalizeRuleLine(rule)
		if rule == "" {
			continue
		}
		normalized = append(normalized, rule)
	}
	return uniqueOrdered(normalized)
}

func normalizeRuleLine(rule string) string {
	rule = strings.TrimSpace(rule)
	if rule == "" || strings.HasPrefix(rule, "#") {
		return ""
	}
	rule = strings.TrimSpace(strings.TrimPrefix(rule, "-"))
	rule = strings.Trim(rule, `"'`)
	if rule == "" || strings.HasPrefix(rule, "#") {
		return ""
	}
	if idx := strings.Index(rule, " #"); idx >= 0 {
		rule = strings.TrimSpace(rule[:idx])
	}
	return normalizeRulePolicyAlias(strings.TrimSpace(rule))
}

func normalizeRulePolicyAlias(rule string) string {
	parts := strings.Split(rule, ",")
	if len(parts) < 3 {
		return rule
	}
	targetIndex := len(parts) - 1
	if strings.EqualFold(strings.TrimSpace(parts[targetIndex]), "no-resolve") && len(parts) >= 4 {
		targetIndex--
	}
	parts[targetIndex] = normalizeRulePolicyTarget(parts[targetIndex])
	return strings.Join(parts, ",")
}

func normalizeRulePolicyTarget(target string) string {
	switch strings.ToUpper(strings.TrimSpace(target)) {
	case "PROXY":
		return groupNodeSelect
	default:
		return strings.TrimSpace(target)
	}
}

func isRulePolicyTarget(target string) bool {
	switch strings.TrimSpace(target) {
	case "DIRECT", "REJECT", groupNodeSelect, groupAutoSelect, groupFinal:
		return true
	default:
		return strings.EqualFold(strings.TrimSpace(target), "PROXY")
	}
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

func normalizePacketEncoding(value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "xudp":
		return "xudp"
	case "packetaddr":
		return "packetaddr"
	default:
		return value
	}
}

func normalizeMieruTransport(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "tcp":
		return "TCP"
	case "udp":
		return "UDP"
	default:
		return strings.TrimSpace(value)
	}
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
	case model.ProtocolSS, model.ProtocolSSR, model.ProtocolVMess, model.ProtocolVLESS, model.ProtocolTrojan, model.ProtocolHysteria, model.ProtocolHysteria2, model.ProtocolTUIC, model.ProtocolAnyTLS, model.ProtocolWireGuard, model.ProtocolMieru, model.ProtocolSOCKS5:
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

func boolPointer(v bool) *bool {
	return &v
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

func mihomoProxyExtraFields(node model.NodeIR) map[string]interface{} {
	out := make(map[string]interface{})
	if fields := rawStringAnyMap(node.Raw, mihomoProxyFieldsRawKey); len(fields) > 0 {
		for key, value := range fields {
			key = strings.TrimSpace(key)
			if key == "" || value == nil {
				continue
			}
			out[key] = cloneAnyValue(value)
		}
	}

	if node.Type == model.ProtocolHysteria {
		addExtraFromRaw(out, node.Raw, "protocol", "protocol")
		addExtraFromRaw(out, node.Raw, "obfsParam", "obfs-param")
	}
	if node.Type == model.ProtocolHysteria || node.Type == model.ProtocolHysteria2 {
		addExtraFromRaw(out, node.Raw, "ports", "ports")
		addExtraFromRaw(out, node.Raw, "mport", "mport")
		addExtraFromRaw(out, node.Raw, "hopInterval", "hop-interval")
		addExtraFromRaw(out, node.Raw, "up", "up")
		addExtraFromRaw(out, node.Raw, "down", "down")
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func rawMihomoProxyFieldString(raw map[string]interface{}, keys ...string) string {
	fields := rawStringAnyMap(raw, mihomoProxyFieldsRawKey)
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

func addExtraFromRaw(out map[string]interface{}, raw map[string]interface{}, rawKey, yamlKey string) {
	if len(raw) == 0 {
		return
	}
	if _, exists := out[yamlKey]; exists {
		return
	}
	value, ok := raw[rawKey]
	if !ok || value == nil {
		return
	}
	out[yamlKey] = cloneAnyValue(value)
}

func rawStringAnyMap(raw map[string]interface{}, key string) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, value := range typed {
			out[fmt.Sprint(key)] = value
		}
		return out
	default:
		return nil
	}
}

func cloneAnyValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, value := range typed {
			out[key] = cloneAnyValue(value)
		}
		return out
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, value := range typed {
			out[fmt.Sprint(key)] = cloneAnyValue(value)
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			out = append(out, cloneAnyValue(item))
		}
		return out
	case []string:
		return append([]string(nil), typed...)
	case []int:
		return append([]int(nil), typed...)
	default:
		return value
	}
}

func renderConfig(cfg mihomoConfig) ([]byte, error) {
	cfg = sanitizeConfigBeforeSerialize(cfg)
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
	addSection("tcp-concurrent", scalarNode(cfg.TCPConcurrent))
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
			_ = enc.Close()
			return nil, fmt.Errorf("encode mihomo section %s: %w", section.key, err)
		}
		if err := enc.Close(); err != nil {
			return nil, fmt.Errorf("close mihomo section %s: %w", section.key, err)
		}
		out.Write(bytes.TrimSuffix(buf.Bytes(), []byte("\n")))
	}
	return []byte(unescapeYAMLUnicodeEscapes(out.String())), nil
}

func sanitizeConfigBeforeSerialize(cfg mihomoConfig) mihomoConfig {
	realProxyNames := make(map[string]struct{}, len(cfg.Proxies))
	for _, proxy := range cfg.Proxies {
		name := strings.TrimSpace(proxy.Name)
		if name != "" {
			realProxyNames[name] = struct{}{}
		}
	}
	cfg.ProxyGroups = sanitizeProxyGroupsBeforeSerialize(cfg.ProxyGroups, realProxyNames)
	return cfg
}

func sanitizeProxyGroupsBeforeSerialize(groups []mihomoProxyGroup, realProxyNames map[string]struct{}) []mihomoProxyGroup {
	out := make([]mihomoProxyGroup, 0, len(groups))
	for _, group := range groups {
		if isAutoProxyGroup(group) {
			group.Proxies = filterProxyRefs(group.Proxies, func(ref string) bool {
				_, ok := realProxyNames[ref]
				return ok
			})
		}
		out = append(out, group)
	}
	return out
}

func buildDNSNode(dns mihomoDNS) *yaml.Node {
	node := mappingNode()
	appendMap(node, "enable", scalarNode(dns.Enable))
	if dns.Listen != "" {
		appendMap(node, "listen", scalarNode(dns.Listen))
	}
	if dns.UseHosts != nil {
		appendMap(node, "use-hosts", scalarNode(*dns.UseHosts))
	}
	if dns.UseSystemHosts != nil {
		appendMap(node, "use-system-hosts", scalarNode(*dns.UseSystemHosts))
	}
	appendMap(node, "ipv6", scalarNode(dns.IPv6))
	if dns.RespectRules != nil {
		appendMap(node, "respect-rules", scalarNode(*dns.RespectRules))
	}
	if dns.EnhancedMode != "" {
		appendMap(node, "enhanced-mode", scalarNode(dns.EnhancedMode))
	}
	if len(dns.DefaultNameserver) > 0 {
		appendMap(node, "default-nameserver", flowStringSeqNode(dns.DefaultNameserver))
	}
	if len(dns.NameserverPolicy) > 0 {
		policy := mappingNode()
		keys := make([]string, 0, len(dns.NameserverPolicy))
		for key := range dns.NameserverPolicy {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendMap(policy, key, dnsPolicyValueNode(dns.NameserverPolicy[key]))
		}
		appendMap(node, "nameserver-policy", policy)
	}
	if len(dns.Nameserver) > 0 {
		appendMap(node, "nameserver", flowStringSeqNode(dns.Nameserver))
	}
	if len(dns.ProxyNameserver) > 0 {
		appendMap(node, "proxy-server-nameserver", flowStringSeqNode(dns.ProxyNameserver))
	}
	if len(dns.DirectNameserver) > 0 {
		appendMap(node, "direct-nameserver", flowStringSeqNode(dns.DirectNameserver))
	}
	if dns.DirectFollowPolicy {
		appendMap(node, "direct-nameserver-follow-policy", scalarNode(dns.DirectFollowPolicy))
	}
	if dns.FakeIPRange != "" {
		appendMap(node, "fake-ip-range", scalarNode(dns.FakeIPRange))
	}
	if len(dns.Fallback) > 0 {
		appendMap(node, "fallback", flowStringSeqNode(dns.Fallback))
	}
	if dns.FallbackFilter != nil {
		filter := flowMappingNode()
		appendMap(filter, "geoip", scalarNode(dns.FallbackFilter.GeoIP))
		if dns.FallbackFilter.GeoIPCode != "" {
			appendMap(filter, "geoip-code", scalarNode(dns.FallbackFilter.GeoIPCode))
		}
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
	return node
}

func dnsPolicyValueNode(values []string) *yaml.Node {
	if len(values) == 1 {
		return scalarNode(values[0])
	}
	return stringSeqNode(values)
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
	for _, key := range []string{"HTTP", "TLS"} {
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
	if proxy.PortRange != "" {
		appendMap(node, "port-range", scalarNode(proxy.PortRange))
	}
	if proxy.Transport != "" {
		appendMap(node, "transport", scalarNode(proxy.Transport))
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
	if proxy.Multiplexing != "" {
		appendMap(node, "multiplexing", scalarNode(proxy.Multiplexing))
	}
	if proxy.HandshakeMode != "" {
		appendMap(node, "handshake-mode", scalarNode(proxy.HandshakeMode))
	}
	if proxy.TrafficPattern != "" {
		appendMap(node, "traffic-pattern", scalarNode(proxy.TrafficPattern))
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
	appendProxyExtraFields(node, proxy.Extra)
	return node
}

func appendProxyExtraFields(node *yaml.Node, extra map[string]interface{}) {
	if len(extra) == 0 {
		return
	}
	keys := make([]string, 0, len(extra))
	for key := range extra {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if mappingNodeHasKey(node, key) {
			continue
		}
		appendMap(node, key, nodeFromAny(extra[key]))
	}
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
	seq := sequenceNode()
	for _, rule := range rules {
		seq.Content = append(seq.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: rule,
			Style: yaml.DoubleQuotedStyle,
		})
	}
	return seq
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

func mappingNodeHasKey(node *yaml.Node, key string) bool {
	if node == nil || node.Kind != yaml.MappingNode {
		return false
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return true
		}
	}
	return false
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
	case float64:
		if typed == float64(int64(typed)) {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(int64(typed), 10)}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(typed, 'f', -1, 64)}
	case float32:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(float64(typed), 'f', -1, 32)}
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
	case string, bool, int, int64, float64, float32:
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
	if dns.UseHosts != nil {
		w.scalar(indent, "use-hosts", *dns.UseHosts)
	}
	if dns.UseSystemHosts != nil {
		w.scalar(indent, "use-system-hosts", *dns.UseSystemHosts)
	}
	w.scalar(indent, "ipv6", dns.IPv6)
	if dns.RespectRules != nil {
		w.scalar(indent, "respect-rules", *dns.RespectRules)
	}
	if dns.EnhancedMode != "" {
		w.scalar(indent, "enhanced-mode", dns.EnhancedMode)
	}
	if len(dns.DefaultNameserver) > 0 {
		w.line(indent, "default-nameserver:")
		for _, item := range dns.DefaultNameserver {
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
			values := dns.NameserverPolicy[key]
			if len(values) == 1 {
				w.line(indent+4, fmt.Sprintf("%s: %s", yamlString(key), yamlString(values[0])))
				continue
			}
			w.line(indent+4, fmt.Sprintf("%s:", yamlString(key)))
			for _, item := range values {
				w.listScalar(indent+8, item)
			}
		}
	}
	if len(dns.Nameserver) > 0 {
		w.line(indent, "nameserver:")
		for _, item := range dns.Nameserver {
			w.listScalar(indent+4, item)
		}
	}
	if len(dns.ProxyNameserver) > 0 {
		w.line(indent, "proxy-server-nameserver:")
		for _, item := range dns.ProxyNameserver {
			w.listScalar(indent+4, item)
		}
	}
	if len(dns.DirectNameserver) > 0 {
		w.line(indent, "direct-nameserver:")
		for _, item := range dns.DirectNameserver {
			w.listScalar(indent+4, item)
		}
	}
	if dns.DirectFollowPolicy {
		w.scalar(indent, "direct-nameserver-follow-policy", dns.DirectFollowPolicy)
	}
	if dns.FakeIPRange != "" {
		w.scalar(indent, "fake-ip-range", dns.FakeIPRange)
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
		if dns.FallbackFilter.GeoIPCode != "" {
			w.scalar(indent+4, "geoip-code", dns.FallbackFilter.GeoIPCode)
		}
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
	for _, key := range []string{"TLS", "HTTP"} {
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
	if proxy.PortRange != "" {
		w.scalar(indent+2, "port-range", proxy.PortRange)
	}
	if proxy.Transport != "" {
		w.scalar(indent+2, "transport", proxy.Transport)
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
	if proxy.Multiplexing != "" {
		w.scalar(indent+2, "multiplexing", proxy.Multiplexing)
	}
	if proxy.HandshakeMode != "" {
		w.scalar(indent+2, "handshake-mode", proxy.HandshakeMode)
	}
	if proxy.TrafficPattern != "" {
		w.scalar(indent+2, "traffic-pattern", proxy.TrafficPattern)
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
