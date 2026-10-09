package sqlgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// SortKey is one ORDER BY column.
type SortKey struct {
	Column string `json:"column"`
	Desc   bool   `json:"desc"`
}

// FilterCondition is one structured filter row from the filter bar.
type FilterCondition struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Value  string `json:"value"`
	Value2 string `json:"value2,omitempty"`
}

// FilterGroup combines conditions with AND or OR.
type FilterGroup struct {
	Logic      string            `json:"logic"` // "and" or "or"
	Conditions []FilterCondition `json:"conditions"`
}

// BrowseSpec describes a table browse query.
type BrowseSpec struct {
	Schema         string
	Table          string
	AllowedColumns []string
	KeyColumns     []string
	Filter         FilterGroup
	RawWhere       string
	Sort           []SortKey
	Limit          int
	Offset         int
}

// BuildBrowse returns SELECT * with optional WHERE, ORDER BY, and LIMIT/OFFSET.
func BuildBrowse(d driver.Dialect, spec BrowseSpec) (string, []any, error) {
	if spec.Schema == "" || spec.Table == "" {
		return "", nil, errors.New("sqlgen: schema and table required")
	}
	table := d.QuoteIdent(spec.Schema) + "." + d.QuoteIdent(spec.Table)
	allow := columnSet(spec.AllowedColumns)

	var args []any
	var where string
	var err error
	if strings.TrimSpace(spec.RawWhere) != "" {
		if err := validateRawWhere(spec.RawWhere); err != nil {
			return "", nil, err
		}
		where = "(" + strings.TrimSpace(spec.RawWhere) + ")"
	} else if len(spec.Filter.Conditions) > 0 {
		where, args, err = buildFilterWhere(d, spec.Filter, allow, args)
		if err != nil {
			return "", nil, err
		}
	}

	order, err := buildOrderBy(d, spec.Sort, spec.KeyColumns, allow)
	if err != nil {
		return "", nil, err
	}

	sqlText := "SELECT * FROM " + table
	if where != "" {
		sqlText += " WHERE " + where
	}
	if order != "" {
		sqlText += " ORDER BY " + order
	}
	if spec.Limit > 0 || spec.Offset > 0 {
		sqlText += " " + d.LimitOffset(spec.Limit, spec.Offset)
	}
	return sqlText, args, nil
}

func columnSet(cols []string) map[string]struct{} {
	m := make(map[string]struct{}, len(cols))
	for _, c := range cols {
		m[c] = struct{}{}
	}
	return m
}

func requireColumn(name string, allow map[string]struct{}) error {
	if len(allow) == 0 {
		return nil
	}
	if _, ok := allow[name]; !ok {
		return fmt.Errorf("sqlgen: unknown column %q", name)
	}
	return nil
}

func buildOrderBy(d driver.Dialect, sort []SortKey, keys []string, allow map[string]struct{}) (string, error) {
	var parts []string
	seen := make(map[string]struct{})
	for _, s := range sort {
		if err := requireColumn(s.Column, allow); err != nil {
			return "", err
		}
		dir := "ASC"
		if s.Desc {
			dir = "DESC"
		}
		parts = append(parts, d.QuoteIdent(s.Column)+" "+dir)
		seen[s.Column] = struct{}{}
	}
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			continue
		}
		parts = append(parts, d.QuoteIdent(k)+" ASC")
	}
	return strings.Join(parts, ", "), nil
}

func buildFilterWhere(d driver.Dialect, g FilterGroup, allow map[string]struct{}, args []any) (string, []any, error) {
	logic := strings.ToLower(strings.TrimSpace(g.Logic))
	if logic == "" {
		logic = "and"
	}
	if logic != "and" && logic != "or" {
		return "", nil, fmt.Errorf("sqlgen: filter logic %q", g.Logic)
	}
	joiner := " " + strings.ToUpper(logic) + " "
	var parts []string
	for _, c := range g.Conditions {
		part, newArgs, err := buildCondition(d, c, allow, args)
		if err != nil {
			return "", nil, err
		}
		args = newArgs
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "", args, nil
	}
	return strings.Join(parts, joiner), args, nil
}

