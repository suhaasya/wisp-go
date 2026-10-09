package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
)

type changeRequest struct {
	Op       string  `json:"op"`
	Row      []any   `json:"row,omitempty"`
	Col      string  `json:"col,omitempty"`
	Value    *string `json:"value,omitempty"`
	Null     bool    `json:"null,omitempty"`
	Original *string `json:"original,omitempty"`
	OrigNull bool    `json:"origNull,omitempty"`
	Default  bool    `json:"default,omitempty"`
}

type changeResponse struct {
	Count     int    `json:"count"`
	InsertKey string `json:"insertKey,omitempty"`
}

func (h *SessionsHandler) serveResults(w http.ResponseWriter, r *http.Request, sid, rest string) {
	switch {
	case rest == "results" && r.Method == http.MethodPost:
		h.postResult(w, r, sid)
	case strings.HasPrefix(rest, "results/") && strings.Contains(rest, "/changes"):
		rid, sub, ok := parseChangesRest(strings.TrimPrefix(rest, "results/"))
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			break
		}
		switch sub {
		case "":
			switch r.Method {
			case http.MethodPost:
				h.postResultChange(w, r, sid, rid)
			case http.MethodDelete:
				h.deleteResultChanges(w, r, sid, rid)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case "sql":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				break
			}
			h.getResultChangesSQL(w, r, sid, rid)
		case "commit":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				break
			}
			h.postResultChangesCommit(w, r, sid, rid)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	case strings.HasPrefix(rest, "results/") && strings.HasSuffix(rest, "/count") && r.Method == http.MethodPost:
		rid := strings.TrimSuffix(strings.TrimPrefix(rest, "results/"), "/count")
		rid = strings.TrimSuffix(rid, "/")
		h.postResultCount(w, r, sid, rid)
	case strings.HasPrefix(rest, "results/") && strings.Contains(rest, "/window") && r.Method == http.MethodGet:
		h.getResultWindow(w, r, sid, rest)
	case strings.HasPrefix(rest, "results/") && strings.Contains(rest, "/value") && r.Method == http.MethodGet:
		h.getResultValue(w, r, sid, rest)
	case strings.HasPrefix(rest, "results/") && r.Method == http.MethodDelete:
		rid := strings.TrimPrefix(rest, "results/")
		rid = strings.TrimSuffix(rid, "/")
		h.deleteResult(w, r, sid, rid)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SessionsHandler) postResultChange(w http.ResponseWriter, r *http.Request, sid, rid string) {
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
	var req changeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", "")
		return
	}
	ch := session.Change{
		Op:       req.Op,
		Row:      req.Row,
		Col:      req.Col,
		Value:    req.Value,
		Null:     req.Null || req.Default,
		Original: req.Original,
		OrigNull: req.OrigNull,
	}
	count, err := s.ApplyResultChange(rid, ch)
	if err != nil {
		if session.IsWriteBlockedReadOnly(err) {
			writeError(w, http.StatusForbidden, "read_only", "Session is read-only.", "")
			return
		}
		writeError(w, http.StatusBadRequest, "change_rejected", err.Error(), "")
		return
	}
	resp := changeResponse{Count: count}
	if req.Op == "insert" {
		resp.InsertKey = s.ResultLastInsertKey(rid)
	}
	writeJSON(w, resp)
}

func (h *SessionsHandler) deleteResultChanges(w http.ResponseWriter, r *http.Request, sid, rid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	if err := s.DiscardResultChanges(rid); err != nil {
		writeError(w, http.StatusNotFound, "result_not_found", "Result not found.", "")
		return
	}
	writeJSON(w, changeResponse{Count: 0})
}

func parseChangesRest(path string) (rid, sub string, ok bool) {
	path = strings.Trim(path, "/")
	i := strings.Index(path, "/changes")
	if i < 0 {
		return "", "", false
	}
	rid = strings.Trim(path[:i], "/")
	tail := strings.TrimPrefix(path[i+len("/changes"):], "/")
	return rid, tail, rid != ""
}

func (h *SessionsHandler) getResultChangesSQL(w http.ResponseWriter, r *http.Request, sid, rid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := s.ResultReviewSQL(ctx, rid)
	if err != nil {
		writeError(w, http.StatusBadRequest, "review_sql_failed", err.Error(), "")
		return
	}
	writeJSON(w, out)
}

type commitChangesRequest struct {
	Confirm      bool   `json:"confirm"`
	TableConfirm string `json:"tableConfirm,omitempty"`
}

func (h *SessionsHandler) postResultChangesCommit(w http.ResponseWriter, r *http.Request, sid, rid string) {
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
	var req commitChangesRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", "")
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	res, err := s.CommitResultChanges(ctx, rid, session.CommitOptions{
		Confirm:      true,
		TableConfirm: req.TableConfirm,
	})
	if err != nil {
		var ce *session.CommitError
		if errors.As(err, &ce) {
			writeError(w, http.StatusBadGateway, "commit_failed", ce.Message,
				fmt.Sprintf("statementIndex=%d conflict=%v", ce.StatementIndex, ce.Conflict))
			return
		}
		if session.IsWriteBlockedReadOnly(err) {
			writeError(w, http.StatusForbidden, "read_only", "Session is read-only.", "")
			return
		}
		writeError(w, http.StatusBadRequest, "commit_rejected", err.Error(), "")
		return
	}
	writeJSON(w, res)
}
