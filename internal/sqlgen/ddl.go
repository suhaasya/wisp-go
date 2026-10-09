package sqlgen

import (
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// CreateTablePostgreSQL builds CREATE TABLE and non-PK indexes from metadata.
func CreateTablePostgreSQL(det *driver.TableDetails) string {
	if det == nil {
		return ""
	}
	qual := quotePGSchemaTable(det.Schema, det.Table)
	var colDefs []string
	for _, col := range det.Columns {
		line := fmt.Sprintf("  %s %s", quotePGIdent(col.Name), col.Type)
		if !col.Nullable {
			line += " NOT NULL"
		}
		if col.Default != "" {
			line += " DEFAULT " + col.Default
		}
		colDefs = append(colDefs, line)
	}
	if len(det.PrimaryKey) > 0 {
		parts := make([]string, len(det.PrimaryKey))
		for i, c := range det.PrimaryKey {
			parts[i] = quotePGIdent(c)
		}
		name := det.Table + "_pkey"
		for _, idx := range det.Indexes {
			if idx.Primary && idx.Name != "" {
				name = idx.Name
				break
			}
		}
		colDefs = append(colDefs, fmt.Sprintf("  CONSTRAINT %s PRIMARY KEY (%s)",
			quotePGIdent(name), strings.Join(parts, ", ")))
	}
	for _, chk := range det.Checks {
		expr := strings.TrimSpace(chk.Expression)
		if expr == "" {
			continue
		}
		name := chk.Name
		if name == "" {
			name = det.Table + "_check"
		}
		colDefs = append(colDefs, fmt.Sprintf("  CONSTRAINT %s CHECK %s",
			quotePGIdent(name), ensureParens(expr)))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "CREATE TABLE %s (\n", qual)
	b.WriteString(strings.Join(colDefs, ",\n"))
	b.WriteString("\n);\n")
	for _, idx := range det.Indexes {
		if idx.Primary {
			continue
		}
		cols := make([]string, len(idx.Columns))
		for i, c := range idx.Columns {
			cols[i] = quotePGIdent(c)
		}
		unique := ""
		if idx.Unique {
			unique = "UNIQUE "
		}
		fmt.Fprintf(&b, "\nCREATE %sINDEX %s ON %s USING btree (%s);\n",
			unique, quotePGIdent(idx.Name), qual, strings.Join(cols, ", "))
	}
	return b.String()
}

func quotePGIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quotePGSchemaTable(schema, table string) string {
	if schema == "" {
		return quotePGIdent(table)
	}
	return quotePGIdent(schema) + "." + quotePGIdent(table)
}

func ensureParens(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "(") {
		return expr
	}
	return "(" + expr + ")"
}
