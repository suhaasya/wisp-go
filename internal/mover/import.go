package mover

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

const MaxImportErrors = 1000

// ImportKind identifies a file import format.
type ImportKind string

const (
	ImportKindCSV   ImportKind = "csv"
	ImportKindJSONL ImportKind = "jsonl"
	ImportKindSQL   ImportKind = "sql"
)

// DetectImportKind picks csv, jsonl, or sql from the file extension.
func DetectImportKind(path string) ImportKind {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json", ".jsonl", ".ndjson":
		return ImportKindJSONL
	case ".sql":
		return ImportKindSQL
	default:
		return ImportKindCSV
	}
}

// FilePreview is step-1 preview for any supported import file.
type FilePreview struct {
	Kind     ImportKind       `json:"kind"`
	Path     string           `json:"path"`
	CSV      *CSVPreview      `json:"csv,omitempty"`
	JSONL    *JSONLPreview    `json:"jsonl,omitempty"`
	SQL      *SQLPreview      `json:"sql,omitempty"`
	Inferred []InferredColumn `json:"inferred,omitempty"`
}

// PreviewImportFile returns up to 100 sample rows and inferred types.
func PreviewImportFile(ctx context.Context, path string) (FilePreview, error) {
	kind := DetectImportKind(path)
	out := FilePreview{Kind: kind, Path: path}
	switch kind {
	case ImportKindJSONL:
		j, err := PreviewJSONL(ctx, path, previewRowCap)
		if err != nil {
			return FilePreview{}, err
		}
		out.JSONL = &j
		out.Inferred = InferColumnTypes(j.Columns, j.Rows)
	case ImportKindSQL:
		s, err := PreviewSQL(ctx, path)
		if err != nil {
			return FilePreview{}, err
		}
		out.SQL = &s
	default:
		c, err := PreviewCSV(ctx, path, previewRowCap)
		if err != nil {
			return FilePreview{}, err
		}
		out.CSV = &c
		out.Inferred = InferColumnTypes(c.Header, c.Rows)
	}
	return out, nil
}

// ColumnMapping maps a file column to a table column.
type ColumnMapping struct {
	FileColumn  string `json:"file_column"`
	TableColumn string `json:"table_column"`
}

// DataImportOptions configures a CSV/JSONL import run.
type DataImportOptions struct {
	Schema      string
	Table       string
	Delimiter   string
	HasHeader   bool
	NullMarkers []string
	BatchSize   int
	StopOnError bool
	Mapping     []ColumnMapping
	FileColumns []string
}

