package sqlgen_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func strPtr(s string) *string { return &s }

func TestBuildCommitStatementsUpdateChangedColumnsOnly(t *testing.T) {
	d := postgres.Dialect()
	plan := sqlgen.CommitPlan{
		Updates: []sqlgen.RowUpdate{{
			KeyColumns: []string{"id"},
			KeyValues:  []any{int64(42)},
			Cells: map[string]sqlgen.CellChange{
				"status": {
					Value:    strPtr("shipped"),
					Original: strPtr("pending"),
				},
			},
		}},
	}
	stmts, err := sqlgen.BuildCommitStatements(d, sqlgen.CommitSpec{
		Schema: "public", Table: "orders", Optimistic: true,
	}, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(stmts) != 1 {
		t.Fatalf("stmts=%d want 1", len(stmts))
	}
	wantSQL := `UPDATE "public"."orders" SET "status" = $1 WHERE "id" = $2 AND "status" = $3`
	if stmts[0].SQL != wantSQL {
		t.Fatalf("sql:\n got %q\nwant %q", stmts[0].SQL, wantSQL)
	}
	wantArgs := []any{"shipped", int64(42), "pending"}
	if len(stmts[0].Args) != len(wantArgs) {
		t.Fatalf("args=%v want %v", stmts[0].Args, wantArgs)
	}
	for i := range wantArgs {
		if stmts[0].Args[i] != wantArgs[i] {
			t.Fatalf("args[%d]=%v want %v", i, stmts[0].Args[i], wantArgs[i])
		}
	}
}

func TestDisplaySQLEqualsManualLiteralSubstitution(t *testing.T) {
	d := postgres.Dialect()
	sqlText := `UPDATE "public"."orders" SET "status" = $1 WHERE "id" = $2 AND "status" = $3`
	args := []any{"shipped", int64(42), "pending"}
	got := sqlgen.DisplaySQL(d, sqlText, args)
	want := `UPDATE "public"."orders" SET "status" = 'shipped' WHERE "id" = 42 AND "status" = 'pending'`
	if got != want {
		t.Fatalf("display:\n got %q\nwant %q", got, want)
	}
}
