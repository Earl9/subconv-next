package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

func TestDeletingNodesImmediatelyUpdatesPublishedSubscriptionWithoutFetching(t *testing.T) {
	for _, mode := range []string{"single", "mixed_batch", "custom"} {
		t.Run(mode, func(t *testing.T) {
			var subscriptionRequests, ruleRequests atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rules" {
					ruleRequests.Add(1)
					fmt.Fprintln(w, "example.org")
					return
				}
				subscriptionRequests.Add(1)
				w.Header().Set("Subscription-Userinfo", "upload=10; download=20; total=1000; expire=1900000000")
				fmt.Fprintln(w, "trojan://fixture@keep.example:443#keep\ntrojan://fixture@one.example:443#drop-one\ntrojan://fixture@two.example:443#drop-two")
			}))
			defer upstream.Close()
			cfg := model.DefaultConfig()
			cfg.Service.AllowLAN, cfg.Service.RefreshOnRequest = true, false
			cfg.Service.AllowInsecurePublic = true
			cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", URL: upstream.URL + "/nodes", Enabled: true, AllowLAN: true}}
			cfg.Render.CustomRules = []model.CustomRule{{Key: "cached-rule", Label: "Cached rule", Enabled: true, SourceType: "http", TargetMode: "direct", Format: "text", Behavior: "domain", URL: upstream.URL + "/rules", Interval: 3600}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/custom", customNodeRequest{Content: "trojan://fixture@custom.example:443#custom-node"}, http.StatusOK)
			published := publishNodesForDeletionTest(t, server, ref.ID)
			before := publishedDeletionRequest(server, published.SubscriptionURL)
			if before.Code != http.StatusOK || before.Header().Get("Subscription-Userinfo") == "" {
				t.Fatalf("initial subscription missing metadata: status=%d", before.Code)
			}
			if subscriptionRequests.Load() != 1 || ruleRequests.Load() != 1 {
				t.Fatal("initial publication did not fetch both sources")
			}
			// Force the rule snapshot to expire; deletion must still be offline.
			err := filepath.WalkDir(server.cacheRootDir(), func(path string, entry os.DirEntry, err error) error {
				if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".meta.json") {
					return err
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				var meta fetcher.CacheMeta
				if err := json.Unmarshal(data, &meta); err != nil {
					return err
				}
				if meta.OriginalURL != upstream.URL+"/rules" {
					return nil
				}
				meta.FetchedAt = time.Now().Add(-24 * time.Hour)
				data, err = json.Marshal(meta)
				if err != nil {
					return err
				}
				return os.WriteFile(path, data, 0600)
			})
			if err != nil {
				t.Fatal(err)
			}
			ids := deletionNodeIDsByName(t, server, ref.ID)
			path, method := "/api/nodes/delete", http.MethodPost
			deletedNames := []string{"drop-one"}
			payload := interface{}(toggleNodesRequest{IDs: []string{ids["drop-one"]}})
			wantCount := 3
			if mode == "mixed_batch" {
				deletedNames = []string{"drop-one", "drop-two", "custom-node"}
				payload = toggleNodesRequest{IDs: []string{ids["drop-one"], ids["drop-two"], ids["custom-node"]}}
				wantCount = 1
			} else if mode == "custom" {
				path, method = "/api/nodes/custom/"+ids["custom-node"], http.MethodDelete
				deletedNames, payload = []string{"custom-node"}, nil
			}
			data := nodeDeletionRequest(t, server, ref.ID, method, path, payload, http.StatusOK)
			var deleted struct {
				PublishedUpdated bool `json:"published_updated"`
				NodeCount        int  `json:"node_count"`
			}
			if err := json.Unmarshal(data, &deleted); err != nil {
				t.Fatal(err)
			}
			if !deleted.PublishedUpdated || deleted.NodeCount != wantCount {
				t.Fatalf("deletion did not report synchronized publication: %s", data)
			}
			after := publishedDeletionRequest(server, published.SubscriptionURL)
			if after.Code != http.StatusOK {
				t.Fatalf("original link status=%d: %s", after.Code, after.Body.String())
			}
			for _, name := range deletedNames {
				if strings.Contains(after.Body.String(), name) {
					t.Fatalf("deleted %s remains in a proxy or group reference", name)
				}
			}
			if !strings.Contains(after.Body.String(), "keep") || !strings.Contains(after.Body.String(), "example.org") {
				t.Fatal("deletion lost remaining nodes or the cached custom rule")
			}
			if after.Header().Get("Subscription-Userinfo") != before.Header().Get("Subscription-Userinfo") {
				t.Fatal("deletion changed the subscription traffic header")
			}
			current, err := server.loadPublishedByID(published.PublishID)
			originalURL, _ := url.Parse(published.SubscriptionURL)
			if err != nil || !strings.Contains(originalURL.Path, current.Meta.Token) || current.Meta.WorkspaceHash != ref.Hash {
				t.Fatal("deletion changed the token or publication binding")
			}
			if subscriptionRequests.Load() != 1 || ruleRequests.Load() != 1 {
				t.Fatalf("deletion fetched the network: subscriptions=%d rules=%d", subscriptionRequests.Load(), ruleRequests.Load())
			}
		})
	}
}

