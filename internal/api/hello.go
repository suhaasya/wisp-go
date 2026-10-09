package api

import (
	"encoding/json"
	"net/http"
)

// HelloResponse is returned by GET /api/v1/hello.
type HelloResponse struct {
	Version  string         `json:"version"`
	CSRF     string         `json:"csrf"`
	Settings map[string]any `json:"settings"`
}

// HelloHandler serves bootstrap metadata for an authenticated session.
func HelloHandler(version string, csrfFromRequest func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		csrf := csrfFromRequest(r)
		if csrf == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HelloResponse{
			Version:  version,
			CSRF:     csrf,
			Settings: map[string]any{},
		})
	}
}
