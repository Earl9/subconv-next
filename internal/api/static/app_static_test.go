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
