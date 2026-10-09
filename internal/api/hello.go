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

// SettingsSnapshot returns bootstrap settings for hello.
type SettingsSnapshot func() (map[string]any, error)

// HelloHandler serves bootstrap metadata for an authenticated session.
func HelloHandler(version string, csrfFromRequest func(*http.Request) string, settings SettingsSnapshot) http.HandlerFunc {
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
		doc := map[string]any{}
		if settings != nil {
			var err error
			doc, err = settings()
			if err != nil {
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HelloResponse{
			Version:  version,
			CSRF:     csrf,
			Settings: doc,
		})
	}
}
