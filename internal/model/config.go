package model

import "strings"

const (
	DefaultUserAgent          = "clash.meta"
	DefaultListenAddr         = "127.0.0.1"
	DefaultListenPort         = 9876
	DefaultLogLevel           = "info"
	DefaultTemplate           = "standard"
	DefaultOutputPath         = "/data/mihomo.yaml"
	DefaultCacheDir           = "/data/cache"
	DefaultStatePath          = "/data/state.json"
	DefaultRefreshInterval    = 3600
	DefaultMaxSubscriptionB   = 5 * 1024 * 1024
	DefaultFetchTimeoutSecond = 15
	DefaultMixedPort          = 7897
	DefaultMode               = "rule"
	DefaultEnhancedMode       = "fake-ip"
)

type Config struct {
	Service            ServiceConfig        `json:"service"`
	Subscriptions      []SubscriptionConfig `json:"subscriptions"`
	Inline             []InlineConfig       `json:"inline"`
	ManualNodesEnabled *bool                `json:"manual_nodes_enabled,omitempty"`
	Render             RenderConfig         `json:"render"`
}

type ServiceConfig struct {
	Enabled                          bool   `json:"enabled"`
	ListenAddr                       string `json:"listen_addr"`
	ListenPort                       int    `json:"listen_port"`
	LogLevel                         string `json:"log_level"`
	Template                         string `json:"template"`
	OutputPath                       string `json:"output_path"`
	CacheDir                         string `json:"cache_dir"`
	StatePath                        string `json:"state_path"`
	RefreshInterval                  int    `json:"refresh_interval"`
	RefreshOnRequest                 bool   `json:"refresh_on_request"`
	StaleIfError                     bool   `json:"stale_if_error"`
	StrictMode                       bool   `json:"strict_mode"`
	WorkspaceTTLSeconds              int    `json:"workspace_ttl_seconds"`
	WorkspaceCleanupIntervalSeconds  int    `json:"workspace_cleanup_interval_seconds,omitempty"`
	PublishedDeleteIfNotAccessedDays int    `json:"published_delete_if_not_accessed_days,omitempty"`
	WorkspaceCleanupInterval         int    `json:"workspace_cleanup_interval,omitempty"`
	PublishedSubscriptionTTLSeconds  int    `json:"published_subscription_ttl_seconds,omitempty"`
	PublicBaseURL                    string `json:"public_base_url,omitempty"`
	AccessToken                      string `json:"access_token,omitempty"`
	SubscriptionToken                string `json:"subscription_token,omitempty"`
	MaxSubscriptionBytes             int    `json:"max_subscription_bytes"`
	FetchTimeoutSeconds              int    `json:"fetch_timeout_seconds"`
	AllowLAN                         bool   `json:"allow_lan"`
}

type SubscriptionConfig struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name"`
	Emoji              string   `json:"emoji,omitempty"`
	SourceLogo         string   `json:"source_logo,omitempty"`
	Enabled            bool     `json:"enabled"`
	URL                string   `json:"url"`
	UserAgent          string   `json:"user_agent"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify"`
	AllowLAN           bool     `json:"allow_lan,omitempty"`
	IncludeKeywords    []string `json:"include_keywords,omitempty"`
	ExcludeKeywords    []string `json:"exclude_keywords,omitempty"`
	ExcludedNodeIDs    []string `json:"excluded_node_ids,omitempty"`
}

type InlineConfig struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Content string `json:"content"`
}

