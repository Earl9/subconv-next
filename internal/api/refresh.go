package api

import (
	"errors"
	"os"
	"strings"
	"time"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
)

var ErrRefreshInProgress = errors.New("refresh is already running")

type refreshOutcome struct {
	Result pipeline.RenderResult
	At     time.Time
}

func (s *Server) startRefresh(reason string) (*refreshOutcome, error) {
	return s.startRefreshForConfig(s.snapshotConfig(), "", reason)
}

func (s *Server) startRefreshForConfig(cfg model.Config, workspaceHash, reason string) (*refreshOutcome, error) {
	s.refreshMu.Lock()
	if s.refreshRunning {
		done := s.refreshDone
		s.refreshMu.Unlock()
		_ = done
		return nil, ErrRefreshInProgress
	}
	s.refreshRunning = true
	s.refreshDone = make(chan struct{})
	s.refreshMu.Unlock()

	s.setRefreshing(true)
	defer s.finishRefresh()

	outcome, err := s.executeRefresh(cfg, workspaceHash, reason)
	if err != nil {
		return nil, err
	}
	return outcome, nil
}

func (s *Server) finishRefresh() {
	s.setRefreshing(false)
	s.refreshMu.Lock()
	done := s.refreshDone
	s.refreshRunning = false
	s.refreshDone = nil
	s.refreshMu.Unlock()
	if done != nil {
		close(done)
	}
}

func (s *Server) waitForRefresh(timeout time.Duration) bool {
	s.refreshMu.Lock()
	done := s.refreshDone
	running := s.refreshRunning
	s.refreshMu.Unlock()
	if !running || done == nil {
		return true
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-timer.C:
		return false
	}
}

func (s *Server) executeRefresh(cfg model.Config, workspaceHash, reason string) (*refreshOutcome, error) {
	if s.refreshBeforeRun != nil {
		s.refreshBeforeRun()
	}
	s.setRefreshStage("fetching")
	result, err := pipeline.RenderConfigWithProgress(cfg, func(stage string) {
		s.setRefreshStage(stage)
	})
	now := time.Now().UTC()
	if err != nil {
		message := err.Error()
		if errors.Is(err, pipeline.ErrNoNodes) {
			message = "no nodes available for rendering"
		}
		s.setRefreshFailure(message, now)
		s.appendLog(reason + " failed: " + message)
		s.appendWorkspaceLog(workspaceHash, reason+" failed: "+message)
		if s.refreshAfterRun != nil {
			s.refreshAfterRun()
		}
		return nil, err
	}
	s.setRefreshStage("writing")
	if err := pipeline.SaveNodeState(cfg, result.State); err != nil {
		s.setRefreshFailure(err.Error(), now)
		s.appendLog(reason + " state save failed: " + err.Error())
		s.appendWorkspaceLog(workspaceHash, reason+" state save failed: "+err.Error())
		if s.refreshAfterRun != nil {
			s.refreshAfterRun()
		}
		return nil, err
	}
	if err := pipeline.WriteRendered(result.OutputPath, result.YAML); err != nil {
		s.setRefreshFailure(err.Error(), now)
		s.appendLog(reason + " write failed: " + err.Error())
		s.appendWorkspaceLog(workspaceHash, reason+" write failed: "+err.Error())
		if s.refreshAfterRun != nil {
			s.refreshAfterRun()
		}
		return nil, err
	}

	s.setRefreshSuccess(result.NodeCount, nodeNames(result.Nodes), result.OutputPath, now)
	if strings.TrimSpace(workspaceHash) != "" {
		s.setWorkspaceStatus(workspaceHash, model.RuntimeStatus{
			StartedAt:                now,
			Running:                  true,
			LastRefreshAt:            now,
			LastSuccessAt:            now,
			NextRefreshAt:            nextRefreshTime(now, cfg),
			RefreshInterval:          effectiveRefreshInterval(cfg),
			YAMLExists:               true,
			YAMLUpdatedAt:            now,
			UpstreamSourceCount:      enabledSubscriptionCount(cfg.Subscriptions),
			NodeCount:                result.NodeCount,
			NodeNames:                nodeNames(result.Nodes),
			EnabledSubscriptionCount: enabledSubscriptionCount(cfg.Subscriptions),
			OutputPath:               result.OutputPath,
		})
	}
	for _, warning := range result.Warnings {
		s.appendLog(reason + " warning: " + warning)
		s.appendWorkspaceLog(workspaceHash, reason+" warning: "+warning)
	}
	for _, parseErr := range result.Errors {
		s.appendLog(reason + " parse error [" + parseErr.Kind + "]: " + parseErr.Message)
		s.appendWorkspaceLog(workspaceHash, reason+" parse error ["+parseErr.Kind+"]: "+parseErr.Message)
	}
	s.appendLog(reason + " succeeded: wrote " + result.OutputPath)
	s.appendWorkspaceLog(workspaceHash, reason+" succeeded: wrote "+result.OutputPath)
	if s.refreshAfterRun != nil {
		s.refreshAfterRun()
	}
	return &refreshOutcome{Result: result, At: now}, nil
}

func (s *Server) StartScheduler(stop <-chan struct{}) {
	go func() {
		cfg := s.snapshotConfig()
		if !yamlFileExists(cfg.Service.OutputPath) {
			_, _ = s.startRefresh("startup refresh")
		}
		for {
			cfg = s.snapshotConfig()
			interval := time.Duration(effectiveRefreshInterval(cfg)) * time.Second
			timer := time.NewTimer(interval)
			select {
			case <-stop:
				timer.Stop()
				return
			case <-timer.C:
				_, err := s.startRefresh("scheduled refresh")
				if err != nil && !errors.Is(err, ErrRefreshInProgress) {
					continue
				}
			}
		}
	}()

	go func() {
		for {
			cfg := s.snapshotConfig()
			interval := time.Duration(workspaceCleanupInterval(cfg)) * time.Second
			if interval <= 0 {
				interval = time.Hour
			}
			timer := time.NewTimer(interval)
			select {
			case <-stop:
				timer.Stop()
				return
			case <-timer.C:
				_ = s.cleanupExpiredWorkspaces()
				_ = s.cleanupStalePublished()
			}
		}
	}()
}

func effectiveRefreshInterval(cfg model.Config) int {
	if cfg.Service.RefreshInterval <= 0 {
		return model.DefaultRefreshInterval
	}
	return cfg.Service.RefreshInterval
}

func nextRefreshTime(now time.Time, cfg model.Config) time.Time {
	return now.UTC().Add(time.Duration(effectiveRefreshInterval(cfg)) * time.Second)
}

func yamlFileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func yamlFileUpdatedAt(path string) time.Time {
	if path == "" {
		return time.Time{}
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return time.Time{}
	}
	return info.ModTime().UTC()
}

func cacheExpired(path string, refreshIntervalSeconds int) bool {
	if !yamlFileExists(path) {
		return true
	}
	if refreshIntervalSeconds <= 0 {
		return false
	}
	updatedAt := yamlFileUpdatedAt(path)
	if updatedAt.IsZero() {
		return true
	}
	return time.Since(updatedAt) > time.Duration(refreshIntervalSeconds)*time.Second
}
