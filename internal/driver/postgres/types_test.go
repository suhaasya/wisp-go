package postgres

import (
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestDecodeFromValueFillsSeparateWindowColumnsManyRows(t *testing.T) {
	meta := []driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "official_email", Type: "text"},
		{Name: "created_at", Type: "timestamp"},
	}
	w := driver.NewWindow(meta, 150)
	for r := 0; r < 27; r++ {
		decodeFromValue(w.Column(1), time.Date(2026, 9, 8, 11, 17, 37, 0, time.UTC), r)
		decodeFromValue(w.Column(2), "user@example.com", r)
		decodeFromValue(w.Column(3), time.Date(2026, 6, 11, 10, 24, 23, 0, time.UTC), r)
	}
	w.Rows = 27
	if got, ok := w.CellText(1, 15); !ok || got == "" {
		t.Fatalf("updated_at row15 = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(2, 20); !ok || got != "user@example.com" {
		t.Fatalf("email row20 = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(3, 16); !ok || len(got) > 40 {
		t.Fatalf("created_at row16 = %q ok=%v", got, ok)
	}
}

func TestDecodeFromValueMultipleTextLikeColumns(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "first_name", Type: "text"},
		{Name: "last_name", Type: "text"},
	}, 1)

	ts1 := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	ts2 := time.Date(2024, 3, 2, 15, 30, 0, 0, time.UTC)
	decodeFromValue(w.Column(0), ts1, 0)
	decodeFromValue(w.Column(1), ts2, 0)
	decodeFromValue(w.Column(2), "Akshay", 0)
	decodeFromValue(w.Column(3), "Smith", 0)
	w.Rows = 1

	if got, ok := w.CellText(1, 0); !ok || got == "" {
		t.Fatalf("updated_at empty ok=%v got=%q", ok, got)
	}
	if got, ok := w.CellText(2, 0); !ok || got != "Akshay" {
		t.Fatalf("first_name = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(3, 0); !ok || got != "Smith" {
		t.Fatalf("last_name = %q ok=%v", got, ok)
	}
}
