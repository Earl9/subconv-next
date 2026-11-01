package pipeline

import (
	"container/list"
	"crypto/sha256"
	"sync"

	"subconv-next/internal/model"
	"subconv-next/internal/parser"
)

// Cache only source parsing, never workspace edits or final node lists. Content
// and all source metadata form the key, so refreshes and source changes take
// effect immediately while pagination/search can reuse unchanged input.
var sourceParseCache = newParseCache(32, 2*1024*1024, 4096)

type parseCacheKey struct {
	Digest [sha256.Size]byte
	Source model.SourceInfo
}

type parseCacheEntry struct {
	key         parseCacheKey
	result      parser.ParseResult
	sourceBytes int
}

type parseCache struct {
	mu                     sync.Mutex
	entries                map[parseCacheKey]*list.Element
	lru                    list.List
	maxEntries, maxBytes   int
	maxNodes, bytes, nodes int
}

func newParseCache(maxEntries, maxBytes, maxNodes int) *parseCache {
	return &parseCache{
		entries:    make(map[parseCacheKey]*list.Element),
		maxEntries: maxEntries, maxBytes: maxBytes, maxNodes: maxNodes,
	}
}

func (cache *parseCache) parse(content []byte, source model.SourceInfo) parser.ParseResult {
	if len(content) > cache.maxBytes || cache.maxEntries <= 0 {
		return parser.ParseContent(content, source)
	}
	key := parseCacheKey{Digest: sha256.Sum256(content), Source: source}
	cache.mu.Lock()
	if element := cache.entries[key]; element != nil {
		cache.lru.MoveToFront(element)
		result := element.Value.(parseCacheEntry).result
		cache.mu.Unlock()
		// Cache entries are immutable. Copy outside the lock, including nested
		// raw protocol fields, so edits cannot affect another request/session.
		return cloneParseResult(result)
	}
	cache.mu.Unlock()

	result := parser.ParseContent(content, source)
	if len(result.Errors) != 0 || len(result.Nodes) == 0 || len(result.Nodes) > cache.maxNodes {
		return result
	}
	entry := parseCacheEntry{key: key, result: cloneParseResult(result), sourceBytes: len(content)}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if element := cache.entries[key]; element != nil {
		cache.lru.MoveToFront(element)
		return result
	}
	for cache.lru.Len() >= cache.maxEntries || cache.bytes+entry.sourceBytes > cache.maxBytes || cache.nodes+len(entry.result.Nodes) > cache.maxNodes {
		oldest := cache.lru.Back()
		old := oldest.Value.(parseCacheEntry)
		delete(cache.entries, old.key)
		cache.bytes -= old.sourceBytes
		cache.nodes -= len(old.result.Nodes)
		cache.lru.Remove(oldest)
	}
	cache.entries[key] = cache.lru.PushFront(entry)
	cache.bytes += entry.sourceBytes
	cache.nodes += len(entry.result.Nodes)
	return result
}

func cloneParseResult(result parser.ParseResult) parser.ParseResult {
	result.Nodes = model.CloneNodes(result.Nodes)
	result.Warnings = append([]string(nil), result.Warnings...)
	result.Errors = append([]parser.ParseError(nil), result.Errors...)
	return result
}
