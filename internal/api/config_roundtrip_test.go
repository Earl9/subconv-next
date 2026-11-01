package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"subconv-next/internal/model"
)

func TestConfigRoundTripPreservesSubscriptionToken(t *testing.T) {
	cfg := model.DefaultConfig()
	originalURL := "https://example.com/sub?token=secret-token&format=clash"
	cfg.Subscriptions = []model.SubscriptionConfig{{ID: "source", Name: "source", Enabled: true, URL: originalURL}}
	server, cfg := newTestServer(t, cfg)
	ref := createWorkspaceRefForTest(t, server, cfg)
	var response configResponse
	data := nodeDeletionRequest(t, server, ref.ID, http.MethodGet, "/api/config", nil, http.StatusOK)
	if strings.Contains(string(data), "secret-token") {
		t.Fatal("GET leaked the subscription token")
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	response.Config.Subscriptions[0].Name = "renamed source"
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusOK)
	stored, err := server.loadWorkspaceConfig(ref)
	if err != nil || stored.Subscriptions[0].URL != originalURL || stored.Subscriptions[0].Name != "renamed source" {
		t.Fatal("saving redacted config lost the original URL or the name edit")
	}
	// A masked URL sent to a new host must not receive the old credential.
	response.Config.Subscriptions[0].URL = strings.Replace(response.Config.Subscriptions[0].URL, "example.com", "other.example", 1)
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusBadRequest)
	response.Config.Subscriptions[0].URL = "https://other.example/sub?token=new-token"
	nodeDeletionRequest(t, server, ref.ID, http.MethodPut, "/api/config", response.Config, http.StatusOK)
	stored, err = server.loadWorkspaceConfig(ref)
	if err != nil || stored.Subscriptions[0].URL != response.Config.Subscriptions[0].URL {
		t.Fatal("an explicit replacement URL was not saved")
	}
}
