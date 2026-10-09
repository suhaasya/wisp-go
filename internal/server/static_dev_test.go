package server_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/suhaasya/wisp-go/internal/server"
)

func TestStatic_devFallbackWhenViteUnreachable(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"index.html": {Data: []byte("<html><head></head><body>fallback</body></html>")},
	}
	st, err := server.NewStatic(fsys, true)
	if err != nil {
		t.Fatal(err)
	}
	st.Vite, err = url.Parse("http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	st.Register(mux)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "fallback") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
