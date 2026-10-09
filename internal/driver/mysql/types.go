package mysql

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func columnsFromTypes(types []*sql.ColumnType) []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(types))
	for i, ct := range types {
		out[i] = driver.ColumnMeta{
			Name: ct.Name(),
			Type: wispTypeName(ct),
		}
	}
	return out
}

func wispTypeName(ct *sql.ColumnType) string {
	db := strings.ToUpper(ct.DatabaseTypeName())
	switch db {
	case "TINYINT":
		if length, ok := ct.Length(); ok && length == 1 {
			return "bool"
		}
		return "int8"
	case "SMALLINT":
		return "int2"
	case "MEDIUMINT", "INT", "INTEGER":
		return "int4"
	case "BIGINT":
		return "int8"
	case "FLOAT":
		return "float4"
	case "DOUBLE", "DOUBLE PRECISION":
		return "float8"
	case "DECIMAL", "NUMERIC":
		return "numeric"
	case "BIT":
		return "text"
	case "JSON":
		return "json"
	case "DATE":
		return "date"
	case "TIME":
		return "time"
	case "DATETIME", "TIMESTAMP":
		return "timestamp"
	default:
		if strings.Contains(db, "BLOB") || strings.Contains(db, "BINARY") {
			return "bytea"
		}
		return "text"
	}
}

func decodeCell(b *driver.ColumnBuilder, ct *sql.ColumnType, raw []byte, row int) {
	if raw == nil {
		b.SetNull(row)
		return
	}
	typ := wispTypeName(ct)
	switch typ {
	case "bool":
		s := string(raw)
		b.SetBool(row, s == "1" || s == "true" || s == "TRUE")
	case "int2", "int4", "int8":
		n, err := strconv.ParseInt(string(raw), 10, 64)
		if err != nil {
			b.SetText(row, string(raw))
			return
		}
		b.SetInt(row, n)
	case "float4", "float8":
		f, err := strconv.ParseFloat(string(raw), 64)
		if err != nil {
			b.SetText(row, string(raw))
			return
		}
		b.SetFloat(row, f)
	default:
		if strings.EqualFold(ct.DatabaseTypeName(), "BIT") {
			b.SetText(row, "0x"+hexEncode(raw))
			return
		}
		b.SetText(row, string(raw))
	}
}

func hexEncode(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hexdigits[c>>4]
		out[i*2+1] = hexdigits[c&0xf]
	}
	return string(out)
}
