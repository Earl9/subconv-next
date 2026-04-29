package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"subconv-next/internal/config"
	"subconv-next/internal/model"
	"subconv-next/internal/storage"
)

type workspaceMeta struct {
	ID                   string    `json:"id"`
	Hash                 string    `json:"hash"`
	CreatedAt            time.Time `json:"created_at"`
	LastAccessAt         time.Time `json:"last_access_at,omitempty"`
	PublishID            string    `json:"publish_id,omitempty"`
	LegacyExpiresAt      time.Time `json:"expires_at,omitempty"`
	LegacyPublishedToken string    `json:"published_token,omitempty"`
	LegacyPublishedAt    time.Time `json:"published_at,omitempty"`
}

type workspaceRef struct {
	ID         string
	Hash       string
	Dir        string
	ConfigPath string
	StatePath  string
	OutputPath string
	CacheDir   string
	MetaPath   string
	Meta       workspaceMeta
}

var (
	errWorkspaceRequired = errors.New("workspace is required")
	errWorkspaceNotFound = errors.New("workspace not found or expired")
)

func (s *Server) baseDataDir() string {
	return filepath.Dir(s.snapshotConfig().Service.StatePath)
}

func (s *Server) workspaceRootDir() string {
	return filepath.Join(s.baseDataDir(), "workspaces")
}

func (s *Server) cacheRootDir() string {
	return filepath.Join(s.baseDataDir(), "cache")
}

func (s *Server) logsDir() string {
	return filepath.Join(s.baseDataDir(), "logs")
}

func (s *Server) publishedRootDir() string {
	return filepath.Join(s.baseDataDir(), "published")
}

func (s *Server) buildWorkspaceRef(id string) workspaceRef {
	hash := sha256Hex(strings.TrimSpace(id))
	cacheDir := filepath.Join(s.cacheRootDir(), "workspaces", hash)
	dir := filepath.Join(s.workspaceRootDir(), hash)
	return workspaceRef{
		ID:         strings.TrimSpace(id),
		Hash:       hash,
		Dir:        dir,
		ConfigPath: filepath.Join(dir, "config.json"),
		StatePath:  filepath.Join(dir, "state.json"),
		OutputPath: filepath.Join(cacheDir, "preview.yaml"),
		CacheDir:   cacheDir,
		MetaPath:   filepath.Join(dir, "meta.json"),
	}
}

func (s *Server) buildWorkspaceRefByHash(hash string) workspaceRef {
	ref := s.buildWorkspaceRef(hash)
	ref.ID = ""
	ref.Hash = strings.TrimSpace(hash)
	ref.Dir = filepath.Join(s.workspaceRootDir(), ref.Hash)
	ref.ConfigPath = filepath.Join(ref.Dir, "config.json")
	ref.StatePath = filepath.Join(ref.Dir, "state.json")
	ref.MetaPath = filepath.Join(ref.Dir, "meta.json")
	ref.CacheDir = filepath.Join(s.cacheRootDir(), "workspaces", ref.Hash)
	ref.OutputPath = filepath.Join(ref.CacheDir, "preview.yaml")
	return ref
}

func (s *Server) createWorkspace() (workspaceRef, error) {
	if err := s.cleanupExpiredWorkspaces(); err != nil {
		return workspaceRef{}, err
	}
	id, err := randomWorkspaceID()
	if err != nil {
		return workspaceRef{}, err
	}
	ref := s.buildWorkspaceRef(id)
	now := time.Now().UTC()
	ref.Meta = workspaceMeta{
		ID:           ref.ID,
		Hash:         ref.Hash,
		CreatedAt:    now,
		LastAccessAt: now,
	}
	cfg := s.workspaceBaseConfig()
	cfg.Service.OutputPath = ref.OutputPath
	cfg.Service.StatePath = ref.StatePath
	cfg.Service.CacheDir = ref.CacheDir
	if err := os.MkdirAll(ref.Dir, 0o755); err != nil {
		return workspaceRef{}, fmt.Errorf("create workspace dir: %w", err)
	}
	if err := os.MkdirAll(ref.CacheDir, 0o755); err != nil {
		return workspaceRef{}, fmt.Errorf("create workspace cache dir: %w", err)
	}
	if err := config.WriteJSON(ref.ConfigPath, cfg); err != nil {
		return workspaceRef{}, fmt.Errorf("write workspace config: %w", err)
	}
	if err := s.saveWorkspaceMeta(ref); err != nil {
		return workspaceRef{}, err
	}
	if err := storage.AtomicWriteFile(ref.StatePath, []byte("{\n}\n"), 0o644); err != nil {
		return workspaceRef{}, fmt.Errorf("write workspace state: %w", err)
	}
	return ref, nil
}

