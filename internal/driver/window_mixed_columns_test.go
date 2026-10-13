package driver_test

import (
	"encoding/json"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestWindowTextColumnsDoNotShareRowSlots(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "created_by", Type: "int4"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "official_email", Type: "text"},
		{Name: "last_modified_by", Type: "int4"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "active", Type: "int8"},
	}, 2)
	w.Column(2).SetText(0, "2026-09-08 11:17:37")
	w.Column(5).SetText(0, "2026-06-11 10:24:23")
	w.Column(3).SetText(0, "user@example.com")
	if got, ok := w.CellText(2, 0); !ok || got != "2026-09-08 11:17:37" {
		t.Fatalf("updated_at = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(3, 0); !ok || got != "user@example.com" {
		t.Fatalf("official_email = %q ok=%v", got, ok)
	}
	if got, ok := w.CellText(5, 0); !ok || got != "2026-06-11 10:24:23" {
		t.Fatalf("created_at = %q ok=%v", got, ok)
	}
}

func TestWindowMarshalManyRowsInLargeCapacity(t *testing.T) {
	const cap = 150
	const n = 27
	meta := []driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "created_by", Type: "int4"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "official_email", Type: "text"},
		{Name: "last_modified_by", Type: "int4"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "active", Type: "int8"},
	}
	w := driver.NewWindow(meta, cap)
	for r := 0; r < n; r++ {
		w.Column(2).SetText(r, "2026-09-08 11:17:37")
		w.Column(3).SetText(r, "user@example.com")
		w.Column(5).SetText(r, "2026-06-11 10:24:23")
	}
	w.Rows = n
	raw, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Data [][]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Data[2][15] != "2026-09-08 11:17:37" {
		t.Fatalf("updated_at row15 = %v", doc.Data[2][15])
	}
	if doc.Data[3][20] != "user@example.com" {
		t.Fatalf("email row20 = %v", doc.Data[3][20])
	}
	got5, _ := doc.Data[5][16].(string)
	if len(got5) > 30 || got5 != "2026-06-11 10:24:23" {
		t.Fatalf("created_at row16 = %q", got5)
	}
}

func TestWindowMarshalMixedIntAndMultipleTimestamps(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "created_by", Type: "int4"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "official_email", Type: "text"},
		{Name: "last_modified_by", Type: "int4"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "active", Type: "int8"},
	}, 1)
	w.Column(0).SetInt(0, 849)
	w.Column(1).SetInt(0, 743)
	w.Column(2).SetText(0, "2026-09-08 11:17:30")
	w.Column(3).SetText(0, "user@example.com")
	w.Column(4).SetInt(0, 743)
	w.Column(5).SetText(0, "2026-03-01 10:00:00")
	w.Column(6).SetInt(0, 1)
	w.Rows = 1

	got, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Data [][]any `json:"data"`
	}
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Data) != 7 {
		t.Fatalf("columns = %d", len(doc.Data))
	}
	if doc.Data[3][0] != "user@example.com" {
		t.Fatalf("official_email = %v", doc.Data[3][0])
	}
	if doc.Data[5][0] != "2026-03-01 10:00:00" {
		t.Fatalf("created_at = %v", doc.Data[5][0])
	}
}
