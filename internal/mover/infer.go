package mover

import (
	"strconv"
	"strings"
)

// InferredColumn is a proposed column name and SQL type.
type InferredColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// InferColumnTypes guesses PostgreSQL-flavoured types from sample rows (header + data).
func InferColumnTypes(header []string, rows [][]string) []InferredColumn {
	out := make([]InferredColumn, len(header))
	for i, name := range header {
		var samples []string
		for _, row := range rows {
			if i < len(row) {
				samples = append(samples, row[i])
			}
		}
		out[i] = InferredColumn{Name: name, Type: inferSQLType(samples)}
	}
	return out
}

func inferSQLType(samples []string) string {
	allInt := true
	allNum := true
	allBool := true
	nonEmpty := 0
	maxLen := 0
	for _, s := range samples {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		nonEmpty++
		if len(s) > maxLen {
			maxLen = len(s)
		}
		if s == "true" || s == "false" || s == "t" || s == "f" {
			continue
		}
		allBool = false
		if _, err := strconv.ParseInt(s, 10, 64); err != nil {
			allInt = false
		}
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			allNum = false
		}
	}
	if nonEmpty == 0 {
		return "text"
	}
	if allBool && nonEmpty > 0 {
		return "boolean"
	}
	if allInt {
		return "bigint"
	}
	if allNum {
		return "double precision"
	}
	if maxLen > 255 {
		return "text"
	}
	return "varchar(255)"
}
