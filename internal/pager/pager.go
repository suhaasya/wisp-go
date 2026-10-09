package pager

import (
	"context"
	"errors"
	"sync"
	"time"

	"crypto/rand"
	"encoding/hex"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

const (
	DefaultWindowRows = 150
	MaxCachedWindows  = 3
	ResultIdleTimeout = 10 * time.Minute
	ExactCountMax     = 100_000
)

const (
	SourceTable = "table"
	SourceQuery = "query"
)

// Source describes how to open a result.
type Source struct {
	Kind       string              `json:"kind"`
	Schema     string              `json:"schema,omitempty"`
	Table      string              `json:"table,omitempty"`
	SQL        string              `json:"sql,omitempty"`
	KeyColumns []string            `json:"-"`
	Columns    []driver.ColumnMeta `json:"-"`
	Filter     sqlgen.FilterGroup  `json:"filter,omitempty"`
	RawWhere   string              `json:"rawWhere,omitempty"`
	Sort       []sqlgen.SortKey    `json:"sort,omitempty"`
}

// Total describes row count kind and value.
type Total struct {
	Kind  string `json:"kind"`
	Value int64  `json:"value"`
}

// OpenResult is returned when a result is opened.
type OpenResult struct {
	RID     string              `json:"rid"`
	Columns []driver.ColumnMeta `json:"columns"`
	Total   Total               `json:"total"`
}

// Manager owns results for one database session (main connection).
type Manager struct {
	mu        sync.Mutex
	results   map[string]*result
	conn      driver.Conn
	engine    string
	mockTotal int
	clock     func() time.Time
	newID     func() string
	overlay   func(rid string) Overlay
}

// NewManager returns a pager for session main conn.
func NewManager(conn driver.Conn, engine string, clock func() time.Time) *Manager {
	if clock == nil {
		clock = time.Now
	}
	return &Manager{
		results: make(map[string]*result),
		conn:    conn,
		engine:  engine,
		clock:   clock,
		newID:   newResultID,
		overlay: func(string) Overlay { return Overlay{} },
	}
}

// SetMockTotal sets row count for mock-engine table/query results (tests).
func (m *Manager) SetMockTotal(n int) {
	m.mockTotal = n
}

// SetOverlayProvider sets staged-edit overlay lookup (LW-029).
func (m *Manager) SetOverlayProvider(fn func(rid string) Overlay) {
	if fn != nil {
		m.overlay = fn
	}
}

// CachedWindowCount returns cached windows for a result (tests).
func (m *Manager) CachedWindowCount(rid string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.results[rid]
	if !ok {
		return 0
	}
	return r.cache.len()
}

// Open starts a new result.
func (m *Manager) Open(ctx context.Context, src Source) (OpenResult, error) {
	if m.conn == nil {
		return OpenResult{}, errors.New("pager: no connection")
	}
	f, total, err := m.newFetcher(ctx, src)
	if err != nil {
		return OpenResult{}, err
	}
	rid := m.newID()
	r := &result{
		id:      rid,
		source:  src,
		fetch:   f,
		total:   total,
		columns: f.Columns(),
		cache:   newWindowCache(MaxCachedWindows),
		lastUse: m.clock(),
	}
	m.mu.Lock()
	m.results[rid] = r
	m.mu.Unlock()
	return OpenResult{
		RID:     rid,
		Columns: r.columns,
		Total:   total,
	}, nil
}

// WindowJSON returns PRD columnar window JSON for a logical offset.
func (m *Manager) WindowJSON(ctx context.Context, rid string, offset, limit int) ([]byte, error) {
	if limit <= 0 {
		limit = DefaultWindowRows
	}
	m.mu.Lock()
	r, ok := m.results[rid]
	if !ok {
		m.mu.Unlock()
		return nil, errNotFound
	}
	r.lastUse = m.clock()
	m.mu.Unlock()

	if w, ok := r.cache.get(offset, limit); ok {
		return buildWindowJSON(r.id, offset, r.total, w, m.overlay(r.id))
	}
	w, err := r.fetch.Fetch(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	if w == nil {
		w = driver.NewWindow(r.columns, 0)
	}
	r.cache.put(offset, limit, w)
	applyOverlay(w, m.overlay(r.id))
	return buildWindowJSON(r.id, offset, r.total, w, m.overlay(r.id))
}

// Close removes one result.
func (m *Manager) Close(rid string) {
	m.mu.Lock()
	r, ok := m.results[rid]
	delete(m.results, rid)
	m.mu.Unlock()
	if ok && r.fetch != nil {
		_ = r.fetch.Close()
	}
}

// CloseAll closes every result.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.results))
	for id := range m.results {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		m.Close(id)
	}
}

// EvictIdle closes results unused longer than ResultIdleTimeout.
func (m *Manager) EvictIdle() {
	cutoff := m.clock().Add(-ResultIdleTimeout)
	m.mu.Lock()
	var idle []string
	for id, r := range m.results {
		if r.lastUse.Before(cutoff) {
			idle = append(idle, id)
		}
	}
	m.mu.Unlock()
	for _, id := range idle {
		m.Close(id)
	}
}

var errNotFound = errors.New("pager: result not found")

func newResultID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

type result struct {
	id      string
	source  Source
	fetch   fetcher
	total   Total
	columns []driver.ColumnMeta
	cache   windowCache
	lastUse time.Time
}

type fetcher interface {
	Columns() []driver.ColumnMeta
	Fetch(ctx context.Context, offset, limit int) (*driver.Window, error)
	Close() error
}

func (m *Manager) newFetcher(ctx context.Context, src Source) (fetcher, Total, error) {
	switch src.Kind {
	case SourceTable:
		return m.newTableFetcher(ctx, src)
	case SourceQuery:
		return m.newQueryFetcher(ctx, src)
	default:
		return nil, Total{}, errors.New("pager: unknown source kind")
	}
}
