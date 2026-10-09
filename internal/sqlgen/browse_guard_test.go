package sqlgen_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestBrowseWithRawWhereClassifiesRead(t *testing.T) {
	sqlText, err := sqlgen.BrowseSQLForGuard(postgres.Dialect(), sqlgen.BrowseSpec{
		Schema:         "public",
		Table:          "orders",
		AllowedColumns: []string{"status"},
		KeyColumns:     []string{"id"},
		RawWhere:       "status = 'paid'",
		Limit:          150,
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.ClassifyStatement(sqlText) != session.StatementRead {
		t.Fatalf("want read, sql=%q", sqlText)
	}
}
