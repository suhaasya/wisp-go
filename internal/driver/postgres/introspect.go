package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

var pgSystemSchemas = map[string]struct{}{
	"pg_catalog": {}, "information_schema": {}, "pg_toast": {},
}

// Introspector reads PostgreSQL catalog metadata via driver.Conn.
type Introspector struct {
	Conn driver.Conn
}

// Databases lists databases on the server (one query).
func (i *Introspector) Databases(ctx context.Context) ([]string, error) {
	const q = `SELECT datname FROM pg_catalog.pg_database
WHERE datallowconn AND datname NOT IN ('template0','template1')
ORDER BY datname`
	return driver.CollectTextColumn(ctx, i.Conn, q)
}

// Schemas lists user-visible schemas in the current database.
func (i *Introspector) Schemas(ctx context.Context) ([]string, error) {
	const q = `SELECT nspname FROM pg_catalog.pg_namespace n
WHERE nspname NOT LIKE 'pg\_%' ESCAPE '\'
AND nspname <> 'information_schema'
ORDER BY nspname`
	return driver.CollectTextColumn(ctx, i.Conn, q)
}

// Objects lists object names of kind in schema (one query per call).
func (i *Introspector) Objects(ctx context.Context, schema string, kind driver.ObjectKind) ([]string, error) {
	if _, skip := pgSystemSchemas[schema]; skip {
		return nil, nil
	}
	switch kind {
	case driver.ObjectKindTable:
		return i.relations(ctx, schema, `'r','p','v','m','f'`)
	case driver.ObjectKindView:
		return i.relations(ctx, schema, `'v'`)
	case driver.ObjectKindMatView:
		return i.relations(ctx, schema, `'m'`)
	case driver.ObjectKindSequence:
		return i.relations(ctx, schema, `'S'`)
	case driver.ObjectKindFunction:
		return i.routines(ctx, schema, "f")
	case driver.ObjectKindProcedure:
		return i.routines(ctx, schema, "p")
	case driver.ObjectKindEnum:
		return i.enums(ctx, schema)
	case driver.ObjectKindTrigger:
		return i.triggers(ctx, schema)
	default:
		return nil, fmt.Errorf("postgres: unknown object kind %q", kind)
	}
}

func (i *Introspector) relations(ctx context.Context, schema, kinds string) ([]string, error) {
	q := fmt.Sprintf(`SELECT c.relname FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relkind IN (%s)
AND NOT c.relispartition
ORDER BY c.relname`, kinds)
	return driver.CollectTextColumn(ctx, i.Conn, q, schema)
}

func (i *Introspector) routines(ctx context.Context, schema, prokind string) ([]string, error) {
	const q = `SELECT p.proname FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = $1 AND p.prokind = $2
ORDER BY p.proname`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema, prokind)
}

func (i *Introspector) enums(ctx context.Context, schema string) ([]string, error) {
	const q = `SELECT t.typname FROM pg_catalog.pg_type t
JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = $1 AND t.typtype = 'e'
ORDER BY t.typname`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema)
}

func (i *Introspector) triggers(ctx context.Context, schema string) ([]string, error) {
	const q = `SELECT DISTINCT t.tgname FROM pg_catalog.pg_trigger t
JOIN pg_catalog.pg_class c ON c.oid = t.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND NOT t.tgisinternal
ORDER BY t.tgname`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema)
}

// TableDetails loads column and constraint metadata for one relation.
func (i *Introspector) TableDetails(ctx context.Context, schema, table string) (*driver.TableDetails, error) {
	kind, err := i.relationKind(ctx, schema, table)
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

func (i *Introspector) relationKind(ctx context.Context, schema, table string) (string, error) {
	const q = `SELECT c.relkind::text FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2 AND NOT c.relispartition`
	rows, err := driver.CollectTextColumn(ctx, i.Conn, q, schema, table)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("postgres: relation not found")
	}
	switch rows[0] {
	case "r", "p":
		return "table", nil
	case "v":
		return "view", nil
	case "m":
		return "matview", nil
	case "f":
		return "foreign", nil
	default:
		return rows[0], nil
	}
}

