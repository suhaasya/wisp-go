package mover_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/mover"
)

func TestPreviewJSONLFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rows.jsonl")
	if err := os.WriteFile(path, []byte("{\"id\":1,\"name\":\"a\"}\n{\"id\":2,\"name\":\"b\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := mover.PreviewJSONL(t.Context(), path, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Columns) != 2 || got.Columns[0] != "id" {
		t.Fatalf("cols=%v", got.Columns)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("rows=%d", len(got.Rows))
	}
}