func TestDeletingUnpublishedNodesDoesNotCreatePublication(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@one.example:443#drop\ntrojan://fixture@keep.example:443#keep"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	ids := deletionNodeIDsByName(t, server, ref.ID)
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{ids["drop"]}}, http.StatusOK)
	var deleted struct {
		PublishedUpdated bool `json:"published_updated"`
	}
	if err := json.Unmarshal(data, &deleted); err != nil {
		t.Fatal(err)
	}
	count, err := server.activePublicationCount()
	if err != nil || count != 0 || deleted.PublishedUpdated {
		t.Fatalf("deleting draft nodes created a publication: count=%d err=%v body=%s", count, err, data)
	}
}

func TestDeletingLastPublishedNodeInvalidatesOldYAML(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@one.example:443#last-node"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	ids := deletionNodeIDsByName(t, server, ref.ID)
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{ids["last-node"]}}, http.StatusOK)
	if rec := publishedDeletionRequest(server, published.SubscriptionURL); rec.Code != http.StatusNotFound {
		t.Fatalf("last deleted node is still downloadable: status=%d body=%s", rec.Code, rec.Body.String())
	}
	current, err := server.loadPublishedByID(published.PublishID)
	if err != nil || current.Meta.Token == "" || current.Meta.Revoked {
		t.Fatal("empty publication lost its reusable token")
	}
	var status statusResponse
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/status", nil, http.StatusOK)
	if err := json.Unmarshal(data, &status); err != nil {
		t.Fatal(err)
	}
	if status.YAMLExists || status.NodeCount != 0 {
		t.Fatalf("empty publication status is stale: %s", data)
	}
}

func TestDeletingNodeRollsBackWhenPublicationCannotBeWritten(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires normal directory permission enforcement")
	}
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@one.example:443#drop\ntrojan://fixture@keep.example:443#keep"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	current, err := server.loadPublishedByID(published.PublishID)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(current.CurrentPath)
	if err != nil {
		t.Fatal(err)
	}
	ids := deletionNodeIDsByName(t, server, ref.ID)
	if err := os.Chmod(current.Dir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(current.Dir, 0700) })
	rec := deletionMutationRequest(t, server, ref.ID, toggleNodesRequest{IDs: []string{ids["drop"]}})
	if rec.Code < 400 {
		t.Fatalf("unwritable publication reported deletion success: %s", rec.Body.String())
	}
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	state, err := pipeline.LoadNodeState(workspaceCfg)
	if err != nil || len(state.DeletedNodes) != 0 {
		t.Fatalf("failed deletion persisted: %#v err=%v", state.DeletedNodes, err)
	}
	after, err := os.ReadFile(current.CurrentPath)
	if err != nil || !bytes.Equal(original, after) {
		t.Fatal("failed deletion changed the existing published YAML")
	}
}

