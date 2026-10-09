package pager

import (
	"context"
)

// ScrollEndWindows walks 1M-row mock result in window steps (bench helper).
func ScrollEndWindows(ctx context.Context, m *Manager, rid string, totalRows, step int) error {
	for off := 0; off < totalRows; off += step {
		if _, err := m.WindowJSON(ctx, rid, off, DefaultWindowRows); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}
