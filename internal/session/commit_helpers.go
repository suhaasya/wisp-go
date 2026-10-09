package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/sqlgen"
	"github.com/suhaasya/wisp-go/internal/store"
)

func needsDeleteConfirm(p store.Profile) bool {
	return p.Env == "production" || p.SafeMode
}

func hasOptimisticCells(plan sqlgen.CommitPlan) bool {
	for _, u := range plan.Updates {
		for _, ch := range u.Cells {
			if ch.OrigNull || ch.Original != nil {
				return true
			}
		}
	}
	return false
}

func isUpdateSQL(sql string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "UPDATE")
}

func formatRowLabel(plan sqlgen.CommitPlan, stmtIndex int, stmts []sqlgen.Stmt) string {
	// Map statement index to update row by counting prior deletes/updates.
	delN := len(plan.Deletes)
	if stmtIndex < delN {
		return formatPK(plan.Deletes[stmtIndex].KeyValues)
	}
	updIdx := stmtIndex - delN
	if updIdx < len(plan.Updates) {
		return formatPK(plan.Updates[updIdx].KeyValues)
	}
	_ = stmts
	return "?"
}

func formatPK(vals []any) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		if v == nil {
			parts[i] = "NULL"
		} else {
			parts[i] = fmt.Sprint(v)
		}
	}
	return strings.Join(parts, ", ")
}

func (s *Session) tableCommitWarning(ctx context.Context, schema, table string) string {
	if s.profile.Engine != "mysql" && s.profile.Engine != "mariadb" {
		return ""
	}
	if s.main == nil {
		return ""
	}
	d := s.main.Dialect()
	q := fmt.Sprintf(
		"SELECT ENGINE FROM information_schema.tables WHERE table_schema = %s AND table_name = %s",
		d.Placeholder(1), d.Placeholder(2),
	)
	rows, err := s.main.Query(ctx, q, schema, table)
	if err != nil {
		return ""
	}
	defer func() { _ = rows.Close() }()
	w, err := rows.NextWindow(1)
	if err != nil || w == nil || w.Rows == 0 {
		return ""
	}
	engine, ok := w.CellString(0, 0)
	if !ok {
		return ""
	}
	if strings.EqualFold(engine, "MyISAM") {
		return "This table uses MyISAM: commits cannot be rolled back if a statement fails partway through."
	}
	return ""
}
