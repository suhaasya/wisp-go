package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/suhaasya/wisp-go/internal/session"
)

type runRequest struct {
	SQL      string `json:"sql"`
	Mode     string `json:"mode"`
	Cursor   int    `json:"cursor,omitempty"`
	SelStart int    `json:"sel_start,omitempty"`
	SelEnd   int    `json:"sel_end,omitempty"`
	Confirm  bool   `json:"confirm,omitempty"`
}

type runResponse struct {
	QID string `json:"qid"`
}

func (h *SessionsHandler) postRun(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req runRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "sql required", "")
		return
	}
	mode := session.RunMode(req.Mode)
	if mode == "" {
		mode = session.RunModeCursor
	}
	qid, err := s.StartRun(r.Context(), session.RunRequest{
		SQL:      req.SQL,
		Mode:     mode,
		Cursor:   req.Cursor,
		SelStart: req.SelStart,
		SelEnd:   req.SelEnd,
		Confirm:  req.Confirm,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "run_failed", "Could not start query run.", err.Error())
		return
	}
	writeJSON(w, runResponse{QID: qid})
}

type explainRequest struct {
	SQL     string `json:"sql"`
	Analyze bool   `json:"analyze,omitempty"`
	Confirm bool   `json:"confirm,omitempty"`
}

func (h *SessionsHandler) postExplain(w http.ResponseWriter, r *http.Request, sid string) {
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
	var req explainRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	out, err := s.Explain(r.Context(), session.ExplainRequest{
		SQL:     req.SQL,
		Analyze: req.Analyze,
		Confirm: req.Confirm,
	})
	if err != nil {
		if errors.Is(err, session.ErrExplainNeedsConfirm()) {
			writeError(w, http.StatusConflict, "needs_confirmation", "This EXPLAIN ANALYZE executes the statement.", "")
			return
		}
		writeError(w, http.StatusBadGateway, "explain_failed", "Could not run EXPLAIN.", err.Error())
		return
	}
	writeJSON(w, out)
}

func (h *SessionsHandler) postCancelRun(w http.ResponseWriter, r *http.Request, sid, qid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	if err := s.CancelRun(qid); err != nil {
		writeError(w, http.StatusNotFound, "query_not_found", "Query run not found.", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
