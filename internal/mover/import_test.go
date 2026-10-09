package mover_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmt"

	"github.com/suhaasya/wisp-go/internal/mover"
)

func TestImportCSVMockBatchInsert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rows.csv")
	if err := os.WriteFile(path, []byte("id,name\n1,one\n2,two\n3,three\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var execs []string
	exec := func(ctx context.Context, sql string, args ...any) error {
		execs = append(execs, sql)
		return nil
	}
	n, errs, err := mover.ImportCSV(t.Context(), mockDialect{}, nil, path, mover.DataImportOptions{
		Schema:      "public",
		Table:       "items",
		Delimiter:   ",",
		HasHeader:   true,
		BatchSize:   2,
		FileColumns: []string{"id", "name"},
		Mapping: []mover.ColumnMapping{
			{FileColumn: "id", TableColumn: "id"},
			{FileColumn: "name", TableColumn: "name"},
		},
	}, exec, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	if n != 3 {
		t.Fatalf("rows=%d", n)
	}
	if len(execs) != 2 {
		t.Fatalf("batches=%d want 2", len(execs))
	}
}

func TestFormatImportErrorsCSVCapped(t *testing.T) {
	var errs []mover.ImportError
	for i := 0; i < 1005; i++ {
		errs = append(errs, mover.ImportError{Line: int64(i + 1), Message: "bad"})
	}
	// FormatImportErrorsCSV doesn't cap - session caps when collecting
	out := mover.FormatImportErrorsCSV(errs[:mover.MaxImportErrors])
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != mover.MaxImportErrors+1 {
		t.Fatalf("lines=%d", len(lines))
	}
}

type mockDialect struct{}

func (mockDialect) Name() string                  { return "mock" }
func (mockDialect) QuoteIdent(name string) string { return `"` + name + `"` }
func (mockDialect) QuoteLiteral(s string) string  { return `'` + s + `'` }
func (mockDialect) Placeholder(n int) string      { return fmt.Sprintf("$%d", n) }

func (mockDialect) LimitOffset(limit, offset int) string { return "" }
func (mockDialect) KeysetPredicate([]string, []any, string) (string, []any, error) {
	return "", nil, nil
}
