package session_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestStreamExportResultScopeRowCount(t *testing.T) {
	const total = 100
	s := session.NewTestSession(store.Profile{Engine: "mock"})
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: total,
			Columns:   []driver.ColumnMeta{{Name: "id", Type: "int8"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s.AdoptMainConn(conn)
	open, err := s.OpenResult(t.Context(), pager.Source{
		Kind: pager.SourceQuery,
		SQL:  "SELECT id FROM t",
	})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	n, err := s.StreamExport(t.Context(), session.ExportRequest{
		Format: "csv",
		Scope:  session.ExportScopeResult,
		RID:    open.RID,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != total {
		t.Fatalf("rows=%d want %d", n, total)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != total+1 {
		t.Fatalf("lines=%d", len(lines))
	}
}
