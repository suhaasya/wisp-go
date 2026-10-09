package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
)

func TestConnectionParseURLSuccess(t *testing.T) {
	h := api.NewConnectionParseHandler()
	body, _ := json.Marshal(map[string]string{
		"url": "postgres://app:secret@db.example.com:5432/shop?sslmode=require",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/parse-url", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Profile struct {
			Database string `json:"database"`
		} `json:"profile"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Profile.Database != "shop" {
		t.Fatalf("db = %q", out.Profile.Database)
	}
	if out.Password != "secret" {
		t.Fatalf("password = %q", out.Password)
	}
}

func TestConnectionParseURLGarbage(t *testing.T) {
	h := api.NewConnectionParseHandler()
	body, _ := json.Marshal(map[string]string{"url": "not a connection string"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/parse-url", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestConnectionParseURLDoesNotEchoInErrorDetail(t *testing.T) {
	h := api.NewConnectionParseHandler()
	secret := "super-secret-pass-xyz"
	body, _ := json.Marshal(map[string]string{"url": secret})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/parse-url", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if strings.Contains(rec.Body.String(), secret) {
		t.Fatalf("error response echoed input secret")
	}
}