func (i *Introspector) columns(ctx context.Context, schema, table string) ([]driver.ColumnDetail, error) {
	const q = `SELECT a.attname,
pg_catalog.format_type(a.atttypid, a.atttypmod),
NOT a.attnotnull,
COALESCE(pg_catalog.pg_get_expr(d.adbin, d.adrelid), ''),
a.attnum
FROM pg_catalog.pg_attribute a
JOIN pg_catalog.pg_class c ON c.oid = a.attrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE n.nspname = $1 AND c.relname = $2
AND a.attnum > 0 AND NOT a.attisdropped
ORDER BY a.attnum`
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
	const q = `SELECT a.attname FROM pg_catalog.pg_index i
JOIN pg_catalog.pg_class c ON c.oid = i.indrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(i.indkey)
WHERE n.nspname = $1 AND c.relname = $2 AND i.indisprimary
ORDER BY array_position(i.indkey, a.attnum)`
	return driver.CollectTextColumn(ctx, i.Conn, q, schema, table)
}

func (i *Introspector) foreignKeys(ctx context.Context, schema, table string) ([]driver.ForeignKeyDetail, error) {
	const q = `SELECT con.conname,
pg_catalog.array_to_string(array_agg(a.attname ORDER BY u.ord), ','),
fn.nspname, fc.relname,
pg_catalog.array_to_string(array_agg(fa.attname ORDER BY u.ord), ',')
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class c ON c.oid = con.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
JOIN unnest(con.conkey) WITH ORDINALITY AS u(attnum, ord) ON true
JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attnum = u.attnum
JOIN pg_catalog.pg_class fc ON fc.oid = con.confrelid
JOIN pg_catalog.pg_namespace fn ON fn.oid = fc.relnamespace
JOIN unnest(con.confkey) WITH ORDINALITY AS fu(attnum, ord) ON fu.ord = u.ord
JOIN pg_catalog.pg_attribute fa ON fa.attrelid = fc.oid AND fa.attnum = fu.attnum
WHERE n.nspname = $1 AND c.relname = $2 AND con.contype = 'f'
GROUP BY con.conname, fn.nspname, fc.relname
ORDER BY con.conname`
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
				Columns:    splitCSV(cols),
				RefSchema:  refSchema,
				RefTable:   refTable,
				RefColumns: splitCSV(refCols),
			})
		}
	}
	return out, nil
}

func (i *Introspector) indexes(ctx context.Context, schema, table string) ([]driver.IndexDetail, error) {
	const q = `SELECT ic.relname,
pg_catalog.array_to_string(array_agg(a.attname ORDER BY u.ord), ','),
i.indisunique, i.indisprimary
FROM pg_catalog.pg_index i
JOIN pg_catalog.pg_class c ON c.oid = i.indrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
JOIN pg_catalog.pg_class ic ON ic.oid = i.indexrelid
JOIN unnest(i.indkey) WITH ORDINALITY AS u(attnum, ord) ON true
JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attnum = u.attnum
WHERE n.nspname = $1 AND c.relname = $2
GROUP BY ic.relname, i.indisunique, i.indisprimary
ORDER BY ic.relname`
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
				Name: name, Columns: splitCSV(cols), Unique: uniq, Primary: pri,
			})
		}
	}
	return out, nil
}

func (i *Introspector) checks(ctx context.Context, schema, table string) ([]driver.CheckDetail, error) {
	const q = `SELECT con.conname, pg_catalog.pg_get_constraintdef(con.oid, true)
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class c ON c.oid = con.conrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2 AND con.contype = 'c'
ORDER BY con.conname`
	rows, err := i.Conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, err
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
	const q = `SELECT t.tgname,
CASE WHEN t.tgtype & 2 = 2 THEN 'BEFORE' WHEN t.tgtype & 64 = 64 THEN 'INSTEAD OF' ELSE 'AFTER' END,
trim(both ' ' from (
CASE WHEN t.tgtype & 4 = 4 THEN 'INSERT ' ELSE '' END ||
CASE WHEN t.tgtype & 8 = 8 THEN 'UPDATE ' ELSE '' END ||
CASE WHEN t.tgtype & 16 = 16 THEN 'DELETE ' ELSE '' END ||
CASE WHEN t.tgtype & 32 = 32 THEN 'TRUNCATE ' ELSE '' END
)),
NOT t.tgenabled = 'D'
FROM pg_catalog.pg_trigger t
JOIN pg_catalog.pg_class c ON c.oid = t.tgrelid
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2 AND NOT t.tgisinternal
ORDER BY t.tgname`
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
			events, _ := w.CellText(2, r)
			en, _ := w.CellBool(3, r)
			out = append(out, driver.TriggerDetail{Name: name, Timing: timing, Events: events, Enabled: en})
		}
	}
	return out, nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