type RenderConfig struct {
	MixedPort               int                      `json:"mixed_port"`
	AllowLAN                bool                     `json:"allow_lan"`
	Mode                    string                   `json:"mode"`
	LogLevel                string                   `json:"log_level"`
	IPv6                    bool                     `json:"ipv6"`
	DNSEnabled              bool                     `json:"dns_enabled"`
	CustomDNS               bool                     `json:"custom_dns,omitempty"`
	EnhancedMode            string                   `json:"enhanced_mode"`
	Emoji                   bool                     `json:"emoji"`
	ShowNodeType            bool                     `json:"show_node_type"`
	IncludeInfoNode         bool                     `json:"include_info_node"`
	ShowInfoNodes           bool                     `json:"show_info_nodes,omitempty"`
	SkipTLSVerify           bool                     `json:"skip_tls_verify"`
	UDP                     bool                     `json:"udp"`
	NodeList                bool                     `json:"node_list"`
	SortNodes               bool                     `json:"sort_nodes"`
	FilterIllegal           bool                     `json:"filter_illegal"`
	InsertURL               bool                     `json:"insert_url"`
	GroupProxyMode          string                   `json:"group_proxy_mode,omitempty"`
	GroupOptions            GroupOptions             `json:"group_options,omitempty"`
	SourcePrefix            bool                     `json:"source_prefix"`
	SourcePrefixFormat      string                   `json:"source_prefix_format,omitempty"`
	SourcePrefixSeparator   string                   `json:"source_prefix_separator,omitempty"`
	NameOptions             NameOptions              `json:"name_options,omitempty"`
	DedupeScope             string                   `json:"dedupe_scope,omitempty"`
	GeodataMode             bool                     `json:"geodata_mode"`
	GeoAutoUpdate           bool                     `json:"geo_auto_update"`
	GeodataLoader           string                   `json:"geodata_loader,omitempty"`
	GeoUpdateInterval       int                      `json:"geo_update_interval,omitempty"`
	GeoxURL                 *GeoxURLConfig           `json:"geox_url,omitempty"`
	IncludeKeywords         string                   `json:"include_keywords,omitempty"`
	ExcludeKeywords         string                   `json:"exclude_keywords,omitempty"`
	OutputFilename          string                   `json:"output_filename,omitempty"`
	SourceMode              string                   `json:"source_mode,omitempty"`
	TemplateRuleMode        string                   `json:"template_rule_mode"`
	ExternalConfig          ExternalConfig           `json:"external_config"`
	RuleMode                string                   `json:"rule_mode"`
	EnabledRules            []string                 `json:"enabled_rules"`
	CustomRules             []CustomRule             `json:"custom_rules"`
	UnifiedDelay            bool                     `json:"unified_delay"`
	TCPConcurrent           bool                     `json:"tcp_concurrent"`
	FindProcessMode         string                   `json:"find_process_mode,omitempty"`
	GlobalClientFingerprint string                   `json:"global_client_fingerprint,omitempty"`
	DNS                     *DNSConfig               `json:"dns,omitempty"`
	Profile                 *ProfileConfig           `json:"profile,omitempty"`
	Sniffer                 *SnifferConfig           `json:"sniffer,omitempty"`
	FinalPolicy             string                   `json:"final_policy,omitempty"`
	AdditionalRules         []string                 `json:"additional_rules"`
	RuleProviders           []RuleProviderConfig     `json:"rule_providers"`
	CustomProxyGroups       []CustomProxyGroupConfig `json:"custom_proxy_groups"`
	SubscriptionInfo        *SubscriptionInfoConfig  `json:"subscription_info,omitempty"`
}

type NameOptions struct {
	KeepRawName           bool   `json:"keep_raw_name"`
	SourcePrefixMode      string `json:"source_prefix_mode,omitempty"`
	SourcePrefixSeparator string `json:"source_prefix_separator,omitempty"`
	DedupeSuffixStyle     string `json:"dedupe_suffix_style,omitempty"`
	ShowSourceEmoji       bool   `json:"show_source_emoji,omitempty"`      // Deprecated: use SourcePrefixMode.
	SourceEmojiSeparator  string `json:"source_emoji_separator,omitempty"` // Deprecated: emoji-only mode always uses a space.
}

type GroupOptions struct {
	EnableRegionGroups           bool   `json:"enable_region_groups"`
	RuleGroupNodeMode            string `json:"rule_group_node_mode,omitempty"`
	IncludeRealNodesInRuleGroups bool   `json:"include_real_nodes_in_rule_groups"`
	SpecialGroupsUseCompact      bool   `json:"special_groups_use_compact"`
}

func DefaultGroupOptions() GroupOptions {
	return GroupOptions{
		EnableRegionGroups:           false,
		RuleGroupNodeMode:            "full",
		IncludeRealNodesInRuleGroups: true,
		SpecialGroupsUseCompact:      true,
	}
}

func NormalizeGroupOptions(opt GroupOptions) GroupOptions {
	switch opt.RuleGroupNodeMode {
	case "compact", "full":
	default:
		opt.RuleGroupNodeMode = "full"
	}
	opt.IncludeRealNodesInRuleGroups = opt.RuleGroupNodeMode == "full"
	opt.SpecialGroupsUseCompact = true
	return opt
}

