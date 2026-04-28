package model

import (
	"fmt"
	"sort"
	"strings"
)

type SubscriptionMeta struct {
	SourceID     string  `json:"source_id,omitempty"`
	SourceName   string  `json:"source_name,omitempty"`
	Upload       int64   `json:"upload,omitempty"`
	Download     int64   `json:"download,omitempty"`
	Total        int64   `json:"total,omitempty"`
	Expire       int64   `json:"expire,omitempty"`
	Used         int64   `json:"used,omitempty"`
	Remaining    int64   `json:"remaining,omitempty"`
	UsedRatio    float64 `json:"used_ratio,omitempty"`
	FromHeader   bool    `json:"from_header,omitempty"`
	FromInfoNode bool    `json:"from_info_node,omitempty"`
	FetchedAt    string  `json:"fetched_at,omitempty"`
}

type AggregatedSubscriptionMeta struct {
	Upload           int64   `json:"upload,omitempty"`
	Download         int64   `json:"download,omitempty"`
	Total            int64   `json:"total,omitempty"`
	Used             int64   `json:"used,omitempty"`
	Remaining        int64   `json:"remaining,omitempty"`
	UsedRatio        float64 `json:"used_ratio,omitempty"`
	Expire           int64   `json:"expire,omitempty"`
	ExpireSourceID   string  `json:"expire_source_id,omitempty"`
	ExpireSourceName string  `json:"expire_source_name,omitempty"`
}

type SubscriptionInfoConfig struct {
	Enabled        bool   `json:"enabled"`
	ExposeHeader   bool   `json:"expose_header"`
	ShowPerSource  bool   `json:"show_per_source"`
	MergeStrategy  string `json:"merge_strategy,omitempty"`
	ExpireStrategy string `json:"expire_strategy,omitempty"`
}

type SubscriptionMetaAggregateOptions struct {
	MergeStrategy  string
	ExpireStrategy string
	SourceOrder    []string
}

func DefaultSubscriptionInfoConfig() *SubscriptionInfoConfig {
	return &SubscriptionInfoConfig{
		Enabled:        true,
		ExposeHeader:   true,
		ShowPerSource:  true,
		MergeStrategy:  "sum",
		ExpireStrategy: "earliest",
	}
}

func NormalizeSubscriptionInfoConfig(cfg *SubscriptionInfoConfig) *SubscriptionInfoConfig {
	defaults := DefaultSubscriptionInfoConfig()
	if cfg == nil {
		return defaults
	}

	out := *cfg
	if strings.TrimSpace(out.MergeStrategy) == "" {
		out.MergeStrategy = defaults.MergeStrategy
	}
	if strings.TrimSpace(out.ExpireStrategy) == "" {
		out.ExpireStrategy = defaults.ExpireStrategy
	}
	return &out
}

func NormalizeSubscriptionMeta(meta SubscriptionMeta) SubscriptionMeta {
	meta.SourceID = sanitizeText(meta.SourceID)
	meta.SourceName = sanitizeText(meta.SourceName)
	meta.FetchedAt = sanitizeText(meta.FetchedAt)

	if meta.Upload < 0 {
		meta.Upload = 0
	}
	if meta.Download < 0 {
		meta.Download = 0
	}
	if meta.Total < 0 {
		meta.Total = 0
	}
	if meta.Expire < 0 {
		meta.Expire = 0
	}
	if meta.Used < 0 {
		meta.Used = 0
	}
	if meta.Remaining < 0 {
		meta.Remaining = 0
	}
	if meta.Upload > 0 || meta.Download > 0 {
		meta.Used = meta.Upload + meta.Download
	}
	switch {
	case meta.Total > 0:
		if meta.Remaining == 0 && meta.Used > 0 {
			meta.Remaining = meta.Total - meta.Used
		}
		if meta.Used > 0 {
			meta.UsedRatio = float64(meta.Used) / float64(meta.Total)
		} else {
			meta.UsedRatio = 0
		}
	default:
		meta.UsedRatio = 0
	}

	return meta
}

