package sqlgen_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestBuildPointLookupCompositeKey(t *testing.T) {
	sqlText, args, err := sqlgen.BuildPointLookup(
		postgres.Dialect(),
		"public",
		"orders",
		"note",
		[]string{"tenant_id", "id"},
		[]any{"1", "42"},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := `SELECT "note" FROM "public"."orders" WHERE "tenant_id" = $1 AND "id" = $2`
	if sqlText != want {
		t.Fatalf("sql=%q", sqlText)
	}
	if len(args) != 2 || args[0] != "1" || args[1] != "42" {
		t.Fatalf("args=%v", args)
	}
}
