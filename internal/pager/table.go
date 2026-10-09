package pager

import (
	"context"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
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
	total, err := m.tableTotal(ctx, src, table)
	if err != nil {
		return nil, Total{}, err
	}
	return newTableBrowseFetcher(m.conn, src, cols), total, nil
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
	if sourceHasBrowseFilter(src) {
		spec := browseSpecFromSource(src)
		sqlText, args, err := sqlgen.BuildTableCount(m.conn.Dialect(), spec)
		if err != nil {
			return Total{}, err
		}
		n, err := scalarInt64(ctx, m.conn, sqlText, args...)
		if err != nil {
			return Total{}, err
		}
		return Total{Kind: "exact", Value: n}, nil
	}
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
