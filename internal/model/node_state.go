package model

import (
	"encoding/json"
	"sort"
	"strings"
)

type NodeState struct {
	NodeOverrides map[string]NodeOverride `json:"node_overrides,omitempty"`
	DisabledNodes []string                `json:"disabled_nodes,omitempty"`
	CustomNodes   []NodeIR                `json:"custom_nodes,omitempty"`
}

type NodeOverride struct {
	Enabled   bool               `json:"enabled"`
	Name      string             `json:"name,omitempty"`
	Region    string             `json:"region,omitempty"`
	Tags      []string           `json:"tags,omitempty"`
	Fields    NodeOverrideFields `json:"fields,omitempty"`
	UpdatedAt string             `json:"updated_at,omitempty"`
}

type NodeOverrideFields struct {
	Server    string                 `json:"server,omitempty"`
	Port      int                    `json:"port,omitempty"`
	UDP       *bool                  `json:"udp,omitempty"`
	TLS       *TLSOptions            `json:"tls,omitempty"`
	Auth      *Auth                  `json:"auth,omitempty"`
	Transport *TransportOptions      `json:"transport,omitempty"`
	WireGuard *WireGuardOptions      `json:"wireguard,omitempty"`
	Raw       map[string]interface{} `json:"raw,omitempty"`
}

func DefaultNodeState() NodeState {
	return NodeState{
		NodeOverrides: map[string]NodeOverride{},
		DisabledNodes: []string{},
		CustomNodes:   []NodeIR{},
	}
}

func NormalizeNodeState(state NodeState) NodeState {
	if state.NodeOverrides == nil {
		state.NodeOverrides = map[string]NodeOverride{}
	}
	if state.DisabledNodes == nil {
		state.DisabledNodes = []string{}
	}
	if state.CustomNodes == nil {
		state.CustomNodes = []NodeIR{}
	}

	state.DisabledNodes = cleanStringSlice(state.DisabledNodes)

	for id, override := range state.NodeOverrides {
		cleanID := sanitizeText(id)
		override.Name = sanitizeText(override.Name)
		override.Region = strings.ToUpper(sanitizeText(override.Region))
		override.Tags = cleanStringSlice(override.Tags)
		if override.Fields.Server != "" {
			override.Fields.Server = strings.ToLower(strings.TrimSpace(override.Fields.Server))
		}
		if override.Fields.TLS != nil {
			override.Fields.TLS.SNI = sanitizeText(override.Fields.TLS.SNI)
			override.Fields.TLS.Fingerprint = sanitizeText(override.Fields.TLS.Fingerprint)
			override.Fields.TLS.ClientFingerprint = sanitizeText(override.Fields.TLS.ClientFingerprint)
			override.Fields.TLS.ALPN = cleanStringSlice(override.Fields.TLS.ALPN)
		}
		if override.Fields.Transport != nil {
			override.Fields.Transport.Network = sanitizeText(override.Fields.Transport.Network)
			override.Fields.Transport.Path = sanitizeText(override.Fields.Transport.Path)
			override.Fields.Transport.Host = sanitizeText(override.Fields.Transport.Host)
			override.Fields.Transport.ServiceName = sanitizeText(override.Fields.Transport.ServiceName)
			override.Fields.Transport.Mode = sanitizeText(override.Fields.Transport.Mode)
			override.Fields.Transport.H2Hosts = cleanStringSlice(override.Fields.Transport.H2Hosts)
		}
		if cleanID == "" {
			delete(state.NodeOverrides, id)
			continue
		}
		if cleanID != id {
			delete(state.NodeOverrides, id)
		}
		state.NodeOverrides[cleanID] = override
	}

	normalizedCustom := make([]NodeIR, 0, len(state.CustomNodes))
	seenCustom := map[string]struct{}{}
	for _, node := range state.CustomNodes {
		if strings.TrimSpace(node.Source.Kind) == "" {
			node.Source.Kind = "custom"
		}
		if strings.TrimSpace(node.Source.ID) == "" {
			node.Source.ID = "custom"
		}
		if strings.TrimSpace(node.Source.Name) == "" {
			node.Source.Name = "手动节点"
		}
		node = NormalizeNode(node)
		if strings.TrimSpace(node.ID) != "" && !strings.HasPrefix(node.ID, "custom-") {
			node.ID = "custom-" + node.ID
		}
		if _, ok := seenCustom[node.ID]; ok {
			continue
		}
		seenCustom[node.ID] = struct{}{}
		normalizedCustom = append(normalizedCustom, node)
	}
	state.CustomNodes = normalizedCustom

	return state
}

func CloneNode(node NodeIR) NodeIR {
	raw, _ := json.Marshal(node)
	var cloned NodeIR
	_ = json.Unmarshal(raw, &cloned)
	return cloned
}

func CloneNodes(nodes []NodeIR) []NodeIR {
	out := make([]NodeIR, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, CloneNode(node))
	}
	return out
}

func DisabledNodeSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = sanitizeText(id)
		if id == "" {
			continue
		}
		set[id] = struct{}{}
	}
	return set
}

func ReplaceRegionTag(tags []string, region string) []string {
	var out []string
	region = strings.ToUpper(sanitizeText(region))
	for _, tag := range cleanStringSlice(tags) {
		if isRegionTag(tag) {
			continue
		}
		out = append(out, tag)
	}
	if region != "" {
		out = append(out, region)
	}
	sort.Strings(out)
	return out
}

func NodeRegionCode(node NodeIR) string {
	for _, tag := range node.Tags {
		upper := strings.ToUpper(strings.TrimSpace(tag))
		if isRegionTag(upper) {
			return upper
		}
	}
	return "OTHER"
}

func NodeRegionLabel(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "HK":
		return "香港"
	case "JP":
		return "日本"
	case "US":
		return "美国"
	case "SG":
		return "新加坡"
	case "TW":
		return "台湾"
	case "KR":
		return "韩国"
	case "GB":
		return "英国"
	case "DE":
		return "德国"
	case "NL":
		return "荷兰"
	case "RU":
		return "俄罗斯"
	default:
		return "其它"
	}
}

func NodeRegionEmoji(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "HK":
		return "🇭🇰"
	case "JP":
		return "🇯🇵"
	case "US":
		return "🇺🇸"
	case "SG":
		return "🇸🇬"
	case "TW":
		return "🇹🇼"
	case "KR":
		return "🇰🇷"
	case "GB":
		return "🇬🇧"
	case "DE":
		return "🇩🇪"
	case "NL":
		return "🇳🇱"
	case "RU":
		return "🇷🇺"
	default:
		return ""
	}
}

func isRegionTag(tag string) bool {
	switch strings.ToUpper(strings.TrimSpace(tag)) {
	case "HK", "JP", "US", "SG", "TW", "KR", "GB", "DE", "NL", "RU", "FR", "CA", "AU", "OTHER":
		return true
	default:
		return false
	}
}
