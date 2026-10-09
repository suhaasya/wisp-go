package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/store"
)

type memBroadcaster struct {
	types []string
}

func (m *memBroadcaster) Broadcast(typ string, data any) {
	m.types = append(m.types, typ)
}

func TestSettingsHandler_getAndPut(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/settings.json"
	st := store.NewSettingsStore(path)
	_, _ = st.Load()
	var bc memBroadcaster
	h := api.NewSettingsHandler(st, &bc)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/settings", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d", rec.Code)
	}

	body, _ := json.Marshal(map[string]any{"theme": "dark", "version": 1})
	rec = httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(bc.types) != 1 || bc.types[0] != "settings" {
		t.Fatalf("broadcast = %v", bc.types)
	}
	loaded, _ := st.Load()
	if loaded.Theme != "dark" {
		t.Fatalf("theme = %q", loaded.Theme)
	}
}

func TestSettingsHandler_putInvalidTheme(t *testing.T) {
	t.Parallel()
	st := store.NewSettingsStore(t.TempDir() + "/settings.json")
	_, _ = st.Load()
	h := api.NewSettingsHandler(st, &memBroadcaster{})

	body, _ := json.Marshal(map[string]any{"theme": "neon"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/settings", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error.Code == "" || resp.Error.Message == "" {
		t.Fatalf("error body = %+v", resp)
	}
}

func TestSettingsHandler_putInvalidJSON(t *testing.T) {
	st := store.NewSettingsStore(t.TempDir() + "/settings.json")
	_, _ = st.Load()
	_ = st.Save(store.DefaultSettings())
	h := api.NewSettingsHandler(st, &memBroadcaster{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/settings", bytes.NewReader([]byte("{")))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	if st.Current().Theme != store.DefaultSettings().Theme {
		t.Fatalf("settings mutated on bad put")
	}
}
