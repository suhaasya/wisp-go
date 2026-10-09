//go:build integration

package mysql_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
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
	for _, stmt := range strings.Split(string(body), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.Exec(t.Context(), stmt); err != nil {
			t.Fatalf("exec fixture stmt: %v\n%s", err, stmt)
		}
	}
}

func TestIntrospectObjectsFixture(t *testing.T) {
	conn := openConn(t, 0)
	execFixture(t, conn, "ci/fixtures/mysql_objects.sql")

	intr := &mysql.Introspector{Conn: conn}
	tables, err := intr.Objects(t.Context(), "wisp_meta", driver.ObjectKindTable)
	if err != nil {
		t.Fatalf("Objects tables: %v", err)
	}
	assertContainsAll(t, tables, []string{"child", "parent"})

	views, err := intr.Objects(t.Context(), "wisp_meta", driver.ObjectKindView)
	if err != nil {
		t.Fatalf("Objects views: %v", err)
	}
	assertContainsAll(t, views, []string{"parent_v"})

	det, err := intr.TableDetails(t.Context(), "wisp_meta", "parent")
	if err != nil {
		t.Fatalf("TableDetails: %v", err)
	}
	if len(det.PrimaryKey) != 1 || det.PrimaryKey[0] != "id" {
		t.Fatalf("pk = %v", det.PrimaryKey)
	}
	child, err := intr.TableDetails(t.Context(), "wisp_meta", "child")
	if err != nil {
		t.Fatalf("child details: %v", err)
	}
	if len(child.ForeignKeys) != 1 {
		t.Fatalf("child fks = %d", len(child.ForeignKeys))
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
