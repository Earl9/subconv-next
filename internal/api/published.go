package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	OutputFilename   string                         `json:"output_filename,omitempty"`
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

var (
	errPublishedLimitReached = errors.New("published subscription limit reached")
	errPublishedInvalidID    = errors.New("invalid published subscription ID")
)

const publishedAccessWriteInterval = 30 * time.Second

type publishedAccessState struct {
	pending       int
	lastAccessAt  time.Time
	lastPersisted time.Time
	flushTimer    *time.Timer
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

func validPublishID(id string) bool {
	id = strings.TrimSpace(id)
	if len(id) < 3 || len(id) > 128 || !strings.HasPrefix(id, "p_") {
		return false
	}
	for _, char := range id {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '_', char == '-':
		default:
			return false
		}
	}
	return true
}

func (s *Server) createPublished(workspaceHash string) (publishedRef, error) {
	s.publishedCreateMu.Lock()
	defer s.publishedCreateMu.Unlock()

	if s.snapshotConfig().Service.PublicConverter && s.maxPublications > 0 {
		count, err := s.activePublicationCount()
		if err != nil {
			return publishedRef{}, err
		}
		if count >= s.maxPublications {
			if err := s.cleanupStalePublished(); err != nil {
				return publishedRef{}, err
			}
			count, err = s.activePublicationCount()
			if err != nil {
				return publishedRef{}, err
			}
			if count >= s.maxPublications {
				return publishedRef{}, errPublishedLimitReached
			}
		}
	}

	token, err := randomSubscriptionToken()
	if err != nil {
		return publishedRef{}, err
	}
	return s.createPublishedWithToken(workspaceHash, "", token)
}

func (s *Server) activePublicationCount() (int, error) {
	entries, err := os.ReadDir(s.publishedRootDir())
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read published root: %w", err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}
	return count, nil
}

