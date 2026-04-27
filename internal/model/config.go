package model

const (
	DefaultUserAgent          = "SubConvNext/0.1"
	DefaultListenAddr         = "0.0.0.0"
	DefaultListenPort         = 9876
	DefaultLogLevel           = "info"
	DefaultTemplate           = "standard"
	DefaultOutputPath         = "/data/mihomo.yaml"
	DefaultCacheDir           = "/data/cache"
	DefaultStatePath          = "/data/state.json"
	DefaultRefreshInterval    = 3600
	DefaultMaxSubscriptionB   = 5 * 1024 * 1024
	DefaultFetchTimeoutSecond = 15
	DefaultMixedPort          = 7890
	DefaultMode               = "rule"
	DefaultEnhancedMode       = "fake-ip"
)

type Config struct {
	Service       ServiceConfig        `json:"service"`
	Subscriptions []SubscriptionConfig `json:"subscriptions"`
	Inline        []InlineConfig       `json:"inline"`
	Render        RenderConfig         `json:"render"`
}

type ServiceConfig struct {
	Enabled              bool   `json:"enabled"`
	ListenAddr           string `json:"listen_addr"`
	ListenPort           int    `json:"listen_port"`
	LogLevel             string `json:"log_level"`
	Template             string `json:"template"`
	OutputPath           string `json:"output_path"`
	CacheDir             string `json:"cache_dir"`
	StatePath            string `json:"state_path"`
	RefreshInterval      int    `json:"refresh_interval"`
	MaxSubscriptionBytes int    `json:"max_subscription_bytes"`
	FetchTimeoutSeconds  int    `json:"fetch_timeout_seconds"`
	AllowLAN             bool   `json:"allow_lan"`
}

type SubscriptionConfig struct {
	Name               string   `json:"name"`
	Enabled            bool     `json:"enabled"`
	URL                string   `json:"url"`
	UserAgent          string   `json:"user_agent"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify"`
	IncludeKeywords    []string `json:"include_keywords,omitempty"`
	ExcludeKeywords    []string `json:"exclude_keywords,omitempty"`
	ExcludedNodeIDs    []string `json:"excluded_node_ids,omitempty"`
}

type InlineConfig struct {
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
	EnhancedMode            string                   `json:"enhanced_mode"`
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
}

type DNSConfig struct {
	Enable            bool                `json:"enable"`
	Listen            string              `json:"listen,omitempty"`
	UseSystemHosts    bool                `json:"use_system_hosts"`
	EnhancedMode      string              `json:"enhanced_mode,omitempty"`
	FakeIPRange       string              `json:"fake_ip_range,omitempty"`
	DefaultNameserver []string            `json:"default_nameserver,omitempty"`
	Nameserver        []string            `json:"nameserver,omitempty"`
	Fallback          []string            `json:"fallback,omitempty"`
	FallbackFilter    *DNSFallbackFilter  `json:"fallback_filter,omitempty"`
	FakeIPFilter      []string            `json:"fake_ip_filter,omitempty"`
	NameserverPolicy  map[string][]string `json:"nameserver_policy,omitempty"`
}

type DNSFallbackFilter struct {
	GeoIP  bool     `json:"geoip"`
	IPCIDR []string `json:"ipcidr,omitempty"`
	Domain []string `json:"domain,omitempty"`
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

func DefaultConfig() Config {
	return Config{
		Service: DefaultServiceConfig(),
		Render:  DefaultRenderConfig(),
	}
}

func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Enabled:              true,
		ListenAddr:           DefaultListenAddr,
		ListenPort:           DefaultListenPort,
		LogLevel:             DefaultLogLevel,
		Template:             DefaultTemplate,
		OutputPath:           DefaultOutputPath,
		CacheDir:             DefaultCacheDir,
		StatePath:            DefaultStatePath,
		RefreshInterval:      DefaultRefreshInterval,
		MaxSubscriptionBytes: DefaultMaxSubscriptionB,
		FetchTimeoutSeconds:  DefaultFetchTimeoutSecond,
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
		MixedPort:         DefaultMixedPort,
		Mode:              DefaultMode,
		LogLevel:          DefaultLogLevel,
		DNSEnabled:        true,
		EnhancedMode:      DefaultEnhancedMode,
		AdditionalRules:   []string{},
		RuleProviders:     []RuleProviderConfig{},
		CustomProxyGroups: []CustomProxyGroupConfig{},
	}
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
