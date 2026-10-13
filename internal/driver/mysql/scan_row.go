package mysql

import (
	"database/sql"

	"github.com/suhaasya/wisp-go/internal/driver"
)

type mysqlScanTarget struct {
	dest  any
	apply func(*driver.ColumnBuilder, int)
}

func mysqlScanTargets(types []*sql.ColumnType) []mysqlScanTarget {
	out := make([]mysqlScanTarget, len(types))
	for i, ct := range types {
		typ := wispTypeName(ct)
		switch typ {
		case "bool":
			v := new(sql.NullBool)
			out[i] = mysqlScanTarget{
				dest: v,
				apply: func(b *driver.ColumnBuilder, row int) {
					if !v.Valid {
						b.SetNull(row)
						return
					}
					b.SetBool(row, v.Bool)
				},
			}
		case "int2", "int4", "int8":
			v := new(sql.NullInt64)
			out[i] = mysqlScanTarget{
				dest: v,
				apply: func(b *driver.ColumnBuilder, row int) {
					if !v.Valid {
						b.SetNull(row)
						return
					}
					b.SetInt(row, v.Int64)
				},
			}
		case "float4", "float8":
			v := new(sql.NullFloat64)
			out[i] = mysqlScanTarget{
				dest: v,
				apply: func(b *driver.ColumnBuilder, row int) {
					if !v.Valid {
						b.SetNull(row)
						return
					}
					b.SetFloat(row, v.Float64)
				},
			}
		default:
			v := new(sql.NullString)
			out[i] = mysqlScanTarget{
				dest: v,
				apply: func(b *driver.ColumnBuilder, row int) {
					if !v.Valid {
						b.SetNull(row)
						return
					}
					b.SetText(row, v.String)
				},
			}
		}
	}
	return out
}

func scanRowInto(w *driver.Window, row int, types []*sql.ColumnType, scan func(dest ...any) error) error {
	targets := mysqlScanTargets(types)
	dest := make([]any, len(targets))
	for i := range targets {
		dest[i] = targets[i].dest
	}
	if err := scan(dest...); err != nil {
		return err
	}
	for i := range targets {
		targets[i].apply(w.Column(i), row)
	}
	return nil
}
