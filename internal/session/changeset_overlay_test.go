package session

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestChangeSetApplyToWindowModifiedCell(t *testing.T) {
	cols := []driver.ColumnMeta{{Name: "id", Type: "int8"}, {Name: "name", Type: "text"}}
	w := driver.NewWindow(cols, 2)
	w.Column(0).SetInt(0, 10)
	w.Column(0).SetInt(1, 11)
	w.Column(1).SetText(0, "a")
	w.Column(1).SetText(1, "b")

	cs := NewChangeSet([]string{"id"})
	row := []any{int64(11)}
	if err := cs.Apply(Change{Op: "set", Row: row, Col: "name", Value: ptr("beta"), Original: ptr("b")}); err != nil {
		t.Fatal(err)
	}
	ov := cs.ApplyToWindow(w, 0, []int{0})
	if len(ov.Mod) != 1 || ov.Mod[0][0] != 1 || ov.Mod[0][1] != 1 {
		t.Fatalf("mod = %#v, want [[1 1]]", ov.Mod)
	}
	got, ok := w.CellText(1, 1)
	if !ok || got != "beta" {
		t.Fatalf("cell = %q ok=%v, want beta", got, ok)
	}
}
