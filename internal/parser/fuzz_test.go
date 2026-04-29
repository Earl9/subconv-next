package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"subconv-next/internal/model"
)

func FuzzParseSubscription(f *testing.F) {
	addParserSeeds(f)
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 64*1024 {
			t.Skip()
		}
		_ = ParseContent([]byte(input), model.SourceInfo{Name: "fuzz", Kind: "fuzz"})
	})
}

func FuzzParseNodeURL(f *testing.F) {
	addParserSeeds(f)
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 16*1024 {
			t.Skip()
		}
		for _, line := range strings.Split(input, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			_, _ = parseURI(line, model.SourceInfo{Name: "fuzz", Kind: "fuzz"})
		}
	})
}

func addParserSeeds(f *testing.F) {
	seeds := []string{
		"ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		"vless://uuid@example.com:443?type=xhttp&security=reality&sni=example.com&pbk=pub&sid=abcd#vless",
		"proxies:\n  - {name: ss, type: ss, server: example.com, port: 443, cipher: aes-256-gcm, password: pass}\n",
	}
	if data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "subscriptions", "protocols.txt")); err == nil {
		seeds = append(seeds, string(data))
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
}
