//go:build integration

package session

import (
	"context"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/store"
)

func openPGRun(t *testing.T) driver.Conn {
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

func TestRunScriptMultiStatementPostgres(t *testing.T) {
	conn := openPGRun(t)
	defer func() { _ = conn.Close() }()
	ctx := context.Background()
	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.main = conn
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		events = append(events, ev)
	})
	qid, err := s.StartRun(ctx, RunRequest{
		SQL:  "SELECT 1 AS a; SELECT 2 AS b;",
		Mode: RunModeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &events, qid)
	rids := 0
	for _, ev := range events {
		if ev.QID == qid && ev.RID != "" {
			rids++
		}
	}
	if rids != 2 {
		t.Fatalf("rids=%d events=%+v", rids, events)
	}
}

func TestCancelRunPostgresSleep(t *testing.T) {
	conn := openPGRun(t)
	defer func() { _ = conn.Close() }()
	ctx := context.Background()
	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.main = conn
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		events = append(events, ev)
	})
	start := time.Now()
	qid, err := s.StartRun(ctx, RunRequest{
		SQL:  "SELECT pg_sleep(30)",
		Mode: RunModeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := s.CancelRun(qid); err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &events, qid)
	elapsed := time.Since(start)
	if elapsed > 500*time.Millisecond {
		t.Fatalf("cancel took %v want <= 500ms", elapsed)
	}
	cancelled := false
	for _, ev := range events {
		if ev.QID == qid && ev.Cancelled {
			cancelled = true
		}
	}
	if !cancelled {
		t.Fatalf("events=%+v", events)
	}
}

func TestRunFiftyStatementsRSSPostgres(t *testing.T) {
	conn := openPGRun(t)
	defer func() { _ = conn.Close() }()
	ctx := context.Background()
	s := NewTestSession(store.Profile{Engine: "postgres", Env: "local"})
	s.main = conn
	sql := ""
	for i := 0; i < 50; i++ {
		if i > 0 {
			sql += ";"
		}
		sql += "SELECT 1"
	}
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		events = append(events, ev)
	})
	qid, err := s.StartRun(ctx, RunRequest{SQL: sql, Mode: RunModeAll})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &events, qid)
	if len(s.QueryRunResultIDs()) > 5 {
		t.Fatalf("too many open results: %d", len(s.QueryRunResultIDs()))
	}
}
