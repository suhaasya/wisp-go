package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
)

func TestDetectLocalHandler_GET(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/detect/local", nil)
	rec := httptest.NewRecorder()
	api.NewDetectHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var out struct {
		Suggestions []any `json:"suggestions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Suggestions == nil {
		t.Fatal("missing suggestions array")
	}
}
