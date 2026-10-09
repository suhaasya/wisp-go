package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
)

type flushWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

func (f flushWriter) Flush() {
	if f.flusher != nil {
		f.flusher.Flush()
	}
}

func (h *SessionsHandler) getExport(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	req, err := parseExportRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_export", err.Error(), "")
		return
	}
	if req.RID == "" {
		writeError(w, http.StatusBadRequest, "invalid_export", "rid required", "")
		return
	}
	s.TouchTabRef(time.Now())

	if strings.TrimSpace(req.FilePath) != "" {
		if h.fs == nil {
			writeError(w, http.StatusInternalServerError, "export_unavailable", "File export is not configured.", "")
			return
		}
		resolved, err := h.fs.ResolveHomePath(req.FilePath)
		if err != nil {
			writeError(w, http.StatusForbidden, "path_forbidden", "Path must stay under your home directory.", err.Error())
			return
		}
		req.FilePath = resolved
		eid, err := s.StartExportFile(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusBadRequest, "export_failed", "Could not start export.", err.Error())
			return
		}
		writeJSON(w, map[string]string{"export_id": eid})
		return
	}

	filename := exportFilename(req)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", exportContentType(req.Format))
	fl, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "export_unavailable", "Streaming is not supported.", "")
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	dst := flushWriter{ResponseWriter: w, flusher: fl}
	_, _ = s.StreamExport(ctx, req, dst)
}

func (h *SessionsHandler) postExportCancel(w http.ResponseWriter, r *http.Request, sid, eid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	if err := s.CancelExport(eid); err != nil {
		writeError(w, http.StatusNotFound, "export_not_found", "Export not found.", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseExportRequest(r *http.Request) (session.ExportRequest, error) {
	q := r.URL.Query()
	req := session.ExportRequest{
		Format:   q.Get("format"),
		Scope:    q.Get("scope"),
		RID:      q.Get("rid"),
		FilePath: q.Get("path"),
		ColMin:   -1,
		ColMax:   -1,
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("invalid offset")
		}
		req.Offset = n
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("invalid limit")
		}
		req.Limit = n
	}
	if v := q.Get("row_min"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("invalid row_min")
		}
		req.RowMin = n
	}
	if v := q.Get("row_max"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("invalid row_max")
		}
		req.RowMax = n
	}
	if v := q.Get("col_min"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("invalid col_min")
		}
		req.ColMin = n
	}
	if v := q.Get("col_max"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("invalid col_max")
		}
		req.ColMax = n
	}
	return req, nil
}

func exportContentType(format string) string {
	switch strings.ToLower(format) {
	case "json":
		return "application/json"
	case "markdown", "md":
		return "text/markdown; charset=utf-8"
	case "sql":
		return "application/sql; charset=utf-8"
	default:
		return "text/csv; charset=utf-8"
	}
}

func exportFilename(req session.ExportRequest) string {
	ext := "csv"
	switch strings.ToLower(req.Format) {
	case "json":
		ext = "json"
	case "sql":
		ext = "sql"
	case "markdown", "md":
		ext = "md"
	}
	return "wisp-export." + ext
}
