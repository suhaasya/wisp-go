package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

var mySystemSchemas = map[string]struct{}{
	"information_schema": {}, "mysql": {}, "performance_schema": {}, "sys": {},
}

// Introspector reads MySQL/MariaDB metadata via information_schema.
type Introspector struct {
	Conn driver.Conn
}

// Databases lists user-visible databases (one query).
func (i *Introspector) Databases(ctx context.Context) ([]string, error) {
	const q = `SELECT SCHEMA_NAME FROM information_schema.SCHEMATA
WHERE SCHEMA_NAME NOT IN ('information_schema','mysql','performance_schema','sys')
ORDER BY SCHEMA_NAME`
	return driver.CollectTextColumn(ctx, i.Conn, q)
}

// Schemas lists databases visible to the session (MySQL has no separate schema namespace).
func (i *Introspector) Schemas(ctx context.Context) ([]string, error) {
	return i.Databases(ctx)
}

// Objects lists object names of kind in schema (one query per call).
func (i *Introspector) Objects(ctx context.Context, schema string, kind driver.ObjectKind) ([]string, error) {
	if _, skip := mySystemSchemas[schema]; skip {
		return nil, nil
	}
	switch kind {
	case driver.ObjectKindTable:
		return i.tables(ctx, schema, "BASE TABLE")
	case driver.ObjectKindView:
		return i.tables(ctx, schema, "VIEW")
	case driver.ObjectKindMatView:
		return nil, nil
	case driver.ObjectKindSequence:
		return nil, nil
	case driver.ObjectKindFunction:
		return i.routines(ctx, schema, "FUNCTION")
	case driver.ObjectKindProcedure:
		return i.routines(ctx, schema, "PROCEDURE")
	case driver.ObjectKindEnum:
		return i.enums(ctx, schema)
	case driver.ObjectKindTrigger:
		return i.schemaTriggers(ctx, schema)
	default:
		return nil, fmt.Errorf("mysql: unknown object kind %q", kind)
	}
}

func (i *Introspector) tables(ctx context.Context, schema, tableType string) ([]string, error) {
	const q = `SELECT TABLE_NAME FROM information_schema.TABLES
WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = ?
ORDER BY TABLE_NAME`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema, tableType)
}

func (i *Introspector) routines(ctx context.Context, schema, routineType string) ([]string, error) {
	const q = `SELECT ROUTINE_NAME FROM information_schema.ROUTINES
WHERE ROUTINE_SCHEMA = ? AND ROUTINE_TYPE = ?
ORDER BY ROUTINE_NAME`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema, routineType)
}

func (i *Introspector) enums(ctx context.Context, schema string) ([]string, error) {
	const q = `SELECT DISTINCT COLUMN_TYPE FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = ? AND DATA_TYPE = 'enum'
ORDER BY COLUMN_TYPE`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema)
}

func (i *Introspector) schemaTriggers(ctx context.Context, schema string) ([]string, error) {
	const q = `SELECT DISTINCT TRIGGER_NAME FROM information_schema.TRIGGERS
WHERE TRIGGER_SCHEMA = ?
ORDER BY TRIGGER_NAME`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema)
}

