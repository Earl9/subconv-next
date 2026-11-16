package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
	"subconv-next/internal/renderer"
	"subconv-next/internal/storage"
)

// The request middleware already holds this workspace's lock. Rebuild an
// existing publication from cached sources before acknowledging a node change.
func (s *Server) saveNodeStateAndPublication(cfg model.Config, state model.NodeState, ref workspaceRef) (bool, int, error) {
	if strings.TrimSpace(ref.Meta.PublishID) == "" {
		return false, 0, pipeline.SaveNodeState(cfg, state)
	}
	published, err := s.loadPublishedByID(ref.Meta.PublishID)
	if errors.Is(err, errWorkspaceNotFound) || (err == nil && (published.Meta.Revoked || strings.TrimSpace(published.Meta.Token) == "")) {
		return false, 0, pipeline.SaveNodeState(cfg, state)
	}
	if err != nil {
		return false, 0, err
	}
	cfg.Service.OutputPath = published.CurrentPath
	finish, err := s.beginConfigRefresh(cfg, ref.Hash)
	if err != nil {
		return false, 0, err
	}
	defer finish()
	s.setRefreshStage("rendering")
	result, err := pipeline.RenderCachedConfig(cfg, state)
	empty := errors.Is(err, pipeline.ErrNoNodes)
	if err != nil && !empty {
		return false, 0, fmt.Errorf("rebuild published subscription: %w", err)
	}

	if !empty {
		stagedPath := filepath.Join(published.Dir, ".node-state.validate")
		defer os.Remove(stagedPath)
		if err := pipeline.WriteRendered(stagedPath, result.YAML); err != nil {
			return false, 0, err
		}
		if s.refreshAfterWrite != nil {
			s.refreshAfterWrite(stagedPath)
		}
		written, err := os.ReadFile(stagedPath)
		if err != nil {
			return false, 0, err
		}
		if err := pipeline.ValidateFinalConfig(written, pipeline.FinalNodeSet{Nodes: result.Nodes}, result.Audit, renderer.OptionsFromConfig(cfg)); err != nil {
			return false, 0, fmt.Errorf("validate published subscription: %w", err)
		}
		result.YAML = written
	}

	oldState, err := snapshotNodePublicationFile(cfg.Service.StatePath)
	if err != nil {
		return false, 0, err
	}
	oldYAML, err := snapshotNodePublicationFile(published.CurrentPath)
	if err != nil {
		return false, 0, err
	}
	oldWorkspaceMeta, err := snapshotNodePublicationFile(ref.MetaPath)
	if err != nil {
		return false, 0, err
	}
	oldPublished := published
	rollback := func(cause error, restoreYAML, restoreMeta bool) error {
		errs := []error{cause, oldState.restore()}
		if restoreYAML {
			errs = append(errs, oldYAML.restore())
		}
		if restoreMeta {
			// Preserve concurrent token rotations and access accounting.
			_, metaErr := s.updatePublishedMeta(published.ID, func(current *publishedRef) error {
				current.Meta.WorkspaceHash = oldPublished.Meta.WorkspaceHash
				current.Meta.UpdatedAt = oldPublished.Meta.UpdatedAt
				current.Meta.OutputFilename = oldPublished.Meta.OutputFilename
				current.Meta.SubscriptionInfo = oldPublished.Meta.SubscriptionInfo
				current.Meta.SourceUserinfo = oldPublished.Meta.SourceUserinfo
				return nil
			})
			errs = append(errs, metaErr, oldWorkspaceMeta.restore())
		}
		return errors.Join(errs...)
	}
	if err := pipeline.SaveNodeState(cfg, result.State); err != nil {
		return false, 0, err
	}
	if empty {
		// Retain the publication capability, but stop serving deleted nodes.
		if err := os.Remove(published.CurrentPath); err != nil && !os.IsNotExist(err) {
			return false, 0, rollback(err, false, false)
		}
	} else if err := pipeline.WriteRendered(published.CurrentPath, result.YAML); err != nil {
		return false, 0, rollback(err, false, false)
	}
	if err := s.finalizePublishedRefresh(&ref, &published, cfg, result); err != nil {
		return false, 0, rollback(err, true, true)
	}
	now := time.Now().UTC()
	s.setRefreshSuccess(result.NodeCount, nodeNames(result.Nodes), published.CurrentPath, now)
	updatedAt := now
	if empty {
		updatedAt = time.Time{}
		s.mu.Lock()
		s.status.YAMLExists = false
		s.status.YAMLUpdatedAt = updatedAt
		s.mu.Unlock()
	}
	s.setWorkspaceStatus(ref.Hash, model.RuntimeStatus{
		StartedAt:                firstNonZeroTime(ref.Meta.CreatedAt, now),
		Running:                  true,
		LastRefreshAt:            now,
		LastSuccessAt:            now,
		NextRefreshAt:            nextRefreshTime(now, cfg),
		RefreshInterval:          effectiveRefreshInterval(cfg),
		YAMLExists:               !empty,
		YAMLUpdatedAt:            updatedAt,
		UpstreamSourceCount:      enabledSubscriptionCount(cfg.Subscriptions),
		NodeCount:                result.NodeCount,
		NodeNames:                nodeNames(result.Nodes),
		EnabledSubscriptionCount: enabledSubscriptionCount(cfg.Subscriptions),
		OutputPath:               published.CurrentPath,
	})
	s.appendWorkspaceLog(ref.Hash, fmt.Sprintf("node state updated published subscription: %d nodes", result.NodeCount))
	return true, result.NodeCount, nil
}

type nodePublicationFile struct {
	path    string
	content []byte
	exists  bool
}

func snapshotNodePublicationFile(path string) (nodePublicationFile, error) {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nodePublicationFile{}, err
	}
	return nodePublicationFile{path: path, content: content, exists: err == nil}, nil
}

func (file nodePublicationFile) restore() error {
	if file.exists {
		return storage.AtomicWriteFile(file.path, file.content, 0o600)
	}
	if err := os.Remove(file.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeNodeDeletionError(w http.ResponseWriter, err error) {
	writeNodePublicationError(w, err, "NODE_DELETE_FAILED", "删除")
}

func writeNodePublicationError(w http.ResponseWriter, err error, code, action string) {
	if errors.Is(err, ErrRefreshInProgress) {
		writeAPIError(w, http.StatusConflict, "REFRESH_IN_PROGRESS", "订阅正在更新，请稍后重试"+action)
		return
	}
	if errors.Is(err, ErrRefreshCapacity) {
		w.Header().Set("Retry-After", "5")
		writeAPIError(w, http.StatusTooManyRequests, "REFRESH_CAPACITY_REACHED", "订阅更新繁忙，请稍后重试"+action)
		return
	}
	writeAPIError(w, http.StatusInternalServerError, code, err.Error())
}
