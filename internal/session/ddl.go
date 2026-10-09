package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

// TableDDL returns CREATE TABLE DDL for the table (PostgreSQL from metadata, MySQL via SHOW CREATE).
func (s *Session) TableDDL(ctx context.Context, schema, table string, refresh bool) (string, error) {
	engine := strings.ToLower(s.Profile().Engine)
	switch engine {
	case "postgres":
		det, err := s.TableMetadata(ctx, schema, table, refresh)
		if err != nil {
			return "", err
		}
		return sqlgen.CreateTablePostgreSQL(det), nil
	case "mysql", "mariadb":
		return s.mysqlShowCreateTable(ctx, schema, table)
	default:
		return "", errBadEngine
	}
}

func (s *Session) mysqlShowCreateTable(ctx context.Context, schema, table string) (string, error) {
	var ddl string
	err := s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		d := conn.Dialect()
		q := fmt.Sprintf("SHOW CREATE TABLE %s.%s",
			d.QuoteIdent(schema), d.QuoteIdent(table))
		var err error
		ddl, err = querySecondTextColumn(ctx, conn, q)
		return err
	})
	return ddl, err
}

func querySecondTextColumn(ctx context.Context, conn driver.Conn, sql string, args ...any) (string, error) {
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	w, err := rows.NextWindow(1)
	if err != nil {
		return "", err
	}
	if w == nil || w.Rows == 0 {
		return "", fmt.Errorf("session: empty result for %q", sql)
	}
	if len(w.Columns) < 2 {
		s, ok := w.CellText(0, 0)
		if !ok {
			return "", fmt.Errorf("session: missing ddl cell")
		}
		return s, nil
	}
	s, ok := w.CellText(1, 0)
	if !ok {
		return "", fmt.Errorf("session: missing ddl cell")
	}
	return s, nil
}
