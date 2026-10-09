package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestPostDDLPreviewDropWithoutCascade(t *testing.T) {
	s := session.NewTestSession(store.Profile{Engine: "postgres"})
	conn, _ := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: 0, Columns: []driver.ColumnMeta{{Name: "x", Type: "text"}}},
	})
	s.AdoptMainConn(conn)
	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	mgr.AdoptSession(s)
	h := api.NewSessionsHandler(mgr, nil)

	body, _ := json.Marshal(map[string]any{
		"action": "drop_table",
		"schema": "wisp_meta",
		"table":  "parent",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/sessions/test/ddl/preview", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		SQL string `json:"sql"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToUpper(out.SQL), "CASCADE") {
		t.Fatalf("sql=%q", out.SQL)
	}
}
