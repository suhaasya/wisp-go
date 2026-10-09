package session

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/suhaasya/wisp-go/internal/mover"
)

// ImportPreviewRequest selects a file to sample.
type ImportPreviewRequest struct {
	Path string
}

// ImportColumnMapping maps file columns to table columns for import.
type ImportColumnMapping struct {
	FileColumn  string `json:"file_column"`
	TableColumn string `json:"table_column"`
}

// ImportStartRequest starts a background import job.
type ImportStartRequest struct {
	Path        string
	Kind        string
	Schema      string
	Table       string
	Delimiter   string
	HasHeader   bool
	NullMarkers []string
	BatchSize   int
	TxMode      string
	StopOnError bool
	Confirm     bool
	Mapping     []ImportColumnMapping
	FileColumns []string
}

// ImportEvent reports import progress over SSE.
type ImportEvent struct {
	IID       string `json:"iid"`
	Rows      int64  `json:"rows"`
	Bytes     int64  `json:"bytes,omitempty"`
	Done      bool   `json:"done,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
	Error     string `json:"error,omitempty"`
}

// PreviewImport reads the first rows from a local file path.
func (s *Session) PreviewImport(ctx context.Context, req ImportPreviewRequest) (mover.FilePreview, error) {
	if strings.TrimSpace(req.Path) == "" {
		return mover.FilePreview{}, errors.New("session: import path required")
	}
	return mover.PreviewImportFile(ctx, req.Path)
}

// StartImport runs an import asynchronously.
func (s *Session) StartImport(parent context.Context, req ImportStartRequest) (string, error) {
	if strings.TrimSpace(req.Path) == "" {
		return "", errors.New("session: import path required")
	}
	iid := newQueryID()
	runCtx, cancel := context.WithCancel(context.WithoutCancel(parent))
	s.importMu.Lock()
	if s.imports == nil {
		s.imports = make(map[string]*importJob)
	}
	s.imports[iid] = &importJob{cancel: cancel, errors: make([]mover.ImportError, 0)}
	s.importMu.Unlock()

	go s.runImport(runCtx, iid, req)
	return iid, nil
}

// CancelImport cancels an in-flight import.
func (s *Session) CancelImport(iid string) error {
	s.importMu.Lock()
	job, ok := s.imports[iid]
	s.importMu.Unlock()
	if !ok {
		return errors.New("session: import not found")
	}
	job.cancel()
	return nil
}

// ImportErrorsCSV returns the error report for a finished import job.
func (s *Session) ImportErrorsCSV(iid string) (string, bool) {
	s.importMu.Lock()
	job, ok := s.imports[iid]
	s.importMu.Unlock()
	if !ok {
		return "", false
	}
	job.mu.Lock()
	out := mover.FormatImportErrorsCSV(job.errors)
	job.mu.Unlock()
	return out, true
}

// ReleaseImport drops a finished import job from memory.
func (s *Session) ReleaseImport(iid string) {
	s.importMu.Lock()
	delete(s.imports, iid)
	s.importMu.Unlock()
}

type importJob struct {
	cancel context.CancelFunc
	mu     sync.Mutex
	errors []mover.ImportError
}

func (s *Session) runImport(ctx context.Context, iid string, req ImportStartRequest) {
	kind := mover.ImportKind(req.Kind)
	if kind == "" {
		kind = mover.DetectImportKind(req.Path)
	}
	conn := s.mainConn()
	if conn == nil {
		s.emitImportEvent(ImportEvent{IID: iid, Error: errSessionNotReady.Error(), Done: true})
		return
	}
	d := conn.Dialect()
	exec := func(ctx context.Context, sql string, args ...any) error {
		check, err := s.Guard(sql, req.Confirm)
		if err != nil {
			return err
		}
		if check == WriteBlocked {
			return ErrWriteBlockedReadOnly()
		}
		if check == WriteNeedsConfirmation {
			return errors.New("session: import needs confirmation")
		}
		_, err = conn.Exec(ctx, sql, args...)
		return err
	}
	var loaded int64
	var importErrs []mover.ImportError
	var err error
	switch kind {
	case mover.ImportKindSQL:
		var stmts int
		stmts, err = mover.ImportSQLFile(ctx, conn, req.Path, func(ctx context.Context, sql string) error {
			return exec(ctx, sql)
		}, func(p mover.SQLImportProgress) {
			s.emitImportEvent(ImportEvent{IID: iid, Bytes: p.BytesRead})
		})
		loaded = int64(stmts)
	default:
		opts := mover.DataImportOptions{
			Schema:      req.Schema,
			Table:       req.Table,
			Delimiter:   req.Delimiter,
			HasHeader:   req.HasHeader,
			NullMarkers: req.NullMarkers,
			BatchSize:   req.BatchSize,
			StopOnError: req.StopOnError,
			Mapping:     moverMappings(req.Mapping),
			FileColumns: req.FileColumns,
		}
		progress := func(p mover.ImportProgress) {
			s.emitImportEvent(ImportEvent{IID: iid, Rows: p.RowsLoaded})
		}
		if kind == mover.ImportKindJSONL {
			loaded, importErrs, err = mover.ImportJSONL(ctx, d, conn, req.Path, opts, exec, progress)
		} else {
			loaded, importErrs, err = mover.ImportCSV(ctx, d, conn, req.Path, opts, exec, progress)
		}
	}
	s.storeImportErrors(iid, importErrs)
	if err != nil {
		if ctx.Err() != nil {
			s.emitImportEvent(ImportEvent{IID: iid, Cancelled: true, Done: true})
			return
		}
		s.emitImportEvent(ImportEvent{IID: iid, Error: err.Error(), Done: true})
		return
	}
	s.emitImportEvent(ImportEvent{IID: iid, Rows: loaded, Done: true})
	s.InvalidateMetadata()
}

func (s *Session) storeImportErrors(iid string, errs []mover.ImportError) {
	s.importMu.Lock()
	job, ok := s.imports[iid]
	s.importMu.Unlock()
	if !ok {
		return
	}
	job.mu.Lock()
	job.errors = append(job.errors, errs...)
	if len(job.errors) > mover.MaxImportErrors {
		job.errors = job.errors[:mover.MaxImportErrors]
	}
	job.mu.Unlock()
}

func (s *Session) emitImportEvent(ev ImportEvent) {
	s.mu.RLock()
	fn := s.onImportBroadcast
	hook := s.onImportEvent
	s.mu.RUnlock()
	if hook != nil {
		hook(ev)
	}
	if fn != nil {
		fn(s.ID(), ev)
	}
}

// SetImportEventHook records import events (tests).
func (s *Session) SetImportEventHook(fn func(ImportEvent)) {
	s.mu.Lock()
	s.onImportEvent = fn
	s.mu.Unlock()
}

func moverMappings(in []ImportColumnMapping) []mover.ColumnMapping {
	out := make([]mover.ColumnMapping, len(in))
	for i, m := range in {
		out[i] = mover.ColumnMapping{FileColumn: m.FileColumn, TableColumn: m.TableColumn}
	}
	return out
}
