package postgres

import (
	"encoding/hex"
	"strconv"

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
		b.SetText(row, string(raw))
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
