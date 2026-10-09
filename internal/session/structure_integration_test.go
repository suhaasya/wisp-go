//go:build integration

package session

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestStructureChangeRoundTripPostgres(t *testing.T) {
	conn := openPGRun(t)
	defer func() { _ = conn.Close() }()

	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.mu.Lock()
	s.main = conn
	s.meta = conn
	s.mu.Unlock()

	schema, table := "wisp_meta", "parent"
	_, err := s.ApplyStructureChange(t.Context(), schema, table, sqlgen.StructureOp{
		Kind: sqlgen.StructureOpAddColumn,
		ColumnDef: sqlgen.ColumnAlterSpec{
			Name: "notes", Type: "text", Nullable: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CommitStructureChanges(t.Context(), schema, table, CommitOptions{Confirm: true})
	if err != nil {
		t.Fatal(err)
	}
	intr := &postgres.Introspector{Conn: conn}
	det, err := intr.TableDetails(t.Context(), schema, table)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range det.Columns {
		if c.Name == "notes" {
			found = true
			if !c.Nullable {
				t.Fatal("notes should be nullable")
			}
		}
	}
	if !found {
		t.Fatal("notes column missing after alter")
	}
	_, _ = conn.Exec(t.Context(), `ALTER TABLE wisp_meta.parent DROP COLUMN IF EXISTS notes`)
}
