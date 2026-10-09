package mock_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
)

func TestMockDriverStreamsOneMillionRows(t *testing.T) {
	const total = 1_000_000
	cols := []driver.ColumnMeta{{Name: "id", Type: "int8"}, {Name: "label", Type: "text"}}
	var d mock.Driver
	conn, err := d.Open(context.Background(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: total, Columns: cols},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	rows, err := conn.Query(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var got int
	for {
		w, err := rows.NextWindow(150)
		if err != nil {
			t.Fatal(err)
		}
		if w == nil {
			break
		}
		got += w.Rows
		w = nil
	}
	if got != total {
		t.Fatalf("rows = %d want %d", got, total)
	}

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	// One window (~150 rows × 2 cols) should not retain 1M rows worth of heap.
	const maxRetainBytes = 8 << 20 // 8 MiB headroom for test runtime
	if delta := int64(after.HeapAlloc) - int64(before.HeapAlloc); delta > maxRetainBytes {
		t.Fatalf("heap grew by %d bytes after streaming 1M rows", delta)
	}
}
