package session

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestResultOverlaySkipsQueryResults(t *testing.T) {
	main, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: 1,
			Columns:   []driver.ColumnMeta{{Name: "id", Type: "int4"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = main.Close() })

	s := NewTestSession(store.Profile{Engine: "mock"})
	s.main = main
	s.state = StateReady
	p := s.ensurePager()

	open, err := p.Open(t.Context(), pager.Source{
		Kind: pager.SourceQuery,
		SQL:  "SELECT 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	s.registerResultChanges(open.RID, []string{"id"}, "public", "user")
	rc, _ := s.resultChangesFor(open.RID)
	empty := ""
	if err := rc.cs.Apply(Change{
		Op: "set", Row: []any{int64(1)}, Col: "last_name", Value: &empty, Original: ptr("x"),
	}); err != nil {
		t.Fatal(err)
	}

	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "last_name", Type: "text"},
	}, 1)
	w.Column(0).SetInt(0, 1)
	w.Column(1).SetText(0, "from-db")
	w.Rows = 1

	ov := s.resultOverlay(open.RID, w, 0)
	if len(ov.Mod) != 0 {
		t.Fatalf("mod = %#v, want none for query results", ov.Mod)
	}
	got, ok := w.CellText(1, 0)
	if !ok || got != "from-db" {
		t.Fatalf("last_name = %q ok=%v, want from-db unchanged", got, ok)
	}
}
