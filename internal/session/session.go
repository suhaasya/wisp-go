package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/store"
)

// Session is a live database connection opened from a profile snapshot.
type Session struct {
	id      string
	profile store.Profile

	mu           sync.RWMutex
	state        State
	lastActivity time.Time
	lastTabRef   time.Time

	main driver.Conn
	meta driver.Conn

	pager     *pager.Manager
	pagerMain driver.Conn
	metaCache *metadataCache

	changeMu      sync.RWMutex
	resultChanges map[string]*resultChanges

	metaInitMu sync.Mutex

	releaseSSH       func()
	releaseMetaSSH   func()
	keepCancel       context.CancelFunc
	connector        Connector
	onState          func(id string, st State, info driver.ServerInfo)
	onMetadataBump   func(id string)
	onQueryEvent     func(QueryRunEvent)
	onQueryBroadcast func(sid string, ev QueryRunEvent)
	onJournalBump    func(profileID string)
	onHistoryRecord  func(sql string, ms int64, rows int64, status string)

	runMu          sync.Mutex
	pendingJournal []JournalBinding
	runs           map[string]*runJob
	queryRunRIDs   []string

	serverInfo driver.ServerInfo
	dbPassword []byte
	sshSec     SSHConnectSecrets
}

// Connector opens database connections for a profile (injected in tests).
type Connector interface {
	Connect(ctx context.Context, p store.Profile, password []byte, ssh SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error)
}

// AdoptMainConn wires a database connection (tests).
func (s *Session) AdoptMainConn(c driver.Conn) {
	s.mu.Lock()
	s.main = c
	s.mu.Unlock()
}

// NewTestSession builds a session shell for guard tests (no live connection).
func NewTestSession(p store.Profile) *Session {
	if p.Engine == "" {
		p.Engine = "postgres"
	}
	return &Session{
		id:           "test",
		profile:      p,
		state:        StateReady,
		lastActivity: time.Now(),
		lastTabRef:   time.Now(),
	}
}

// IsWriteBlockedReadOnly reports read-only write guard rejections.
func IsWriteBlockedReadOnly(err error) bool {
	return errors.Is(err, errWriteBlockedReadOnly)
}

// ID returns the session identifier.
func (s *Session) ID() string { return s.id }

// State returns the current lifecycle state.
func (s *Session) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Profile returns the frozen profile snapshot.
func (s *Session) Profile() store.Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile
}

// ServerInfo returns details from the main connection.
func (s *Session) ServerInfo() driver.ServerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serverInfo
}

func (s *Session) setState(st State) {
	s.mu.Lock()
	s.state = st
	info := s.serverInfo
	s.mu.Unlock()
	if s.onState != nil {
		s.onState(s.id, st, info)
	}
}

// TouchTabRef marks that a browser tab still references this session.
func (s *Session) TouchTabRef(now time.Time) {
	s.mu.Lock()
	s.lastTabRef = now
	s.mu.Unlock()
}

func (s *Session) lastTabRefAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastTabRef
}

func (s *Session) keepAliveSec() int {
	if s.profile.Advanced.KeepAliveSec > 0 {
		return s.profile.Advanced.KeepAliveSec
	}
	return defaultKeepSec
}

func (s *Session) withMetaConn(ctx context.Context, fn func(context.Context, driver.Conn) error) error {
	s.metaInitMu.Lock()
	defer s.metaInitMu.Unlock()
	run := func() error {
		conn, err := s.openMetaLocked()
		if err != nil {
			return err
		}
		return fn(ctx, conn)
	}
	err := run()
	if err != nil && isConnError(err) {
		s.closeMetaLocked()
		return run()
	}
	return err
}

func (s *Session) openMetaLocked() (driver.Conn, error) {
	if s.meta != nil {
		return s.meta, nil
	}
	if s.main == nil {
		return nil, errors.New("session: main connection missing")
	}
	connectCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	meta, _, release, err := s.connector.Connect(connectCtx, s.profile, s.dbPassword, s.sshSec)
	if err != nil {
		return nil, err
	}
	s.meta = meta
	s.releaseMetaSSH = release
	return s.meta, nil
}

func (s *Session) closeMetaLocked() {
	if s.meta != nil {
		_ = s.meta.Close()
		s.meta = nil
	}
	if s.releaseMetaSSH != nil {
		s.releaseMetaSSH()
		s.releaseMetaSSH = nil
	}
}

func (s *Session) closeConns() {
	s.metaInitMu.Lock()
	defer s.metaInitMu.Unlock()
	s.closeMetaLocked()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.main != nil {
		_ = s.main.Close()
		s.main = nil
	}
	if s.releaseSSH != nil {
		s.releaseSSH()
		s.releaseSSH = nil
	}
}

func (s *Session) shutdown() {
	if s.keepCancel != nil {
		s.keepCancel()
		s.keepCancel = nil
	}
	s.closeResults()
	s.closeConns()
	s.setState(StateClosed)
}
