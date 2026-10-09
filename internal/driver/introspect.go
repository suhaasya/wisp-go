package driver

import "context"

// ObjectKind is a schema object category for sidebar browsing.
type ObjectKind string

const (
	ObjectKindTable     ObjectKind = "table"
	ObjectKindView      ObjectKind = "view"
	ObjectKindMatView   ObjectKind = "matview"
	ObjectKindFunction  ObjectKind = "function"
	ObjectKindProcedure ObjectKind = "procedure"
	ObjectKindSequence  ObjectKind = "sequence"
	ObjectKindEnum      ObjectKind = "enum"
	ObjectKindTrigger   ObjectKind = "trigger"
)

// ColumnDetail describes one table column.
type ColumnDetail struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	Default    string `json:"default,omitempty"`
	Comment    string `json:"comment,omitempty"`
	OrdinalPos int    `json:"ordinal"`
}

// ForeignKeyDetail describes a foreign key constraint.
type ForeignKeyDetail struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefSchema  string   `json:"ref_schema"`
	RefTable   string   `json:"ref_table"`
	RefColumns []string `json:"ref_columns"`
}

// IndexDetail describes an index.
type IndexDetail struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary,omitempty"`
}

// CheckDetail describes a check constraint.
type CheckDetail struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

// TriggerDetail describes a trigger on a table.
type TriggerDetail struct {
	Name    string `json:"name"`
	Timing  string `json:"timing"`
	Events  string `json:"events"`
	Enabled bool   `json:"enabled"`
}

// TableDetails is full metadata for one table or view.
type TableDetails struct {
	Schema      string             `json:"schema"`
	Table       string             `json:"table"`
	Kind        string             `json:"kind"`
	Columns     []ColumnDetail     `json:"columns"`
	PrimaryKey  []string           `json:"primary_key,omitempty"`
	ForeignKeys []ForeignKeyDetail `json:"foreign_keys,omitempty"`
	Indexes     []IndexDetail      `json:"indexes,omitempty"`
	Checks      []CheckDetail      `json:"checks,omitempty"`
	Triggers    []TriggerDetail    `json:"triggers,omitempty"`
}

// Introspector loads schema metadata through a live connection.
type Introspector interface {
	Databases(ctx context.Context) ([]string, error)
	Schemas(ctx context.Context) ([]string, error)
	Objects(ctx context.Context, schema string, kind ObjectKind) ([]string, error)
	TableDetails(ctx context.Context, schema, table string) (*TableDetails, error)
}
