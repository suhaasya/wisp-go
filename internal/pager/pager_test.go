package pager_test

import (
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/pager"
)

func mockManager(t *testing.T, totalRows int) *pager.Manager {
	t.Helper()
	var d mock.Driver
	conn, err := d.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: totalRows,
			Columns: []driver.ColumnMeta{
				{Name: "id", Type: "int8"},
				{Name: "name", Type: "text"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return pager.NewManager(conn, "mock", func() time.Time { return time.Unix(0, 0) })
}

func TestManagerWindowCacheMaxThree(t *testing.T) {
	m := mockManager(t, 10_000)
	ctx := t.Context()
	open, err := m.Open(ctx, pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "t",
		KeyColumns: []string{"id"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, off := range []int{0, 150, 300, 450, 600} {
		if _, err := m.WindowJSON(ctx, open.RID, off, 150); err != nil {
			t.Fatalf("offset %d: %v", off, err)
		}
	}
	if m.CachedWindowCount(open.RID) > 3 {
		t.Fatalf("cached windows = %d, want <= 3", m.CachedWindowCount(open.RID))
	}
}

func TestKeysetJumpToEndUnder300ms(t *testing.T) {
	const total = 1_000_000
	m := mockManager(t, total)
	m.SetMockTotal(total)
	ctx := t.Context()
	open, err := m.Open(ctx, pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "t",
		KeyColumns: []string{"id"},
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if _, err := m.WindowJSON(ctx, open.RID, total-150, 150); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 300*time.Millisecond {
		t.Fatalf("jump to end took %v, want < 300ms", elapsed)
	}
}
