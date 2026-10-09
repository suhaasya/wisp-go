package pager

import (
	"context"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

// CellValueResult is the JSON body for GET .../value.
type CellValueResult struct {
	Value     string `json:"value"`
	Truncated bool   `json:"truncated"`
	Full      bool   `json:"full"`
	JSON      bool   `json:"json,omitempty"`
}

// CellValue returns the display or full cell value for a logical row and column name.
func (m *Manager) CellValue(ctx context.Context, rid string, row int, colName string) (CellValueResult, error) {
	if row < 0 {
		return CellValueResult{}, fmt.Errorf("pager: row out of range")
	}
	m.mu.Lock()
	r, ok := m.results[rid]
	if !ok {
		m.mu.Unlock()
		return CellValueResult{}, errNotFound
	}
	r.lastUse = m.clock()
	m.mu.Unlock()

	colIdx := -1
	for i, c := range r.columns {
		if c.Name == colName {
			colIdx = i
			break
		}
	}
	if colIdx < 0 {
		return CellValueResult{}, fmt.Errorf("pager: unknown column %q", colName)
	}

	offset := (row / DefaultWindowRows) * DefaultWindowRows
	w, err := m.windowAt(ctx, r, offset, DefaultWindowRows)
	if err != nil {
		return CellValueResult{}, err
	}
	local := row - offset
	if local < 0 || local >= w.Rows {
		return CellValueResult{}, fmt.Errorf("pager: row not loaded")
	}
	preview, ok := w.CellString(colIdx, local)
	if !ok {
		return CellValueResult{Value: "", Full: true}, nil
	}
	if !w.IsTruncated(colIdx, local) {
		return CellValueResult{Value: preview, Full: true, JSON: looksLikeJSON(preview)}, nil
	}
	full, err := m.fetchFullCellValue(ctx, r, w, local, colName)
	if err != nil {
		return CellValueResult{}, err
	}
	return CellValueResult{Value: full, Truncated: true, Full: true, JSON: looksLikeJSON(full)}, nil
}

func (m *Manager) windowAt(ctx context.Context, r *result, offset, limit int) (*driver.Window, error) {
	if w, ok := r.cache.get(offset, limit); ok {
		return w, nil
	}
	w, err := r.fetch.Fetch(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	if w == nil {
		w = driver.NewWindow(r.columns, 0)
	}
	r.cache.put(offset, limit, w)
	applyOverlay(w, m.overlay(r.id))
	return w, nil
}

func (m *Manager) fetchFullCellValue(ctx context.Context, r *result, w *driver.Window, localRow int, colName string) (string, error) {
	if r.source.Kind != SourceTable {
		return "", fmt.Errorf("pager: full value only for table browse")
	}
	keys := r.source.KeyColumns
	if len(keys) == 0 {
		return "", fmt.Errorf("pager: no primary key for full value fetch")
	}
	var keyVals []any
	for _, k := range keys {
		ki := columnIndex(w.Columns, k)
		if ki < 0 {
			return "", fmt.Errorf("pager: key column %q missing", k)
		}
		s, ok := w.CellString(ki, localRow)
		if !ok {
			return "", fmt.Errorf("pager: key value null")
		}
		keyVals = append(keyVals, s)
	}
	sqlText, args, err := sqlgen.BuildPointLookup(
		m.conn.Dialect(),
		r.source.Schema,
		r.source.Table,
		colName,
		keys,
		keyVals,
	)
	if err != nil {
		return "", err
	}
	rows, err := m.conn.Query(ctx, sqlText, args...)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	out, err := rows.NextWindow(1)
	if err != nil {
		return "", err
	}
	if out == nil || out.Rows == 0 {
		return "", fmt.Errorf("pager: empty full value")
	}
	s, ok := out.CellString(0, 0)
	if !ok {
		return "", fmt.Errorf("pager: could not read full value")
	}
	return s, nil
}

func columnIndex(cols []driver.ColumnMeta, name string) int {
	for i, c := range cols {
		if c.Name == name {
			return i
		}
	}
	return -1
}

func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}
