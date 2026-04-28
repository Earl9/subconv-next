package renderer

import "strings"

type RuleProviderSpec struct {
	Name     string
	Type     string
	Behavior string
	URL      string
	Path     string
	Interval int
	Format   string
}

type RuleSpec struct {
	Provider    string
	TargetGroup string
	NoResolve   bool
}

type RuleCategory struct {
	Key       string
	GroupName string
	Providers []RuleProviderSpec
	Rules     []RuleSpec
}

var ruleOutputOrder = []string{
	"adblock",
	"ai",
	"youtube",
	"education",
	"cloud",
	"google",
	"private",
	"domestic",
	"telegram",
	"github",
	"microsoft",
	"apple",
	"twitter",
	"meta",
	"discord",
	"social",
	"netflix",
	"disney",
	"western_streaming",
	"asian_streaming",
	"steam",
	"pc_gaming",
	"console_gaming",
	"dev",
	"storage",
	"payment",
	"crypto",
	"news",
	"shopping",
	"non_cn",
}

var ruleModePresets = map[string][]string{
	"minimal":  {"private", "domestic", "non_cn"},
	"balanced": {"private", "domestic", "microsoft", "apple", "google", "github", "ai", "telegram", "streaming", "non_cn"},
	"full":     ruleOutputOrder,
}

var legacyRuleExpansions = map[string][]string{
	"streaming": {"netflix", "disney", "western_streaming", "asian_streaming"},
	"gaming":    {"steam", "pc_gaming", "console_gaming"},
	"finance":   {"payment", "crypto"},
	"social":    {"twitter", "meta", "discord", "social"},
	"bilibili":  {"asian_streaming"},
}

