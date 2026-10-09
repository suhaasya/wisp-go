package api

import (
	"context"
	"net/http"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
)

// DetectHandler serves GET /api/v1/detect/local.
type DetectHandler struct{}

// NewDetectHandler returns a local database detection handler.
func NewDetectHandler() http.Handler {
	return DetectHandler{}
}

func (DetectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET required", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	suggestions := session.DetectLocal(ctx)
	writeJSON(w, map[string]any{"suggestions": suggestions})
}