func (s *Server) workspaceBaseConfig() model.Config {
	base := s.snapshotConfig()
	cfg := model.DefaultConfig()
	cfg.Service.LogLevel = base.Service.LogLevel
	cfg.Service.Template = base.Service.Template
	cfg.Service.RefreshInterval = base.Service.RefreshInterval
	cfg.Service.RefreshOnRequest = base.Service.RefreshOnRequest
	cfg.Service.StaleIfError = base.Service.StaleIfError
	cfg.Service.StrictMode = base.Service.StrictMode
	cfg.Service.MaxSubscriptionBytes = base.Service.MaxSubscriptionBytes
	cfg.Service.FetchTimeoutSeconds = base.Service.FetchTimeoutSeconds
	cfg.Service.AllowLAN = base.Service.AllowLAN
	cfg.Service.WorkspaceTTLSeconds = base.Service.WorkspaceTTLSeconds
	cfg.Service.WorkspaceCleanupIntervalSeconds = base.Service.WorkspaceCleanupIntervalSeconds
	cfg.Service.WorkspaceCleanupInterval = base.Service.WorkspaceCleanupInterval
	cfg.Service.PublishedDeleteIfNotAccessedDays = base.Service.PublishedDeleteIfNotAccessedDays
	return cfg
}

func (s *Server) loadWorkspace(id string) (workspaceRef, error) {
	ref, err := s.loadWorkspaceNoTouch(id)
	if err != nil {
		return workspaceRef{}, err
	}
	if err := s.touchWorkspace(&ref); err != nil {
		return workspaceRef{}, err
	}
	return ref, nil
}

func (s *Server) loadWorkspaceNoTouch(id string) (workspaceRef, error) {
	if strings.TrimSpace(id) == "" {
		return workspaceRef{}, errWorkspaceRequired
	}
	ref := s.buildWorkspaceRef(id)
	data, err := os.ReadFile(ref.MetaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return workspaceRef{}, errWorkspaceNotFound
		}
		return workspaceRef{}, fmt.Errorf("read workspace meta: %w", err)
	}
	if err := json.Unmarshal(data, &ref.Meta); err != nil {
		return workspaceRef{}, fmt.Errorf("decode workspace meta: %w", err)
	}
	if strings.TrimSpace(ref.Meta.ID) == "" {
		ref.Meta.ID = ref.ID
	}
	if strings.TrimSpace(ref.Meta.Hash) == "" {
		ref.Meta.Hash = ref.Hash
	}
	if ref.Meta.LastAccessAt.IsZero() {
		ref.Meta.LastAccessAt = legacyWorkspaceLastAccess(ref.Meta, s.snapshotConfig().Service.WorkspaceTTLSeconds)
	}
	if s.workspaceExpired(ref.Meta) {
		_ = s.removeWorkspace(ref)
		return workspaceRef{}, errWorkspaceNotFound
	}
	return ref, nil
}

func (s *Server) loadWorkspaceByHash(hash string) (workspaceRef, error) {
	ref := s.buildWorkspaceRefByHash(hash)
	data, err := os.ReadFile(ref.MetaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return workspaceRef{}, errWorkspaceNotFound
		}
		return workspaceRef{}, fmt.Errorf("read workspace meta: %w", err)
	}
	if err := json.Unmarshal(data, &ref.Meta); err != nil {
		return workspaceRef{}, fmt.Errorf("decode workspace meta: %w", err)
	}
	ref.ID = firstNonEmptyString(ref.Meta.ID, ref.ID)
	ref.Hash = firstNonEmptyString(ref.Meta.Hash, hash)
	ref.Dir = filepath.Join(s.workspaceRootDir(), ref.Hash)
	ref.ConfigPath = filepath.Join(ref.Dir, "config.json")
	ref.StatePath = filepath.Join(ref.Dir, "state.json")
	ref.MetaPath = filepath.Join(ref.Dir, "meta.json")
	ref.CacheDir = filepath.Join(s.cacheRootDir(), "workspaces", ref.Hash)
	ref.OutputPath = filepath.Join(ref.CacheDir, "preview.yaml")
	if ref.Meta.LastAccessAt.IsZero() {
		ref.Meta.LastAccessAt = legacyWorkspaceLastAccess(ref.Meta, s.snapshotConfig().Service.WorkspaceTTLSeconds)
	}
	return ref, nil
}

func (s *Server) touchWorkspace(ref *workspaceRef) error {
	if ref == nil {
		return nil
	}
	ref.Meta.LastAccessAt = time.Now().UTC()
	return s.saveWorkspaceMeta(*ref)
}