func TestDeletingNodeRejectsMissingSourceCacheWithoutPartialPublication(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/one" {
			fmt.Fprintln(w, "trojan://fixture@one.example:443#drop\ntrojan://fixture@keep-one.example:443#keep-one")
			return
		}
		fmt.Fprintln(w, "trojan://fixture@keep-two.example:443#keep-two")
	}))
	defer upstream.Close()
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "one", Name: "one", Enabled: true, URL: upstream.URL + "/one", AllowLAN: true}, {ID: "two", Name: "two", Enabled: true, URL: upstream.URL + "/two", AllowLAN: true}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	current, err := server.loadPublishedByID(published.PublishID)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(current.CurrentPath)
	if err != nil {
		t.Fatal(err)
	}
	ids := deletionNodeIDsByName(t, server, ref.ID)
	digest := sha256.Sum256([]byte(upstream.URL + "/two"))
	cacheKey := hex.EncodeToString(digest[:])
	if err := filepath.WalkDir(server.cacheRootDir(), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && (entry.Name() == cacheKey+".body" || entry.Name() == cacheKey+".meta.json") {
			return os.Remove(path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec := deletionMutationRequest(t, server, ref.ID, toggleNodesRequest{IDs: []string{ids["drop"]}})
	if rec.Code < 400 {
		t.Fatalf("missing cache allowed a partial publication: %s", rec.Body.String())
	}
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	state, err := pipeline.LoadNodeState(workspaceCfg)
	if err != nil || len(state.DeletedNodes) != 0 {
		t.Fatal("failed partial rebuild persisted deletion state")
	}
	after, err := os.ReadFile(current.CurrentPath)
	if err != nil || !bytes.Equal(original, after) || requests.Load() != 2 {
		t.Fatal("missing-cache deletion changed publication or fetched upstream")
	}
}

func TestDeletingNodePublicationFailureLeavesStateAndOutputUnchanged(t *testing.T) {
	for _, failure := range []string{"staged_validation", "metadata_finalize"} {
		t.Run(failure, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Service.RefreshOnRequest = false
			cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@drop.example:443#drop\ntrojan://fixture@keep.example:443#keep"}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			published := publishNodesForDeletionTest(t, server, ref.ID)
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
			ids := deletionNodeIDsByName(t, server, ref.ID)
			var restoreMeta func()
			server.refreshAfterWrite = func(staged string) {
				if failure == "staged_validation" {
					if err := os.WriteFile(staged, []byte("proxies: []\nrules: [MATCH,DIRECT]\n"), 0600); err != nil {
						t.Fatal(err)
					}
					return
				}
				// Fail the metadata commit after state and YAML can be written.
				backup := current.MetaPath + ".before-failure"
				if err := os.Rename(current.MetaPath, backup); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(current.MetaPath, 0700); err != nil {
					t.Fatal(err)
				}
				restoreMeta = func() {
					if err := os.Remove(current.MetaPath); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(backup, current.MetaPath); err != nil {
						t.Fatal(err)
					}
				}
			}
			rec := deletionMutationRequest(t, server, ref.ID, toggleNodesRequest{IDs: []string{ids["drop"]}})
			server.refreshAfterWrite = nil
			if restoreMeta != nil {
				restoreMeta()
			}
			if rec.Code < 400 {
				t.Fatalf("failed publication reported successful deletion: %s", rec.Body.String())
			}
			afterState, err := os.ReadFile(ref.StatePath)
			if err != nil || !bytes.Equal(beforeState, afterState) {
				t.Fatalf("%s failure changed node state: %v", failure, err)
			}
			afterYAML, err := os.ReadFile(current.CurrentPath)
			if err != nil || !bytes.Equal(beforeYAML, afterYAML) {
				t.Fatalf("%s failure changed published YAML: %v", failure, err)
			}
			// A failed publication must release the refresh guard for a retry.
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{ids["drop"]}}, http.StatusOK)
		})
	}
}

func TestDeletingNodeWaitsForConflictingPublicationRefresh(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@drop.example:443#drop\ntrojan://fixture@keep.example:443#keep"}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	current, err := server.loadPublishedByID(published.PublishID)
	if err != nil {
		t.Fatal(err)
	}
	workspaceCfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	workspaceCfg.Service.OutputPath = current.CurrentPath
	ids := deletionNodeIDsByName(t, server, ref.ID)
	beforeState, err := os.ReadFile(ref.StatePath)
	if err != nil {
		t.Fatal(err)
	}
	beforeYAML, err := os.ReadFile(current.CurrentPath)
	if err != nil {
		t.Fatal(err)
	}
	finish, err := server.beginConfigRefresh(workspaceCfg, "another-workspace")
	if err != nil {
		t.Fatal(err)
	}
	rec := deletionMutationRequest(t, server, ref.ID, toggleNodesRequest{IDs: []string{ids["drop"]}})
	finish()
	if rec.Code != http.StatusConflict {
		t.Fatalf("conflicting output refresh status=%d body=%s", rec.Code, rec.Body.String())
	}
	afterState, err := os.ReadFile(ref.StatePath)
	if err != nil || !bytes.Equal(beforeState, afterState) {
		t.Fatal("refresh conflict persisted deletion")
	}
	afterYAML, err := os.ReadFile(current.CurrentPath)
	if err != nil || !bytes.Equal(beforeYAML, afterYAML) {
		t.Fatal("refresh conflict changed published output")
	}
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/nodes/delete", toggleNodesRequest{IDs: []string{ids["drop"]}}, http.StatusOK)
}

