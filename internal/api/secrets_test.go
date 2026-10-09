package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestSecretsUnlockWrongPasswordNoSecretInResponse(t *testing.T) {
	dir := t.TempDir()
	paths := store.Paths{ConfigDir: dir, DataDir: filepath.Join(dir, "data")}
	st, err := store.OpenSecretsStore(paths, store.FakeKeyringProbe(store.KeyringProbeUnavailable))
	if err != nil {
		t.Fatal(err)
	}
	master := "correct-master-pass"
	if err := st.UnlockMaster([]byte(master)); err != nil {
		t.Fatal(err)
	}
	if err := st.Set("p1", store.SecretKindDBPassword, []byte("stored-db-pass")); err != nil {
		t.Fatal(err)
	}

	h := api.NewSecretsHandler(st)
	body, _ := json.Marshal(map[string]string{"master_password": "wrong-master-pass"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/secrets/unlock", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	respBody := rec.Body.String()
	if strings.Contains(respBody, "stored-db-pass") || strings.Contains(respBody, "wrong-master-pass") {
		t.Fatalf("response leaked secret material: %s", respBody)
	}
}

func TestSecretsUnlockSuccessNoEcho(t *testing.T) {
	dir := t.TempDir()
	paths := store.Paths{ConfigDir: dir, DataDir: filepath.Join(dir, "data")}
	st, err := store.OpenSecretsStore(paths, store.FakeKeyringProbe(store.KeyringProbeUnavailable))
	if err != nil {
		t.Fatal(err)
	}
	master := "my-master-key"
	h := api.NewSecretsHandler(st)

	body, _ := json.Marshal(map[string]string{"master_password": master})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/secrets/unlock", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), master) {
		t.Fatalf("response echoed master password: %s", rec.Body.String())
	}
	var out struct {
		Unlocked bool `json:"unlocked"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil || !out.Unlocked {
		t.Fatalf("body = %+v err=%v", out, err)
	}
}
