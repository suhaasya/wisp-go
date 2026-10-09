package mover

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sort"
)

// JSONLPreview is the first objects in a JSON-lines file.
type JSONLPreview struct {
	Encoding  string     `json:"encoding"`
	Columns   []string   `json:"columns"`
	Rows      [][]string `json:"rows"`
	SizeBytes int64      `json:"size_bytes"`
}

// PreviewJSONL reads up to maxRows objects from path.
func PreviewJSONL(ctx context.Context, path string, maxRows int) (JSONLPreview, error) {
	if maxRows <= 0 || maxRows > previewRowCap {
		maxRows = previewRowCap
	}
	f, err := openImportFile(path)
	if err != nil {
		return JSONLPreview{}, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return JSONLPreview{}, err
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	colSet := map[string]struct{}{}
	var rows []map[string]string
	line := 0
	for sc.Scan() {
		if err := ctx.Err(); err != nil {
			return JSONLPreview{}, err
		}
		line++
		lineText := sc.Bytes()
		if len(lineText) == 0 {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal(lineText, &obj); err != nil {
			return JSONLPreview{}, err
		}
		row := map[string]string{}
		for k, v := range obj {
			colSet[k] = struct{}{}
			row[k] = jsonScalarString(v)
		}
		rows = append(rows, row)
		if len(rows) >= maxRows {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return JSONLPreview{}, err
	}
	cols := make([]string, 0, len(colSet))
	for k := range colSet {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	outRows := make([][]string, len(rows))
	for i, row := range rows {
		outRows[i] = make([]string, len(cols))
		for j, c := range cols {
			outRows[i][j] = row[c]
		}
	}
	return JSONLPreview{
		Encoding:  "utf-8",
		Columns:   cols,
		Rows:      outRows,
		SizeBytes: st.Size(),
	}, nil
}

func jsonScalarString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// JSONLReader streams JSON lines from a file.
type JSONLReader struct {
	sc   *bufio.Scanner
	f    *os.File
	line int64
}

// OpenJSONL opens path for streaming import.
func OpenJSONL(path string) (*JSONLReader, error) {
	f, err := openImportFile(path)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	return &JSONLReader{sc: sc, f: f}, nil
}

// Read decodes the next object into a string map.
func (j *JSONLReader) Read() (map[string]string, int64, error) {
	for j.sc.Scan() {
		j.line++
		b := j.sc.Bytes()
		if len(b) == 0 {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal(b, &obj); err != nil {
			return nil, j.line, err
		}
		out := make(map[string]string, len(obj))
		for k, v := range obj {
			out[k] = jsonScalarString(v)
		}
		return out, j.line, nil
	}
	if err := j.sc.Err(); err != nil {
		return nil, j.line, err
	}
	return nil, j.line, io.EOF
}

// Close closes the file.
func (j *JSONLReader) Close() error {
	return j.f.Close()
}
