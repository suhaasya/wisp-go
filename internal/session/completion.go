package session

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

var errCompletionSchemaRequired = errors.New("session: schema required")

const maxCompletionTables = 500

// CompletionColumn is one column for SQL editor autocomplete.
type CompletionColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// CompletionTable is table metadata for autocomplete.
type CompletionTable struct {
	Name    string             `json:"name"`
	Columns []CompletionColumn `json:"columns"`
}

// CompletionSchemaResponse is returned by GET .../complete.
type CompletionSchemaResponse struct {
	Schema string            `json:"schema"`
	Tables []CompletionTable `json:"tables"`
}

// CompletionSchema builds autocomplete metadata for one schema from the metadata cache.
func (s *Session) CompletionSchema(ctx context.Context, schema string, extraTables []string) (CompletionSchemaResponse, error) {
	if schema == "" {
		return CompletionSchemaResponse{}, errCompletionSchemaRequired
	}
	if len(extraTables) == 0 {
		tables, err := s.ListObjects(ctx, schema, driver.ObjectKindTable, false)
		if err != nil {
			return CompletionSchemaResponse{}, err
		}
		views, err := s.ListObjects(ctx, schema, driver.ObjectKindView, false)
		if err != nil {
			return CompletionSchemaResponse{}, err
		}
		names := mergeObjectNames(tables, views, nil)
		if len(names) > maxCompletionTables {
			names = names[:maxCompletionTables]
		}
		out := CompletionSchemaResponse{Schema: schema, Tables: make([]CompletionTable, 0, len(names))}
		for _, name := range names {
			out.Tables = append(out.Tables, CompletionTable{Name: name, Columns: []CompletionColumn{}})
		}
		return out, nil
	}

	names := mergeObjectNames(nil, nil, extraTables)
	if len(names) > maxCompletionTables {
		names = names[:maxCompletionTables]
	}
	out := CompletionSchemaResponse{Schema: schema, Tables: []CompletionTable{}}
	for _, name := range names {
		det, err := s.TableMetadata(ctx, schema, name, false)
		if err != nil {
			continue
		}
		tbl := CompletionTable{Name: name}
		for _, c := range det.Columns {
			tbl.Columns = append(tbl.Columns, CompletionColumn{Name: c.Name, Type: c.Type})
		}
		out.Tables = append(out.Tables, tbl)
	}
	return out, nil
}

func mergeObjectNames(lists ...[]string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(names []string) {
		for _, n := range names {
			n = strings.TrimSpace(n)
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	for _, list := range lists {
		add(list)
	}
	sort.Strings(out)
	return out
}
