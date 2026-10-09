package mover

import (
	"io"
	"strings"
)

// CSVWriter writes RFC4180-style CSV rows.
type CSVWriter struct {
	w io.Writer
}

// NewCSVWriter returns a CSV row writer.
func NewCSVWriter(w io.Writer) *CSVWriter {
	return &CSVWriter{w: w}
}

// FormatCSVField quotes fields that contain comma, quote, or newline.
func FormatCSVField(s string) string {
	if s == "" {
		return ""
	}
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// WriteRow writes one CSV record.
func (c *CSVWriter) WriteRow(fields []string) error {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = FormatCSVField(f)
	}
	_, err := io.WriteString(c.w, strings.Join(parts, ",")+"\n")
	return err
}
