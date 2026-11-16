package api

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"subconv-next/internal/model"
)

func TestEnsureWorkspacePublishedRefKeepsTokenWhenEmptyPublicationHasNoYAML(t *testing.T) {
	server, cfg := newTestServer(t, model.DefaultConfig())
	ref := createWorkspaceRefForTest(t, server, cfg)
	first, created, err := server.ensureWorkspacePublishedRef(&ref)
	if err != nil || !created {
		t.Fatalf("initial ensure = id %q created %v err %v", first.ID, created, err)
	}
	if err := os.Remove(first.CurrentPath); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	second, created, err := server.ensureWorkspacePublishedRef(&ref)
	if err != nil {
		t.Fatalf("re-ensure error = %v", err)
	}
	if created || second.ID != first.ID || second.Meta.Token != first.Meta.Token {
		t.Fatalf("empty publication was replaced: first=%+v second=%+v created=%v", first.Meta, second.Meta, created)
	}
}

func TestRegeneratingAfterDeletingLastNodeKeepsOriginalSubscriptionLink(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@one.example:443#last-node"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	first := publishNodesForDeletionTest(t, server, ref.ID)
	ids := deletionNodeIDsByName(t, server, ref.ID)
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{ids["last-node"]}}, http.StatusOK)
	if rec := publishedDeletionRequest(server, first.SubscriptionURL); rec.Code != http.StatusNotFound {
		t.Fatalf("empty publication returned %d, want 404", rec.Code)
	}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/enable", toggleNodesRequest{IDs: []string{ids["last-node"]}}, http.StatusOK)
	var regenerated refreshResponse
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/refresh", nil, http.StatusOK)
	if err := json.Unmarshal(data, &regenerated); err != nil {
		t.Fatal(err)
	}
	if regenerated.PublishID != first.PublishID || regenerated.SubscriptionURL != first.SubscriptionURL || regenerated.NodeCount != 1 {
		t.Fatalf("regeneration changed publication: old id=%s new id=%s count=%d", first.PublishID, regenerated.PublishID, regenerated.NodeCount)
	}
	if rec := publishedDeletionRequest(server, first.SubscriptionURL); rec.Code != http.StatusOK {
		t.Fatalf("original link did not recover: status=%d", rec.Code)
	}
}
