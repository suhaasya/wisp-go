package sqlgen_test

import (
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestBuildStructureAlterPostgreSQLAddColumn(t *testing.T) {
	assertPGAlter(t, sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind: sqlgen.StructureOpAddColumn,
			ColumnDef: sqlgen.ColumnAlterSpec{
				Name: "notes", Type: "text", Nullable: true,
			},
		}},
	}, `ALTER TABLE "wisp_meta"."parent" ADD COLUMN "notes" text`)
}

func TestBuildStructureAlterPostgreSQLAlterTypeUsing(t *testing.T) {
	assertPGAlter(t, sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind:   sqlgen.StructureOpAlterColumnType,
			Column: "status",
			ColumnDef: sqlgen.ColumnAlterSpec{
				Name: "status", Type: "varchar(20)",
			},
			TypeUsing: `"status"::varchar(20)`,
		}},
	}, `ALTER TABLE "wisp_meta"."parent" ALTER COLUMN "status" TYPE varchar(20) USING "status"::varchar(20)`)
}

func TestBuildStructureAlterPostgreSQLSetNotNull(t *testing.T) {
	assertPGAlter(t, sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind:      sqlgen.StructureOpAlterColumnNullable,
			Column:    "status",
			ColumnDef: sqlgen.ColumnAlterSpec{Name: "status", Nullable: false},
		}},
	}, `ALTER TABLE "wisp_meta"."parent" ALTER COLUMN "status" SET NOT NULL`)
}

func TestBuildStructureAlterPostgreSQLCreateIndexConcurrentlyOutsideTx(t *testing.T) {
	d := postgres.Dialect()
	batch, err := sqlgen.BuildStructureAlterStatements(d, pgSpec(), parentFixtureDetails(), sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind:            sqlgen.StructureOpAddIndex,
			IndexName:       "idx_parent_status",
			IndexColumns:    []string{"status"},
			IndexConcurrent: true,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.InTransaction) != 0 {
		t.Fatalf("in tx=%d want 0", len(batch.InTransaction))
	}
	if len(batch.AfterCommit) != 1 {
		t.Fatalf("after=%d want 1", len(batch.AfterCommit))
	}
	want := `CREATE INDEX CONCURRENTLY "idx_parent_status" ON "wisp_meta"."parent" ("status")`
	got := sqlgen.DisplaySQL(d, batch.AfterCommit[0].SQL, batch.AfterCommit[0].Args)
	if normalizeDDL(got) != normalizeDDL(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func assertPGAlter(t *testing.T, plan sqlgen.StructureAlterPlan, want string) {
	t.Helper()
	d := postgres.Dialect()
	batch, err := sqlgen.BuildStructureAlterStatements(d, pgSpec(), parentFixtureDetails(), plan)
	if err != nil {
		t.Fatal(err)
	}
	got := displayBatch(d, batch)
	if normalizeDDL(got) != normalizeDDL(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func pgSpec() sqlgen.StructureSpec {
	return sqlgen.StructureSpec{Schema: "wisp_meta", Table: "parent", Engine: "postgres"}
}

func displayBatch(d driver.Dialect, batch sqlgen.StructureAlterBatch) string {
	var parts []string
	for _, st := range batch.InTransaction {
		parts = append(parts, sqlgen.DisplaySQL(d, st.SQL, st.Args))
	}
	for _, st := range batch.AfterCommit {
		parts = append(parts, sqlgen.DisplaySQL(d, st.SQL, st.Args))
	}
	return strings.Join(parts, "\n\n")
}
