// Package mysql implements the Wisp driver for MySQL and MariaDB.
package mysql

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql" // register "mysql" driver
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver"
)

var mysqlTLSReg atomic.Uint64
var mysqlNetReg atomic.Uint64

// Driver connects to MySQL/MariaDB over TCP.
type Driver struct{}

func (Driver) Name() string { return "mysql" }

// Open returns a session-scoped connection (one underlying *sql.Conn).
func (Driver) Open(ctx context.Context, p driver.ConnectParams) (driver.Conn, error) {
	mode := strings.ToLower(strings.TrimSpace(p.TLS.Mode))
	if mode == "prefer" {
		conn, err := openMySQL(ctx, p, true)
		if err == nil {
			return conn, nil
		}
		return openMySQL(ctx, p, false)
	}
	useTLS := mode != "" && mode != "disable"
	return openMySQL(ctx, p, useTLS)
}

func openMySQL(ctx context.Context, p driver.ConnectParams, useTLS bool) (driver.Conn, error) {
	dsn, err := buildDSN(p, useTLS)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	sconn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	c := &sqlConn{
		db:      db,
		conn:    sconn,
		dsn:     dsn,
		cancel:  cancelRegistry{dsn: dsn},
		dialect: myDialect{engine: "mysql"},
	}
	if err := c.loadConnectionID(ctx); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := c.detectEngine(ctx); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func buildDSN(p driver.ConnectParams, useTLS bool) (string, error) {
	if p.Host == "" {
		return "", errors.New("mysql: host required")
	}
	if p.Port == 0 {
		p.Port = 3306
	}
	cfg := mysqldriver.Config{
		User:                 p.User,
		Passwd:               string(p.Password),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", p.Host, p.Port),
		DBName:               p.DB,
		ParseTime:            false,
		InterpolateParams:    false,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}
	if p.AllowPublicKeyRetrieval {
		cfg.Params["allowPublicKeyRetrieval"] = "true"
	}
	if p.ConnectTimeout > 0 {
		cfg.Timeout = p.ConnectTimeout
		cfg.ReadTimeout = p.ConnectTimeout
		cfg.WriteTimeout = p.ConnectTimeout
	}
	if useTLS {
		name, err := registerMySQLTLS(p.TLS)
		if err != nil {
			return "", err
		}
		cfg.TLSConfig = name
	}
	if p.DialContext != nil {
		name, err := registerMySQLNet(p.DialContext)
		if err != nil {
			return "", err
		}
		cfg.Net = name
	}
	return cfg.FormatDSN(), nil
}

func registerMySQLNet(dial func(context.Context, string, string) (net.Conn, error)) (string, error) {
	name := fmt.Sprintf("wisp-net-%d", mysqlNetReg.Add(1))
	mysqldriver.RegisterDialContext(name, func(ctx context.Context, addr string) (net.Conn, error) {
		return dial(ctx, "tcp", addr)
	})
	return name, nil
}

func registerMySQLTLS(p driver.TLSParams) (string, error) {
	tlsCfg, err := driver.ClientTLSConfig(p)
	if err != nil {
		return "", err
	}
	if tlsCfg == nil {
		return "", nil
	}
	name := fmt.Sprintf("wisp-%d", mysqlTLSReg.Add(1))
	if err := mysqldriver.RegisterTLSConfig(name, tlsCfg); err != nil {
		return "", err
	}
	return name, nil
}

type sqlConn struct {
	db      *sql.DB
	conn    *sql.Conn
	dsn     string
	cancel  cancelRegistry
	dialect myDialect
}

func (c *sqlConn) loadConnectionID(ctx context.Context) error {
	var id int64
	if err := c.conn.QueryRowContext(ctx, "SELECT CONNECTION_ID()").Scan(&id); err != nil {
		return err
	}
	c.cancel.set(id)
	return nil
}

func (c *sqlConn) detectEngine(ctx context.Context) error {
	var version string
	if err := c.conn.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(version), "mariadb") {
		c.dialect.engine = "mariadb"
	}
	return nil
}

func (c *sqlConn) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	if err := c.conn.PingContext(ctx); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func (c *sqlConn) ServerInfo(ctx context.Context) (driver.ServerInfo, error) {
	var version, db, user string
	err := c.conn.QueryRowContext(ctx, "SELECT VERSION(), DATABASE(), USER()").Scan(&version, &db, &user)
	if err != nil {
		return driver.ServerInfo{}, err
	}
	if idx := strings.Index(user, "@"); idx >= 0 {
		user = user[:idx]
	}
	return driver.ServerInfo{
		Version:  version,
		Engine:   c.dialect.engine,
		Database: db,
		User:     user,
		TLS:      c.tlsStatus(ctx),
	}, nil
}

func (c *sqlConn) tlsStatus(ctx context.Context) string {
	var out tls.ConnectionState
	var ok bool
	_ = c.conn.Raw(func(dc any) error {
		out, ok = mysqlTLSState(dc)
		return nil
	})
	if !ok || !out.HandshakeComplete {
		return c.tlsFromShow(ctx)
	}
	return driver.FormatTLSLine(&out)
}

