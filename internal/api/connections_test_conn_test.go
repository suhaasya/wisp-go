package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestConnectionTestWrongPortRefused(t *testing.T) {
	dir := t.TempDir()
	paths := store.Paths{ConfigDir: dir, DataDir: filepath.Join(dir, "data")}
	sec, err := store.OpenSecretsStore(paths, store.FakeKeyringProbe(store.KeyringProbeUnavailable))
	if err != nil {
		t.Fatal(err)
	}
	h := api.NewConnectionTestHandler(sec)
	body, _ := json.Marshal(map[string]any{
		"engine":    "postgres",
		"transport": map[string]any{"kind": "tcp", "host": "127.0.0.1", "port": 1},
		"user":      "u",
		"database":  "d",
		"password":  "x",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/test", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Error.Code != "connection_refused" {
		t.Fatalf("code = %q body=%s", out.Error.Code, rec.Body.String())
	}
}

func TestConnectionTestOmitsStoredPassword(t *testing.T) {
	dir := t.TempDir()
	paths := store.Paths{ConfigDir: dir, DataDir: filepath.Join(dir, "data")}
	sec, err := store.OpenSecretsStore(paths, store.FakeKeyringProbe(store.KeyringProbeUnavailable))
	if err != nil {
		t.Fatal(err)
	}
	master := "master-pass-12345"
	if err := sec.UnlockMaster([]byte(master)); err != nil {
		t.Fatal(err)
	}
	if err := sec.Set("prof-1", store.SecretKindDBPassword, []byte("secret-db")); err != nil {
		t.Fatal(err)
	}
	h := api.NewConnectionTestHandler(sec)
	body, _ := json.Marshal(map[string]any{
		"profile_id": "prof-1",
		"engine":     "postgres",
		"transport":  map[string]any{"kind": "tcp", "host": "127.0.0.1", "port": 1},
		"user":       "u",
		"database":   "d",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/test", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	resp := rec.Body.String()
	if bytes.Contains(rec.Body.Bytes(), []byte("secret-db")) {
		t.Fatalf("response leaked password: %s", resp)
	}
}