func buildCondition(d driver.Dialect, c FilterCondition, allow map[string]struct{}, args []any) (string, []any, error) {
	if err := requireColumn(c.Column, allow); err != nil {
		return "", nil, err
	}
	col := d.QuoteIdent(c.Column)
	op := strings.ToLower(strings.TrimSpace(c.Op))
	switch op {
	case "=", "eq":
		args = append(args, c.Value)
		return col + " = " + d.Placeholder(len(args)), args, nil
	case "!=", "<>", "neq":
		args = append(args, c.Value)
		return col + " <> " + d.Placeholder(len(args)), args, nil
	case "<", "lt":
		args = append(args, c.Value)
		return col + " < " + d.Placeholder(len(args)), args, nil
	case ">", "gt":
		args = append(args, c.Value)
		return col + " > " + d.Placeholder(len(args)), args, nil
	case "like":
		args = append(args, c.Value)
		return col + " LIKE " + d.Placeholder(len(args)), args, nil
	case "ilike":
		kw := "ILIKE"
		if isMySQLFamily(d) {
			kw = "LIKE"
		}
		args = append(args, c.Value)
		return col + " " + kw + " " + d.Placeholder(len(args)), args, nil
	case "in":
		vals, err := parseINList(c.Value)
		if err != nil {
			return "", nil, err
		}
		if len(vals) == 0 {
			return "", nil, errors.New("sqlgen: IN requires values")
		}
		var ph []string
		for _, v := range vals {
			args = append(args, v)
			ph = append(ph, d.Placeholder(len(args)))
		}
		return col + " IN (" + strings.Join(ph, ", ") + ")", args, nil
	case "is null", "is_null", "null":
		return col + " IS NULL", args, nil
	case "between":
		args = append(args, c.Value, c.Value2)
		return col + " BETWEEN " + d.Placeholder(len(args)-1) + " AND " + d.Placeholder(len(args)), args, nil
	default:
		return "", nil, fmt.Errorf("sqlgen: operator %q", c.Op)
	}
}

func isMySQLFamily(d driver.Dialect) bool {
	n := strings.ToLower(d.Name())
	return n == "mysql" || n == "mariadb"
}

func parseINList(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, unquoteLiteral(p))
	}
	return out, nil
}

func unquoteLiteral(s string) string {
	if len(s) >= 2 && ((s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"')) {
		return s[1 : len(s)-1]
	}
	return s
}

func validateRawWhere(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("sqlgen: empty raw WHERE")
	}
	if strings.Contains(s, ";") {
		return errors.New("sqlgen: raw WHERE must not contain ;")
	}
	upper := strings.ToUpper(s)
	for _, kw := range []string{"INSERT ", "UPDATE ", "DELETE ", "DROP ", "CREATE ", "ALTER ", "TRUNCATE "} {
		if strings.Contains(upper, kw) {
			return fmt.Errorf("sqlgen: raw WHERE must not contain %s", strings.TrimSpace(kw))
		}
	}
	return nil
}

// BrowseSQLForGuard returns the SQL string used to classify table browse (read guard).
func BrowseSQLForGuard(d driver.Dialect, spec BrowseSpec) (string, error) {
	sqlText, _, err := BuildBrowse(d, spec)
	return sqlText, err
}

// BuildPointLookup selects one column for a row matching key columns.
func BuildPointLookup(d driver.Dialect, schema, table, selectCol string, keyCols []string, keyVals []any) (string, []any, error) {
	if schema == "" || table == "" || selectCol == "" {
		return "", nil, errors.New("sqlgen: schema, table and column required")
	}
	if len(keyCols) != len(keyVals) || len(keyCols) == 0 {
		return "", nil, errors.New("sqlgen: key columns and values required")
	}
	tbl := d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
	col := d.QuoteIdent(selectCol)
	var args []any
	var parts []string
	for i, k := range keyCols {
		args = append(args, keyVals[i])
		parts = append(parts, d.QuoteIdent(k)+" = "+d.Placeholder(len(args)))
	}
	sqlText := "SELECT " + col + " FROM " + tbl + " WHERE " + strings.Join(parts, " AND ")
	return sqlText, args, nil
}

// BuildTableCount returns COUNT(*) with the same WHERE as BuildBrowse.
func BuildTableCount(d driver.Dialect, spec BrowseSpec) (string, []any, error) {
	if spec.Schema == "" || spec.Table == "" {
		return "", nil, errors.New("sqlgen: schema and table required")
	}
	table := d.QuoteIdent(spec.Schema) + "." + d.QuoteIdent(spec.Table)
	allow := columnSet(spec.AllowedColumns)

	var args []any
	var where string
	var err error
	if strings.TrimSpace(spec.RawWhere) != "" {
		if err := validateRawWhere(spec.RawWhere); err != nil {
			return "", nil, err
		}
		where = "(" + strings.TrimSpace(spec.RawWhere) + ")"
	} else if len(spec.Filter.Conditions) > 0 {
		where, args, err = buildFilterWhere(d, spec.Filter, allow, args)
		if err != nil {
			return "", nil, err
		}
	}
	sqlText := "SELECT COUNT(*) FROM " + table
	if where != "" {
		sqlText += " WHERE " + where
	}
	return sqlText, args, nil
}