func mysqlTLSState(dc any) (tls.ConnectionState, bool) {
	v := reflect.ValueOf(dc)
	if !v.IsValid() {
		return tls.ConnectionState{}, false
	}
	m := v.MethodByName("TLSConnectionState")
	if !m.IsValid() {
		return tls.ConnectionState{}, false
	}
	res := m.Call(nil)
	if len(res) != 2 {
		return tls.ConnectionState{}, false
	}
	st, _ := res[0].Interface().(tls.ConnectionState)
	ok, _ := res[1].Interface().(bool)
	return st, ok
}

func (c *sqlConn) tlsFromShow(ctx context.Context) string {
	rows, err := c.conn.QueryContext(ctx, "SHOW SESSION STATUS WHERE Variable_name IN ('Ssl_version','Ssl_cipher')")
	if err != nil {
		return ""
	}
	defer func() { _ = rows.Close() }()
	var version, cipher string
	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err != nil {
			return ""
		}
		switch name {
		case "Ssl_version":
			version = val
		case "Ssl_cipher":
			cipher = val
		}
	}
	if version == "" {
		return ""
	}
	if cipher == "" {
		return version
	}
	return version + ", " + cipher
}

func (c *sqlConn) Exec(ctx context.Context, sqlText string, args ...any) (driver.ExecResult, error) {
	res, err := c.conn.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return driver.ExecResult{}, err
	}
	n, _ := res.RowsAffected()
	return driver.ExecResult{RowsAffected: n}, nil
}

func (c *sqlConn) Query(ctx context.Context, sqlText string, args ...any) (driver.Rows, error) {
	qctx, stop := watchCancel(ctx, &c.cancel)
	rows, err := c.conn.QueryContext(qctx, sqlText, args...)
	if err != nil {
		stop()
		return nil, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		stop()
		_ = rows.Close()
		return nil, err
	}
	return &sqlRows{
		rows:    rows,
		types:   types,
		columns: columnsFromTypes(types),
		stop:    stop,
	}, nil
}

func (c *sqlConn) Begin(ctx context.Context) (driver.Tx, error) {
	tx, err := c.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx: tx}, nil
}

func (c *sqlConn) Dialect() driver.Dialect { return c.dialect }

func (c *sqlConn) Close() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

type sqlTx struct {
	tx *sql.Tx
}

func (t *sqlTx) Exec(ctx context.Context, sql string, args ...any) (driver.ExecResult, error) {
	res, err := t.tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return driver.ExecResult{}, err
	}
	n, _ := res.RowsAffected()
	return driver.ExecResult{RowsAffected: n}, nil
}

func (t *sqlTx) Commit(ctx context.Context) error   { return t.tx.Commit() }
func (t *sqlTx) Rollback(ctx context.Context) error { return t.tx.Rollback() }

type sqlRows struct {
	rows    *sql.Rows
	types   []*sql.ColumnType
	columns []driver.ColumnMeta
	stop    func()
	closed  bool
}

func (r *sqlRows) Columns() []driver.ColumnMeta {
	out := make([]driver.ColumnMeta, len(r.columns))
	copy(out, r.columns)
	return out
}

func (r *sqlRows) NextWindow(max int) (*driver.Window, error) {
	if r.closed || r.rows == nil {
		return nil, nil
	}
	if max <= 0 {
		max = 150
	}
	w := driver.NewWindow(r.columns, max)
	n := 0
	dest := make([]sql.RawBytes, len(r.types))
	scanArgs := make([]any, len(r.types))
	for i := range dest {
		scanArgs[i] = &dest[i]
	}
	for n < max {
		if !r.rows.Next() {
			if err := r.rows.Err(); err != nil {
				return nil, err
			}
			break
		}
		if err := r.rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		for i, ct := range r.types {
			raw := dest[i]
			if raw != nil {
				cp := make([]byte, len(raw))
				copy(cp, raw)
				raw = cp
			}
			decodeCell(w.Column(i), ct, raw, n)
		}
		n++
	}
	if n == 0 {
		return nil, nil
	}
	w.Rows = n
	return w, nil
}

func (r *sqlRows) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.stop != nil {
		r.stop()
	}
	if r.rows != nil {
		return r.rows.Close()
	}
	return nil
}

type myDialect struct {
	engine string
}

// Dialect returns the MySQL/MariaDB SQL dialect (for sqlgen and tests).
func Dialect(engine string) driver.Dialect {
	if engine == "" {
		engine = "mysql"
	}
	return myDialect{engine: engine}
}

func (d myDialect) Name() string { return d.engine }

func (d myDialect) QuoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func (d myDialect) QuoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

func (d myDialect) Placeholder(n int) string {
	return "?"
}

func (d myDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (d myDialect) KeysetPredicate(columns []string, values []any, dir string) (string, []any, error) {
	return "", nil, errors.New("mysql: KeysetPredicate not implemented")
}