// TableDetails loads column and constraint metadata for one table.
func (i *Introspector) TableDetails(ctx context.Context, schema, table string) (*driver.TableDetails, error) {
	kind, err := i.tableKind(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	cols, err := i.columns(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	pk, err := i.primaryKey(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	fks, err := i.foreignKeys(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	idx, err := i.indexes(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	checks, err := i.checks(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	trigs, err := i.tableTriggers(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	return &driver.TableDetails{
		Schema:      schema,
		Table:       table,
		Kind:        kind,
		Columns:     cols,
		PrimaryKey:  pk,
		ForeignKeys: fks,
		Indexes:     idx,
		Checks:      checks,
		Triggers:    trigs,
	}, nil
}

func (i *Introspector) tableKind(ctx context.Context, schema, table string) (string, error) {
	const q = `SELECT TABLE_TYPE FROM information_schema.TABLES
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`
	rows, err := driver.CollectTextColumn(ctx, i.Conn, q, schema, table)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("mysql: table not found")
	}
	switch rows[0] {
	case "BASE TABLE":
		return "table", nil
	case "VIEW":
		return "view", nil
	default:
		return strings.ToLower(rows[0]), nil
	}
}

func (i *Introspector) columns(ctx context.Context, schema, table string) ([]driver.ColumnDetail, error) {
	const q = `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE = 'YES',
IFNULL(COLUMN_DEFAULT, ''), ORDINAL_POSITION
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
ORDER BY ORDINAL_POSITION`
	rows, err := i.Conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []driver.ColumnDetail
	for {
		w, err := rows.NextWindow(200)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			name, _ := w.CellText(0, r)
			typ, _ := w.CellText(1, r)
			nullable, _ := w.CellBool(2, r)
			def, _ := w.CellText(3, r)
			ord, _ := w.CellInt(4, r)
			out = append(out, driver.ColumnDetail{
				Name: name, Type: typ, Nullable: nullable, Default: def, OrdinalPos: int(ord),
			})
		}
	}
	return out, nil
}

func (i *Introspector) primaryKey(ctx context.Context, schema, table string) ([]string, error) {
	const q = `SELECT COLUMN_NAME FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND INDEX_NAME = 'PRIMARY'
ORDER BY SEQ_IN_INDEX`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema, table)
}

func (i *Introspector) foreignKeys(ctx context.Context, schema, table string) ([]driver.ForeignKeyDetail, error) {
	const q = `SELECT k.CONSTRAINT_NAME,
GROUP_CONCAT(k.COLUMN_NAME ORDER BY k.ORDINAL_POSITION),
k.REFERENCED_TABLE_SCHEMA, k.REFERENCED_TABLE_NAME,
GROUP_CONCAT(k.REFERENCED_COLUMN_NAME ORDER BY k.ORDINAL_POSITION)
FROM information_schema.KEY_COLUMN_USAGE k
WHERE k.TABLE_SCHEMA = ? AND k.TABLE_NAME = ? AND k.REFERENCED_TABLE_NAME IS NOT NULL
GROUP BY k.CONSTRAINT_NAME, k.REFERENCED_TABLE_SCHEMA, k.REFERENCED_TABLE_NAME
ORDER BY k.CONSTRAINT_NAME`
	rows, err := i.Conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []driver.ForeignKeyDetail
	for {
		w, err := rows.NextWindow(50)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			name, _ := w.CellText(0, r)
			cols, _ := w.CellText(1, r)
			refSchema, _ := w.CellText(2, r)
			refTable, _ := w.CellText(3, r)
			refCols, _ := w.CellText(4, r)
			out = append(out, driver.ForeignKeyDetail{
				Name:       name,
				Columns:    strings.Split(cols, ","),
				RefSchema:  refSchema,
				RefTable:   refTable,
				RefColumns: strings.Split(refCols, ","),
			})
		}
	}
	return out, nil
}

func (i *Introspector) indexes(ctx context.Context, schema, table string) ([]driver.IndexDetail, error) {
	const q = `SELECT INDEX_NAME,
GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX),
NOT NON_UNIQUE, INDEX_NAME = 'PRIMARY'
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
GROUP BY INDEX_NAME, NON_UNIQUE
ORDER BY INDEX_NAME`
	rows, err := i.Conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []driver.IndexDetail
	for {
		w, err := rows.NextWindow(50)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			name, _ := w.CellText(0, r)
			cols, _ := w.CellText(1, r)
			uniq, _ := w.CellBool(2, r)
			pri, _ := w.CellBool(3, r)
			out = append(out, driver.IndexDetail{
				Name: name, Columns: strings.Split(cols, ","), Unique: uniq, Primary: pri,
			})
		}
	}
	return out, nil
}

func (i *Introspector) checks(ctx context.Context, schema, table string) ([]driver.CheckDetail, error) {
	const q = `SELECT tc.CONSTRAINT_NAME, cc.CHECK_CLAUSE
FROM information_schema.TABLE_CONSTRAINTS tc
JOIN information_schema.CHECK_CONSTRAINTS cc
ON tc.CONSTRAINT_SCHEMA = cc.CONSTRAINT_SCHEMA AND tc.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
WHERE tc.TABLE_SCHEMA = ? AND tc.TABLE_NAME = ? AND tc.CONSTRAINT_TYPE = 'CHECK'
ORDER BY tc.CONSTRAINT_NAME`
	rows, err := i.Conn.Query(ctx, q, schema, table)
	if err != nil {
		// CHECK_CONSTRAINTS absent on older MySQL; treat as no checks.
		return nil, nil
	}
	defer func() { _ = rows.Close() }()
	var out []driver.CheckDetail
	for {
		w, err := rows.NextWindow(50)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			name, _ := w.CellText(0, r)
			expr, _ := w.CellText(1, r)
			out = append(out, driver.CheckDetail{Name: name, Expression: expr})
		}
	}
	return out, nil
}

func (i *Introspector) tableTriggers(ctx context.Context, schema, table string) ([]driver.TriggerDetail, error) {
	const q = `SELECT TRIGGER_NAME, ACTION_TIMING, EVENT_MANIPULATION, ACTION_ORDER > 0
FROM information_schema.TRIGGERS
WHERE TRIGGER_SCHEMA = ? AND EVENT_OBJECT_TABLE = ?
ORDER BY TRIGGER_NAME`
	rows, err := i.Conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []driver.TriggerDetail
	for {
		w, err := rows.NextWindow(50)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			name, _ := w.CellText(0, r)
			timing, _ := w.CellText(1, r)
			event, _ := w.CellText(2, r)
			en, _ := w.CellBool(3, r)
			out = append(out, driver.TriggerDetail{Name: name, Timing: timing, Events: event, Enabled: en})
		}
	}
	return out, nil
}
