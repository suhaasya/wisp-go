package pager_test

import (
	"context"
	"testing"

	"github.com/suhaasya/wisp-go/internal/pager"
)

func TestManagerCellValuePreview(t *testing.T) {
	m := mockManager(t, 100)
	ctx := context.Background()
	open, err := m.Open(ctx, pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "t",
		KeyColumns: []string{"id"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.WindowJSON(ctx, open.RID, 0, 150); err != nil {
		t.Fatal(err)
	}
	out, err := m.CellValue(ctx, open.RID, 0, "name")
	if err != nil {
		t.Fatal(err)
	}
	if out.Value == "" {
		t.Fatal("expected preview value")
	}
	if !out.Full {
		t.Fatal("want full preview when not truncated")
	}
}
