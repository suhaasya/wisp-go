package session

import (
	"context"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
)

// ResultOpenResponse is the POST /results JSON body.
type ResultOpenResponse struct {
	RID     string              `json:"rid"`
	Columns []driver.ColumnMeta `json:"columns"`
	Total   pager.Total         `json:"total"`
}

// OpenResultFromRequest opens a result from API fields.
func (s *Session) OpenResultFromRequest(ctx context.Context, kind, schema, table, sql string) (ResultOpenResponse, error) {
	src := pager.Source{Kind: kind, Schema: schema, Table: table, SQL: sql}
	out, err := s.OpenResult(ctx, src)
	if err != nil {
		return ResultOpenResponse{}, err
	}
	return ResultOpenResponse{RID: out.RID, Columns: out.Columns, Total: out.Total}, nil
}
