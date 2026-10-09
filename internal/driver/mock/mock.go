// Package mock provides a deterministic driver for tests and UI development.
package mock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// Driver is a deterministic row generator for tests and UI development.
type Driver struct{}

func (Driver) Name() string { return "mock" }

// Config selects the shape of mock results.
type Config struct {
	TotalRows int
	Columns   []driver.ColumnMeta
	ExecLog   *[]string
	PingFail  func() bool
}

// Open returns a connection that serves Config rows.
func (Driver) Open(ctx context.Context, p driver.ConnectParams) (driver.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg, ok := p.Extra.(Config)
	if !ok {
		return nil, errors.New("mock: ConnectParams.Extra must be mock.Config")
	}
	if cfg.TotalRows < 0 || len(cfg.Columns) == 0 {
		return nil, errors.New("mock: invalid Config")
	}
	return &conn{cfg: cfg}, nil
}

type conn struct {
	cfg Config
}

func (c *conn) Ping(ctx context.Context) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if c.cfg.PingFail != nil && c.cfg.PingFail() {
		return 0, errors.New("mock: ping failed")
	}
	return time.Millisecond, nil
}

func (c *conn) ServerInfo(ctx context.Context) (driver.ServerInfo, error) {
	if err := ctx.Err(); err != nil {
		return driver.ServerInfo{}, err
	}
	return driver.ServerInfo{Version: "mock-1.0", Engine: "mock"}, nil
}

func (c *conn) Exec(ctx context.Context, sql string, args ...any) (driver.ExecResult, error) {
	if err := ctx.Err(); err != nil {
		return driver.ExecResult{}, err
	}
	if c.cfg.ExecLog != nil {
		*c.cfg.ExecLog = append(*c.cfg.ExecLog, sql)
	}
	return driver.ExecResult{}, nil
}

func (c *conn) Query(ctx context.Context, sql string, args ...any) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &rows{
		cfg:    c.cfg,
		cursor: 0,
	}, nil
}

func (c *conn) Begin(ctx context.Context) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New("mock: Begin not implemented")
}

func (c *conn) Dialect() driver.Dialect { return dialect{} }

func (c *conn) Close() error { return nil }

type rows struct {
	cfg    Config
	cursor int
}

func (r *rows) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(r.cfg.Columns))
	copy(out, r.cfg.Columns)
	return out
}

func (r *rows) NextWindow(max int) (*driver.Window, error) {
	if max <= 0 {
		max = 150
	}
	if r.cursor >= r.cfg.TotalRows {
		return nil, nil
	}
	n := max
	if remain := r.cfg.TotalRows - r.cursor; remain < n {
		n = remain
	}
	w := driver.NewWindow(r.cfg.Columns, n)
	for ci, col := range r.cfg.Columns {
		b := w.Column(ci)
		for row := 0; row < n; row++ {
			globalRow := r.cursor + row
			if err := fillCell(b, col, row, globalRow, ci); err != nil {
				return nil, err
			}
		}
	}
	r.cursor += n
	return w, nil
}

func fillCell(b *driver.ColumnBuilder, col driver.ColumnMeta, localRow, globalRow, colIdx int) error {
	switch columnKindForType(col.Type) {
	case colKindInt:
		b.SetInt(localRow, int64(globalRow)*1000+int64(colIdx))
	case colKindText:
		b.SetText(localRow, fmt.Sprintf("r%dc%d-%020d", globalRow, colIdx, globalRow))
	default:
		b.SetText(localRow, fmt.Sprintf("v%d", globalRow))
	}
	return nil
}

func (r *rows) Close() error { return nil }

type colKind int

const (
	colKindInt colKind = iota
	colKindText
)

func columnKindForType(typ string) colKind {
	switch typ {
	case "int2", "int4", "int8", "integer", "bigint":
		return colKindInt
	default:
		return colKindText
	}
}

type dialect struct{}

func (dialect) Name() string { return "mock" }

func (dialect) QuoteIdent(name string) string { return `"` + name + `"` }

func (dialect) QuoteLiteral(s string) string { return `'` + s + `'` }

func (dialect) Placeholder(n int) string { return fmt.Sprintf("$%d", n) }

func (dialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (dialect) KeysetPredicate(columns []string, values []any, dir string) (string, []any, error) {
	return "", nil, errors.New("mock: KeysetPredicate not implemented")
}
