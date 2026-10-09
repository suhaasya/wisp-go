package session

import (
	"context"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
)

func (s *Session) ensureMetaPager(ctx context.Context) (*pager.Manager, error) {
	if err := s.withMetaConn(ctx, func(context.Context, driver.Conn) error { return nil }); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.meta == nil {
		return nil, errSessionNotReady
	}
	if s.metaPager != nil && s.metaPagerConn == s.meta {
		return s.metaPager, nil
	}
	if s.metaPager != nil {
		stale := s.metaPager
		s.metaPager = nil
		s.metaPagerConn = nil
		go stale.CloseAll()
	}
	s.metaPager = pager.NewManager(s.meta, s.profile.Engine, time.Now)
	s.metaPagerConn = s.meta
	return s.metaPager, nil
}

func (s *Session) metaPagerForRID(rid string) *pager.Manager {
	s.mu.RLock()
	mp := s.metaPager
	s.mu.RUnlock()
	if mp == nil {
		return nil
	}
	if _, _, err := mp.ResultMeta(rid); err != nil {
		return nil
	}
	return mp
}

func (s *Session) pagerForRID(rid string) *pager.Manager {
	if mp := s.metaPagerForRID(rid); mp != nil {
		return mp
	}
	if p := s.ensurePager(); p != nil {
		if _, _, err := p.ResultMeta(rid); err == nil {
			return p
		}
	}
	return nil
}
