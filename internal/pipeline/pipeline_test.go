package pipeline

import (
	"testing"

	"subconv-next/internal/model"
)

func TestFilterSubscriptionNodes(t *testing.T) {
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{
			Name:   "[anytls]JP Tokyo",
			Type:   model.ProtocolAnyTLS,
			Server: "jp.example.com",
			Port:   443,
			Tags:   []string{"日本"},
		}),
		model.NormalizeNode(model.NodeIR{
			Name:   "[anytls]US New York",
			Type:   model.ProtocolAnyTLS,
			Server: "us.example.com",
			Port:   443,
			Tags:   []string{"美国"},
		}),
		model.NormalizeNode(model.NodeIR{
			Name:   "[vless]JP Test",
			Type:   model.ProtocolVLESS,
			Server: "test.example.com",
			Port:   8443,
			Tags:   []string{"日本"},
		}),
	}

	filtered, dropped := filterSubscriptionNodes(nodes, model.SubscriptionConfig{
		IncludeKeywords: []string{"jp", "日本"},
		ExcludeKeywords: []string{"test"},
	})

	if dropped != 2 {
		t.Fatalf("dropped = %d, want 2", dropped)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].Name != "[anytls]JP Tokyo" {
		t.Fatalf("filtered[0].Name = %q, want %q", filtered[0].Name, "[anytls]JP Tokyo")
	}
}

func TestFilterSubscriptionNodesManualExcludedIDs(t *testing.T) {
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{
			Name:   "[anytls]JP Tokyo",
			Type:   model.ProtocolAnyTLS,
			Server: "jp.example.com",
			Port:   443,
		}),
		model.NormalizeNode(model.NodeIR{
			Name:   "[anytls]US New York",
			Type:   model.ProtocolAnyTLS,
			Server: "us.example.com",
			Port:   443,
		}),
	}

	filtered, dropped := filterSubscriptionNodes(nodes, model.SubscriptionConfig{
		ExcludedNodeIDs: []string{nodes[1].ID},
	})

	if dropped != 1 {
		t.Fatalf("dropped = %d, want 1", dropped)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].ID != nodes[0].ID {
		t.Fatalf("filtered[0].ID = %q, want %q", filtered[0].ID, nodes[0].ID)
	}
}

func TestCollectNodesSkipsInlineWhenManualNodesDisabled(t *testing.T) {
	cfg := model.DefaultConfig()
	disabled := false
	cfg.ManualNodesEnabled = &disabled
	cfg.Inline = []model.InlineConfig{
		{
			Name:    "manual",
			Enabled: true,
			Content: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#manual",
		},
	}

	result := CollectNodes(cfg)
	if len(result.Nodes) != 0 {
		t.Fatalf("CollectNodes() nodes = %d, want 0 when manual nodes disabled", len(result.Nodes))
	}
}

func TestNormalizeKeywords(t *testing.T) {
	got := normalizeKeywords([]string{" JP, 香港 ", "日本\nJP", "", "香港"})
	want := []string{"jp", "香港", "日本"}

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q; got=%v", i, got[i], want[i], got)
		}
	}
}