var ruleCatalog = map[string]RuleCategory{
	"adblock": {
		Key:       "adblock",
		GroupName: "🛑 广告拦截",
		Providers: []RuleProviderSpec{
			geositeProvider("category-ads-all"),
		},
		Rules: []RuleSpec{
			{Provider: "category-ads-all", TargetGroup: "🛑 广告拦截"},
		},
	},
	"ai": {
		Key:       "ai",
		GroupName: "🤖 AI 服务",
		Providers: []RuleProviderSpec{
			geositeProvider("category-ai-chat-!cn"),
			geositeProvider("openai"),
			geositeProvider("anthropic"),
		},
		Rules: []RuleSpec{
			{Provider: "category-ai-chat-!cn", TargetGroup: "🤖 AI 服务"},
			{Provider: "openai", TargetGroup: "🤖 AI 服务"},
			{Provider: "anthropic", TargetGroup: "🤖 AI 服务"},
		},
	},
	"youtube": {
		Key:       "youtube",
		GroupName: "📹 油管视频",
		Providers: []RuleProviderSpec{
			geositeProvider("youtube"),
		},
		Rules: []RuleSpec{
			{Provider: "youtube", TargetGroup: "📹 油管视频"},
		},
	},
	"education": {
		Key:       "education",
		GroupName: "📚 教育学术",
		Providers: []RuleProviderSpec{
			geositeProvider("category-scholar-!cn"),
			geositeProvider("coursera"),
			geositeProvider("udemy"),
			geositeProvider("edx"),
			geositeProvider("khanacademy"),
			geositeProvider("wikimedia"),
		},
		Rules: []RuleSpec{
			{Provider: "category-scholar-!cn", TargetGroup: "📚 教育学术"},
			{Provider: "coursera", TargetGroup: "📚 教育学术"},
			{Provider: "udemy", TargetGroup: "📚 教育学术"},
			{Provider: "edx", TargetGroup: "📚 教育学术"},
			{Provider: "khanacademy", TargetGroup: "📚 教育学术"},
			{Provider: "wikimedia", TargetGroup: "📚 教育学术"},
		},
	},
	"cloud": {
		Key:       "cloud",
		GroupName: "☁️ 云服务",
		Providers: []RuleProviderSpec{
			geositeProvider("aws"),
			geositeProvider("azure"),
			geositeProvider("cloudflare"),
			geositeProvider("digitalocean"),
			geositeProvider("vercel"),
			geositeProvider("netlify"),
			geoipProvider("cloudflare-ip", "cloudflare"),
		},
		Rules: []RuleSpec{
			{Provider: "aws", TargetGroup: "☁️ 云服务"},
			{Provider: "azure", TargetGroup: "☁️ 云服务"},
			{Provider: "cloudflare", TargetGroup: "☁️ 云服务"},
			{Provider: "digitalocean", TargetGroup: "☁️ 云服务"},
			{Provider: "vercel", TargetGroup: "☁️ 云服务"},
			{Provider: "netlify", TargetGroup: "☁️ 云服务"},
			{Provider: "cloudflare-ip", TargetGroup: "☁️ 云服务", NoResolve: true},
		},
	},
	"google": {
		Key:       "google",
		GroupName: "🔍 谷歌服务",
		Providers: []RuleProviderSpec{
			geositeProvider("google"),
			geoipProvider("google-ip", "google"),
		},
		Rules: []RuleSpec{
			{Provider: "google", TargetGroup: "🔍 谷歌服务"},
			{Provider: "google-ip", TargetGroup: "🔍 谷歌服务", NoResolve: true},
		},
	},
	"private": {
		Key:       "private",
		GroupName: "🏠 私有网络",
		Providers: []RuleProviderSpec{
			geositeProvider("private"),
			geoipProvider("private-ip", "private"),
		},
		Rules: []RuleSpec{
			{Provider: "private", TargetGroup: "🏠 私有网络"},
			{Provider: "private-ip", TargetGroup: "🏠 私有网络", NoResolve: true},
		},
	},
	"domestic": {
		Key:       "domestic",
		GroupName: "🔒 国内服务",
		Providers: []RuleProviderSpec{
			geositeProvider("geolocation-cn"),
			geoipProvider("cn-ip", "cn"),
			geositeProvider("cn"),
		},
		Rules: []RuleSpec{
			{Provider: "geolocation-cn", TargetGroup: "🔒 国内服务"},
			{Provider: "cn-ip", TargetGroup: "🔒 国内服务", NoResolve: true},
		},
	},
	"telegram": {
		Key:       "telegram",
		GroupName: "📲 电报消息",
		Providers: []RuleProviderSpec{
			geositeProvider("telegram"),
			geoipProvider("telegram-ip", "telegram"),
		},
		Rules: []RuleSpec{
			{Provider: "telegram", TargetGroup: "📲 电报消息"},
			{Provider: "telegram-ip", TargetGroup: "📲 电报消息", NoResolve: true},
		},
	},
	"github": {
		Key:       "github",
		GroupName: "🐱 代码托管",
		Providers: []RuleProviderSpec{
			geositeProvider("github"),
			geositeProvider("gitlab"),
			geositeProvider("atlassian"),
		},
		Rules: []RuleSpec{
			{Provider: "github", TargetGroup: "🐱 代码托管"},
			{Provider: "gitlab", TargetGroup: "🐱 代码托管"},
			{Provider: "atlassian", TargetGroup: "🐱 代码托管"},
		},
	},
	"microsoft": {
		Key:       "microsoft",
		GroupName: "Ⓜ️ 微软服务",
		Providers: []RuleProviderSpec{
			geositeProvider("microsoft"),
			geositeProvider("onedrive"),
		},
		Rules: []RuleSpec{
			{Provider: "microsoft", TargetGroup: "Ⓜ️ 微软服务"},
			{Provider: "onedrive", TargetGroup: "Ⓜ️ 微软服务"},
		},
	},
	"apple": {
		Key:       "apple",
		GroupName: "🍏 苹果服务",
		Providers: []RuleProviderSpec{
			geositeProvider("apple"),
			geositeProvider("icloud"),
		},
		Rules: []RuleSpec{
			{Provider: "apple", TargetGroup: "🍏 苹果服务"},
			{Provider: "icloud", TargetGroup: "🍏 苹果服务"},
		},
	},
	"twitter": {
		Key:       "twitter",
		GroupName: "🐦 推特/X",
		Providers: []RuleProviderSpec{
			geositeProvider("twitter"),
			geoipProvider("twitter-ip", "twitter"),
		},
		Rules: []RuleSpec{
			{Provider: "twitter", TargetGroup: "🐦 推特/X"},
			{Provider: "twitter-ip", TargetGroup: "🐦 推特/X", NoResolve: true},
		},
	},
	"meta": {
		Key:       "meta",
		GroupName: "📘 Meta 系",
		Providers: []RuleProviderSpec{
			geositeProvider("facebook"),
			geositeProvider("instagram"),
			geositeProvider("whatsapp"),
			geoipProvider("facebook-ip", "facebook"),
		},
		Rules: []RuleSpec{
			{Provider: "facebook", TargetGroup: "📘 Meta 系"},
			{Provider: "instagram", TargetGroup: "📘 Meta 系"},
			{Provider: "whatsapp", TargetGroup: "📘 Meta 系"},
			{Provider: "facebook-ip", TargetGroup: "📘 Meta 系", NoResolve: true},
		},
	},
	"discord": {
		Key:       "discord",
		GroupName: "🎙️ Discord",
		Providers: []RuleProviderSpec{
			geositeProvider("discord"),
		},
		Rules: []RuleSpec{
			{Provider: "discord", TargetGroup: "🎙️ Discord"},
		},
	},
	"social": {
		Key:       "social",
		GroupName: "💬 其他社交",
		Providers: []RuleProviderSpec{
			geositeProvider("tiktok"),
			geositeProvider("line"),
			geositeProvider("reddit"),
			geositeProvider("linkedin"),
			geositeProvider("snap"),
			geositeProvider("pinterest"),
			geositeProvider("tumblr"),
		},
		Rules: []RuleSpec{
			{Provider: "tiktok", TargetGroup: "💬 其他社交"},
			{Provider: "line", TargetGroup: "💬 其他社交"},
			{Provider: "reddit", TargetGroup: "💬 其他社交"},
			{Provider: "linkedin", TargetGroup: "💬 其他社交"},
			{Provider: "snap", TargetGroup: "💬 其他社交"},
			{Provider: "pinterest", TargetGroup: "💬 其他社交"},
			{Provider: "tumblr", TargetGroup: "💬 其他社交"},
		},
	},
	"netflix": {
		Key:       "netflix",
		GroupName: "🎬 奈飞",
		Providers: []RuleProviderSpec{
			geositeProvider("netflix"),
			geoipProvider("netflix-ip", "netflix"),
		},
		Rules: []RuleSpec{
			{Provider: "netflix", TargetGroup: "🎬 奈飞"},
			{Provider: "netflix-ip", TargetGroup: "🎬 奈飞", NoResolve: true},
		},
	},
	"disney": {
		Key:       "disney",
		GroupName: "🏰 迪士尼+",
		Providers: []RuleProviderSpec{
			geositeProvider("disney"),
		},
		Rules: []RuleSpec{
			{Provider: "disney", TargetGroup: "🏰 迪士尼+"},
		},
	},
	"western_streaming": {
		Key:       "western_streaming",
		GroupName: "📺 欧美流媒体",
		Providers: []RuleProviderSpec{
			geositeProvider("hbo"),
			geositeProvider("hulu"),
			geositeProvider("primevideo"),
			geositeProvider("apple-tvplus"),
			geositeProvider("spotify"),
			geositeProvider("twitch"),
			geositeProvider("dazn"),
		},
		Rules: []RuleSpec{
			{Provider: "hbo", TargetGroup: "📺 欧美流媒体"},
			{Provider: "hulu", TargetGroup: "📺 欧美流媒体"},
			{Provider: "primevideo", TargetGroup: "📺 欧美流媒体"},
			{Provider: "apple-tvplus", TargetGroup: "📺 欧美流媒体"},
			{Provider: "spotify", TargetGroup: "📺 欧美流媒体"},
			{Provider: "twitch", TargetGroup: "📺 欧美流媒体"},
			{Provider: "dazn", TargetGroup: "📺 欧美流媒体"},
		},
	},
	"asian_streaming": {
		Key:       "asian_streaming",
		GroupName: "🎌 亚洲流媒体",
		Providers: []RuleProviderSpec{
			geositeProvider("bahamut"),
			geositeProvider("biliintl"),
			geositeProvider("niconico"),
			geositeProvider("abema"),
			geositeProvider("viu"),
			geositeProvider("kktv"),
		},
		Rules: []RuleSpec{
			{Provider: "bahamut", TargetGroup: "🎌 亚洲流媒体"},
			{Provider: "biliintl", TargetGroup: "🎌 亚洲流媒体"},
			{Provider: "niconico", TargetGroup: "🎌 亚洲流媒体"},
			{Provider: "abema", TargetGroup: "🎌 亚洲流媒体"},
			{Provider: "viu", TargetGroup: "🎌 亚洲流媒体"},
			{Provider: "kktv", TargetGroup: "🎌 亚洲流媒体"},
		},
	},
	"steam": {
		Key:       "steam",
		GroupName: "🎮 Steam",
		Providers: []RuleProviderSpec{
			geositeProvider("steam"),
		},
		Rules: []RuleSpec{
			{Provider: "steam", TargetGroup: "🎮 Steam"},
		},
	},
	"pc_gaming": {
		Key:       "pc_gaming",
		GroupName: "🖥️ PC 游戏",
		Providers: []RuleProviderSpec{
			geositeProvider("epicgames"),
			geositeProvider("ea"),
			geositeProvider("ubisoft"),
			geositeProvider("blizzard"),
			geositeProvider("gog"),
			geositeProvider("riot"),
		},
		Rules: []RuleSpec{
			{Provider: "epicgames", TargetGroup: "🖥️ PC 游戏"},
			{Provider: "ea", TargetGroup: "🖥️ PC 游戏"},
			{Provider: "ubisoft", TargetGroup: "🖥️ PC 游戏"},
			{Provider: "blizzard", TargetGroup: "🖥️ PC 游戏"},
			{Provider: "gog", TargetGroup: "🖥️ PC 游戏"},
			{Provider: "riot", TargetGroup: "🖥️ PC 游戏"},
		},
	},
	"console_gaming": {
		Key:       "console_gaming",
		GroupName: "🎯 主机游戏",
		Providers: []RuleProviderSpec{
			geositeProvider("playstation"),
			geositeProvider("xbox"),
			geositeProvider("nintendo"),
		},
		Rules: []RuleSpec{
			{Provider: "playstation", TargetGroup: "🎯 主机游戏"},
			{Provider: "xbox", TargetGroup: "🎯 主机游戏"},
			{Provider: "nintendo", TargetGroup: "🎯 主机游戏"},
		},
	},
	"dev": {
		Key:       "dev",
		GroupName: "🛠️ 开发工具",
		Providers: []RuleProviderSpec{
			geositeProvider("docker"),
			geositeProvider("npmjs"),
			geositeProvider("jetbrains"),
			geositeProvider("stackexchange"),
		},
		Rules: []RuleSpec{
			{Provider: "docker", TargetGroup: "🛠️ 开发工具"},
			{Provider: "npmjs", TargetGroup: "🛠️ 开发工具"},
			{Provider: "jetbrains", TargetGroup: "🛠️ 开发工具"},
			{Provider: "stackexchange", TargetGroup: "🛠️ 开发工具"},
		},
	},
	"storage": {
		Key:       "storage",
		GroupName: "💾 网盘存储",
		Providers: []RuleProviderSpec{
			geositeProvider("dropbox"),
			geositeProvider("notion"),
		},
		Rules: []RuleSpec{
			{Provider: "dropbox", TargetGroup: "💾 网盘存储"},
			{Provider: "notion", TargetGroup: "💾 网盘存储"},
		},
	},
	"payment": {
		Key:       "payment",
		GroupName: "💳 支付平台",
		Providers: []RuleProviderSpec{
			geositeProvider("paypal"),
			geositeProvider("stripe"),
			geositeProvider("wise"),
		},
		Rules: []RuleSpec{
			{Provider: "paypal", TargetGroup: "💳 支付平台"},
			{Provider: "stripe", TargetGroup: "💳 支付平台"},
			{Provider: "wise", TargetGroup: "💳 支付平台"},
		},
	},
	"crypto": {
		Key:       "crypto",
		GroupName: "₿ 加密货币",
		Providers: []RuleProviderSpec{
			geositeProvider("binance"),
		},
		Rules: []RuleSpec{
			{Provider: "binance", TargetGroup: "₿ 加密货币"},
		},
	},
	"news": {
		Key:       "news",
		GroupName: "📰 新闻资讯",
		Providers: []RuleProviderSpec{
			geositeProvider("bbc"),
			geositeProvider("cnn"),
			geositeProvider("nytimes"),
			geositeProvider("wsj"),
			geositeProvider("bloomberg"),
		},
		Rules: []RuleSpec{
			{Provider: "bbc", TargetGroup: "📰 新闻资讯"},
			{Provider: "cnn", TargetGroup: "📰 新闻资讯"},
			{Provider: "nytimes", TargetGroup: "📰 新闻资讯"},
			{Provider: "wsj", TargetGroup: "📰 新闻资讯"},
			{Provider: "bloomberg", TargetGroup: "📰 新闻资讯"},
		},
	},
	"shopping": {
		Key:       "shopping",
		GroupName: "🛒 海淘购物",
		Providers: []RuleProviderSpec{
			geositeProvider("amazon"),
			geositeProvider("ebay"),
		},
		Rules: []RuleSpec{
			{Provider: "amazon", TargetGroup: "🛒 海淘购物"},
			{Provider: "ebay", TargetGroup: "🛒 海淘购物"},
		},
	},
	"non_cn": {
		Key:       "non_cn",
		GroupName: "🌍 非中国",
		Providers: []RuleProviderSpec{
			geositeProvider("geolocation-!cn"),
		},
		Rules: []RuleSpec{
			{Provider: "geolocation-!cn", TargetGroup: "🌍 非中国"},
		},
	},
}

