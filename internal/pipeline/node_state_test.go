package pipeline

import (
	"bytes"
	"path/filepath"
	"testing"

	"subconv-next/internal/model"
)

func TestApplyNodeOverrides(t *testing.T) {
	node := model.NormalizeNode(model.NodeIR{
		Name:   "old-name",
		Type:   model.ProtocolAnyTLS,
		Server: "example.com",
		Port:   443,
		Auth:   model.Auth{Password: "old-pass"},
		TLS: model.TLSOptions{
			Enabled:           true,
			SNI:               "old.example.com",
			ClientFingerprint: "chrome",
		},
		UDP: model.Bool(true),
	})

	overridden := applyNodeOverrides([]model.NodeIR{node}, map[string]model.NodeOverride{
		node.ID: {
			Name:   "new-name",
			Region: "JP",
			Tags:   []string{"jp"},
			Fields: model.NodeOverrideFields{
				Server: "override.example.com",
				Port:   8443,
				UDP:    model.Bool(false),
				TLS: &model.TLSOptions{
					Enabled:           true,
					SNI:               "override.example.com",
					ClientFingerprint: "firefox",
				},
				Auth: &model.Auth{Password: "new-pass"},
				Raw:  map[string]interface{}{"idleSessionTimeout": "30s"},
			},
		},
	})

	got := overridden[0]
	if got.Name != "new-name" || got.Server != "override.example.com" || got.Port != 8443 {
		t.Fatalf("override result = %#v, want name/server/port replaced", got)
	}
	if got.UDP == nil || *got.UDP {
		t.Fatalf("override UDP = %#v, want false", got.UDP)
	}
	if got.TLS.SNI != "override.example.com" || got.TLS.ClientFingerprint != "firefox" {
		t.Fatalf("override TLS = %#v, want replacement", got.TLS)
	}
	if got.Auth.Password != "new-pass" {
		t.Fatalf("override Auth.Password = %q, want %q", got.Auth.Password, "new-pass")
	}
	if got.ID != node.ID {
		t.Fatalf("override ID = %q, want original %q", got.ID, node.ID)
	}
}

func TestFilterDisabledNodes(t *testing.T) {
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{Name: "a", Type: model.ProtocolSS, Server: "a.example.com", Port: 1}),
		model.NormalizeNode(model.NodeIR{Name: "b", Type: model.ProtocolSS, Server: "b.example.com", Port: 1}),
	}

	got := filterDisabledNodes(nodes, []string{nodes[1].ID})
	if len(got) != 1 || got[0].ID != nodes[0].ID {
		t.Fatalf("filterDisabledNodes() = %#v, want only first node", got)
	}
}

func TestCollectNodesWithStateIncludesCustomNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}

	custom := model.NormalizeNode(model.NodeIR{
		Name:   "custom-vless",
		Type:   model.ProtocolVLESS,
		Server: "custom.example.com",
		Port:   443,
		Auth:   model.Auth{UUID: "uuid-1"},
		Source: model.SourceInfo{ID: "custom", Name: "manual", Kind: "custom"},
	})
	custom.ID = "custom-" + custom.ID

	result := CollectNodesWithState(cfg, model.NodeState{
		NodeOverrides: map[string]model.NodeOverride{},
		DisabledNodes: []string{},
		CustomNodes:   []model.NodeIR{custom},
	}, true, false)

	if len(result.Nodes) != 2 {
		t.Fatalf("len(result.Nodes) = %d, want 2", len(result.Nodes))
	}
	foundCustom := false
	for _, node := range result.Nodes {
		if node.ID == custom.ID {
			foundCustom = true
			break
		}
	}
	if !foundCustom {
		t.Fatalf("custom node %#v not found in %#v", custom, result.Nodes)
	}
}

func TestRenderConfigUsesOverrideValues(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}

	base := collectNodes(cfg, true)
	nodeID := base.Nodes[0].ID
	err := SaveNodeState(cfg, model.NodeState{
		NodeOverrides: map[string]model.NodeOverride{
			nodeID: {
				Name: "renamed-ss",
				Fields: model.NodeOverrideFields{
					Server: "override.example.com",
					Port:   9443,
					Auth:   &model.Auth{Password: "pass"},
					Raw:    map[string]interface{}{"method": "aes-256-gcm"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	result, err := RenderConfig(cfg)
	if err != nil {
		t.Fatalf("RenderConfig() error = %v", err)
	}
	if !bytes.Contains(result.YAML, []byte(`renamed-ss`)) {
		t.Fatalf("rendered YAML = %q, want renamed node", string(result.YAML))
	}
	if !bytes.Contains(result.YAML, []byte("server: override.example.com")) {
		t.Fatalf("rendered YAML = %q, want overridden server", string(result.YAML))
	}
}

func TestRenderConfigSkipsDisabledAndIncludesCustomNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.StatePath = filepath.Join(t.TempDir(), "state.json")
	cfg.Service.OutputPath = filepath.Join(t.TempDir(), "mihomo.yaml")
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
		},
	}

	base := collectNodes(cfg, true)
	nodeID := base.Nodes[0].ID
	custom := model.NormalizeNode(model.NodeIR{
		Name:   "custom-vless",
		Type:   model.ProtocolVLESS,
		Server: "custom.example.com",
		Port:   443,
		Auth:   model.Auth{UUID: "uuid-1"},
		TLS:    model.TLSOptions{Enabled: true, SNI: "custom.example.com"},
		Source: model.SourceInfo{Name: "manual", Kind: "custom"},
	})
	custom.ID = "custom-" + custom.ID

	err := SaveNodeState(cfg, model.NodeState{
		DisabledNodes: []string{nodeID},
		CustomNodes:   []model.NodeIR{custom},
	})
	if err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	result, err := RenderConfig(cfg)
	if err != nil {
		t.Fatalf("RenderConfig() error = %v", err)
	}
	if bytes.Contains(result.YAML, []byte(`ss-node`)) {
		t.Fatalf("rendered YAML = %q, disabled node should be excluded", string(result.YAML))
	}
	if !bytes.Contains(result.YAML, []byte(`custom-vless`)) {
		t.Fatalf("rendered YAML = %q, custom node should be included", string(result.YAML))
	}
}

func TestApplyRenderPreferencesFiltersInvalidSSNode(t *testing.T) {
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{
			Name:   "invalid-ss",
			Type:   model.ProtocolSS,
			Server: "example.com",
			Port:   443,
			Auth:   model.Auth{Password: "secret"},
		}),
	}

	got := applyRenderPreferences(nodes, model.RenderConfig{
		FilterIllegal: true,
		UDP:           true,
	})
	if len(got) != 0 {
		t.Fatalf("applyRenderPreferences() = %#v, want invalid ss node filtered", got)
	}
}
