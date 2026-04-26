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
	Service       ServiceConfig       `json:"service"`
	Subscriptions []SubscriptionConfig `json:"subscriptions"`
	Inline        []InlineConfig      `json:"inline"`
	Render        RenderConfig        `json:"render"`
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
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	URL                string `json:"url"`
	UserAgent          string `json:"user_agent"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

type InlineConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Content string `json:"content"`
}

type RenderConfig struct {
	MixedPort    int    `json:"mixed_port"`
	AllowLAN     bool   `json:"allow_lan"`
	Mode         string `json:"mode"`
	LogLevel     string `json:"log_level"`
	IPv6         bool   `json:"ipv6"`
	DNSEnabled   bool   `json:"dns_enabled"`
	EnhancedMode string `json:"enhanced_mode"`
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
		MixedPort:    DefaultMixedPort,
		Mode:         DefaultMode,
		LogLevel:     DefaultLogLevel,
		DNSEnabled:   true,
		EnhancedMode: DefaultEnhancedMode,
	}
}
