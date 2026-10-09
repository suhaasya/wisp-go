package pager

import (
	"context"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

type tableBrowseFetcher struct {
	conn       driver.Conn
	schema     string
	table      string
	keyColumns []string
	columns    []driver.ColumnMeta
	filter     sqlgen.FilterGroup
	rawWhere   string
	sort       []sqlgen.SortKey
}

func newTableBrowseFetcher(conn driver.Conn, src Source, columns []driver.ColumnMeta) *tableBrowseFetcher {
	if len(columns) == 0 {
		columns = tableColumns(src)
	}
	return &tableBrowseFetcher{
		conn:       conn,
		schema:     src.Schema,
		table:      src.Table,
		keyColumns: append([]string(nil), src.KeyColumns...),
		columns:    columns,
		filter:     src.Filter,
		rawWhere:   src.RawWhere,
		sort:       append([]sqlgen.SortKey(nil), src.Sort...),
	}
}

func (f *tableBrowseFetcher) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(f.columns))
	copy(out, f.columns)
	return out
}

func (f *tableBrowseFetcher) Close() error { return nil }

func (f *tableBrowseFetcher) Fetch(ctx context.Context, offset, limit int) (*driver.Window, error) {
	spec := f.browseSpec(offset, limit)
	sqlText, args, err := sqlgen.BuildBrowse(f.conn.Dialect(), spec)
	if err != nil {
		return nil, err
	}
	return f.queryWindow(ctx, sqlText, limit, args...)
}

func (f *tableBrowseFetcher) browseSpec(offset, limit int) sqlgen.BrowseSpec {
	var allowed []string
	for _, c := range f.columns {
		allowed = append(allowed, c.Name)
	}
	return sqlgen.BrowseSpec{
		Schema:         f.schema,
		Table:          f.table,
		AllowedColumns: allowed,
		KeyColumns:     f.keyColumns,
		Filter:         f.filter,
		RawWhere:       f.rawWhere,
		Sort:           f.sort,
		Limit:          limit,
		Offset:         offset,
	}
}

func (f *tableBrowseFetcher) queryWindow(ctx context.Context, sqlText string, limit int, args ...any) (*driver.Window, error) {
	rows, err := f.conn.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if scanned := rows.Columns(); len(scanned) > 0 {
		f.columns = scanned
	}
	return rows.NextWindow(limit)
}

// BrowseSQLSpec builds sqlgen input from a table result source.
func BrowseSQLSpec(src Source) sqlgen.BrowseSpec {
	return browseSpecFromSource(src)
}

func browseSpecFromSource(src Source) sqlgen.BrowseSpec {
	var allowed []string
	for _, c := range src.Columns {
		allowed = append(allowed, c.Name)
	}
	return sqlgen.BrowseSpec{
		Schema:         src.Schema,
		Table:          src.Table,
		AllowedColumns: allowed,
		KeyColumns:     src.KeyColumns,
		Filter:         src.Filter,
		RawWhere:       src.RawWhere,
		Sort:           src.Sort,
	}
}

func sourceHasBrowseFilter(src Source) bool {
	if strings.TrimSpace(src.RawWhere) != "" {
		return true
	}
	return len(src.Filter.Conditions) > 0
}
