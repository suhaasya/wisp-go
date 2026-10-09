package session

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

// ReviewSQLResult is returned by GET .../changes/sql.
type ReviewSQLResult struct {
	SQL                string `json:"sql"`
	Count              int    `json:"count"`
	Statements         int    `json:"statements"`
	Table              string `json:"table"`
	Schema             string `json:"schema"`
	NeedsDeleteConfirm bool   `json:"needsDeleteConfirm,omitempty"`
	Warning            string `json:"warning,omitempty"`
}

// CommitOptions controls commit confirmation.
type CommitOptions struct {
	Confirm      bool
	TableConfirm string
}

// CommitResult summarizes a successful commit.
type CommitResult struct {
	Committed int `json:"committed"`
}

// CommitError is returned when commit fails mid-batch.
type CommitError struct {
	StatementIndex int
	Message        string
	Conflict       bool
}

func (e *CommitError) Error() string { return e.Message }

// ResultReviewSQL builds display SQL for staged edits.
func (s *Session) ResultReviewSQL(ctx context.Context, rid string) (ReviewSQLResult, error) {
	rc, meta, err := s.resultMeta(rid)
	if err != nil {
		return ReviewSQLResult{}, err
	}
	if rc.cs.Count() == 0 {
		return ReviewSQLResult{}, errors.New("session: no staged changes")
	}
	plan := rc.cs.CommitPlan()
	d := s.main.Dialect()
	optimistic := hasOptimisticCells(plan)
	stmts, err := sqlgen.BuildCommitStatements(d, sqlgen.CommitSpec{
		Schema: meta.schema, Table: meta.table, Optimistic: optimistic,
	}, plan)
	if err != nil {
		return ReviewSQLResult{}, err
	}
	var parts []string
	for _, st := range stmts {
		parts = append(parts, sqlgen.DisplaySQL(d, st.SQL, st.Args))
	}
	out := ReviewSQLResult{
		SQL:        strings.Join(parts, ";\n\n") + ";\n",
		Count:      rc.cs.Count(),
		Statements: len(stmts),
		Schema:     meta.schema,
		Table:      meta.table,
	}
	if len(plan.Deletes) > 0 && needsDeleteConfirm(s.profile) {
		out.NeedsDeleteConfirm = true
	}
	if warn := s.tableCommitWarning(ctx, meta.schema, meta.table); warn != "" {
		out.Warning = warn
	}
	return out, nil
}

// CommitResultChanges runs staged SQL in one transaction.
func (s *Session) CommitResultChanges(ctx context.Context, rid string, opts CommitOptions) (CommitResult, error) {
	if s.profile.ReadOnly {
		return CommitResult{}, ErrWriteBlockedReadOnly()
	}
	rc, meta, err := s.resultMeta(rid)
	if err != nil {
		return CommitResult{}, err
	}
	if rc.cs.Count() == 0 {
		return CommitResult{}, errors.New("session: no staged changes")
	}
	plan := rc.cs.CommitPlan()
	if len(plan.Deletes) > 0 && needsDeleteConfirm(s.profile) {
		if opts.TableConfirm != meta.table {
			return CommitResult{}, errors.New("session: table name confirmation required")
		}
	}
	optimistic := hasOptimisticCells(plan)
	d := s.main.Dialect()
	stmts, err := sqlgen.BuildCommitStatements(d, sqlgen.CommitSpec{
		Schema: meta.schema, Table: meta.table, Optimistic: optimistic,
	}, plan)
	if err != nil {
		return CommitResult{}, err
	}
	for _, st := range stmts {
		check, err := s.Guard(st.SQL, opts.Confirm)
		if err != nil {
			return CommitResult{}, err
		}
		if check == WriteBlocked {
			return CommitResult{}, ErrWriteBlockedReadOnly()
		}
		if check == WriteNeedsConfirmation {
			return CommitResult{}, errors.New("session: commit needs confirmation")
		}
	}
	tx, err := s.main.Begin(ctx)
	if err != nil {
		return CommitResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for i, st := range stmts {
		res, err := tx.Exec(ctx, st.SQL, st.Args...)
		if err != nil {
			return CommitResult{}, &CommitError{StatementIndex: i, Message: err.Error()}
		}
		if optimistic && isUpdateSQL(st.SQL) && res.RowsAffected == 0 {
			rowLabel := formatRowLabel(plan, i, stmts)
			return CommitResult{}, &CommitError{
				StatementIndex: i,
				Message:        fmt.Sprintf("Row %s was changed by someone else", rowLabel),
				Conflict:       true,
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return CommitResult{}, err
	}
	n := rc.cs.Count()
	rc.cs.Discard()
	if p := s.ensurePager(); p != nil {
		p.InvalidateWindows(rid)
	}
	return CommitResult{Committed: n}, nil
}

type resultMeta struct {
	schema string
	table  string
}

func (s *Session) resultMeta(rid string) (*resultChanges, resultMeta, error) {
	rc, ok := s.resultChangesFor(rid)
	if !ok || rc == nil || rc.cs == nil {
		return nil, resultMeta{}, errors.New("session: unknown result")
	}
	if rc.schema == "" || rc.table == "" {
		return nil, resultMeta{}, errors.New("session: result is not a table browse")
	}
	return rc, resultMeta{schema: rc.schema, table: rc.table}, nil
}
