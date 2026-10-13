package postgres

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suhaasya/wisp-go/internal/driver"
)

type scanTarget struct {
	dest  any
	apply func(*driver.ColumnBuilder, int)
}

func intScanTarget() scanTarget {
	v := new(int64)
	return scanTarget{
		dest: v,
		apply: func(b *driver.ColumnBuilder, row int) {
			b.SetInt(row, *v)
		},
	}
}

func floatScanTarget() scanTarget {
	v := new(float64)
	return scanTarget{
		dest: v,
		apply: func(b *driver.ColumnBuilder, row int) {
			b.SetFloat(row, *v)
		},
	}
}

func boolScanTarget() scanTarget {
	v := new(bool)
	return scanTarget{
		dest: v,
		apply: func(b *driver.ColumnBuilder, row int) {
			b.SetBool(row, *v)
		},
	}
}

func textScanTarget() scanTarget {
	t := new(pgtype.Text)
	return scanTarget{
		dest: t,
		apply: func(b *driver.ColumnBuilder, row int) {
			if !t.Valid {
				b.SetNull(row)
				return
			}
			b.SetText(row, t.String)
		},
	}
}

func scanTargets(fields []pgconn.FieldDescription) []scanTarget {
	out := make([]scanTarget, len(fields))
	for i, fd := range fields {
		switch fd.DataTypeOID {
		case pgtype.Int2OID, pgtype.Int4OID, pgtype.Int8OID:
			out[i] = intScanTarget()
		case pgtype.Float4OID, pgtype.Float8OID:
			out[i] = floatScanTarget()
		case pgtype.BoolOID:
			out[i] = boolScanTarget()
		case pgtype.TimestampOID, pgtype.DateOID:
			out[i] = timestampScanTarget()
		case pgtype.TimestamptzOID:
			out[i] = timestamptzScanTarget()
		default:
			out[i] = textScanTarget()
		}
	}
	return out
}

func timestampScanTarget() scanTarget {
	t := new(pgtype.Timestamp)
	return scanTarget{
		dest: t,
		apply: func(b *driver.ColumnBuilder, row int) {
			if !t.Valid {
				b.SetNull(row)
				return
			}
			b.SetText(row, t.Time.Format("2006-01-02 15:04:05.999999999"))
		},
	}
}

func timestamptzScanTarget() scanTarget {
	t := new(pgtype.Timestamptz)
	return scanTarget{
		dest: t,
		apply: func(b *driver.ColumnBuilder, row int) {
			if !t.Valid {
				b.SetNull(row)
				return
			}
			b.SetText(row, t.Time.Format("2006-01-02 15:04:05.999999999-07:00"))
		},
	}
}

func scanRowInto(w *driver.Window, row int, targets []scanTarget, scan func(dest ...any) error) error {
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
