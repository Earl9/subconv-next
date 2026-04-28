package pipeline

import (
	"testing"

	"subconv-next/internal/model"
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

func TestApplySourcePrefixNaming(t *testing.T) {
	node := model.NormalizeNode(model.NodeIR{
		Name:   "JP Tokyo",
		Type:   model.ProtocolAnyTLS,
		Server: "example.com",
		Port:   443,
		Source: model.SourceInfo{Name: "主力机场"},
	})
	got := ApplySourcePrefix(node, model.RenderConfig{
		SourcePrefix:       true,
		SourcePrefixFormat: "[{source}] {name}",
	})
	if got.Name != "[主力机场] JP Tokyo" {
		t.Fatalf("ApplySourcePrefix() = %q, want %q", got.Name, "[主力机场] JP Tokyo")
	}
}

func TestApplySourcePrefixNoDuplicate(t *testing.T) {
	node := model.NormalizeNode(model.NodeIR{
		Name:   "[主力机场] JP Tokyo",
		Type:   model.ProtocolAnyTLS,
		Server: "example.com",
		Port:   443,
		Source: model.SourceInfo{Name: "主力机场"},
	})
	got := ApplySourcePrefix(node, model.RenderConfig{
		SourcePrefix:       true,
		SourcePrefixFormat: "[{source}] {name}",
	})
	if got.Name != "[主力机场] JP Tokyo" {
		t.Fatalf("ApplySourcePrefix() duplicated prefix: %q", got.Name)
	}
}

func TestApplySourcePrefixDisabled(t *testing.T) {
	node := model.NormalizeNode(model.NodeIR{
		Name:   "JP Tokyo",
		Type:   model.ProtocolAnyTLS,
		Server: "example.com",
		Port:   443,
		Source: model.SourceInfo{Name: "主力机场"},
	})
	got := ApplySourcePrefix(node, model.RenderConfig{
		SourcePrefix:       false,
		SourcePrefixFormat: "[{source}] {name}",
	})
	if got.Name != "JP Tokyo" {
		t.Fatalf("ApplySourcePrefix() changed name with source_prefix=false: %q", got.Name)
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
