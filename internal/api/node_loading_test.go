package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"subconv-next/internal/model"
)

func TestNodeListIncludesValidationForOffPageNodes(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://test@keep.example:443#keep\nvless://test@warning.example:443?type=xhttp&security=tls#warning"}}
	server, cfg := newTestServer(t, cfg)
	w := createWorkspaceForTest(t, server, cfg)
	var list nodeListResponse
	data := nodeDeletionRequest(t, server, w, http.MethodGet, "/api/nodes?page_size=1&q=keep", nil, http.StatusOK)
	if err := json.Unmarshal(data, &list); err != nil {
		t.Fatal(err)
	}
	var validation validateNodesResponse
	data = nodeDeletionRequest(t, server, w, http.MethodPost, "/api/nodes/validate", nil, http.StatusOK)
	if err := json.Unmarshal(data, &validation); err != nil {
		t.Fatal(err)
	}
	if len(list.Nodes) != 1 || len(list.ValidationWarnings) != 1 || len(validation.Warnings) != 1 || list.ValidationWarnings[0] != validation.Warnings[0] {
		t.Fatal("paginated node response lost global validation warnings")
	}
}

func TestNodeEditingReusesSubscriptionsUntilRefresh(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := "original-node"
		host := "example.com"
		if requests.Add(1) > 1 {
			name = "updated-node"
			host = "updated.example.com"
		}
		_, _ = w.Write([]byte("trojan://test-password@" + host + ":443#" + name))
	}))
	defer upstream.Close()

	cfg := model.DefaultConfig()
	cfg.Subscriptions = []model.SubscriptionConfig{{
		ID: "source", Name: "source", Enabled: true, URL: upstream.URL, AllowLAN: true,
	}}
	server, cfg := newTestServer(t, cfg)
	workspaceID := createWorkspaceForTest(t, server, cfg)
	handler := server.Handler()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(method, withWorkspace(path, workspaceID), strings.NewReader(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %s: status = %d, body = %s", method, path, rec.Code, rec.Body.String())
		}
		return rec
	}
	listNodes := func(path string) nodeListResponse {
		t.Helper()
		var response nodeListResponse
		if err := json.Unmarshal(request(http.MethodGet, path, "").Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if !response.OK || len(response.Nodes) != 1 {
			t.Fatalf("nodes = %+v, want one node", response)
		}
		return response
	}

	request(http.MethodPost, "/api/refresh", "")
	nodes := listNodes("/api/nodes?page=1&page_size=25")
	if nodes.Nodes[0].Name != "original-node" || requests.Load() != 1 {
		t.Fatalf("node list refetched upstream: name = %q, requests = %d", nodes.Nodes[0].Name, requests.Load())
	}
	id := nodes.Nodes[0].ID
	request(http.MethodGet, "/api/nodes/"+id, "")
	request(http.MethodPut, "/api/nodes/"+id+"/override", `{"enabled":true,"name":"renamed-node"}`)
	nodes = listNodes("/api/nodes?all=1&q=renamed")
	if nodes.Nodes[0].Name != "renamed-node" || !nodes.Nodes[0].Modified {
		t.Fatalf("node override was not applied: %+v", nodes.Nodes[0])
	}
	request(http.MethodPost, "/api/nodes/validate", "")
	request(http.MethodGet, "/api/nodes/deleted", "")
	preview := request(http.MethodGet, "/api/preview-yaml", "")
	if !strings.Contains(preview.Body.String(), "original-node") {
		t.Fatal("preview did not return the generated YAML")
	}
	if requests.Load() != 1 {
		t.Fatalf("editing or preview refetched upstream: requests = %d, want 1", requests.Load())
	}

	request(http.MethodPost, "/api/refresh", "")
	nodes = listNodes("/api/nodes?all=1")
	if requests.Load() != 2 {
		t.Fatalf("explicit refresh requests = %d, want 2", requests.Load())
	}
	if nodes.Nodes[0].Name != "updated-node" {
		t.Fatalf("node after refresh = %q, want updated-node", nodes.Nodes[0].Name)
	}
}
