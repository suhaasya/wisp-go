package session

import (
	"context"
	"errors"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
)

var errSessionNotReady = errors.New("session: not ready")

func (s *Session) ensurePager() *pager.Manager {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pager == nil && s.main != nil {
		s.pager = pager.NewManager(s.main, s.profile.Engine, time.Now)
	}
	return s.pager
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
	return p.Open(ctx, src)
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

// CloseResult closes one result.
func (s *Session) CloseResult(rid string) {
	if p := s.ensurePager(); p != nil {
		p.Close(rid)
	}
}

func (s *Session) closeResults() {
	s.mu.Lock()
	p := s.pager
	s.pager = nil
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
