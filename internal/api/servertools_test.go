package api_test

import (
	"bytes"
	"context"
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

type stMetaConnector struct{}

func (stMetaConnector) Connect(ctx context.Context, _ store.Profile, _ []byte, _ session.SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error) {
	cfg := mock.Config{
		TotalRows: 2,
		Columns: []driver.ColumnMeta{
			{Name: "pid", Type: "text"},
			{Name: "wisp_highlight", Type: "text"},
		},
	}
	conn, err := mock.Driver{}.Open(ctx, driver.ConnectParams{Extra: cfg})
	if err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	return conn, driver.ServerInfo{Engine: "postgres"}, func() {}, nil
}

func TestPostServerToolsProcesses(t *testing.T) {
	main, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: 1, Columns: []driver.ColumnMeta{{Name: "x", Type: "text"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = main.Close() })

	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	mgr.SetConnector(stMetaConnector{})
	s := session.NewTestSession(store.Profile{Engine: "postgres"})
	s.AdoptMainConn(main)
	mgr.AdoptSession(s)

	h := api.NewSessionsHandler(mgr, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/sessions/test/server-tools/processes", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		RID string `json:"rid"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.RID == "" {
		t.Fatal("empty rid")
	}
}

func TestPostServerToolsCancelReadOnly(t *testing.T) {
	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	s := session.NewTestSession(store.Profile{Engine: "postgres", ReadOnly: true})
	mgr.AdoptSession(s)

	h := api.NewSessionsHandler(mgr, nil)
	body, _ := json.Marshal(map[string]any{"terminate": false})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/sessions/test/server-tools/processes/99/cancel",
		bytes.NewReader(body),
	)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
