package pipeline

import (
	"testing"
	"time"

	"subconv-next/internal/fetcher"
	"subconv-next/internal/model"
)

func TestInfoNodeBelongsToSource(t *testing.T) {
	sourceA := model.SourceInfo{ID: "source-1", Name: "主力机场", Kind: "subscription"}
	sourceB := model.SourceInfo{ID: "source-2", Name: "备用机场", Kind: "subscription"}
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{Name: "剩余流量: 170 GB", Type: model.ProtocolSS, Source: sourceA}),
		model.NormalizeNode(model.NodeIR{Name: "总流量: 200 GB", Type: model.ProtocolSS, Source: sourceA}),
		model.NormalizeNode(model.NodeIR{Name: "普通节点", Type: model.ProtocolSS, Source: sourceB}),
	}

	metaA := parseInfoNodeMeta(sourceA, nodes, nil, time.Unix(1770000000, 0).UTC())
	metaB := parseInfoNodeMeta(sourceB, nodes, nil, time.Unix(1770000000, 0).UTC())

	if !metaA.hasTotal || !metaA.hasRemaining || metaA.meta.Total != 200*1024*1024*1024 {
		t.Fatalf("metaA = %#v, want source-1 info extracted", metaA)
	}
	if metaB.hasTotal || metaB.hasRemaining || metaB.meta.Total != 0 {
		t.Fatalf("metaB = %#v, source-2 should not inherit source-1 info", metaB)
	}
}

func TestHeaderPriorityOverInfoNode(t *testing.T) {
	source := model.SourceInfo{ID: "source-1", Name: "主力机场", Kind: "subscription"}
	fetched := fetcher.FetchedSubscription{
		SubscriptionUserinfo: "upload=0; download=32212254720; total=214748364800; expire=1779235200",
		FetchedAt:            time.Unix(1770000000, 0).UTC(),
	}
	nodes := []model.NodeIR{
		model.NormalizeNode(model.NodeIR{Name: "总流量: 100 GB", Type: model.ProtocolSS, Source: source}),
		model.NormalizeNode(model.NodeIR{Name: "剩余流量: 90 GB", Type: model.ProtocolSS, Source: source}),
		model.NormalizeNode(model.NodeIR{Name: "到期时间: 2026-06-01", Type: model.ProtocolSS, Source: source}),
	}

	meta, ok := buildSourceSubscriptionMeta(source, fetched, nodes)
	if !ok {
		t.Fatalf("buildSourceSubscriptionMeta() ok = false, want true")
	}
	if meta.Total != 214748364800 {
		t.Fatalf("meta.Total = %d, want header total", meta.Total)
	}
	if meta.Expire != 1779235200 {
		t.Fatalf("meta.Expire = %d, want header expire", meta.Expire)
	}
	if meta.Used != 32212254720 {
		t.Fatalf("meta.Used = %d, want header used from download", meta.Used)
	}
	if !meta.FromHeader || !meta.FromInfoNode {
		t.Fatalf("meta = %#v, want both header and info node markers", meta)
	}
}

func TestParseSubscriptionUserinfoHeader(t *testing.T) {
	source := model.SourceInfo{ID: "source-1", Name: "SecOne", Kind: "subscription"}
	fetchedAt := time.Unix(1770000000, 0).UTC()

	parsed := parseSubscriptionUserinfoHeader(
		"upload=1073741824; download=2147483648; total=10737418240; expire=1779235200",
		source,
		fetchedAt,
	)
	meta := model.NormalizeSubscriptionMeta(parsed.meta)

	if !parsed.hasUpload || !parsed.hasDownload || !parsed.hasTotal || !parsed.hasExpire {
		t.Fatalf("parsed flags = %#v, want upload/download/total/expire", parsed)
	}
	if meta.SourceID != "source-1" || meta.SourceName != "SecOne" || !meta.FromHeader {
		t.Fatalf("meta source/header = %#v", meta)
	}
	if meta.Upload != 1073741824 || meta.Download != 2147483648 {
		t.Fatalf("meta upload/download = %#v", meta)
	}
	if meta.Used != 3221225472 || meta.Total != 10737418240 || meta.Remaining != 7516192768 {
		t.Fatalf("meta usage = %#v", meta)
	}
	if meta.Expire != 1779235200 {
		t.Fatalf("meta.Expire = %d, want 1779235200", meta.Expire)
	}
	if meta.FetchedAt != fetchedAt.Format(time.RFC3339) {
		t.Fatalf("meta.FetchedAt = %q, want %q", meta.FetchedAt, fetchedAt.Format(time.RFC3339))
	}
}
