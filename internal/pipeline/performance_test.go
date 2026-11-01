package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"subconv-next/internal/model"
)

func benchmarkNodeConfig(b *testing.B, format string, count int) model.Config {
	b.Helper()
	var content strings.Builder
	if format == "yaml" {
		content.WriteString("proxies:\n")
	}
	for i := 0; i < count; i++ {
		if format == "yaml" {
			fmt.Fprintf(&content, "  - {name: 'HK node-%04d', type: vless, server: node-%d.example, port: 443, uuid: benchmark-uuid, tls: true, servername: node-%d.example, network: ws, ws-opts: {path: /proxy, headers: {Host: cdn.example}}}\n", i, i, i)
		} else {
			fmt.Fprintf(&content, "trojan://benchmark@node-%d.example:443#HK-node-%04d\n", i, i)
		}
	}
	root := b.TempDir()
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(root, "state.json")
	cfg.Service.CacheDir = filepath.Join(root, "cache")
	cfg.Service.OutputPath = filepath.Join(root, "output.yaml")
	cfg.Inline = []model.InlineConfig{{ID: "benchmark", Name: "benchmark", Enabled: true, Content: content.String()}}
	return cfg
}

func BenchmarkNodePipeline(b *testing.B) {
	for _, format := range []string{"uri", "yaml"} {
		for _, count := range []int{100, 1000} {
			b.Run(fmt.Sprintf("%s/%d", format, count), func(b *testing.B) {
				cfg := benchmarkNodeConfig(b, format, count)
				b.Run("editor", func(b *testing.B) {
					state := model.DefaultNodeState()
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						result := CollectNodesWithState(cfg, state, true, false)
						if len(result.Nodes) != count || len(result.Errors) != 0 {
							b.Fatalf("unexpected node collection: nodes=%d errors=%v", len(result.Nodes), result.Errors)
						}
					}
				})
				b.Run("generate", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						result, err := RenderConfig(cfg)
						if err != nil || result.NodeCount != count {
							b.Fatalf("unexpected generation: nodes=%d err=%v", result.NodeCount, err)
						}
					}
				})
			})
		}
	}
}
