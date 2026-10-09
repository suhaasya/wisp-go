package session

import (
	"context"
	"regexp"
)

var reSimpleFrom = regexp.MustCompile(`(?is)\bfrom\s+(?:ONLY\s+)?("([^"]+)"|` + "`([^`]+)`" + `|([a-zA-Z_][\w$]*)(?:\.([a-zA-Z_][\w$]*))?)\b`)

// singleTableFromSelect reports one unqualified FROM target for simple SELECTs.
func singleTableFromSelect(sql string) (schema, table string, ok bool) {
	if ClassifyStatement(sql) != StatementRead {
		return "", "", false
	}
	m := reSimpleFrom.FindStringSubmatch(sql)
	if m == nil {
		return "", "", false
	}
	if m[2] != "" {
		return "public", m[2], true
	}
	if m[3] != "" {
		return "public", m[3], true
	}
	if m[5] != "" {
		return m[4], m[5], true
	}
	if m[4] != "" {
		return "public", m[4], true
	}
	return "", "", false
}

func (s *Session) queryResultSource(ctx context.Context, sql string) (schema, table string, keyCols []string) {
	schema, table, ok := singleTableFromSelect(sql)
	if !ok {
		return "", "", nil
	}
	if schema == "" {
		schema = "public"
	}
	det, err := s.TableMetadata(ctx, schema, table, false)
	if err != nil || len(det.PrimaryKey) == 0 {
		return "", "", nil
	}
	return schema, table, append([]string(nil), det.PrimaryKey...)
}
