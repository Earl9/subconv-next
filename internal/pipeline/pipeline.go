package pipeline

import (
	"context"
	"errors"
	"fmt"
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
				Name: sub.Name,
				Kind: "subscription",
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

	result.Nodes = model.NormalizeNodes(result.Nodes)
	return result
}

func RenderConfig(cfg model.Config) (RenderResult, error) {
	collected := CollectNodes(cfg)
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
