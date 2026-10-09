package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestTableDDLPostgreSQLFromCachedMetadata(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	cache := s.ensureMetaCache()
	det := &driver.TableDetails{
		Schema: "wisp_meta",
		Table:  "parent",
		Columns: []driver.ColumnDetail{
			{Name: "id", Type: "bigint", Nullable: false},
		},
		PrimaryKey: []string{"id"},
	}
	cache.mu.Lock()
	cache.tables.put(tableDetailsKey{"wisp_meta", "parent"}, det)
	cache.mu.Unlock()

	ddl, err := s.TableDDL(context.Background(), "wisp_meta", "parent", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ddl, "CREATE TABLE") || !strings.Contains(ddl, "parent") {
		t.Fatalf("unexpected ddl: %s", ddl)
	}
}

func TestTableMetadataCachedSecondCallUnder200ms(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	cache := s.ensureMetaCache()
	det := &driver.TableDetails{Schema: "public", Table: "t1", Columns: []driver.ColumnDetail{{Name: "id", Type: "int"}}}
	cache.mu.Lock()
	cache.tables.put(tableDetailsKey{"public", "t1"}, det)
	cache.mu.Unlock()

	ctx := context.Background()
	start := time.Now()
	if _, err := s.TableMetadata(ctx, "public", "t1", false); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("cached TableMetadata took %v, want < 200ms", elapsed)
	}
}
