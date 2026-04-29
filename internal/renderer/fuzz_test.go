package renderer

import (
	"os"
	"path/filepath"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/parser"
)

func FuzzRenderMihomo(f *testing.F) {
	seeds := []string{
		"ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		"vless://uuid@example.com:443?type=xhttp&security=reality&sni=example.com&pbk=pub&sid=abcd#vless",
	}
	if data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "subscriptions", "protocols.txt")); err == nil {
		seeds = append(seeds, string(data))
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 64*1024 {
			t.Skip()
		}
		result := parser.ParseContent([]byte(input), model.SourceInfo{Name: "fuzz", Kind: "fuzz"})
		if len(result.Nodes) == 0 {
			return
		}
		nodes := result.Nodes
		if len(nodes) > 64 {
			nodes = nodes[:64]
		}
		_, _ = RenderMihomo(nodes, model.DefaultRenderOptions())
	})
}
