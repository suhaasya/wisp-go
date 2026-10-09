// Package postgres implements the Wisp driver for PostgreSQL via pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/suhaasya/wisp-go/internal/driver"
)

// Driver connects to PostgreSQL over TCP with pgx.
type Driver struct{}

func (Driver) Name() string { return "postgres" }

// Open dials PostgreSQL using driver.ConnectParams.
func (Driver) Open(ctx context.Context, p driver.ConnectParams) (driver.Conn, error) {
	cfg, err := connConfig(p)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pg, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &pgConn{pg: pg, dialect: pgDialect{}}, nil
}

func connConfig(p driver.ConnectParams) (*pgx.ConnConfig, error) {
	if p.Host == "" {
		return nil, errors.New("postgres: host required")
	}
	if p.Port == 0 {
		p.Port = 5432
	}
	user := url.UserPassword(p.User, string(p.Password))
	userURL := user.String()
	if len(p.Password) == 0 {
		userURL = url.User(p.User).String()
	}
	connURL := fmt.Sprintf("postgres://%s@%s:%d/%s", userURL, p.Host, p.Port, url.PathEscape(p.DB))
	cfg, err := pgx.ParseConfig(connURL)
	if err != nil {
		return nil, err
	}
	if cfg.RuntimeParams == nil {
		cfg.RuntimeParams = map[string]string{}
	}
	cfg.RuntimeParams["application_name"] = "wisp"
	if p.StatementTimeout > 0 {
		ms := p.StatementTimeout.Milliseconds()
		if ms < 1 {
			ms = 1
		}
		cfg.RuntimeParams["statement_timeout"] = fmt.Sprintf("%d", ms)
	}
	if p.ConnectTimeout > 0 {
		cfg.ConnectTimeout = p.ConnectTimeout
	}
	if p.PgPassFile != "" && len(p.Password) == 0 {
		cfg.RuntimeParams["passfile"] = p.PgPassFile
	}
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	return cfg, nil
}

type pgConn struct {
	pg      *pgx.Conn
	dialect pgDialect
}

func (c *pgConn) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	if err := c.pg.Ping(ctx); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func (c *pgConn) ServerInfo(ctx context.Context) (driver.ServerInfo, error) {
	var version, db, user string
	if err := c.pg.QueryRow(ctx, "SELECT version(), current_database(), current_user").Scan(&version, &db, &user); err != nil {
		return driver.ServerInfo{}, err
	}
	return driver.ServerInfo{
		Version:  version,
		Engine:   "postgres",
		Database: db,
		User:     user,
	}, nil
}

func (c *pgConn) Exec(ctx context.Context, sql string, args ...any) (driver.ExecResult, error) {
	tag, err := c.pg.Exec(ctx, sql, args...)
	if err != nil {
		return driver.ExecResult{}, err
	}
	return driver.ExecResult{RowsAffected: tag.RowsAffected()}, nil
}

func (c *pgConn) Query(ctx context.Context, sql string, args ...any) (driver.Rows, error) {
	qctx, stop := watchCancel(ctx, c.pg.PgConn())
	defer stop()
	rows, err := c.pg.Query(qctx, sql, args...)
	if err != nil {
		return nil, err
	}
	fields := rows.FieldDescriptions()
	return &pgRows{
		rows:    rows,
		columns: columnsFromFields(fields),
		fields:  fields,
	}, nil
}

func (c *pgConn) Begin(ctx context.Context) (driver.Tx, error) {
	tx, err := c.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgTx{tx: tx}, nil
}

func (c *pgConn) Dialect() driver.Dialect { return c.dialect }

func (c *pgConn) Close() error {
	if c.pg != nil {
		_ = c.pg.Close(context.Background())
	}
	return nil
}

type pgTx struct {
	tx pgx.Tx
}

func (t *pgTx) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t *pgTx) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }

type pgRows struct {
	rows    pgx.Rows
	columns []driver.ColumnMeta
	fields  []pgconn.FieldDescription
}

func (r *pgRows) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(r.columns))
	copy(out, r.columns)
	return out
}

func (r *pgRows) NextWindow(max int) (*driver.Window, error) {
	if max <= 0 {
		max = 150
	}
	if r.rows == nil {
		return nil, nil
	}
	w := driver.NewWindow(r.columns, max)
	n := 0
	for n < max {
		if !r.rows.Next() {
			err := r.rows.Err()
			if n == 0 {
				return nil, err
			}
			w.Rows = n
			return w, err
		}
		raw := r.rows.RawValues()
		for i, fd := range r.fields {
			decodeIntoColumn(w.Column(i), fd.DataTypeOID, raw[i], n)
		}
		n++
	}
	w.Rows = n
	return w, r.rows.Err()
}

func (r *pgRows) Close() error {
	if r.rows != nil {
		r.rows.Close()
	}
	return nil
}

type pgDialect struct{}

func (pgDialect) Name() string { return "postgres" }

func (pgDialect) QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (pgDialect) QuoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

func (pgDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (pgDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (pgDialect) KeysetPredicate(columns []string, values []any, dir string) (string, []any, error) {
	return "", nil, errors.New("postgres: KeysetPredicate not implemented")
}
