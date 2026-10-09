package sqlgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// StructureSpec identifies the table being altered.
type StructureSpec struct {
	Schema     string
	Table      string
	Engine     string
	MySQLMajor int
}

// ColumnAlterSpec is one column definition for ADD/MODIFY.
type ColumnAlterSpec struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable,omitempty"`
	Default  string `json:"default,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// StructureOpKind is a structure change operation.
type StructureOpKind string

const (
	StructureOpAddColumn           StructureOpKind = "add_column"
	StructureOpDropColumn          StructureOpKind = "drop_column"
	StructureOpRenameColumn        StructureOpKind = "rename_column"
	StructureOpAlterColumnType     StructureOpKind = "alter_column_type"
	StructureOpAlterColumnNullable StructureOpKind = "alter_column_nullable"
	StructureOpAlterColumnDefault  StructureOpKind = "alter_column_default"
	StructureOpAlterColumnComment  StructureOpKind = "alter_column_comment"
	StructureOpAddIndex            StructureOpKind = "add_index"
	StructureOpDropIndex           StructureOpKind = "drop_index"
	StructureOpAddForeignKey       StructureOpKind = "add_foreign_key"
	StructureOpDropForeignKey      StructureOpKind = "drop_foreign_key"
)

// StructureOp is one staged structure edit.
type StructureOp struct {
	Kind StructureOpKind `json:"kind,omitempty"`

	Column    string          `json:"column,omitempty"`
	NewName   string          `json:"new_name,omitempty"`
	ColumnDef ColumnAlterSpec `json:"column_def,omitempty"`
	Previous  ColumnAlterSpec `json:"previous,omitempty"`

	IndexName       string   `json:"index_name,omitempty"`
	IndexColumns    []string `json:"index_columns,omitempty"`
	IndexUnique     bool     `json:"index_unique,omitempty"`
	IndexConcurrent bool     `json:"index_concurrent,omitempty"`

	FKName       string   `json:"fk_name,omitempty"`
	FKColumns    []string `json:"fk_columns,omitempty"`
	FKRefSchema  string   `json:"fk_ref_schema,omitempty"`
	FKRefTable   string   `json:"fk_ref_table,omitempty"`
	FKRefColumns []string `json:"fk_ref_columns,omitempty"`

	TypeUsing string `json:"type_using,omitempty"`
}

// StructureAlterPlan is ordered structure DDL.
type StructureAlterPlan struct {
	Ops []StructureOp
}

// StructureAlterBatch splits statements for transaction boundaries.
type StructureAlterBatch struct {
	InTransaction []Stmt
	AfterCommit   []Stmt
}

// BuildStructureAlterStatements generates ALTER DDL for one plan.
func BuildStructureAlterStatements(d driver.Dialect, spec StructureSpec, baseline *driver.TableDetails, plan StructureAlterPlan) (StructureAlterBatch, error) {
	if spec.Schema == "" || spec.Table == "" {
		return StructureAlterBatch{}, errors.New("sqlgen: schema and table required")
	}
	engine := strings.ToLower(spec.Engine)
	if engine == "" {
		engine = strings.ToLower(d.Name())
	}
	var batch StructureAlterBatch
	for _, op := range plan.Ops {
		switch engine {
		case "postgres", "postgresql":
			st, outTx, err := buildPostgresStructureOp(d, spec, baseline, op)
			if err != nil {
				return StructureAlterBatch{}, err
			}
			if st.SQL != "" {
				batch.InTransaction = append(batch.InTransaction, st)
			}
			if outTx.SQL != "" {
				batch.AfterCommit = append(batch.AfterCommit, outTx)
			}
		case "mysql", "mariadb":
			st, err := buildMySQLStructureOp(d, spec, baseline, op, spec.MySQLMajor)
			if err != nil {
				return StructureAlterBatch{}, err
			}
			if st.SQL != "" {
				batch.InTransaction = append(batch.InTransaction, st)
			}
		default:
			return StructureAlterBatch{}, fmt.Errorf("sqlgen: unsupported engine %q", engine)
		}
	}
	return batch, nil
}

func qualTable(d driver.Dialect, schema, table string) string {
	return d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
}

func buildPostgresStructureOp(d driver.Dialect, spec StructureSpec, baseline *driver.TableDetails, op StructureOp) (inTx Stmt, outTx Stmt, err error) {
	table := qualTable(d, spec.Schema, spec.Table)
	switch op.Kind {
	case StructureOpAddColumn:
		col := op.ColumnDef
		if col.Name == "" || col.Type == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: add column needs name and type")
		}
		line := fmt.Sprintf("%s %s", d.QuoteIdent(col.Name), col.Type)
		if !col.Nullable {
			line += " NOT NULL"
		}
		if col.Default != "" {
			line += " DEFAULT " + col.Default
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, line)}, Stmt{}, nil
	case StructureOpDropColumn:
		if op.Column == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: drop column needs name")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, d.QuoteIdent(op.Column))}, Stmt{}, nil
	case StructureOpRenameColumn:
		if op.Column == "" || op.NewName == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: rename column needs old and new name")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			table, d.QuoteIdent(op.Column), d.QuoteIdent(op.NewName))}, Stmt{}, nil
	case StructureOpAlterColumnType:
		if op.Column == "" || op.ColumnDef.Type == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: alter type needs column and type")
		}
		using := strings.TrimSpace(op.TypeUsing)
		if using == "" {
			using = fmt.Sprintf("%s::%s", d.QuoteIdent(op.Column), op.ColumnDef.Type)
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s",
			table, d.QuoteIdent(op.Column), op.ColumnDef.Type, using)}, Stmt{}, nil
	case StructureOpAlterColumnNullable:
		if op.Column == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: alter nullable needs column")
		}
		if op.ColumnDef.Nullable {
			return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", table, d.QuoteIdent(op.Column))}, Stmt{}, nil
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", table, d.QuoteIdent(op.Column))}, Stmt{}, nil
	case StructureOpAlterColumnDefault:
		if op.Column == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: alter default needs column")
		}
		if op.ColumnDef.Default == "" {
			return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT", table, d.QuoteIdent(op.Column))}, Stmt{}, nil
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
			table, d.QuoteIdent(op.Column), op.ColumnDef.Default)}, Stmt{}, nil
	case StructureOpAlterColumnComment:
		if op.Column == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: comment needs column")
		}
		col := qualTable(d, spec.Schema, spec.Table) + "." + d.QuoteIdent(op.Column)
		lit := d.QuoteLiteral(op.ColumnDef.Comment)
		return Stmt{SQL: fmt.Sprintf("COMMENT ON COLUMN %s IS %s", col, lit)}, Stmt{}, nil
	case StructureOpAddIndex:
		if op.IndexName == "" || len(op.IndexColumns) == 0 {
			return Stmt{}, Stmt{}, errors.New("sqlgen: add index needs name and columns")
		}
		cols := quoteIdents(d, op.IndexColumns)
		unique := ""
		if op.IndexUnique {
			unique = "UNIQUE "
		}
		var sql string
		if op.IndexConcurrent {
			sql = fmt.Sprintf("CREATE %sINDEX CONCURRENTLY %s ON %s (%s)",
				unique, d.QuoteIdent(op.IndexName), table, strings.Join(cols, ", "))
		} else {
			sql = fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)",
				unique, d.QuoteIdent(op.IndexName), table, strings.Join(cols, ", "))
		}
		if op.IndexConcurrent {
			return Stmt{}, Stmt{SQL: sql}, nil
		}
		return Stmt{SQL: sql}, Stmt{}, nil
	case StructureOpDropIndex:
		if op.IndexName == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: drop index needs name")
		}
		return Stmt{SQL: fmt.Sprintf("DROP INDEX %s", d.QuoteIdent(op.IndexName))}, Stmt{}, nil
	case StructureOpAddForeignKey:
		if op.FKName == "" || len(op.FKColumns) == 0 || op.FKRefTable == "" || len(op.FKRefColumns) == 0 {
			return Stmt{}, Stmt{}, errors.New("sqlgen: incomplete foreign key")
		}
		refSchema := op.FKRefSchema
		if refSchema == "" {
			refSchema = spec.Schema
		}
		ref := qualTable(d, refSchema, op.FKRefTable)
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			table, d.QuoteIdent(op.FKName), strings.Join(quoteIdents(d, op.FKColumns), ", "),
			ref, strings.Join(quoteIdents(d, op.FKRefColumns), ", "))}, Stmt{}, nil
	case StructureOpDropForeignKey:
		if op.FKName == "" {
			return Stmt{}, Stmt{}, errors.New("sqlgen: drop fk needs name")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", table, d.QuoteIdent(op.FKName))}, Stmt{}, nil
	default:
		return Stmt{}, Stmt{}, fmt.Errorf("sqlgen: unsupported postgres op %q", op.Kind)
	}
}

func quoteIdents(d driver.Dialect, names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = d.QuoteIdent(n)
	}
	return out
}

func buildMySQLStructureOp(d driver.Dialect, spec StructureSpec, baseline *driver.TableDetails, op StructureOp, major int) (Stmt, error) {
	table := qualTable(d, spec.Schema, spec.Table)
	if major <= 0 {
		major = 8
	}
	switch op.Kind {
	case StructureOpAddColumn:
		col := op.ColumnDef
		if col.Name == "" || col.Type == "" {
			return Stmt{}, errors.New("sqlgen: add column needs name and type")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, mysqlColumnLine(d, col))}, nil
	case StructureOpDropColumn:
		if op.Column == "" {
			return Stmt{}, errors.New("sqlgen: drop column needs name")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, d.QuoteIdent(op.Column))}, nil
	case StructureOpRenameColumn:
		if op.Column == "" || op.NewName == "" {
			return Stmt{}, errors.New("sqlgen: rename column needs old and new name")
		}
		if major >= 8 {
			return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
				table, d.QuoteIdent(op.Column), d.QuoteIdent(op.NewName))}, nil
		}
		prev := op.Previous
		if prev.Type == "" {
			prev = columnAlterFromBaseline(baseline, op.Column)
		}
		prev.Name = op.NewName
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN %s %s",
			table, d.QuoteIdent(op.Column), mysqlColumnLine(d, prev))}, nil
	case StructureOpAlterColumnType, StructureOpAlterColumnNullable, StructureOpAlterColumnDefault:
		col := op.ColumnDef
		if col.Name == "" {
			col.Name = op.Column
		}
		if col.Type == "" {
			col = columnAlterFromBaseline(baseline, op.Column)
			if op.Kind == StructureOpAlterColumnType {
				col.Type = op.ColumnDef.Type
			}
			if op.Kind == StructureOpAlterColumnNullable {
				col.Nullable = op.ColumnDef.Nullable
			}
			if op.Kind == StructureOpAlterColumnDefault {
				col.Default = op.ColumnDef.Default
			}
		}
		if col.Name == "" || col.Type == "" {
			return Stmt{}, errors.New("sqlgen: modify column needs full definition")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", table, mysqlColumnLine(d, col))}, nil
	case StructureOpAddIndex:
		if op.IndexName == "" || len(op.IndexColumns) == 0 {
			return Stmt{}, errors.New("sqlgen: add index needs name and columns")
		}
		unique := ""
		if op.IndexUnique {
			unique = "UNIQUE "
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ADD %sINDEX %s (%s)",
			table, unique, d.QuoteIdent(op.IndexName), strings.Join(quoteIdents(d, op.IndexColumns), ", "))}, nil
	case StructureOpDropIndex:
		if op.IndexName == "" {
			return Stmt{}, errors.New("sqlgen: drop index needs name")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", table, d.QuoteIdent(op.IndexName))}, nil
	case StructureOpAddForeignKey:
		if op.FKName == "" || len(op.FKColumns) == 0 || op.FKRefTable == "" || len(op.FKRefColumns) == 0 {
			return Stmt{}, errors.New("sqlgen: incomplete foreign key")
		}
		refSchema := op.FKRefSchema
		if refSchema == "" {
			refSchema = spec.Schema
		}
		ref := qualTable(d, refSchema, op.FKRefTable)
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			table, d.QuoteIdent(op.FKName), strings.Join(quoteIdents(d, op.FKColumns), ", "),
			ref, strings.Join(quoteIdents(d, op.FKRefColumns), ", "))}, nil
	case StructureOpDropForeignKey:
		if op.FKName == "" {
			return Stmt{}, errors.New("sqlgen: drop fk needs name")
		}
		return Stmt{SQL: fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", table, d.QuoteIdent(op.FKName))}, nil
	case StructureOpAlterColumnComment:
		return Stmt{}, nil
	default:
		return Stmt{}, fmt.Errorf("sqlgen: unsupported mysql op %q", op.Kind)
	}
}

func mysqlColumnLine(d driver.Dialect, col ColumnAlterSpec) string {
	line := fmt.Sprintf("%s %s", d.QuoteIdent(col.Name), col.Type)
	if !col.Nullable {
		line += " NOT NULL"
	}
	if col.Default != "" {
		def := col.Default
		if !strings.Contains(def, "(") && !strings.HasPrefix(def, "'") && def != "NULL" {
			def = d.QuoteLiteral(strings.Trim(def, "'"))
		}
		line += " DEFAULT " + def
	}
	return line
}

func columnAlterFromBaseline(baseline *driver.TableDetails, name string) ColumnAlterSpec {
	if baseline == nil {
		return ColumnAlterSpec{Name: name}
	}
	for _, c := range baseline.Columns {
		if c.Name == name {
			return ColumnAlterSpec{
				Name: c.Name, Type: c.Type, Nullable: c.Nullable, Default: c.Default, Comment: c.Comment,
			}
		}
	}
	return ColumnAlterSpec{Name: name}
}
