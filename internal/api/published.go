package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"subconv-next/internal/model"
	"subconv-next/internal/pipeline"
	"subconv-next/internal/storage"
)

type publishedMeta struct {
	PublishID        string                         `json:"publish_id"`
	Token            string                         `json:"token,omitempty"`
	TokenHash        string                         `json:"token_hash"`
	TokenHint        string                         `json:"token_hint,omitempty"`
	CreatedAt        time.Time                      `json:"created_at"`
	UpdatedAt        time.Time                      `json:"updated_at"`
	LastAccessAt     time.Time                      `json:"last_access_at,omitempty"`
	AccessCount      int                            `json:"access_count"`
	Revoked          bool                           `json:"revoked"`
	WorkspaceHash    string                         `json:"workspace_hash,omitempty"`
	RotatedAt        time.Time                      `json:"rotated_at,omitempty"`
	SubscriptionInfo *publishedSubscriptionUserinfo `json:"subscription_userinfo,omitempty"`
	SourceUserinfo   []publishedSourceUserinfo      `json:"source_userinfo,omitempty"`
}

type publishedSubscriptionUserinfo struct {
	Upload        int64     `json:"upload"`
	Download      int64     `json:"download"`
	Total         int64     `json:"total"`
	Expire        int64     `json:"expire,omitempty"`
	Sources       int       `json:"sources"`
	UpdatedAt     time.Time `json:"updated_at"`
	HeaderEnabled bool      `json:"header_enabled,omitempty"`
}

type publishedSourceUserinfo struct {
	SourceID      string `json:"source_id,omitempty"`
	SourceName    string `json:"source_name,omitempty"`
	SourceURLHost string `json:"source_url_host,omitempty"`
	Upload        int64  `json:"upload,omitempty"`
	Download      int64  `json:"download,omitempty"`
	Total         int64  `json:"total,omitempty"`
	Expire        int64  `json:"expire,omitempty"`
	Available     bool   `json:"available"`
	FromHeader    bool   `json:"from_header,omitempty"`
	FromInfoNode  bool   `json:"from_info_node,omitempty"`
	FetchedAt     string `json:"fetched_at,omitempty"`
}

