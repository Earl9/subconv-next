package pipeline

import (
	"fmt"
	"strings"

	"subconv-next/internal/model"
	"subconv-next/internal/nodestate"
	"subconv-next/internal/parser"
)

type NodeValidationWarning struct {
	NodeID  string `json:"node_id"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type StatefulCollectResult struct {
	Nodes              []model.NodeIR
	Warnings           []string
	Errors             []parser.ParseError
	ValidationWarnings []NodeValidationWarning
	SubscriptionMeta   map[string]model.SubscriptionMeta
}

func LoadNodeState(cfg model.Config) (model.NodeState, error) {
	return nodestate.Load(cfg.Service.StatePath)
}

func SaveNodeState(cfg model.Config, state model.NodeState) error {
	return nodestate.Save(cfg.Service.StatePath, state)
}

func CollectNodesWithState(cfg model.Config, state model.NodeState, applyFilters bool, excludeDisabled bool) StatefulCollectResult {
	collected := collectNodes(cfg, applyFilters)
	state = model.NormalizeNodeState(state)

	nodes := model.CloneNodes(collected.Nodes)
	if len(state.CustomNodes) > 0 {
		nodes = append(nodes, model.CloneNodes(state.CustomNodes)...)
		nodes = model.NormalizeNodesWithScope(nodes, cfg.Render.DedupeScope)
	}

	nodes = applyNodeOverrides(nodes, state.NodeOverrides)
	nodes, validationWarnings := validateNodes(nodes)
	if excludeDisabled {
		nodes = filterDisabledNodes(nodes, state.DisabledNodes)
	}

	return StatefulCollectResult{
		Nodes:              nodes,
		Warnings:           collected.Warnings,
		Errors:             collected.Errors,
		ValidationWarnings: validationWarnings,
		SubscriptionMeta:   cloneSubscriptionMetaMap(collected.SubscriptionMeta),
	}
}

func applyNodeOverrides(nodes []model.NodeIR, overrides map[string]model.NodeOverride) []model.NodeIR {
	if len(nodes) == 0 || len(overrides) == 0 {
		return nodes
	}

	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		override, ok := overrides[node.ID]
		if !ok {
			out = append(out, node)
			continue
		}

		originalID := node.ID
		if strings.TrimSpace(override.Name) != "" {
			node.Name = override.Name
			if node.Raw == nil {
				node.Raw = make(map[string]interface{})
			}
			node.Raw["_overrideName"] = true
		}
		if len(override.Tags) > 0 {
			node.Tags = append([]string(nil), override.Tags...)
		}
		if strings.TrimSpace(override.Region) != "" {
			node.Tags = model.ReplaceRegionTag(node.Tags, override.Region)
		}
		if strings.TrimSpace(override.Fields.Server) != "" {
			node.Server = override.Fields.Server
		}
		if override.Fields.Port > 0 {
			node.Port = override.Fields.Port
		}
		if override.Fields.UDP != nil {
			v := *override.Fields.UDP
			node.UDP = &v
		}
		if override.Fields.TLS != nil {
			tlsCopy := *override.Fields.TLS
			node.TLS = tlsCopy
		}
		if override.Fields.Auth != nil {
			authCopy := *override.Fields.Auth
			node.Auth = authCopy
		}
		if override.Fields.Transport != nil {
			transportCopy := *override.Fields.Transport
			node.Transport = transportCopy
		}
		if override.Fields.WireGuard != nil {
			wgCopy := *override.Fields.WireGuard
			node.WireGuard = &wgCopy
		}
		if override.Fields.Raw != nil {
			rawCopy := make(map[string]interface{}, len(override.Fields.Raw))
			for key, value := range override.Fields.Raw {
				rawCopy[key] = value
			}
			node.Raw = rawCopy
		}

		node = model.NormalizeNode(node)
		node.ID = originalID
		out = append(out, node)
	}
	return out
}

func filterDisabledNodes(nodes []model.NodeIR, disabledIDs []string) []model.NodeIR {
	if len(nodes) == 0 || len(disabledIDs) == 0 {
		return nodes
	}
	disabled := model.DisabledNodeSet(disabledIDs)
	out := make([]model.NodeIR, 0, len(nodes))
	for _, node := range nodes {
		if _, ok := disabled[node.ID]; ok {
			continue
		}
		out = append(out, node)
	}
	return out
}

func validateNodes(nodes []model.NodeIR) ([]model.NodeIR, []NodeValidationWarning) {
	if len(nodes) == 0 {
		return nodes, nil
	}

	out := make([]model.NodeIR, 0, len(nodes))
	var warnings []NodeValidationWarning
	for _, node := range nodes {
		nodeWarnings := append([]string(nil), node.Warnings...)

		if node.Type == model.ProtocolVLESS && strings.EqualFold(strings.TrimSpace(node.Transport.Network), "xhttp") {
			if strings.TrimSpace(node.Transport.Path) == "" {
				message := "vless network=xhttp 缺少 xhttp-opts.path"
				nodeWarnings = append(nodeWarnings, message)
				warnings = append(warnings, NodeValidationWarning{NodeID: node.ID, Level: "warning", Message: message})
			}
		}
		if node.Type == model.ProtocolAnyTLS && strings.TrimSpace(node.TLS.ClientFingerprint) == "" {
			message := "anytls 缺少 client-fingerprint"
			nodeWarnings = append(nodeWarnings, message)
			warnings = append(warnings, NodeValidationWarning{NodeID: node.ID, Level: "warning", Message: message})
		}
		if strings.TrimSpace(node.Server) == "" || node.Port <= 0 {
			message := fmt.Sprintf("%s 缺少有效 server/port", node.Type)
			nodeWarnings = append(nodeWarnings, message)
			warnings = append(warnings, NodeValidationWarning{NodeID: node.ID, Level: "warning", Message: message})
		}

		node.Warnings = uniqueWarnings(nodeWarnings)
		out = append(out, node)
	}
	return out, warnings
}

func uniqueWarnings(values []string) []string {
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
