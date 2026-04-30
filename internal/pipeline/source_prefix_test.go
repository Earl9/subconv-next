package pipeline

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"subconv-next/internal/model"
	"subconv-next/internal/renderer"
)

func TestCollectNodesPreservesSubscriptionSourceNames(t *testing.T) {
	nodes := model.NormalizeNodesWithScope([]model.NodeIR{
		{
			Name:   "A",
			Type:   model.ProtocolSS,
			Server: "a.example.com",
			Port:   443,
			Source: model.SourceInfo{ID: "sub-main", Name: "主力机场", Kind: "subscription", URLHash: "aaaaaaaaaaaa"},
		},
		{
			Name:   "B",
			Type:   model.ProtocolSS,
			Server: "b.example.com",
			Port:   443,
			Source: model.SourceInfo{ID: "sub-backup", Name: "备用订阅", Kind: "subscription", URLHash: "bbbbbbbbbbbb"},
		},
	}, "none")

	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}
	if nodes[0].Source.Name == nodes[1].Source.Name {
		t.Fatalf("source names should differ: %#v", nodes)
	}
	if nodes[0].Source.ID == "" || nodes[0].Source.URLHash == "" {
		t.Fatalf("source metadata missing: %#v", nodes[0].Source)
	}
}

func TestBuildYamlNodeNameKeepsRawName(t *testing.T) {
	rawName := "[anytls]JP Osaka Oracle"
	got := model.BuildYamlNodeName(rawName, model.SourceInfo{}, model.NameOptions{
		KeepRawName:           true,
		SourcePrefixMode:      "none",
		SourcePrefixSeparator: "｜",
	})
	if got != rawName {
		t.Fatalf("BuildYamlNodeName() = %q, want raw name %q", got, rawName)
	}
}

func TestBuildYamlNodeNameEmojiNamePrefix(t *testing.T) {
	rawName := "[anytls]JP Osaka Oracle"
	got := model.BuildYamlNodeName(rawName, model.SourceInfo{Name: "SecOne", Emoji: "⚡"}, model.NameOptions{
		KeepRawName:           true,
		SourcePrefixMode:      "emoji_name",
		SourcePrefixSeparator: "｜",
	})
	if got != "⚡ SecOne｜[anytls]JP Osaka Oracle" {
		t.Fatalf("BuildYamlNodeName() = %q", got)
	}
	if !contains(got, rawName) {
		t.Fatalf("BuildYamlNodeName() did not preserve raw name: %q", got)
	}
}

func TestBuildYamlNodeNameNoProtocolRewrite(t *testing.T) {
	rawName := "[hysteria2]香港02[HY2]"
	got := model.BuildYamlNodeName(rawName, model.SourceInfo{Name: "CokeCloud", Emoji: "🔥"}, model.NameOptions{
		KeepRawName:           true,
		SourcePrefixMode:      "emoji_name",
		SourcePrefixSeparator: "｜",
	})
	if got != "🔥 CokeCloud｜[hysteria2]香港02[HY2]" {
		t.Fatalf("BuildYamlNodeName() = %q", got)
	}
}

