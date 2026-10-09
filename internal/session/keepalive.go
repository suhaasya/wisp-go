package session

import (
	"context"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
)

const (
	reconnectMaxBackoff = 30 * time.Second
	reconnectGiveUp     = 5 * time.Minute
)

var reconnectBackoffSteps = []time.Duration{
	500 * time.Millisecond,
	time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
}

func (s *Session) runKeepAlive(ctx context.Context, sleep func(time.Duration)) {
	interval := time.Duration(s.keepAliveSec()) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.keepAliveOnce(ctx, sleep)
		}
	}
}

func (s *Session) keepAliveOnce(ctx context.Context, sleep func(time.Duration)) {
	st := s.State()
	if st == StateBusy || st == StateClosed || st == StateFailed || st == StateReconnecting {
		return
	}
	err := s.withMetaConn(ctx, func(ctx context.Context, meta driver.Conn) error {
		_, err := meta.Ping(ctx)
		return err
	})
	if err != nil {
		s.startReconnect(ctx, sleep, err)
	}
}

func (s *Session) startReconnect(ctx context.Context, sleep func(time.Duration), cause error) {
	_ = cause
	if s.State() == StateReconnecting || s.State() == StateClosed {
		return
	}
	s.setState(StateReconnecting)
	deadline := time.Now().Add(reconnectGiveUp)
	step := 0
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return
		}
		wait := reconnectMaxBackoff
		if step < len(reconnectBackoffSteps) {
			wait = reconnectBackoffSteps[step]
		}
		if wait > reconnectMaxBackoff {
			wait = reconnectMaxBackoff
		}
		sleep(wait)
		step++
		if s.tryReconnect(ctx) {
			return
		}
	}
	s.setState(StateFailed)
}

func (s *Session) tryReconnect(ctx context.Context) bool {
	s.closeResults()
	s.closeConns()
	main, info, release, err := s.connector.Connect(ctx, s.profile, s.dbPassword, s.sshSec)
	if err != nil {
		return false
	}
	if _, err := main.Ping(ctx); err != nil {
		_ = main.Close()
		release()
		return false
	}
	s.mu.Lock()
	s.main = main
	s.serverInfo = info
	s.releaseSSH = release
	s.mu.Unlock()
	s.InvalidateMetadata()
	s.setState(StateReady)
	return true
}
