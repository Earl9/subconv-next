package api

import (
	"strings"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

func pruneDeletedNodeStateForConfigChange(oldCfg, newCfg model.Config, state model.NodeState) (model.NodeState, bool, int) {
	state = model.NormalizeNodeState(state)
	removedSources := removedSubscriptionSourceIDs(oldCfg, newCfg)
	if len(removedSources) == 0 || len(state.DeletedNodes) == 0 {
		return pruneRemovedSubscriptionMeta(state, activeSubscriptionSourceIDs(newCfg), removedSources)
	}

	changed := rememberDeletedNodeSourcesFromAudit(&state)
	pruneUnknownMissing := false
	if hasDeletedNodeWithoutSources(state) {
		oldCollected := pipeline.CollectNodesWithState(oldCfg, state, true, false)
		if rememberDeletedNodeSources(&state, oldCollected.Nodes) {
			changed = true
		}
		pruneUnknownMissing = len(oldCollected.Errors) == 0
	}
	state, pruned := pruneDeletedNodeState(
		state,
		activeSubscriptionSourceIDs(newCfg),
		nil,
		pruneUnknownMissing,
	)
	if pruned > 0 {
		changed = true
	}
	state, metaChanged, _ := pruneRemovedSubscriptionMeta(state, activeSubscriptionSourceIDs(newCfg), removedSources)
	return state, changed || metaChanged, pruned
}

func hasDeletedNodeWithoutSources(state model.NodeState) bool {
	for _, nodeID := range state.DeletedNodes {
		if len(state.DeletedNodeSources[nodeID]) == 0 {
			return true
		}
	}
	return false
}

func pruneDeletedNodeStateForCurrentConfig(cfg model.Config, state model.NodeState, currentNodes []model.NodeIR, pruneUnknownMissing bool) (model.NodeState, bool, int) {
	state = model.NormalizeNodeState(state)
	changed := rememberDeletedNodeSources(&state, currentNodes)
	if rememberDeletedNodeSourcesFromAudit(&state) {
		changed = true
	}
	state, pruned := pruneDeletedNodeState(
		state,
		activeSubscriptionSourceIDs(cfg),
		nodeIDSet(currentNodes),
		pruneUnknownMissing,
	)
	return state, changed || pruned > 0, pruned
}

func pruneDeletedNodeStateForActiveConfig(cfg model.Config, state model.NodeState) model.NodeState {
	state = model.NormalizeNodeState(state)
	_ = rememberDeletedNodeSourcesFromAudit(&state)
	state, _ = pruneDeletedNodeState(state, activeSubscriptionSourceIDs(cfg), nil, false)
	return state
}

func pruneRestoredDraftDeletedNodeState(cfg model.Config, state model.NodeState) model.NodeState {
	state = pruneDeletedNodeStateForActiveConfig(cfg, state)
	if !hasDeletedNodeWithoutSources(state) {
		return state
	}
	collected := pipeline.CollectNodesWithState(cfg, state, true, false)
	state, _, _ = pruneDeletedNodeStateForCurrentConfig(
		cfg,
		state,
		collected.Nodes,
		len(collected.Errors) == 0,
	)
	return state
}

func rememberDeletedNodeSources(state *model.NodeState, nodes []model.NodeIR) bool {
	if state == nil || len(state.DeletedNodes) == 0 || len(nodes) == 0 {
		return false
	}
	deleted := model.DisabledNodeSet(state.DeletedNodes)
	if state.DeletedNodeSources == nil {
		state.DeletedNodeSources = map[string][]string{}
	}
	changed := false
	for _, node := range nodes {
		if _, ok := deleted[node.ID]; !ok {
			continue
		}
		sources := subscriptionSourceIDs(node)
		if len(sources) == 0 {
			continue
		}
		merged := uniqueStrings(append(append([]string(nil), state.DeletedNodeSources[node.ID]...), sources...))
		if !equalStrings(state.DeletedNodeSources[node.ID], merged) {
			state.DeletedNodeSources[node.ID] = merged
			changed = true
		}
	}
	return changed
}

func rememberDeletedNodeSourcesFromAudit(state *model.NodeState) bool {
	if state == nil || len(state.DeletedNodes) == 0 {
		return false
	}
	deleted := model.DisabledNodeSet(state.DeletedNodes)
	if state.DeletedNodeSources == nil {
		state.DeletedNodeSources = map[string][]string{}
	}
	changed := false
	for _, excluded := range state.LastAudit.ExcludedNodes {
		if !strings.EqualFold(strings.TrimSpace(excluded.Reason), "deleted_node") {
			continue
		}
		id := strings.TrimSpace(excluded.ID)
		if _, ok := deleted[id]; !ok {
			continue
		}
		sourceID := strings.TrimSpace(excluded.Source.ID)
		if sourceID == "" || !strings.EqualFold(strings.TrimSpace(excluded.Source.Kind), "subscription") {
			continue
		}
		merged := uniqueStrings(append(append([]string(nil), state.DeletedNodeSources[id]...), sourceID))
		if !equalStrings(state.DeletedNodeSources[id], merged) {
			state.DeletedNodeSources[id] = merged
			changed = true
		}
	}
	return changed
}

func pruneDeletedNodeState(state model.NodeState, activeSources, currentNodeIDs map[string]struct{}, pruneUnknownMissing bool) (model.NodeState, int) {
	state = model.NormalizeNodeState(state)
	kept := make([]string, 0, len(state.DeletedNodes))
	prunedIDs := make(map[string]struct{})
	for _, id := range state.DeletedNodes {
		if _, current := currentNodeIDs[id]; current {
			kept = append(kept, id)
			continue
		}
		sources := state.DeletedNodeSources[id]
		if len(sources) == 0 && !pruneUnknownMissing {
			kept = append(kept, id)
			continue
		}
		if len(sources) > 0 && intersectsStringSet(sources, activeSources) {
			kept = append(kept, id)
			continue
		}
		prunedIDs[id] = struct{}{}
		delete(state.DeletedNodeSources, id)
	}
	if len(prunedIDs) == 0 {
		return state, 0
	}
	state.DeletedNodes = kept
	filteredAudit := state.LastAudit.ExcludedNodes[:0]
	for _, excluded := range state.LastAudit.ExcludedNodes {
		if _, pruned := prunedIDs[excluded.ID]; pruned {
			continue
		}
		filteredAudit = append(filteredAudit, excluded)
	}
	state.LastAudit.ExcludedNodes = filteredAudit
	return model.NormalizeNodeState(state), len(prunedIDs)
}

func pruneRemovedSubscriptionMeta(state model.NodeState, activeSources, removedSources map[string]struct{}) (model.NodeState, bool, int) {
	if len(removedSources) == 0 || len(state.SubscriptionMeta) == 0 {
		return state, false, 0
	}
	count := 0
	for sourceID := range state.SubscriptionMeta {
		if _, active := activeSources[sourceID]; active {
			continue
		}
		if _, removed := removedSources[sourceID]; !removed {
			continue
		}
		delete(state.SubscriptionMeta, sourceID)
		count++
	}
	return state, count > 0, count
}

func activeSubscriptionSourceIDs(cfg model.Config) map[string]struct{} {
	active := make(map[string]struct{}, len(cfg.Subscriptions))
	for _, subscription := range cfg.Subscriptions {
		if id := strings.TrimSpace(subscription.ID); id != "" {
			active[id] = struct{}{}
		}
	}
	return active
}

func removedSubscriptionSourceIDs(oldCfg, newCfg model.Config) map[string]struct{} {
	active := activeSubscriptionSourceIDs(newCfg)
	removed := map[string]struct{}{}
	for sourceID := range activeSubscriptionSourceIDs(oldCfg) {
		if _, stillActive := active[sourceID]; !stillActive {
			removed[sourceID] = struct{}{}
		}
	}
	return removed
}

func subscriptionSourceIDs(node model.NodeIR) []string {
	var sources []string
	appendSource := func(source model.SourceInfo) {
		if !strings.EqualFold(strings.TrimSpace(source.Kind), "subscription") {
			return
		}
		if id := strings.TrimSpace(source.ID); id != "" {
			sources = append(sources, id)
		}
	}
	appendSource(node.Source)
	for _, source := range node.Sources {
		appendSource(source)
	}
	return uniqueStrings(sources)
}

func nodeIDSet(nodes []model.NodeIR) map[string]struct{} {
	set := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if id := strings.TrimSpace(node.ID); id != "" {
			set[id] = struct{}{}
		}
	}
	return set
}

func intersectsStringSet(values []string, set map[string]struct{}) bool {
	for _, value := range values {
		if _, ok := set[strings.TrimSpace(value)]; ok {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
