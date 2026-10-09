package driver

import "context"

// CollectTextColumn reads one text column from all rows of a query.
func CollectTextColumn(ctx context.Context, conn Conn, sql string, args ...any) ([]string, error) {
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		w, err := rows.NextWindow(500)
		if err != nil {
			return nil, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			s, ok := w.CellText(0, r)
			if !ok {
				out = append(out, "")
				continue
			}
			out = append(out, s)
		}
	}
	return out, nil
}