// ImportError is one failed row or statement.
type ImportError struct {
	Line    int64  `json:"line"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
}

// ImportProgress reports rows loaded so far.
type ImportProgress struct {
	RowsLoaded int64
	BytesRead  int64
}

// ExecImport runs one parameterized statement (session wraps with Guard).
type ExecImport func(ctx context.Context, sql string, args ...any) error

// ImportCSV loads path into schema.table using batched INSERTs or COPY when available.
func ImportCSV(ctx context.Context, d driver.Dialect, conn driver.Conn, path string, opts DataImportOptions, exec ExecImport, onProgress func(ImportProgress)) (int64, []ImportError, error) {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 500
	}
	if conn != nil {
		if copier, ok := conn.(driver.CopyFromConn); ok && d.Name() == "postgres" {
			if n, err := importCSVCopyFrom(ctx, copier, path, opts, onProgress); err == nil {
				return n, nil, nil
			}
		}
	}
	r, err := OpenCSV(path, opts.Delimiter, opts.HasHeader)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = r.Close() }()
	return importDelimited(ctx, d, opts, r.Read, onProgress, exec)
}

func importCSVCopyFrom(ctx context.Context, copier driver.CopyFromConn, path string, opts DataImportOptions, onProgress func(ImportProgress)) (int64, error) {
	tableCols := mappedTableColumns(opts)
	if len(tableCols) == 0 {
		return 0, errors.New("mover: column mapping required")
	}
	r, err := OpenCSV(path, opts.Delimiter, opts.HasHeader)
	if err != nil {
		return 0, err
	}
	defer func() { _ = r.Close() }()
	src := &csvCopySource{read: r.Read, opts: opts}
	n, err := copier.CopyFrom(ctx, opts.Schema, opts.Table, tableCols, src)
	if err != nil {
		return 0, err
	}
	if onProgress != nil {
		onProgress(ImportProgress{RowsLoaded: n})
	}
	return n, nil
}

type csvCopySource struct {
	read func() ([]string, error)
	opts DataImportOptions
}

func (c *csvCopySource) Next() ([]any, error) {
	row, err := c.read()
	if err != nil {
		return nil, err
	}
	return mapRowValues(row, c.opts)
}

// ImportJSONL loads a JSON-lines file.
func ImportJSONL(ctx context.Context, d driver.Dialect, _ driver.Conn, path string, opts DataImportOptions, exec ExecImport, onProgress func(ImportProgress)) (int64, []ImportError, error) {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 500
	}
	jr, err := OpenJSONL(path)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = jr.Close() }()
	readRow := func() ([]string, error) {
		obj, line, err := jr.Read()
		if err != nil {
			return nil, err
		}
		_ = line
		out := make([]string, len(opts.FileColumns))
		for i, col := range opts.FileColumns {
			out[i] = obj[col]
		}
		return out, nil
	}
	return importDelimited(ctx, d, opts, readRow, onProgress, exec)
}

func importDelimited(
	ctx context.Context,
	d driver.Dialect,
	opts DataImportOptions,
	readRow func() ([]string, error),
	onProgress func(ImportProgress),
	exec ExecImport,
) (int64, []ImportError, error) {
	tableCols := mappedTableColumns(opts)
	if len(tableCols) == 0 {
		return 0, nil, errors.New("mover: column mapping required")
	}
	var loaded int64
	var errs []ImportError
	var batch [][]any
	line := int64(1)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		st, err := sqlgen.BuildMultiRowInsert(d, opts.Schema, opts.Table, tableCols, batch)
		if err != nil {
			return err
		}
		if err := exec(ctx, st.SQL, st.Args...); err != nil {
			return err
		}
		loaded += int64(len(batch))
		batch = batch[:0]
		if onProgress != nil {
			onProgress(ImportProgress{RowsLoaded: loaded})
		}
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return loaded, errs, err
		}
		row, err := readRow()
		if errors.Is(err, io.EOF) {
			break
		}
		line++
		if err != nil {
			errs = appendImportError(errs, ImportError{Line: line, Message: err.Error()})
			if opts.StopOnError || len(errs) >= MaxImportErrors {
				return loaded, errs, err
			}
			continue
		}
		vals, err := mapRowValues(row, opts)
		if err != nil {
			errs = appendImportError(errs, ImportError{Line: line, Message: err.Error()})
			if opts.StopOnError {
				return loaded, errs, err
			}
			continue
		}
		batch = append(batch, vals)
		if len(batch) >= opts.BatchSize {
			if err := flush(); err != nil {
				errs = appendImportError(errs, ImportError{Line: line, Message: err.Error()})
				if opts.StopOnError {
					return loaded, errs, err
				}
			}
		}
	}
	if err := flush(); err != nil {
		return loaded, errs, err
	}
	return loaded, errs, nil
}

func mappedTableColumns(opts DataImportOptions) []string {
	var cols []string
	for _, m := range opts.Mapping {
		if m.TableColumn != "" {
			cols = append(cols, m.TableColumn)
		}
	}
	return cols
}

func mapRowValues(row []string, opts DataImportOptions) ([]any, error) {
	var out []any
	for _, m := range opts.Mapping {
		if m.TableColumn == "" {
			continue
		}
		fileIdx := fileColumnIndex(opts.FileColumns, m.FileColumn)
		if fileIdx < 0 || fileIdx >= len(row) {
			return nil, errors.New("mover: missing file column")
		}
		s := row[fileIdx]
		if isNullMarker(s, opts.NullMarkers) {
			out = append(out, nil)
		} else {
			out = append(out, s)
		}
	}
	return out, nil
}

func fileColumnIndex(cols []string, name string) int {
	for i, c := range cols {
		if c == name {
			return i
		}
	}
	return -1
}

func isNullMarker(s string, markers []string) bool {
	for _, m := range markers {
		if s == m {
			return true
		}
	}
	return s == ""
}

func appendImportError(errs []ImportError, e ImportError) []ImportError {
	if len(errs) >= MaxImportErrors {
		return errs
	}
	return append(errs, e)
}

// FormatImportErrorsCSV renders errors for download.
func FormatImportErrorsCSV(errs []ImportError) string {
	var b strings.Builder
	b.WriteString("line,column,message\n")
	for _, e := range errs {
		b.WriteString(strconv.FormatInt(e.Line, 10))
		b.WriteByte(',')
		b.WriteString(FormatCSVField(e.Column))
		b.WriteByte(',')
		b.WriteString(FormatCSVField(e.Message))
		b.WriteByte('\n')
	}
	return b.String()
}
