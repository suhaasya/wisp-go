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

// OpenResultFromRequest opens a result from API source JSON.
func (s *Session) OpenResultFromRequest(ctx context.Context, src pager.Source) (ResultOpenResponse, error) {
	out, err := s.OpenResult(ctx, src)
	if err != nil {
		return ResultOpenResponse{}, err
	}
	return ResultOpenResponse{RID: out.RID, Columns: out.Columns, Total: out.Total}, nil
}
