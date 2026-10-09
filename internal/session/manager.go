package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/store"
)

const tabIdleClose = 30 * time.Minute

// Broadcaster emits SSE session events.
type Broadcaster interface {
	Broadcast(typ string, data any)
}

// Manager owns live database sessions.
type Manager struct {
	mu        sync.Mutex
	sessions  map[string]*Session
	profiles  *store.ProfileStore
	secrets   store.Secrets
	broadcast Broadcaster
	connector Connector
	clock     func() time.Time
	sleep     func(time.Duration)
	newID     func() string

	stopEvict context.CancelFunc

	journal   *JournalStore
	workspace *WorkspaceStore
	history   *store.HistoryStore
}

// NewManager returns a session manager wired to profiles and secrets.
func NewManager(profiles *store.ProfileStore, secrets store.Secrets, bc Broadcaster) *Manager {
	m := &Manager{
		sessions:  make(map[string]*Session),
		profiles:  profiles,
		secrets:   secrets,
		broadcast: bc,
		connector: DefaultConnector{},
		clock:     time.Now,
		sleep:     time.Sleep,
		newID:     store.NewUUIDv7,
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.stopEvict = cancel
	go m.evictLoop(ctx)
	return m
}

// SetJournal wires crash-safe persistence (LW-034).
func (m *Manager) SetJournal(j *JournalStore, ws *WorkspaceStore) {
	m.mu.Lock()
	m.journal = j
	m.workspace = ws
	m.mu.Unlock()
}

// SetHistory wires query history persistence.
func (m *Manager) SetHistory(h *store.HistoryStore) {
	m.mu.Lock()
	m.history = h
	m.mu.Unlock()
}

// SetConnector replaces the connector (tests).
func (m *Manager) SetConnector(c Connector) {
	m.mu.Lock()
	m.connector = c
	m.mu.Unlock()
}

// SetClock replaces time source (tests).
func (m *Manager) SetClock(clock func() time.Time) {
	m.mu.Lock()
	m.clock = clock
	m.mu.Unlock()
}

// SetSleep replaces sleep (tests).
func (m *Manager) SetSleep(sleep func(time.Duration)) {
	m.mu.Lock()
	m.sleep = sleep
	m.mu.Unlock()
}

// CloseAll closes every session (shutdown).
func (m *Manager) CloseAll(ctx context.Context) error {
	_ = ctx
	m.mu.Lock()
	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	profileIDs := make(map[string]struct{})
	m.mu.Lock()
	for _, id := range ids {
		if s, ok := m.sessions[id]; ok {
			profileIDs[s.profile.ID] = struct{}{}
		}
	}
	journal := m.journal
	m.mu.Unlock()
	for _, id := range ids {
		_ = m.Delete(id)
	}
	if journal != nil {
		for pid := range profileIDs {
			_ = journal.Clear(pid)
		}
	}
	if m.stopEvict != nil {
		m.stopEvict()
	}
	return nil
}

// AdoptSession registers a session shell (tests).
func (m *Manager) AdoptSession(s *Session) {
	if s == nil {
		return
	}
	if s.onMetadataBump == nil && m.broadcast != nil {
		s.onMetadataBump = func(sid string) {
			m.broadcast.Broadcast("schema", map[string]any{"sid": sid})
		}
	}
	if s.onQueryBroadcast == nil && m.broadcast != nil {
		s.onQueryBroadcast = func(sid string, ev QueryRunEvent) {
			m.broadcast.Broadcast("query", queryEventPayload(sid, ev))
		}
	}
	if s.onExportBroadcast == nil && m.broadcast != nil {
		s.onExportBroadcast = func(sid string, ev ExportEvent) {
			m.broadcast.Broadcast("export", map[string]any{
				"sid":       sid,
				"eid":       ev.EID,
				"rows":      ev.Rows,
				"done":      ev.Done,
				"cancelled": ev.Cancelled,
				"error":     ev.Error,
			})
		}
	}
	m.mu.Lock()
	m.sessions[s.ID()] = s
	m.mu.Unlock()
}

// Get returns a live session by id.
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	return s, ok
}

// Delete closes and removes a session.
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	if !ok {
		return errors.New("session not found")
	}
	s.shutdown()
	return nil
}