func geositeProvider(name string) RuleProviderSpec {
	return RuleProviderSpec{
		Name:     name,
		Type:     "http",
		Behavior: "domain",
		URL:      "https://github.com/MetaCubeX/meta-rules-dat/raw/refs/heads/meta/geo/geosite/" + name + ".mrs",
		Path:     "./ruleset/" + name + ".mrs",
		Interval: 86400,
		Format:   "mrs",
	}
}

func geoipProvider(name, remoteName string) RuleProviderSpec {
	if strings.TrimSpace(remoteName) == "" {
		remoteName = strings.TrimSuffix(name, "-ip")
	}
	return RuleProviderSpec{
		Name:     name,
		Type:     "http",
		Behavior: "ipcidr",
		URL:      "https://github.com/MetaCubeX/meta-rules-dat/raw/refs/heads/meta/geo/geoip/" + remoteName + ".mrs",
		Path:     "./ruleset/" + name + ".mrs",
		Interval: 86400,
		Format:   "mrs",
	}
}

func orderedRuleCategories(enabledRules []string) []RuleCategory {
	enabled := make(map[string]struct{}, len(enabledRules))
	for _, key := range normalizeEnabledRuleKeys(enabledRules) {
		enabled[key] = struct{}{}
	}

	out := make([]RuleCategory, 0, len(enabled))
	for _, key := range ruleOutputOrder {
		if _, ok := enabled[key]; !ok {
			continue
		}
		if category, exists := ruleCatalog[key]; exists {
			out = append(out, category)
		}
	}
	return out
}

