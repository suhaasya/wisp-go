package session_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

type mockConnector struct {
	execLog  *[]string
	pingFail *atomic.Bool
	opens    atomic.Int32
}

func (m *mockConnector) Connect(ctx context.Context, p store.Profile, _ []byte, _ session.SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error) {
	m.opens.Add(1)
	cfg := mock.Config{
		TotalRows: 1,
		Columns:   []driver.ColumnMeta{{Name: "id", Type: "int"}},
		ExecLog:   m.execLog,
		PingFail: func() bool {
			return m.pingFail != nil && m.pingFail.Load()
		},
	}
	conn, err := mock.Driver{}.Open(ctx, driver.ConnectParams{Extra: cfg})
	if err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	if sql := session.ReadOnlySessionSQL(p.Engine); sql != "" && p.ReadOnly {
		if _, err := conn.Exec(ctx, sql); err != nil {
			_ = conn.Close()
			return nil, driver.ServerInfo{}, func() {}, err
		}
	}
	return conn, driver.ServerInfo{Engine: p.Engine, Version: "mock"}, func() {}, nil
}

func TestManagerReadOnlySetsDBSession(t *testing.T) {
	var log []string
	mc := &mockConnector{execLog: &log}
	prof := store.Profile{
		ID: "p1", Name: "x", Engine: "postgres", Env: "local",
		Transport: store.Transport{Kind: "tcp", Host: "127.0.0.1", Port: 5432},
		User:      "u", Database: "d", PasswordMode: "none", ReadOnly: true,
	}
	st := store.NewProfileStore(t.TempDir() + "/connections.json")
	created, err := st.Create(prof)
	if err != nil {
		t.Fatal(err)
	}

	mgr := session.NewManager(st, nil, nil)
	mgr.SetConnector(mc)
	s, _, err := mgr.Create(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mgr.Delete(s.ID()) }()
	found := false
	for _, line := range log {
		if line == "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY" {
			found = true
		}
	}
	if !found {
		t.Fatalf("read-only SQL not applied: %v", log)
	}
	check, err := s.Guard("INSERT INTO t VALUES (1)", false)
	if err == nil || check != session.WriteBlocked {
		t.Fatalf("guard=%v err=%v", check, err)
	}
}

func TestKeepAliveReconnects(t *testing.T) {
	var fail atomic.Bool
	mc := &mockConnector{pingFail: &fail}
	prof := store.Profile{
		ID: "p1", Name: "x", Engine: "postgres", Env: "local",
		Transport: store.Transport{Kind: "tcp", Host: "127.0.0.1", Port: 5432},
		User:      "u", Database: "d", PasswordMode: "none",
		Advanced: store.ProfileAdvanced{KeepAliveSec: 1},
	}
	st := store.NewProfileStore(t.TempDir() + "/connections.json")
	created, err := st.Create(prof)
	if err != nil {
		t.Fatal(err)
	}

	mgr := session.NewManager(st, nil, nil)
	mgr.SetConnector(mc)
	mgr.SetSleep(func(time.Duration) {})
	s, _, err := mgr.Create(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mgr.Delete(s.ID()) }()

	fail.Store(true)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if s.State() == session.StateReconnecting {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if s.State() != session.StateReconnecting {
		t.Fatalf("state=%q want reconnecting", s.State())
	}
	fail.Store(false)
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if s.State() == session.StateReady {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if s.State() != session.StateReady {
		t.Fatalf("state=%q want ready", s.State())
	}
	if mc.opens.Load() < 2 {
		t.Fatalf("opens=%d want reconnect open", mc.opens.Load())
	}
}
