package model

import "testing"

func TestAggregateSubscriptionMetaSum(t *testing.T) {
	aggregate := AggregateSubscriptionMeta(map[string]SubscriptionMeta{
		"source-1": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID:   "source-1",
			SourceName: "主力机场",
			Download:   30 * 1024 * 1024 * 1024,
			Total:      200 * 1024 * 1024 * 1024,
			Expire:     1779235200,
		}),
		"source-2": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID:   "source-2",
			SourceName: "备用机场",
			Download:   10 * 1024 * 1024 * 1024,
			Total:      100 * 1024 * 1024 * 1024,
			Expire:     1780185600,
		}),
	}, SubscriptionMetaAggregateOptions{
		MergeStrategy:  "sum",
		ExpireStrategy: "earliest",
		SourceOrder:    []string{"source-1", "source-2"},
	})

	if aggregate.Total != 300*1024*1024*1024 {
		t.Fatalf("aggregate.Total = %d, want %d", aggregate.Total, int64(300*1024*1024*1024))
	}
	if aggregate.Used != 40*1024*1024*1024 {
		t.Fatalf("aggregate.Used = %d, want %d", aggregate.Used, int64(40*1024*1024*1024))
	}
	if aggregate.Remaining != 260*1024*1024*1024 {
		t.Fatalf("aggregate.Remaining = %d, want %d", aggregate.Remaining, int64(260*1024*1024*1024))
	}
	if aggregate.Expire != 1779235200 || aggregate.ExpireSourceID != "source-1" || aggregate.ExpireSourceName != "主力机场" {
		t.Fatalf("aggregate expire = %#v, want earliest from source-1", aggregate)
	}
}

func TestAggregateSubscriptionMetaMissingTotalAndExpire(t *testing.T) {
	aggregate := AggregateSubscriptionMeta(map[string]SubscriptionMeta{
		"source-1": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID:   "source-1",
			SourceName: "主力机场",
			Download:   30 * 1024 * 1024 * 1024,
		}),
		"source-2": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID:   "source-2",
			SourceName: "备用机场",
			Download:   10 * 1024 * 1024 * 1024,
			Total:      100 * 1024 * 1024 * 1024,
			Expire:     1780185600,
		}),
	}, SubscriptionMetaAggregateOptions{
		MergeStrategy:  "sum",
		ExpireStrategy: "latest",
		SourceOrder:    []string{"source-1", "source-2"},
	})

	if aggregate.Total != 100*1024*1024*1024 {
		t.Fatalf("aggregate.Total = %d, want only source-2 total", aggregate.Total)
	}
	if aggregate.Used != 40*1024*1024*1024 {
		t.Fatalf("aggregate.Used = %d, want used sum", aggregate.Used)
	}
	if aggregate.Expire != 1780185600 || aggregate.ExpireSourceID != "source-2" {
		t.Fatalf("aggregate expire = %#v, want latest from source-2", aggregate)
	}
}

func TestAggregateSubscriptionMetaFirst(t *testing.T) {
	aggregate := AggregateSubscriptionMeta(map[string]SubscriptionMeta{
		"source-1": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID:   "source-1",
			SourceName: "主力机场",
			Download:   30 * 1024 * 1024 * 1024,
			Total:      200 * 1024 * 1024 * 1024,
			Expire:     1779235200,
		}),
		"source-2": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID:   "source-2",
			SourceName: "备用机场",
			Download:   10 * 1024 * 1024 * 1024,
			Total:      100 * 1024 * 1024 * 1024,
			Expire:     1780185600,
		}),
	}, SubscriptionMetaAggregateOptions{
		MergeStrategy:  "first",
		ExpireStrategy: "latest",
		SourceOrder:    []string{"source-1", "source-2"},
	})

	if aggregate.Total != 200*1024*1024*1024 || aggregate.Used != 30*1024*1024*1024 {
		t.Fatalf("aggregate = %#v, want first source only", aggregate)
	}
	if aggregate.Expire != 1779235200 || aggregate.ExpireSourceID != "source-1" {
		t.Fatalf("aggregate expire = %#v, want first source expire", aggregate)
	}
}

func TestAggregateSubscriptionMetaNone(t *testing.T) {
	aggregate := AggregateSubscriptionMeta(map[string]SubscriptionMeta{
		"source-1": NormalizeSubscriptionMeta(SubscriptionMeta{
			SourceID: "source-1",
			Total:    100,
			Download: 10,
		}),
	}, SubscriptionMetaAggregateOptions{
		MergeStrategy:  "none",
		ExpireStrategy: "earliest",
		SourceOrder:    []string{"source-1"},
	})

	if aggregate.Total != 0 || aggregate.Used != 0 || aggregate.Expire != 0 {
		t.Fatalf("aggregate = %#v, want zero aggregate for merge_strategy=none", aggregate)
	}
}