func (s *Server) createPublishedWithToken(workspaceHash, publishID, token string) (publishedRef, error) {
	if strings.TrimSpace(publishID) == "" {
		var err error
		publishID, err = randomPublishID()
		if err != nil {
			return publishedRef{}, err
		}
	}
	if !validPublishID(publishID) {
		return publishedRef{}, errPublishedInvalidID
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
	if err := os.MkdirAll(ref.Dir, 0o700); err != nil {
		return publishedRef{}, fmt.Errorf("create published dir: %w", err)
	}
	if err := s.savePublishedMeta(ref); err != nil {
		return publishedRef{}, err
	}
	return ref, nil
}

func (s *Server) loadPublishedByID(id string) (publishedRef, error) {
	s.publishedMetaMu.Lock()
	defer s.publishedMetaMu.Unlock()
	return s.loadPublishedByIDUnlocked(id)
}

func (s *Server) loadPublishedByIDUnlocked(id string) (publishedRef, error) {
	id = strings.TrimSpace(id)
	if !validPublishID(id) {
		return publishedRef{}, errWorkspaceNotFound
	}
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
	ref.Meta.PublishID = firstNonEmptyString(ref.Meta.PublishID, ref.ID)
	if !validPublishID(ref.Meta.PublishID) || ref.Meta.PublishID != ref.ID {
		return publishedRef{}, fmt.Errorf("decode published meta: %w", errPublishedInvalidID)
	}
	return ref, nil
}

func (s *Server) savePublishedMeta(ref publishedRef) error {
	ref.Meta.PublishID = firstNonEmptyString(ref.Meta.PublishID, ref.ID)
	if !validPublishID(ref.ID) || ref.Meta.PublishID != ref.ID {
		return errPublishedInvalidID
	}
	canonical := s.buildPublishedRef(ref.ID)
	ref.Dir = canonical.Dir
	ref.CurrentPath = canonical.CurrentPath
	ref.MetaPath = canonical.MetaPath
	s.publishedMetaMu.Lock()
	oldTokenHash := ""
	if current, err := s.loadPublishedByIDUnlocked(ref.ID); err == nil {
		oldTokenHash = current.Meta.TokenHash
	}
	err := s.savePublishedMetaUnlocked(ref)
	s.publishedMetaMu.Unlock()
	if err != nil {
		return err
	}
	newTokenHash := strings.TrimSpace(ref.Meta.TokenHash)
	if token := strings.TrimSpace(ref.Meta.Token); token != "" {
		newTokenHash = sha256Hex(token)
	}
	s.replacePublishedTokenIndex(oldTokenHash, newTokenHash, firstNonEmptyString(ref.Meta.PublishID, ref.ID))
	return nil
}

func (s *Server) savePublishedMetaUnlocked(ref publishedRef) error {
	ref.Meta.PublishID = firstNonEmptyString(ref.Meta.PublishID, ref.ID)
	if !validPublishID(ref.ID) || ref.Meta.PublishID != ref.ID {
		return errPublishedInvalidID
	}
	canonical := s.buildPublishedRef(ref.ID)
	ref.Dir = canonical.Dir
	ref.CurrentPath = canonical.CurrentPath
	ref.MetaPath = canonical.MetaPath
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
	if err := storage.AtomicWriteFile(ref.MetaPath, data, 0o600); err != nil {
		return fmt.Errorf("write published meta: %w", err)
	}
	return nil
}

func (s *Server) updatePublishedMeta(id string, update func(*publishedRef) error) (publishedRef, error) {
	s.publishedMetaMu.Lock()
	published, err := s.loadPublishedByIDUnlocked(id)
	if err != nil {
		s.publishedMetaMu.Unlock()
		return publishedRef{}, err
	}
	oldTokenHash := published.Meta.TokenHash
	if update != nil {
		if err := update(&published); err != nil {
			s.publishedMetaMu.Unlock()
			return publishedRef{}, err
		}
	}
	if err := s.savePublishedMetaUnlocked(published); err != nil {
		s.publishedMetaMu.Unlock()
		return publishedRef{}, err
	}
	s.publishedMetaMu.Unlock()
	s.replacePublishedTokenIndex(oldTokenHash, published.Meta.TokenHash, published.ID)
	return published, nil
}

func (s *Server) replacePublishedTokenIndex(oldTokenHash, newTokenHash, publishID string) {
	oldTokenHash = strings.TrimSpace(oldTokenHash)
	newTokenHash = strings.TrimSpace(newTokenHash)
	publishID = strings.TrimSpace(publishID)
	s.publishedIndexMu.Lock()
	defer s.publishedIndexMu.Unlock()
	if oldTokenHash != "" && oldTokenHash != newTokenHash && s.publishedTokenIndex[oldTokenHash] == publishID {
		delete(s.publishedTokenIndex, oldTokenHash)
	}
	if newTokenHash != "" && publishID != "" {
		s.publishedTokenIndex[newTokenHash] = publishID
	}
}

func (s *Server) removePublishedTokenIndex(tokenHash, publishID string) {
	s.publishedIndexMu.Lock()
	defer s.publishedIndexMu.Unlock()
	if s.publishedTokenIndex[strings.TrimSpace(tokenHash)] == strings.TrimSpace(publishID) {
		delete(s.publishedTokenIndex, strings.TrimSpace(tokenHash))
	}
}

func (s *Server) indexedPublishID(tokenHash string) string {
	s.publishedIndexMu.RLock()
	defer s.publishedIndexMu.RUnlock()
	return s.publishedTokenIndex[strings.TrimSpace(tokenHash)]
}

func (s *Server) ensurePublishedTokenIndex() error {
	s.publishedIndexMu.RLock()
	loaded := s.publishedIndexLoaded
	s.publishedIndexMu.RUnlock()
	if loaded {
		return nil
	}

	entries, err := os.ReadDir(s.publishedRootDir())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	discovered := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		published, err := s.loadPublishedByID(entry.Name())
		if err != nil || published.Meta.Revoked || strings.TrimSpace(published.Meta.TokenHash) == "" {
			continue
		}
		discovered[published.Meta.TokenHash] = published.ID
	}

	s.publishedIndexMu.Lock()
	for tokenHash, publishID := range discovered {
		if _, exists := s.publishedTokenIndex[tokenHash]; !exists {
			s.publishedTokenIndex[tokenHash] = publishID
		}
	}
	s.publishedIndexLoaded = true
	s.publishedIndexMu.Unlock()
	return nil
}

