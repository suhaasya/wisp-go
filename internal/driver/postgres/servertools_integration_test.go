//go:build integration

package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
)

func TestProcessListBlockedHighlight(t *testing.T) {
	holder := openConn(t, 0)
	execFixture(t, holder, "ci/fixtures/pg_blocking_locks.sql")

	waiter := openConn(t, 0)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := holder.Exec(ctx, "BEGIN")
		if err != nil {
			t.Errorf("begin: %v", err)
			return
		}
		_, err = holder.Exec(ctx, "SELECT id FROM wisp_lock.hold WHERE id = 1 FOR UPDATE")
		if err != nil {
			t.Errorf("lock: %v", err)
			return
		}
		time.Sleep(3 * time.Second)
		_, _ = holder.Exec(ctx, "ROLLBACK")
	}()

	time.Sleep(200 * time.Millisecond)
	blockDone := make(chan struct{})
	go func() {
		defer close(blockDone)
		_, _ = waiter.Exec(ctx, "UPDATE wisp_lock.hold SET id = 1 WHERE id = 1")
	}()

	time.Sleep(500 * time.Millisecond)
	rows, err := holder.Query(ctx, postgres.ProcessListSQL())
	if err != nil {
		t.Fatalf("process list: %v", err)
	}
	defer rows.Close()
	foundBlocked := false
	for {
		w, err := rows.NextWindow(50)
		if err != nil {
			t.Fatalf("window: %v", err)
		}
		if w == nil || w.Rows == 0 {
			break
		}
		highlightCol := -1
		for i, c := range w.Columns {
			if c.Name == "wisp_highlight" {
				highlightCol = i
				break
			}
		}
		if highlightCol < 0 {
			t.Fatal("missing wisp_highlight column")
		}
		for r := 0; r < w.Rows; r++ {
			if text, ok := w.CellText(highlightCol, r); ok && text == "blocked" {
				foundBlocked = true
			}
		}
	}
	if !foundBlocked {
		t.Fatal("expected at least one blocked process row")
	}
	_, _ = holder.Exec(ctx, "ROLLBACK")
	<-blockDone
	wg.Wait()
}
