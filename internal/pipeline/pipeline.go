package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
	"subconv-next/internal/parser"
	"subconv-next/internal/renderer"
	"subconv-next/internal/storage"
)

var ErrNoNodes = errors.New("no nodes available for rendering")

type CollectResult struct {
	Nodes    []model.NodeIR
	Warnings []string
	Errors   []parser.ParseError
}

type RenderResult struct {
	Nodes      []model.NodeIR
	NodeCount  int
	YAML       []byte
	OutputPath string
	Warnings   []string
	Errors     []parser.ParseError
}

func CollectNodes(cfg model.Config) CollectResult {
	return collectNodes(cfg, true)
}

func CollectPreviewNodes(cfg model.Config) CollectResult {
	return collectNodes(cfg, false)
}

func collectNodes(cfg model.Config, applyFilters bool) CollectResult {
	var result CollectResult

	for _, inline := range cfg.Inline {
		if !inline.Enabled || strings.TrimSpace(inline.Content) == "" {
			continue
		}

		parsed := parser.ParseContent([]byte(inline.Content), model.SourceInfo{
			ID:   inline.ID,
			Name: inline.Name,
			Kind: "inline",
		})
		result.Nodes = append(result.Nodes, parsed.Nodes...)
		result.Warnings = append(result.Warnings, parsed.Warnings...)
		result.Errors = append(result.Errors, parsed.Errors...)
	}

	enabledSubscriptions := 0
	for _, sub := range cfg.Subscriptions {
		if sub.Enabled {
			enabledSubscriptions++
		}
	}

	if enabledSubscriptions > 0 {
		remoteFetcher := fetcher.New(fetcher.OptionsFromConfig(cfg))
		for _, sub := range cfg.Subscriptions {
			if !sub.Enabled {
				continue
			}

			fetched, warnings, err := remoteFetcher.Fetch(context.Background(), fetcher.Source{
				Name:               sub.Name,
				URL:                sub.URL,
				UserAgent:          sub.UserAgent,
				Enabled:            sub.Enabled,
				InsecureSkipVerify: sub.InsecureSkipVerify,
			})
			result.Warnings = append(result.Warnings, warnings...)
			if err != nil {
				result.Errors = append(result.Errors, parser.ParseError{
					Kind:    "FETCH_FAILED",
					Message: fmt.Sprintf("%s: %v", sub.Name, err),
				})
				continue
			}

			parsed := parser.ParseContent(fetched.Content, model.SourceInfo{
				ID:      sub.ID,
				Name:    sub.Name,
				Kind:    "subscription",
				URLHash: sourceURLHash(sub.URL),
			})
			if applyFilters {
				filteredNodes, filteredCount := filterSubscriptionNodes(parsed.Nodes, sub)
				if filteredCount > 0 {
					result.Warnings = append(result.Warnings, fmt.Sprintf("%s: 已按手动过滤条件排除 %d 个节点", sub.Name, filteredCount))
				}
				result.Nodes = append(result.Nodes, filteredNodes...)
			} else {
				result.Nodes = append(result.Nodes, parsed.Nodes...)
			}
			result.Warnings = append(result.Warnings, parsed.Warnings...)
			result.Errors = append(result.Errors, parsed.Errors...)
		}
	}

	result.Nodes = model.NormalizeNodesWithScope(result.Nodes, cfg.Render.DedupeScope)
	return result
}

func RenderConfig(cfg model.Config) (RenderResult, error) {
	state, err := LoadNodeState(cfg)
	if err != nil {
		return RenderResult{}, err
	}

	collected := CollectNodesWithState(cfg, state, true, true)
	collected.Nodes = applyRenderPreferences(collected.Nodes, cfg.Render)
	if len(collected.Nodes) == 0 {
		return RenderResult{
			Warnings: collected.Warnings,
			Errors:   collected.Errors,
		}, ErrNoNodes
	}

	rendered, err := renderer.RenderMihomo(collected.Nodes, renderer.OptionsFromConfig(cfg))
	if err != nil {
		return RenderResult{
			Nodes:     collected.Nodes,
			NodeCount: len(collected.Nodes),
			Warnings:  collected.Warnings,
			Errors:    collected.Errors,
		}, err
	}

	return RenderResult{
		Nodes:      collected.Nodes,
		NodeCount:  len(collected.Nodes),
		YAML:       appendTrailingNewline(rendered),
		OutputPath: cfg.Service.OutputPath,
		Warnings:   collected.Warnings,
		Errors:     collected.Errors,
	}, nil
}