type DNSConfig struct {
	Enable             bool                `json:"enable"`
	Listen             string              `json:"listen,omitempty"`
	UseHosts           bool                `json:"use_hosts"`
	UseSystemHosts     bool                `json:"use_system_hosts"`
	RespectRules       bool                `json:"respect_rules,omitempty"`
	EnhancedMode       string              `json:"enhanced_mode,omitempty"`
	FakeIPRange        string              `json:"fake_ip_range,omitempty"`
	DefaultNameserver  []string            `json:"default_nameserver,omitempty"`
	Nameserver         []string            `json:"nameserver,omitempty"`
	ProxyNameserver    []string            `json:"proxy_server_nameserver,omitempty"`
	DirectNameserver   []string            `json:"direct_nameserver,omitempty"`
	DirectFollowPolicy bool                `json:"direct_nameserver_follow_policy,omitempty"`
	Fallback           []string            `json:"fallback,omitempty"`
	FallbackFilter     *DNSFallbackFilter  `json:"fallback_filter,omitempty"`
	FakeIPFilter       []string            `json:"fake_ip_filter,omitempty"`
	NameserverPolicy   map[string][]string `json:"nameserver_policy,omitempty"`
}

type DNSFallbackFilter struct {
	GeoIP     bool     `json:"geoip"`
	GeoIPCode string   `json:"geoip_code,omitempty"`
	IPCIDR    []string `json:"ipcidr,omitempty"`
	Domain    []string `json:"domain,omitempty"`
}

type ProfileConfig struct {
	StoreSelected bool `json:"store_selected"`
	StoreFakeIP   bool `json:"store_fake_ip"`
}

type SnifferConfig struct {
	Enable      bool           `json:"enable"`
	ParsePureIP bool           `json:"parse_pure_ip"`
	TLS         *SniffProtocol `json:"tls,omitempty"`
	HTTP        *SniffHTTP     `json:"http,omitempty"`
	QUIC        *SniffProtocol `json:"quic,omitempty"`
}

type SniffProtocol struct {
	Ports []string `json:"ports,omitempty"`
}

type SniffHTTP struct {
	Ports               []string `json:"ports,omitempty"`
	OverrideDestination bool     `json:"override_destination"`
}

type RuleProviderConfig struct {
	Name      string              `json:"name"`
	Type      string              `json:"type"`
	URL       string              `json:"url,omitempty"`
	Path      string              `json:"path,omitempty"`
	Interval  int                 `json:"interval,omitempty"`
	Proxy     string              `json:"proxy,omitempty"`
	Behavior  string              `json:"behavior"`
	Format    string              `json:"format,omitempty"`
	SizeLimit int64               `json:"size_limit,omitempty"`
	Headers   map[string][]string `json:"headers,omitempty"`
	Payload   []string            `json:"payload,omitempty"`
	Policy    string              `json:"policy"`
	NoResolve bool                `json:"no_resolve,omitempty"`
	Enabled   bool                `json:"enabled"`
}

type CustomProxyGroupConfig struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Members  []string `json:"members"`
	URL      string   `json:"url,omitempty"`
	Interval int      `json:"interval,omitempty"`
	Enabled  bool     `json:"enabled"`
}

type CustomRule struct {
	Key            string   `json:"key"`
	Label          string   `json:"label"`
	Icon           string   `json:"icon,omitempty"`
	Emoji          string   `json:"emoji,omitempty"` // legacy alias
	Enabled        bool     `json:"enabled"`
	TargetMode     string   `json:"target_mode,omitempty"`
	TargetGroup    string   `json:"target_group,omitempty"`
	SourceType     string   `json:"source_type,omitempty"`
	Behavior       string   `json:"behavior,omitempty"`
	Format         string   `json:"format,omitempty"`
	URL            string   `json:"url,omitempty"`
	Path           string   `json:"path,omitempty"`
	Interval       int      `json:"interval,omitempty"`
	Payload        []string `json:"payload,omitempty"`
	InsertPosition string   `json:"insert_position,omitempty"`
	NoResolve      bool     `json:"no_resolve,omitempty"`
}

type ExternalConfig struct {
	TemplateKey   string `json:"template_key"`
	TemplateLabel string `json:"template_label"`
	CustomURL     string `json:"custom_url,omitempty"`
}

type GeoxURLConfig struct {
	GeoIP   string `json:"geoip,omitempty"`
	GeoSite string `json:"geosite,omitempty"`
	MMDB    string `json:"mmdb,omitempty"`
	ASN     string `json:"asn,omitempty"`
}

