package pipeline

import (
	"strings"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/renderer"
)

func TestBuildFinalNodesDisabledAndDeletedNotRendered(t *testing.T) {
	cfg := model.DefaultConfig()
	rawNodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{Name: "keep", Type: model.ProtocolSS, Server: "a.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}}),
		model.NormalizeNode(model.NodeIR{Name: "disabled", Type: model.ProtocolSS, Server: "b.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}}),
		model.NormalizeNode(model.NodeIR{Name: "deleted", Type: model.ProtocolSS, Server: "c.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}}),
	}
	rawNodes[0].ID = model.StableNodeID(rawNodes[0])
	rawNodes[1].ID = model.StableNodeID(rawNodes[1])
	rawNodes[2].ID = model.StableNodeID(rawNodes[2])

	finalSet, audit, err := BuildFinalNodes(cfg, model.NodeState{
		DisabledNodes: []string{rawNodes[1].ID},
		DeletedNodes:  []string{rawNodes[2].ID},
	}, rawNodes)
	if err != nil {
		t.Fatalf("BuildFinalNodes() error = %v", err)
	}
	if len(finalSet.Nodes) != 1 || finalSet.Nodes[0].Name != "keep" {
		t.Fatalf("final nodes = %#v, want only keep", finalSet.Nodes)
	}
	reasons := map[string]struct{}{}
	for _, item := range audit.ExcludedNodes {
		reasons[item.Reason] = struct{}{}
	}
	if _, ok := reasons["disabled_node"]; !ok {
		t.Fatalf("audit missing disabled_node: %#v", audit.ExcludedNodes)
	}
	if _, ok := reasons["deleted_node"]; !ok {
		t.Fatalf("audit missing deleted_node: %#v", audit.ExcludedNodes)
	}
}

func TestBuildFinalNodesExcludeKeywordAndInfoNode(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.ExcludeKeywords = "test"
	rawNodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{Name: "剩余流量：100GB", Type: model.ProtocolSS, Source: model.SourceInfo{Name: "SecOne"}}),
		model.NormalizeNode(model.NodeIR{Name: "test-node", Type: model.ProtocolSS, Server: "a.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}}),
		model.NormalizeNode(model.NodeIR{Name: "ok-node", Type: model.ProtocolSS, Server: "b.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}}),
	}
	finalSet, audit, err := BuildFinalNodes(cfg, model.DefaultNodeState(), rawNodes)
	if err != nil {
		t.Fatalf("BuildFinalNodes() error = %v", err)
	}
	if len(finalSet.Nodes) != 1 || finalSet.Nodes[0].Name != "ok-node" {
		t.Fatalf("final nodes = %#v, want only ok-node", finalSet.Nodes)
	}
	reasonCount := map[string]int{}
	for _, item := range audit.ExcludedNodes {
		reasonCount[item.Reason]++
	}
	if reasonCount["info_node"] == 0 || reasonCount["exclude_keyword_matched"] == 0 {
		t.Fatalf("audit reasons = %#v, want info_node and exclude_keyword_matched", reasonCount)
	}
}

func TestValidateOutputNoLeakDetectsExcludedLeak(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "ok-node", Type: model.ProtocolSS, Server: "b.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	audit := model.AuditReport{
		ExcludedNodes: []model.ExcludedNode{
			{Name: "leaked-node", Reason: "exclude_keyword_matched", Source: model.SourceInfo{Name: "SecOne"}},
		},
	}

	rendered, err := renderer.RenderMihomo([]model.NodeIR{
		node,
		model.NormalizeNode(model.NodeIR{Name: "leaked-node", Type: model.ProtocolSS, Server: "c.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}}),
	}, renderer.OptionsFromConfig(cfg))
	if err != nil {
		t.Fatalf("RenderMihomo() error = %v", err)
	}

	if err := ValidateOutputNoLeak(rendered, finalSet, audit, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "leaked-node") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want leaked-node failure", err)
	}
}

func TestValidateFinalNodeSetRejectsFilteredAndInvalidNodes(t *testing.T) {
	invalid := model.NormalizeNode(model.NodeIR{
		Name: "invalid-node",
		Type: model.ProtocolSS,
		Raw:  map[string]interface{}{"method": "aes-256-gcm"},
	})
	if err := ValidateFinalNodeSet(FinalNodeSet{Nodes: []model.NodeIR{invalid}}, model.AuditReport{}); err == nil || !strings.Contains(err.Error(), "invalid-node") {
		t.Fatalf("ValidateFinalNodeSet() error = %v, want invalid-node failure", err)
	}

	disabled := model.NormalizeNode(model.NodeIR{
		Name:   "disabled-node",
		Type:   model.ProtocolSS,
		Server: "disabled.example.com",
		Port:   443,
		Auth:   model.Auth{Password: "p"},
		Raw:    map[string]interface{}{"method": "aes-256-gcm"},
	})
	audit := model.AuditReport{ExcludedNodes: []model.ExcludedNode{
		{Name: "disabled-node", Reason: "disabled_node"},
		{Name: "deleted-node", Reason: "deleted_node"},
		{Name: "excluded-node", Reason: "exclude_keyword_matched"},
		{Name: "invalid-node", Reason: "invalid_node"},
	}}
	if err := ValidateFinalNodeSet(FinalNodeSet{Nodes: []model.NodeIR{disabled}}, audit); err == nil || !strings.Contains(err.Error(), "disabled-node") {
		t.Fatalf("ValidateFinalNodeSet() error = %v, want disabled-node leak failure", err)
	}
}

func TestValidateOutputNoLeakInfoNodeNotInURLTest(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.IncludeInfoNode = true
	cfg.Render.ShowInfoNodes = true
	infoNode := model.NormalizeNode(model.NodeIR{Name: "剩余流量：100GB", Type: model.ProtocolSS, Server: "info.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Source: model.SourceInfo{Name: "SecOne"}, Raw: map[string]interface{}{"method": "aes-256-gcm", "_infoNode": true}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{infoNode}}
	yamlBytes := []byte(`
proxies:
  - {name: "剩余流量：100GB", type: ss, server: info.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "⚡ 自动选择", type: url-test, proxies: ["剩余流量：100GB"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🐟 漏网之鱼
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "must only reference real node") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want global auto real-node-only failure", err)
	}
}

func TestValidateOutputNoLeakGlobalAutoOnlyRealNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.IncludeInfoNode = true
	cfg.Render.ShowInfoNodes = true
	realNode := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	infoNode := model.NormalizeNode(model.NodeIR{Name: "剩余流量：100GB", Type: model.ProtocolSS, Server: "info.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Source: model.SourceInfo{Name: "SecOne"}, Raw: map[string]interface{}{"method": "aes-256-gcm", "_infoNode": true}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{realNode, infoNode}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
  - {name: "剩余流量：100GB", type: ss, server: info.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "US Node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node", "剩余流量：100GB"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🚀 节点选择
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "must only reference real node") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want global auto real-node-only failure", err)
	}
}

func TestValidateOutputNoLeakRejectsUnsupportedBuiltinReference(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["US Node", REJECT-DROP]}
rules:
  - MATCH,🚀 节点选择
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "REJECT-DROP") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want unsupported builtin reference failure", err)
	}
}

func TestValidateFinalConfigRequiresMainSelectEssentials(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🚀 节点选择
`)
	if err := ValidateFinalConfig(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), `must contain "REJECT"`) {
		t.Fatalf("ValidateFinalConfig() error = %v, want missing REJECT failure", err)
	}
}

func TestValidateFinalConfigRequiresMainSelectRealNode(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🚀 节点选择
`)
	if err := ValidateFinalConfig(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "must include at least one real node") {
		t.Fatalf("ValidateFinalConfig() error = %v, want missing main real node failure", err)
	}
}

func TestValidateFinalConfigSpecialGroupsStayCompact(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "US Node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node"], url: "https://www.gstatic.com/generate_204", interval: 300}
  - {name: "🛑 广告拦截", type: select, proxies: [REJECT, DIRECT, "🚀 节点选择", "US Node"]}
rules:
  - RULE-SET,adblock,🛑 广告拦截
  - MATCH,🚀 节点选择
rule-providers:
  adblock: {type: http, behavior: domain, url: "https://example.com/adblock.yaml", path: ./ruleset/adblock.yaml}
`)
	if err := ValidateFinalConfig(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "must stay compact") {
		t.Fatalf("ValidateFinalConfig() error = %v, want special compact failure", err)
	}
}

func TestValidateFinalConfigRejectsFilteredGroupReference(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	audit := model.AuditReport{
		ExcludedNodes: []model.ExcludedNode{
			{Name: "disabled-node", Reason: "disabled_node"},
		},
	}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "US Node", "disabled-node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🚀 节点选择
`)
	if err := ValidateFinalConfig(yamlBytes, finalSet, audit, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "disabled-node") {
		t.Fatalf("ValidateFinalConfig() error = %v, want filtered node reference failure", err)
	}
}

func TestValidateOutputNoLeakRejectsRegionGroupGlobalAuto(t *testing.T) {
	cfg := model.DefaultConfig()
	hkNode := model.NormalizeNode(model.NodeIR{Name: "HK Node", Type: model.ProtocolSS, Server: "hk.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	jpNode := model.NormalizeNode(model.NodeIR{Name: "JP Node", Type: model.ProtocolSS, Server: "jp.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{hkNode, jpNode}}
	yamlBytes := []byte(`
proxies:
  - {name: "HK Node", type: ss, server: hk.example.com, port: 443, cipher: aes-256-gcm, password: p}
  - {name: "JP Node", type: ss, server: jp.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", "🇭🇰 香港", "🇯🇵 日本"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["HK Node", "JP Node"]}
  - {name: "🇭🇰 香港", type: select, proxies: ["⚡ 自动选择", DIRECT, "HK Node"]}
  - {name: "🇯🇵 日本", type: select, proxies: [DIRECT, "JP Node"]}
rules:
  - MATCH,🚀 节点选择
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "region proxy-group") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want region proxy-group failure", err)
	}
}

func TestValidateOutputNoLeakRequiresRuleGroupRealNodeInFullMode(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "US Node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node"], url: "https://www.gstatic.com/generate_204", interval: 300}
  - {name: "🤖 AI 服务", type: select, proxies: ["🚀 节点选择", "⚡ 自动选择", DIRECT, REJECT]}
rules:
  - RULE-SET,openai,🤖 AI 服务
  - MATCH,🚀 节点选择
rule-providers:
  openai: {type: http, behavior: domain, url: "https://example.com/openai.yaml", path: ./ruleset/openai.yaml}
`)
	opts := renderer.OptionsFromConfig(cfg)
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{RuleGroupNodeMode: "full"})
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, opts); err == nil || !strings.Contains(err.Error(), "must include at least one real node") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want missing real node failure", err)
	}
}

func TestValidateOutputNoLeakAllowsRuleGroupCompactMode(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "US Node", Type: model.ProtocolSS, Server: "us.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "US Node", type: ss, server: us.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "US Node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["US Node"], url: "https://www.gstatic.com/generate_204", interval: 300}
  - {name: "🤖 AI 服务", type: select, proxies: ["🚀 节点选择", "⚡ 自动选择", DIRECT, REJECT]}
rules:
  - RULE-SET,openai,🤖 AI 服务
  - MATCH,🚀 节点选择
rule-providers:
  openai: {type: http, behavior: domain, url: "https://example.com/openai.yaml", path: ./ruleset/openai.yaml}
`)
	opts := renderer.OptionsFromConfig(cfg)
	opts.GroupOptions = model.NormalizeGroupOptions(model.GroupOptions{RuleGroupNodeMode: "compact"})
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, opts); err != nil {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want nil", err)
	}
}

func TestValidateOutputNoLeakRejectsEarlyMatch(t *testing.T) {
	cfg := model.DefaultConfig()
	node := model.NormalizeNode(model.NodeIR{Name: "ok-node", Type: model.ProtocolSS, Server: "b.example.com", Port: 443, Auth: model.Auth{Password: "p"}, Raw: map[string]interface{}{"method": "aes-256-gcm"}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{node}}
	yamlBytes := []byte(`
proxies:
  - {name: "ok-node", type: ss, server: b.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "🚀 节点选择", type: select, proxies: ["⚡ 自动选择", DIRECT, REJECT, "ok-node"]}
  - {name: "⚡ 自动选择", type: url-test, proxies: ["ok-node"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🚀 节点选择
  - DOMAIN,example.com,🚀 节点选择
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "MATCH rule must be last") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want MATCH last failure", err)
	}
}
