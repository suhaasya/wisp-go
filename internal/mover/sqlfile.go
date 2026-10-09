package mover

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

// SQLImportProgress reports byte progress for a SQL script import.
type SQLImportProgress struct {
	BytesRead  int64
	BytesTotal int64
	StmtIndex  int
	StmtTotal  int
}

// ImportSQLFile executes statements from path sequentially.
func ImportSQLFile(
	ctx context.Context,
	conn driver.Conn,
	path string,
	exec func(ctx context.Context, sql string) error,
	onProgress func(SQLImportProgress),
) (int, error) {
	f, err := openImportFile(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	total := st.Size()
	var buf strings.Builder
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var read int64
	for sc.Scan() {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		line := sc.Text()
		read += int64(len(line)) + 1
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return 0, err
	}
	script := buf.String()
	ranges := sqlgen.SplitStatements(script)
	done := 0
	for i, r := range ranges {
		if err := ctx.Err(); err != nil {
			return done, err
		}
		stmt := strings.TrimSpace(script[r.Start:r.End])
		if stmt == "" {
			continue
		}
		if err := exec(ctx, stmt); err != nil {
			return done, err
		}
		done++
		if onProgress != nil {
			onProgress(SQLImportProgress{
				BytesRead:  read,
				BytesTotal: total,
				StmtIndex:  i + 1,
				StmtTotal:  len(ranges),
			})
		}
	}
	return done, nil
}

// SQLPreview describes a SQL file without executing it.
type SQLPreview struct {
	SizeBytes int64  `json:"size_bytes"`
	Sample    string `json:"sample"`
	StmtCount int    `json:"stmt_count"`
}

// PreviewSQL reads the start of a SQL file and counts statements in the first 256 KiB sample.
func PreviewSQL(ctx context.Context, path string) (SQLPreview, error) {
	f, err := openImportFile(path)
	if err != nil {
		return SQLPreview{}, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return SQLPreview{}, err
	}
	sample, err := io.ReadAll(io.LimitReader(f, 256<<10))
	if err != nil {
		return SQLPreview{}, err
	}
	if err := ctx.Err(); err != nil {
		return SQLPreview{}, err
	}
	text := string(sample)
	ranges := sqlgen.SplitStatements(text)
	return SQLPreview{
		SizeBytes: st.Size(),
		Sample:    strings.TrimSpace(text),
		StmtCount: len(ranges),
	}, nil
}