func TestPublicationRefreshKeepsOutputLockedUntilMetadataIsFinalized(t *testing.T) {
	for _, mode := range []string{"manual", "scheduled"} {
		t.Run(mode, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Service.RefreshOnRequest = false
			cfg.Inline = []model.InlineConfig{{ID: "inline", Enabled: true, Content: "trojan://fixture@keep.example:443#keep"}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			published := publishNodesForDeletionTest(t, server, ref.ID)
			current, err := server.loadPublishedByID(published.PublishID)
			if err != nil {
				t.Fatal(err)
			}
			workspaceCfg, err := server.loadWorkspaceConfig(ref)
			if err != nil {
				t.Fatal(err)
			}
			workspaceCfg.Service.OutputPath = current.CurrentPath
			blocked := make(chan struct{})
			server.refreshAfterRun = func() {
				// executeRefresh has written YAML; block only the subsequent
				// metadata commit so a second workspace cannot overtake it.
				server.publishedMetaMu.Lock()
				close(blocked)
			}
			done := make(chan error, 1)
			go func() {
				if mode == "scheduled" {
					_, _, err := server.refreshPublishedWorkspace(ref, "scheduled test")
					done <- err
					return
				}
				rec := httptest.NewRecorder()
				server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, withWorkspace("/api/refresh", ref.ID), nil))
				if rec.Code != http.StatusOK {
					done <- fmt.Errorf("manual refresh status=%d: %s", rec.Code, rec.Body.String())
					return
				}
				done <- nil
			}()
			select {
			case <-blocked:
			case err := <-done:
				t.Fatalf("refresh did not reach metadata commit: %v", err)
			case <-time.After(2 * time.Second):
				t.Fatal("refresh did not reach metadata commit")
			}
			var unlockOnce sync.Once
			unblock := func() { unlockOnce.Do(server.publishedMetaMu.Unlock) }
			defer unblock()
			deadline := time.Now().Add(100 * time.Millisecond)
			var overtaken bool
			for time.Now().Before(deadline) {
				finish, err := server.beginConfigRefresh(workspaceCfg, "competing-workspace")
				if err == nil {
					finish()
					overtaken = true
					break
				}
				if !errors.Is(err, ErrRefreshInProgress) {
					t.Fatalf("unexpected conflicting refresh error: %v", err)
				}
				time.Sleep(time.Millisecond)
			}
			unblock()
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if overtaken {
				t.Fatal("another workspace acquired the output before publication metadata was finalized")
			}
			finish, err := server.beginConfigRefresh(workspaceCfg, "competing-workspace")
			if err != nil {
				t.Fatalf("completed publication retained output lock: %v", err)
			}
			finish()
		})
	}
}

func publishNodesForDeletionTest(t *testing.T, server *Server, workspace string) refreshResponse {
	t.Helper()
	var result refreshResponse
	data := nodeDeletionRequest(t, server, workspace, http.MethodPost, "/api/refresh", nil, http.StatusOK)
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func deletionNodeIDsByName(t *testing.T, server *Server, workspace string) map[string]string {
	t.Helper()
	var list nodeListResponse
	data := nodeDeletionRequest(t, server, workspace, http.MethodGet, "/api/nodes?all=1", nil, http.StatusOK)
	if err := json.Unmarshal(data, &list); err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, node := range list.Nodes {
		ids[node.Name] = node.ID
	}
	return ids
}

func publishedDeletionRequest(server *Server, rawURL string) *httptest.ResponseRecorder {
	parsed, _ := url.Parse(rawURL)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, parsed.RequestURI(), nil))
	return rec
}

func deletionMutationRequest(t *testing.T, server *Server, workspace string, payload toggleNodesRequest) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, withWorkspace("/api/nodes/delete", workspace), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	return rec
}
