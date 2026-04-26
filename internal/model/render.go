package model

type RenderOptions struct {
	Template     string `json:"template"`
	MixedPort    int    `json:"mixed_port"`
	AllowLAN     bool   `json:"allow_lan"`
	Mode         string `json:"mode"`
	LogLevel     string `json:"log_level"`
	IPv6         bool   `json:"ipv6"`
	DNSEnabled   bool   `json:"dns_enabled"`
	EnhancedMode string `json:"enhanced_mode"`
}

func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		Template:     DefaultTemplate,
		MixedPort:    DefaultMixedPort,
		Mode:         DefaultMode,
		LogLevel:     DefaultLogLevel,
		DNSEnabled:   true,
		EnhancedMode: DefaultEnhancedMode,
	}
}
