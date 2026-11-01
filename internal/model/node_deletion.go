package model

import (
	"encoding/json"
	"strings"
)

// A connection ID deliberately changes when an endpoint or credential changes.
// Deletion also remembers an unambiguous source/name identity so subscription
// refreshes cannot silently resurrect the same entry with a new connection ID.
// Keep this separate from connection IDs and deduplication: names are not unique
// across subscriptions, or even necessarily within one subscription.
func nodeDeletionSignatures(node NodeIR) []string {
	if strings.TrimSpace(node.Name) == "" {
		return nil
	}
	var signatures []string
	for _, source := range append([]SourceInfo{node.Source}, node.Sources...) {
		if source.Kind != "subscription" || source.ID == "" || source.URLHash == "" {
			continue
		}
		key, _ := json.Marshal([]string{source.ID, source.URLHash, string(node.Type), node.Name})
		signatures = append(signatures, sha256Hex(string(key)))
	}
	return cleanStringSlice(signatures)
}

func uniqueNodeDeletionSignatures(nodes []NodeIR) map[string]string {
	owners := make(map[string]string)
	for _, node := range nodes {
		for _, signature := range nodeDeletionSignatures(node) {
			if owner, exists := owners[signature]; !exists {
				owners[signature] = node.ID
			} else if owner != node.ID {
				owners[signature] = "" // Ambiguous names must never hide another node.
			}
		}
	}
	return owners
}

// RememberDeletedNodeSignatures must receive original, undecorated names, before
// user overrides. Only deleted subscription entries with unique names qualify.
func RememberDeletedNodeSignatures(state *NodeState, nodes []NodeIR) {
	if state.DeletedNodeSignatures == nil {
		state.DeletedNodeSignatures = map[string][]string{}
	}
	deleted := DisabledNodeSet(state.DeletedNodes)
	for signature, id := range uniqueNodeDeletionSignatures(nodes) {
		if _, ok := deleted[id]; ok && id != "" {
			state.DeletedNodeSignatures[id] = mergeUniqueStrings(state.DeletedNodeSignatures[id], []string{signature})
		}
	}
}

// ResolveDeletedNodes maps current connection IDs to their persisted deletion
// records. Callers use this both to hide entries and to restore their old records.
func ResolveDeletedNodes(state NodeState, nodes []NodeIR) map[string][]string {
	deleted := DisabledNodeSet(state.DeletedNodes)
	resolved := make(map[string][]string)
	for _, node := range nodes {
		if _, ok := deleted[node.ID]; ok {
			resolved[node.ID] = []string{node.ID}
		}
	}
	if len(state.DeletedNodeSignatures) == 0 {
		return resolved
	}
	owners := uniqueNodeDeletionSignatures(nodes)
	for id, signatures := range state.DeletedNodeSignatures {
		if _, ok := deleted[id]; !ok {
			continue
		}
		for _, signature := range signatures {
			if currentID := owners[signature]; currentID != "" {
				resolved[currentID] = mergeUniqueStrings(resolved[currentID], []string{id})
			}
		}
	}
	return resolved
}
