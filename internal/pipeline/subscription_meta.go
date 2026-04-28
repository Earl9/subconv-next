package pipeline

import (
	"bytes"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
	"subconv-next/internal/parser"
)

type parsedSubscriptionMeta struct {
	meta         model.SubscriptionMeta
	hasUpload    bool
	hasDownload  bool
	hasTotal     bool
	hasExpire    bool
	hasUsed      bool
	hasRemaining bool
}

var (
	usedTrafficPattern      = regexp.MustCompile(`(?i)(?:已用(?:流量)?|已使用|used)\s*[:：]?\s*([0-9]+(?:\.[0-9]+)?)\s*([kmgtpe]?b)`)
	remainingTrafficPattern = regexp.MustCompile(`(?i)(?:剩余(?:流量)?|剩餘(?:流量)?|remaining)\s*[:：]?\s*([0-9]+(?:\.[0-9]+)?)\s*([kmgtpe]?b)`)
	totalTrafficPattern     = regexp.MustCompile(`(?i)(?:总(?:流量|量)?|總(?:流量|量)?|total)\s*[:：]?\s*([0-9]+(?:\.[0-9]+)?)\s*([kmgtpe]?b)`)
	expireDatePattern       = regexp.MustCompile(`(?i)(?:到期(?:时间)?|過期(?:時間)?|过期(?:时间)?|expire(?:d)?(?:\s*time)?)\s*[:：]?\s*([0-9]{4}[-/.][0-9]{1,2}[-/.][0-9]{1,2}(?:[ T][0-9]{1,2}:[0-9]{2}(?::[0-9]{2})?)?)`)
	expireUnixPattern       = regexp.MustCompile(`(?i)(?:到期(?:时间)?|過期(?:時間)?|过期(?:时间)?|expire(?:d)?(?:\s*time)?)\s*[:：]?\s*([0-9]{10,13})`)
)

func buildSourceSubscriptionMeta(source model.SourceInfo, fetched fetcher.FetchedSubscription, parsedNodes []model.NodeIR) (model.SubscriptionMeta, bool) {
	headerMeta := parseSubscriptionUserinfoHeader(fetched.SubscriptionUserinfo, source, fetched.FetchedAt)
	infoMeta := parseInfoNodeMeta(source, parsedNodes, fetched.Content, fetched.FetchedAt)
	merged := mergeSourceSubscriptionMeta(source, headerMeta, infoMeta, fetched.FetchedAt)
	if merged.SourceID == "" && merged.SourceName == "" && merged.Total == 0 && merged.Used == 0 && merged.Remaining == 0 && merged.Expire == 0 && merged.Upload == 0 && merged.Download == 0 {
		return model.SubscriptionMeta{}, false
	}
	return model.NormalizeSubscriptionMeta(merged), true
}

func parseSubscriptionUserinfoHeader(raw string, source model.SourceInfo, fetchedAt time.Time) parsedSubscriptionMeta {
	out := parsedSubscriptionMeta{
		meta: model.SubscriptionMeta{
			SourceID:   source.ID,
			SourceName: source.Name,
			FromHeader: true,
			FetchedAt:  fetchedAt.UTC().Format(time.RFC3339),
		},
	}

	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		value := strings.TrimSpace(kv[1])
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "upload":
			out.meta.Upload = n
			out.hasUpload = true
		case "download":
			out.meta.Download = n
			out.hasDownload = true
		case "total":
			out.meta.Total = n
			out.hasTotal = true
		case "expire":
			out.meta.Expire = n
			out.hasExpire = true
		}
	}

	if out.hasUpload || out.hasDownload {
		out.meta.Used = out.meta.Upload + out.meta.Download
		out.hasUsed = true
	}
	if out.hasTotal && out.hasUsed {
		out.meta.Remaining = out.meta.Total - out.meta.Used
		out.hasRemaining = true
	}
	return out
}

