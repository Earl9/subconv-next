package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

func TestDeletedSubscriptionNodesStayDeletedAfterEndpointRotation(t *testing.T) {
	for _, scope := range []string{"global", "per_source", "none"} {
		t.Run(scope, func(t *testing.T) {
			var revision atomic.Int32
			names := []string{"剩余流量：100 GB", "获取使用教程", "移动推荐使用优化节点"}
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for i, name := range names {
					fmt.Fprintf(w, "trojan://secret-%d@node-%d-%d.example:443#%s\n", revision.Load(), i, revision.Load(), url.PathEscape(name))
				}
				fmt.Fprintln(w, "trojan://keep@keep.example:443#keep")
			}))
			defer upstream.Close()
			cfg := model.DefaultConfig()
			cfg.Render.DedupeScope = scope
			cfg.Render.ShowInfoNodes, cfg.Render.IncludeInfoNode = true, true
			cfg.Subscriptions = []model.SubscriptionConfig{{ID: "rotating", Name: "rotating", URL: upstream.URL, Enabled: true, AllowLAN: true}}
			server, cfg := newTestServer(t, cfg)
			workspace := createWorkspaceForTest(t, server, cfg)
			request := func(method, path string, body interface{}) []byte {
				return nodeDeletionRequest(t, server, workspace, method, path, body, http.StatusOK)
			}
			list := func(path string) nodeListResponse {
				var result nodeListResponse
				if err := json.Unmarshal(request(http.MethodGet, path, nil), &result); err != nil {
					t.Fatal(err)
				}
				return result
			}
			request(http.MethodPost, "/api/refresh", nil)
			initial := list("/api/nodes?all=1")
			if len(initial.Nodes) != 4 {
				t.Fatalf("initial nodes = %d", len(initial.Nodes))
			}
			var ids []string
			for _, node := range initial.Nodes {
				if node.Name != "keep" {
					ids = append(ids, node.ID)
				}
			}
			// A user rename must not replace the original subscription identity.
			request(http.MethodPut, "/api/nodes/"+ids[1]+"/override", nodeOverrideRequest{Enabled: true, Name: "my renamed entry"})
			request(http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: ids})
			var saved nodeStateResponse
			if err := json.Unmarshal(request(http.MethodGet, "/api/nodes/state", nil), &saved); err != nil {
				t.Fatal(err)
			}
			if len(saved.State.DeletedNodeSignatures) != 3 {
				t.Fatalf("persisted identities = %d, want 3", len(saved.State.DeletedNodeSignatures))
			}
			// Exercise JSON draft export/restore, then real upstream fetching.
			request(http.MethodPut, "/api/nodes/state", nodeStateRequest{State: saved.State})
			for version := int32(1); version <= 2; version++ {
				revision.Store(version)
				var refreshed refreshResponse
				if err := json.Unmarshal(request(http.MethodPost, "/api/refresh", nil), &refreshed); err != nil {
					t.Fatal(err)
				}
				if refreshed.NodeCount != 1 {
					t.Fatalf("refresh %d resurrected deleted nodes: count=%d", version, refreshed.NodeCount)
				}
				if nodes := list("/api/nodes?all=1"); len(nodes.Nodes) != 1 || nodes.Nodes[0].Name != "keep" {
					t.Fatalf("refresh %d editor nodes = %#v", version, nodes.Nodes)
				}
			}
			var renamed bulkRenameResponse
			if err := json.Unmarshal(request(http.MethodPost, "/api/nodes/bulk-rename", bulkRenameRequest{Scope: "current_filtered", Mode: "add_prefix", Prefix: "kept-"}), &renamed); err != nil {
				t.Fatal(err)
			}
			if renamed.Changed != 1 {
				t.Fatalf("bulk rename included hidden deleted nodes: %d changes", renamed.Changed)
			}
			var deleted deletedNodeListResponse
			if err := json.Unmarshal(request(http.MethodGet, "/api/nodes/deleted", nil), &deleted); err != nil {
				t.Fatal(err)
			}
			if len(deleted.Nodes) != 3 || len(deleted.MissingIDs) != 0 {
				t.Fatalf("deleted entries not restorable: %#v", deleted)
			}
			var currentIDs []string
			for _, node := range deleted.Nodes {
				if containsString(ids, node.ID) {
					t.Fatal("test did not rotate connection IDs")
				}
				currentIDs = append(currentIDs, node.ID)
			}
			request(http.MethodPost, "/api/nodes/enable", toggleNodesRequest{IDs: currentIDs})
			if nodes := list("/api/nodes?all=1"); len(nodes.Nodes) != 4 {
				t.Fatalf("restore returned %d nodes, want 4", len(nodes.Nodes))
			}
			saved = nodeStateResponse{}
			if err := json.Unmarshal(request(http.MethodGet, "/api/nodes/state", nil), &saved); err != nil {
				t.Fatal(err)
			}
			if len(saved.State.DeletedNodes) != 0 || len(saved.State.DeletedNodeSignatures) != 0 {
				t.Fatalf("restore retained old deletion records: %#v", saved.State)
			}
		})
	}
}

