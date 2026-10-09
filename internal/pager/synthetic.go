package pager

import (
	"context"
	"fmt"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// syntheticTableFetcher materializes deterministic rows for mock-engine tests.
type syntheticTableFetcher struct {
	columns []driver.ColumnMeta
	total   int
}

func newSyntheticTableFetcher(columns []driver.ColumnMeta, total int) *syntheticTableFetcher {
	return &syntheticTableFetcher{columns: columns, total: total}
}

func (f *syntheticTableFetcher) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(f.columns))
	copy(out, f.columns)
	return out
}

func (f *syntheticTableFetcher) Close() error { return nil }

func (f *syntheticTableFetcher) Fetch(ctx context.Context, offset, limit int) (*driver.Window, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if offset >= f.total {
		return driver.NewWindow(f.columns, 0), nil
	}
	n := limit
	if remain := f.total - offset; remain < n {
		n = remain
	}
	w := driver.NewWindow(f.columns, n)
	for ci, col := range f.columns {
		b := w.Column(ci)
		for row := 0; row < n; row++ {
			global := offset + row
			switch col.Type {
			case "int8", "int4", "integer", "bigint":
				b.SetInt(row, int64(global))
			default:
				b.SetText(row, fmt.Sprintf("row-%d", global))
			}
		}
	}
	return w, nil
}
