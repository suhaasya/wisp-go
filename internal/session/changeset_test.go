package session

import "testing"

func ptr(s string) *string { return &s }

func TestChangeSetEditAfterInsertUpdatesInsert(t *testing.T) {
	cs := NewChangeSet([]string{"id"})
	if err := cs.Apply(Change{Op: "insert"}); err != nil {
		t.Fatal(err)
	}
	insKey := cs.LastInsertKey()
	if insKey == "" {
		t.Fatal("expected insert key")
	}
	row := []any{insKey}
	if err := cs.Apply(Change{Op: "set", Row: row, Col: "name", Value: ptr("A")}); err != nil {
		t.Fatal(err)
	}
	if err := cs.Apply(Change{Op: "set", Row: row, Col: "name", Value: ptr("B")}); err != nil {
		t.Fatal(err)
	}
	if cs.Count() != 1 {
		t.Fatalf("count = %d, want 1", cs.Count())
	}
	v, ok := cs.InsertCell(insKey, "name")
	if !ok || v != "B" {
		t.Fatalf("insert name = %q ok=%v, want B", v, ok)
	}
}

func TestChangeSetDeleteAfterInsertRemovesInsert(t *testing.T) {
	cs := NewChangeSet([]string{"id"})
	if err := cs.Apply(Change{Op: "insert"}); err != nil {
		t.Fatal(err)
	}
	key := cs.LastInsertKey()
	if err := cs.Apply(Change{Op: "delete", Row: []any{key}}); err != nil {
		t.Fatal(err)
	}
	if cs.Count() != 0 {
		t.Fatalf("count = %d, want 0", cs.Count())
	}
}

func TestChangeSetEditDeletedRowRefused(t *testing.T) {
	cs := NewChangeSet([]string{"id"})
	row := []any{int64(42)}
	if err := cs.Apply(Change{Op: "delete", Row: row}); err != nil {
		t.Fatal(err)
	}
	err := cs.Apply(Change{Op: "set", Row: row, Col: "name", Value: ptr("x"), Original: ptr("y")})
	if err == nil {
		t.Fatal("expected error editing deleted row")
	}
}

func TestChangeSetRevertToOriginalRemovesUpdate(t *testing.T) {
	cs := NewChangeSet([]string{"id"})
	row := []any{int64(1)}
	orig := "alpha"
	if err := cs.Apply(Change{Op: "set", Row: row, Col: "name", Value: ptr("beta"), Original: &orig}); err != nil {
		t.Fatal(err)
	}
	if cs.Count() != 1 {
		t.Fatalf("count = %d, want 1", cs.Count())
	}
	if err := cs.Apply(Change{Op: "set", Row: row, Col: "name", Value: &orig, Original: &orig}); err != nil {
		t.Fatal(err)
	}
	if cs.Count() != 0 {
		t.Fatalf("count = %d, want 0 after revert", cs.Count())
	}
}
