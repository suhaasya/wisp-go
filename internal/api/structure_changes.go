package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

type structureChangeRequest struct {
	Op string `json:"op"`
	sqlgen.StructureOp
}

type structureChangeResponse struct {
	Count int `json:"count"`
}

func (h *SessionsHandler) serveStructureChanges(w http.ResponseWriter, r *http.Request, sid, schema, table, rest string) {
	switch rest {
	case "changes":
		switch r.Method {
		case http.MethodPost:
			h.postStructureChange(w, r, sid, schema, table)
		case http.MethodDelete:
			h.deleteStructureChanges(w, r, sid, schema, table)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "changes/sql":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.getStructureChangesSQL(w, r, sid, schema, table)
	case "changes/commit":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.postStructureChangesCommit(w, r, sid, schema, table)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SessionsHandler) postStructureChange(w http.ResponseWriter, r *http.Request, sid, schema, table string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Could not read body.", "")
		return
	}
	var req structureChangeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON.", err.Error())
		return
	}
	op := req.StructureOp
	if req.Op != "" {
		op.Kind = sqlgen.StructureOpKind(req.Op)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	count, err := s.ApplyStructureChange(ctx, schema, table, op)
	if err != nil {
		writeChangeError(w, err)
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, structureChangeResponse{Count: count})
}

func (h *SessionsHandler) deleteStructureChanges(w http.ResponseWriter, r *http.Request, sid, schema, table string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	if err := s.DiscardStructureChanges(schema, table); err != nil {
		writeError(w, http.StatusBadGateway, "discard_failed", "Could not discard changes.", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) getStructureChangesSQL(w http.ResponseWriter, r *http.Request, sid, schema, table string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := s.StructureReviewSQL(ctx, schema, table)
	if err != nil {
		writeError(w, http.StatusBadGateway, "review_failed", "Could not build SQL.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}

type structureCommitRequest struct {
	Confirm      bool   `json:"confirm"`
	TableConfirm string `json:"tableConfirm"`
}

func (h *SessionsHandler) postStructureChangesCommit(w http.ResponseWriter, r *http.Request, sid, schema, table string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	var req structureCommitRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	out, err := s.CommitStructureChanges(ctx, schema, table, session.CommitOptions{
		Confirm: req.Confirm, TableConfirm: req.TableConfirm,
	})
	if err != nil {
		writeChangeError(w, err)
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}

func writeChangeError(w http.ResponseWriter, err error) {
	if session.IsWriteBlockedReadOnly(err) {
		writeError(w, http.StatusForbidden, "read_only", "Session is read-only.", "")
		return
	}
	writeError(w, http.StatusBadGateway, "change_failed", err.Error(), "")
}
