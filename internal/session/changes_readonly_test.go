package session

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
)

func TestApplyResultChangeReadOnly(t *testing.T) {
	s := NewTestSession(store.Profile{ReadOnly: true, Engine: "postgres"})
	s.registerResultChanges("rid-1", []string{"id"})
	v := "x"
	_, err := s.ApplyResultChange("rid-1", Change{Op: "set", Row: []any{1}, Col: "name", Value: &v})
	if !IsWriteBlockedReadOnly(err) {
		t.Fatalf("err=%v want read-only", err)
	}
}
