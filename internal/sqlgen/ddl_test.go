package sqlgen_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestCreateTablePostgreSQLMatchesGolden(t *testing.T) {
	det := parentFixtureDetails()
	got := sqlgen.CreateTablePostgreSQL(det)
	want := readGolden(t, "parent_pg.golden")
	if normalizeDDL(got) != normalizeDDL(want) {
		t.Fatalf("ddl mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func parentFixtureDetails() *driver.TableDetails {
	return &driver.TableDetails{
		Schema: "wisp_meta",
		Table:  "parent",
		Columns: []driver.ColumnDetail{
			{Name: "id", Type: "bigint", Nullable: false, OrdinalPos: 1},
			{Name: "status", Type: "text", Nullable: false, Default: "'new'::text", OrdinalPos: 2},
		},
		PrimaryKey: []string{"id"},
		Checks: []driver.CheckDetail{
			{Name: "parent_status_check", Expression: `(status IN ('new', 'done'))`},
		},
		Indexes: []driver.IndexDetail{
			{Name: "idx_parent_status", Columns: []string{"status"}, Unique: false},
		},
	}
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	path := filepath.Join(filepath.Dir(file), "testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func normalizeDDL(s string) string {
	s = strings.ReplaceAll(s, `"`, "")
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, strings.ToLower(line))
	}
	return strings.Join(out, "\n")
}
