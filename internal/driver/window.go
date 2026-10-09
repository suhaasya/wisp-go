package driver

import (
	"bytes"
	"strconv"
	"sync"
)

const maxCellBytes = 4096

var jsonBufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

type columnKind int

const (
	colInt columnKind = iota
	colFloat
	colBool
	colText
)

type columnData struct {
	meta      ColumnMeta
	kind      columnKind
	ints      []int64
	floats    []float64
	bools     []bool
	offsetOff int
	offsetLen int
	nulls     []int
	truncated []int
}

func (w *Window) bindColumnData(meta ColumnMeta, rows int, textSlot *int) columnData {
	kind := columnKindForType(meta.Type)
	cd := columnData{meta: meta, kind: kind}
	switch kind {
	case colInt:
		cd.ints = make([]int64, rows)
	case colFloat:
		cd.floats = make([]float64, rows)
	case colBool:
		cd.bools = make([]bool, rows)
	case colText:
		cd.offsetOff = *textSlot
		cd.offsetLen = rows + 1
		*textSlot += rows + 1
	}
	return cd
}

func columnKindForType(typ string) columnKind {
	switch typ {
	case "int2", "int4", "int8", "integer", "bigint", "serial", "bigserial":
		return colInt
	case "float4", "float8", "double precision", "real":
		return colFloat
	case "bool", "boolean":
		return colBool
	default:
		return colText
	}
}

// Window is a column-wise block of rows.
type Window struct {
	Columns     []ColumnMeta
	Rows        int
	arena       []byte
	textOffsets []int
	cols        []columnData
}

// ArenaLen returns bytes in the shared text arena (for tests and diagnostics).
func (w *Window) ArenaLen() int { return len(w.arena) }

// ResetWindow clears row values and reuses allocations for the same column layout.
func ResetWindow(w *Window, cols []ColumnMeta, rows int) {
	if w == nil || len(w.cols) != len(cols) || w.Rows != rows {
		return
	}
	w.Columns = cols
	w.arena = w.arena[:0]
	for i := range w.textOffsets {
		w.textOffsets[i] = 0
	}
	for i := range w.cols {
		cd := &w.cols[i]
		cd.nulls = cd.nulls[:0]
		cd.truncated = cd.truncated[:0]
		switch cd.kind {
		case colInt:
			clear(cd.ints)
		case colFloat:
			clear(cd.floats)
		case colBool:
			clear(cd.bools)
		}
	}
}

// NewWindow allocates a window with n rows and column metadata.
func NewWindow(cols []ColumnMeta, rows int) *Window {
	w := &Window{Columns: cols, Rows: rows}
	textCols := 0
	for _, m := range cols {
		if columnKindForType(m.Type) == colText {
			textCols++
		}
	}
	if textCols > 0 && rows > 0 {
		w.textOffsets = make([]int, textCols*(rows+1))
		w.arena = make([]byte, 0, textCols*rows*32)
	}
	w.cols = make([]columnData, len(cols))
	slot := 0
	for i, m := range cols {
		w.cols[i] = w.bindColumnData(m, rows, &slot)
	}
	return w
}

// ColumnBuilder writes one column in a window.
type ColumnBuilder struct {
	w    *Window
	data *columnData
}

// Column returns mutable storage for column i.
func (w *Window) Column(i int) *ColumnBuilder {
	return &ColumnBuilder{w: w, data: &w.cols[i]}
}

func (b *ColumnBuilder) SetInt(row int, v int64) {
	b.data.ints[row] = v
}

func (b *ColumnBuilder) SetFloat(row int, v float64) {
	b.data.floats[row] = v
}

func (b *ColumnBuilder) SetBool(row int, v bool) {
	b.data.bools[row] = v
}

func (b *ColumnBuilder) SetText(row int, s string) {
	b.appendText(row, s)
}

func (b *ColumnBuilder) SetNull(row int) {
	b.data.nulls = appendUniqueInt(b.data.nulls, row)
}

func (b *ColumnBuilder) textOffsets() []int {
	return b.w.textOffsets[b.data.offsetOff : b.data.offsetOff+b.data.offsetLen]
}

