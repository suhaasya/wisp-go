package sqlgen_test

import (
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestBuildBrowsePostgresSortAndPKTiebreaker(t *testing.T) {
	sqlText, args, err := sqlgen.BuildBrowse(postgres.Dialect(), sqlgen.BrowseSpec{
		Schema:         "public",
		Table:          "orders",
		AllowedColumns: []string{"id", "created_at"},
		KeyColumns:     []string{"id"},
		Sort:           []sqlgen.SortKey{{Column: "created_at", Desc: true}},
		Limit:          150,
		Offset:         0,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `SELECT * FROM "public"."orders" ORDER BY "created_at" DESC, "id" ASC LIMIT 150 OFFSET 0`
	if sqlText != want {
		t.Fatalf("sql:\n got %q\nwant %q", sqlText, want)
	}
	if len(args) != 0 {
		t.Fatalf("args=%v want none", args)
	}
}

func TestBuildBrowsePostgresEqBindsValue(t *testing.T) {
	payload := `x' OR '1'='1`
	sqlText, args, err := sqlgen.BuildBrowse(postgres.Dialect(), sqlgen.BrowseSpec{
		Schema:         "public",
		Table:          "orders",
		AllowedColumns: []string{"status"},
		KeyColumns:     []string{"id"},
		Filter: sqlgen.FilterGroup{
			Logic: "and",
			Conditions: []sqlgen.FilterCondition{
				{Column: "status", Op: "=", Value: payload},
			},
		},
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sqlText, payload) {
		t.Fatalf("value must not appear in SQL: %q", sqlText)
	}
	if len(args) != 1 || args[0] != payload {
		t.Fatalf("args=%v want [%q]", args, payload)
	}
	if !strings.Contains(sqlText, `"status" = $1`) {
		t.Fatalf("sql=%q", sqlText)
	}
}

func TestBuildBrowseMySQLIlikeUsesLike(t *testing.T) {
	sqlText, args, err := sqlgen.BuildBrowse(mysql.Dialect("mysql"), sqlgen.BrowseSpec{
		Schema:         "shop",
		Table:          "items",
		AllowedColumns: []string{"name"},
		KeyColumns:     []string{"id"},
		Filter: sqlgen.FilterGroup{
			Logic: "and",
			Conditions: []sqlgen.FilterCondition{
				{Column: "name", Op: "ilike", Value: "%a%"},
			},
		},
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToUpper(sqlText), "ILIKE") {
		t.Fatalf("mysql must not use ILIKE: %q", sqlText)
	}
	if !strings.Contains(sqlText, "LIKE") {
		t.Fatalf("sql=%q", sqlText)
	}
	if len(args) != 1 || args[0] != "%a%" {
		t.Fatalf("args=%v", args)
	}
}
