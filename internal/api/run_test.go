package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestPostRunReturnsQID(t *testing.T) {
	cfg := mock.Config{
		TotalRows: 1,
		Columns:   []driver.ColumnMeta{{Name: "id", Type: "int8"}},
	}
	conn, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	s := session.NewTestSession(store.Profile{Engine: "mock"})
	s.AdoptMainConn(conn)
	mgr.AdoptSession(s)

	h := api.NewSessionsHandler(mgr, nil)
	body, _ := json.Marshal(map[string]any{
		"sql":  "SELECT 1",
		"mode": "all",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/sessions/test/run", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		QID string `json:"qid"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.QID == "" {
		t.Fatal("empty qid")
	}
}
