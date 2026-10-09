package session

import (
	"context"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestCompletionSchemaFromCache(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	cache.objects[objectListKey{"public", driver.ObjectKindTable}] = []string{"orders"}
	cache.objects[objectListKey{"public", driver.ObjectKindView}] = []string{}
	cache.tables.put(tableDetailsKey{"public", "orders"}, &driver.TableDetails{
		Schema: "public",
		Table:  "orders",
		Columns: []driver.ColumnDetail{
			{Name: "id", Type: "int4", OrdinalPos: 1},
			{Name: "status", Type: "text", OrdinalPos: 2},
		},
	})
	cache.mu.Unlock()

	out, err := s.CompletionSchema(context.Background(), "public", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Schema != "public" || len(out.Tables) != 1 {
		t.Fatalf("out=%+v", out)
	}
	if out.Tables[0].Name != "orders" || len(out.Tables[0].Columns) != 2 {
		t.Fatalf("table=%+v", out.Tables[0])
	}
}

func TestCompletionSchemaExtraTables(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	cache.objects[objectListKey{"public", driver.ObjectKindTable}] = []string{"orders"}
	cache.objects[objectListKey{"public", driver.ObjectKindView}] = []string{}
	cache.tables.put(tableDetailsKey{"public", "orders"}, &driver.TableDetails{
		Schema: "public", Table: "orders",
		Columns: []driver.ColumnDetail{{Name: "id", Type: "int4", OrdinalPos: 1}},
	})
	cache.tables.put(tableDetailsKey{"public", "items"}, &driver.TableDetails{
		Schema: "public", Table: "items",
		Columns: []driver.ColumnDetail{{Name: "sku", Type: "text", OrdinalPos: 1}},
	})
	cache.mu.Unlock()

	out, err := s.CompletionSchema(context.Background(), "public", []string{"items"})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tbl := range out.Tables {
		names[tbl.Name] = true
	}
	if !names["orders"] || !names["items"] {
		t.Fatalf("tables=%+v", out.Tables)
	}
}
