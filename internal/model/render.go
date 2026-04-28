package model

type RenderOptions struct {
	Template                string                   `json:"template"`
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
}

func DefaultRenderOptions() RenderOptions {
	defaults := DefaultRenderConfig()
	return RenderOptions{
		Template:                DefaultTemplate,
		MixedPort:               DefaultMixedPort,
		AllowLAN:                true,
		Mode:                    DefaultMode,
		LogLevel:                DefaultLogLevel,
		DNSEnabled:              true,
		EnhancedMode:            DefaultEnhancedMode,
		Emoji:                   true,
		ShowNodeType:            true,
		IncludeInfoNode:         true,
		UDP:                     true,
		FilterIllegal:           true,
		GroupProxyMode:          "compact",
		SourcePrefix:            true,
		SourcePrefixFormat:      "[{source}] {name}",
		SourcePrefixSeparator:   " ",
		DedupeScope:             "global",
		GeodataMode:             true,
		GeoAutoUpdate:           true,
		GeodataLoader:           "standard",
		GeoUpdateInterval:       24,
		UnifiedDelay:            true,
		TCPConcurrent:           true,
		FindProcessMode:         "strict",
		GlobalClientFingerprint: "chrome",
		OutputFilename:          "mihomo.yaml",
		TemplateRuleMode:        "rules",
		GeoxURL:                 defaults.GeoxURL,
		DNS:                     defaults.DNS,
		Profile:                 defaults.Profile,
		Sniffer:                 defaults.Sniffer,
		ExternalConfig: ExternalConfig{
			TemplateKey:   "none",
			TemplateLabel: "不选择，由接口提供方提供",
		},
		RuleMode:     "custom",
		EnabledRules: []string{},
		CustomRules:  []CustomRule{},
	}
}
