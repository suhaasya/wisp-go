package mysql

import (
	"context"
	"fmt"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// ProcessListSQL returns SHOW FULL PROCESSLIST as a selectable query with wisp_highlight.
func ProcessListSQL() string {
	return `
SELECT
  CAST(id AS CHAR) AS id,
  user,
  host,
  db,
  command,
  time AS seconds,
  state,
  LEFT(info, 200) AS info,
  CASE
    WHEN state LIKE '%lock%' OR state LIKE 'Waiting%' THEN 'blocked'
    WHEN time > 30 THEN 'long_running'
    ELSE ''
  END AS wisp_highlight
FROM information_schema.processlist
WHERE id <> CONNECTION_ID()
ORDER BY id`
}

// CancelProcess sends KILL QUERY for a connection id.
func CancelProcess(ctx context.Context, conn driver.Conn, id int64) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("KILL QUERY %d", id))
	return err
}

// TerminateProcess sends KILL for a connection id.
func TerminateProcess(ctx context.Context, conn driver.Conn, id int64) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("KILL %d", id))
	return err
}

// VariablesSQL lists system variables.
func VariablesSQL() string {
	return `SHOW VARIABLES`
}

// UsersSQL lists users and hosts (read-only).
func UsersSQL() string {
	return `SELECT user, host FROM mysql.user ORDER BY user, host`
}

// FormatCancelSQL returns SQL for guard checks.
func FormatCancelSQL(id int64, terminate bool) string {
	if terminate {
		return fmt.Sprintf("KILL %d", id)
	}
	return fmt.Sprintf("KILL QUERY %d", id)
}