func WriteRendered(path string, data []byte) error {
	return storage.AtomicWriteFile(path, data, 0o644)
}

func appendTrailingNewline(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	return append(data, '\n')
}

func filterSubscriptionNodes(nodes []model.NodeIR, sub model.SubscriptionConfig) ([]model.NodeIR, int) {
	include := normalizeKeywords(sub.IncludeKeywords)
	exclude := normalizeKeywords(sub.ExcludeKeywords)
	manualExcluded := normalizeExactValues(sub.ExcludedNodeIDs)
	if len(include) == 0 && len(exclude) == 0 && len(manualExcluded) == 0 {
		return nodes, 0
	}

	out := make([]model.NodeIR, 0, len(nodes))
	filteredCount := 0
	for _, node := range nodes {
		if !matchesSubscriptionFilters(node, include, exclude, manualExcluded) {
			filteredCount++
			continue
		}
		out = append(out, node)
	}
	return out, filteredCount
}

func matchesSubscriptionFilters(node model.NodeIR, include, exclude, manualExcluded []string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		node.Name,
		node.Server,
		string(node.Type),
		strings.Join(node.Tags, " "),
	}, " "))

	if len(include) > 0 {
		matched := false
		for _, keyword := range include {
			if strings.Contains(haystack, keyword) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	for _, keyword := range exclude {
		if strings.Contains(haystack, keyword) {
			return false
		}
	}
	for _, id := range manualExcluded {
		if node.ID == id {
			return false
		}
	}
	return true
}

func applyRenderPreferences(nodes []model.NodeIR, render model.RenderConfig) []model.NodeIR {
	if len(nodes) == 0 {
		return nodes
	}

	includeMatcher := newNameMatcher(render.IncludeKeywords)
	excludeMatcher := newNameMatcher(render.ExcludeKeywords)

	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		if render.FilterIllegal && !isRenderableNode(node) {
			continue
		}
		if !render.IncludeInfoNode && looksLikeInfoNode(node.Name) {
			continue
		}
		if includeMatcher != nil && !includeMatcher(node.Name) {
			continue
		}
		if excludeMatcher != nil && excludeMatcher(node.Name) {
			continue
		}
		if render.SkipTLSVerify {
			node.TLS.Insecure = true
		}
		if render.UDP {
			node.UDP = model.Bool(true)
		}
		node = ApplySourcePrefix(node, render)
		if render.ShowNodeType {
			node.Name = withNodeTypePrefix(node)
		}
		if render.Emoji {
			node.Name = withEmojiPrefix(node)
		}
		out = append(out, model.NormalizeNode(node))
	}

	if render.SortNodes {
		sort.SliceStable(out, func(i, j int) bool {
			return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
		})
	}

	return model.NormalizeNodesWithScope(out, render.DedupeScope)
}

func ApplySourcePrefix(node model.NodeIR, render model.RenderConfig) model.NodeIR {
	if !render.SourcePrefix {
		return node
	}

	sourceName := strings.TrimSpace(node.Source.Name)
	name := strings.TrimSpace(node.Name)
	if sourceName == "" || name == "" {
		return node
	}
	forced := rawBool(node.Raw, "_sourcePrefixForced")
	if strings.HasPrefix(name, "["+sourceName+"]") && !forced {
		return node
	}
	if rawBool(node.Raw, "_overrideName") && !forced {
		return node
	}

	format := strings.TrimSpace(render.SourcePrefixFormat)
	if format == "" {
		format = "[{source}] {name}"
	}
	replacer := strings.NewReplacer(
		"{source}", sourceName,
		"{name}", name,
		"{type}", strings.ToLower(string(node.Type)),
		"{region}", model.NodeRegionCode(node),
	)
	node.Name = strings.TrimSpace(replacer.Replace(format))
	return node
}

