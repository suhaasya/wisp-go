package driver

// Dialect generates engine-specific SQL fragments.
type Dialect interface {
	Name() string
	QuoteIdent(name string) string
	QuoteLiteral(s string) string
	Placeholder(n int) string
	LimitOffset(limit, offset int) string
	KeysetPredicate(columns []string, values []any, dir string) (string, []any, error)
}
