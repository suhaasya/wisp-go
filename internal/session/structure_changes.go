package session

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
	"github.com/suhaasya/wisp-go/internal/store"
)

type structureChangeSet struct {
	schema string
	table  string
	ops    []sqlgen.StructureOp
}

func (s *Session) ensureStructureChanges(schema, table string) *structureChangeSet {
	s.changeMu.Lock()
	defer s.changeMu.Unlock()
	if s.structureChanges == nil {
		s.structureChanges = make(map[string]*structureChangeSet)
	}
	key := structureKey(schema, table)
	if cs, ok := s.structureChanges[key]; ok {
		return cs
	}
	cs := &structureChangeSet{schema: schema, table: table}
	s.structureChanges[key] = cs
	return cs
}

func structureKey(schema, table string) string {
	return schema + "\x00" + table
}

// ApplyStructureChange stages one structure edit.
func (s *Session) ApplyStructureChange(ctx context.Context, schema, table string, op sqlgen.StructureOp) (int, error) {
	if s.profile.ReadOnly {
		return 0, ErrWriteBlockedReadOnly()
	}
	if schema == "" || table == "" {
		return 0, errors.New("session: schema and table required")
	}
	cs := s.ensureStructureChanges(schema, table)
	cs.ops = append(cs.ops, op)
	return len(cs.ops), nil
}

// DiscardStructureChanges clears staged structure edits.
func (s *Session) DiscardStructureChanges(schema, table string) error {
	s.changeMu.Lock()
	defer s.changeMu.Unlock()
	delete(s.structureChanges, structureKey(schema, table))
	return nil
}

// StructureChangeCount returns staged structure op count.
func (s *Session) StructureChangeCount(schema, table string) int {
	s.changeMu.RLock()
	cs := s.structureChanges[structureKey(schema, table)]
	s.changeMu.RUnlock()
	if cs == nil {
		return 0
	}
	return len(cs.ops)
}

// StructureReviewSQL builds display SQL for staged structure edits.
func (s *Session) StructureReviewSQL(ctx context.Context, schema, table string) (ReviewSQLResult, error) {
	cs := s.getStructureChanges(schema, table)
	if cs == nil || len(cs.ops) == 0 {
		return ReviewSQLResult{Schema: schema, Table: table}, nil
	}
	det, err := s.TableMetadata(ctx, schema, table, false)
	if err != nil {
		return ReviewSQLResult{}, err
	}
	if s.main == nil {
		return ReviewSQLResult{}, errors.New("session: not connected")
	}
	d := s.main.Dialect()
	spec := structureSpec(s, schema, table)
	batch, err := sqlgen.BuildStructureAlterStatements(d, spec, det, sqlgen.StructureAlterPlan{Ops: cs.ops})
	if err != nil {
		return ReviewSQLResult{}, err
	}
	parts, stmts := displayStructureBatch(d, batch)
	out := ReviewSQLResult{
		SQL:        strings.Join(parts, ";\n\n") + ";\n",
		Count:      len(cs.ops),
		Statements: stmts,
		Schema:     schema,
		Table:      table,
	}
	if needsStructureDestructiveConfirm(s.profile, cs.ops) {
		out.NeedsDeleteConfirm = true
	}
	if warn := s.structureReviewWarnings(ctx, schema, table, cs.ops); warn != "" {
		out.Warning = warn
	}
	return out, nil
}

