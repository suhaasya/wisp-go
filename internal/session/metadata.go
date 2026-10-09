package session

import (
	"context"
	"sync"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
)

const maxTableDetailsCache = 2000

type objectListKey struct {
	schema string
	kind   driver.ObjectKind
}

type metadataCache struct {
	mu      sync.Mutex
	objects map[objectListKey][]string
	tables  tableDetailsLRU
}

func newMetadataCache() *metadataCache {
	return &metadataCache{
		objects: make(map[objectListKey][]string),
		tables:  newTableDetailsLRU(maxTableDetailsCache),
	}
}

func (c *metadataCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.objects = make(map[objectListKey][]string)
	c.tables.clear()
}

func (c *metadataCache) objectListLen() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.objects)
}

func (c *metadataCache) tableDetailsLen() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tables.len()
}

type tableDetailsLRU struct {
	max   int
	order []tableDetailsKey
	items map[tableDetailsKey]*driver.TableDetails
}

type tableDetailsKey struct {
	schema string
	table  string
}

func newTableDetailsLRU(max int) tableDetailsLRU {
	return tableDetailsLRU{
		max:   max,
		items: make(map[tableDetailsKey]*driver.TableDetails),
	}
}

func (l *tableDetailsLRU) get(k tableDetailsKey) (*driver.TableDetails, bool) {
	v, ok := l.items[k]
	if !ok {
		return nil, false
	}
	for i, key := range l.order {
		if key == k {
			copy(l.order[i:], l.order[i+1:])
			l.order = l.order[:len(l.order)-1]
			break
		}
	}
	l.order = append(l.order, k)
	return v, true
}

func (l *tableDetailsLRU) put(k tableDetailsKey, v *driver.TableDetails) {
	if _, ok := l.items[k]; ok {
		l.get(k)
	} else if len(l.order) >= l.max && l.max > 0 {
		evict := l.order[0]
		l.order = l.order[1:]
		delete(l.items, evict)
	}
	l.items[k] = v
	l.order = append(l.order, k)
}

func (l *tableDetailsLRU) clear() {
	l.order = nil
	l.items = make(map[tableDetailsKey]*driver.TableDetails)
}

func (l *tableDetailsLRU) len() int {
	return len(l.items)
}

// InvalidateMetadata clears cached schema lists and table details.
func (s *Session) InvalidateMetadata() {
	if s == nil {
		return
	}
	s.mu.Lock()
	cache := s.metaCache
	s.mu.Unlock()
	if cache != nil {
		cache.invalidateAll()
	}
	if s.onMetadataBump != nil {
		s.onMetadataBump(s.id)
	}
}

// SeedTableMetadataForTest inserts cached table details (tests only).
func (s *Session) SeedTableMetadataForTest(schema, table string, det *driver.TableDetails) {
	if s == nil || det == nil {
		return
	}
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	cache.tables.put(tableDetailsKey{schema: schema, table: table}, det)
	cache.mu.Unlock()
}

// MetadataCacheStats returns cache sizes (for tests).
func (s *Session) MetadataCacheStats() (objectLists, tableDetails int) {
	s.mu.RLock()
	cache := s.metaCache
	s.mu.RUnlock()
	if cache == nil {
		return 0, 0
	}
	return cache.objectListLen(), cache.tableDetailsLen()
}

func (s *Session) ensureMetaCache() *metadataCache {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metaCache == nil {
		s.metaCache = newMetadataCache()
	}
	return s.metaCache
}

func introspectorFor(engine string, conn driver.Conn) driver.Introspector {
	switch engine {
	case "postgres":
		return &postgres.Introspector{Conn: conn}
	case "mysql", "mariadb":
		return &mysql.Introspector{Conn: conn}
	default:
		return nil
	}
}

// SchemaSnapshot lists databases and schemas for the connected engine.
func (s *Session) SchemaSnapshot(ctx context.Context) (databases, schemas []string, err error) {
	engine := s.Profile().Engine
	err = s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		intr := introspectorFor(engine, conn)
		if intr == nil {
			return errBadEngine
		}
		switch engine {
		case "postgres":
			schemas, err = intr.Schemas(ctx)
			if err != nil {
				return err
			}
			databases, err = intr.Databases(ctx)
			if err != nil {
				databases = nil
				err = nil
			}
			return nil
		default:
			databases, err = intr.Databases(ctx)
			return err
		}
	})
	return databases, schemas, err
}

// ListObjects returns cached object names for a schema and kind.
func (s *Session) ListObjects(ctx context.Context, schema string, kind driver.ObjectKind, refresh bool) ([]string, error) {
	if refresh {
		s.InvalidateMetadata()
	}
	cache := s.ensureMetaCache()
	key := objectListKey{schema: schema, kind: kind}
	cache.mu.Lock()
	if names, ok := cache.objects[key]; ok && !refresh {
		out := append([]string(nil), names...)
		cache.mu.Unlock()
		return out, nil
	}
	cache.mu.Unlock()

	var names []string
	err := s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		intr := introspectorFor(s.Profile().Engine, conn)
		if intr == nil {
			return errBadEngine
		}
		var err error
		names, err = intr.Objects(ctx, schema, kind)
		return err
	})
	if err != nil {
		return nil, err
	}
	cache.mu.Lock()
	cache.objects[key] = append([]string(nil), names...)
	out := append([]string(nil), names...)
	cache.mu.Unlock()
	return out, nil
}

// TableMetadata returns table details, using the LRU cache unless refresh is set.
func (s *Session) TableMetadata(ctx context.Context, schema, table string, refresh bool) (*driver.TableDetails, error) {
	if refresh {
		s.InvalidateMetadata()
	}
	cache := s.ensureMetaCache()
	tkey := tableDetailsKey{schema: schema, table: table}
	cache.mu.Lock()
	if det, ok := cache.tables.get(tkey); ok && !refresh {
		cache.mu.Unlock()
		return det, nil
	}
	cache.mu.Unlock()

	var det *driver.TableDetails
	err := s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		intr := introspectorFor(s.Profile().Engine, conn)
		if intr == nil {
			return errBadEngine
		}
		var err error
		det, err = intr.TableDetails(ctx, schema, table)
		return err
	})
	if err != nil {
		return nil, err
	}
	cache.mu.Lock()
	cache.tables.put(tkey, det)
	cache.mu.Unlock()
	return det, nil
}
