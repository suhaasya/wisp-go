package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestScanRowIntoMultipleTextColumns(t *testing.T) {
	fields := []pgconn.FieldDescription{
		{Name: "id", DataTypeOID: pgtype.Int4OID},
		{Name: "created_by", DataTypeOID: pgtype.Int4OID},
		{Name: "updated_at", DataTypeOID: pgtype.TimestampOID},
		{Name: "official_email", DataTypeOID: pgtype.VarcharOID},
		{Name: "last_modified_by", DataTypeOID: pgtype.Int4OID},
		{Name: "created_at", DataTypeOID: pgtype.TimestampOID},
		{Name: "active", DataTypeOID: pgtype.Int8OID},
	}
	targets := scanTargets(fields)
	w := driver.NewWindow(columnsFromFields(fields), 2)

	err := scanRowInto(w, 0, targets, func(dest ...any) error {
		*dest[0].(*int64) = 849
		*dest[1].(*int64) = 743
		dest[2].(*pgtype.Timestamp).Time = mustParseTime("2026-09-08 11:17:37")
		dest[2].(*pgtype.Timestamp).Valid = true
		dest[3].(*pgtype.Text).Valid = false
		*dest[4].(*int64) = 743
		dest[5].(*pgtype.Timestamp).Time = mustParseTime("2026-06-11 10:24:23")
		dest[5].(*pgtype.Timestamp).Valid = true
		*dest[6].(*int64) = 1
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = scanRowInto(w, 1, targets, func(dest ...any) error {
		*dest[0].(*int64) = 851
		*dest[1].(*int64) = 1
		dest[2].(*pgtype.Timestamp).Time = mustParseTime("2026-06-12 16:43:16")
		dest[2].(*pgtype.Timestamp).Valid = true
		dest[3].(*pgtype.Text).String = "Sanity_116ABC@gmail.com"
		dest[3].(*pgtype.Text).Valid = true
		*dest[4].(*int64) = 1
		dest[5].(*pgtype.Timestamp).Time = mustParseTime("2026-06-11 10:24:24")
		dest[5].(*pgtype.Timestamp).Valid = true
		*dest[6].(*int64) = 1
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	w.Rows = 2

	if got, ok := w.CellText(2, 0); !ok || got != "2026-09-08 11:17:37" {
		t.Fatalf("updated_at row0 = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(2, 1); !ok || got != "2026-06-12 16:43:16" {
		t.Fatalf("updated_at row1 = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(3, 0); ok {
		t.Fatalf("official_email row0 should be null, got %q", got)
	}
	if got, ok := w.CellText(3, 1); !ok || got != "Sanity_116ABC@gmail.com" {
		t.Fatalf("official_email row1 = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(5, 1); !ok || got != "2026-06-11 10:24:24" {
		t.Fatalf("created_at row1 = %q ok=%v", got, ok)
	}

	raw, err := w.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if contains := string(raw); len(contains) > 0 {
		if idx := findSubstr(contains, "Sanity_116ABC@gmail.com2026"); idx >= 0 {
			t.Fatalf("JSON concatenates email into next timestamp: ...%s...", contains[idx:idx+60])
		}
		if idx := findSubstr(contains, "16:43:16Sanity"); idx >= 0 {
			t.Fatalf("JSON concatenates updated_at and email in one column value")
		}
	}
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		panic(err)
	}
	return t
}

func findSubstr(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
