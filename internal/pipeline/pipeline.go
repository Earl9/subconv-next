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
			result.Nodes = append(result.Nodes, parsed.Nodes...)
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
			Nodes:    collected.Nodes,
			NodeCount: len(collected.Nodes),
			Warnings: collected.Warnings,
			Errors:   collected.Errors,
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