func (s *Server) ensureWorkspacePublishedRef(ref *workspaceRef) (publishedRef, bool, error) {
	if ref == nil {
		return publishedRef{}, false, fmt.Errorf("workspace ref is nil")
	}
	if strings.TrimSpace(ref.Meta.PublishID) != "" {
		published, err := s.loadPublishedByID(ref.Meta.PublishID)
		if err == nil && (publishedRestorable(published) || emptyPublishedBoundToWorkspace(published, ref.Hash)) {
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

// A node deletion may intentionally remove current.yaml when the publication
// becomes empty. Keep its existing bearer token bound to the same workspace so
// a later generation updates that publication instead of creating a new link.
func emptyPublishedBoundToWorkspace(published publishedRef, workspaceHash string) bool {
	if published.Meta.Revoked || strings.TrimSpace(published.Meta.Token) == "" {
		return false
	}
	bound := strings.TrimSpace(published.Meta.WorkspaceHash)
	workspaceHash = strings.TrimSpace(workspaceHash)
	return workspaceHash != "" && bound == workspaceHash
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
	if err := storage.AtomicWriteFile(published.CurrentPath, yamlBytes, 0o600); err != nil {
		return publishedRef{}, fmt.Errorf("write migrated published yaml: %w", err)
	}
	published, err = s.updatePublishedMeta(published.ID, func(current *publishedRef) error {
		current.Meta.CreatedAt = firstNonZeroTime(legacy.CreatedAt, current.Meta.CreatedAt)
		current.Meta.UpdatedAt = firstNonZeroTime(legacy.UpdatedAt, current.Meta.UpdatedAt)
		return nil
	})
	if err != nil {
		return publishedRef{}, err
	}
	_ = os.Remove(metaPath)
	_ = os.Remove(yamlPath)
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
	info, sources := buildPublishedSubscriptionUserinfo(cfg, result.SubscriptionMeta, now)
	updated, err := s.updatePublishedMeta(published.ID, func(current *publishedRef) error {
		if ref != nil {
			current.Meta.WorkspaceHash = firstNonEmptyString(ref.Hash, current.Meta.WorkspaceHash)
		}
		current.Meta.UpdatedAt = now
		current.Meta.OutputFilename = publishedOutputFilenameFromConfig(cfg)
		current.Meta.SubscriptionInfo = info
		current.Meta.SourceUserinfo = sources
		return nil
	})
	if err != nil {
		if errors.Is(err, errWorkspaceNotFound) {
			_ = os.RemoveAll(published.Dir)
		}
		return err
	}
	*published = updated
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

func (s *Server) loadPublishedYAML(token string) ([]byte, publishedRef, bool, error) {
	published, err := s.loadPublishedByToken(token)
	if err != nil {
		return nil, publishedRef{}, false, err
	}
	if published.Meta.Revoked {
		return nil, publishedRef{}, false, errWorkspaceNotFound
	}
	published = s.refreshPublishedOnRequest(published)
	published = s.ensurePublishedSubscriptionUserinfo(published)
	data, err := os.ReadFile(published.CurrentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, publishedRef{}, false, errWorkspaceNotFound
		}
		return nil, publishedRef{}, false, err
	}
	published, shouldLog := s.recordPublishedAccess(published)
	return data, published, shouldLog, nil
}

func (s *Server) recordPublishedAccess(published publishedRef) (publishedRef, bool) {
	now := time.Now().UTC()
	s.publishedAccessMu.Lock()
	state := s.publishedAccess[published.ID]
	if state == nil {
		state = &publishedAccessState{}
		s.publishedAccess[published.ID] = state
	}
	state.pending++
	state.lastAccessAt = now
	shouldPersist := state.lastPersisted.IsZero() || now.Sub(state.lastPersisted) >= publishedAccessWriteInterval
	count := 0
	if shouldPersist {
		count = state.pending
		state.pending = 0
		state.lastPersisted = now
		if state.flushTimer != nil {
			state.flushTimer.Stop()
			state.flushTimer = nil
		}
	} else if state.flushTimer == nil {
		delay := publishedAccessWriteInterval - now.Sub(state.lastPersisted)
		state.flushTimer = time.AfterFunc(delay, func() { s.flushPublishedAccess(published.ID) })
	}
	s.publishedAccessMu.Unlock()

	if !shouldPersist {
		return published, false
	}
	updated, err := s.persistPublishedAccess(published.ID, count, now)
	if err != nil {
		s.requeuePublishedAccess(published.ID, count, now)
		return published, false
	}
	return updated, true
}

func (s *Server) flushPublishedAccess(publishID string) {
	s.publishedAccessMu.Lock()
	state := s.publishedAccess[publishID]
	if state == nil {
		s.publishedAccessMu.Unlock()
		return
	}
	count := state.pending
	lastAccessAt := state.lastAccessAt
	state.pending = 0
	state.lastPersisted = time.Now().UTC()
	state.flushTimer = nil
	s.publishedAccessMu.Unlock()
	if count == 0 {
		return
	}
	if _, err := s.persistPublishedAccess(publishID, count, lastAccessAt); err != nil {
		s.requeuePublishedAccess(publishID, count, lastAccessAt)
	}
}

func (s *Server) persistPublishedAccess(publishID string, count int, lastAccessAt time.Time) (publishedRef, error) {
	return s.updatePublishedMeta(publishID, func(current *publishedRef) error {
		if lastAccessAt.After(current.Meta.LastAccessAt) {
			current.Meta.LastAccessAt = lastAccessAt
		}
		current.Meta.AccessCount += count
		return nil
	})
}

func (s *Server) requeuePublishedAccess(publishID string, count int, lastAccessAt time.Time) {
	s.publishedAccessMu.Lock()
	defer s.publishedAccessMu.Unlock()
	state := s.publishedAccess[publishID]
	if state == nil {
		state = &publishedAccessState{}
		s.publishedAccess[publishID] = state
	}
	state.pending += count
	if lastAccessAt.After(state.lastAccessAt) {
		state.lastAccessAt = lastAccessAt
	}
	if state.flushTimer == nil {
		state.flushTimer = time.AfterFunc(publishedAccessWriteInterval, func() { s.flushPublishedAccess(publishID) })
	}
}

func (s *Server) forgetPublishedAccess(publishID string) {
	s.publishedAccessMu.Lock()
	defer s.publishedAccessMu.Unlock()
	if state := s.publishedAccess[publishID]; state != nil && state.flushTimer != nil {
		state.flushTimer.Stop()
	}
	delete(s.publishedAccess, publishID)
}

func (s *Server) refreshPublishedOnRequest(published publishedRef) publishedRef {
	ref, cfg, _, ok := s.loadPublishedWorkspaceState(published)
	if !ok || !cfg.Service.RefreshOnRequest {
		return published
	}
	if !cacheExpired(published.CurrentPath, effectiveRefreshInterval(cfg)) {
		return published
	}
	_, refreshed, err := s.refreshPublishedWorkspace(ref, "published subscription refresh")
	if err != nil {
		if errors.Is(err, ErrRefreshInProgress) || errors.Is(err, ErrRefreshCapacity) {
			s.appendLog("published subscription refresh skipped: publish=" + published.ID + " refresh already running")
			s.appendWorkspaceLog(ref.Hash, "published subscription refresh skipped: refresh already running")
			return published
		}
		s.appendLog("published subscription refresh failed: publish=" + published.ID + " error=" + err.Error())
		s.appendWorkspaceLog(ref.Hash, "published subscription refresh failed: "+err.Error())
		return published
	}
	return refreshed
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

	updated, err := s.updatePublishedMeta(published.ID, func(current *publishedRef) error {
		if current.Meta.SubscriptionInfo != nil && current.Meta.SubscriptionInfo.Total > 0 {
			return nil
		}
		current.Meta.WorkspaceHash = firstNonEmptyString(current.Meta.WorkspaceHash, ref.Hash)
		current.Meta.SubscriptionInfo = info
		current.Meta.SourceUserinfo = sources
		return nil
	})
	if err != nil {
		s.appendLog("published subscription userinfo restore failed: publish=" + published.ID + " error=" + err.Error())
		return published
	}
	published = updated
	s.appendLog(fmt.Sprintf("published subscription userinfo restored: publish=%s token_hint=%s sources=%d", published.ID, published.Meta.TokenHint, info.Sources))
	return published
}

func (s *Server) loadPublishedWorkspaceState(published publishedRef) (workspaceRef, model.Config, model.NodeState, bool) {
	return s.loadPublishedWorkspaceStateWithLock(published, "")
}

func (s *Server) loadPublishedWorkspaceStateWithLock(published publishedRef, lockedWorkspaceHash string) (workspaceRef, model.Config, model.NodeState, bool) {
	if hash := strings.TrimSpace(published.Meta.WorkspaceHash); hash != "" {
		if ref, cfg, state, ok := s.loadWorkspaceStateByHashWithLock(hash, lockedWorkspaceHash); ok {
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
		ref, cfg, state, ok := s.loadWorkspaceStateByHashWithLock(entry.Name(), lockedWorkspaceHash)
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
	return s.loadWorkspaceStateByHashWithLock(hash, "")
}

func (s *Server) loadWorkspaceStateByHashWithLock(hash, lockedWorkspaceHash string) (workspaceRef, model.Config, model.NodeState, bool) {
	unlock := func() {}
	if strings.TrimSpace(hash) != strings.TrimSpace(lockedWorkspaceHash) {
		unlock = s.lockWorkspaceHash(hash)
	}
	defer unlock()
	return s.loadWorkspaceStateByHashUnlocked(hash)
}

func (s *Server) loadWorkspaceStateByHashUnlocked(hash string) (workspaceRef, model.Config, model.NodeState, bool) {
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
	if publishID := s.indexedPublishID(tokenHash); publishID != "" {
		published, err := s.loadPublishedByID(publishID)
		if err == nil && published.Meta.TokenHash == tokenHash && !published.Meta.Revoked {
			return published, nil
		}
		s.removePublishedTokenIndex(tokenHash, publishID)
	}
	if err := s.ensurePublishedTokenIndex(); err != nil {
		return publishedRef{}, err
	}
	if publishID := s.indexedPublishID(tokenHash); publishID != "" {
		published, err := s.loadPublishedByID(publishID)
		if err == nil && published.Meta.TokenHash == tokenHash && !published.Meta.Revoked {
			return published, nil
		}
		s.removePublishedTokenIndex(tokenHash, publishID)
	}
	if published, err := s.migrateLegacyPublishedToken(token, ""); err == nil {
		return published, nil
	}
	return publishedRef{}, errWorkspaceNotFound
}

func (s *Server) rotatePublishedToken(publishID string) (publishedRef, error) {
	token, err := randomSubscriptionToken()
	if err != nil {
		return publishedRef{}, err
	}
	now := time.Now().UTC()
	return s.updatePublishedMeta(publishID, func(published *publishedRef) error {
		published.Meta.Token = token
		published.Meta.TokenHash = sha256Hex(token)
		published.Meta.TokenHint = publishedTokenHint(token)
		published.Meta.UpdatedAt = now
		published.Meta.RotatedAt = now
		published.Meta.Revoked = false
		return nil
	})
}

func (s *Server) deletePublished(publishID string) error {
	return s.deletePublishedForWorkspace(publishID, "")
}

func (s *Server) deletePublishedForWorkspace(publishID, lockedWorkspaceHash string) error {
	_, err := s.removePublishedIfWithWorkspace(publishID, lockedWorkspaceHash, func(publishedRef) bool { return true })
	return err
}

func (s *Server) removePublishedIf(publishID string, shouldRemove func(publishedRef) bool) (bool, error) {
	return s.removePublishedIfWithWorkspace(publishID, "", shouldRemove)
}

func (s *Server) removePublishedIfWithWorkspace(publishID, lockedWorkspaceHash string, shouldRemove func(publishedRef) bool) (bool, error) {
	s.publishedMetaMu.Lock()
	published, err := s.loadPublishedByIDUnlocked(publishID)
	if err != nil {
		s.publishedMetaMu.Unlock()
		return false, err
	}
	if shouldRemove != nil && !shouldRemove(published) {
		s.publishedMetaMu.Unlock()
		return false, nil
	}
	err = os.RemoveAll(published.Dir)
	s.publishedMetaMu.Unlock()
	if err != nil {
		return false, err
	}
	s.removePublishedTokenIndex(published.Meta.TokenHash, published.ID)
	s.forgetPublishedAccess(published.ID)
	if err := s.clearPublishedAssociations(publishID, lockedWorkspaceHash); err != nil {
		return true, err
	}
	return true, nil
}

func (s *Server) clearPublishedAssociations(publishID, lockedWorkspaceHash string) error {
	root := s.workspaceRootDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var updateErrors []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		unlock := func() {}
		if entry.Name() != strings.TrimSpace(lockedWorkspaceHash) {
			unlock = s.lockWorkspaceHash(entry.Name())
		}
		ref, err := s.loadWorkspaceByHash(entry.Name())
		if err != nil {
			unlock()
			continue
		}
		if ref.Meta.PublishID != publishID {
			unlock()
			continue
		}
		ref.Meta.PublishID = ""
		ref.Meta.LegacyPublishedToken = ""
		ref.Meta.LegacyPublishedAt = time.Time{}
		if err := s.saveWorkspaceMeta(ref); err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("clear publish association from workspace %s: %w", entry.Name(), err))
		}
		unlock()
	}
	return errors.Join(updateErrors...)
}

func (s *Server) cleanupStalePublished() error {
	service := s.snapshotConfig().Service
	days := service.PublishedDeleteIfNotAccessedDays
	if service.PublicConverter && days <= 0 {
		days = 30
	}
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
	var cleanupErrors []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		_, err := s.removePublishedIf(entry.Name(), func(published publishedRef) bool {
			lastAccess := firstNonZeroTime(published.Meta.LastAccessAt, published.Meta.UpdatedAt, published.Meta.CreatedAt)
			return !lastAccess.IsZero() && !lastAccess.After(cutoff)
		})
		if err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("remove stale published item %s: %w", entry.Name(), err))
		}
	}
	return errors.Join(cleanupErrors...)
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

func publishedURL(origin, token string, filenames ...string) string {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	token = strings.TrimSpace(token)
	if origin == "" || token == "" {
		return ""
	}
	filename := "mihomo.yaml"
	if len(filenames) > 0 {
		filename = publishedOutputFilename(filenames[0])
	}
	return origin + "/s/" + url.PathEscape(token) + "/" + url.PathEscape(filename)
}

func publishedTokenFromSubscriptionURL(rawURL string) (string, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	path := strings.Trim(parsed.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[0] != "s" || strings.TrimSpace(parts[2]) == "" {
		return "", false
	}
	token, err := url.PathUnescape(parts[1])
	if err != nil {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

func publishedOutputFilenameFromConfig(cfg model.Config) string {
	return publishedOutputFilename(cfg.Render.OutputFilename)
}

func publishedOutputFilename(value string) string {
	return subscriptionOutputFilename(value)
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value.UTC()
		}
	}
	return time.Time{}
}
