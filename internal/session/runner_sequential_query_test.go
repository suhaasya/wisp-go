package session

import (
	"context"
	"sync"
	"testing"
)

func TestTwoSequentialSelectRunsBothComplete(t *testing.T) {
	var logMu sync.Mutex
	var log []string
	s := mockSessionConn(t, &log, &logMu)
	var evMu sync.Mutex
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		evMu.Lock()
		events = append(events, ev)
		evMu.Unlock()
	})

	qid1, err := s.StartRun(context.Background(), RunRequest{
		SQL:  "SELECT 1",
		Mode: RunModeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &evMu, &events, qid1)

	qid2, err := s.StartRun(context.Background(), RunRequest{
		SQL:  "SELECT 2",
		Mode: RunModeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &evMu, &events, qid2)

	evMu.Lock()
	defer evMu.Unlock()
	var rids []string
	for _, ev := range events {
		if ev.QID == qid2 && ev.RID != "" {
			rids = append(rids, ev.RID)
		}
	}
	if len(rids) != 1 {
		t.Fatalf("second run rids=%v events=%+v", rids, events)
	}
	if len(s.QueryRunResultIDs()) == 0 {
		t.Fatal("expected query run rids tracked")
	}
}

func TestTwoSequentialSelectRunsLoadWindows(t *testing.T) {
	var logMu sync.Mutex
	s := mockSessionConn(t, nil, &logMu)
	ctx := context.Background()
	var evMu sync.Mutex
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		evMu.Lock()
		events = append(events, ev)
		evMu.Unlock()
	})

	qid1, err := s.StartRun(ctx, RunRequest{SQL: "SELECT 1", Mode: RunModeAll})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &evMu, &events, qid1)
	rid1 := resultRID(t, &evMu, &events, qid1)
	if _, err := s.ResultWindowJSON(ctx, rid1, 0, 150); err != nil {
		t.Fatalf("first window: %v", err)
	}

	qid2, err := s.StartRun(ctx, RunRequest{SQL: "SELECT 2", Mode: RunModeAll})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &evMu, &events, qid2)
	rid2 := resultRID(t, &evMu, &events, qid2)
	if _, err := s.ResultWindowJSON(ctx, rid2, 0, 150); err != nil {
		t.Fatalf("second window: %v", err)
	}
	if _, err := s.ResultWindowJSON(ctx, rid1, 0, 150); err != nil {
		t.Fatalf("first window after second run: %v", err)
	}
}

func resultRID(t *testing.T, mu *sync.Mutex, events *[]QueryRunEvent, qid string) string {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	for _, ev := range *events {
		if ev.QID == qid && ev.RID != "" {
			return ev.RID
		}
	}
	t.Fatalf("no rid for qid %q in %+v", qid, *events)
	return ""
}
