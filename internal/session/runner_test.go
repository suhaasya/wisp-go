package session

import (
	"context"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

func mockSessionConn(t *testing.T, execLog *[]string) *Session {
	t.Helper()
	cfg := mock.Config{
		TotalRows: 3,
		Columns: []driver.ColumnMeta{
			{Name: "id", Type: "int8"},
			{Name: "name", Type: "text"},
		},
		ExecLog: execLog,
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

func waitQueryDone(t *testing.T, events *[]QueryRunEvent, qid string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, ev := range *events {
			if ev.QID == qid && ev.Done {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("run %q did not finish", qid)
}

func TestRunScriptSequentialMock(t *testing.T) {
	var log []string
	s := mockSessionConn(t, &log)
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		events = append(events, ev)
	})

	sql := "SELECT 1;\nUPDATE t SET x = 1;\nSELECT 2"
	qid, err := s.StartRun(context.Background(), RunRequest{
		SQL:  sql,
		Mode: RunModeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &events, qid)

	var rids int
	for _, ev := range events {
		if ev.QID != qid {
			continue
		}
		if ev.RID != "" {
			rids++
		}
	}
	if rids != 2 {
		t.Fatalf("result tabs=%d want 2 events=%+v", rids, events)
	}
	if len(log) != 1 {
		t.Fatalf("exec log=%v want one UPDATE", log)
	}
}
