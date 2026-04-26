package renderer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/parser"
)

func TestRenderGoldenLiteBasic(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "ss.txt"))
	got := mustRender(t, nodes, model.RenderOptions{
		Template:     "lite",
		MixedPort:    7890,
		Mode:         "rule",
		LogLevel:     "info",
		DNSEnabled:   true,
		EnhancedMode: "fake-ip",
	})
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "lite-basic.yaml"), got)
}

func TestRenderGoldenStandardVLESSReality(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "vless-reality.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-vless-reality.yaml"), got)
}

func TestRenderGoldenStandardHy2TUIC(t *testing.T) {
	nodes := append(
		mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "hy2.txt")),
		mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "tuic-v5.txt"))...,
	)
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-hy2-tuic.yaml"), got)
}

func TestRenderGoldenStandardAnyTLS(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "anytls.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-anytls.yaml"), got)
}

func TestRenderGoldenStandardWireGuard(t *testing.T) {
	nodes := mustParseFile(t, filepath.Join("..", "..", "testdata", "nodes", "wireguard-uri.txt"))
	got := mustRender(t, nodes, standardRenderOptions())
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "standard-wireguard.yaml"), got)
}

func TestRenderGoldenDedupeRenamed(t *testing.T) {
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{
			Name:   "dup",
			Type:   model.ProtocolTrojan,
			Server: "one.example.com",
			Port:   443,
			Auth:   model.Auth{Password: "a"},
			TLS:    model.TLSOptions{Enabled: true, SNI: "one.example.com"},
			UDP:    model.Bool(true),
		}),
		model.NormalizeNode(model.NodeIR{
			Name:   "dup",
			Type:   model.ProtocolTrojan,
			Server: "two.example.com",
			Port:   443,
			Auth:   model.Auth{Password: "b"},
			TLS:    model.TLSOptions{Enabled: true, SNI: "two.example.com"},
			UDP:    model.Bool(true),
		}),
	}

	got := mustRender(t, nodes, model.RenderOptions{
		Template:   "lite",
		MixedPort:  7890,
		Mode:       "rule",
		LogLevel:   "info",
		DNSEnabled: false,
	})
	assertGoldenFile(t, filepath.Join("..", "..", "testdata", "golden", "dedupe-renamed.yaml"), got)
}

func mustParseFile(t *testing.T, path string) []model.NodeIR {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	result := parser.ParseContent(content, model.SourceInfo{Name: filepath.Base(path), Kind: "test"})
	if len(result.Errors) != 0 {
		t.Fatalf("ParseContent(%q) errors = %#v", path, result.Errors)
	}
	return result.Nodes
}

func mustRender(t *testing.T, nodes []model.NodeIR, opts model.RenderOptions) []byte {
	t.Helper()

	got, err := RenderMihomo(nodes, opts)
	if err != nil {
		t.Fatalf("RenderMihomo() error = %v", err)
	}
	return got
}

func assertGoldenFile(t *testing.T, path string, got []byte) {
	t.Helper()

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func standardRenderOptions() model.RenderOptions {
	return model.RenderOptions{
		Template:     "standard",
		MixedPort:    7890,
		Mode:         "rule",
		LogLevel:     "info",
		DNSEnabled:   true,
		EnhancedMode: "fake-ip",
	}
}
