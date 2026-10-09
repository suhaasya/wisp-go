package session

import (
	"context"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestRunScriptCapsQueryResultsAtFive(t *testing.T) {
	cfg := mock.Config{
		TotalRows: 1,
		Columns: []driver.ColumnMeta{
			{Name: "id", Type: "int8"},
		},
	}
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	s := NewTestSession(store.Profile{Engine: "mock"})
	s.main = conn

	sql := ""
	for i := 0; i < 6; i++ {
		if i > 0 {
			sql += ";"
		}
		sql += "SELECT 1"
	}
	var events []QueryRunEvent
	s.SetQueryEventHook(func(ev QueryRunEvent) {
		events = append(events, ev)
	})
	qid, err := s.StartRun(context.Background(), RunRequest{SQL: sql, Mode: RunModeAll})
	if err != nil {
		t.Fatal(err)
	}
	waitQueryDone(t, &events, qid)
	ids := s.QueryRunResultIDs()
	if len(ids) != 5 {
		t.Fatalf("kept %d rids want 5: %v", len(ids), ids)
	}
}
