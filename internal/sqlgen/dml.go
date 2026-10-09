package sqlgen

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// CommitSpec identifies the table being committed.
type CommitSpec struct {
	Schema     string
	Table      string
	Optimistic bool
}

// CellChange is one staged column value for DML.
type CellChange struct {
	Null     bool
	Value    *string
	OrigNull bool
	Original *string
}

// RowUpdate is an UPDATE for one row.
type RowUpdate struct {
	KeyColumns []string
	KeyValues  []any
	Cells      map[string]CellChange
}

// RowDelete is a DELETE for one row.
type RowDelete struct {
	KeyColumns []string
	KeyValues  []any
	// Original row snapshot for optimistic DELETE (optional per column in Optimistic mode).
	OrigCells map[string]CellChange
}

// RowInsert is an INSERT for one new row.
type RowInsert struct {
	Cells map[string]CellChange
}

// CommitPlan is ordered DML for one table change set.
type CommitPlan struct {
	Deletes []RowDelete
	Updates []RowUpdate
	Inserts []RowInsert
}

// Stmt is one parameterized statement.
type Stmt struct {
	SQL  string
	Args []any
}

// BuildCommitStatements returns DELETE, UPDATE, INSERT statements in that order.
func BuildCommitStatements(d driver.Dialect, spec CommitSpec, plan CommitPlan) ([]Stmt, error) {
	if spec.Schema == "" || spec.Table == "" {
		return nil, errors.New("sqlgen: schema and table required")
	}
	table := d.QuoteIdent(spec.Schema) + "." + d.QuoteIdent(spec.Table)
	var out []Stmt
	for _, del := range plan.Deletes {
		st, err := buildDelete(d, table, spec.Optimistic, del)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	for _, upd := range plan.Updates {
		st, err := buildUpdate(d, table, spec.Optimistic, upd)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	for _, ins := range plan.Inserts {
		st, err := buildInsert(d, table, ins)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}

func buildUpdate(d driver.Dialect, table string, optimistic bool, upd RowUpdate) (Stmt, error) {
	if len(upd.KeyColumns) == 0 || len(upd.KeyColumns) != len(upd.KeyValues) {
		return Stmt{}, errors.New("sqlgen: update key mismatch")
	}
	if len(upd.Cells) == 0 {
		return Stmt{}, errors.New("sqlgen: update with no columns")
	}
	cols := sortedKeys(upd.Cells)
	var setParts []string
	var args []any
	n := 1
	for _, col := range cols {
		ch := upd.Cells[col]
		setParts = append(setParts, fmt.Sprintf("%s = %s", d.QuoteIdent(col), d.Placeholder(n)))
		n++
		if ch.Null {
			args = append(args, nil)
		} else if ch.Value != nil {
			args = append(args, *ch.Value)
		} else {
			args = append(args, nil)
		}
	}
	where, wArgs, next := pkWhere(d, upd.KeyColumns, upd.KeyValues, n)
	args = append(args, wArgs...)
	n = next
	if optimistic {
		op, oArgs, _ := optimisticAnd(d, cols, upd.Cells, n)
		if op != "" {
			where += op
			args = append(args, oArgs...)
		}
	}
	sqlText := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setParts, ", "), where)
	return Stmt{SQL: sqlText, Args: args}, nil
}

func buildDelete(d driver.Dialect, table string, optimistic bool, del RowDelete) (Stmt, error) {
	if len(del.KeyColumns) == 0 || len(del.KeyColumns) != len(del.KeyValues) {
		return Stmt{}, errors.New("sqlgen: delete key mismatch")
	}
	where, args, n := pkWhere(d, del.KeyColumns, del.KeyValues, 1)
	if optimistic && len(del.OrigCells) > 0 {
		cols := sortedKeys(del.OrigCells)
		op, oArgs, _ := optimisticAnd(d, cols, del.OrigCells, n)
		if op != "" {
			where += op
			args = append(args, oArgs...)
		}
	}
	sqlText := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	return Stmt{SQL: sqlText, Args: args}, nil
}

func buildInsert(d driver.Dialect, table string, ins RowInsert) (Stmt, error) {
	if len(ins.Cells) == 0 {
		return Stmt{}, errors.New("sqlgen: insert with no columns")
	}
	cols := sortedKeys(ins.Cells)
	var colNames []string
	var ph []string
	var args []any
	for i, col := range cols {
		colNames = append(colNames, d.QuoteIdent(col))
		ph = append(ph, d.Placeholder(i+1))
		ch := ins.Cells[col]
		if ch.Null {
			args = append(args, nil)
		} else if ch.Value != nil {
			args = append(args, *ch.Value)
		} else {
			args = append(args, nil)
		}
	}
	sqlText := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(colNames, ", "),
		strings.Join(ph, ", "),
	)
	return Stmt{SQL: sqlText, Args: args}, nil
}

func pkWhere(d driver.Dialect, keyCols []string, keyVals []any, start int) (string, []any, int) {
	var parts []string
	var args []any
	n := start
	for i, col := range keyCols {
		parts = append(parts, fmt.Sprintf("%s = %s", d.QuoteIdent(col), d.Placeholder(n)))
		args = append(args, keyVals[i])
		n++
	}
	return strings.Join(parts, " AND "), args, n
}

func optimisticAnd(d driver.Dialect, cols []string, cells map[string]CellChange, start int) (string, []any, int) {
	var parts []string
	var args []any
	n := start
	for _, col := range cols {
		ch := cells[col]
		if ch.OrigNull {
			parts = append(parts, fmt.Sprintf("%s IS NULL", d.QuoteIdent(col)))
			continue
		}
		if ch.Original == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = %s", d.QuoteIdent(col), d.Placeholder(n)))
		args = append(args, *ch.Original)
		n++
	}
	if len(parts) == 0 {
		return "", nil, n
	}
	return " AND " + strings.Join(parts, " AND "), args, n
}

func sortedKeys(m map[string]CellChange) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[i] > out[j] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// DisplaySQL inlines bound arguments using QuoteLiteral / numeric formatting (review dialog).
func DisplaySQL(d driver.Dialect, sqlText string, args []any) string {
	if len(args) == 0 {
		return sqlText
	}
	if strings.Contains(sqlText, "?") {
		var b strings.Builder
		parts := strings.Split(sqlText, "?")
		for i, part := range parts {
			b.WriteString(part)
			if i < len(args) {
				b.WriteString(formatArg(d, args[i]))
			}
		}
		return b.String()
	}
	out := sqlText
	for i := len(args); i >= 1; i-- {
		ph := d.Placeholder(i)
		out = strings.Replace(out, ph, formatArg(d, args[i-1]), 1)
	}
	return out
}

func formatArg(d driver.Dialect, v any) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case string:
		return d.QuoteLiteral(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return d.QuoteLiteral(fmt.Sprint(v))
	}
}
