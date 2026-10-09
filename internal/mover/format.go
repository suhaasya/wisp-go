package mover

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// Format is an export encoding.
type Format string

const (
	FormatCSV       Format = "csv"
	FormatJSON      Format = "json"
	FormatSQLInsert Format = "sql"
	FormatMarkdown  Format = "markdown"
)

// Options configures export writers.
type Options struct {
	Table         string
	IncludeHeader bool
	ColStart      int // negative: from first column
	ColEnd        int // inclusive; negative: through last column
	OnProgress    func(rowsWritten int64)
}

type rowFormatter interface {
	WriteHeader(w io.Writer, cols []driver.ColumnMeta) error
	WriteRow(w io.Writer, cols []driver.ColumnMeta, cells []string) error
	Close(w io.Writer) error
}

func newFormatter(format Format, opts Options, cols []driver.ColumnMeta) rowFormatter {
	switch format {
	case FormatJSON:
		return &jsonFormatter{cols: cols, first: true}
	case FormatSQLInsert:
		return &sqlInsertFormatter{table: opts.Table, cols: cols}
	case FormatMarkdown:
		return &markdownFormatter{cols: cols, wroteHeader: false}
	default:
		return &csvFormatter{header: opts.IncludeHeader, wroteHeader: false}
	}
}

type csvFormatter struct {
	header      bool
	wroteHeader bool
}

func (f *csvFormatter) WriteHeader(w io.Writer, cols []driver.ColumnMeta) error {
	if !f.header || f.wroteHeader {
		return nil
	}
	f.wroteHeader = true
	return NewCSVWriter(w).WriteRow(columnNames(cols))
}

func (f *csvFormatter) WriteRow(w io.Writer, _ []driver.ColumnMeta, cells []string) error {
	return NewCSVWriter(w).WriteRow(cells)
}

func (f *csvFormatter) Close(w io.Writer) error { return nil }

type jsonFormatter struct {
	cols  []driver.ColumnMeta
	first bool
}

func (f *jsonFormatter) WriteHeader(w io.Writer, _ []driver.ColumnMeta) error {
	_, err := io.WriteString(w, "[")
	return err
}

func (f *jsonFormatter) WriteRow(w io.Writer, cols []driver.ColumnMeta, cells []string) error {
	obj := make(map[string]string, len(cols))
	for i, col := range cols {
		v := ""
		if i < len(cells) {
			v = cells[i]
		}
		obj[col.Name] = v
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if !f.first {
		if _, err := io.WriteString(w, ","); err != nil {
			return err
		}
	}
	f.first = false
	_, err = w.Write(b)
	return err
}

func (f *jsonFormatter) Close(w io.Writer) error {
	if f.first {
		_, err := io.WriteString(w, "[]")
		return err
	}
	_, err := io.WriteString(w, "]")
	return err
}

type sqlInsertFormatter struct {
	table string
	cols  []driver.ColumnMeta
}

func (f *sqlInsertFormatter) WriteHeader(w io.Writer, _ []driver.ColumnMeta) error { return nil }

func (f *sqlInsertFormatter) WriteRow(w io.Writer, cols []driver.ColumnMeta, cells []string) error {
	table := f.table
	if table == "" {
		table = "exported"
	}
	names := columnNames(cols)
	vals := make([]string, len(cells))
	for i, c := range cells {
		if c == "" {
			vals[i] = "NULL"
		} else {
			vals[i] = quoteSQLString(c)
		}
	}
	line := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
		table, strings.Join(names, ", "), strings.Join(vals, ", "))
	_, err := io.WriteString(w, line)
	return err
}

func (f *sqlInsertFormatter) Close(w io.Writer) error { return nil }

type markdownFormatter struct {
	cols        []driver.ColumnMeta
	wroteHeader bool
}

func (f *markdownFormatter) WriteHeader(w io.Writer, cols []driver.ColumnMeta) error {
	return f.writeHeaderRow(w, cols)
}

func (f *markdownFormatter) WriteRow(w io.Writer, cols []driver.ColumnMeta, cells []string) error {
	if !f.wroteHeader {
		if err := f.writeHeaderRow(w, cols); err != nil {
			return err
		}
	}
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = strings.ReplaceAll(c, "|", `\|`)
	}
	_, err := io.WriteString(w, "| "+strings.Join(parts, " | ")+" |\n")
	return err
}

func (f *markdownFormatter) writeHeaderRow(w io.Writer, cols []driver.ColumnMeta) error {
	names := columnNames(cols)
	if _, err := io.WriteString(w, "| "+strings.Join(names, " | ")+" |\n"); err != nil {
		return err
	}
	sep := make([]string, len(names))
	for i := range sep {
		sep[i] = "---"
	}
	if _, err := io.WriteString(w, "| "+strings.Join(sep, " | ")+" |\n"); err != nil {
		return err
	}
	f.wroteHeader = true
	return nil
}

func (f *markdownFormatter) Close(w io.Writer) error { return nil }

func columnNames(cols []driver.ColumnMeta) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name
	}
	return out
}

func quoteSQLString(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}
