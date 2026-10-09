package sqlgen_test

import (
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestBuildMultiRowInsertPostgresPlaceholders(t *testing.T) {
	d := postgres.Dialect()
	st, err := sqlgen.BuildMultiRowInsert(d, "public", "t", []string{"a", "b"}, [][]any{
		{1, "x"},
		{2, "y"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(st.SQL, "INSERT INTO") || !strings.Contains(st.SQL, "VALUES") {
		t.Fatalf("sql=%q", st.SQL)
	}
	if len(st.Args) != 4 {
		t.Fatalf("args=%d", len(st.Args))
	}
	if st.Args[0] != 1 || st.Args[3] != "y" {
		t.Fatalf("args=%v", st.Args)
	}
}
