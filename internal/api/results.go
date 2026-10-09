package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

type openResultRequest struct {
	Source pagerSourceDTO `json:"source"`
}

type pagerSourceDTO struct {
	Kind     string         `json:"kind"`
	Schema   string         `json:"schema,omitempty"`
	Table    string         `json:"table,omitempty"`
	SQL      string         `json:"sql,omitempty"`
	Filter   filterGroupDTO `json:"filter,omitempty"`
	RawWhere string         `json:"rawWhere,omitempty"`
	Sort     []sortKeyDTO   `json:"sort,omitempty"`
}

type filterGroupDTO struct {
	Logic      string               `json:"logic"`
	Conditions []filterConditionDTO `json:"conditions"`
}

type filterConditionDTO struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Value  string `json:"value"`
	Value2 string `json:"value2,omitempty"`
}

type sortKeyDTO struct {
	Column string `json:"column"`
	Desc   bool   `json:"desc"`
}

func sourceFromDTO(d pagerSourceDTO) pager.Source {
	var conds []sqlgen.FilterCondition
	for _, c := range d.Filter.Conditions {
		conds = append(conds, sqlgen.FilterCondition{
			Column: c.Column,
			Op:     c.Op,
			Value:  c.Value,
			Value2: c.Value2,
		})
	}
	var sort []sqlgen.SortKey
	for _, s := range d.Sort {
		sort = append(sort, sqlgen.SortKey{Column: s.Column, Desc: s.Desc})
	}
	return pager.Source{
		Kind:     d.Kind,
		Schema:   d.Schema,
		Table:    d.Table,
		SQL:      d.SQL,
		Filter:   sqlgen.FilterGroup{Logic: d.Filter.Logic, Conditions: conds},
		RawWhere: d.RawWhere,
		Sort:     sort,
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
	out, err := s.OpenResultFromRequest(ctx, sourceFromDTO(req.Source))
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

func (h *SessionsHandler) getResultValue(w http.ResponseWriter, r *http.Request, sid, rest string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	path := strings.TrimPrefix(rest, "results/")
	path = strings.TrimSuffix(path, "/value")
	path = strings.Trim(path, "/")
	rid := path
	row, _ := strconv.Atoi(r.URL.Query().Get("row"))
	col := r.URL.Query().Get("col")
	if col == "" {
		writeError(w, http.StatusBadRequest, "invalid_query", "col required", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	out, err := s.ResultCellValue(ctx, rid, row, col)
	if err != nil {
		writeError(w, http.StatusBadGateway, "value_failed", "Could not load value.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
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
