package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/suhaasya/wisp-go/internal/server"
)

func TestStatic_devEmbeddedAssetBeforeVite(t *testing.T) {
	t.Parallel()
	vite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html>vite</html>"))
	}))
	t.Cleanup(vite.Close)

	fsys := fstest.MapFS{
		"index.html":             {Data: []byte("<html><head></head><body>app</body></html>")},
		"assets/chunk-abc123.js": {Data: []byte("export const x = 1;")},
	}
	st, err := server.NewStaticWithVite(fsys, true, vite.URL)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	st.Register(mux)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/chunk-abc123.js", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Fatalf("Content-Type = %q, want javascript", ct)
	}
	if body := rec.Body.String(); body != "export const x = 1;" {
		t.Fatalf("body = %q", body)
	}
}

func TestStatic_cacheHeaders(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"index.html":              {Data: []byte("<html><head></head><body>app</body></html>")},
		"assets/index-abc123.js":  {Data: []byte("console.log(1)")},
		"assets/index-abc123.css": {Data: []byte("body{}")},
	}
	st, err := server.NewStatic(fsys, false)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	st.Register(mux)

	reqAsset, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/index-abc123.js", nil)
	recAsset := httptest.NewRecorder()
	mux.ServeHTTP(recAsset, reqAsset)
	if recAsset.Code != http.StatusOK {
		t.Fatalf("asset status = %d", recAsset.Code)
	}
	if cc := recAsset.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("asset Cache-Control = %q", cc)
	}

	reqIndex, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	recIndex := httptest.NewRecorder()
	mux.ServeHTTP(recIndex, reqIndex)
	if recIndex.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("index Cache-Control = %q", recIndex.Header().Get("Cache-Control"))
	}
}
