package pager

import (
	"context"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

type keysetTableFetcher struct {
	conn    driver.Conn
	table   string
	keys    []string
	columns []driver.ColumnMeta
}

func newKeysetTableFetcher(conn driver.Conn, table string, keys []string, columns []driver.ColumnMeta) *keysetTableFetcher {
	return &keysetTableFetcher{conn: conn, table: table, keys: keys, columns: columns}
}

func (f *keysetTableFetcher) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(f.columns))
	copy(out, f.columns)
	return out
}

func (f *keysetTableFetcher) Close() error { return nil }

func (f *keysetTableFetcher) Fetch(ctx context.Context, offset, limit int) (*driver.Window, error) {
	d := f.conn.Dialect()
	order := strings.Join(f.keys, ", ")
	sqlText := fmt.Sprintf("SELECT * FROM %s ORDER BY %s %s", f.table, order, d.LimitOffset(limit, offset))
	return f.queryWindow(ctx, sqlText, limit)
}

func (f *keysetTableFetcher) queryWindow(ctx context.Context, sqlText string, limit int) (*driver.Window, error) {
	rows, err := f.conn.Query(ctx, sqlText)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return rows.NextWindow(limit)
}
