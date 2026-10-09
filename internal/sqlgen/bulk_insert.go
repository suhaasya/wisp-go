package sqlgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// BuildMultiRowInsert builds one INSERT with multiple VALUES clauses.
func BuildMultiRowInsert(d driver.Dialect, schema, table string, columns []string, rows [][]any) (Stmt, error) {
	if schema == "" || table == "" {
		return Stmt{}, errors.New("sqlgen: schema and table required")
	}
	if len(columns) == 0 {
		return Stmt{}, errors.New("sqlgen: columns required")
	}
	if len(rows) == 0 {
		return Stmt{}, errors.New("sqlgen: rows required")
	}
	quoted := d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
	var colNames []string
	for _, c := range columns {
		colNames = append(colNames, d.QuoteIdent(c))
	}
	var valueGroups []string
	var args []any
	n := 1
	for _, row := range rows {
		if len(row) != len(columns) {
			return Stmt{}, errors.New("sqlgen: row width mismatch")
		}
		var ph []string
		for range columns {
			ph = append(ph, d.Placeholder(n))
			n++
		}
		valueGroups = append(valueGroups, "("+strings.Join(ph, ", ")+")")
		args = append(args, row...)
	}
	sqlText := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		quoted,
		strings.Join(colNames, ", "),
		strings.Join(valueGroups, ", "),
	)
	return Stmt{SQL: sqlText, Args: args}, nil
}