func (b *ColumnBuilder) appendText(row int, s string) {
	truncated := false
	if len(s) > maxCellBytes {
		s = s[:maxCellBytes]
		truncated = true
	}
	start := len(b.w.arena)
	b.w.arena = append(b.w.arena, s...)
	off := b.textOffsets()
	off[row] = start
	off[row+1] = len(b.w.arena)
	if truncated {
		b.data.truncated = appendUniqueInt(b.data.truncated, row)
	}
}

func appendUniqueInt(list []int, v int) []int {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

func (w *Window) MarshalJSON() ([]byte, error) {
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jsonBufferPool.Put(buf)
	buf.Grow(len(w.arena) + len(w.Columns)*48 + 128)

	buf.WriteByte('{')
	writeColumnMetaJSON(buf, w.Columns)
	buf.WriteString(`,"data":`)
	writeDataJSON(buf, w.arena, w.textOffsets, w.cols, w.Rows)
	buf.WriteString(`,"nulls":`)
	writeIndexMatrixJSON(buf, w.cols, w.Rows, func(c *columnData) []int { return c.nulls })
	buf.WriteString(`,"truncated":`)
	writeIndexMatrixJSON(buf, w.cols, w.Rows, func(c *columnData) []int { return c.truncated })
	buf.WriteByte('}')

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func writeColumnMetaJSON(buf *bytes.Buffer, cols []ColumnMeta) {
	buf.WriteString(`"columns":[`)
	for i, c := range cols {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(`{"name":`)
		writeJSONStringFromString(buf, c.Name)
		buf.WriteString(`,"type":`)
		writeJSONStringFromString(buf, c.Type)
		buf.WriteByte('}')
	}
	buf.WriteByte(']')
}

func writeJSONStringFromString(buf *bytes.Buffer, s string) {
	writeJSONString(buf, []byte(s))
}

func writeDataJSON(buf *bytes.Buffer, arena []byte, textOffsets []int, cols []columnData, rows int) {
	buf.WriteByte('[')
	for ci, col := range cols {
		if ci > 0 {
			buf.WriteByte(',')
		}
		writeColumnValuesJSON(buf, arena, textOffsets, &col, rows)
	}
	buf.WriteByte(']')
}

func writeColumnValuesJSON(buf *bytes.Buffer, arena []byte, textOffsets []int, col *columnData, rows int) {
	buf.WriteByte('[')
	switch col.kind {
	case colInt:
		for r := 0; r < rows; r++ {
			if r > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(strconv.FormatInt(col.ints[r], 10))
		}
	case colFloat:
		for r := 0; r < rows; r++ {
			if r > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(strconv.FormatFloat(col.floats[r], 'g', -1, 64))
		}
	case colBool:
		for r := 0; r < rows; r++ {
			if r > 0 {
				buf.WriteByte(',')
			}
			if col.bools[r] {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		}
	case colText:
		off := textOffsets[col.offsetOff : col.offsetOff+col.offsetLen]
		for r := 0; r < rows; r++ {
			if r > 0 {
				buf.WriteByte(',')
			}
			writeJSONString(buf, arena[off[r]:off[r+1]])
		}
	}
	buf.WriteByte(']')
}

func writeIndexMatrixJSON(buf *bytes.Buffer, cols []columnData, rows int, pick func(*columnData) []int) {
	buf.WriteByte('[')
	for ci := range cols {
		if ci > 0 {
			buf.WriteByte(',')
		}
		idx := pick(&cols[ci])
		if len(idx) == 0 {
			buf.WriteString("[]")
			continue
		}
		buf.WriteByte('[')
		for i, v := range idx {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(strconv.Itoa(v))
		}
		buf.WriteByte(']')
	}
	buf.WriteByte(']')
}

func writeJSONString(buf *bytes.Buffer, b []byte) {
	buf.WriteByte('"')
	for _, c := range b {
		switch c {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if c < 0x20 {
				buf.WriteString(`\u00`)
				buf.WriteByte(hexDigit(c >> 4))
				buf.WriteByte(hexDigit(c & 0xf))
			} else {
				buf.WriteByte(c)
			}
		}
	}
	buf.WriteByte('"')
}

func hexDigit(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'a' + n - 10
}
