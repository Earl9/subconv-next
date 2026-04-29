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
	if len(finalSet.Nodes) != 1 || finalSet.Nodes[0].Name != "[ss] keep" {
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
	if len(finalSet.Nodes) != 1 || finalSet.Nodes[0].Name != "[ss] ok-node" {
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

func TestValidateOutputNoLeakInfoNodeNotInURLTest(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Render.IncludeInfoNode = true
	cfg.Render.ShowInfoNodes = true
	infoNode := model.NormalizeNode(model.NodeIR{Name: "剩余流量：100GB", Type: model.ProtocolSS, Source: model.SourceInfo{Name: "SecOne"}, Raw: map[string]interface{}{"_infoNode": true}})
	finalSet := FinalNodeSet{Nodes: []model.NodeIR{infoNode}}
	yamlBytes := []byte(`
proxies:
  - {name: "剩余流量：100GB", type: ss, server: info.example.com, port: 443, cipher: aes-256-gcm, password: p}
proxy-groups:
  - {name: "⚡ 自动选择", type: url-test, proxies: ["剩余流量：100GB"], url: "https://www.gstatic.com/generate_204", interval: 300}
rules:
  - MATCH,🐟 漏网之鱼
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "info node") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want info node url-test failure", err)
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
  - {name: "🚀 节点选择", type: select, proxies: ["ok-node", DIRECT]}
rules:
  - MATCH,🚀 节点选择
  - DOMAIN,example.com,🚀 节点选择
`)
	if err := ValidateOutputNoLeak(yamlBytes, finalSet, model.AuditReport{}, renderer.OptionsFromConfig(cfg)); err == nil || !strings.Contains(err.Error(), "MATCH rule must be last") {
		t.Fatalf("ValidateOutputNoLeak() error = %v, want MATCH last failure", err)
	}
}
