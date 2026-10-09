package sqlgen_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestBuildStructureAlterMySQLModifyColumnFullDefinition(t *testing.T) {
	d := mysql.Dialect("mysql")
	baseline := parentFixtureDetails()
	baseline.Columns = append(baseline.Columns, driver.ColumnDetail{
		Name: "total", Type: "decimal(10,2)", Nullable: false, Default: "0.00", OrdinalPos: 3,
	})
	batch, err := sqlgen.BuildStructureAlterStatements(d, sqlgen.StructureSpec{
		Schema: "wisp_meta", Table: "parent", Engine: "mysql", MySQLMajor: 8,
	}, baseline, sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind:   sqlgen.StructureOpAlterColumnType,
			Column: "total",
			ColumnDef: sqlgen.ColumnAlterSpec{
				Name: "total", Type: "decimal(12,2)", Nullable: false, Default: "0.00",
			},
			Previous: sqlgen.ColumnAlterSpec{
				Name: "total", Type: "decimal(10,2)", Nullable: false, Default: "0.00",
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := displayBatch(d, batch)
	want := "ALTER TABLE `wisp_meta`.`parent` MODIFY COLUMN `total` decimal(12,2) NOT NULL DEFAULT '0.00'"
	if normalizeDDL(got) != normalizeDDL(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildStructureAlterMySQLRenameColumn80(t *testing.T) {
	d := mysql.Dialect("mysql")
	batch, err := sqlgen.BuildStructureAlterStatements(d, sqlgen.StructureSpec{
		Schema: "wisp_meta", Table: "parent", Engine: "mysql", MySQLMajor: 8,
	}, parentFixtureDetails(), sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind: sqlgen.StructureOpRenameColumn, Column: "status", NewName: "state",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := displayBatch(d, batch)
	want := "ALTER TABLE `wisp_meta`.`parent` RENAME COLUMN `status` TO `state`"
	if normalizeDDL(got) != normalizeDDL(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildStructureAlterMySQLRenameColumn57UsesChange(t *testing.T) {
	d := mysql.Dialect("mysql")
	batch, err := sqlgen.BuildStructureAlterStatements(d, sqlgen.StructureSpec{
		Schema: "wisp_meta", Table: "parent", Engine: "mysql", MySQLMajor: 5,
	}, parentFixtureDetails(), sqlgen.StructureAlterPlan{
		Ops: []sqlgen.StructureOp{{
			Kind: sqlgen.StructureOpRenameColumn, Column: "status", NewName: "state",
			Previous: sqlgen.ColumnAlterSpec{Name: "status", Type: "text", Nullable: false, Default: "'new'"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := displayBatch(d, batch)
	if !containsAll(got, "CHANGE COLUMN", "`status`", "`state`", "text") {
		t.Fatalf("got %q", got)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !stringsContains(s, p) {
			return false
		}
	}
	return true
}

func stringsContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
