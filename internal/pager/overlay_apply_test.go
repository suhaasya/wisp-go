package pager_test

import (
	"encoding/json"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/session"
)

func TestWindowJSONStagedOverlay(t *testing.T) {
	m := mockManager(t, 100)
	ctx := t.Context()
	open, err := m.Open(ctx, sourceTableFixture())
	if err != nil {
		t.Fatal(err)
	}
	cs := session.NewChangeSet([]string{"id"})
	orig := "row-1"
	newVal := "staged-name"
	if err := cs.Apply(session.Change{
		Op: "set", Row: []any{float64(1)}, Col: "name", Value: &newVal, Original: &orig,
	}); err != nil {
		t.Fatal(err)
	}
	m.SetOverlayProvider(func(rid string, w *driver.Window, offset int) pager.Overlay {
		if rid != open.RID {
			return pager.Overlay{}
		}
		return cs.ApplyToWindow(w, offset, []int{0})
	})
	raw, err := m.WindowJSON(ctx, open.RID, 0, 150)
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Overlay struct {
			Mod [][]int `json:"mod"`
		} `json:"overlay"`
		Data [][]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Overlay.Mod) != 1 || body.Overlay.Mod[0][0] != 1 || body.Overlay.Mod[0][1] != 1 {
		t.Fatalf("overlay mod = %#v", body.Overlay.Mod)
	}
	if body.Data[1][1] != "staged-name" {
		t.Fatalf("name = %v", body.Data[1][1])
	}
}

func sourceTableFixture() pager.Source {
	return pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "t",
		KeyColumns: []string{"id"},
	}
}
