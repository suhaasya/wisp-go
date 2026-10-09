//go:build integration

package session

import (
	"context"
	"errors"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestCommitRollbackPostgres(t *testing.T) {
	conn := openPGCommit(t)
	defer func() { _ = conn.Close() }()
	ctx := context.Background()
	_, err := conn.Exec(ctx, `DROP TABLE IF EXISTS wisp_commit_test`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, `CREATE TABLE wisp_commit_test (id int PRIMARY KEY, val text NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, `INSERT INTO wisp_commit_test VALUES (1, 'a'), (2, 'b')`)
	if err != nil {
		t.Fatal(err)
	}

	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.main = conn
	s.registerResultChanges("r1", []string{"id"}, "public", "wisp_commit_test")

	v1, v2 := "aa", "bb"
	orig1, orig2 := "a", "b"
	if _, err := s.ApplyResultChange("r1", Change{Op: "set", Row: []any{1}, Col: "val", Value: &v1, Original: &orig1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ApplyResultChange("r1", Change{Op: "set", Row: []any{2}, Col: "val", Value: &v2, Original: &orig2}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ApplyResultChange("r1", Change{Op: "insert"}); err != nil {
		t.Fatal(err)
	}
	key := s.ResultLastInsertKey("r1")
	dup := "1"
	valZ := "z"
	if _, err := s.ApplyResultChange("r1", Change{Op: "set", Row: []any{key}, Col: "id", Value: &dup}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ApplyResultChange("r1", Change{Op: "set", Row: []any{key}, Col: "val", Value: &valZ}); err != nil {
		t.Fatal(err)
	}

	_, err = s.CommitResultChanges(ctx, "r1", CommitOptions{Confirm: true})
	if err == nil {
		t.Fatal("expected commit error")
	}
	var ce *CommitError
	if !errors.As(err, &ce) {
		t.Fatalf("err=%v", err)
	}

	rows, err := conn.Query(ctx, `SELECT val FROM wisp_commit_test WHERE id = 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	w, err := rows.NextWindow(1)
	if err != nil || w == nil || w.Rows != 1 {
		t.Fatal("expected row 1")
	}
	got, _ := w.CellString(0, 0)
	if got != "a" {
		t.Fatalf("val=%q want a unchanged after rollback", got)
	}
}

func openPGCommit(t *testing.T) driver.Conn {
	t.Helper()
	var d postgres.Driver
	conn, err := d.Open(t.Context(), driver.ConnectParams{
		Host: "127.0.0.1", Port: 54329, DB: "wisp_bench", User: "wisp",
		Password: []byte("wisp"),
	})
	if err != nil {
		t.Skip(err)
	}
	return conn
}
