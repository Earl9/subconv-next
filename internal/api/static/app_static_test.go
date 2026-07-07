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

func TestSubscriptionLANAllowHooksExist(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function isPrivateSubscriptionURL",
		"function sourceLANWarning",
		"data-sub-field=\"allow_lan\"",
		"允许局域网订阅地址",
		"allow_lan: item.allow_lan === true",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing LAN subscription hook %q", needle)
		}
	}
}

func TestCustomDNSHooksExist(t *testing.T) {
	app := readAppJS(t)
	index, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatalf("ReadFile(index.html) error = %v", err)
	}

	requiredApp := []string{
		"custom_dns: false",
		"function normalizeDNSConfigForFrontend",
		"function readDNSFormConfig",
		"function parseNameserverPolicyText",
		"function splitDNSListItems",
		"splitDNSListItems(line).forEach",
		"custom_dns: customDNS",
		"dns: dnsConfig",
		"custom-dns-enabled",
	}
	for _, needle := range requiredApp {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing custom DNS hook %q", needle)
		}
	}

	requiredIndex := []string{
		"id=\"dns-settings-panel\"",
		"id=\"custom-dns-enabled\"",
		"value=\"127.0.0.1:5335\"",
		"value=\"198.18.0.1/16\"",
		"id=\"dns-use-hosts\" type=\"checkbox\" checked",
		"id=\"dns-nameserver\"",
		">223.5.5.5\n119.29.29.29</textarea>",
		"id=\"dns-fallback\"",
		">https://dns.alidns.com/dns-query</textarea>",
		"id=\"dns-nameserver-policy\"",
	}
	indexText := string(index)
	for _, needle := range requiredIndex {
		if !strings.Contains(indexText, needle) {
			t.Fatalf("index.html missing custom DNS hook %q", needle)
		}
	}
	rulePanelIndex := strings.Index(indexText, "id=\"template-mode-panel\"")
	dnsPanelIndex := strings.Index(indexText, "id=\"dns-settings-panel\"")
	if rulePanelIndex < 0 || dnsPanelIndex < 0 || dnsPanelIndex < rulePanelIndex {
		t.Fatalf("dns-settings-panel should be after rule/template settings")
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

func TestNodeEditorBulkDeleteHooksExist(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"function renderNodeBulkActionBar",
		"bulk-delete-selected-nodes-btn",
		"bulk-delete-current-filter-btn",
		"function openBulkDeleteDialog",
		"function performBulkDeleteNodes",
		"const ok = await performBulkDeleteNodes(config.nodes || config.nodeIds || [])",
		"nodes: targets",
		"function normalizeBulkDeleteTargets",
		"action: \"bulk-delete-nodes\"",
		"fetchJSON(\"/api/nodes/delete\"",
		"`/api/nodes/custom/${encodeURIComponent(nodeId)}`",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing bulk-delete hook %q", needle)
		}
	}
}

func TestNodeEditorDeletedRestoreHooksExist(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"restore-deleted-nodes-btn",
		"restore-deleted-dialog",
		"function openDeletedNodesDialog",
		"fetchJSON(\"/api/nodes/deleted\")",
		"data-restore-deleted-node",
		"function renderDeletedNodesDialog",
		"function restoreAllDeletedNodesFromDialog",
		"function restoreDeletedNodes",
		"fetchJSON(\"/api/nodes/enable\"",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing deleted-node restore hook %q", needle)
		}
	}
}

func TestNodeEditorPerformanceHooksExist(t *testing.T) {
	app := readAppJS(t)

	required := []string{
		"nodeById: new Map()",
		"function buildNodeIndexes",
		"node._searchText = buildNodeSearchText(node)",
		"function bindNodeTableDelegates",
		"body.dataset.delegatesBound === \"1\"",
		"function refreshNodeSelectionUI",
		"function buildNodeQueryParams",
		"function loadAllFilteredNodes",
		"state.selectedNodeCache.set(node.id, node)",
		"const localNode = state.nodeById.get(nodeId)",
	}
	for _, needle := range required {
		if !strings.Contains(app, needle) {
			t.Fatalf("app.js missing node editor performance hook %q", needle)
		}
	}

	openStart := strings.Index(app, "async function openNodeDialog")
	openEnd := strings.Index(app[openStart:], "\n}\n")
	if openStart < 0 || openEnd < 0 {
		t.Fatalf("openNodeDialog function not found")
	}
	openBody := app[openStart : openStart+openEnd]
	if strings.Contains(openBody, "state.allNodes.find") || strings.Contains(openBody, "state.nodes.find") {
		t.Fatalf("openNodeDialog should use indexed lookup instead of scanning nodes:\n%s", openBody)
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
		"function sourceNamePreviewLabel",
		"`命名预览：${sourceNamePreview(item)}`",
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