// CommitStructureChanges runs staged structure DDL.
func (s *Session) CommitStructureChanges(ctx context.Context, schema, table string, opts CommitOptions) (CommitResult, error) {
	if s.profile.ReadOnly {
		return CommitResult{}, ErrWriteBlockedReadOnly()
	}
	cs, err := s.structureOps(schema, table)
	if err != nil {
		return CommitResult{}, err
	}
	review, err := s.StructureReviewSQL(ctx, schema, table)
	if err != nil {
		return CommitResult{}, err
	}
	if review.NeedsDeleteConfirm {
		if opts.TableConfirm != table {
			return CommitResult{}, errors.New("session: table name confirmation required")
		}
	}
	det, err := s.TableMetadata(ctx, schema, table, false)
	if err != nil {
		return CommitResult{}, err
	}
	d := s.main.Dialect()
	spec := structureSpec(s, schema, table)
	batch, err := sqlgen.BuildStructureAlterStatements(d, spec, det, sqlgen.StructureAlterPlan{Ops: cs.ops})
	if err != nil {
		return CommitResult{}, err
	}
	for _, st := range batch.InTransaction {
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
	for _, st := range batch.AfterCommit {
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
	engine := strings.ToLower(s.profile.Engine)
	if engine == "mysql" || engine == "mariadb" {
		return s.commitStructureMySQL(ctx, schema, table, batch)
	}
	return s.commitStructurePostgres(ctx, schema, table, batch)
}

func (s *Session) getStructureChanges(schema, table string) *structureChangeSet {
	s.changeMu.RLock()
	cs := s.structureChanges[structureKey(schema, table)]
	s.changeMu.RUnlock()
	return cs
}

func (s *Session) structureOps(schema, table string) (*structureChangeSet, error) {
	cs := s.getStructureChanges(schema, table)
	if cs == nil || len(cs.ops) == 0 {
		return nil, errors.New("session: no staged structure changes")
	}
	return cs, nil
}

func structureSpec(s *Session, schema, table string) sqlgen.StructureSpec {
	spec := sqlgen.StructureSpec{Schema: schema, Table: table, Engine: s.profile.Engine}
	if major := mysqlMajorVersion(s.serverInfo.Version); major > 0 {
		spec.MySQLMajor = major
	}
	return spec
}

func mysqlMajorVersion(version string) int {
	i := 0
	for i < len(version) && version[i] >= '0' && version[i] <= '9' {
		if version[i] == '.' {
			break
		}
		i++
	}
	if i == 0 {
		return 0
	}
	var major int
	_, _ = fmt.Sscanf(version[:i], "%d", &major)
	return major
}

func displayStructureBatch(d driver.Dialect, batch sqlgen.StructureAlterBatch) ([]string, int) {
	var parts []string
	n := 0
	for _, st := range batch.InTransaction {
		parts = append(parts, sqlgen.DisplaySQL(d, st.SQL, st.Args))
		n++
	}
	for _, st := range batch.AfterCommit {
		parts = append(parts, sqlgen.DisplaySQL(d, st.SQL, st.Args))
		n++
	}
	return parts, n
}

func needsStructureDestructiveConfirm(p store.Profile, ops []sqlgen.StructureOp) bool {
	if !needsDeleteConfirm(p) {
		return false
	}
	for _, op := range ops {
		if isDestructiveStructureOp(op) {
			return true
		}
	}
	return false
}

func isDestructiveStructureOp(op sqlgen.StructureOp) bool {
	switch op.Kind {
	case sqlgen.StructureOpDropColumn, sqlgen.StructureOpDropIndex, sqlgen.StructureOpDropForeignKey:
		return true
	case sqlgen.StructureOpAlterColumnType:
		return typeChangeMayNarrow(op.Previous.Type, op.ColumnDef.Type)
	default:
		return false
	}
}

func typeChangeMayNarrow(from, to string) bool {
	from = strings.ToLower(strings.TrimSpace(from))
	to = strings.ToLower(strings.TrimSpace(to))
	if from == to {
		return false
	}
	if strings.Contains(from, "text") && !strings.Contains(to, "text") {
		return true
	}
	if strings.HasPrefix(from, "varchar") && strings.HasPrefix(to, "char") {
		return true
	}
	return false
}

func (s *Session) structureReviewWarnings(ctx context.Context, schema, table string, ops []sqlgen.StructureOp) string {
	var warns []string
	engine := strings.ToLower(s.profile.Engine)
	if engine == "mysql" || engine == "mariadb" {
		warns = append(warns, "MySQL DDL auto-commits each statement and cannot be rolled back as a single transaction.")
	}
	for _, op := range ops {
		if op.Kind == sqlgen.StructureOpAlterColumnNullable && !op.ColumnDef.Nullable {
			if s.columnHasNulls(ctx, schema, table, op.Column) {
				warns = append(warns, fmt.Sprintf("Column %q has NULL values; SET NOT NULL may fail.", op.Column))
			}
		}
		if op.Kind == sqlgen.StructureOpAlterColumnType && typeChangeMayNarrow(op.Previous.Type, op.ColumnDef.Type) {
			warns = append(warns, fmt.Sprintf("Type change on %q may narrow stored values.", op.Column))
		}
	}
	return strings.Join(warns, " ")
}

func (s *Session) columnHasNulls(ctx context.Context, schema, table, column string) bool {
	if s.main == nil || column == "" {
		return false
	}
	d := s.main.Dialect()
	q := fmt.Sprintf("SELECT 1 FROM %s.%s WHERE %s IS NULL LIMIT 1",
		d.QuoteIdent(schema), d.QuoteIdent(table), d.QuoteIdent(column))
	rows, err := s.main.Query(ctx, q)
	if err != nil {
		return false
	}
	defer func() { _ = rows.Close() }()
	w, err := rows.NextWindow(1)
	if err != nil || w == nil || w.Rows == 0 {
		return false
	}
	return true
}

func (s *Session) commitStructurePostgres(ctx context.Context, schema, table string, batch sqlgen.StructureAlterBatch) (CommitResult, error) {
	tx, err := s.main.Begin(ctx)
	if err != nil {
		return CommitResult{}, err
	}
	n := 0
	for _, st := range batch.InTransaction {
		if _, err := tx.Exec(ctx, st.SQL, st.Args...); err != nil {
			_ = tx.Rollback(ctx)
			return CommitResult{}, err
		}
		n++
	}
	if err := tx.Commit(ctx); err != nil {
		return CommitResult{}, err
	}
	for _, st := range batch.AfterCommit {
		if _, err := s.main.Exec(ctx, st.SQL, st.Args...); err != nil {
			return CommitResult{}, err
		}
		n++
	}
	s.finishStructureCommit(schema, table, n)
	return CommitResult{Committed: n}, nil
}

func (s *Session) commitStructureMySQL(ctx context.Context, schema, table string, batch sqlgen.StructureAlterBatch) (CommitResult, error) {
	n := 0
	for _, st := range append(batch.InTransaction, batch.AfterCommit...) {
		if _, err := s.main.Exec(ctx, st.SQL, st.Args...); err != nil {
			return CommitResult{}, err
		}
		n++
	}
	s.finishStructureCommit(schema, table, n)
	return CommitResult{Committed: n}, nil
}

func (s *Session) finishStructureCommit(schema, table string, n int) {
	_ = n
	_ = s.DiscardStructureChanges(schema, table)
	s.InvalidateMetadata()
}
