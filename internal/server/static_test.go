package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/suhaasya/wisp-go/internal/server"
)

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
