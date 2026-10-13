package driver_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestCloneWindowPreservesMultipleTextColumns(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "official_email", Type: "text"},
		{Name: "created_at", Type: "timestamp"},
	}, 2)
	w.Column(0).SetInt(0, 1)
	w.Column(0).SetInt(1, 2)
	w.Column(1).SetText(0, "2026-09-08 11:17:30")
	w.Column(1).SetText(1, "2026-09-09 12:00:00")
	w.Column(2).SetText(0, "a@b.c")
	w.Column(2).SetText(1, "c@d.e")
	w.Column(3).SetText(0, "2026-01-01 00:00:00")
	w.Column(3).SetText(1, "2026-02-02 00:00:00")
	w.Rows = 2

	c := driver.CloneWindow(w)
	if got, ok := c.CellText(2, 1); !ok || got != "c@d.e" {
		t.Fatalf("email row1 = %q ok=%v", got, ok)
	}
	if got, ok := c.CellText(3, 0); !ok || got != "2026-01-01 00:00:00" {
		t.Fatalf("created_at row0 = %q ok=%v", got, ok)
	}
}
