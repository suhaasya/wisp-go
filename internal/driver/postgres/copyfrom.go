package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/suhaasya/wisp-go/internal/driver"
)

func (c *pgConn) CopyFrom(ctx context.Context, schema, table string, columns []string, src driver.CopyFromSource) (int64, error) {
	return c.pg.CopyFrom(ctx, pgx.Identifier{schema, table}, columns, pgx.CopyFromFunc(func() ([]any, error) {
		return src.Next()
	}))
}
