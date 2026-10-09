package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestPostImportPreview(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "sample.csv")
	if err := os.WriteFile(path, []byte("a\n1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := session.NewTestSession(store.Profile{Engine: "mock"})
	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	mgr.AdoptSession(s)
	fs := api.NewFSHandler(home)
	h := api.NewSessionsHandler(mgr, fs)

	body, _ := json.Marshal(map[string]string{"path": "~/sample.csv"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/sessions/test/import/preview", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rec.Code, rec.Body.String())
	}
}
