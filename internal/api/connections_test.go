package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/store"
)

type memSecrets struct {
	data map[string]map[store.SecretKind][]byte
}

func newMemSecrets() *memSecrets {
	return &memSecrets{data: make(map[string]map[store.SecretKind][]byte)}
}

func (m *memSecrets) Get(profileID string, kind store.SecretKind) ([]byte, bool, error) {
	if m.data[profileID] == nil {
		return nil, false, nil
	}
	v, ok := m.data[profileID][kind]
	if !ok {
		return nil, false, nil
	}
	out := make([]byte, len(v))
	copy(out, v)
	return out, true, nil
}

func (m *memSecrets) Set(profileID string, kind store.SecretKind, value []byte) error {
	if m.data[profileID] == nil {
		m.data[profileID] = make(map[store.SecretKind][]byte)
	}
	cp := make([]byte, len(value))
	copy(cp, value)
	m.data[profileID][kind] = cp
	return nil
}

func (m *memSecrets) Delete(profileID string, kind store.SecretKind) error {
	if m.data[profileID] != nil {
		delete(m.data[profileID], kind)
	}
	return nil
}

func (m *memSecrets) DeleteAll(profileID string) error {
	delete(m.data, profileID)
	return nil
}

func profileJSONBody(t *testing.T, extra map[string]any) []byte {
	t.Helper()
	base := map[string]any{
		"name":          "Staging DB",
		"folder":        "Work",
		"engine":        "postgres",
		"transport":     map[string]any{"kind": "tcp", "host": "db.example.com", "port": 5432},
		"user":          "app",
		"database":      "appdb",
		"password_mode": "keychain",
		"ssl":           map[string]string{"mode": "disable"},
		"env":           "staging",
		"read_only":     true,
	}
	for k, v := range extra {
		base[k] = v
	}
	body, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestConnectionsGETEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	h := api.NewConnectionsHandler(profiles, newMemSecrets())
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/connections", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatalf("unexpected password in list: %s", rec.Body.String())
	}
}

func TestConnectionsPOSTStoresSecretNotOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	sec := newMemSecrets()
	h := api.NewConnectionsHandler(profiles, sec)

	const dbPass = "super-secret-db-pass"
	body := profileJSONBody(t, map[string]any{"password": dbPass})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	resp := rec.Body.String()
	if strings.Contains(resp, dbPass) {
		t.Fatalf("response leaked password: %s", resp)
	}
	var created struct {
		ID          string `json:"id"`
		HasPassword bool   `json:"has_password"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if !created.HasPassword {
		t.Fatal("expected has_password true")
	}
	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(onDisk), dbPass) || strings.Contains(string(onDisk), `"password"`) {
		t.Fatalf("disk file contains password material: %s", onDisk)
	}
	got, ok, _ := sec.Get(created.ID, store.SecretKindDBPassword)
	if !ok || string(got) != dbPass {
		t.Fatalf("secret not stored")
	}
}

func TestConnectionsPUTPasswordAbsentUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	sec := newMemSecrets()
	h := api.NewConnectionsHandler(profiles, sec)

	createBody := profileJSONBody(t, map[string]any{"password": "keep-me"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections", bytes.NewReader(createBody))
	h.ServeHTTP(rec, req)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	updateBody := profileJSONBody(t, map[string]any{"name": "Renamed"})
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/connections/"+created.ID, bytes.NewReader(updateBody))
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d", rec2.Code)
	}
	got, ok, _ := sec.Get(created.ID, store.SecretKindDBPassword)
	if !ok || string(got) != "keep-me" {
		t.Fatal("password should remain")
	}
}

func TestConnectionsPUTEmptyPasswordDeletes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	sec := newMemSecrets()
	h := api.NewConnectionsHandler(profiles, sec)

	createBody := profileJSONBody(t, map[string]any{"password": "gone-soon"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections", bytes.NewReader(createBody))
	h.ServeHTTP(rec, req)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	updateBody := profileJSONBody(t, map[string]any{"password": ""})
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/connections/"+created.ID, bytes.NewReader(updateBody))
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d", rec2.Code)
	}
	var out struct {
		HasPassword bool `json:"has_password"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &out)
	if out.HasPassword {
		t.Fatal("expected has_password false")
	}
	_, ok, _ := sec.Get(created.ID, store.SecretKindDBPassword)
	if ok {
		t.Fatal("password should be deleted")
	}
}

func TestConnectionsDELETERemovesSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	sec := newMemSecrets()
	h := api.NewConnectionsHandler(profiles, sec)

	createBody := profileJSONBody(t, map[string]any{"password": "x"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections", bytes.NewReader(createBody))
	h.ServeHTTP(rec, req)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/connections/"+created.ID, nil)
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec2.Code)
	}
	if _, ok := sec.data[created.ID]; ok {
		t.Fatal("secrets map entry should be gone")
	}
}

func TestConnectionsDuplicateCopiesSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	sec := newMemSecrets()
	h := api.NewConnectionsHandler(profiles, sec)

	createBody := profileJSONBody(t, map[string]any{"password": "copy-me", "ssh_password": "ssh-copy"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections", bytes.NewReader(createBody))
	h.ServeHTTP(rec, req)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/"+created.ID+"/duplicate", nil)
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec2.Code, rec2.Body.String())
	}
	var dup struct {
		ID          string `json:"id"`
		HasPassword bool   `json:"has_password"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &dup)
	if dup.ID == created.ID || !dup.HasPassword {
		t.Fatalf("dup = %+v", dup)
	}
	db, ok, _ := sec.Get(dup.ID, store.SecretKindDBPassword)
	if !ok || string(db) != "copy-me" {
		t.Fatal("db password not copied")
	}
	ssh, ok, _ := sec.Get(dup.ID, store.SecretKindSSHPassword)
	if !ok || string(ssh) != "ssh-copy" {
		t.Fatal("ssh password not copied")
	}
}

func TestConnectionsResponseNeverContainsPasswordField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	profiles := store.NewProfileStore(path)
	if _, err := profiles.Load(); err != nil {
		t.Fatal(err)
	}
	h := api.NewConnectionsHandler(profiles, newMemSecrets())
	body := profileJSONBody(t, map[string]any{"password": "must-not-appear"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	raw := rec.Body.String()
	if strings.Contains(raw, "must-not-appear") {
		t.Fatalf("leaked: %s", raw)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["password"]; ok {
		t.Fatal("password key in response")
	}
}
