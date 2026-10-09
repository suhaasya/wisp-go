package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type cancelRegistry struct {
	mu  sync.Mutex
	dsn string
	id  int64
}

func (r *cancelRegistry) set(id int64) {
	r.mu.Lock()
	r.id = id
	r.mu.Unlock()
}

func (r *cancelRegistry) killQuery() {
	r.mu.Lock()
	id := r.id
	dsn := r.dsn
	r.mu.Unlock()
	if id == 0 || dsn == "" {
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return
	}
	db.SetMaxOpenConns(1)
	defer func() { _ = db.Close() }()
	kctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = db.ExecContext(kctx, fmt.Sprintf("KILL QUERY %d", id))
}

func watchCancel(ctx context.Context, reg *cancelRegistry) (context.Context, func()) {
	if ctx == nil {
		return context.Background(), func() {}
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				reg.killQuery()
			}
		case <-done:
		}
	}()
	return ctx, func() { close(done) }
}
