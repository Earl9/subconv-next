package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

func TestRestoringDeletedNodesImmediatelyUpdatesPublishedSubscription(t *testing.T) {
	for _, mode := range []string{"single", "bulk", "all"} {
		t.Run(mode, func(t *testing.T) {
			var requests atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.Header().Set("Subscription-Userinfo", "upload=10; download=20; total=1000; expire=1900000000")
				fmt.Fprintln(w, "trojan://fixture@keep.example:443#keep\ntrojan://fixture@one.example:443#restore-one\ntrojan://fixture@two.example:443#restore-two")
			}))
			defer upstream.Close()
			cfg := model.DefaultConfig()
			cfg.Service.RefreshOnRequest = false
			cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", Enabled: true, AllowLAN: true, URL: upstream.URL}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			published := publishNodesForDeletionTest(t, server, ref.ID)
			before := publishedDeletionRequest(server, published.SubscriptionURL)
			ids := deletionNodeIDsByName(t, server, ref.ID)
			deleteIDs := []string{ids["restore-one"], ids["restore-two"]}
			if mode == "all" {
				deleteIDs = append(deleteIDs, ids["keep"])
			}
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: deleteIDs}, http.StatusOK)
			restoreIDs := deleteIDs
			wantCount := 3
			if mode == "single" {
				restoreIDs = []string{ids["restore-one"]}
				wantCount = 2
			}
			if mode == "all" {
				if rec := publishedDeletionRequest(server, published.SubscriptionURL); rec.Code != http.StatusNotFound {
					t.Fatalf("deleted all nodes still returns %d", rec.Code)
				}
				var deleted deletedNodeListResponse
				data := nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/nodes/deleted", nil, http.StatusOK)
				if err := json.Unmarshal(data, &deleted); err != nil {
					t.Fatal(err)
				}
				restoreIDs = nil
				for _, node := range deleted.Nodes {
					restoreIDs = append(restoreIDs, node.ID)
				}
			}
			data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/enable", toggleNodesRequest{IDs: restoreIDs}, http.StatusOK)
			var restored struct {
				PublishedUpdated bool `json:"published_updated"`
				NodeCount        int  `json:"node_count"`
			}
			if err := json.Unmarshal(data, &restored); err != nil {
				t.Fatal(err)
			}
			if !restored.PublishedUpdated || restored.NodeCount != wantCount {
				t.Fatalf("restore did not synchronize publication: %s", data)
			}
			after := publishedDeletionRequest(server, published.SubscriptionURL)
			if after.Code != http.StatusOK || !strings.Contains(after.Body.String(), "restore-one") {
				t.Fatalf("restored node absent from original link: status=%d", after.Code)
			}
			if mode == "single" && strings.Contains(after.Body.String(), "restore-two") {
				t.Fatal("restoring one node restored an unrelated deletion")
			}
			if mode != "single" && !strings.Contains(after.Body.String(), "restore-two") {
				t.Fatal("bulk restoration lost a selected node")
			}
			if before.Header().Get("Subscription-Userinfo") != after.Header().Get("Subscription-Userinfo") {
				t.Fatal("restoration changed traffic metadata")
			}
			current, err := server.loadPublishedByID(published.PublishID)
			if err != nil || !strings.Contains(published.SubscriptionURL, current.Meta.Token) {
				t.Fatal("restoration changed subscription token")
			}
			if requests.Load() != 1 {
				t.Fatalf("restoration refetched upstream %d times", requests.Load())
			}
		})
	}
}