func DefaultConfig() Config {
	manualNodesEnabled := true
	return Config{
		Service:            DefaultServiceConfig(),
		ManualNodesEnabled: &manualNodesEnabled,
		Render:             DefaultRenderConfig(),
	}
}

func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Enabled:                         true,
		ListenAddr:                      DefaultListenAddr,
		ListenPort:                      DefaultListenPort,
		LogLevel:                        DefaultLogLevel,
		Template:                        DefaultTemplate,
		OutputPath:                      DefaultOutputPath,
		CacheDir:                        DefaultCacheDir,
		StatePath:                       DefaultStatePath,
		RefreshInterval:                 DefaultRefreshInterval,
		RefreshOnRequest:                true,
		StaleIfError:                    true,
		StrictMode:                      true,
		WorkspaceTTLSeconds:             86400,
		WorkspaceCleanupIntervalSeconds: 3600,
		WorkspaceCleanupInterval:        3600,
		MaxSubscriptionBytes:            DefaultMaxSubscriptionB,
		FetchTimeoutSeconds:             DefaultFetchTimeoutSecond,
	}
}

func DefaultSubscriptionConfig() SubscriptionConfig {
	return SubscriptionConfig{
		Enabled:   true,
		UserAgent: DefaultUserAgent,
	}
}

func DefaultInlineConfig() InlineConfig {
	return InlineConfig{
		Enabled: true,
	}
}

func DefaultRenderConfig() RenderConfig {
	return RenderConfig{
		MixedPort:             DefaultMixedPort,
		AllowLAN:              true,
		Mode:                  DefaultMode,
		LogLevel:              DefaultLogLevel,
		DNSEnabled:            true,
		EnhancedMode:          DefaultEnhancedMode,
		Emoji:                 false,
		ShowNodeType:          false,
		IncludeInfoNode:       false,
		ShowInfoNodes:         false,
		UDP:                   true,
		FilterIllegal:         true,
		GroupProxyMode:        "compact",
		GroupOptions:          DefaultGroupOptions(),
		SourcePrefix:          true,
		SourcePrefixFormat:    "{emoji} {name}",
		SourcePrefixSeparator: "｜",
		NameOptions:           DefaultNameOptions(),
		DedupeScope:           "global",
		OutputFilename:        "mihomo.yaml",
		SourceMode:            "rules",
		TemplateRuleMode:      "rules",
		DNS:                   DefaultDNSConfig(),
		Profile: &ProfileConfig{
			StoreSelected: true,
			StoreFakeIP:   false,
		},
		Sniffer: &SnifferConfig{
			Enable:      true,
			ParsePureIP: false,
			HTTP: &SniffHTTP{
				Ports:               []string{"80", "8080-8880"},
				OverrideDestination: true,
			},
			TLS: &SniffProtocol{
				Ports: []string{"443", "8443"},
			},
		},
		ExternalConfig: ExternalConfig{
			TemplateKey:   "none",
			TemplateLabel: "不选择，由接口提供方提供",
			CustomURL:     "",
		},
		RuleMode:                "balanced",
		EnabledRules:            []string{},
		CustomRules:             []CustomRule{},
		UnifiedDelay:            true,
		TCPConcurrent:           false,
		FindProcessMode:         "strict",
		GlobalClientFingerprint: "chrome",
		AdditionalRules:         []string{},
		RuleProviders:           []RuleProviderConfig{},
		CustomProxyGroups:       []CustomProxyGroupConfig{},
		SubscriptionInfo:        DefaultSubscriptionInfoConfig(),
	}
}

func DefaultDNSConfig() *DNSConfig {
	return &DNSConfig{
		Enable:         true,
		Listen:         "127.0.0.1:5335",
		UseHosts:       true,
		UseSystemHosts: false,
		RespectRules:   false,
		EnhancedMode:   "fake-ip",
		FakeIPRange:    "198.18.0.1/16",
		DefaultNameserver: []string{
			"223.5.5.5",
			"119.29.29.29",
		},
		Nameserver: []string{
			"223.5.5.5",
			"119.29.29.29",
		},
		Fallback: []string{
			"https://dns.alidns.com/dns-query",
		},
		FakeIPFilter: []string{
			"*.lan",
			"*.local",
			"localhost",
			"*.msftconnecttest.com",
			"*.msftncsi.com",
			"*.nintendo.net",
			"*.playstation.net",
			"*.xboxlive.com",
			"stun.*",
			"time.*",
		},
	}
}

