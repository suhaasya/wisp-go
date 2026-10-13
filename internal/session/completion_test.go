package session

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestCompletionSchemaEmptyTablesJSON(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	cache.objects[objectListKey{"public", driver.ObjectKindTable}] = []string{}
	cache.objects[objectListKey{"public", driver.ObjectKindView}] = []string{}
	cache.mu.Unlock()

	out, err := s.CompletionSchema(context.Background(), "public", nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) {
		t.Fatalf("invalid json: %s", raw)
	}
	var decoded struct {
		Tables json.RawMessage `json:"tables"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if string(decoded.Tables) != "[]" {
		t.Fatalf("tables json=%s want []", decoded.Tables)
	}
}

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
	if out.Tables[0].Name != "orders" || len(out.Tables[0].Columns) != 0 {
		t.Fatalf("bootstrap table=%+v want name only", out.Tables[0])
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
	if len(out.Tables) != 1 || !names["items"] {
		t.Fatalf("tables=%+v want items only", out.Tables)
	}
	if len(out.Tables[0].Columns) != 1 {
		t.Fatalf("items columns=%+v", out.Tables[0].Columns)
	}
}

func TestCompletionSchemaBootstrapManyTablesNamesOnly(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	cache := s.ensureMetaCache()
	cache.mu.Lock()
	tableNames := make([]string, 120)
	for i := range tableNames {
		tableNames[i] = "tbl_" + strconv.Itoa(i)
	}
	cache.objects[objectListKey{"public", driver.ObjectKindTable}] = tableNames
	cache.objects[objectListKey{"public", driver.ObjectKindView}] = []string{}
	cache.mu.Unlock()

	out, err := s.CompletionSchema(context.Background(), "public", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Tables) != 120 {
		t.Fatalf("tables=%d want 120", len(out.Tables))
	}
	for _, tbl := range out.Tables {
		if len(tbl.Columns) != 0 {
			t.Fatalf("%s has columns on bootstrap", tbl.Name)
		}
	}
}
