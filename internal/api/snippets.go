package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/suhaasya/wisp-go/internal/store"
)

// SnippetsHandler serves saved SQL snippets.
type SnippetsHandler struct {
	snippets *store.SnippetsStore
}

// NewSnippetsHandler returns the snippets handler.
func NewSnippetsHandler(snippets *store.SnippetsStore) *SnippetsHandler {
	return &SnippetsHandler{snippets: snippets}
}

// ServeHTTP implements GET/POST /api/v1/snippets.
func (h *SnippetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		doc, err := h.snippets.Load()
		if err != nil {
			writeError(w, http.StatusBadGateway, "snippets_load_failed", "Could not load snippets.", err.Error())
			return
		}
		writeJSON(w, doc)
	case http.MethodPost:
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
			return
		}
		var sn store.Snippet
		if err := json.Unmarshal(body, &sn); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
			return
		}
		if err := h.snippets.Upsert(sn); err != nil {
			writeError(w, http.StatusBadRequest, "snippets_save_failed", "Could not save snippet.", err.Error())
			return
		}
		writeJSON(w, sn)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
