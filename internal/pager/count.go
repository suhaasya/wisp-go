package pager

import (
	"context"
	"fmt"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func estimateTableRows(ctx context.Context, conn driver.Conn, engine, schema, table string) (int64, error) {
	switch engine {
	case "postgres":
		const q = `SELECT COALESCE(c.reltuples::bigint, 0)
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2`
		return scalarInt64(ctx, conn, q, schema, table)
	case "mysql", "mariadb":
		const q = `SELECT TABLE_ROWS FROM information_schema.TABLES
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`
		return scalarInt64(ctx, conn, q, schema, table)
	default:
		return 0, nil
	}
}

func exactCount(ctx context.Context, conn driver.Conn, quotedTable, where string) (int64, error) {
	sqlText := "SELECT COUNT(*) FROM " + quotedTable
	if where != "" {
		sqlText += " WHERE " + where
	}
	return scalarInt64(ctx, conn, sqlText)
}

func scalarInt64(ctx context.Context, conn driver.Conn, sqlText string, args ...any) (int64, error) {
	rows, err := conn.Query(ctx, sqlText, args...)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	w, err := rows.NextWindow(1)
	if err != nil {
		return 0, err
	}
	if w == nil || w.Rows == 0 {
		return 0, fmt.Errorf("pager: empty count")
	}
	v, ok := w.CellInt(0, 0)
	if !ok {
		return 0, fmt.Errorf("pager: count not int")
	}
	return v, nil
}

// ExactCountResult runs cancellable COUNT for a result table browse.
func (m *Manager) ExactCount(ctx context.Context, rid string) (Total, error) {
	m.mu.Lock()
	r, ok := m.results[rid]
	m.mu.Unlock()
	if !ok {
		return Total{}, errNotFound
	}
	if r.source.Kind != SourceTable {
		return Total{}, fmt.Errorf("pager: count only for table results")
	}
	d := m.conn.Dialect()
	table := d.QuoteIdent(r.source.Schema) + "." + d.QuoteIdent(r.source.Table)
	n, err := exactCount(ctx, m.conn, table, "")
	if err != nil {
		return Total{}, err
	}
	total := Total{Kind: "exact", Value: n}
	m.mu.Lock()
	if rr, ok := m.results[rid]; ok {
		rr.total = total
	}
	m.mu.Unlock()
	return total, nil
}