func nodeDeletionRequest(t *testing.T, server *Server, workspace, method, path string, body interface{}, status int) []byte {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, withWorkspace(path, workspace), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s: status %d, want %d: %s", method, path, rec.Code, status, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func TestBulkDeleteMixedNodesAcrossDedupeScopes(t *testing.T) {
	for _, scope := range []string{"global", "per_source", "none"} {
		t.Run(scope, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Render.DedupeScope = scope
			cfg.Inline = []model.InlineConfig{{ID: "inline", Name: "inline", Enabled: true, Content: "trojan://test@info.example:443#剩余流量\ntrojan://test@delete.example:443#delete-me\ntrojan://test@keep.example:443#keep"}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/custom", customNodeRequest{Content: "trojan://test@custom1.example:443#custom1\ntrojan://test@custom2.example:443#custom2\ntrojan://test@custom3.example:443#custom3"}, http.StatusOK)
			var list nodeListResponse
			data := nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/nodes?all=1", nil, http.StatusOK)
			if err := json.Unmarshal(data, &list); err != nil || len(list.Nodes) != 6 {
				t.Fatalf("initial list = %s, err=%v", data, err)
			}
			var ids []string
			for _, node := range list.Nodes {
				if node.Source.Kind == "custom" || node.Name == "delete-me" {
					ids = append(ids, node.ID)
				}
			}
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/disable", toggleNodesRequest{IDs: ids}, http.StatusOK)
			// A stale selection must fail before any node is removed.
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: append(append([]string{}, ids...), "missing")}, http.StatusNotFound)
			data = nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/nodes?all=1", nil, http.StatusOK)
			if err := json.Unmarshal(data, &list); err != nil || len(list.Nodes) != 6 {
				t.Fatalf("failed batch partially deleted nodes: %s", data)
			}
			// Duplicate IDs count once; disabled nodes remain deletable.
			data = nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: append(ids, ids[0])}, http.StatusOK)
			var deleted deleteNodesResponse
			if err := json.Unmarshal(data, &deleted); err != nil || deleted.DeletedCount != 4 || len(deleted.DeletedIDs) != 4 {
				t.Fatalf("unexpected deletion result: %s", data)
			}
			data = nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/nodes?all=1", nil, http.StatusOK)
			if err := json.Unmarshal(data, &list); err != nil || len(list.Nodes) != 2 {
				t.Fatalf("nodes remain after deletion: %s", data)
			}
			workspaceCfg, err := server.loadWorkspaceConfig(ref)
			if err != nil {
				t.Fatal(err)
			}
			state, err := pipeline.LoadNodeState(workspaceCfg)
			if err != nil || len(state.CustomNodes) != 0 || len(state.DeletedNodes) != 1 || len(state.DisabledNodes) != 0 {
				t.Fatalf("invalid persisted deletion state: %#v, err=%v", state, err)
			}
			collected := pipeline.CollectNodesWithState(workspaceCfg, state, true, false)
			final, _, err := pipeline.BuildFinalNodes(workspaceCfg, state, collected.Nodes)
			if err != nil || len(final.Nodes) != 1 || final.Nodes[0].Server != "keep.example" {
				t.Fatalf("deleted node returned on generation: %#v, err=%v", final.Nodes, err)
			}
		})
	}
}

func TestDeleteCustomNodeUsesEditorID(t *testing.T) {
	for _, scope := range []string{"global", "per_source", "none"} {
		t.Run(scope, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Render.DedupeScope = scope
			server, cfg := newTestServer(t, cfg)
			workspace := createWorkspaceForTest(t, server, cfg)
			nodeDeletionRequest(t, server, workspace, http.MethodPost, "/api/nodes/custom", customNodeRequest{Content: "trojan://test@manual.example:443#manual"}, http.StatusOK)
			var list nodeListResponse
			data := nodeDeletionRequest(t, server, workspace, http.MethodGet, "/api/nodes?all=1", nil, http.StatusOK)
			if err := json.Unmarshal(data, &list); err != nil || len(list.Nodes) != 1 {
				t.Fatalf("initial list: %s", data)
			}
			path := "/api/nodes/custom/" + list.Nodes[0].ID
			nodeDeletionRequest(t, server, workspace, http.MethodDelete, path, nil, http.StatusOK)
			nodeDeletionRequest(t, server, workspace, http.MethodDelete, path, nil, http.StatusNotFound)
			data = nodeDeletionRequest(t, server, workspace, http.MethodGet, "/api/nodes?all=1", nil, http.StatusOK)
			if err := json.Unmarshal(data, &list); err != nil || len(list.Nodes) != 0 {
				t.Fatalf("custom node survived deletion: %s", data)
			}
		})
	}
}
