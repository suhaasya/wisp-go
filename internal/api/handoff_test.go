package api_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
)

func TestHandoff_rejectsMissingSecret(t *testing.T) {
	t.Parallel()
	h := api.NewHandoffHandler(filepath.Join(t.TempDir(), "handoff.secret"), func() string {
		return "http://127.0.0.1:1/auth/exchange?t=abc"
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/handoff", http.NoBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestHandoff_rejectsWrongSecret(t *testing.T) {
	t.Parallel()
	secretPath := filepath.Join(t.TempDir(), "handoff.secret")
	if err := os.WriteFile(secretPath, []byte("correct"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := api.NewHandoffHandler(secretPath, func() string {
		return "http://127.0.0.1:1/auth/exchange?t=abc"
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/handoff", bytes.NewReader(nil))
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestHandoff_returnsLoginURL(t *testing.T) {
	t.Parallel()
	const wantURL = "http://127.0.0.1:8080/auth/exchange?t=token1"
	secretPath := filepath.Join(t.TempDir(), "handoff.secret")
	if err := os.WriteFile(secretPath, []byte("s3cret"), 0o600); err != nil {
		t.Fatal(err)
	}
	h := api.NewHandoffHandler(secretPath, func() string { return wantURL })
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/handoff", http.NoBody)
	req.Header.Set("Authorization", "Bearer s3cret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !bytes.Contains([]byte(body), []byte(wantURL)) {
		t.Fatalf("body = %q, want URL %q", body, wantURL)
	}
}