func TestRestoringDeletedNodeAfterEndpointRotationUpdatesPublicationWithoutRefresh(t *testing.T) {
	var revision, requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		fmt.Fprintf(w, "trojan://fixture-%d@node-%d.example:443#restore-me\ntrojan://fixture@keep.example:443#keep", revision.Load(), revision.Load())
	}))
	defer upstream.Close()
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", Enabled: true, AllowLAN: true, URL: upstream.URL}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	oldID := deletionNodeIDsByName(t, server, ref.ID)["restore-me"]
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{oldID}}, http.StatusOK)
	revision.Store(1)
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/refresh", nil, http.StatusOK)
	var deleted deletedNodeListResponse
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/nodes/deleted", nil, http.StatusOK)
	if err := json.Unmarshal(data, &deleted); err != nil {
		t.Fatal(err)
	}
	if len(deleted.Nodes) != 1 || deleted.Nodes[0].ID == oldID {
		t.Fatal("fixture did not rotate the deleted node identity")
	}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/enable", toggleNodesRequest{IDs: []string{deleted.Nodes[0].ID}}, http.StatusOK)
	rec := publishedDeletionRequest(server, published.SubscriptionURL)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "node-1.example") {
		t.Fatal("current restored endpoint was not published immediately")
	}
	if requests.Load() != 2 {
		t.Fatalf("restoration fetched the network: %d requests", requests.Load())
	}
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	state, err := pipeline.LoadNodeState(workspaceCfg)
	if err != nil || len(state.DeletedNodes) != 0 || len(state.DeletedNodeSignatures) != 0 {
		t.Fatal("restoration left stale deletion records")
	}
}

func TestRestoringNodeFailurePreservesDeletionAndPublication(t *testing.T) {
	for _, failure := range []string{"validation", "missing_cache"} {
		t.Run(failure, func(t *testing.T) {
			var requests atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				fmt.Fprintln(w, "trojan://fixture@restore.example:443#restore\ntrojan://fixture@keep.example:443#keep")
			}))
			defer upstream.Close()
			cfg := model.DefaultConfig()
			cfg.Service.RefreshOnRequest = false
			cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", Enabled: true, AllowLAN: true, URL: upstream.URL}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			published := publishNodesForDeletionTest(t, server, ref.ID)
			ids := deletionNodeIDsByName(t, server, ref.ID)
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{ids["restore"]}}, http.StatusOK)
			current, err := server.loadPublishedByID(published.PublishID)
			if err != nil {
				t.Fatal(err)
			}
			beforeState, err := os.ReadFile(ref.StatePath)
			if err != nil {
				t.Fatal(err)
			}
			beforeYAML, err := os.ReadFile(current.CurrentPath)
			if err != nil {
				t.Fatal(err)
			}
			if failure == "validation" {
				server.refreshAfterWrite = func(staged string) {
					if err := os.WriteFile(staged, []byte("proxies: []\nrules: [MATCH,DIRECT]\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				digest := sha256.Sum256([]byte(upstream.URL))
				key := hex.EncodeToString(digest[:])
				if err := filepath.WalkDir(server.cacheRootDir(), func(path string, entry os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if !entry.IsDir() && (entry.Name() == key+".body" || entry.Name() == key+".meta.json") {
						return os.Remove(path)
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			rec := restorationMutationRequest(t, server, ref.ID, []string{ids["restore"]})
			server.refreshAfterWrite = nil
			if rec.Code < 400 {
				t.Fatalf("failed restoration reported success: %s", rec.Body.String())
			}
			afterState, err := os.ReadFile(ref.StatePath)
			if err != nil || !bytes.Equal(beforeState, afterState) {
				t.Fatal("failed restoration removed deletion records")
			}
			afterYAML, err := os.ReadFile(current.CurrentPath)
			if err != nil || !bytes.Equal(beforeYAML, afterYAML) {
				t.Fatal("failed restoration changed subscription output")
			}
			if requests.Load() != 1 {
				t.Fatal("failed restoration fetched the network")
			}
		})
	}
}

func TestRestoringUnpublishedNodeDoesNotCreatePublication(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@restore.example:443#restore"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	id := deletionNodeIDsByName(t, server, ref.ID)["restore"]
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{id}}, http.StatusOK)
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/enable", toggleNodesRequest{IDs: []string{id}}, http.StatusOK)
	var response struct {
		PublishedUpdated bool `json:"published_updated"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	count, err := server.activePublicationCount()
	if err != nil || count != 0 || response.PublishedUpdated {
		t.Fatalf("restoring draft node created a publication: count=%d err=%v", count, err)
	}
}

func restorationMutationRequest(t *testing.T, server *Server, workspace string, ids []string) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(toggleNodesRequest{IDs: ids})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/nodes/enable", workspace), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(rec, req)
	return rec
}