func resolveEnabledRules(ruleMode string, enabledRules []string) []string {
	ruleMode = strings.ToLower(strings.TrimSpace(ruleMode))
	if ruleMode == "" {
		ruleMode = "custom"
	}
	if ruleMode == "full" {
		return uniqueOrdered(ruleOutputOrder)
	}
	if preset, ok := ruleModePresets[ruleMode]; ok {
		return normalizeEnabledRuleKeys(preset)
	}
	return normalizeEnabledRuleKeys(enabledRules)
}

func orderedProviderNames(enabledRules []string) []string {
	var names []string
	seen := make(map[string]struct{})
	for _, category := range orderedRuleCategories(enabledRules) {
		for _, provider := range category.Providers {
			if _, ok := seen[provider.Name]; ok {
				continue
			}
			seen[provider.Name] = struct{}{}
			names = append(names, provider.Name)
		}
	}
	return names
}

func normalizeEnabledRuleKeys(enabledRules []string) []string {
	var out []string
	seen := make(map[string]struct{})

	var visit func(string)
	visit = func(key string) {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			return
		}
		if expanded, ok := legacyRuleExpansions[key]; ok {
			for _, item := range expanded {
				if item == key {
					if _, ok := ruleCatalog[item]; ok {
						if _, exists := seen[item]; !exists {
							seen[item] = struct{}{}
							out = append(out, item)
						}
					}
					continue
				}
				visit(item)
			}
			return
		}
		if _, ok := ruleCatalog[key]; !ok {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}

	for _, key := range enabledRules {
		visit(key)
	}
	return out
}
