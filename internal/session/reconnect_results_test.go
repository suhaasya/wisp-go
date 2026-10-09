package session

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/store"
)

// closedAwareConn marks itself closed so Query fails like database/sql after Close.
type closedAwareConn struct {
	driver.Conn
	closed atomic.Bool
}

func (c *closedAwareConn) Close() error {
	c.closed.Store(true)
	return c.Conn.Close()
}

func (c *closedAwareConn) Query(ctx context.Context, sql string, args ...any) (driver.Rows, error) {
	if c.closed.Load() {
		return nil, errors.New("sql: connection is already closed")
	}
	return c.Conn.Query(ctx, sql, args...)
}

func (c *closedAwareConn) Exec(ctx context.Context, sql string, args ...any) (driver.ExecResult, error) {
	if c.closed.Load() {
		return driver.ExecResult{}, errors.New("sql: connection is already closed")
	}
	return c.Conn.Exec(ctx, sql, args...)
}

type countingMockConnector struct {
	opens atomic.Int32
}

func (m *countingMockConnector) Connect(ctx context.Context, _ store.Profile, _ []byte, _ SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error) {
	m.opens.Add(1)
	cfg := mock.Config{
		TotalRows: 100,
		Columns: []driver.ColumnMeta{
			{Name: "id", Type: "int8"},
			{Name: "name", Type: "text"},
		},
	}
	raw, err := mock.Driver{}.Open(ctx, driver.ConnectParams{Extra: cfg})
	if err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	return &closedAwareConn{Conn: raw}, driver.ServerInfo{Engine: "mock"}, func() {}, nil
}

func TestTryReconnectOpenResultUsesNewMain(t *testing.T) {
	mc := &countingMockConnector{}
	s := &Session{
		profile:   store.Profile{Engine: "mock"},
		state:     StateReady,
		connector: mc,
	}
	ctx := t.Context()
	src := pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "items",
		KeyColumns: []string{"id"},
	}
	main, _, release, err := mc.Connect(ctx, s.profile, nil, SSHConnectSecrets{})
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	s.main = main
	if _, err := s.OpenResult(ctx, src); err != nil {
		t.Fatalf("first open: %v", err)
	}
	if s.pager == nil {
		t.Fatal("pager should exist after open")
	}
	if !s.tryReconnect(ctx) {
		t.Fatal("tryReconnect failed")
	}
	if s.pager != nil {
		t.Fatal("pager should be cleared on reconnect")
	}
	if _, err := s.OpenResult(ctx, src); err != nil {
		t.Fatalf("open after reconnect: %v", err)
	}
}

func TestEnsurePagerRebindsWhenMainReplaced(t *testing.T) {
	mc := &countingMockConnector{}
	s := &Session{
		profile:   store.Profile{Engine: "mock"},
		state:     StateReady,
		connector: mc,
	}
	ctx := t.Context()
	src := pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "items",
		KeyColumns: []string{"id"},
	}
	main1, _, _, err := mc.Connect(ctx, s.profile, nil, SSHConnectSecrets{})
	if err != nil {
		t.Fatal(err)
	}
	w1 := main1.(*closedAwareConn)
	s.main = main1
	if _, err := s.OpenResult(ctx, src); err != nil {
		t.Fatalf("first open: %v", err)
	}
	main2, _, _, err := mc.Connect(ctx, s.profile, nil, SSHConnectSecrets{})
	if err != nil {
		t.Fatal(err)
	}
	_ = w1.Close()
	s.main = main2
	if _, err := s.OpenResult(ctx, src); err != nil {
		t.Fatalf("open after main swap: %v", err)
	}
}
