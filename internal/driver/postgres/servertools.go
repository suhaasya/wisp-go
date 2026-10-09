package postgres

import (
	"context"
	"fmt"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// ProcessListSQL returns pg_stat_activity with wisp_highlight for the process grid.
func ProcessListSQL() string {
	return `
SELECT
  pid::text AS pid,
  usename,
  COALESCE(state, '') AS state,
  COALESCE(wait_event_type, '') AS wait_event_type,
  COALESCE(left(query, 200), '') AS query,
  GREATEST(0, EXTRACT(EPOCH FROM (now() - query_start))::int) AS seconds,
  CASE
    WHEN wait_event_type = 'Lock' THEN 'blocked'
    WHEN EXTRACT(EPOCH FROM (now() - query_start)) > interval '30 seconds' THEN 'long_running'
    ELSE ''
  END AS wisp_highlight
FROM pg_stat_activity
WHERE datname = current_database()
  AND pid <> pg_backend_pid()
ORDER BY pid`
}

// CancelBackend sends pg_cancel_backend.
func CancelBackend(ctx context.Context, conn driver.Conn, pid int64) error {
	_, err := conn.Exec(ctx, "SELECT pg_cancel_backend($1)", pid)
	return err
}

// TerminateBackend sends pg_terminate_backend.
func TerminateBackend(ctx context.Context, conn driver.Conn, pid int64) error {
	_, err := conn.Exec(ctx, "SELECT pg_terminate_backend($1)", pid)
	return err
}

// VariablesSQL lists server settings for the current database cluster.
func VariablesSQL() string {
	return `SELECT name, setting FROM pg_settings ORDER BY name`
}

// UsersSQL lists roles and membership (read-only).
func UsersSQL() string {
	return `
SELECT rolname AS user_name,
       rolsuper::text AS superuser,
       rolcreatedb::text AS create_db,
       rolcanlogin::text AS can_login
FROM pg_roles
ORDER BY rolname`
}

// FormatCancelSQL returns guarded cancel/terminate SQL for the write guard.
func FormatCancelSQL(pid int64, terminate bool) string {
	if terminate {
		return fmt.Sprintf("SELECT pg_terminate_backend(%d)", pid)
	}
	return fmt.Sprintf("SELECT pg_cancel_backend(%d)", pid)
}
