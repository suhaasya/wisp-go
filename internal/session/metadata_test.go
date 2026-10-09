package session

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestTableDetailsLRUEvictsAtMax(t *testing.T) {
	lru := newTableDetailsLRU(3)
	put := func(schema, table string) {
		lru.put(tableDetailsKey{schema, table}, &driver.TableDetails{Schema: schema, Table: table})
	}
	put("s", "a")
	put("s", "b")
	put("s", "c")
	if lru.len() != 3 {
		t.Fatalf("len = %d, want 3", lru.len())
	}
	put("s", "d")
	if lru.len() != 3 {
		t.Fatalf("after evict len = %d, want 3", lru.len())
	}
	if _, ok := lru.get(tableDetailsKey{"s", "a"}); ok {
		t.Fatal("expected oldest entry evicted")
	}
	if _, ok := lru.get(tableDetailsKey{"s", "d"}); !ok {
		t.Fatal("expected newest entry present")
	}
}

func TestGuardDDLClearsMetadataCache(t *testing.T) {
	s := NewTestSession(storeProfile())
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	cache.objects[objectListKey{"public", driver.ObjectKindTable}] = []string{"t1"}
	cache.mu.Unlock()
	check, err := s.Guard("DROP TABLE t", false)
	if err != nil || check != WriteAllowed {
		t.Fatalf("guard: check=%v err=%v", check, err)
	}
	if lists, _ := s.MetadataCacheStats(); lists != 0 {
		t.Fatalf("expected cache cleared, lists=%d", lists)
	}
}

func TestMetadataCacheInvalidateClears(t *testing.T) {
	s := NewTestSession(storeProfile())
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	cache.objects[objectListKey{"public", driver.ObjectKindTable}] = []string{"t1"}
	cache.tables.put(tableDetailsKey{"public", "t1"}, &driver.TableDetails{})
	cache.mu.Unlock()
	s.InvalidateMetadata()
	lists, tables := s.MetadataCacheStats()
	if lists != 0 || tables != 0 {
		t.Fatalf("stats lists=%d tables=%d, want 0/0", lists, tables)
	}
}

func storeProfile() store.Profile {
	return store.Profile{Engine: "postgres"}
}
