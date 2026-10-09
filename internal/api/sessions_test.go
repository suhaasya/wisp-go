package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestSessionsCreateRequiresProfile(t *testing.T) {
	st := store.NewProfileStore(t.TempDir() + "/connections.json")
	mgr := session.NewManager(st, nil, nil)
	h := api.NewSessionsHandler(mgr, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/sessions", bytes.NewReader([]byte(`{}`)))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}
