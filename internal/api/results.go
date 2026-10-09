package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type openResultRequest struct {
	Source resultSourceDTO `json:"source"`
}

type resultSourceDTO struct {
	Kind   string `json:"kind"`
	Schema string `json:"schema,omitempty"`
	Table  string `json:"table,omitempty"`
	SQL    string `json:"sql,omitempty"`
}

func (h *SessionsHandler) serveResults(w http.ResponseWriter, r *http.Request, sid, rest string) {
	switch {
	case rest == "results" && r.Method == http.MethodPost:
		h.postResult(w, r, sid)
	case strings.HasPrefix(rest, "results/") && strings.HasSuffix(rest, "/count") && r.Method == http.MethodPost:
		rid := strings.TrimSuffix(strings.TrimPrefix(rest, "results/"), "/count")
		rid = strings.TrimSuffix(rid, "/")
		h.postResultCount(w, r, sid, rid)
	case strings.HasPrefix(rest, "results/") && strings.Contains(rest, "/window") && r.Method == http.MethodGet:
		h.getResultWindow(w, r, sid, rest)
	case strings.HasPrefix(rest, "results/") && r.Method == http.MethodDelete:
		rid := strings.TrimPrefix(rest, "results/")
		rid = strings.TrimSuffix(rid, "/")
		h.deleteResult(w, r, sid, rid)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SessionsHandler) postResult(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req openResultRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	out, err := s.OpenResultFromRequest(ctx, req.Source.Kind, req.Source.Schema, req.Source.Table, req.Source.SQL)
	if err != nil {
		writeError(w, http.StatusBadGateway, "result_open_failed", "Could not open result.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}

func (h *SessionsHandler) getResultWindow(w http.ResponseWriter, r *http.Request, sid, rest string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	path := strings.TrimPrefix(rest, "results/")
	path = strings.TrimSuffix(path, "/window")
	path = strings.Trim(path, "/")
	rid := path
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	raw, err := s.ResultWindowJSON(ctx, rid, offset, limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, "window_failed", "Could not load window.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeRawJSON(w, raw)
}

func (h *SessionsHandler) postResultCount(w http.ResponseWriter, r *http.Request, sid, rid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	total, err := s.ExactResultCount(ctx, rid)
	if err != nil {
		writeError(w, http.StatusBadGateway, "count_failed", "Could not count rows.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, total)
}

func (h *SessionsHandler) deleteResult(w http.ResponseWriter, r *http.Request, sid, rid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	s.CloseResult(rid)
	w.WriteHeader(http.StatusNoContent)
}
