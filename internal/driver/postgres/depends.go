package postgres

import (
	"context"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// DropDependent is an object that depends on a table (shown before DROP).
type DropDependent struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
}

// ListDropDependents returns views and materialized views that depend on the table.
func ListDropDependents(ctx context.Context, conn driver.Conn, schema, table string) ([]DropDependent, error) {
	const q = `
SELECT depns.nspname, dep.relname, dep.relkind
FROM pg_catalog.pg_class ref
JOIN pg_catalog.pg_namespace refns ON refns.oid = ref.relnamespace
JOIN pg_catalog.pg_depend d ON d.refobjid = ref.oid AND d.deptype IN ('a', 'n')
JOIN pg_catalog.pg_class dep ON dep.oid = d.objid
JOIN pg_catalog.pg_namespace depns ON depns.oid = dep.relnamespace
WHERE refns.nspname = $1 AND ref.relname = $2
  AND dep.relkind IN ('v', 'm')
  AND dep.oid <> ref.oid
ORDER BY 1, 2`
	rows, err := conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []DropDependent
	seen := map[string]struct{}{}
	for {
		w, err := rows.NextWindow(200)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			sch, ok0 := w.CellText(0, r)
			name, ok1 := w.CellText(1, r)
			kind, ok2 := w.CellText(2, r)
			if !ok0 || !ok1 || !ok2 {
				continue
			}
			key := sch + "\x00" + name
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, DropDependent{
				Schema: sch,
				Name:   name,
				Kind:   relKindLabel(kind),
			})
		}
	}
	return out, nil
}

func relKindLabel(k string) string {
	switch k {
	case "v":
		return "view"
	case "m":
		return "matview"
	default:
		return k
	}
}
