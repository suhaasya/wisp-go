package mover

import (
	"bufio"
	"context"
	"io"

	"github.com/suhaasya/wisp-go/internal/driver"
)

const exportWindowRows = 500

// StreamRows reads from an open row cursor and writes formatted export data.
func StreamRows(ctx context.Context, rows driver.Rows, format Format, opts Options, dst io.Writer) (written int64, err error) {
	buf := bufio.NewWriterSize(dst, 64<<10)
	defer func() { _ = buf.Flush() }()

	fullCols := rows.Columns()
	outCols, _ := applyColSlice(fullCols, nil, opts.ColStart, opts.ColEnd)
	fmttr := newFormatter(format, opts, outCols)
	if err := fmttr.WriteHeader(buf, outCols); err != nil {
		return 0, err
	}
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		w, err := rows.NextWindow(exportWindowRows)
		if err != nil {
			return written, err
		}
		if w == nil || w.Rows == 0 {
			break
		}
		for r := 0; r < w.Rows; r++ {
			cells := make([]string, len(fullCols))
			for c := 0; c < len(fullCols); c++ {
				s, ok := w.CellText(c, r)
				if !ok {
					cells[c] = ""
				} else {
					cells[c] = s
				}
			}
			_, cells = applyColSlice(fullCols, cells, opts.ColStart, opts.ColEnd)
			if err := fmttr.WriteRow(buf, outCols, cells); err != nil {
				return written, err
			}
			written++
		}
		if err := buf.Flush(); err != nil {
			return written, err
		}
		if opts.OnProgress != nil {
			opts.OnProgress(written)
		}
		if flushable, ok := dst.(interface{ Flush() }); ok {
			flushable.Flush()
		}
	}
	if err := fmttr.Close(buf); err != nil {
		return written, err
	}
	return written, buf.Flush()
}

// StreamQuery opens sql on conn and streams formatted rows to dst.
func StreamQuery(ctx context.Context, conn driver.Conn, sql string, args []any, format Format, opts Options, dst io.Writer) (int64, error) {
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	return StreamRows(ctx, rows, format, opts, dst)
}

func applyColSlice(cols []driver.ColumnMeta, cells []string, start, end int) ([]driver.ColumnMeta, []string) {
	if start < 0 && end < 0 {
		if cells == nil {
			return cols, nil
		}
		return cols, cells
	}
	n := len(cols)
	if n == 0 {
		return cols, cells
	}
	if end < 0 || end >= n {
		end = n - 1
	}
	if start < 0 {
		start = 0
	}
	if start > end {
		return nil, nil
	}
	outCols := cols[start : end+1]
	if cells == nil {
		return outCols, nil
	}
	return outCols, cells[start : end+1]
}
