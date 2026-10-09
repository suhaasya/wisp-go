package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/server"
)

func TestRequestHandoffLoginURL(t *testing.T) {
	t.Parallel()
	const wantURL = "http://127.0.0.1:9/auth/exchange?t=x"
	secretPath := filepath.Join(t.TempDir(), "handoff.secret")
	secret := []byte("handoff-key")
	if err := os.WriteFile(secretPath, secret, 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/internal/handoff", api.NewHandoffHandler(secretPath, func() string { return wantURL }))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	got, err := server.RequestHandoffLoginURL(context.Background(), srv.URL, secret)
	if err != nil {
		t.Fatalf("RequestHandoffLoginURL: %v", err)
	}
	if got != wantURL {
		t.Fatalf("url = %q, want %q", got, wantURL)
	}
}
