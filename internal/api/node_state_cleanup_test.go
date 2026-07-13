package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

func TestPruneDeletedNodeStateKeepsSharedAndUnknownNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "sub-active"}}
	state := model.NodeState{
		DeletedNodes: []string{"removed", "shared", "legacy", "unknown"},
		DeletedNodeSources: map[string][]string{
			"removed": {"sub-removed"},
			"shared":  {"sub-removed", "sub-active"},
		},
		LastAudit: model.AuditReport{ExcludedNodes: []model.ExcludedNode{
			{
				ID:     "legacy",
				Reason: "deleted_node",
				Source: model.SourceInfo{ID: "sub-removed", Kind: "subscription"},
			},
		}},
	}

	state, changed, pruned := pruneDeletedNodeStateForCurrentConfig(cfg, state, nil, false)
	if !changed || pruned != 2 {
		t.Fatalf("cleanup changed=%v pruned=%d, want true/2", changed, pruned)
	}
	if containsString(state.DeletedNodes, "removed") || containsString(state.DeletedNodes, "legacy") {
		t.Fatalf("removed-source records remain: %#v", state.DeletedNodes)
	}
	if !containsString(state.DeletedNodes, "shared") || !containsString(state.DeletedNodes, "unknown") {
		t.Fatalf("shared or unknown records were pruned: %#v", state.DeletedNodes)
	}
	if _, ok := state.DeletedNodeSources["shared"]; !ok {
		t.Fatalf("shared node source evidence was removed: %#v", state.DeletedNodeSources)
	}
}

func TestPruneDeletedNodeStateDropsLegacyUnknownWhenSourcesLoaded(t *testing.T) {
	cfg := model.DefaultConfig()
	state := model.NodeState{DeletedNodes: []string{"legacy-missing"}}

	state, changed, pruned := pruneDeletedNodeStateForCurrentConfig(cfg, state, nil, true)
	if !changed || pruned != 1 || len(state.DeletedNodes) != 0 {
		t.Fatalf("cleanup changed=%v pruned=%d state=%#v", changed, pruned, state)
	}
}

func TestRestoredDraftPrunesLegacyUnknownDeletedNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	state := pruneRestoredDraftDeletedNodeState(cfg, model.NodeState{
		DeletedNodes: []string{"legacy-missing"},
	})
	if len(state.DeletedNodes) != 0 {
		t.Fatalf("restored draft deleted nodes = %#v, want empty", state.DeletedNodes)
	}
}

func TestConfigCleanupUsesStoredSourcesWithoutRefetchingSubscriptions(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("unused"))
	}))
	defer upstream.Close()

	oldCfg := model.DefaultConfig()
	oldCfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-removed", Name: "removed", Enabled: true, URL: upstream.URL},
	}
	newCfg := oldCfg
	newCfg.Subscriptions = nil
	state := model.NodeState{
		DeletedNodes: []string{"node-removed"},
		DeletedNodeSources: map[string][]string{
			"node-removed": {"sub-removed"},
		},
	}

	state, changed, pruned := pruneDeletedNodeStateForConfigChange(oldCfg, newCfg, state)
	if !changed || pruned != 1 || len(state.DeletedNodes) != 0 {
		t.Fatalf("cleanup changed=%v pruned=%d state=%#v", changed, pruned, state)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("upstream requests = %d, want 0 when source evidence is stored", got)
	}
}

func TestConfigUpdatePrunesDeletedNodesFromRemovedSubscription(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Subscriptions = []model.SubscriptionConfig{
		{ID: "sub-removed", Name: "removed", Enabled: false, URL: "https://example.invalid/sub"},
	}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	if err := pipeline.SaveNodeState(workspaceCfg, model.NodeState{
		DeletedNodes: []string{"node-removed"},
		DeletedNodeSources: map[string][]string{
			"node-removed": {"sub-removed"},
		},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	nextCfg := workspaceCfg
	nextCfg.Subscriptions = []model.SubscriptionConfig{}
	body, err := json.Marshal(nextCfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, withWorkspace("/api/config", ref.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("config update status = %d; body=%s", rec.Code, rec.Body.String())
	}

	updatedCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("load updated config: %v", err)
	}
	updatedState, err := pipeline.LoadNodeState(updatedCfg)
	if err != nil {
		t.Fatalf("LoadNodeState() error = %v", err)
	}
	if len(updatedState.DeletedNodes) != 0 || len(updatedState.DeletedNodeSources) != 0 {
		t.Fatalf("stale deleted state remains: %#v", updatedState)
	}
}

func TestDeletedNodesListPrunesLegacyRemovedSubscriptionRecord(t *testing.T) {
	cfg := model.DefaultConfig()
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatalf("loadWorkspaceConfig() error = %v", err)
	}
	if err := pipeline.SaveNodeState(workspaceCfg, model.NodeState{
		DeletedNodes: []string{"legacy-node"},
		LastAudit: model.AuditReport{ExcludedNodes: []model.ExcludedNode{
			{
				ID:     "legacy-node",
				Reason: "deleted_node",
				Source: model.SourceInfo{ID: "sub-removed", Kind: "subscription"},
			},
		}},
	}); err != nil {
		t.Fatalf("SaveNodeState() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, withWorkspace("/api/nodes/deleted", ref.ID), nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("deleted nodes status = %d; body=%s", rec.Code, rec.Body.String())
	}
	var response deletedNodeListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Nodes) != 0 || len(response.MissingIDs) != 0 {
		t.Fatalf("deleted response = %#v, want no stale records", response)
	}
}
