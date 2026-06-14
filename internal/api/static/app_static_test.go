package static

import (
	"os"
	"strings"
	"testing"
)

func readAppJS(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("app.js")
	if err != nil {
		t.Fatalf("ReadFile(app.js) error = %v", err)
	}
	return string(data)
}

func TestLocalDraftDoesNotPersistFullSubscriptionTokenOrURL(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function draftPublishRefFromPublished",
		"publish_id: publishID",
		"token_hint:",
		"updated_at:",
		"payload.publish_ref = publishRef",
		"access_token: \"\"",
		"subscription_token: \"\"",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing %q", needle)
		}
	}

	draftRefStart := strings.Index(app, "function draftPublishRefFromPublished")
	draftRefEnd := strings.Index(app[draftRefStart:], "\n}\n")
	if draftRefStart < 0 || draftRefEnd < 0 {
		t.Fatalf("draftPublishRefFromPublished function not found")
	}
	draftRefBody := app[draftRefStart : draftRefStart+draftRefEnd]
	for _, forbidden := range []string{"subscription_url", ".url", "token:"} {
		if strings.Contains(draftRefBody, forbidden) {
			t.Fatalf("draftPublishRefFromPublished persists forbidden %q:\n%s", forbidden, draftRefBody)
		}
	}
}

func TestRestoreFromPublishedLinkDoesNotPersistFullPublishedURL(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"restore-from-published",
		"function restoreWorkspaceFromPublishedLink",
		"function buildRestoredPublishedDraftPayload",
		"draftPublishRefFromPublished(published)",
		"restore-published-url-input",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing restore-from-published marker %q", needle)
		}
	}

	payloadStart := strings.Index(app, "function buildRestoredPublishedDraftPayload")
	payloadEnd := strings.Index(app[payloadStart:], "\n}\n")
	if payloadStart < 0 || payloadEnd < 0 {
		t.Fatalf("buildRestoredPublishedDraftPayload function not found")
	}
	payloadBody := app[payloadStart : payloadStart+payloadEnd]
	for _, forbidden := range []string{"subscription_url", ".url", "publishedURL"} {
		if strings.Contains(payloadBody, forbidden) {
			t.Fatalf("buildRestoredPublishedDraftPayload persists forbidden %q:\n%s", forbidden, payloadBody)
		}
	}
}

func TestRestoreFromPublishedDoesNotUseBrowserConfirm(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"async function restoreWorkspaceFromPublishedLink",
		"findLocalDraftByPublishID(response.publish?.publish_id)",
		"已从订阅链接恢复为新的本机草稿。",
		"已从订阅链接更新已有本机草稿。",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing restore confirm guard %q", needle)
		}
	}

	restoreStart := strings.Index(app, "async function restoreWorkspaceFromPublishedLink")
	restoreEnd := strings.Index(app[restoreStart:], "\n}\n")
	if restoreStart < 0 || restoreEnd < 0 {
		t.Fatalf("restoreWorkspaceFromPublishedLink function not found")
	}
	restoreBody := app[restoreStart : restoreStart+restoreEnd]
	if strings.Contains(restoreBody, "window.confirm") {
		t.Fatalf("restoreWorkspaceFromPublishedLink should not use browser confirm:\n%s", restoreBody)
	}
}

func TestMultipleLocalDraftStorageHooksExist(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"const LOCAL_DRAFTS_STORAGE_KEY = \"SUBCONV_LOCAL_DRAFTS\"",
		"function loadDraftStore",
		"function saveLocalDraftPayload",
		"function loadLocalDrafts",
		"function renderDraftManager",
		"draft-manager-dialog",
		"localStorage.removeItem(LOCAL_DRAFT_STORAGE_KEY)",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing multiple-draft hook %q", needle)
		}
	}

	if strings.Contains(app, "localStorage.setItem(LOCAL_DRAFT_STORAGE_KEY") {
		t.Fatalf("app.js should not write new drafts to legacy single-draft key")
	}
}

func TestRestoreFromPublishedDedupesByPublishID(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function findLocalDraftByPublishID",
		"existingDraft?.draft_id || createLocalDraftID()",
		"preservedLocalDraftName(existingDraft)",
		"draftID: existingDraft?.draft_id || draft.draft_id",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing published-link dedupe marker %q", needle)
		}
	}
}

func TestLocalDraftAutoNameFiltersDefaultSourceNames(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function preservedLocalDraftName",
		"function isGenericLocalDraftName",
		"/^source-\\d+$/i.test(value)",
		"function localDraftSubscriptionLabel",
		"subscriptionURLHostLabel(item?.url)",
		"function formatDraftNameTime",
		"`${baseName} · ${timeLabel}`",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing local-draft naming marker %q", needle)
		}
	}
}

func TestSourceModeIsMutuallyExclusiveInGeneratedConfig(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function switchSourceMode(mode)",
		"const nextMode = mode === \"template\" ? \"template\" : \"rules\"",
		"state.activeSourceMode = nextMode",
		"state.activeSourceMode === \"template\" ? \"template\" : \"rules\"",
		"source_mode: activeMode",
		"template_rule_mode: activeMode",
		"!activeOnly || activeMode === \"template\"",
		"!activeOnly || activeMode === \"rules\"",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing source-mode guard %q", needle)
		}
	}
}

func TestManualNodeUIStateHooksExist(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function renderInlineManager()",
		"manual-node-status",
		"已解析 ·",
		"解析失败",
		"已禁用",
		"function previewInlineEntry",
		"function validateManualNodesBeforeGenerate",
		"data-inline-action=\"preview\"",
		"data-inline-action=\"clear\"",
		"data-inline-action=\"delete\"",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing manual-node hook %q", needle)
		}
	}
}

func TestNodeEditorCacheInvalidatesAfterGenerate(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"nodeCacheStale: false",
		"if (state.nodeCacheStale || !state.allNodes.length) loadNodes();",
		"function markNodeCacheStale()",
		"markNodeCacheStale();",
		"state.nodeCacheStale = false;",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing node-cache invalidation marker %q", needle)
		}
	}
}

func TestEmojiSourceNamePreviewMatchesDefaultNaming(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"source_prefix_mode: \"emoji_name\"",
		"const SOURCE_PREFIX_SEPARATOR = \"｜\"",
		"function buildSourcePrefix",
		"if (emoji && sourceName) return `${emoji} ${sourceName}`",
		"return `${prefix}${SOURCE_PREFIX_SEPARATOR}${SOURCE_NAME_PREVIEW}`",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing naming preview guard %q", needle)
		}
	}
}

func TestBuiltinRulesExposeOneDriveSeparately(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		`{ key: "onedrive", label: "微软云盘" }`,
		`onedrive: "cloud"`,
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing OneDrive rule UI marker %q", needle)
		}
	}

	oneDrivePos := strings.Index(app, `"onedrive"`)
	microsoftPos := strings.Index(app, `"microsoft"`)
	if oneDrivePos < 0 || microsoftPos < 0 || oneDrivePos > microsoftPos {
		t.Fatalf("OneDrive rule should be listed before Microsoft rule")
	}
}
