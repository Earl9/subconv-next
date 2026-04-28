package templatecatalog

import "strings"

type Preset struct {
	Key            string
	Group          string
	Label          string
	Description    string
	Template       string
	RuleMode       string
	EnabledRules   []string
	GroupProxyMode string
}

var presetList = []Preset{
	{
		Key:            "subconv_lite",
		Group:          "SubConv Next",
		Label:          "SubConv Next Lite",
		Description:    "只生成基础代理组，不展开业务分流规则。",
		Template:       "lite",
		RuleMode:       "custom",
		EnabledRules:   []string{},
		GroupProxyMode: "compact",
	},
	{
		Key:            "subconv_standard",
		Group:          "SubConv Next",
		Label:          "SubConv Next Standard",
		Description:    "内置标准模板，规则覆盖日常常用服务，代理组保持紧凑。",
		Template:       "standard",
		RuleMode:       "balanced",
		GroupProxyMode: "compact",
	},
	{
		Key:            "subconv_full",
		Group:          "SubConv Next",
		Label:          "SubConv Next Full",
		Description:    "完整规则覆盖，并为业务组追加完整节点选择。",
		Template:       "full",
		RuleMode:       "full",
		GroupProxyMode: "full",
	},
	{
		Key:            "cm_online",
		Group:          "CM_Online 兼容",
		Label:          "CM_Online 默认版",
		Description:    "偏向日常分流，覆盖 Google、Telegram、苹果、微软和基础国际流量。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"private", "domestic", "google", "youtube", "telegram", "microsoft", "apple", "social", "non_cn"},
		GroupProxyMode: "compact",
	},
	{
		Key:            "cm_online_game",
		Group:          "CM_Online 兼容",
		Label:          "CM_Online_Game",
		Description:    "在默认分流基础上补充游戏平台和常见流媒体。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"private", "domestic", "google", "youtube", "telegram", "microsoft", "apple", "social", "streaming", "gaming", "non_cn"},
		GroupProxyMode: "compact",
	},
	{
		Key:            "cm_online_multi_country",
		Group:          "CM_Online 兼容",
		Label:          "CM_Online_MultiCountry",
		Description:    "启用地区组，适合按国家快速切换节点来源。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"private", "domestic", "google", "youtube", "telegram", "microsoft", "apple", "github", "ai", "streaming", "social", "non_cn"},
		GroupProxyMode: "regional",
	},
	{
		Key:            "cm_online_multi_country_cf",
		Group:          "CM_Online 兼容",
		Label:          "CM_Online_MultiCountry_CF",
		Description:    "多地区模板，额外补充云服务分流，适合 Worker / CDN 相关节点。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"private", "domestic", "google", "youtube", "telegram", "microsoft", "apple", "github", "ai", "streaming", "social", "cloud", "non_cn"},
		GroupProxyMode: "regional",
	},
	{
		Key:            "cm_online_full",
		Group:          "CM_Online 兼容",
		Label:          "CM_Online_Full",
		Description:    "完整规则集和全量业务组节点选择。",
		Template:       "full",
		RuleMode:       "full",
		GroupProxyMode: "full",
	},
	{
		Key:            "cm_online_full_cf",
		Group:          "CM_Online 兼容",
		Label:          "CM_Online_Full_CF",
		Description:    "完整规则集，附带云服务分流和全量节点选择。",
		Template:       "full",
		RuleMode:       "custom",
		EnabledRules:   []string{"adblock", "microsoft", "ai", "apple", "bilibili", "social", "youtube", "streaming", "google", "gaming", "private", "education", "domestic", "finance", "telegram", "cloud", "github", "non_cn"},
		GroupProxyMode: "full",
	},
	{
		Key:            "acl4ssr_online_mini",
		Group:          "ACL4SSR 兼容",
		Label:          "ACL4SSR_Online_Mini",
		Description:    "精简国际分流模板，只保留基础常用业务。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"private", "domestic", "google", "telegram", "non_cn"},
		GroupProxyMode: "compact",
	},
	{
		Key:            "acl4ssr_online",
		Group:          "ACL4SSR 兼容",
		Label:          "ACL4SSR_Online",
		Description:    "兼顾广告拦截、常用国际服务和流媒体的通用模板。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"adblock", "private", "domestic", "google", "youtube", "telegram", "social", "streaming", "non_cn"},
		GroupProxyMode: "compact",
	},
	{
		Key:            "acl4ssr_online_adblock",
		Group:          "ACL4SSR 兼容",
		Label:          "ACL4SSR_Online_AdblockPlus",
		Description:    "在通用模板基础上扩展广告和更多业务分流。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"adblock", "private", "domestic", "google", "youtube", "telegram", "social", "streaming", "gaming", "finance", "github", "non_cn"},
		GroupProxyMode: "compact",
	},
	{
		Key:            "acl4ssr_online_full",
		Group:          "ACL4SSR 兼容",
		Label:          "ACL4SSR_Online_Full",
		Description:    "完整业务分流模板，适合偏向大而全的配置风格。",
		Template:       "full",
		RuleMode:       "full",
		GroupProxyMode: "full",
	},
	{
		Key:            "blackmatrix7_basic",
		Group:          "BlackMatrix7 风格",
		Label:          "BlackMatrix7 Basic",
		Description:    "偏向开发、云服务和基础国际业务分流。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"adblock", "private", "domestic", "google", "github", "telegram", "cloud", "non_cn"},
		GroupProxyMode: "compact",
	},
	{
		Key:            "blackmatrix7_streaming",
		Group:          "BlackMatrix7 风格",
		Label:          "BlackMatrix7 Streaming",
		Description:    "以视频、音乐和国际访问为主的模板。",
		Template:       "standard",
		RuleMode:       "custom",
		EnabledRules:   []string{"private", "domestic", "google", "youtube", "streaming", "social", "non_cn"},
		GroupProxyMode: "regional",
	},
	{
		Key:            "blackmatrix7_global",
		Group:          "BlackMatrix7 风格",
		Label:          "BlackMatrix7 Global",
		Description:    "全量规则和全量节点选择，适合重度分流场景。",
		Template:       "full",
		RuleMode:       "full",
		GroupProxyMode: "full",
	},
}

var presetByKey = func() map[string]Preset {
	out := make(map[string]Preset, len(presetList))
	for _, preset := range presetList {
		out[preset.Key] = clonePreset(preset)
	}
	return out
}()

func Lookup(key string) (Preset, bool) {
	preset, ok := presetByKey[normalizeKey(key)]
	if !ok {
		return Preset{}, false
	}
	return clonePreset(preset), true
}

func Resolve(key, serviceTemplate string) Preset {
	key = normalizeKey(key)
	switch key {
	case "", "none", "custom_url":
		return DefaultForServiceTemplate(serviceTemplate)
	default:
		if preset, ok := Lookup(key); ok {
			return preset
		}
		return DefaultForServiceTemplate(serviceTemplate)
	}
}

func DefaultForServiceTemplate(serviceTemplate string) Preset {
	switch normalizeKey(serviceTemplate) {
	case "lite":
		return clonePreset(presetByKey["subconv_lite"])
	case "full":
		return clonePreset(presetByKey["subconv_full"])
	default:
		return clonePreset(presetByKey["subconv_standard"])
	}
}

func IsKnownKey(key string) bool {
	switch normalizeKey(key) {
	case "", "none", "custom_url":
		return true
	default:
		_, ok := presetByKey[normalizeKey(key)]
		return ok
	}
}

func clonePreset(preset Preset) Preset {
	preset.EnabledRules = append([]string(nil), preset.EnabledRules...)
	return preset
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
