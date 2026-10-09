package pager_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/suhaasya/wisp-go/internal/pager"
)

func referenceIDs(total, offset, limit int) []int64 {
	if offset >= total {
		return nil
	}
	n := limit
	if remain := total - offset; remain < n {
		n = remain
	}
	out := make([]int64, n)
	for i := range out {
		out[i] = int64(offset + i)
	}
	return out
}

func windowIDs(t *testing.T, m *pager.Manager, rid string, offset, limit int) []int64 {
	t.Helper()
	raw, err := m.WindowJSON(t.Context(), rid, offset, limit)
	if err != nil {
		t.Fatalf("WindowJSON: %v", err)
	}
	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(body.Data) == 0 {
		return nil
	}
	var col0 []int64
	if err := json.Unmarshal(body.Data[0], &col0); err != nil {
		t.Fatalf("col0: %v", err)
	}
	return col0
}

func TestPropertyRandomWindowsMatchReference(t *testing.T) {
	const total = 100_000
	m := mockManager(t, total)
	m.SetMockTotal(total)
	ctx := t.Context()
	open, err := m.Open(ctx, pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "t",
		KeyColumns: []string{"id"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(99))
	for i := 0; i < 200; i++ {
		offset := rng.Intn(total)
		limit := 1 + rng.Intn(150)
		want := referenceIDs(total, offset, limit)
		got := windowIDs(t, m, open.RID, offset, limit)
		if len(got) != len(want) {
			t.Fatalf("trial %d offset=%d limit=%d: len got=%d want=%d", i, offset, limit, len(got), len(want))
		}
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("trial %d row %d: got %d want %d", i, j, got[j], want[j])
			}
		}
	}
}
