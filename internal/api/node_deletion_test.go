package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

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
