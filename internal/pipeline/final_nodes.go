package pipeline

import (
	"sort"
	"strings"

	"subconv-next/internal/model"
)

type FinalNodeSet struct {
	Nodes []model.NodeIR
}

func BuildFinalNodes(cfg model.Config, state model.NodeState, rawNodes []model.NodeIR) (FinalNodeSet, model.AuditReport, error) {
	state = model.NormalizeNodeState(state)

	candidates := append([]model.NodeIR{}, model.CloneNodes(rawNodes)...)
	if len(state.CustomNodes) > 0 {
		candidates = append(candidates, model.CloneNodes(state.CustomNodes)...)
	}
	candidates = model.NormalizeNodesNoDedupe(candidates)

	audit := model.AuditReport{
		RawCount:      len(candidates),
		ExcludedNodes: []model.ExcludedNode{},
		Warnings:      []model.AuditWarning{},
	}

	includeMatcher := newNameMatcher(cfg.Render.IncludeKeywords)
	excludeMatcher := newNameMatcher(cfg.Render.ExcludeKeywords)

	filtered := make([]model.NodeIR, 0, len(candidates))
	for _, node := range candidates {
		node = markInfoNode(node)
		if isInfoNode(node) && !(cfg.Render.ShowInfoNodes && cfg.Render.IncludeInfoNode) {
			auditExclude(&audit, node, "info_node")
			continue
		}
		if includeMatcher != nil && !includeMatcher(node.Name) {
			auditExclude(&audit, node, "include_keyword_not_matched")
			continue
		}
		if excludeMatcher != nil && excludeMatcher(node.Name) {
			auditExclude(&audit, node, "exclude_keyword_matched")
			continue
		}
		if cfg.Render.FilterIllegal && !isRenderableNode(node) {
			auditExclude(&audit, node, "invalid_node")
			continue
		}
		filtered = append(filtered, node)
	}

	deduped := dedupeFinalNodes(filtered, cfg.Render.DedupeScope, &audit)
	for index := range deduped {
		deduped[index].ID = model.StableNodeIDWithScope(deduped[index], cfg.Render.DedupeScope, index)
		if strings.EqualFold(strings.TrimSpace(deduped[index].Source.Kind), "custom") && !strings.HasPrefix(deduped[index].ID, "custom-") {
			deduped[index].ID = "custom-" + deduped[index].ID
		}
	}

	deduped = applyNodeOverrides(deduped, state.NodeOverrides)

	disabledSet := model.DisabledNodeSet(state.DisabledNodes)
	deletedSet := model.DisabledNodeSet(state.DeletedNodes)

	finalNodes := make([]model.NodeIR, 0, len(deduped))
	for _, node := range deduped {
		if _, ok := disabledSet[node.ID]; ok {
			auditExclude(&audit, node, "disabled_node")
			continue
		}
		if _, ok := deletedSet[node.ID]; ok {
			auditExclude(&audit, node, "deleted_node")
			continue
		}
		finalNodes = append(finalNodes, node)
	}

	finalNodes = applyNodeDecorations(finalNodes, cfg.Render)
	finalNodes = ensureUniqueFinalNodeNames(finalNodes, cfg.Render)

	audit.FinalCount = len(finalNodes)
	audit.ExcludedCount = len(audit.ExcludedNodes)

	return FinalNodeSet{Nodes: finalNodes}, audit, nil
}

func markInfoNode(node model.NodeIR) model.NodeIR {
	if !looksLikeInfoNode(node.Name) {
		return node
	}
	if node.Raw == nil {
		node.Raw = map[string]interface{}{}
	}
	node.Raw["_infoNode"] = true
	return node
}

func isInfoNode(node model.NodeIR) bool {
	return rawBool(node.Raw, "_infoNode") || looksLikeInfoNode(node.Name)
}

func auditExclude(audit *model.AuditReport, node model.NodeIR, reason string) {
	audit.ExcludedNodes = append(audit.ExcludedNodes, model.ExcludedNode{
		ID:     strings.TrimSpace(node.ID),
		Name:   strings.TrimSpace(node.Name),
		Source: node.Source,
		Reason: reason,
	})
}

func dedupeFinalNodes(nodes []model.NodeIR, scope string, audit *model.AuditReport) []model.NodeIR {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "none" {
		return append([]model.NodeIR(nil), nodes...)
	}

	seen := make(map[string]int, len(nodes))
	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		key := model.StableNodeIDWithScope(node, scope, 0)
		if idx, ok := seen[key]; ok {
			existing := out[idx]
			existing.Tags = mergeUniqueStrings(existing.Tags, node.Tags)
			existing.Warnings = mergeUniqueStrings(existing.Warnings, append([]string{}, node.Warnings...))
			existing.Sources = mergeNodeSources(existing, node)
			out[idx] = existing
			auditExclude(audit, node, "duplicate")
			continue
		}
		seen[key] = len(out)
		out = append(out, node)
	}
	return out
}

func mergeNodeSources(existing, duplicate model.NodeIR) []model.SourceInfo {
	combined := append([]model.SourceInfo{}, model.MergeSourcesForView(existing)...)
	combined = append(combined, model.MergeSourcesForView(duplicate)...)
	seen := map[string]struct{}{}
	out := make([]model.SourceInfo, 0, len(combined))
	for _, source := range combined {
		key := strings.Join([]string{source.ID, source.Name, source.Emoji, source.Kind, source.URLHash}, "|")
		if key == "||||" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, source)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func applyNodeDecorations(nodes []model.NodeIR, render model.RenderConfig) []model.NodeIR {
	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		if render.SkipTLSVerify {
			node.TLS.Insecure = true
		}
		if render.UDP {
			node.UDP = model.Bool(true)
		}
		node.Name = model.BuildYamlNodeName(node.Name, node.Source, model.EffectiveNameOptions(render))
		out = append(out, node)
	}
	if render.SortNodes {
		sort.SliceStable(out, func(i, j int) bool {
			return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
		})
	}
	return out
}

func ensureUniqueFinalNodeNames(nodes []model.NodeIR, render model.RenderConfig) []model.NodeIR {
	return model.EnsureUniqueProxyNames(nodes, model.EffectiveNameOptions(render).DedupeSuffixStyle)
}

func mergeUniqueStrings(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, value := range append(append([]string{}, a...), b...) {
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
	sort.Strings(out)
	return out
}