type legacyPublishedMeta struct {
	TokenHash     string    `json:"token_hash"`
	WorkspaceHash string    `json:"workspace_hash,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
}

type publishedRef struct {
	ID          string
	Dir         string
	CurrentPath string
	MetaPath    string
	Meta        publishedMeta
}

func (s *Server) buildPublishedRef(id string) publishedRef {
	id = strings.TrimSpace(id)
	dir := filepath.Join(s.publishedRootDir(), id)
	return publishedRef{
		ID:          id,
		Dir:         dir,
		CurrentPath: filepath.Join(dir, "current.yaml"),
		MetaPath:    filepath.Join(dir, "meta.json"),
	}
}

func (s *Server) createPublished(workspaceHash string) (publishedRef, error) {
	token, err := randomSubscriptionToken()
	if err != nil {
		return publishedRef{}, err
	}
	return s.createPublishedWithToken(workspaceHash, "", token)
}

func (s *Server) createPublishedWithToken(workspaceHash, publishID, token string) (publishedRef, error) {
	if strings.TrimSpace(publishID) == "" {
		var err error
		publishID, err = randomPublishID()
		if err != nil {
			return publishedRef{}, err
		}
	}
	ref := s.buildPublishedRef(publishID)
	now := time.Now().UTC()
	ref.Meta = publishedMeta{
		PublishID:     ref.ID,
		Token:         strings.TrimSpace(token),
		TokenHash:     sha256Hex(token),
		TokenHint:     publishedTokenHint(token),
		CreatedAt:     now,
		UpdatedAt:     now,
		WorkspaceHash: strings.TrimSpace(workspaceHash),
	}
	if err := os.MkdirAll(ref.Dir, 0o755); err != nil {
		return publishedRef{}, fmt.Errorf("create published dir: %w", err)
	}
	if err := s.savePublishedMeta(ref); err != nil {
		return publishedRef{}, err
	}
	return ref, nil
}

func (s *Server) loadPublishedByID(id string) (publishedRef, error) {
	ref := s.buildPublishedRef(id)
	data, err := os.ReadFile(ref.MetaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return publishedRef{}, errWorkspaceNotFound
		}
		return publishedRef{}, fmt.Errorf("read published meta: %w", err)
	}
	if err := json.Unmarshal(data, &ref.Meta); err != nil {
		return publishedRef{}, fmt.Errorf("decode published meta: %w", err)
	}
	ref.ID = firstNonEmptyString(ref.Meta.PublishID, ref.ID)
	ref.Dir = filepath.Join(s.publishedRootDir(), ref.ID)
	ref.CurrentPath = filepath.Join(ref.Dir, "current.yaml")
	ref.MetaPath = filepath.Join(ref.Dir, "meta.json")
	return ref, nil
}

func (s *Server) savePublishedMeta(ref publishedRef) error {
	ref.Meta.PublishID = firstNonEmptyString(ref.Meta.PublishID, ref.ID)
	ref.Meta.Token = strings.TrimSpace(ref.Meta.Token)
	if ref.Meta.Token != "" {
		ref.Meta.TokenHash = sha256Hex(ref.Meta.Token)
	}
	ref.Meta.TokenHint = firstNonEmptyString(ref.Meta.TokenHint, publishedTokenHint(ref.Meta.Token))
	if ref.Meta.CreatedAt.IsZero() {
		ref.Meta.CreatedAt = time.Now().UTC()
	}
	if ref.Meta.UpdatedAt.IsZero() {
		ref.Meta.UpdatedAt = ref.Meta.CreatedAt
	}
	data, err := json.MarshalIndent(ref.Meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal published meta: %w", err)
	}
	data = append(data, '\n')
	if err := storage.AtomicWriteFile(ref.MetaPath, data, 0o644); err != nil {
		return fmt.Errorf("write published meta: %w", err)
	}
	return nil
}

func (s *Server) ensureWorkspacePublishedRef(ref *workspaceRef) (publishedRef, bool, error) {
	if ref == nil {
		return publishedRef{}, false, fmt.Errorf("workspace ref is nil")
	}
	if strings.TrimSpace(ref.Meta.PublishID) != "" {
		published, err := s.loadPublishedByID(ref.Meta.PublishID)
		if err == nil && publishedRestorable(published) {
			return published, false, nil
		}
		ref.Meta.PublishID = ""
		ref.Meta.LegacyPublishedToken = ""
		ref.Meta.LegacyPublishedAt = time.Time{}
		_ = s.saveWorkspaceMeta(*ref)
	}
	if strings.TrimSpace(ref.Meta.LegacyPublishedToken) != "" {
		published, err := s.migrateWorkspaceLegacyPublished(ref, ref.Meta.LegacyPublishedToken)
		if err == nil {
			return published, false, nil
		}
	}
	published, err := s.createPublished(ref.Hash)
	if err != nil {
		return publishedRef{}, false, err
	}
	ref.Meta.PublishID = published.ID
	ref.Meta.LegacyPublishedToken = ""
	ref.Meta.LegacyPublishedAt = time.Time{}
	if err := s.saveWorkspaceMeta(*ref); err != nil {
		_ = os.RemoveAll(published.Dir)
		return publishedRef{}, false, err
	}
	return published, true, nil
}

func (s *Server) migrateWorkspaceLegacyPublished(ref *workspaceRef, token string) (publishedRef, error) {
	published, err := s.migrateLegacyPublishedToken(token, ref.Hash)
	if err != nil {
		published, err = s.createPublishedWithToken(ref.Hash, "", token)
		if err != nil {
			return publishedRef{}, err
		}
	}
	ref.Meta.PublishID = published.ID
	ref.Meta.LegacyPublishedToken = ""
	ref.Meta.LegacyPublishedAt = time.Time{}
	if err := s.saveWorkspaceMeta(*ref); err != nil {
		return publishedRef{}, err
	}
	return published, nil
}

func (s *Server) migrateLegacyPublishedToken(token, workspaceHash string) (publishedRef, error) {
	tokenHash := sha256Hex(token)
	root := s.publishedRootDir()
	metaPath := filepath.Join(root, tokenHash+".json")
	yamlPath := filepath.Join(root, tokenHash+".yaml")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return publishedRef{}, errWorkspaceNotFound
	}
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		return publishedRef{}, errWorkspaceNotFound
	}
	var legacy legacyPublishedMeta
	if err := json.Unmarshal(metaBytes, &legacy); err != nil {
		return publishedRef{}, errWorkspaceNotFound
	}
	workspaceHash = firstNonEmptyString(workspaceHash, legacy.WorkspaceHash)
	published, err := s.createPublishedWithToken(workspaceHash, "", token)
	if err != nil {
		return publishedRef{}, err
	}
	published.Meta.CreatedAt = firstNonZeroTime(legacy.CreatedAt, published.Meta.CreatedAt)
	published.Meta.UpdatedAt = firstNonZeroTime(legacy.UpdatedAt, published.Meta.UpdatedAt)
	if err := storage.AtomicWriteFile(published.CurrentPath, yamlBytes, 0o644); err != nil {
		return publishedRef{}, fmt.Errorf("write migrated published yaml: %w", err)
	}
	if err := s.savePublishedMeta(published); err != nil {
		return publishedRef{}, err
	}
	_ = os.Remove(metaPath)
	_ = os.Remove(yamlPath)
	if strings.TrimSpace(workspaceHash) != "" {
		if ref, err := s.loadWorkspaceByHash(workspaceHash); err == nil {
			ref.Meta.PublishID = published.ID
			ref.Meta.LegacyPublishedToken = ""
			ref.Meta.LegacyPublishedAt = time.Time{}
			_ = s.saveWorkspaceMeta(ref)
		}
	}
	return published, nil
}

func (s *Server) releaseWorkspacePublishedRef(ref *workspaceRef, published publishedRef, created bool) {
	if !created || ref == nil {
		return
	}
	if ref.Meta.PublishID == published.ID {
		ref.Meta.PublishID = ""
		_ = s.saveWorkspaceMeta(*ref)
	}
	_ = os.RemoveAll(published.Dir)
}

func (s *Server) finalizePublishedRefresh(ref *workspaceRef, published *publishedRef, cfg model.Config, result pipeline.RenderResult) error {
	if published == nil {
		return nil
	}
	now := time.Now().UTC()
	published.Meta.WorkspaceHash = firstNonEmptyString(ref.Hash, published.Meta.WorkspaceHash)
	published.Meta.UpdatedAt = now
	published.Meta.SubscriptionInfo, published.Meta.SourceUserinfo = buildPublishedSubscriptionUserinfo(cfg, result.SubscriptionMeta, now)
	if err := s.savePublishedMeta(*published); err != nil {
		return err
	}
	if ref != nil {
		ref.Meta.PublishID = published.ID
		ref.Meta.LegacyPublishedToken = ""
		ref.Meta.LegacyPublishedAt = time.Time{}
		return s.saveWorkspaceMeta(*ref)
	}
	return nil
}

func buildPublishedSubscriptionUserinfo(cfg model.Config, metas map[string]model.SubscriptionMeta, updatedAt time.Time) (*publishedSubscriptionUserinfo, []publishedSourceUserinfo) {
	sources := pipeline.BuildSubscriptionMetaSources(cfg, metas)
	sourceHosts := subscriptionSourceHosts(cfg)
	sourceUserinfo := make([]publishedSourceUserinfo, 0, len(sources))

	var (
		upload     int64
		download   int64
		total      int64
		expire     int64
		validCount int
	)

	for _, meta := range sources {
		meta = model.NormalizeSubscriptionMeta(meta)
		available := meta.FromHeader
		sourceUserinfo = append(sourceUserinfo, publishedSourceUserinfo{
			SourceID:      meta.SourceID,
			SourceName:    meta.SourceName,
			SourceURLHost: firstNonEmptyString(sourceHosts.byID[meta.SourceID], sourceHosts.byName[meta.SourceName]),
			Upload:        meta.Upload,
			Download:      meta.Download,
			Total:         meta.Total,
			Expire:        meta.Expire,
			Available:     available,
			FromHeader:    meta.FromHeader,
			FromInfoNode:  meta.FromInfoNode,
			FetchedAt:     meta.FetchedAt,
		})

		if !available || meta.Total <= 0 {
			continue
		}
		validCount++
		upload += meta.Upload
		download += meta.Download
		total += meta.Total
		if meta.Expire > 0 && (expire == 0 || meta.Expire < expire) {
			expire = meta.Expire
		}
	}

	if validCount == 0 {
		return nil, sourceUserinfo
	}

	aggregate := model.AggregatedSubscriptionMeta{
		Upload:   upload,
		Download: download,
		Total:    total,
		Used:     upload + download,
		Expire:   expire,
	}
	info := &publishedSubscriptionUserinfo{
		Upload:        upload,
		Download:      download,
		Total:         total,
		Expire:        expire,
		Sources:       validCount,
		UpdatedAt:     updatedAt.UTC(),
		HeaderEnabled: model.FormatSubscriptionUserinfoHeader(aggregate) != "",
	}
	return info, sourceUserinfo
}

type subscriptionSourceHostIndex struct {
	byID   map[string]string
	byName map[string]string
}

func subscriptionSourceHosts(cfg model.Config) subscriptionSourceHostIndex {
	index := subscriptionSourceHostIndex{
		byID:   map[string]string{},
		byName: map[string]string{},
	}
	for _, sub := range cfg.Subscriptions {
		host := subscriptionURLHost(sub.URL)
		if host == "" {
			continue
		}
		if sourceID := strings.TrimSpace(sub.ID); sourceID != "" {
			index.byID[sourceID] = host
		}
		if name := strings.TrimSpace(sub.Name); name != "" {
			index.byName[name] = host
		}
	}
	return index
}

func subscriptionURLHost(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func (s *Server) loadPublishedYAML(token string) ([]byte, publishedRef, error) {
	published, err := s.loadPublishedByToken(token)
	if err != nil {
		return nil, publishedRef{}, err
	}
	if published.Meta.Revoked {
		return nil, publishedRef{}, errWorkspaceNotFound
	}
	published = s.ensurePublishedSubscriptionUserinfo(published)
	data, err := os.ReadFile(published.CurrentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, publishedRef{}, errWorkspaceNotFound
		}
		return nil, publishedRef{}, err
	}
	published.Meta.LastAccessAt = time.Now().UTC()
	published.Meta.AccessCount++
	_ = s.savePublishedMeta(published)
	return data, published, nil
}

func (s *Server) ensurePublishedSubscriptionUserinfo(published publishedRef) publishedRef {
	if published.Meta.SubscriptionInfo != nil && published.Meta.SubscriptionInfo.Total > 0 {
		return published
	}

	ref, cfg, state, ok := s.loadPublishedWorkspaceState(published)
	if !ok {
		return published
	}

	info, sources := buildPublishedSubscriptionUserinfo(cfg, state.SubscriptionMeta, time.Now().UTC())
	if info == nil || info.Total <= 0 {
		return published
	}

	published.Meta.WorkspaceHash = firstNonEmptyString(published.Meta.WorkspaceHash, ref.Hash)
	published.Meta.SubscriptionInfo = info
	published.Meta.SourceUserinfo = sources
	if err := s.savePublishedMeta(published); err != nil {
		s.appendLog("published subscription userinfo restore failed: publish=" + published.ID + " error=" + err.Error())
		return published
	}
	s.appendLog(fmt.Sprintf("published subscription userinfo restored: publish=%s token_hint=%s sources=%d", published.ID, published.Meta.TokenHint, info.Sources))
	return published
}

func (s *Server) loadPublishedWorkspaceState(published publishedRef) (workspaceRef, model.Config, model.NodeState, bool) {
	if hash := strings.TrimSpace(published.Meta.WorkspaceHash); hash != "" {
		if ref, cfg, state, ok := s.loadWorkspaceStateByHash(hash); ok {
			return ref, cfg, state, true
		}
	}

	root := s.workspaceRootDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		return workspaceRef{}, model.Config{}, model.NodeState{}, false
	}
	var bestRef workspaceRef
	var bestCfg model.Config
	var bestState model.NodeState
	var bestAccess time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ref, cfg, state, ok := s.loadWorkspaceStateByHash(entry.Name())
		if !ok || ref.Meta.PublishID != published.ID {
			continue
		}
		if len(state.SubscriptionMeta) == 0 {
			continue
		}
		if bestRef.Hash == "" || ref.Meta.LastAccessAt.After(bestAccess) {
			bestRef = ref
			bestCfg = cfg
			bestState = state
			bestAccess = ref.Meta.LastAccessAt
		}
	}
	if bestRef.Hash == "" {
		return workspaceRef{}, model.Config{}, model.NodeState{}, false
	}
	return bestRef, bestCfg, bestState, true
}

func (s *Server) loadWorkspaceStateByHash(hash string) (workspaceRef, model.Config, model.NodeState, bool) {
	ref, err := s.loadWorkspaceByHash(hash)
	if err != nil {
		return workspaceRef{}, model.Config{}, model.NodeState{}, false
	}
	cfg, err := s.loadWorkspaceConfig(ref)
	if err != nil {
		return workspaceRef{}, model.Config{}, model.NodeState{}, false
	}
	state, err := pipeline.LoadNodeState(cfg)
	if err != nil {
		return workspaceRef{}, model.Config{}, model.NodeState{}, false
	}
	return ref, cfg, state, true
}

func (s *Server) loadPublishedByToken(token string) (publishedRef, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return publishedRef{}, errWorkspaceNotFound
	}
	tokenHash := sha256Hex(token)
	root := s.publishedRootDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return publishedRef{}, errWorkspaceNotFound
		}
		return publishedRef{}, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		published, err := s.loadPublishedByID(entry.Name())
		if err != nil {
			continue
		}
		if published.Meta.TokenHash == tokenHash && !published.Meta.Revoked {
			return published, nil
		}
	}
	if published, err := s.migrateLegacyPublishedToken(token, ""); err == nil {
		return published, nil
	}
	return publishedRef{}, errWorkspaceNotFound
}

func (s *Server) rotatePublishedToken(publishID string) (publishedRef, error) {
	published, err := s.loadPublishedByID(publishID)
	if err != nil {
		return publishedRef{}, err
	}
	token, err := randomSubscriptionToken()
	if err != nil {
		return publishedRef{}, err
	}
	now := time.Now().UTC()
	published.Meta.Token = token
	published.Meta.TokenHash = sha256Hex(token)
	published.Meta.TokenHint = publishedTokenHint(token)
	published.Meta.UpdatedAt = now
	published.Meta.RotatedAt = now
	published.Meta.Revoked = false
	if err := s.savePublishedMeta(published); err != nil {
		return publishedRef{}, err
	}
	return published, nil
}

func (s *Server) deletePublished(publishID string) error {
	published, err := s.loadPublishedByID(publishID)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(published.Dir); err != nil {
		return err
	}
	return s.clearPublishedAssociations(publishID)
}

func (s *Server) clearPublishedAssociations(publishID string) error {
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
		if ref.Meta.PublishID != publishID {
			continue
		}
		ref.Meta.PublishID = ""
		ref.Meta.LegacyPublishedToken = ""
		ref.Meta.LegacyPublishedAt = time.Time{}
		_ = s.saveWorkspaceMeta(ref)
	}
	return nil
}

func (s *Server) cleanupStalePublished() error {
	days := s.snapshotConfig().Service.PublishedDeleteIfNotAccessedDays
	if days <= 0 {
		return nil
	}
	root := s.publishedRootDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		published, err := s.loadPublishedByID(entry.Name())
		if err != nil {
			continue
		}
		lastAccess := firstNonZeroTime(published.Meta.LastAccessAt, published.Meta.UpdatedAt, published.Meta.CreatedAt)
		if lastAccess.IsZero() || lastAccess.After(cutoff) {
			continue
		}
		_ = os.RemoveAll(published.Dir)
		_ = s.clearPublishedAssociations(published.ID)
	}
	return nil
}

func randomPublishID() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate publish id: %w", err)
	}
	return "p_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func randomSubscriptionToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate subscription token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func publishedTokenHint(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return token
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func publishedURL(origin, token string) string {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	token = strings.TrimSpace(token)
	if origin == "" || token == "" {
		return ""
	}
	return origin + "/s/" + token + "/mihomo.yaml"
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value.UTC()
		}
	}
	return time.Time{}
}
