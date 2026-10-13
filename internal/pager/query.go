package pager

import (
	"context"
	"errors"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

type queryFetcher struct {
	conn    driver.Conn
	sql     string
	columns []driver.ColumnMeta
	rows    driver.Rows
	pos     int
}

func (m *Manager) newQueryFetcher(ctx context.Context, src Source) (fetcher, Total, error) {
	if stringsTrim(src.SQL) == "" {
		return nil, Total{}, errors.New("pager: query sql required")
	}
	if m.engine == "mock" {
		cols := []driver.ColumnMeta{{Name: "id", Type: "int8"}, {Name: "name", Type: "text"}}
		total := Total{Kind: "exact", Value: 10_000}
		return newSyntheticTableFetcher(cols, int(total.Value)), total, nil
	}
	rows, err := m.conn.Query(ctx, src.SQL)
	if err != nil {
		return nil, Total{}, err
	}
	cols := rows.Columns()
	f := &queryFetcher{conn: m.conn, sql: src.SQL, columns: cols, rows: rows}
	total := Total{Kind: "estimate", Value: 0}
	return f, total, nil
}

// updateQueryTotalAfterFetch adjusts row count for ad-hoc queries (unknown until rows are read).
func updateQueryTotalAfterFetch(r *result, offset, rows, limit int) {
	if r.source.Kind != SourceQuery {
		return
	}
	end := int64(offset + rows)
	if rows < limit {
		r.total = Total{Kind: "exact", Value: end}
		return
	}
	if r.total.Kind == "exact" {
		return
	}
	next := end + 1
	if r.total.Value < next {
		r.total = Total{Kind: "estimate", Value: next}
	}
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

func (f *queryFetcher) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(f.columns))
	copy(out, f.columns)
	return out
}

func (f *queryFetcher) Close() error {
	if f.rows != nil {
		err := f.rows.Close()
		f.rows = nil
		return err
	}
	return nil
}

func (f *queryFetcher) Fetch(ctx context.Context, offset, limit int) (*driver.Window, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.rows == nil {
		rows, err := f.conn.Query(ctx, f.sql)
		if err != nil {
			return nil, err
		}
		f.rows = rows
		f.pos = 0
	}
	if offset < f.pos {
		_ = f.rows.Close()
		rows, err := f.conn.Query(ctx, f.sql)
		if err != nil {
			return nil, err
		}
		f.rows = rows
		f.pos = 0
	}
	for f.pos < offset {
		skip := 500
		if remain := offset - f.pos; remain < skip {
			skip = remain
		}
		w, err := f.rows.NextWindow(skip)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			return driver.NewWindow(f.columns, 0), nil
		}
		f.pos += w.Rows
	}
	w, err := f.rows.NextWindow(limit)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return driver.NewWindow(f.columns, 0), nil
	}
	f.pos += w.Rows
	return w, nil
}