func parseInfoNodeMeta(source model.SourceInfo, nodes []model.NodeIR, content []byte, fetchedAt time.Time) parsedSubscriptionMeta {
	texts := make([]string, 0, len(nodes)+8)
	for _, node := range nodes {
		if !sameSource(node.Source, source) {
			continue
		}
		if looksLikeInfoNode(node.Name) {
			texts = append(texts, node.Name)
		}
	}
	texts = append(texts, extractMetaLinesFromContent(content)...)

	out := parsedSubscriptionMeta{
		meta: model.SubscriptionMeta{
			SourceID:     source.ID,
			SourceName:   source.Name,
			FromInfoNode: true,
			FetchedAt:    fetchedAt.UTC().Format(time.RFC3339),
		},
	}

	for _, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if value, ok := extractTrafficValue(text, usedTrafficPattern); ok {
			out.meta.Used = value
			out.hasUsed = true
		}
		if value, ok := extractTrafficValue(text, remainingTrafficPattern); ok {
			out.meta.Remaining = value
			out.hasRemaining = true
		}
		if value, ok := extractTrafficValue(text, totalTrafficPattern); ok {
			out.meta.Total = value
			out.hasTotal = true
		}
		if expire, ok := extractExpireValue(text); ok {
			out.meta.Expire = expire
			out.hasExpire = true
		}
	}

	if out.hasTotal && out.hasRemaining && !out.hasUsed {
		out.meta.Used = out.meta.Total - out.meta.Remaining
		out.hasUsed = true
	}
	if out.hasTotal && out.hasUsed && !out.hasRemaining {
		out.meta.Remaining = out.meta.Total - out.meta.Used
		out.hasRemaining = true
	}
	if !out.hasTotal && out.hasUsed && out.hasRemaining {
		out.meta.Total = out.meta.Used + out.meta.Remaining
		out.hasTotal = true
	}
	return out
}

func mergeSourceSubscriptionMeta(source model.SourceInfo, headerMeta, infoMeta parsedSubscriptionMeta, fetchedAt time.Time) model.SubscriptionMeta {
	merged := model.SubscriptionMeta{
		SourceID:   source.ID,
		SourceName: source.Name,
		FetchedAt:  fetchedAt.UTC().Format(time.RFC3339),
	}

	if infoMeta.hasUpload {
		merged.Upload = infoMeta.meta.Upload
	}
	if infoMeta.hasDownload {
		merged.Download = infoMeta.meta.Download
	}
	if infoMeta.hasTotal {
		merged.Total = infoMeta.meta.Total
	}
	if infoMeta.hasExpire {
		merged.Expire = infoMeta.meta.Expire
	}
	if infoMeta.hasUsed {
		merged.Used = infoMeta.meta.Used
	}
	if infoMeta.hasRemaining {
		merged.Remaining = infoMeta.meta.Remaining
	}
	merged.FromInfoNode = infoMeta.hasUpload || infoMeta.hasDownload || infoMeta.hasTotal || infoMeta.hasExpire || infoMeta.hasUsed || infoMeta.hasRemaining

	if headerMeta.hasUpload {
		merged.Upload = headerMeta.meta.Upload
	}
	if headerMeta.hasDownload {
		merged.Download = headerMeta.meta.Download
	}
	if headerMeta.hasTotal {
		merged.Total = headerMeta.meta.Total
	}
	if headerMeta.hasExpire {
		merged.Expire = headerMeta.meta.Expire
	}
	if headerMeta.hasUsed {
		merged.Used = headerMeta.meta.Used
	}
	if headerMeta.hasRemaining {
		merged.Remaining = headerMeta.meta.Remaining
	}
	merged.FromHeader = headerMeta.hasUpload || headerMeta.hasDownload || headerMeta.hasTotal || headerMeta.hasExpire
	return merged
}

func extractMetaLinesFromContent(content []byte) []string {
	decoded := content
	if parser.Detect(content) == parser.InputKindBase64 {
		if raw, err := parser.DecodeBase64Bytes(content); err == nil {
			decoded = raw
		}
	}

	lines := bytes.Split(decoded, []byte{'\n'})
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		text := strings.TrimSpace(string(line))
		if text == "" {
			continue
		}
		if looksLikeInfoNode(text) {
			out = append(out, decodeInfoText(text))
		}
	}
	return out
}

func decodeInfoText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if decoded, err := url.PathUnescape(value); err == nil {
		return decoded
	}
	return value
}

func extractTrafficValue(text string, pattern *regexp.Regexp) (int64, bool) {
	match := pattern.FindStringSubmatch(text)
	if len(match) != 3 {
		return 0, false
	}
	return parseHumanTraffic(match[1], match[2])
}

func parseHumanTraffic(number, unit string) (int64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(number), 64)
	if err != nil {
		return 0, false
	}

	multiplier := float64(1)
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "", "B":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "PB":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 0, false
	}
	return int64(value * multiplier), true
}

