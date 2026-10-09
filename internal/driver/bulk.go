package driver

import "context"

// CopyFromSource streams rows for COPY / bulk load.
type CopyFromSource interface {
	Next() ([]any, error)
}

// CopyFromConn loads rows through a server-specific fast path (PostgreSQL COPY FROM).
type CopyFromConn interface {
	CopyFrom(ctx context.Context, schema, table string, columns []string, src CopyFromSource) (int64, error)
}
