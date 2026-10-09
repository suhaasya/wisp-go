package driver

import (
	"context"
	"net"
	"time"

	"github.com/suhaasya/wisp-go/internal/transport"
)

// Driver opens database connections for one engine.
type Driver interface {
	Open(ctx context.Context, p ConnectParams) (Conn, error)
	Name() string
}

// Conn is a live database connection.
type Conn interface {
	Ping(ctx context.Context) (time.Duration, error)
	ServerInfo(ctx context.Context) (ServerInfo, error)
	Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	Begin(ctx context.Context) (Tx, error)
	Dialect() Dialect
	Close() error
}

// Rows streams result windows.
type Rows interface {
	Columns() []ColumnMeta
	NextWindow(max int) (*Window, error)
	Close() error
}

// Tx is an open transaction.
type Tx interface {
	Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ConnectParams holds non-secret connection options (secrets resolved elsewhere).
type ConnectParams struct {
	Host  string
	Port  int
	DB    string
	User  string
	Extra any // engine-specific; mock.Config for mock driver

	// TransportKind is tcp, socket, or pipe (from profile transport.kind).
	TransportKind string
	// TransportPath is the unix socket file or named pipe identifier.
	TransportPath string

	Password         []byte
	PgPassFile       string
	ConnectTimeout   time.Duration
	StatementTimeout time.Duration

	// MySQL/MariaDB (LW-014)
	AllowPublicKeyRetrieval bool

	TLS TLSParams

	DialContext func(ctx context.Context, network, address string) (net.Conn, error)
	ViaSSH      bool
}

// TLSParams mirrors transport TLS settings on ConnectParams.
type TLSParams = transport.TLSParams

// ServerInfo describes the connected server.
type ServerInfo struct {
	Version  string
	Engine   string
	Database string
	User     string
	TLS      string
	ViaSSH   bool
}

// ExecResult summarizes a non-query statement.
type ExecResult struct {
	RowsAffected int64
}

// ColumnMeta describes one result column.
type ColumnMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