func TestBuildYamlNodeNamePrefixModes(t *testing.T) {
	rawName := "[anytls]JP Osaka Oracle"
	source := model.SourceInfo{Name: "SecOne", Emoji: "⚡"}
	tests := []struct {
		name string
		mode string
		want string
	}{
		{name: "emoji only", mode: "emoji", want: "⚡ [anytls]JP Osaka Oracle"},
		{name: "name only", mode: "name", want: "SecOne｜[anytls]JP Osaka Oracle"},
		{name: "none", mode: "none", want: rawName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.BuildYamlNodeName(rawName, source, model.NameOptions{
				KeepRawName:           true,
				SourcePrefixMode:      tt.mode,
				SourcePrefixSeparator: "｜",
			})
			if got != tt.want {
				t.Fatalf("BuildYamlNodeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureUniqueProxyNamesUsesHashSuffix(t *testing.T) {
	nodes := []model.NodeIR{
		{Name: "⚡ SecOne｜[anytls]JP Osaka Oracle", Type: model.ProtocolSS},
		{Name: "⚡ SecOne｜[anytls]JP Osaka Oracle", Type: model.ProtocolSS},
	}
	got := model.EnsureUniqueProxyNames(nodes, "#n")
	if got[0].Name != "⚡ SecOne｜[anytls]JP Osaka Oracle" || got[1].Name != "⚡ SecOne｜[anytls]JP Osaka Oracle #2" {
		t.Fatalf("EnsureUniqueProxyNames() = %#v", got)
	}
}

func TestBuildFinalNodesDisableSourceEmoji(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.SourcePrefix = false
	cfg.Render.NameOptions.SourcePrefixMode = "none"
	rawName := "[anytls]JP Osaka Oracle"
	nodes := []model.NodeIR{validSSNode(rawName, "a.example.com", model.SourceInfo{Name: "SecOne", Emoji: "⚡"})}
	finalSet, _, err := BuildFinalNodes(cfg, model.NodeState{}, nodes)
	if err != nil {
		t.Fatalf("BuildFinalNodes() error = %v", err)
	}
	if len(finalSet.Nodes) != 1 || finalSet.Nodes[0].Name != rawName {
		t.Fatalf("final name = %#v, want %q", finalSet.Nodes, rawName)
	}
}

func TestProxyGroupsReferenceFinalUniqueNames(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.SourcePrefix = true
	cfg.Render.NameOptions = model.DefaultNameOptions()
	cfg.Render.DedupeScope = "none"
	nodes := []model.NodeIR{
		validSSNode("[anytls]JP Osaka Oracle", "a.example.com", model.SourceInfo{Name: "SecOne", Emoji: "⚡"}),
		validSSNode("[anytls]JP Osaka Oracle", "b.example.com", model.SourceInfo{Name: "SecOne", Emoji: "⚡"}),
	}
	finalSet, _, err := BuildFinalNodes(cfg, model.NodeState{}, nodes)
	if err != nil {
		t.Fatalf("BuildFinalNodes() error = %v", err)
	}
	if got, want := finalSet.Nodes[0].Name, "⚡ SecOne｜[anytls]JP Osaka Oracle"; got != want {
		t.Fatalf("first final name = %q, want %q", got, want)
	}
	if got, want := finalSet.Nodes[1].Name, "⚡ SecOne｜[anytls]JP Osaka Oracle #2"; got != want {
		t.Fatalf("second final name = %q, want %q", got, want)
	}

	rendered, err := renderer.RenderMihomo(finalSet.Nodes, renderer.OptionsFromConfig(cfg))
	if err != nil {
		t.Fatalf("RenderMihomo() error = %v", err)
	}
	var out struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
		ProxyGroups []struct {
			Name    string   `yaml:"name"`
			Proxies []string `yaml:"proxies"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(rendered, &out); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	proxyNames := map[string]struct{}{}
	for _, proxy := range out.Proxies {
		proxyNames[proxy.Name] = struct{}{}
	}
	for _, want := range []string{"⚡ SecOne｜[anytls]JP Osaka Oracle", "⚡ SecOne｜[anytls]JP Osaka Oracle #2"} {
		if _, ok := proxyNames[want]; !ok {
			t.Fatalf("rendered proxies missing %q: %#v", want, out.Proxies)
		}
	}
	foundGroupRef := false
	for _, group := range out.ProxyGroups {
		for _, proxyName := range group.Proxies {
			if proxyName == "⚡ SecOne｜[anytls]JP Osaka Oracle #2" {
				foundGroupRef = true
			}
		}
	}
	if !foundGroupRef {
		t.Fatalf("proxy-groups did not reference final #2 name: %#v", out.ProxyGroups)
	}
}

func TestDedupeGlobalMergesSources(t *testing.T) {
	nodes := model.NormalizeNodesWithScope([]model.NodeIR{
		{
			Name:   "same",
			Type:   model.ProtocolSS,
			Server: "same.example.com",
			Port:   443,
			Source: model.SourceInfo{ID: "sub-a", Name: "A", Kind: "subscription"},
		},
		{
			Name:   "same",
			Type:   model.ProtocolSS,
			Server: "same.example.com",
			Port:   443,
			Source: model.SourceInfo{ID: "sub-b", Name: "B", Kind: "subscription"},
		},
	}, "global")
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(nodes))
	}
	if len(model.MergeSourcesForView(nodes[0])) != 2 {
		t.Fatalf("expected merged sources, got %#v", nodes[0].Sources)
	}
}

func TestDedupePerSourceKeepsDuplicates(t *testing.T) {
	nodes := model.NormalizeNodesWithScope([]model.NodeIR{
		{
			Name:   "same",
			Type:   model.ProtocolSS,
			Server: "same.example.com",
			Port:   443,
			Source: model.SourceInfo{ID: "sub-a", Name: "A", Kind: "subscription"},
		},
		{
			Name:   "same",
			Type:   model.ProtocolSS,
			Server: "same.example.com",
			Port:   443,
			Source: model.SourceInfo{ID: "sub-b", Name: "B", Kind: "subscription"},
		},
	}, "per_source")
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}
	if nodes[0].ID == nodes[1].ID {
		t.Fatalf("per_source dedupe should keep different IDs per source")
	}
}

func validSSNode(name, server string, source model.SourceInfo) model.NodeIR {
	return model.NormalizeNode(model.NodeIR{
		Name:   name,
		Type:   model.ProtocolSS,
		Server: server,
		Port:   443,
		Auth:   model.Auth{Password: "password"},
		Source: source,
		Raw:    map[string]interface{}{"method": "aes-256-gcm"},
	})
}

func contains(value, needle string) bool {
	return strings.Contains(value, needle)
}
