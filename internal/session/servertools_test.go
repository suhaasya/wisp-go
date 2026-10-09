package session

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/store"
)

type metaMockConnector struct{}

func (metaMockConnector) Connect(ctx context.Context, _ store.Profile, _ []byte, _ SSHConnectSecrets) (driver.Conn, driver.ServerInfo, func(), error) {
	cfg := mock.Config{
		TotalRows: 3,
		Columns: []driver.ColumnMeta{
			{Name: "pid", Type: "text"},
			{Name: "wisp_highlight", Type: "text"},
		},
	}
	conn, err := mock.Driver{}.Open(ctx, driver.ConnectParams{Extra: cfg})
	if err != nil {
		return nil, driver.ServerInfo{}, func() {}, err
	}
	return conn, driver.ServerInfo{Engine: "postgres"}, func() {}, nil
}

func TestOpenServerToolsProcessesMetaPagerWindow(t *testing.T) {
	main, err := mock.Driver{}.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{TotalRows: 1, Columns: []driver.ColumnMeta{{Name: "x", Type: "text"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = main.Close() })

	s := &Session{
		id:        "test",
		profile:   store.Profile{Engine: "postgres"},
		state:     StateReady,
		main:      main,
		connector: metaMockConnector{},
	}
	ctx := t.Context()
	open, err := s.OpenServerToolsProcesses(ctx)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if open.RID == "" {
		t.Fatal("empty rid")
	}
	raw, err := s.ResultWindowJSON(ctx, open.RID, 0, 10)
	if err != nil {
		t.Fatalf("window: %v", err)
	}
	var win struct {
		Columns []driver.ColumnMeta `json:"columns"`
	}
	if err := json.Unmarshal(raw, &win); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(win.Columns) < 2 {
		t.Fatalf("columns=%d", len(win.Columns))
	}
}

func TestCancelServerProcessReadOnly(t *testing.T) {
	s := NewTestSession(store.Profile{Engine: "postgres", ReadOnly: true})
	err := s.CancelServerProcess(t.Context(), 42, false)
	if !IsWriteBlockedReadOnly(err) {
		t.Fatalf("err=%v", err)
	}
}
