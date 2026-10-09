package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
)

func TestHelloHandler_returnsVersionCSRFAndSettings(t *testing.T) {
	t.Parallel()
	h := api.HelloHandler("1.2.3", func(_ *http.Request) string { return "tok-abc" }, func() (map[string]any, error) {
		return map[string]any{"theme": "system"}, nil
	})
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/hello", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body api.HelloResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Version != "1.2.3" || body.CSRF != "tok-abc" {
		t.Fatalf("body = %+v", body)
	}
	if body.Settings["theme"] != "system" {
		t.Fatalf("settings = %+v", body.Settings)
	}
}
