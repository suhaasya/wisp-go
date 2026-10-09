package session

import (
	"context"
	"errors"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

var errSessionNotReady = errors.New("session: not ready")

func (s *Session) mainConn() driver.Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.main
}

func (s *Session) ensurePager() *pager.Manager {
	s.mu.Lock()
	if s.main == nil {
		s.mu.Unlock()
		return nil
	}
	if s.pager != nil && s.pagerMain != s.main {
		stale := s.pager
		s.pager = nil
		s.pagerMain = nil
		s.mu.Unlock()
		stale.CloseAll()
		s.mu.Lock()
	}
	if s.pager == nil {
		s.pager = pager.NewManager(s.main, s.profile.Engine, time.Now)
		s.pagerMain = s.main
		s.pager.SetOverlayProvider(s.resultOverlay)
	}
	p := s.pager
	s.mu.Unlock()
	return p
}

// OpenResult opens a table browse or query result on the main connection.
func (s *Session) OpenResult(ctx context.Context, src pager.Source) (pager.OpenResult, error) {
	if src.Kind == pager.SourceTable {
		det, err := s.TableMetadata(ctx, src.Schema, src.Table, false)
		if err == nil {
			if len(det.PrimaryKey) > 0 && len(src.KeyColumns) == 0 {
				src.KeyColumns = append([]string(nil), det.PrimaryKey...)
			}
			if len(src.Columns) == 0 {
				for _, c := range det.Columns {
					src.Columns = append(src.Columns, driver.ColumnMeta{Name: c.Name, Type: c.Type})
				}
			}
		}
		main := s.mainConn()
		if main == nil {
			return pager.OpenResult{}, errSessionNotReady
		}
		sqlText, err := sqlgen.BrowseSQLForGuard(main.Dialect(), pager.BrowseSQLSpec(src))
		if err != nil {
			return pager.OpenResult{}, err
		}
		check, err := s.Guard(sqlText, false)
		if err != nil {
			return pager.OpenResult{}, err
		}
		if check == WriteBlocked {
			return pager.OpenResult{}, ErrWriteBlockedReadOnly()
		}
	}
	if src.Kind == pager.SourceQuery {
		check, err := s.Guard(src.SQL, false)
		if err != nil {
			return pager.OpenResult{}, err
		}
		if check == WriteBlocked {
			return pager.OpenResult{}, ErrWriteBlockedReadOnly()
		}
	}
	p := s.ensurePager()
	if p == nil {
		return pager.OpenResult{}, errSessionNotReady
	}
	open, err := p.Open(ctx, src)
	if err != nil {
		return pager.OpenResult{}, err
	}
	s.registerResultChanges(open.RID, src.KeyColumns, src.Schema, src.Table)
	s.wirePagerOverlay()
	return open, nil
}

// ResultWindowJSON returns encoded window bytes for a result.
func (s *Session) ResultWindowJSON(ctx context.Context, rid string, offset, limit int) ([]byte, error) {
	p := s.ensurePager()
	if p == nil {
		return nil, errSessionNotReady
	}
	return p.WindowJSON(ctx, rid, offset, limit)
}

// ExactResultCount runs a cancellable COUNT for a result.
func (s *Session) ExactResultCount(ctx context.Context, rid string) (pager.Total, error) {
	p := s.ensurePager()
	if p == nil {
		return pager.Total{}, errSessionNotReady
	}
	return p.ExactCount(ctx, rid)
}

// ResultCellValue returns a full or preview cell value for the detail panel.
func (s *Session) ResultCellValue(ctx context.Context, rid string, row int, col string) (pager.CellValueResult, error) {
	p := s.ensurePager()
	if p == nil {
		return pager.CellValueResult{}, errSessionNotReady
	}
	return p.CellValue(ctx, rid, row, col)
}

// CloseResult closes one result.
func (s *Session) CloseResult(rid string) {
	s.dropResultChanges(rid)
	if p := s.ensurePager(); p != nil {
		p.Close(rid)
	}
}

func (s *Session) closeResults() {
	s.mu.Lock()
	p := s.pager
	s.pager = nil
	s.pagerMain = nil
	s.mu.Unlock()
	if p != nil {
		p.CloseAll()
	}
}

// EvictIdleResults closes pager results unused for 10 minutes.
func (s *Session) EvictIdleResults() {
	if p := s.ensurePager(); p != nil {
		p.EvictIdle()
	}
}
