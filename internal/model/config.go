package model

const (
	DefaultUserAgent          = "SubConvNext/1.0"
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
	Service       ServiceConfig        `json:"service"`
	Subscriptions []SubscriptionConfig `json:"subscriptions"`
	Inline        []InlineConfig       `json:"inline"`
	Render        RenderConfig         `json:"render"`
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
	AccessToken                      string `json:"access_token,omitempty"`
	SubscriptionToken                string `json:"subscription_token,omitempty"`
	MaxSubscriptionBytes             int    `json:"max_subscription_bytes"`
	FetchTimeoutSeconds              int    `json:"fetch_timeout_seconds"`
	AllowLAN                         bool   `json:"allow_lan"`
}

type SubscriptionConfig struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name"`
	SourceLogo         string   `json:"source_logo,omitempty"`
	Enabled            bool     `json:"enabled"`
	URL                string   `json:"url"`
	UserAgent          string   `json:"user_agent"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify"`
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
	SourcePrefix            bool                     `json:"source_prefix"`
	SourcePrefixFormat      string                   `json:"source_prefix_format,omitempty"`
	SourcePrefixSeparator   string                   `json:"source_prefix_separator,omitempty"`
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
	return Config{
		Service: DefaultServiceConfig(),
		Render:  DefaultRenderConfig(),
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
		Emoji:                 true,
		ShowNodeType:          true,
		IncludeInfoNode:       false,
		ShowInfoNodes:         false,
		UDP:                   true,
		FilterIllegal:         true,
		GroupProxyMode:        "compact",
		SourcePrefix:          true,
		SourcePrefixFormat:    "[{source}] {name}",
		SourcePrefixSeparator: " ",
		DedupeScope:           "global",
		GeodataMode:           true,
		GeoAutoUpdate:         true,
		GeodataLoader:         "standard",
		GeoUpdateInterval:     24,
		OutputFilename:        "mihomo.yaml",
		SourceMode:            "rules",
		TemplateRuleMode:      "rules",
		GeoxURL: &GeoxURLConfig{
			GeoIP:   "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geoip.dat",
			GeoSite: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geosite.dat",
			MMDB:    "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/country.mmdb",
			ASN:     "https://github.com/xishang0128/geoip/releases/download/latest/GeoLite2-ASN.mmdb",
		},
		DNS: &DNSConfig{
			Enable:         true,
			Listen:         "127.0.0.1:5335",
			UseSystemHosts: false,
			EnhancedMode:   "fake-ip",
			FakeIPRange:    "198.18.0.1/16",
			DefaultNameserver: []string{
				"180.76.76.76",
				"182.254.118.118",
				"8.8.8.8",
				"180.184.2.2",
			},
			Nameserver: []string{
				"180.76.76.76",
				"119.29.29.29",
				"180.184.1.1",
				"223.5.5.5",
				"8.8.8.8",
				"https://223.6.6.6/dns-query#h3=true",
				"https://dns.alidns.com/dns-query",
				"https://cloudflare-dns.com/dns-query",
				"https://doh.pub/dns-query",
			},
			Fallback: []string{
				"https://000000.dns.nextdns.io/dns-query#h3=true",
				"https://dns.alidns.com/dns-query",
				"https://doh.pub/dns-query",
				"https://public.dns.iij.jp/dns-query",
				"https://101.101.101.101/dns-query",
				"https://208.67.220.220/dns-query",
				"tls://8.8.4.4",
				"tls://1.0.0.1:853",
				"https://cloudflare-dns.com/dns-query",
				"https://dns.google/dns-query",
			},
			FallbackFilter: &DNSFallbackFilter{
				GeoIP: true,
				IPCIDR: []string{
					"240.0.0.0/4",
					"0.0.0.0/32",
					"127.0.0.1/32",
				},
				Domain: []string{
					"+.google.com",
					"+.facebook.com",
					"+.twitter.com",
					"+.youtube.com",
					"+.xn--ngstr-lra8j.com",
					"+.google.cn",
					"+.googleapis.cn",
					"+.googleapis.com",
					"+.gvt1.com",
				},
			},
			FakeIPFilter: []string{
				"*.lan",
				"stun.*.*.*",
				"stun.*.*",
				"time.windows.com",
				"time.nist.gov",
				"time.apple.com",
				"time.asia.apple.com",
				"*.ntp.org.cn",
				"*.openwrt.pool.ntp.org",
				"time1.cloud.tencent.com",
				"time.ustc.edu.cn",
				"pool.ntp.org",
				"ntp.ubuntu.com",
				"ntp.aliyun.com",
				"ntp1.aliyun.com",
				"ntp2.aliyun.com",
				"ntp3.aliyun.com",
				"ntp4.aliyun.com",
				"ntp5.aliyun.com",
				"ntp6.aliyun.com",
				"ntp7.aliyun.com",
				"time1.aliyun.com",
				"time2.aliyun.com",
				"time3.aliyun.com",
				"time4.aliyun.com",
				"time5.aliyun.com",
				"time6.aliyun.com",
				"time7.aliyun.com",
				"*.time.edu.cn",
				"time1.apple.com",
				"time2.apple.com",
				"time3.apple.com",
				"time4.apple.com",
				"time5.apple.com",
				"time6.apple.com",
				"time7.apple.com",
				"time1.google.com",
				"time2.google.com",
				"time3.google.com",
				"time4.google.com",
				"music.163.com",
				"*.music.163.com",
				"*.126.net",
				"musicapi.taihe.com",
				"music.taihe.com",
				"songsearch.kugou.com",
				"trackercdn.kugou.com",
				"*.kuwo.cn",
				"api-jooxtt.sanook.com",
				"api.joox.com",
				"joox.com",
				"y.qq.com",
				"*.y.qq.com",
				"streamoc.music.tc.qq.com",
				"mobileoc.music.tc.qq.com",
				"isure.stream.qqmusic.qq.com",
				"dl.stream.qqmusic.qq.com",
				"aqqmusic.tc.qq.com",
				"amobile.music.tc.qq.com",
				"*.xiami.com",
				"*.music.migu.cn",
				"music.migu.cn",
				"*.msftconnecttest.com",
				"*.msftncsi.com",
				"localhost.ptlogin2.qq.com",
				"*.*.*.srv.nintendo.net",
				"*.*.stun.playstation.net",
				"xbox.*.*.microsoft.com",
				"*.ipv6.microsoft.com",
				"*.*.xboxlive.com",
				"speedtest.cros.wr.pvp.net",
			},
		},
		Profile: &ProfileConfig{
			StoreSelected: true,
			StoreFakeIP:   false,
		},
		Sniffer: &SnifferConfig{
			Enable:      true,
			ParsePureIP: true,
			HTTP: &SniffHTTP{
				Ports:               []string{"80", "8080-8880"},
				OverrideDestination: true,
			},
			QUIC: &SniffProtocol{
				Ports: []string{"443", "8443"},
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
		RuleMode:                "custom",
		EnabledRules:            []string{},
		CustomRules:             []CustomRule{},
		UnifiedDelay:            true,
		TCPConcurrent:           true,
		FindProcessMode:         "strict",
		GlobalClientFingerprint: "chrome",
		AdditionalRules:         []string{},
		RuleProviders:           []RuleProviderConfig{},
		CustomProxyGroups:       []CustomProxyGroupConfig{},
		SubscriptionInfo:        DefaultSubscriptionInfoConfig(),
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
