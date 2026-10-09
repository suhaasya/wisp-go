package session

import (
	"context"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/pager"
)

// ServerToolsVariable is one server variable row.
type ServerToolsVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ServerToolsUser is one user/grant summary row (read-only).
type ServerToolsUser struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

func processListSQL(engine string) string {
	switch engine {
	case "postgres":
		return postgres.ProcessListSQL()
	case "mysql", "mariadb":
		return mysql.ProcessListSQL()
	default:
		return postgres.ProcessListSQL()
	}
}

func cancelProcessSQL(engine string, pid int64, terminate bool) string {
	switch engine {
	case "postgres":
		return postgres.FormatCancelSQL(pid, terminate)
	case "mysql", "mariadb":
		return mysql.FormatCancelSQL(pid, terminate)
	default:
		return postgres.FormatCancelSQL(pid, terminate)
	}
}

// OpenServerToolsProcesses opens or refreshes the process list on the metadata connection.
func (s *Session) OpenServerToolsProcesses(ctx context.Context) (pager.OpenResult, error) {
	sqlText := processListSQL(s.profile.Engine)
	if _, err := s.Guard(sqlText, false); err != nil {
		return pager.OpenResult{}, err
	}
	mp, err := s.ensureMetaPager(ctx)
	if err != nil {
		return pager.OpenResult{}, err
	}
	s.mu.Lock()
	if old := s.serverToolsProcessRID; old != "" {
		mp.Close(old)
	}
	s.mu.Unlock()
	open, err := mp.Open(ctx, pager.Source{Kind: pager.SourceQuery, SQL: sqlText})
	if err != nil {
		return pager.OpenResult{}, err
	}
	s.mu.Lock()
	s.serverToolsProcessRID = open.RID
	s.mu.Unlock()
	return open, nil
}

// CancelServerProcess cancels or terminates a server backend (metadata connection).
func (s *Session) CancelServerProcess(ctx context.Context, pid int64, terminate bool) error {
	if s.Profile().ReadOnly {
		return errWriteBlockedReadOnly
	}
	sqlText := cancelProcessSQL(s.profile.Engine, pid, terminate)
	check, err := s.Guard(sqlText, false)
	if err != nil {
		return err
	}
	if check == WriteBlocked {
		return errWriteBlockedReadOnly
	}
	return s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		switch s.profile.Engine {
		case "postgres":
			if terminate {
				return postgres.TerminateBackend(ctx, conn, pid)
			}
			return postgres.CancelBackend(ctx, conn, pid)
		case "mysql", "mariadb":
			if terminate {
				return mysql.TerminateProcess(ctx, conn, pid)
			}
			return mysql.CancelProcess(ctx, conn, pid)
		default:
			return postgres.TerminateBackend(ctx, conn, pid)
		}
	})
}

func variablesSQL(engine string) string {
	switch engine {
	case "postgres":
		return postgres.VariablesSQL()
	case "mysql", "mariadb":
		return mysql.VariablesSQL()
	default:
		return postgres.VariablesSQL()
	}
}

func usersSQL(engine string) string {
	switch engine {
	case "postgres":
		return postgres.UsersSQL()
	case "mysql", "mariadb":
		return mysql.UsersSQL()
	default:
		return postgres.UsersSQL()
	}
}

// ServerToolsVariables lists variables from the metadata connection, filtered by search.
func (s *Session) ServerToolsVariables(ctx context.Context, search string) ([]ServerToolsVariable, error) {
	sqlText := variablesSQL(s.profile.Engine)
	var out []ServerToolsVariable
	err := s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		rows, err := conn.Query(ctx, sqlText)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for {
			w, err := rows.NextWindow(200)
			if err != nil {
				return err
			}
			if w == nil || w.Rows == 0 {
				break
			}
			nameCol, valCol := 0, 1
			if s.profile.Engine == "mysql" || s.profile.Engine == "mariadb" {
				// SHOW VARIABLES: Variable_name, Value
				for i, c := range w.Columns {
					if c.Name == "Variable_name" {
						nameCol = i
					}
					if c.Name == "Value" {
						valCol = i
					}
				}
			}
			for r := 0; r < w.Rows; r++ {
				name := strings.ToLower(cellText(w, r, nameCol))
				val := cellText(w, r, valCol)
				if search != "" && !strings.Contains(name, strings.ToLower(search)) {
					continue
				}
				out = append(out, ServerToolsVariable{
					Name:  cellText(w, r, nameCol),
					Value: val,
				})
			}
		}
		return nil
	})
	return out, err
}

// ServerToolsUsers returns user/grant rows (read-only).
func (s *Session) ServerToolsUsers(ctx context.Context) (ServerToolsUser, error) {
	sqlText := usersSQL(s.profile.Engine)
	var out ServerToolsUser
	err := s.withMetaConn(ctx, func(ctx context.Context, conn driver.Conn) error {
		rows, err := conn.Query(ctx, sqlText)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		w, err := rows.NextWindow(500)
		if err != nil {
			return err
		}
		if w == nil {
			return nil
		}
		for _, c := range w.Columns {
			out.Columns = append(out.Columns, c.Name)
		}
		for r := 0; r < w.Rows; r++ {
			var row []string
			for c := 0; c < len(w.Columns); c++ {
				row = append(row, cellText(w, r, c))
			}
			out.Rows = append(out.Rows, row)
		}
		return nil
	})
	return out, err
}
