package pipeline

import (
	"fmt"
	"sync"
	"testing"

	"subconv-next/internal/model"
)

func TestParseCacheIsolatesContentSourcesAndMutableNodes(t *testing.T) {
	cache := newParseCache(4, 8192, 10)
	source := model.SourceInfo{ID: "source", Name: "original source", Kind: "subscription", URLHash: "url"}
	content := []byte("proxies:\n  - {name: original, type: vless, server: one.example, port: 443, uuid: test, udp: true, network: ws, ws-opts: {path: /original, headers: {Host: original.example}}}\n")
	first := cache.parse(content, source)
	if len(first.Nodes) != 1 || len(first.Errors) != 0 {
		t.Fatalf("parse failed: %v", first.Errors)
	}
	first.Nodes[0].Name = "changed by first caller"
	*first.Nodes[0].UDP = false
	first.Nodes[0].Raw["_mihomoProxyFields"].(map[string]interface{})["ws-opts"].(map[string]interface{})["path"] = "/changed"
	second := cache.parse(content, source)
	if second.Nodes[0].Name != "original" || !(*second.Nodes[0].UDP) {
		t.Fatal("a caller modified the cached node")
	}
	if second.Nodes[0].Raw["_mihomoProxyFields"].(map[string]interface{})["ws-opts"].(map[string]interface{})["path"] != "/original" {
		t.Fatal("nested extension fields were shared with the cache")
	}
	second.Nodes[0].Name = "changed cache hit"
	if cache.parse(content, source).Nodes[0].Name != "original" {
		t.Fatal("a cache hit modified the cached node")
	}
	other := source
	other.ID, other.Name, other.URLHash = "other", "other source", "other-url"
	if node := cache.parse(content, other).Nodes[0]; node.Source != other {
		t.Fatal("cache reused another source's metadata")
	}
	updated := []byte("trojan://test@two.example:443#updated")
	if node := cache.parse(updated, source).Nodes[0]; node.Name != "updated" || node.Server != "two.example" {
		t.Fatal("changed upstream content returned stale nodes")
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := cache.parse(content, source)
			*result.Nodes[0].UDP = false
		}()
	}
	wg.Wait()
	if !(*cache.parse(content, source).Nodes[0].UDP) {
		t.Fatal("concurrent callers shared cache state")
	}
}

func TestParseCacheBoundsAndEviction(t *testing.T) {
	for _, limits := range [][3]int{{2, 8192, 10}, {10, 100, 10}, {10, 8192, 2}} {
		cache := newParseCache(limits[0], limits[1], limits[2])
		for i := 0; i < 8; i++ {
			content := []byte(fmt.Sprintf("trojan://test@node-%d.example:443#node-%d", i, i))
			cache.parse(content, model.SourceInfo{})
			if cache.lru.Len() > limits[0] || cache.bytes > limits[1] || cache.nodes > limits[2] {
				t.Fatal("parse cache exceeded its entry/input/node budget")
			}
		}
		if cache.lru.Len() >= 8 {
			t.Fatal("cache failed to evict old entries")
		}
		before := cache.lru.Len()
		cache.parse([]byte("not a subscription"), model.SourceInfo{})
		if cache.lru.Len() != before {
			t.Fatal("invalid content was cached")
		}
	}
}
