package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

const maxQueryRunResults = 5

// RunMode selects which statements to execute from a script.
type RunMode string

const (
	RunModeCursor    RunMode = "cursor"
	RunModeSelection RunMode = "selection"
	RunModeAll       RunMode = "all"
)

// RunRequest starts a script run.
type RunRequest struct {
	SQL      string
	Mode     RunMode
	Cursor   int
	SelStart int
	SelEnd   int
	Confirm  bool
}

// QueryRunEvent reports progress for one statement or run completion.
type QueryRunEvent struct {
	QID          string `json:"qid"`
	Done         bool   `json:"done,omitempty"`
	Cancelled    bool   `json:"cancelled,omitempty"`
	StmtIndex    int    `json:"stmt_index"`
	StmtTotal    int    `json:"stmt_total"`
	StmtStart    int    `json:"stmt_start,omitempty"`
	RID          string `json:"rid,omitempty"`
	RowsAffected int64  `json:"rows_affected,omitempty"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDetail  string `json:"error_detail,omitempty"`
	Position     int    `json:"position,omitempty"`
	NeedsConfirm bool   `json:"needs_confirm,omitempty"`
}

type runJob struct {
	cancel context.CancelFunc
}

func newQueryID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// SetQueryEventHook records run events (tests).
func (s *Session) SetQueryEventHook(fn func(QueryRunEvent)) {
	s.mu.Lock()
	s.onQueryEvent = fn
	s.mu.Unlock()
}

func (s *Session) emitQueryEvent(ev QueryRunEvent) {
	s.mu.RLock()
	fn := s.onQueryEvent
	sid := s.id
	s.mu.RUnlock()
	if fn != nil {
		fn(ev)
	}
	if s.onQueryBroadcast != nil {
		s.onQueryBroadcast(sid, ev)
	}
}

// StartRun executes a script asynchronously and returns a query id.
func (s *Session) StartRun(parent context.Context, req RunRequest) (string, error) {
	if s == nil {
		return "", errors.New("session: nil")
	}
	stmts, err := statementsForRun(req)
	if err != nil {
		return "", err
	}
	if len(stmts) == 0 {
		return "", errors.New("session: no statements to run")
	}
	qid := newQueryID()
	runCtx, cancel := context.WithCancel(context.WithoutCancel(parent))
	s.runMu.Lock()
	if s.runs == nil {
		s.runs = make(map[string]*runJob)
	}
	s.runs[qid] = &runJob{cancel: cancel}
	s.runMu.Unlock()

	go s.runScript(runCtx, qid, req.Confirm, stmts)
	return qid, nil
}

// CancelRun cancels an in-flight script.
func (s *Session) CancelRun(qid string) error {
	s.runMu.Lock()
	job, ok := s.runs[qid]
	s.runMu.Unlock()
	if !ok {
		return errors.New("session: query not found")
	}
	job.cancel()
	return nil
}

type stmtSlice struct {
	text  string
	start int
}

func statementsForRun(req RunRequest) ([]stmtSlice, error) {
	sql := req.SQL
	switch req.Mode {
	case RunModeAll, "":
		ranges := sqlgen.SplitStatements(sql)
		out := make([]stmtSlice, 0, len(ranges))
		for _, r := range ranges {
			text := strings.TrimSpace(sql[r.Start:r.End])
			if text == "" {
				continue
			}
			out = append(out, stmtSlice{text: text, start: r.Start})
		}
		return out, nil
	case RunModeCursor:
		r, ok := sqlgen.StatementAtOffset(sql, req.Cursor)
		if !ok {
			return nil, errors.New("session: no statement at cursor")
		}
		text := strings.TrimSpace(sql[r.Start:r.End])
		if text == "" {
			return nil, errors.New("session: empty statement at cursor")
		}
		return []stmtSlice{{text: text, start: r.Start}}, nil
	case RunModeSelection:
		if req.SelStart < 0 || req.SelEnd <= req.SelStart || req.SelEnd > len(sql) {
			return nil, errors.New("session: invalid selection")
		}
		sel := strings.TrimSpace(sql[req.SelStart:req.SelEnd])
		if sel == "" {
			return nil, errors.New("session: empty selection")
		}
		ranges := sqlgen.SplitStatements(sel)
		out := make([]stmtSlice, 0, len(ranges))
		for _, r := range ranges {
			text := strings.TrimSpace(sel[r.Start:r.End])
			if text == "" {
				continue
			}
			out = append(out, stmtSlice{text: text, start: req.SelStart + r.Start})
		}
		return out, nil
	default:
		return nil, errors.New("session: unknown run mode")
	}
}

func (s *Session) runScript(ctx context.Context, qid string, confirm bool, stmts []stmtSlice) {
	defer func() {
		s.runMu.Lock()
		delete(s.runs, qid)
		s.runMu.Unlock()
	}()
	s.setState(StateBusy)
	defer s.setState(StateReady)

	total := len(stmts)
	for i, st := range stmts {
		if err := ctx.Err(); err != nil {
			s.emitQueryEvent(QueryRunEvent{
				QID: qid, Done: true, Cancelled: true,
				StmtIndex: i, StmtTotal: total,
			})
			return
		}
		ev := s.runOneStatement(ctx, qid, confirm, i, total, st)
		s.emitQueryEvent(ev)
		if ev.NeedsConfirm || ev.Error != "" {
			s.emitQueryEvent(QueryRunEvent{QID: qid, Done: true, StmtIndex: i, StmtTotal: total})
			return
		}
	}
	s.emitQueryEvent(QueryRunEvent{QID: qid, Done: true, StmtIndex: total, StmtTotal: total})
}

func (s *Session) runOneStatement(ctx context.Context, qid string, confirm bool, idx, total int, st stmtSlice) QueryRunEvent {
	ev := QueryRunEvent{
		QID:       qid,
		StmtIndex: idx + 1,
		StmtTotal: total,
		StmtStart: st.start,
	}
	check, err := s.Guard(st.text, confirm)
	if err != nil {
		ev.Error = "guard_failed"
		ev.ErrorDetail = err.Error()
		ev.Position = st.start
		return ev
	}
	if check == WriteNeedsConfirmation {
		ev.NeedsConfirm = true
		ev.Position = st.start
		return ev
	}
	if check == WriteBlocked {
		ev.Error = "write_blocked"
		ev.ErrorDetail = ErrWriteBlockedReadOnly().Error()
		ev.Position = st.start
		return ev
	}

	start := time.Now()
	cls := ClassifyStatement(st.text)
	if cls == StatementRead {
		schema, table, keys := s.queryResultSource(ctx, st.text)
		open, err := s.OpenResult(ctx, pager.Source{
			Kind:       pager.SourceQuery,
			SQL:        st.text,
			Schema:     schema,
			Table:      table,
			KeyColumns: keys,
		})
		if err != nil {
			ev.Error = "result_open_failed"
			ev.ErrorDetail = err.Error()
			ev.Position = st.start
			ev.DurationMs = time.Since(start).Milliseconds()
			return ev
		}
		s.trackQueryRunResult(open.RID)
		ev.RID = open.RID
		ev.DurationMs = time.Since(start).Milliseconds()
		return ev
	}

	conn := s.mainConn()
	if conn == nil {
		ev.Error = "session_not_ready"
		ev.Position = st.start
		return ev
	}
	res, err := conn.Exec(ctx, st.text)
	ev.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		ev.Error = "exec_failed"
		ev.ErrorDetail = err.Error()
		ev.Position = st.start
		return ev
	}
	ev.RowsAffected = res.RowsAffected
	return ev
}

func (s *Session) trackQueryRunResult(rid string) {
	s.runMu.Lock()
	s.queryRunRIDs = append(s.queryRunRIDs, rid)
	var toClose []string
	for len(s.queryRunRIDs) > maxQueryRunResults {
		toClose = append(toClose, s.queryRunRIDs[0])
		s.queryRunRIDs = s.queryRunRIDs[1:]
	}
	s.runMu.Unlock()
	for _, old := range toClose {
		s.CloseResult(old)
	}
}

// QueryRunResultIDs returns rids kept for query runs (tests).
func (s *Session) QueryRunResultIDs() []string {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	out := append([]string(nil), s.queryRunRIDs...)
	return out
}