func extractExpireValue(text string) (int64, bool) {
	if match := expireUnixPattern.FindStringSubmatch(text); len(match) == 2 {
		expire, err := strconv.ParseInt(match[1], 10, 64)
		if err == nil {
			if len(match[1]) == 13 {
				return expire / 1000, true
			}
			return expire, true
		}
	}
	if match := expireDatePattern.FindStringSubmatch(text); len(match) == 2 {
		for _, layout := range []string{
			"2006-01-02",
			"2006/01/02",
			"2006.01.02",
			"2006-1-2",
			"2006/1/2",
			"2006.1.2",
			"2006-01-02 15:04",
			"2006/01/02 15:04",
			"2006.01.02 15:04",
			"2006-01-02 15:04:05",
			"2006/01/02 15:04:05",
			"2006.01.02 15:04:05",
			time.RFC3339,
		} {
			if ts, err := time.ParseInLocation(layout, match[1], time.UTC); err == nil {
				return ts.UTC().Unix(), true
			}
		}
	}
	return 0, false
}

func sameSource(a, b model.SourceInfo) bool {
	return strings.TrimSpace(a.ID) != "" && strings.TrimSpace(a.ID) == strings.TrimSpace(b.ID)
}

func subscriptionSourceOrder(cfg model.Config) []string {
	order := make([]string, 0, len(cfg.Subscriptions))
	for _, sub := range cfg.Subscriptions {
		if !sub.Enabled {
			continue
		}
		sourceID := strings.TrimSpace(sub.ID)
		if sourceID == "" {
			continue
		}
		order = append(order, sourceID)
	}
	return order
}

func AggregateSubscriptionMetaForConfig(cfg model.Config, metas map[string]model.SubscriptionMeta) model.AggregatedSubscriptionMeta {
	info := model.NormalizeSubscriptionInfoConfig(cfg.Render.SubscriptionInfo)
	return model.AggregateSubscriptionMeta(metas, model.SubscriptionMetaAggregateOptions{
		MergeStrategy:  info.MergeStrategy,
		ExpireStrategy: info.ExpireStrategy,
		SourceOrder:    subscriptionSourceOrder(cfg),
	})
}

func BuildSubscriptionMetaSources(cfg model.Config, stored map[string]model.SubscriptionMeta) []model.SubscriptionMeta {
	stored = cloneSubscriptionMetaMap(stored)
	sources := make([]model.SubscriptionMeta, 0, len(cfg.Subscriptions)+len(stored))
	seen := make(map[string]struct{}, len(cfg.Subscriptions)+len(stored))

	for _, sub := range cfg.Subscriptions {
		sourceID := strings.TrimSpace(sub.ID)
		if sourceID == "" {
			continue
		}
		meta, ok := stored[sourceID]
		if !ok {
			meta = model.SubscriptionMeta{
				SourceID:   sourceID,
				SourceName: sub.Name,
			}
		}
		meta.SourceID = sourceID
		meta.SourceName = firstNonEmptyString(meta.SourceName, sub.Name)
		sources = append(sources, model.NormalizeSubscriptionMeta(meta))
		seen[sourceID] = struct{}{}
	}

	for sourceID, meta := range stored {
		if _, ok := seen[sourceID]; ok {
			continue
		}
		meta.SourceID = firstNonEmptyString(meta.SourceID, sourceID)
		sources = append(sources, model.NormalizeSubscriptionMeta(meta))
	}

	return sources
}

func cloneSubscriptionMetaMap(values map[string]model.SubscriptionMeta) map[string]model.SubscriptionMeta {
	if len(values) == 0 {
		return map[string]model.SubscriptionMeta{}
	}
	out := make(map[string]model.SubscriptionMeta, len(values))
	for key, value := range values {
		out[key] = model.NormalizeSubscriptionMeta(value)
	}
	return out
}

func BuildSubscriptionMetaHeader(cfg model.Config, aggregate model.AggregatedSubscriptionMeta) string {
	info := model.NormalizeSubscriptionInfoConfig(cfg.Render.SubscriptionInfo)
	if info == nil || !info.Enabled || !info.ExposeHeader {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(info.MergeStrategy), "none") {
		return ""
	}
	return model.FormatSubscriptionUserinfoHeader(aggregate)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
