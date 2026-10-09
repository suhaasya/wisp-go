package session

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/mover"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

const (
	ExportScopeResult    = "result"
	ExportScopePage      = "page"
	ExportScopeSelection = "selection"
)

// ExportRequest selects rows and format for a streaming export.
type ExportRequest struct {
	Format     string
	Scope      string
	RID        string
	Offset     int
	Limit      int
	RowMin     int
	RowMax     int
	ColMin     int
	ColMax     int
	FilePath   string
	OnProgress func(rows int64)
}

// ExportEvent reports file-export progress over SSE.
type ExportEvent struct {
	EID       string `json:"eid"`
	Rows      int64  `json:"rows"`
	Done      bool   `json:"done,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
	Error     string `json:"error,omitempty"`
}

// StreamExport runs a read-only export to w using a dedicated query cursor.
func (s *Session) StreamExport(ctx context.Context, req ExportRequest, w io.Writer) (int64, error) {
	conn := s.mainConn()
	if conn == nil {
		return 0, errSessionNotReady
	}
	sqlText, args, tableName, err := s.exportSQL(ctx, req)
	if err != nil {
		return 0, err
	}
	check, err := s.Guard(sqlText, false)
	if err != nil {
		return 0, err
	}
	if check == WriteBlocked {
		return 0, ErrWriteBlockedReadOnly()
	}
	format, err := parseExportFormat(req.Format)
	if err != nil {
		return 0, err
	}
	opts := mover.Options{
		Table:         tableName,
		IncludeHeader: true,
		ColStart:      colExportStart(req.ColMin),
		ColEnd:        colExportEnd(req.ColMax),
		OnProgress:    req.OnProgress,
	}
	return mover.StreamQuery(ctx, conn, sqlText, args, format, opts, w)
}

// StartExportFile writes export rows to path asynchronously and returns an export id.
func (s *Session) StartExportFile(parent context.Context, req ExportRequest) (string, error) {
	if strings.TrimSpace(req.FilePath) == "" {
		return "", errors.New("session: export path required")
	}
	eid := newQueryID()
	runCtx, cancel := context.WithCancel(context.WithoutCancel(parent))
	s.exportMu.Lock()
	if s.exports == nil {
		s.exports = make(map[string]*exportJob)
	}
	s.exports[eid] = &exportJob{cancel: cancel}
	s.exportMu.Unlock()

	go func() {
		defer func() {
			s.exportMu.Lock()
			delete(s.exports, eid)
			s.exportMu.Unlock()
		}()
		f, err := os.Create(req.FilePath)
		if err != nil {
			s.emitExportEvent(ExportEvent{EID: eid, Error: err.Error(), Done: true})
			return
		}
		defer func() { _ = f.Close() }()

		reqCopy := req
		reqCopy.OnProgress = func(n int64) {
			s.emitExportEvent(ExportEvent{EID: eid, Rows: n})
		}
		n, err := s.StreamExport(runCtx, reqCopy, f)
		if err != nil {
			if runCtx.Err() != nil {
				s.emitExportEvent(ExportEvent{EID: eid, Cancelled: true, Done: true})
				_ = os.Remove(req.FilePath)
				return
			}
			s.emitExportEvent(ExportEvent{EID: eid, Error: err.Error(), Done: true})
			return
		}
		s.emitExportEvent(ExportEvent{EID: eid, Rows: n, Done: true})
	}()
	return eid, nil
}

// CancelExport cancels a file export job.
func (s *Session) CancelExport(eid string) error {
	s.exportMu.Lock()
	job, ok := s.exports[eid]
	s.exportMu.Unlock()
	if !ok {
		return errors.New("session: export not found")
	}
	job.cancel()
	return nil
}

type exportJob struct {
	cancel context.CancelFunc
}

func (s *Session) emitExportEvent(ev ExportEvent) {
	s.mu.RLock()
	fn := s.onExportBroadcast
	sid := s.id
	s.mu.RUnlock()
	if fn != nil {
		fn(sid, ev)
	}
}

func (s *Session) exportSQL(ctx context.Context, req ExportRequest) (string, []any, string, error) {
	_ = ctx
	p := s.ensurePager()
	if p == nil {
		return "", nil, "", errSessionNotReady
	}
	src, _, err := p.ResultMeta(req.RID)
	if err != nil {
		return "", nil, "", err
	}
	conn := s.mainConn()
	if conn == nil {
		return "", nil, "", errSessionNotReady
	}
	d := conn.Dialect()
	scope := req.Scope
	if scope == "" {
		scope = ExportScopeResult
	}
	tableName := exportTableLabel(d, src)
	switch src.Kind {
	case pager.SourceTable:
		spec := pager.BrowseSQLSpec(src)
		off, lim := exportWindow(scope, req)
		spec.Offset = off
		spec.Limit = lim
		sqlText, args, err := sqlgen.BuildBrowse(d, spec)
		return sqlText, args, tableName, err
	case pager.SourceQuery:
		sqlText := strings.TrimSpace(strings.TrimSuffix(src.SQL, ";"))
		off, lim := exportWindow(scope, req)
		if off > 0 || lim > 0 {
			sqlText = wrapLimitedQuery(d, sqlText, lim, off)
		}
		return sqlText, nil, tableName, nil
	default:
		return "", nil, "", errors.New("session: unsupported export source")
	}
}

func exportTableLabel(d driver.Dialect, src pager.Source) string {
	if src.Kind != pager.SourceTable {
		return ""
	}
	if src.Schema == "" {
		return d.QuoteIdent(src.Table)
	}
	return d.QuoteIdent(src.Schema) + "." + d.QuoteIdent(src.Table)
}

func exportWindow(scope string, req ExportRequest) (offset, limit int) {
	switch scope {
	case ExportScopePage:
		return req.Offset, req.Limit
	case ExportScopeSelection:
		if req.RowMax < req.RowMin {
			return 0, 0
		}
		return req.RowMin, req.RowMax - req.RowMin + 1
	default:
		return 0, 0
	}
}

func wrapLimitedQuery(d driver.Dialect, inner string, limit, offset int) string {
	q := "SELECT * FROM (" + inner + ") AS _wisp_export"
	if limit > 0 || offset > 0 {
		q += " " + d.LimitOffset(limit, offset)
	}
	return q
}

func colExportStart(min int) int {
	if min < 0 {
		return -1
	}
	return min
}

func colExportEnd(max int) int {
	return max
}

func parseExportFormat(raw string) (mover.Format, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "csv", "":
		return mover.FormatCSV, nil
	case "json":
		return mover.FormatJSON, nil
	case "sql":
		return mover.FormatSQLInsert, nil
	case "markdown", "md":
		return mover.FormatMarkdown, nil
	default:
		return "", errors.New("session: unknown export format")
	}
}
