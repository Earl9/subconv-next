package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"subconv-next/internal/model"
)

func TestWorkspaceStatusRecoversPublishedNodesAfterRestartWithoutFetching(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		fmt.Fprintln(w, "trojan://fixture@one.example:443#one\ntrojan://fixture@two.example:443#two")
	}))
	defer upstream.Close()
	cfg := model.DefaultConfig()
	cfg.Service.AllowLAN, cfg.Service.AllowInsecurePublic, cfg.Service.RefreshOnRequest = true, true, false
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", URL: upstream.URL, Enabled: true, AllowLAN: true}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	published := publishNodesForDeletionTest(t, server, ref.ID)
	wantNames := workspaceStatusForTest(t, server, ref.ID).NodeNames
	pub, err := server.loadPublishedByID(published.PublishID)
	if err != nil {
		t.Fatal(err)
	}
	updatedAt := yamlFileUpdatedAt(pub.CurrentPath)
	beforeRequests := requests.Load()
	restarted := NewServer("restarted", cfg)
	status := workspaceStatusForTest(t, restarted, ref.ID)
	if !status.YAMLExists || status.NodeCount != 2 || !reflect.DeepEqual(status.NodeNames, wantNames) {
		t.Fatalf("status lost existing published nodes after restart: %+v", status)
	}
	if status.YAMLUpdatedAt != formatTime(updatedAt) || status.LastSuccessAt != formatTime(updatedAt) || status.LastError != "" {
		t.Fatalf("status lost persisted output time or invented a failure: %+v", status)
	}
	if requests.Load() != beforeRequests {
		t.Fatal("reading status fetched upstream subscriptions")
	}
}

func TestWorkspaceFailedRefreshKeepsLastPublicationAndDoesNotLeakStatus(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(fmt.Sprintf("restart=%t", restart), func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Service.RefreshOnRequest = false
			cfg.Inline = []model.InlineConfig{{ID: "source", Enabled: true, Content: "trojan://fixture@keep.example:443#keep\ntrojan://fixture@second.example:443#second"}}
			server, cfg := newTestServer(t, cfg)
			ref := createWorkspaceRefForTest(t, server, cfg)
			published := publishNodesForDeletionTest(t, server, ref.ID)
			wantNames := workspaceStatusForTest(t, server, ref.ID).NodeNames
			pub, err := server.loadPublishedByID(published.PublishID)
			if err != nil {
				t.Fatal(err)
			}
			oldOutputTime := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
			if err := os.Chtimes(pub.CurrentPath, oldOutputTime, oldOutputTime); err != nil {
				t.Fatal(err)
			}
			beforeYAML, err := os.ReadFile(pub.CurrentPath)
			if err != nil {
				t.Fatal(err)
			}
			if restart {
				server = NewServer("restarted", cfg)
			}
			before := workspaceStatusForTest(t, server, ref.ID)
			otherConfig := cfg
			otherConfig.Inline = []model.InlineConfig{{ID: "other", Enabled: true, Content: "trojan://fixture@other.example:443#other"}}
			other := createWorkspaceRefForTest(t, server, otherConfig)
			publishNodesForDeletionTest(t, server, other.ID)
			otherBefore := workspaceStatusForTest(t, server, other.ID)
			writeInvalidWorkspaceRuleForStatusTest(t, server, ref)
			nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/refresh", nil, http.StatusBadRequest)
			after := workspaceStatusForTest(t, server, ref.ID)
			if !after.YAMLExists || after.NodeCount != 2 || !reflect.DeepEqual(after.NodeNames, wantNames) {
				t.Fatalf("failed refresh lost last downloadable nodes: %+v", after)
			}
			if after.YAMLUpdatedAt != formatTime(oldOutputTime) || after.LastSuccessAt != before.LastSuccessAt {
				t.Fatalf("failed refresh replaced last successful timestamps: before=%+v after=%+v", before, after)
			}
			if !strings.Contains(after.LastError, "custom_rules[0].payload") || !strings.Contains(after.LastError, "第 1 行") || after.LastRefreshAt == "" || after.Refreshing {
				t.Fatalf("failed refresh was not reported in workspace status: %+v", after)
			}
			if _, err := time.Parse(time.RFC3339, after.LastRefreshAt); err != nil {
				t.Fatalf("invalid last refresh time: %v", err)
			}
			if otherAfter := workspaceStatusForTest(t, server, other.ID); !reflect.DeepEqual(otherBefore, otherAfter) {
				t.Fatalf("another workspace inherited failed refresh status: before=%+v after=%+v", otherBefore, otherAfter)
			}
			download := publishedDeletionRequest(server, published.SubscriptionURL)
			if download.Code != http.StatusOK || !bytes.Equal(download.Body.Bytes(), beforeYAML) {
				t.Fatal("failed refresh changed the last downloadable publication")
			}
		})
	}
}

func TestUnpublishedWorkspaceStatusRemainsEmptyAfterRestartAndFailedRefresh(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.RefreshOnRequest = false
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	server = NewServer("restarted", cfg)
	otherConfig := cfg
	otherConfig.Inline = []model.InlineConfig{{ID: "other", Enabled: true, Content: "trojan://fixture@other.example:443#other"}}
	other := createWorkspaceRefForTest(t, server, otherConfig)
	publishNodesForDeletionTest(t, server, other.ID)
	assertEmpty := func(status statusResponse) {
		t.Helper()
		if status.YAMLExists || status.NodeCount != 0 || len(status.NodeNames) != 0 || status.YAMLUpdatedAt != "" || status.LastSuccessAt != "" {
			t.Fatalf("unpublished workspace inherited output from another workspace: %+v", status)
		}
	}
	assertEmpty(workspaceStatusForTest(t, server, ref.ID))
	nodeDeletionRequest(t, server, ref.ID, http.MethodPost, "/api/refresh", nil, http.StatusBadRequest)
	status := workspaceStatusForTest(t, server, ref.ID)
	assertEmpty(status)
	if status.LastError == "" || status.LastRefreshAt == "" || status.Refreshing {
		t.Fatalf("unpublished workspace lost failed refresh status: %+v", status)
	}
}

func workspaceStatusForTest(t *testing.T, server *Server, workspace string) statusResponse {
	t.Helper()
	data := nodeDeletionRequest(t, server, workspace, http.MethodGet, "/api/status", nil, http.StatusOK)
	var status statusResponse
	if err := json.Unmarshal(data, &status); err != nil {
		t.Fatal(err)
	}
	return status
}

func writeInvalidWorkspaceRuleForStatusTest(t *testing.T, server *Server, ref workspaceRef) {
	t.Helper()
	cfg, err := server.loadWorkspaceConfig(ref)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Render.CustomRules = []model.CustomRule{{Key: "legacy", Label: "Legacy", Enabled: true, SourceType: "inline", Behavior: "classical", Format: "text", Payload: []string{"DOMAIN-SUFFIX,"}}}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ref.ConfigPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
