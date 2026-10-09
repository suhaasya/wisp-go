package session

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
	"github.com/suhaasya/wisp-go/internal/store"
)

// DDLAction is a catalog DDL operation from the UI.
type DDLAction string

const (
	DDLCreateTable  DDLAction = "create_table"
	DDLCreateSchema DDLAction = "create_schema"
	DDLTruncate     DDLAction = "truncate"
	DDLDropTable    DDLAction = "drop_table"
	DDLRenameTable  DDLAction = "rename_table"
)

// TableWizardColumn is one column in the new-table wizard.
type TableWizardColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
	Primary  bool   `json:"primary,omitempty"`
}

// TableWizardIndex is one index in the new-table wizard.
type TableWizardIndex struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}

// TableWizardInput is the new-table wizard payload.
type TableWizardInput struct {
	Table   string              `json:"table"`
	Columns []TableWizardColumn `json:"columns"`
	Indexes []TableWizardIndex  `json:"indexes,omitempty"`
}

// DDLPreviewRequest describes a DDL preview/execute.
type DDLPreviewRequest struct {
	Action  DDLAction         `json:"action"`
	Schema  string            `json:"schema"`
	Table   string            `json:"table,omitempty"`
	NewName string            `json:"new_name,omitempty"`
	Cascade bool              `json:"cascade,omitempty"`
	Wizard  *TableWizardInput `json:"wizard,omitempty"`
}

// DDLDependent is a dependent object listed before DROP.
type DDLDependent struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
}

// DDLPreviewResult is preview SQL and safety metadata.
type DDLPreviewResult struct {
	SQL               string         `json:"sql"`
	Dependents        []DDLDependent `json:"dependents,omitempty"`
	NeedsTypedConfirm bool           `json:"needs_typed_confirm"`
	ConfirmName       string         `json:"confirm_name,omitempty"`
	Warning           string         `json:"warning,omitempty"`
}

// DDLExecuteOptions controls DDL execution confirmation.
type DDLExecuteOptions struct {
	Confirm       bool
	ObjectConfirm string
}

// PreviewDDL builds SQL and dependency metadata without executing.
func (s *Session) PreviewDDL(ctx context.Context, req DDLPreviewRequest) (DDLPreviewResult, error) {
	if s.main == nil {
		return DDLPreviewResult{}, errors.New("session: not connected")
	}
	d := s.main.Dialect()
	engine := s.profile.Engine
	sqlText, confirmName, err := s.buildDDL(ctx, d, engine, req)
	if err != nil {
		return DDLPreviewResult{}, err
	}
	out := DDLPreviewResult{SQL: sqlText, ConfirmName: confirmName}
	if req.Action == DDLDropTable && strings.EqualFold(engine, "postgres") {
		deps, err := postgres.ListDropDependents(ctx, s.main, req.Schema, req.Table)
		if err != nil {
			return DDLPreviewResult{}, err
		}
		for _, dep := range deps {
			out.Dependents = append(out.Dependents, DDLDependent{
				Schema: dep.Schema, Name: dep.Name, Kind: dep.Kind,
			})
		}
	}
	if needsDDLTypedConfirm(s.profile, req.Action) {
		out.NeedsTypedConfirm = true
	}
	if strings.EqualFold(engine, "mysql") || strings.EqualFold(engine, "mariadb") {
		if isDestructiveDDLAction(req.Action) {
			out.Warning = "MySQL DDL auto-commits and cannot be rolled back."
		}
	}
	return out, nil
}

// ExecuteDDL runs one DDL statement after guard and typed confirmation checks.
func (s *Session) ExecuteDDL(ctx context.Context, req DDLPreviewRequest, opts DDLExecuteOptions) error {
	if s.profile.ReadOnly {
		return ErrWriteBlockedReadOnly()
	}
	preview, err := s.PreviewDDL(ctx, req)
	if err != nil {
		return err
	}
	if preview.NeedsTypedConfirm && opts.ObjectConfirm != preview.ConfirmName {
		return errors.New("session: object name confirmation required")
	}
	check, err := s.Guard(preview.SQL, opts.Confirm)
	if err != nil {
		return err
	}
	if check == WriteBlocked {
		return ErrWriteBlockedReadOnly()
	}
	if check == WriteNeedsConfirmation {
		return errors.New("session: ddl needs confirmation")
	}
	if _, err := s.main.Exec(ctx, preview.SQL); err != nil {
		return err
	}
	s.InvalidateMetadata()
	return nil
}

func (s *Session) buildDDL(ctx context.Context, d driver.Dialect, engine string, req DDLPreviewRequest) (sqlText, confirmName string, err error) {
	_ = ctx
	switch req.Action {
	case DDLCreateTable:
		if req.Wizard == nil || req.Schema == "" {
			return "", "", errors.New("session: wizard and schema required")
		}
		det := wizardToDetails(req.Schema, req.Wizard)
		if strings.EqualFold(engine, "mysql") || strings.EqualFold(engine, "mariadb") {
			return sqlgen.CreateTableMySQL(det), req.Wizard.Table, nil
		}
		return sqlgen.CreateTablePostgreSQL(det), req.Wizard.Table, nil
	case DDLCreateSchema:
		if req.Schema == "" {
			return "", "", errors.New("session: name required")
		}
		return sqlgen.CreateSchemaSQL(d, engine, req.Schema), req.Schema, nil
	case DDLTruncate:
		if req.Schema == "" || req.Table == "" {
			return "", "", errors.New("session: schema and table required")
		}
		return sqlgen.TruncateTableSQL(d, req.Schema, req.Table), req.Table, nil
	case DDLDropTable:
		if req.Schema == "" || req.Table == "" {
			return "", "", errors.New("session: schema and table required")
		}
		cascade := req.Cascade && strings.EqualFold(engine, "postgres")
		return sqlgen.DropTableSQL(d, req.Schema, req.Table, cascade), req.Table, nil
	case DDLRenameTable:
		if req.Schema == "" || req.Table == "" || req.NewName == "" {
			return "", "", errors.New("session: schema, table and new_name required")
		}
		return sqlgen.RenameTableSQL(d, engine, req.Schema, req.Table, req.NewName), req.Table, nil
	default:
		return "", "", fmt.Errorf("session: unknown ddl action %q", req.Action)
	}
}

func wizardToDetails(schema string, w *TableWizardInput) *driver.TableDetails {
	var pk []string
	cols := make([]driver.ColumnDetail, 0, len(w.Columns))
	for i, c := range w.Columns {
		if c.Primary {
			pk = append(pk, c.Name)
		}
		cols = append(cols, driver.ColumnDetail{
			Name: c.Name, Type: c.Type, Nullable: c.Nullable, Default: c.Default, OrdinalPos: i + 1,
		})
	}
	var idxs []driver.IndexDetail
	for _, ix := range w.Indexes {
		idxs = append(idxs, driver.IndexDetail{Name: ix.Name, Columns: ix.Columns, Unique: ix.Unique})
	}
	return &driver.TableDetails{
		Schema: schema, Table: w.Table, Columns: cols, PrimaryKey: pk, Indexes: idxs,
	}
}

func needsDDLTypedConfirm(p store.Profile, action DDLAction) bool {
	if !needsDeleteConfirm(p) {
		return false
	}
	return isDestructiveDDLAction(action)
}

func isDestructiveDDLAction(action DDLAction) bool {
	switch action {
	case DDLTruncate, DDLDropTable, DDLRenameTable:
		return true
	default:
		return false
	}
}