// Create opens a session for profileID.
func (m *Manager) Create(ctx context.Context, profileID string) (*Session, driver.ServerInfo, error) {
	doc, err := m.profiles.Load()
	if err != nil {
		return nil, driver.ServerInfo{}, err
	}
	var prof *store.Profile
	for i := range doc.Connections {
		if doc.Connections[i].ID == profileID {
			prof = &doc.Connections[i]
			break
		}
	}
	if prof == nil {
		return nil, driver.ServerInfo{}, errors.New("profile not found")
	}
	pw, err := m.loadDBPassword(profileID)
	if err != nil {
		return nil, driver.ServerInfo{}, err
	}
	sshSec, err := m.loadSSHSecrets(profileID)
	if err != nil {
		return nil, driver.ServerInfo{}, err
	}
	snapshot := *prof
	id := m.newID()
	s := &Session{
		id:         id,
		profile:    snapshot,
		state:      StateConnecting,
		connector:  m.connector,
		dbPassword: pw,
		sshSec:     sshSec,
		onState:    m.emitState,
		onMetadataBump: func(sid string) {
			if m.broadcast != nil {
				m.broadcast.Broadcast("schema", map[string]any{"sid": sid})
			}
		},
		onQueryBroadcast: func(sid string, ev QueryRunEvent) {
			if m.broadcast != nil {
				m.broadcast.Broadcast("query", queryEventPayload(sid, ev))
			}
		},
		onExportBroadcast: func(sid string, ev ExportEvent) {
			if m.broadcast != nil {
				m.broadcast.Broadcast("export", map[string]any{
					"sid":       sid,
					"eid":       ev.EID,
					"rows":      ev.Rows,
					"done":      ev.Done,
					"cancelled": ev.Cancelled,
					"error":     ev.Error,
				})
			}
		},
		onJournalBump: func(pid string) { m.scheduleJournal(pid) },
		onHistoryRecord: func(sql string, ms int64, rows int64, status string) {
			m.RecordHistory(snapshot, store.HistoryEntry{
				TS:     m.clock().UnixMilli(),
				SQL:    sql,
				MS:     ms,
				Rows:   rows,
				Status: status,
			})
		},
	}
	if m.journal != nil {
		if doc, err := m.journal.Restore(profileID); err == nil {
			s.pendingJournal = doc.Bindings
		}
	}
	now := m.clock()
	s.lastActivity = now
	s.lastTabRef = now

	main, info, release, err := m.connector.Connect(ctx, snapshot, pw, sshSec)
	if err != nil {
		return nil, driver.ServerInfo{}, err
	}
	s.main = main
	s.serverInfo = info
	s.releaseSSH = release
	s.setState(StateReady)

	keepCtx, keepCancel := context.WithCancel(context.Background())
	s.keepCancel = keepCancel
	go s.runKeepAlive(keepCtx, m.sleep)

	m.mu.Lock()
	m.sessions[id] = s
	m.mu.Unlock()
	return s, info, nil
}

func (m *Manager) loadDBPassword(profileID string) ([]byte, error) {
	if m.secrets == nil {
		return nil, nil
	}
	b, ok, err := m.secrets.Get(profileID, store.SecretKindDBPassword)
	if err != nil {
		return nil, err
	}
	if ok {
		return b, nil
	}
	return nil, nil
}

func (m *Manager) loadSSHSecrets(profileID string) (SSHConnectSecrets, error) {
	out := SSHConnectSecrets{}
	if m.secrets == nil {
		return out, nil
	}
	if b, ok, err := m.secrets.Get(profileID, store.SecretKindSSHPassword); err != nil {
		return out, err
	} else if ok {
		out.Password = b
	}
	if b, ok, err := m.secrets.Get(profileID, store.SecretKindSSHPassphrase); err != nil {
		return out, err
	} else if ok {
		out.Passphrase = b
	}
	return out, nil
}

// ScheduleJournal debounces a journal flush for profileID.
func (m *Manager) ScheduleJournal(profileID string) {
	m.scheduleJournal(profileID)
}

func (m *Manager) scheduleJournal(profileID string) {
	m.mu.Lock()
	j := m.journal
	ws := m.workspace
	m.mu.Unlock()
	if j == nil || ws == nil {
		return
	}
	j.ScheduleFlush(profileID, func() (JournalDocument, error) {
		st, err := ws.Load(profileID)
		if err != nil {
			return JournalDocument{}, err
		}
		m.mu.Lock()
		var list []*Session
		for _, sess := range m.sessions {
			if sess.profile.ID == profileID {
				list = append(list, sess)
			}
		}
		m.mu.Unlock()
		return SnapshotJournal(profileID, st, list), nil
	})
}

func (m *Manager) RecordHistory(profile store.Profile, entry store.HistoryEntry) {
	m.mu.Lock()
	h := m.history
	m.mu.Unlock()
	if h == nil {
		return
	}
	_ = h.Append(profile, entry)
}

func (m *Manager) emitState(id string, st State, info driver.ServerInfo) {
	if m.broadcast == nil {
		return
	}
	m.broadcast.Broadcast("session", map[string]any{
		"sid":    id,
		"state":  string(st),
		"engine": info.Engine,
	})
}

func (m *Manager) evictLoop(ctx context.Context) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			m.evictIdle()
		}
	}
}

func (m *Manager) evictIdle() {
	now := m.clock()
	m.mu.Lock()
	var idle []string
	for id, s := range m.sessions {
		if now.Sub(s.lastTabRefAt()) >= tabIdleClose {
			idle = append(idle, id)
		}
	}
	m.mu.Unlock()
	for _, id := range idle {
		_ = m.Delete(id)
	}
	m.mu.Lock()
	for _, s := range m.sessions {
		s.EvictIdleResults()
	}
	m.mu.Unlock()
}
