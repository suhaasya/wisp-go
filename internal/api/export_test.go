package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestGetExportStreamsCSV(t *testing.T) {
	const total = 12
	s := session.NewTestSession(store.Profile{Engine: "mock"})
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: total,
			Columns: []driver.ColumnMeta{
				{Name: "id", Type: "int8"},
				{Name: "note", Type: "text"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s.AdoptMainConn(conn)
	open, err := s.OpenResult(t.Context(), pager.Source{
		Kind: pager.SourceQuery,
		SQL:  "SELECT id, note FROM items",
	})
	if err != nil {
		t.Fatal(err)
	}

	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	mgr.AdoptSession(s)
	h := api.NewSessionsHandler(mgr, nil)

	url := "/api/v1/sessions/test/export?format=csv&scope=result&rid=" + open.RID
	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("disposition=%q", rec.Header().Get("Content-Disposition"))
	}
	body := rec.Body.String()
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	if len(lines) != total+1 {
		t.Fatalf("lines=%d want %d header+rows", len(lines), total+1)
	}
	if !strings.HasPrefix(lines[0], "id,") {
		t.Fatalf("header=%q", lines[0])
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
}

func (flushRecorder) Flush() {}
