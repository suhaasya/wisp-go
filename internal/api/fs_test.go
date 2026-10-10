package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
)

func TestFSListHomeRelative(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sub := filepath.Join(home, "keys")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "id_rsa"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := api.NewFSHandler(home)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/fs/list?path=~/keys", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Entries []struct {
			Name string `json:"name"`
			Dir  bool   `json:"dir"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Entries) != 1 || out.Entries[0].Name != "id_rsa" {
		t.Fatalf("entries = %+v", out.Entries)
	}
}

func TestFSListRejectsOutsideHome(t *testing.T) {
	home := t.TempDir()
	outside, err := filepath.Abs(filepath.Join(home, "..", "wisp-fs-outside"))
	if err != nil {
		t.Fatal(err)
	}
	h := api.NewFSHandler(home)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/fs/list?path="+url.QueryEscape(outside), nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}
