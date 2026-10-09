package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSP_devAllowsInlineStylesForVite(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	setCSPDefault(rec, true)
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "style-src 'self' 'unsafe-inline'") {
		t.Fatalf("dev CSP = %q, want inline styles for Vite", csp)
	}
}

func TestCSP_productionDoesNotAllowUnsafeInlineStyles(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	setCSPDefault(rec, false)
	csp := rec.Header().Get("Content-Security-Policy")
	if strings.Contains(csp, "unsafe-inline") {
		t.Fatalf("prod CSP = %q, should not allow unsafe-inline styles by default", csp)
	}
}
