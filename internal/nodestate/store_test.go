package nodestate

import (
	"path/filepath"
	"testing"

	"subconv-next/internal/model"
)

func TestSaveLoadPreservesPerSourceSubscriptionMeta(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	state := model.NodeState{
		NodeOverrides: map[string]model.NodeOverride{},
		DisabledNodes: []string{},
		CustomNodes:   []model.NodeIR{},
		SubscriptionMeta: map[string]model.SubscriptionMeta{
			"source-1": {
				SourceID:   "source-1",
				SourceName: "主力机场",
				Total:      200 * 1024 * 1024 * 1024,
				Used:       30 * 1024 * 1024 * 1024,
				Expire:     1779235200,
				FromHeader: true,
			},
			"source-2": {
				SourceID:     "source-2",
				SourceName:   "备用机场",
				Total:        100 * 1024 * 1024 * 1024,
				Used:         10 * 1024 * 1024 * 1024,
				Expire:       1780185600,
				FromInfoNode: true,
			},
		},
	}

	if err := Save(path, state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.SubscriptionMeta) != 2 {
		t.Fatalf("len(loaded.SubscriptionMeta) = %d, want 2", len(loaded.SubscriptionMeta))
	}
	if loaded.SubscriptionMeta["source-1"].SourceName != "主力机场" || loaded.SubscriptionMeta["source-2"].SourceName != "备用机场" {
		t.Fatalf("loaded.SubscriptionMeta = %#v, want source metadata preserved", loaded.SubscriptionMeta)
	}
	if loaded.SubscriptionMeta["source-1"].Expire == loaded.SubscriptionMeta["source-2"].Expire {
		t.Fatalf("loaded metas should remain independent: %#v", loaded.SubscriptionMeta)
	}
}
