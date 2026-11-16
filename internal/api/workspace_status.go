package api

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"subconv-next/internal/model"
)

// Recover publication details from this workspace's current output after a
// restart. Reading status must neither fetch upstream sources nor inherit the
// server-wide status of whichever workspace happened to refresh most recently.
func workspaceStatusFromOutput(cfg model.Config, startedAt, now time.Time) model.RuntimeStatus {
	status := model.RuntimeStatus{
		StartedAt:                startedAt,
		Running:                  true,
		RefreshInterval:          effectiveRefreshInterval(cfg),
		NextRefreshAt:            nextRefreshTime(now, cfg),
		UpstreamSourceCount:      enabledSubscriptionCount(cfg.Subscriptions),
		EnabledSubscriptionCount: enabledSubscriptionCount(cfg.Subscriptions),
		OutputPath:               cfg.Service.OutputPath,
	}
	file, err := os.Open(cfg.Service.OutputPath)
	if err != nil {
		return status
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return status
	}
	status.YAMLExists = true
	status.YAMLUpdatedAt = info.ModTime().UTC()
	var output struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
	}
	if err := yaml.NewDecoder(file).Decode(&output); err != nil {
		return status
	}
	status.NodeCount = len(output.Proxies)
	for _, proxy := range output.Proxies {
		status.NodeNames = append(status.NodeNames, proxy.Name)
	}
	if status.NodeCount > 0 {
		status.LastSuccessAt = status.YAMLUpdatedAt
	}
	return status
}

func (s *Server) setConfigRefreshFailure(cfg model.Config, workspaceHash, message string, at time.Time) {
	s.setRefreshFailure(message, at)
	if strings.TrimSpace(workspaceHash) == "" {
		return
	}
	previous := s.snapshotWorkspaceStatus(workspaceHash)
	status := workspaceStatusFromOutput(cfg, firstNonZeroTime(previous.StartedAt, at), at)
	if !previous.LastSuccessAt.IsZero() {
		status.LastSuccessAt = previous.LastSuccessAt
	}
	status.LastRefreshAt = at
	status.LastError = message
	s.setWorkspaceStatus(workspaceHash, status)
}
