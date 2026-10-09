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

// DropTableSQL builds DROP TABLE (CASCADE only when cascade is true).
func DropTableSQL(d driver.Dialect, schema, table string, cascade bool) string {
	qual := d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
	if cascade {
		return fmt.Sprintf("DROP TABLE %s CASCADE", qual)
	}
	return fmt.Sprintf("DROP TABLE %s", qual)
}

// TruncateTableSQL builds TRUNCATE TABLE.
func TruncateTableSQL(d driver.Dialect, schema, table string) string {
	qual := d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
	return fmt.Sprintf("TRUNCATE TABLE %s", qual)
}

// RenameTableSQL renames a table.
func RenameTableSQL(d driver.Dialect, engine, schema, table, newName string) string {
	qual := d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
	switch strings.ToLower(engine) {
	case "mysql", "mariadb":
		return fmt.Sprintf("RENAME TABLE %s TO %s", qual, d.QuoteIdent(schema)+"."+d.QuoteIdent(newName))
	default:
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", qual, d.QuoteIdent(newName))
	}
}

// CreateSchemaSQL creates a schema (PostgreSQL) or database (MySQL).
func CreateSchemaSQL(d driver.Dialect, engine, name string) string {
	switch strings.ToLower(engine) {
	case "mysql", "mariadb":
		return fmt.Sprintf("CREATE DATABASE %s", d.QuoteIdent(name))
	default:
		return fmt.Sprintf("CREATE SCHEMA %s", d.QuoteIdent(name))
	}
}

// CreateTableMySQL builds CREATE TABLE for MySQL/MariaDB from metadata.
func CreateTableMySQL(det *driver.TableDetails) string {
	if det == nil {
		return ""
	}
	qual := "`" + det.Schema + "`.`" + det.Table + "`"
	var colDefs []string
	for _, col := range det.Columns {
		line := fmt.Sprintf("  `%s` %s", col.Name, col.Type)
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
			parts[i] = "`" + c + "`"
		}
		colDefs = append(colDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(parts, ", ")))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "CREATE TABLE %s (\n", qual)
	b.WriteString(strings.Join(colDefs, ",\n"))
	b.WriteString("\n)")
	for _, idx := range det.Indexes {
		if idx.Primary {
			continue
		}
		cols := make([]string, len(idx.Columns))
		for i, c := range idx.Columns {
			cols[i] = "`" + c + "`"
		}
		unique := ""
		if idx.Unique {
			unique = "UNIQUE "
		}
		fmt.Fprintf(&b, ";\nCREATE %sINDEX `%s` ON %s (%s)", unique, idx.Name, qual, strings.Join(cols, ", "))
	}
	b.WriteString(";")
	return b.String()
}

func ensureParens(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "(") {
		return expr
	}
	return "(" + expr + ")"
}