func newNameMatcher(pattern string) func(string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
	}
	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString
	}
	lower := strings.ToLower(pattern)
	return func(value string) bool {
		return strings.Contains(strings.ToLower(value), lower)
	}
}

func isRenderableNode(node model.NodeIR) bool {
	if node.Type == model.ProtocolWireGuard && node.WireGuard != nil && len(node.WireGuard.Peers) > 0 {
		return true
	}
	return strings.TrimSpace(node.Server) != "" && node.Port > 0
}

func looksLikeInfoNode(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	patterns := []string{
		"剩余流量",
		"流量",
		"到期",
		"官网",
		"套餐",
		"订阅",
		"expire",
		"traffic",
		"subscription",
	}
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func withNodeTypePrefix(node model.NodeIR) string {
	name := strings.TrimSpace(node.Name)
	prefix := "[" + strings.ToLower(string(node.Type)) + "]"
	lowerName := strings.ToLower(name)
	if strings.HasPrefix(lowerName, prefix+" ") || lowerName == prefix {
		return name
	}
	if strings.HasPrefix(name, "[") {
		if end := strings.Index(name, "]"); end > 0 {
			head := strings.TrimSpace(name[:end+1])
			tail := strings.TrimSpace(name[end+1:])
			if strings.HasPrefix(strings.ToLower(tail), prefix+" ") || strings.ToLower(tail) == prefix {
				return name
			}
			return strings.TrimSpace(head + " " + prefix + " " + tail)
		}
	}
	return strings.TrimSpace(prefix + " " + name)
}

func withEmojiPrefix(node model.NodeIR) string {
	name := strings.TrimSpace(node.Name)
	if strings.HasPrefix(name, "🇭🇰") || strings.HasPrefix(name, "🇯🇵") || strings.HasPrefix(name, "🇺🇸") || strings.HasPrefix(name, "🇸🇬") ||
		strings.HasPrefix(name, "🇹🇼") || strings.HasPrefix(name, "🇰🇷") || strings.HasPrefix(name, "🇩🇪") || strings.HasPrefix(name, "🇬🇧") ||
		strings.HasPrefix(name, "🇳🇱") || strings.HasPrefix(name, "🇷🇺") ||
		strings.HasPrefix(name, "🇫🇷") || strings.HasPrefix(name, "🇨🇦") || strings.HasPrefix(name, "🇦🇺") || strings.HasPrefix(name, "🇨🇳") {
		return name
	}

	emojiByTag := map[string]string{
		"HK": "🇭🇰",
		"JP": "🇯🇵",
		"US": "🇺🇸",
		"SG": "🇸🇬",
		"TW": "🇹🇼",
		"KR": "🇰🇷",
		"DE": "🇩🇪",
		"GB": "🇬🇧",
		"NL": "🇳🇱",
		"RU": "🇷🇺",
		"FR": "🇫🇷",
		"CA": "🇨🇦",
		"AU": "🇦🇺",
		"CN": "🇨🇳",
	}
	for _, tag := range node.Tags {
		if emoji := emojiByTag[tag]; emoji != "" {
			return emoji + " " + name
		}
	}
	return name
}

func normalizeKeywords(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, item := range strings.FieldsFunc(value, func(r rune) bool {
			return r == '\n' || r == '\r' || r == ',' || r == ';'
		}) {
			item = strings.ToLower(strings.TrimSpace(item))
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}

func normalizeExactValues(values []string) []string {
	if len(values) == 0 {
		return nil
	}

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

func sourceURLHash(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])[:12]
}

func rawBool(raw map[string]interface{}, key string) bool {
	if len(raw) == 0 {
		return false
	}
	value, ok := raw[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || typed == "1"
	default:
		return false
	}
}
