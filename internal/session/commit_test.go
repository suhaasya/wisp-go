package session

import (
	"context"
	"errors"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestCommitResultChangesRollsBackOnMidBatchError(t *testing.T) {
	var log []string
	cfg := mock.Config{
		Columns:      []driver.ColumnMeta{{Name: "id", Type: "int8"}, {Name: "name", Type: "text"}},
		TxFailOnExec: 2,
		ExecLog:      &log,
	}
	conn, err := mock.Driver{}.Open(context.Background(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.main = conn

	s.changeMu.Lock()
	s.resultChanges = map[string]*resultChanges{
		"rid-1": {
			cs:       NewChangeSet([]string{"id"}),
			schema:   "public",
			table:    "orders",
			editable: true,
		},
	}
	rc := s.resultChanges["rid-1"]
	s.changeMu.Unlock()

	v1 := "a"
	v2 := "b"
	orig1 := "x"
	orig2 := "y"
	if err := rc.cs.Apply(Change{Op: "set", Row: []any{int64(1)}, Col: "name", Value: &v1, Original: &orig1}); err != nil {
		t.Fatal(err)
	}
	if err := rc.cs.Apply(Change{Op: "set", Row: []any{int64(2)}, Col: "name", Value: &v2, Original: &orig2}); err != nil {
		t.Fatal(err)
	}

	_, err = s.CommitResultChanges(context.Background(), "rid-1", CommitOptions{Confirm: true})
	var ce *CommitError
	if !errors.As(err, &ce) {
		t.Fatalf("err=%v want CommitError", err)
	}
	if ce.StatementIndex != 1 {
		t.Fatalf("index=%d want 1", ce.StatementIndex)
	}
	if len(log) != 2 {
		t.Fatalf("exec log=%v want 2 statements before fail", log)
	}
}

func TestCommitOptimisticConflict(t *testing.T) {
	var log []string
	cfg := mock.Config{
		Columns:        []driver.ColumnMeta{{Name: "id", Type: "int8"}, {Name: "name", Type: "text"}},
		ExecLog:        &log,
		TxRowsAffected: map[int]int64{1: 0},
	}
	conn, err := mock.Driver{}.Open(context.Background(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.main = conn
	s.changeMu.Lock()
	s.resultChanges = map[string]*resultChanges{
		"rid-1": {
			cs:       NewChangeSet([]string{"id"}),
			schema:   "public",
			table:    "orders",
			editable: true,
		},
	}
	rc := s.resultChanges["rid-1"]
	s.changeMu.Unlock()

	v := "new"
	orig := "old"
	if err := rc.cs.Apply(Change{Op: "set", Row: []any{int64(5)}, Col: "name", Value: &v, Original: &orig}); err != nil {
		t.Fatal(err)
	}
	_, err = s.CommitResultChanges(context.Background(), "rid-1", CommitOptions{Confirm: true})
	var ce *CommitError
	if !errors.As(err, &ce) || !ce.Conflict {
		t.Fatalf("err=%v want conflict CommitError", err)
	}
}
