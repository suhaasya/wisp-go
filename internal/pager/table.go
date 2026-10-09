package pager

import (
	"context"
	"fmt"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func (m *Manager) newTableFetcher(ctx context.Context, src Source) (fetcher, Total, error) {
	cols := tableColumns(src)
	if m.engine == "mock" {
		n := m.mockTotal
		if n <= 0 {
			n = 10_000
		}
		total := Total{Kind: "exact", Value: int64(n)}
		return newSyntheticTableFetcher(cols, n), total, nil
	}
	d := m.conn.Dialect()
	table := d.QuoteIdent(src.Schema) + "." + d.QuoteIdent(src.Table)
	var keys []string
	for _, k := range src.KeyColumns {
		keys = append(keys, d.QuoteIdent(k))
	}
	total, err := m.tableTotal(ctx, src, table)
	if err != nil {
		return nil, Total{}, err
	}
	if len(keys) > 0 {
		return newKeysetTableFetcher(m.conn, table, keys, cols), total, nil
	}
	return newOffsetTableFetcher(m.conn, table, cols), total, nil
}

func tableColumns(src Source) []driver.ColumnMeta {
	if len(src.Columns) > 0 {
		out := make([]driver.ColumnMeta, len(src.Columns))
		copy(out, src.Columns)
		return out
	}
	return []driver.ColumnMeta{
		{Name: "id", Type: "int8"},
		{Name: "name", Type: "text"},
	}
}

func (m *Manager) tableTotal(ctx context.Context, src Source, quotedTable string) (Total, error) {
	est, err := estimateTableRows(ctx, m.conn, m.engine, src.Schema, src.Table)
	if err != nil {
		est = 0
	}
	if est > 0 && est < ExactCountMax {
		n, err := exactCount(ctx, m.conn, quotedTable, "")
		if err == nil {
			return Total{Kind: "exact", Value: n}, nil
		}
	}
	if est > 0 {
		return Total{Kind: "estimate", Value: est}, nil
	}
	n, err := exactCount(ctx, m.conn, quotedTable, "")
	if err != nil {
		return Total{}, err
	}
	return Total{Kind: "exact", Value: n}, nil
}

type offsetTableFetcher struct {
	conn    driver.Conn
	table   string
	columns []driver.ColumnMeta
}

func newOffsetTableFetcher(conn driver.Conn, table string, columns []driver.ColumnMeta) *offsetTableFetcher {
	if len(columns) == 0 {
		columns = []driver.ColumnMeta{{Name: "id", Type: "int8"}}
	}
	return &offsetTableFetcher{conn: conn, table: table, columns: columns}
}

func (f *offsetTableFetcher) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(f.columns))
	copy(out, f.columns)
	return out
}

func (f *offsetTableFetcher) Close() error { return nil }

func (f *offsetTableFetcher) Fetch(ctx context.Context, offset, limit int) (*driver.Window, error) {
	d := f.conn.Dialect()
	sqlText := fmt.Sprintf("SELECT * FROM %s %s", f.table, d.LimitOffset(limit, offset))
	rows, err := f.conn.Query(ctx, sqlText)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if scanned := rows.Columns(); len(scanned) > 0 {
		f.columns = scanned
	}
	return rows.NextWindow(limit)
}
