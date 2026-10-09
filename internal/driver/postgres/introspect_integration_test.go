//go:build integration

package postgres_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func execFixture(t *testing.T, conn driver.Conn, relPath string) {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	for _, stmt := range splitSQL(string(body)) {
		if stmt == "" {
			continue
		}
		if _, err := conn.Exec(t.Context(), stmt); err != nil {
			t.Fatalf("exec fixture stmt: %v\n%s", err, stmt)
		}
	}
}

func splitSQL(s string) []string {
	var out []string
	var b strings.Builder
	inDollar := false
	for i := 0; i < len(s); i++ {
		if !inDollar && i+1 < len(s) && s[i] == '$' && s[i+1] == '$' {
			inDollar = true
			b.WriteByte(s[i])
			b.WriteByte(s[i+1])
			i++
			continue
		}
		if inDollar && i+1 < len(s) && s[i] == '$' && s[i+1] == '$' {
			inDollar = false
			b.WriteByte(s[i])
			b.WriteByte(s[i+1])
			i++
			continue
		}
		if !inDollar && s[i] == ';' {
			out = append(out, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		}
		b.WriteByte(s[i])
	}
	if tail := strings.TrimSpace(b.String()); tail != "" {
		out = append(out, tail)
	}
	return out
}

func TestIntrospectObjectsFixture(t *testing.T) {
	conn := openConn(t, 0)
	execFixture(t, conn, "ci/fixtures/pg_objects.sql")

	intr := &postgres.Introspector{Conn: conn}
	tables, err := intr.Objects(t.Context(), "wisp_meta", driver.ObjectKindTable)
	if err != nil {
		t.Fatalf("Objects tables: %v", err)
	}
	wantTables := []string{"child", "parent", "parted"}
	assertContainsAll(t, tables, wantTables)
	for _, hide := range []string{"parted_2024"} {
		for _, name := range tables {
			if name == hide {
				t.Fatalf("partition child %q should be hidden", hide)
			}
		}
	}

	views, err := intr.Objects(t.Context(), "wisp_meta", driver.ObjectKindView)
	if err != nil {
		t.Fatalf("Objects views: %v", err)
	}
	assertContainsAll(t, views, []string{"parent_v"})

	det, err := intr.TableDetails(t.Context(), "wisp_meta", "parent")
	if err != nil {
		t.Fatalf("TableDetails: %v", err)
	}
	if len(det.Columns) < 2 {
		t.Fatalf("columns = %d", len(det.Columns))
	}
	if len(det.PrimaryKey) != 1 || det.PrimaryKey[0] != "id" {
		t.Fatalf("pk = %v", det.PrimaryKey)
	}
	if len(det.ForeignKeys) != 0 {
		t.Fatalf("parent should have no FKs")
	}
	child, err := intr.TableDetails(t.Context(), "wisp_meta", "child")
	if err != nil {
		t.Fatalf("child details: %v", err)
	}
	if len(child.ForeignKeys) != 1 {
		t.Fatalf("child fks = %d", len(child.ForeignKeys))
	}
}

func TestIntrospectManyTablesUnder300ms(t *testing.T) {
	conn := openConn(t, 0)
	execFixture(t, conn, "ci/fixtures/many_tables.sql")
	intr := &postgres.Introspector{Conn: conn}
	start := time.Now()
	names, err := intr.Objects(t.Context(), "wisp_meta", driver.ObjectKindTable)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Objects: %v", err)
	}
	if len(names) != 500 {
		t.Fatalf("table count = %d, want 500", len(names))
	}
	if elapsed > 300*time.Millisecond {
		t.Fatalf("list 500 tables took %v, want < 300ms", elapsed)
	}
}

func assertContainsAll(t *testing.T, have []string, want []string) {
	t.Helper()
	set := make(map[string]struct{}, len(have))
	for _, h := range have {
		set[h] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			t.Fatalf("missing %q in %v", w, have)
		}
	}
}
