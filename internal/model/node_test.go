package model

import "testing"

func TestNormalizeNode(t *testing.T) {
	node := NodeIR{
		Type:   ProtocolVLESS,
		Name:   "  香港 HK\t\n",
		Server: "Example.COM",
		Port:   443,
		Auth: Auth{
			UUID: "uuid-1",
		},
	}

	got := NormalizeNode(node)
	if got.Name != "  香港 HK\t\n" {
		t.Fatalf("Name = %q, want raw name %q", got.Name, "  香港 HK\t\n")
	}
	if got.Server != "example.com" {
		t.Fatalf("Server = %q, want %q", got.Server, "example.com")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "HK" {
		t.Fatalf("Tags = %#v, want %#v", got.Tags, []string{"HK"})
	}
	if got.ID == "" {
		t.Fatalf("ID is empty")
	}
}

func TestDedupeNodes(t *testing.T) {
	nodes := []NodeIR{
		NormalizeNode(NodeIR{
			Type:   ProtocolTrojan,
			Name:   "node-a",
			Server: "example.com",
			Port:   443,
			Auth:   Auth{Password: "secret"},
			Tags:   []string{"US"},
			Source: SourceInfo{Name: "first"},
		}),
		NormalizeNode(NodeIR{
			Type:   ProtocolTrojan,
			Name:   "node-b",
			Server: "EXAMPLE.COM",
			Port:   443,
			Auth:   Auth{Password: "secret"},
			Tags:   []string{"HK"},
			Source: SourceInfo{Name: "second"},
		}),
	}

	got := DedupeNodes(nodes)
	if len(got) != 1 {
		t.Fatalf("len(DedupeNodes) = %d, want 1", len(got))
	}
	if len(got[0].Tags) != 2 {
		t.Fatalf("Tags = %#v, want merged tags", got[0].Tags)
	}
	if len(got[0].Warnings) != 1 || got[0].Warnings[0] != "duplicate skipped from source second" {
		t.Fatalf("Warnings = %#v, want duplicate warning", got[0].Warnings)
	}
}

func TestInferRegionTagsNoFalsePositiveFromNodeSuffix(t *testing.T) {
	got := inferRegionTags("ss-node")
	if len(got) != 0 {
		t.Fatalf("inferRegionTags(ss-node) = %#v, want empty", got)
	}
}

func TestStableNodeIDIgnoresDisplayNameButTracksSecretDigest(t *testing.T) {
	base := NormalizeNode(NodeIR{
		Type:   ProtocolVLESS,
		Name:   "Node A",
		Server: "example.com",
		Port:   443,
		Auth:   Auth{UUID: "uuid-1"},
		Source: SourceInfo{Name: "source-1", Kind: "subscription"},
	})

	renamed := base
	renamed.ID = ""
	renamed.Name = "Node B"
	if StableNodeID(base) != StableNodeID(renamed) {
		t.Fatalf("StableNodeID changed after rename: %q != %q", StableNodeID(base), StableNodeID(renamed))
	}

	changedSecret := base
	changedSecret.ID = ""
	changedSecret.Auth.UUID = "uuid-2"
	if StableNodeID(base) == StableNodeID(changedSecret) {
		t.Fatalf("StableNodeID did not change after secret fingerprint changed")
	}
	if StableNodeID(base) == "uuid-1" || StableNodeID(base) == "uuid-2" {
		t.Fatalf("StableNodeID leaked secret content: %q", StableNodeID(base))
	}
}

func TestNormalizeGroupOptionsDisablesRegionGroupsForV1(t *testing.T) {
	got := NormalizeGroupOptions(GroupOptions{EnableRegionGroups: true})
	if got.EnableRegionGroups {
		t.Fatalf("EnableRegionGroups = true, want false for V1")
	}
	if got.RuleGroupNodeMode != "full" {
		t.Fatalf("RuleGroupNodeMode = %q, want full", got.RuleGroupNodeMode)
	}
	if !got.IncludeRealNodesInRuleGroups {
		t.Fatalf("IncludeRealNodesInRuleGroups = false, want true")
	}
	if !got.SpecialGroupsUseCompact {
		t.Fatalf("SpecialGroupsUseCompact = false, want true")
	}
}

func TestNormalizeGroupOptionsCompactRuleGroups(t *testing.T) {
	got := NormalizeGroupOptions(GroupOptions{RuleGroupNodeMode: "compact"})
	if got.RuleGroupNodeMode != "compact" {
		t.Fatalf("RuleGroupNodeMode = %q, want compact", got.RuleGroupNodeMode)
	}
	if got.IncludeRealNodesInRuleGroups {
		t.Fatalf("IncludeRealNodesInRuleGroups = true, want false")
	}
}
