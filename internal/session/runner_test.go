package session

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

func mockSessionConn(t *testing.T, execLog *[]string, execLogMu *sync.Mutex) *Session {
	t.Helper()
	cfg := mock.Config{
		TotalRows: 3,
		Columns: []driver.ColumnMeta{
			{Name: "id", Type: "int8"},
			{Name: "name", Type: "text"},
		},
		ExecLog:   execLog,
		ExecLogMu: execLogMu,
	}
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	s := NewTestSession(store.Profile{Engine: "mock", Env: "local"})
	s.main = conn
	return s
}

func waitQueryDone(t *testing.T, mu *sync.Mutex, events *[]QueryRunEvent, qid string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		done := false
		for _, ev := range *events {
			if ev.QID == qid && ev.Done {
				done = true
				break
			}
		}
		mu.Unlock()
		if done {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("run %q did not finish", qid)
}

func TestRunScriptSequentialMock(t *testing.T) {
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

	sql := "SELECT 1;\nUPDATE t SET x = 1;\nSELECT 2"
	qid, err := s.StartRun(context.Background(), RunRequest{
		SQL:  sql,
		Mode: RunModeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &evMu, &events, qid)

	evMu.Lock()
	var rids int
	for _, ev := range events {
		if ev.QID != qid {
			continue
		}
		if ev.RID != "" {
			rids++
		}
	}
	evMu.Unlock()
	if rids != 2 {
		t.Fatalf("result tabs=%d want 2 events=%+v", rids, events)
	}
	logMu.Lock()
	nLog := len(log)
	logCopy := append([]string(nil), log...)
	logMu.Unlock()
	if nLog != 1 {
		t.Fatalf("exec log=%v want one UPDATE", logCopy)
	}
}