func AggregateSubscriptionMeta(metas map[string]SubscriptionMeta, opts SubscriptionMetaAggregateOptions) AggregatedSubscriptionMeta {
	ordered := orderedSubscriptionMetas(metas, opts.SourceOrder)
	mergeStrategy := strings.ToLower(strings.TrimSpace(opts.MergeStrategy))
	if mergeStrategy == "" {
		mergeStrategy = "sum"
	}

	switch mergeStrategy {
	case "none":
		return AggregatedSubscriptionMeta{}
	case "first":
		for _, meta := range ordered {
			meta = NormalizeSubscriptionMeta(meta)
			aggregate := AggregatedSubscriptionMeta{
				Upload:    meta.Upload,
				Download:  meta.Download,
				Total:     meta.Total,
				Used:      meta.Used,
				Remaining: meta.Remaining,
				UsedRatio: meta.UsedRatio,
				Expire:    meta.Expire,
			}
			if meta.Expire > 0 {
				aggregate.ExpireSourceID = meta.SourceID
				aggregate.ExpireSourceName = meta.SourceName
			}
			return aggregate
		}
		return AggregatedSubscriptionMeta{}
	default:
		var aggregate AggregatedSubscriptionMeta
		for _, meta := range ordered {
			meta = NormalizeSubscriptionMeta(meta)
			aggregate.Upload += meta.Upload
			aggregate.Download += meta.Download
			if meta.Total > 0 {
				aggregate.Total += meta.Total
			}
		}
		aggregate.Used = aggregate.Upload + aggregate.Download
		if aggregate.Total > 0 {
			aggregate.Remaining = aggregate.Total - aggregate.Used
			aggregate.UsedRatio = float64(aggregate.Used) / float64(aggregate.Total)
		}
		expire, sourceID, sourceName := aggregateExpire(ordered, opts)
		aggregate.Expire = expire
		aggregate.ExpireSourceID = sourceID
		aggregate.ExpireSourceName = sourceName
		return aggregate
	}
}

func FormatSubscriptionUserinfoHeader(aggregate AggregatedSubscriptionMeta) string {
	parts := []string{
		fmt.Sprintf("upload=%d", aggregate.Upload),
		fmt.Sprintf("download=%d", aggregate.Download),
	}
	if aggregate.Total > 0 {
		parts = append(parts, fmt.Sprintf("total=%d", aggregate.Total))
	}
	if aggregate.Expire > 0 {
		parts = append(parts, fmt.Sprintf("expire=%d", aggregate.Expire))
	}
	return strings.Join(parts, "; ")
}

func orderedSubscriptionMetas(metas map[string]SubscriptionMeta, sourceOrder []string) []SubscriptionMeta {
	if len(metas) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(metas))
	out := make([]SubscriptionMeta, 0, len(metas))
	for _, sourceID := range sourceOrder {
		sourceID = sanitizeText(sourceID)
		if sourceID == "" {
			continue
		}
		meta, ok := metas[sourceID]
		if !ok {
			continue
		}
		seen[sourceID] = struct{}{}
		out = append(out, NormalizeSubscriptionMeta(meta))
	}

	extraIDs := make([]string, 0, len(metas))
	for sourceID := range metas {
		if _, ok := seen[sourceID]; ok {
			continue
		}
		extraIDs = append(extraIDs, sourceID)
	}
	sort.Strings(extraIDs)
	for _, sourceID := range extraIDs {
		out = append(out, NormalizeSubscriptionMeta(metas[sourceID]))
	}

	return out
}

func aggregateExpire(ordered []SubscriptionMeta, opts SubscriptionMetaAggregateOptions) (int64, string, string) {
	expireStrategy := strings.ToLower(strings.TrimSpace(opts.ExpireStrategy))
	if expireStrategy == "" {
		expireStrategy = "earliest"
	}

	switch expireStrategy {
	case "first":
		if len(ordered) == 0 {
			return 0, "", ""
		}
		first := NormalizeSubscriptionMeta(ordered[0])
		if first.Expire <= 0 {
			return 0, "", ""
		}
		return first.Expire, first.SourceID, first.SourceName
	case "latest":
		var (
			expire     int64
			sourceID   string
			sourceName string
		)
		for _, meta := range ordered {
			meta = NormalizeSubscriptionMeta(meta)
			if meta.Expire <= 0 {
				continue
			}
			if meta.Expire >= expire {
				expire = meta.Expire
				sourceID = meta.SourceID
				sourceName = meta.SourceName
			}
		}
		return expire, sourceID, sourceName
	default:
		var (
			expire     int64
			sourceID   string
			sourceName string
		)
		for _, meta := range ordered {
			meta = NormalizeSubscriptionMeta(meta)
			if meta.Expire <= 0 {
				continue
			}
			if expire == 0 || meta.Expire < expire {
				expire = meta.Expire
				sourceID = meta.SourceID
				sourceName = meta.SourceName
			}
		}
		return expire, sourceID, sourceName
	}
}
