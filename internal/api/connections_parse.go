package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/suhaasya/wisp-go/internal/store"
)

const connectionParsePath = "/api/v1/connections/parse-url"

// ConnectionParseHandler serves POST /connections/parse-url.
type ConnectionParseHandler struct{}

// NewConnectionParseHandler returns a URL parse handler.
func NewConnectionParseHandler() *ConnectionParseHandler {
	return &ConnectionParseHandler{}
}

type parseURLRequest struct {
	URL string `json:"url"`
}

type parseURLResponse struct {
	Profile  store.Profile `json:"profile"`
	Password string        `json:"password,omitempty"`
	Warnings []string      `json:"warnings,omitempty"`
}

// ServeHTTP implements http.Handler.
func (h *ConnectionParseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != connectionParsePath {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req parseURLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	parsed, err := store.ParseConnectionURL(req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "parse_failed", "Could not parse connection URL", err.Error())
		return
	}
	writeJSON(w, parseURLResponse{
		Profile:  parsed.Profile,
		Password: parsed.Password,
		Warnings: parsed.Warnings,
	})
}
