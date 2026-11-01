package pipeline

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"

	"subconv-next/internal/model"
	"subconv-next/internal/parser"
	"subconv-next/internal/renderer"
)

func TestHysteria2ObfsNormalizationThroughRendering(t *testing.T) {
	for _, format := range []string{"uri", "yaml"} {
		for _, obfs := range []string{"none", "NoNe", "SALAMANDER", "GeCkO"} {
			t.Run(format+"/"+obfs, func(t *testing.T) {
				content := fmt.Sprintf("hy2://test@hy2.example:443?obfs=%s&obfs-password=secret#test", obfs)
				if format == "yaml" {
					content = fmt.Sprintf("proxies:\n  - {name: test, type: hysteria2, server: hy2.example, port: 443, password: test, obfs: %s, obfs-password: secret}\n", obfs)
				}
				parsed := parser.ParseContent([]byte(content), model.SourceInfo{})
				cfg := model.DefaultConfig()
				final, _, err := BuildFinalNodes(cfg, model.DefaultNodeState(), parsed.Nodes)
				if err != nil || len(final.Nodes) != 1 {
					t.Fatalf("valid obfs alias was filtered: nodes=%d, err=%v", len(final.Nodes), err)
				}
				output, err := renderer.RenderMihomo(final.Nodes, renderer.OptionsFromConfig(cfg))
				if err != nil {
					t.Fatal(err)
				}
				var rendered struct{ Proxies []map[string]interface{} }
				if err := yaml.Unmarshal(output, &rendered); err != nil || len(rendered.Proxies) != 1 {
					t.Fatalf("invalid rendered proxies: %v", err)
				}
				proxy := rendered.Proxies[0]
				switch obfs {
				case "none", "NoNe":
					if proxy["obfs"] != nil || proxy["obfs-password"] != nil {
						t.Fatalf("disabled obfs returned through YAML passthrough: %#v", proxy)
					}
				case "SALAMANDER":
					if proxy["obfs"] != "salamander" || proxy["obfs-password"] != "secret" {
						t.Fatalf("incorrect normalized obfs: %#v", proxy)
					}
				case "GeCkO":
					if proxy["obfs"] != "gecko" || proxy["obfs-password"] != "secret" {
						t.Fatalf("incorrect normalized obfs: %#v", proxy)
					}
				}
			})
		}
	}
}

func TestHysteria2ObfsValidation(t *testing.T) {
	cases := []struct {
		name         string
		obfs         string
		obfsPassword string
		authPassword string
		want         bool
	}{
		{name: "no obfuscation", authPassword: "auth", want: true},
		{name: "salamander with password", obfs: "salamander", obfsPassword: "obfs-secret", authPassword: "auth", want: true},
		{name: "gecko with password", obfs: "gecko", obfsPassword: "obfs-secret", authPassword: "auth", want: true},
		{name: "salamander missing password", obfs: "salamander", authPassword: "auth"},
		{name: "gecko missing password", obfs: "gecko", authPassword: "auth"},
		{name: "blank obfs password", obfs: "salamander", obfsPassword: "  ", authPassword: "auth"},
		{name: "unsupported obfuscation", obfs: "unknown", obfsPassword: "obfs-secret", authPassword: "auth"},
		{name: "missing authentication", obfs: "salamander", obfsPassword: "obfs-secret"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			node := model.NodeIR{
				Name: "hy2", Type: model.ProtocolHysteria2,
				Server: "hy.example.com", Port: 443,
				Auth: model.Auth{Password: tc.authPassword},
				Raw:  map[string]interface{}{"obfs": tc.obfs, "obfsPassword": tc.obfsPassword},
			}
			if got := isRenderableNode(node); got != tc.want {
				t.Fatalf("isRenderableNode() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildFinalNodesFiltersInvalidHysteria2(t *testing.T) {
	cfg := model.DefaultConfig()
	invalid := model.NormalizeNode(model.NodeIR{
		Name: "invalid-hy2", Type: model.ProtocolHysteria2,
		Server: "hy.example.com", Port: 443,
		Auth: model.Auth{Password: "auth"},
		Raw:  map[string]interface{}{"obfs": "salamander"},
	})
	valid := model.NormalizeNode(model.NodeIR{
		Name: "valid-hy2", Type: model.ProtocolHysteria2,
		Server: "valid.example.com", Port: 443,
		Auth: model.Auth{Password: "auth"},
		Raw:  map[string]interface{}{"obfs": "salamander", "obfsPassword": "secret"},
	})
	finalSet, audit, err := BuildFinalNodes(cfg, model.DefaultNodeState(), []model.NodeIR{invalid, valid})
	if err != nil {
		t.Fatal(err)
	}
	if len(finalSet.Nodes) != 1 || finalSet.Nodes[0].Name != "valid-hy2" {
		t.Fatalf("unexpected final nodes: %#v", finalSet.Nodes)
	}
	if len(audit.ExcludedNodes) != 1 || audit.ExcludedNodes[0].Reason != "invalid_node" {
		t.Fatalf("unexpected audit exclusions: %#v", audit.ExcludedNodes)
	}
}

func TestBuildFinalNodesRevalidatesOverrides(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{
		Name: "changed-hy2", Type: model.ProtocolHysteria2,
		Server: "hy.example.com", Port: 443,
		Auth: model.Auth{Password: "auth"},
	})
	state := model.DefaultNodeState()
	state.NodeOverrides = map[string]model.NodeOverride{
		node.ID: {
			Fields: model.NodeOverrideFields{
				Raw: map[string]interface{}{"obfs": "salamander"},
			},
		},
	}
	finalSet, audit, err := BuildFinalNodes(cfg, state, []model.NodeIR{node})
	if err != nil {
		t.Fatal(err)
	}
	if len(finalSet.Nodes) != 0 {
		t.Fatalf("overridden invalid node leaked: %#v", finalSet.Nodes)
	}
	if len(audit.ExcludedNodes) != 1 || audit.ExcludedNodes[0].Reason != "invalid_node" {
		t.Fatalf("override must be excluded: %#v", audit.ExcludedNodes)
	}
}
