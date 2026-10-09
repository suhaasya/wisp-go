package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestWorkspaceHandlerPutGet(t *testing.T) {
	dir := t.TempDir()
	paths := store.PathsFromDirs(dir, dir)
	_ = paths.EnsureConfig()
	ps := store.NewProfileStore(paths.ConnectionsJSON())
	_, err := ps.Create(store.Profile{
		ID: "p1", Name: "Local", Engine: "postgres", Env: "local", PasswordMode: "none",
		Transport: store.Transport{Kind: "tcp", Host: "127.0.0.1", Port: 5432},
	})
	if err != nil {
		t.Fatal(err)
	}
	ws := session.NewWorkspaceStore(paths, ps)
	h := api.NewWorkspaceHandler(ws)

	payload, _ := json.Marshal(session.WorkspaceState{
		Version: 1, ActiveTabID: "t1",
		Tabs: []session.WorkspaceTabState{{ID: "t1", Kind: session.TabKindQuery, QueryText: "select 2"}},
	})
	rec := httptest.NewRecorder()
	put := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/workspace/p1", bytes.NewReader(payload))
	put.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, put)
	if rec.Code != http.StatusOK {
		t.Fatalf("put=%d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/workspace/p1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get=%d", rec.Code)
	}
	var out session.WorkspaceState
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Tabs[0].QueryText != "select 2" {
		t.Fatalf("query=%q", out.Tabs[0].QueryText)
	}
}
