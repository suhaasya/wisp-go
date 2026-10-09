package session_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/mover"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestStartImportReadOnlyBlocked(t *testing.T) {
	s := session.NewTestSession(store.Profile{Engine: "mock", ReadOnly: true})
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: 0, Columns: []driver.ColumnMeta{{Name: "x", Type: "text"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s.AdoptMainConn(conn)
	dir := t.TempDir()
	path := filepath.Join(dir, "t.csv")
	if err := os.WriteFile(path, []byte("id\n1\n2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan session.ImportEvent, 1)
	s.SetImportEventHook(func(ev session.ImportEvent) {
		if ev.Done {
			done <- ev
		}
	})
	_, err = s.StartImport(t.Context(), session.ImportStartRequest{
		Path:        path,
		Kind:        "csv",
		Schema:      "public",
		Table:       "t",
		Mapping:     []session.ImportColumnMapping{{FileColumn: "id", TableColumn: "id"}},
		FileColumns: []string{"id"},
		BatchSize:   1,
	})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-done:
		if ev.Error == "" {
			t.Fatalf("expected error, got rows=%d", ev.Rows)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestPreviewImportCSV(t *testing.T) {
	s := session.NewTestSession(store.Profile{Engine: "mock"})
	dir := t.TempDir()
	path := filepath.Join(dir, "t.csv")
	if err := os.WriteFile(path, []byte("a\n1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prev, err := s.PreviewImport(t.Context(), session.ImportPreviewRequest{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	if prev.Kind != mover.ImportKindCSV {
		t.Fatalf("kind=%q", prev.Kind)
	}
}
