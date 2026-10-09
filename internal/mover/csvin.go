package mover

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const previewRowCap = 100

// CSVPreview is the first rows of a CSV/TSV file on disk.
type CSVPreview struct {
	Encoding  string     `json:"encoding"`
	Delimiter string     `json:"delimiter"`
	Header    []string   `json:"header"`
	Rows      [][]string `json:"rows"`
	HasHeader bool       `json:"has_header"`
	SizeBytes int64      `json:"size_bytes"`
}

// PreviewCSV reads up to maxRows data rows from path (plus one header row when detected).
func PreviewCSV(ctx context.Context, path string, maxRows int) (CSVPreview, error) {
	if maxRows <= 0 || maxRows > previewRowCap {
		maxRows = previewRowCap
	}
	f, err := openImportFile(path)
	if err != nil {
		return CSVPreview{}, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return CSVPreview{}, err
	}
	raw, err := io.ReadAll(io.LimitReader(f, 512*1024))
	if err != nil {
		return CSVPreview{}, err
	}
	if err := ctx.Err(); err != nil {
		return CSVPreview{}, err
	}
	text, enc := decodeImportText(raw)
	delimiter := detectCSVDelimiter(firstLine(text))
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = rune(delimiter[0])
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	all, err := r.ReadAll()
	if err != nil {
		return CSVPreview{}, err
	}
	if len(all) == 0 {
		return CSVPreview{Encoding: enc, Delimiter: delimiter, SizeBytes: st.Size(), HasHeader: true}, nil
	}
	header := all[0]
	data := all[1:]
	if len(data) > maxRows {
		data = data[:maxRows]
	}
	return CSVPreview{
		Encoding:  enc,
		Delimiter: delimiter,
		Header:    header,
		Rows:      data,
		HasHeader: true,
		SizeBytes: st.Size(),
	}, nil
}

func decodeImportText(b []byte) (string, string) {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	if utf8.Valid(b) {
		return string(b), "utf-8"
	}
	return string(b), "binary"
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSuffix(s[:i], "\r")
	}
	return strings.TrimSuffix(s, "\r")
}

func detectCSVDelimiter(line string) string {
	if strings.Count(line, "\t") >= strings.Count(line, ",") && strings.Contains(line, "\t") {
		return "\t"
	}
	if strings.Count(line, ";") > strings.Count(line, ",") {
		return ";"
	}
	return ","
}

// CSVReader streams CSV records from path after optional header skip.
type CSVReader struct {
	r         *csv.Reader
	f         *os.File
	delimiter rune
	done      bool
}

// OpenCSV opens path for streaming import.
func OpenCSV(path string, delimiter string, skipHeader bool) (*CSVReader, error) {
	f, err := openImportFile(path)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(f, 3))
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, err
	}
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		if _, err := f.Seek(3, io.SeekStart); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	if delimiter == "" {
		delimiter = ","
	}
	cr := csv.NewReader(f)
	cr.Comma = rune(delimiter[0])
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	out := &CSVReader{r: cr, f: f, delimiter: rune(delimiter[0])}
	if skipHeader {
		if _, err := cr.Read(); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	return out, nil
}

// Read returns the next row or io.EOF when finished.
func (c *CSVReader) Read() ([]string, error) {
	if c.done {
		return nil, io.EOF
	}
	row, err := c.r.Read()
	if errors.Is(err, io.EOF) {
		c.done = true
		return nil, io.EOF
	}
	return row, err
}

// Close closes the underlying file.
func (c *CSVReader) Close() error {
	return c.f.Close()
}

// LineNumber approximates the source line for error reporting (1-based, after header).
func (c *CSVReader) LineNumber() int64 {
	return 0
}
