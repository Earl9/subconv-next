package pipeline

import (
	"testing"

	"subconv-next/internal/model"
)

func TestCustomNodeIdentitySurvivesDedupeScope(t *testing.T) {
	for _, scope := range []string{"global", "per_source", "none"} {
		t.Run(scope, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Render.DedupeScope = scope
			state := model.NormalizeNodeState(model.NodeState{CustomNodes: []model.NodeIR{{
				Name: "manual", Type: model.ProtocolTrojan, Server: "manual.example", Port: 443,
				Auth: model.Auth{Password: "test"}, Source: model.SourceInfo{Kind: "custom"},
			}}})
			id := state.CustomNodes[0].ID
			state.NodeOverrides[id] = model.NodeOverride{Name: "renamed"}
			collected := CollectNodesWithState(cfg, state, true, false)
			if len(collected.Nodes) != 1 || collected.Nodes[0].ID != id || collected.Nodes[0].Name != "renamed" {
				t.Fatalf("editor lost stored custom identity or override: %#v", collected.Nodes)
			}
			state.DisabledNodes = []string{id}
			final, _, err := BuildFinalNodes(cfg, state, nil)
			if err != nil || len(final.Nodes) != 0 {
				t.Fatalf("disabled custom node rendered: %#v, err=%v", final.Nodes, err)
			}
		})
	}
}

func TestNoDedupeDeletionSurvivesInfoNodeFiltering(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.DedupeScope = "none"
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://test@info.example:443#剩余流量\ntrojan://test@delete.example:443#delete-me\ntrojan://test@keep.example:443#keep"}}
	collected := CollectNodesWithState(cfg, model.DefaultNodeState(), true, false)
	if len(collected.Nodes) != 3 {
		t.Fatalf("collected %d nodes, want 3", len(collected.Nodes))
	}
	state := model.DefaultNodeState()
	state.DeletedNodes = []string{collected.Nodes[1].ID}
	final, _, err := BuildFinalNodes(cfg, state, collected.Nodes)
	if err != nil || len(final.Nodes) != 1 || final.Nodes[0].Server != "keep.example" {
		t.Fatalf("deleted node reappeared after filtering: %#v, err=%v", final.Nodes, err)
	}
}

func TestCustomDuplicateRemainsIndependentlyEditable(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.SourcePrefix = false
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://test@shared.example:443#inline-copy"}}
	raw := CollectNodes(cfg).Nodes
	custom := model.CloneNodes(raw)[0]
	custom.ID = "custom-" + custom.ID
	custom.Name = "custom-copy"
	custom.Source = model.SourceInfo{Kind: "custom"}
	custom.Sources = nil
	state := model.NormalizeNodeState(model.NodeState{CustomNodes: []model.NodeIR{custom}})
	collected := CollectNodesWithState(cfg, state, true, false)
	if len(collected.Nodes) != 2 || collected.Nodes[0].ID == collected.Nodes[1].ID {
		t.Fatalf("custom copy is not independently editable: %#v", collected.Nodes)
	}
	state.DeletedNodes = []string{raw[0].ID}
	state.NodeOverrides[custom.ID] = model.NodeOverride{Name: "kept-custom"}
	final, _, err := BuildFinalNodes(cfg, state, raw)
	if err != nil || len(final.Nodes) != 1 || final.Nodes[0].Name != "kept-custom" || final.Nodes[0].ID != custom.ID {
		t.Fatalf("deleting inline copy suppressed custom node: %#v, err=%v", final.Nodes, err)
	}
}
