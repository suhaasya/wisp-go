package session

import (
	"context"
	"errors"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

type ctxCheckConnector struct {
	connect func(ctx context.Context) error
}

func (c ctxCheckConnector) openMetaConn() (driver.Conn, error) {
	var d mock.Driver
	return d.Open(context.Background(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: 1,
			Columns:   []driver.ColumnMeta{{Name: "x", Type: "text"}},
		},
	})
}

func (c ctxCheckConnector) Connect(ctx context.Context, _ store.Profile, _ []byte, _ SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error) {
	if err := c.connect(ctx); err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	conn, err := c.openMetaConn()
	if err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	return conn, driver.ServerInfo{Engine: "postgres"}, func() {}, nil
}

func TestMetadataConnNotOpenedWithCanceledContext(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	main, err := ctxCheckConnector{}.openMetaConn()
	if err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	s.main = main
	s.connector = ctxCheckConnector{
		connect: func(ctx context.Context) error {
			if ctx.Err() != nil {
				return errors.New("connect used canceled context")
			}
			return nil
		},
	}
	s.mu.Unlock()

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	if err := s.withMetaConn(canceled, func(_ context.Context, conn driver.Conn) error {
		if conn == nil {
			t.Fatal("expected meta conn")
		}
		return nil
	}); err != nil {
		t.Fatalf("withMetaConn: %v", err)
	}
}

func TestMetadataConnRetriesAfterFailedOpen(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres"})
	main, err := ctxCheckConnector{}.openMetaConn()
	if err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	s.main = main
	attempts := 0
	s.connector = ctxCheckConnector{
		connect: func(ctx context.Context) error {
			attempts++
			if attempts == 1 {
				return errors.New("temporary")
			}
			return nil
		},
	}
	s.mu.Unlock()

	if err := s.withMetaConn(context.Background(), func(context.Context, driver.Conn) error { return nil }); err == nil {
		t.Fatal("expected first failure")
	}
	if err := s.withMetaConn(context.Background(), func(context.Context, driver.Conn) error { return nil }); err != nil {
		t.Fatalf("second open: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}
