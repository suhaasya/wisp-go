package session

import (
	"context"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestStructureReviewSQLFromStagedOps(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	s.SeedTableMetadataForTest("wisp_meta", "parent", &driver.TableDetails{
		Schema: "wisp_meta", Table: "parent",
		Columns: []driver.ColumnDetail{{Name: "id", Type: "bigint", Nullable: false}},
	})
	cfg := mock.Config{
		TotalRows: 0,
		Columns:   []driver.ColumnMeta{{Name: "x", Type: "int8"}},
	}
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	s.AdoptMainConn(conn)

	_, err = s.ApplyStructureChange(context.Background(), "wisp_meta", "parent", sqlgen.StructureOp{
		Kind:      sqlgen.StructureOpAddColumn,
		ColumnDef: sqlgen.ColumnAlterSpec{Name: "notes", Type: "text", Nullable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	review, err := s.StructureReviewSQL(context.Background(), "wisp_meta", "parent")
	if err != nil {
		t.Fatal(err)
	}
	if review.Count != 1 || !strings.Contains(review.SQL, "ADD COLUMN") {
		t.Fatalf("review=%+v", review)
	}
}

func TestStructureReviewSQLMySQLAutoCommitWarning(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "mysql"})
	s.SeedTableMetadataForTest("wisp_meta", "parent", &driver.TableDetails{
		Schema: "wisp_meta", Table: "parent",
		Columns: []driver.ColumnDetail{{Name: "id", Type: "bigint"}},
	})
	conn, _ := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: 0, Columns: []driver.ColumnMeta{{Name: "x", Type: "int8"}}},
	})
	s.AdoptMainConn(conn)
	_, _ = s.ApplyStructureChange(context.Background(), "wisp_meta", "parent", sqlgen.StructureOp{
		Kind:      sqlgen.StructureOpAddColumn,
		ColumnDef: sqlgen.ColumnAlterSpec{Name: "x", Type: "int", Nullable: true},
	})
	review, err := s.StructureReviewSQL(context.Background(), "wisp_meta", "parent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(review.Warning, "auto-commits") {
		t.Fatalf("warning=%q", review.Warning)
	}
}

func TestStructureDestructiveRequiresTableConfirmOnProduction(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres", Env: "production"})
	s.SeedTableMetadataForTest("public", "orders", &driver.TableDetails{
		Schema: "public", Table: "orders",
		Columns: []driver.ColumnDetail{{Name: "id", Type: "bigint"}},
	})
	conn, _ := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: 0, Columns: []driver.ColumnMeta{{Name: "x", Type: "int8"}}},
	})
	s.AdoptMainConn(conn)
	_, _ = s.ApplyStructureChange(context.Background(), "public", "orders", sqlgen.StructureOp{
		Kind: sqlgen.StructureOpDropColumn, Column: "legacy",
	})
	review, err := s.StructureReviewSQL(context.Background(), "public", "orders")
	if err != nil {
		t.Fatal(err)
	}
	if !review.NeedsDeleteConfirm {
		t.Fatal("expected destructive confirm")
	}
	_, err = s.CommitStructureChanges(context.Background(), "public", "orders", CommitOptions{Confirm: true})
	if err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("commit err=%v", err)
	}
}
