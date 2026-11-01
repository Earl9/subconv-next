package model

import "testing"

func TestDeletionIdentityDoesNotHideOtherSourcesOrAmbiguousNames(t *testing.T) {
	base := NodeIR{ID: "original", Name: "same name", Type: ProtocolTrojan, Source: SourceInfo{ID: "source", Kind: "subscription", URLHash: "url-one"}}
	state := DefaultNodeState()
	state.DeletedNodes = []string{base.ID}
	RememberDeletedNodeSignatures(&state, []NodeIR{base})
	rotated := base
	rotated.ID = "rotated"
	if len(ResolveDeletedNodes(state, []NodeIR{rotated})[rotated.ID]) != 1 {
		t.Fatal("endpoint rotation lost deletion")
	}
	for _, change := range []func(*NodeIR){
		func(n *NodeIR) { n.Source.ID = "other-source" },
		func(n *NodeIR) { n.Source.URLHash = "replacement-url" },
		func(n *NodeIR) { n.Source.Kind = "custom" },
		func(n *NodeIR) { n.Type = ProtocolVLESS },
		func(n *NodeIR) { n.Name = "other name" },
	} {
		other := rotated
		change(&other)
		if len(ResolveDeletedNodes(state, []NodeIR{other})) != 0 {
			t.Fatalf("unrelated node was hidden: %#v", other)
		}
	}
	duplicate := rotated
	duplicate.ID = "duplicate"
	if len(ResolveDeletedNodes(state, []NodeIR{rotated, duplicate})) != 0 {
		t.Fatal("ambiguous name hid distinct nodes after refresh")
	}
	ambiguous := DefaultNodeState()
	ambiguous.DeletedNodes = []string{base.ID}
	RememberDeletedNodeSignatures(&ambiguous, []NodeIR{base, duplicate})
	if len(ambiguous.DeletedNodeSignatures) != 0 {
		t.Fatal("ambiguous original names created a broad deletion rule")
	}
	if len(ResolveDeletedNodes(ambiguous, []NodeIR{base, duplicate})) != 1 {
		t.Fatal("exact ID deletion no longer works with duplicate names")
	}
	state.DeletedNodes = nil
	state = NormalizeNodeState(state)
	if len(state.DeletedNodeSignatures) != 0 {
		t.Fatal("restored deletion retained signatures")
	}
}