func (s *Server) workspaceExpiresAt(meta workspaceMeta) time.Time {
	ttl := s.snapshotConfig().Service.WorkspaceTTLSeconds
	if ttl <= 0 {
		return time.Time{}
	}
	lastAccessAt := meta.LastAccessAt
	if lastAccessAt.IsZero() {
		lastAccessAt = legacyWorkspaceLastAccess(meta, ttl)
	}
	return lastAccessAt.UTC().Add(time.Duration(ttl) * time.Second)
}

func (s *Server) workspaceExpired(meta workspaceMeta) bool {
	expiresAt := s.workspaceExpiresAt(meta)
	if expiresAt.IsZero() {
		return false
	}
	return !expiresAt.After(time.Now().UTC())
}

func legacyWorkspaceLastAccess(meta workspaceMeta, ttlSeconds int) time.Time {
	if !meta.LastAccessAt.IsZero() {
		return meta.LastAccessAt.UTC()
	}
	if !meta.LegacyExpiresAt.IsZero() && ttlSeconds > 0 {
		last := meta.LegacyExpiresAt.UTC().Add(-time.Duration(ttlSeconds) * time.Second)
		if last.After(time.Time{}) {
			return last
		}
	}
	if !meta.CreatedAt.IsZero() {
		return meta.CreatedAt.UTC()
	}
	return time.Now().UTC()
}

func (s *Server) saveWorkspaceMeta(ref workspaceRef) error {
	ref.Meta.ID = firstNonEmptyString(ref.Meta.ID, ref.ID)
	ref.Meta.Hash = firstNonEmptyString(ref.Meta.Hash, ref.Hash)
	if ref.Meta.CreatedAt.IsZero() {
		ref.Meta.CreatedAt = time.Now().UTC()
	}
	if ref.Meta.LastAccessAt.IsZero() {
		ref.Meta.LastAccessAt = ref.Meta.CreatedAt
	}
	data, err := json.MarshalIndent(ref.Meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workspace meta: %w", err)
	}
	data = append(data, '\n')
	if err := storage.AtomicWriteFile(ref.MetaPath, data, 0o644); err != nil {
		return fmt.Errorf("write workspace meta: %w", err)
	}
	return nil
}

func (s *Server) deleteWorkspace(id string) error {
	ref, err := s.loadWorkspaceNoTouch(id)
	if err != nil {
		return err
	}
	return s.removeWorkspace(ref)
}

func (s *Server) removeWorkspace(ref workspaceRef) error {
	_ = os.Remove(ref.OutputPath)
	_ = os.RemoveAll(ref.CacheDir)
	if err := os.RemoveAll(ref.Dir); err != nil {
		return err
	}
	s.forgetWorkspace(ref.Hash)
	return nil
}

func (s *Server) forgetWorkspace(workspaceHash string) {
	if strings.TrimSpace(workspaceHash) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.workspaceStatus, workspaceHash)
	delete(s.workspaceLogs, workspaceHash)
}

func (s *Server) cleanupExpiredWorkspaces() error {
	root := s.workspaceRootDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ref, err := s.loadWorkspaceByHash(entry.Name())
		if err != nil {
			continue
		}
		if s.workspaceExpired(ref.Meta) {
			_ = s.removeWorkspace(ref)
		}
	}
	return nil
}

func randomWorkspaceID() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate workspace id: %w", err)
	}
	return "w_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func workspaceIDFromRequest(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("workspace"))
}

func (s *Server) requireWorkspace(r *http.Request) (workspaceRef, error) {
	return s.loadWorkspace(workspaceIDFromRequest(r))
}

func (s *Server) loadWorkspaceConfig(ref workspaceRef) (model.Config, error) {
	cfg, err := config.Load(ref.ConfigPath)
	if err != nil {
		return model.Config{}, err
	}
	cfg.Service.StatePath = ref.StatePath
	cfg.Service.CacheDir = ref.CacheDir
	if strings.TrimSpace(ref.Meta.PublishID) != "" {
		if _, err := s.loadPublishedByID(ref.Meta.PublishID); err == nil {
			cfg.Service.OutputPath = s.buildPublishedRef(ref.Meta.PublishID).CurrentPath
			return cfg, nil
		}
	}
	cfg.Service.OutputPath = ref.OutputPath
	return cfg, nil
}

func workspaceCleanupInterval(cfg model.Config) int {
	if cfg.Service.WorkspaceCleanupIntervalSeconds > 0 {
		return cfg.Service.WorkspaceCleanupIntervalSeconds
	}
	if cfg.Service.WorkspaceCleanupInterval > 0 {
		return cfg.Service.WorkspaceCleanupInterval
	}
	return 3600
}
