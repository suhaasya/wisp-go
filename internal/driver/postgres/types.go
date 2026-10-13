package postgres

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suhaasya/wisp-go/internal/driver"
)

func columnsFromFields(fields []pgconn.FieldDescription) []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(fields))
	for i, f := range fields {
		out[i] = driver.ColumnMeta{Name: f.Name, Type: typeNameForOID(f.DataTypeOID)}
	}
	return out
}

func typeNameForOID(oid uint32) string {
	switch oid {
	case pgtype.Int2OID:
		return "int2"
	case pgtype.Int4OID:
		return "int4"
	case pgtype.Int8OID:
		return "int8"
	case pgtype.Float4OID:
		return "float4"
	case pgtype.Float8OID:
		return "float8"
	case pgtype.BoolOID:
		return "bool"
	case pgtype.NumericOID:
		return "numeric"
	case pgtype.TextOID, pgtype.VarcharOID, pgtype.BPCharOID, pgtype.NameOID:
		return "text"
	case pgtype.ByteaOID:
		return "bytea"
	case pgtype.JSONOID:
		return "json"
	case pgtype.JSONBOID:
		return "jsonb"
	case pgtype.UUIDOID:
		return "uuid"
	case pgtype.DateOID:
		return "date"
	case pgtype.TimeOID:
		return "time"
	case pgtype.TimestampOID:
		return "timestamp"
	case pgtype.TimestamptzOID:
		return "timestamptz"
	default:
		return "text"
	}
}

// decodeFromValue maps pgx Values() output into a window column (handles TOAST and binary codecs).
func decodeFromValue(b *driver.ColumnBuilder, v any, row int) {
	if v == nil {
		b.SetNull(row)
		return
	}
	switch x := v.(type) {
	case bool:
		b.SetBool(row, x)
	case int64:
		b.SetInt(row, x)
	case int32:
		b.SetInt(row, int64(x))
	case int16:
		b.SetInt(row, int64(x))
	case int8:
		b.SetInt(row, int64(x))
	case int:
		b.SetInt(row, int64(x))
	case uint64:
		b.SetInt(row, int64(x))
	case float64:
		b.SetFloat(row, x)
	case float32:
		b.SetFloat(row, float64(x))
	case string:
		b.SetText(row, x)
	case []byte:
		b.SetText(row, string(x))
	case pgtype.Text:
		if !x.Valid {
			b.SetNull(row)
			return
		}
		b.SetText(row, x.String)
	case time.Time:
		b.SetText(row, x.Format("2006-01-02 15:04:05.999999999"))
	case pgtype.Timestamp:
		if !x.Valid {
			b.SetNull(row)
			return
		}
		b.SetText(row, x.Time.Format("2006-01-02 15:04:05.999999999"))
	case pgtype.Timestamptz:
		if !x.Valid {
			b.SetNull(row)
			return
		}
		b.SetText(row, x.Time.Format("2006-01-02 15:04:05.999999999-07:00"))
	default:
		b.SetText(row, fmt.Sprint(x))
	}
}

func decodeIntoColumn(b *driver.ColumnBuilder, oid uint32, raw []byte, row int) {
	if raw == nil {
		b.SetNull(row)
		return
	}
	switch oid {
	case pgtype.Int2OID, pgtype.Int4OID, pgtype.Int8OID:
		n, err := parseInt64(raw)
		if err != nil {
			b.SetText(row, string(raw))
			return
		}
		b.SetInt(row, n)
	case pgtype.Float4OID, pgtype.Float8OID:
		f, err := parseFloat64(raw)
		if err != nil {
			b.SetText(row, string(raw))
			return
		}
		b.SetFloat(row, f)
	case pgtype.BoolOID:
		b.SetBool(row, bytesEqual(raw, "t") || bytesEqual(raw, "true"))
	case pgtype.ByteaOID:
		b.SetText(row, decodeBytea(raw))
	default:
		cp := append([]byte(nil), raw...)
		b.SetText(row, string(cp))
	}
}

func parseInt64(raw []byte) (int64, error) {
	return strconv.ParseInt(string(raw), 10, 64)
}

func parseFloat64(raw []byte) (float64, error) {
	return strconv.ParseFloat(string(raw), 64)
}

func bytesEqual(a []byte, s string) bool {
	return len(a) == len(s) && string(a) == s
}

func decodeBytea(raw []byte) string {
	if len(raw) >= 2 && raw[0] == '\\' && raw[1] == 'x' {
		b, err := hex.DecodeString(string(raw[2:]))
		if err == nil {
			return string(b)
		}
	}
	return string(raw)
}