func NormalizeDNSConfig(dns *DNSConfig, custom bool) *DNSConfig {
	if !custom {
		return DefaultDNSConfig()
	}
	defaults := DefaultDNSConfig()
	if dns == nil {
		dns = defaults
	} else {
		dns = CloneDNSConfig(dns)
	}

	dns.Enable = true
	dns.Listen = firstNonEmptyTrimmed(dns.Listen, defaults.Listen)
	dns.EnhancedMode = firstNonEmptyTrimmed(dns.EnhancedMode, defaults.EnhancedMode)
	dns.FakeIPRange = firstNonEmptyTrimmed(dns.FakeIPRange, defaults.FakeIPRange)
	dns.DefaultNameserver = trimStringList(dns.DefaultNameserver)
	if len(dns.DefaultNameserver) == 0 {
		dns.DefaultNameserver = append([]string(nil), defaults.DefaultNameserver...)
	}
	dns.Nameserver = trimStringList(dns.Nameserver)
	if len(dns.Nameserver) == 0 {
		dns.Nameserver = append([]string(nil), defaults.Nameserver...)
	}
	dns.ProxyNameserver = trimStringList(dns.ProxyNameserver)
	dns.DirectNameserver = trimStringList(dns.DirectNameserver)
	dns.Fallback = trimStringList(dns.Fallback)
	dns.FakeIPFilter = trimStringList(dns.FakeIPFilter)
	if dns.FakeIPFilter == nil {
		dns.FakeIPFilter = append([]string(nil), defaults.FakeIPFilter...)
	}
	dns.NameserverPolicy = normalizeDNSNameserverPolicy(dns.NameserverPolicy)
	if dns.FallbackFilter != nil {
		dns.FallbackFilter.GeoIPCode = strings.TrimSpace(dns.FallbackFilter.GeoIPCode)
		dns.FallbackFilter.IPCIDR = trimStringList(dns.FallbackFilter.IPCIDR)
		dns.FallbackFilter.Domain = trimStringList(dns.FallbackFilter.Domain)
	}
	return dns
}

func CloneDNSConfig(dns *DNSConfig) *DNSConfig {
	if dns == nil {
		return nil
	}
	cloned := *dns
	cloned.DefaultNameserver = append([]string(nil), dns.DefaultNameserver...)
	cloned.Nameserver = append([]string(nil), dns.Nameserver...)
	cloned.ProxyNameserver = append([]string(nil), dns.ProxyNameserver...)
	cloned.DirectNameserver = append([]string(nil), dns.DirectNameserver...)
	cloned.Fallback = append([]string(nil), dns.Fallback...)
	cloned.FakeIPFilter = append([]string(nil), dns.FakeIPFilter...)
	if dns.NameserverPolicy != nil {
		cloned.NameserverPolicy = make(map[string][]string, len(dns.NameserverPolicy))
		for key, values := range dns.NameserverPolicy {
			cloned.NameserverPolicy[key] = append([]string(nil), values...)
		}
	}
	if dns.FallbackFilter != nil {
		filter := *dns.FallbackFilter
		filter.IPCIDR = append([]string(nil), dns.FallbackFilter.IPCIDR...)
		filter.Domain = append([]string(nil), dns.FallbackFilter.Domain...)
		cloned.FallbackFilter = &filter
	}
	return &cloned
}

func firstNonEmptyTrimmed(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}

func trimStringList(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeDNSNameserverPolicy(policy map[string][]string) map[string][]string {
	if policy == nil {
		return nil
	}
	out := make(map[string][]string, len(policy))
	for key, values := range policy {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		trimmedValues := trimStringList(values)
		if len(trimmedValues) == 0 {
			continue
		}
		out[trimmedKey] = trimmedValues
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func DefaultRuleProviderConfig() RuleProviderConfig {
	return RuleProviderConfig{
		Type:     "http",
		Behavior: "classical",
		Format:   "yaml",
		Interval: 86400,
		Proxy:    "DIRECT",
		Policy:   "节点选择",
		Enabled:  true,
	}
}

func DefaultCustomProxyGroupConfig() CustomProxyGroupConfig {
	return CustomProxyGroupConfig{
		Type:     "select",
		Members:  []string{},
		Interval: 300,
		Enabled:  true,
	}
}
