package session

import (
	"context"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestPreviewDropTableNoCascadeByDefault(t *testing.T) {
	s := sessionWithMockConn(t)
	out, err := s.PreviewDDL(context.Background(), DDLPreviewRequest{
		Action: DDLDropTable, Schema: "wisp_meta", Table: "parent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToUpper(out.SQL), "CASCADE") {
		t.Fatalf("sql=%q", out.SQL)
	}
}

func TestExecuteDropWrongConfirmRejected(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres", Env: "production"})
	var log []string
	conn, _ := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: 0,
			Columns:   []driver.ColumnMeta{{Name: "x", Type: "text"}},
			ExecLog:   &log,
		},
	})
	s.AdoptMainConn(conn)
	err := s.ExecuteDDL(context.Background(), DDLPreviewRequest{
		Action: DDLDropTable, Schema: "public", Table: "orders",
	}, DDLExecuteOptions{Confirm: true, ObjectConfirm: "wrong"})
	if err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("err=%v log=%v", err, log)
	}
	if len(log) != 0 {
		t.Fatalf("exec called: %v", log)
	}
}

func sessionWithMockConn(t *testing.T) *Session {
	t.Helper()
	s := NewTestSession(store.Profile{Engine: "postgres"})
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: 0,
			Columns:   []driver.ColumnMeta{{Name: "x", Type: "text"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s.AdoptMainConn(conn)
	return s
}
