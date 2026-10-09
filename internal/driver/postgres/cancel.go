package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

func watchCancel(ctx context.Context, pg *pgconn.PgConn) (context.Context, func()) {
	if ctx == nil {
		return context.Background(), func() {}
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				_ = pg.CancelRequest(ctx) //nolint:gosec // G118: cancel RPC after client ctx canceled
			}
		case <-done:
		}
	}()
	return ctx, func() { close(done) }
}
