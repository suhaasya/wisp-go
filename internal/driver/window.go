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
		if rows > 0 {
			cd.offsetLen = 2 * rows
			*textSlot += 2 * rows
		}
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

// GrowWindow returns a copy of w with extra empty rows appended.
func GrowWindow(w *Window, extra int) *Window {
	if w == nil || extra <= 0 {
		return CloneWindow(w)
	}
	out := NewWindow(w.Columns, w.Rows+extra)
	if len(w.arena) > 0 {
		out.arena = append([]byte(nil), w.arena...)
		copy(out.textOffsets, w.textOffsets[:len(w.textOffsets)])
	}
	for ci := range w.cols {
		src := &w.cols[ci]
		dst := out.Column(ci)
		dst.data.nulls = append([]int(nil), src.nulls...)
		dst.data.truncated = append([]int(nil), src.truncated...)
		switch src.kind {
		case colInt:
			copy(dst.data.ints[:w.Rows], src.ints)
		case colFloat:
			copy(dst.data.floats[:w.Rows], src.floats)
		case colBool:
			copy(dst.data.bools[:w.Rows], src.bools)
		case colText:
			if w.Rows > 0 {
				n := 2 * w.Rows
				offSrc := src.offsetOff
				offDst := dst.data.offsetOff
				copy(out.textOffsets[offDst:offDst+n], w.textOffsets[offSrc:offSrc+n])
			}
		}
	}
	return out
}

// CloneWindow deep-copies a window for overlay application without mutating cache.
func CloneWindow(w *Window) *Window {
	if w == nil {
		return nil
	}
	out := NewWindow(w.Columns, w.Rows)
	if len(w.arena) > 0 {
		out.arena = append([]byte(nil), w.arena...)
		copy(out.textOffsets, w.textOffsets)
	}
	for ci := range w.cols {
		src := &w.cols[ci]
		dst := out.Column(ci)
		dst.data.nulls = append([]int(nil), src.nulls...)
		dst.data.truncated = append([]int(nil), src.truncated...)
		switch src.kind {
		case colInt:
			copy(dst.data.ints, src.ints)
		case colFloat:
			copy(dst.data.floats, src.floats)
		case colBool:
			copy(dst.data.bools, src.bools)
		}
	}
	return out
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
		w.textOffsets = make([]int, textCols*2*rows)
		if rows > 0 {
			w.arena = make([]byte, 0, textCols*rows*32)
		}
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

// CellInt reads an int64 column at row (for tests and pager).
func (w *Window) CellInt(col, row int) (int64, bool) {
	if col < 0 || col >= len(w.cols) || row < 0 || row >= w.Rows {
		return 0, false
	}
	cd := &w.cols[col]
	if cd.kind != colInt {
		return 0, false
	}
	if isNullRow(cd.nulls, row) {
		return 0, false
	}
	return cd.ints[row], true
}

// CellFloat reads a float64 column at row.
func (w *Window) CellFloat(col, row int) (float64, bool) {
	if col < 0 || col >= len(w.cols) || row < 0 || row >= w.Rows {
		return 0, false
	}
	cd := &w.cols[col]
	if cd.kind != colFloat {
		return 0, false
	}
	if isNullRow(cd.nulls, row) {
		return 0, false
	}
	return cd.floats[row], true
}

// CellBool reads a bool column at row.
func (w *Window) CellBool(col, row int) (bool, bool) {
	if col < 0 || col >= len(w.cols) || row < 0 || row >= w.Rows {
		return false, false
	}
	cd := &w.cols[col]
	if cd.kind != colBool {
		return false, false
	}
	if isNullRow(cd.nulls, row) {
		return false, false
	}
	return cd.bools[row], true
}

// IsTruncated reports whether a text cell was cut to maxCellBytes in the window.
func (w *Window) IsTruncated(col, row int) bool {
	if col < 0 || col >= len(w.cols) || row < 0 || row >= w.Rows {
		return false
	}
	cd := &w.cols[col]
	for _, r := range cd.truncated {
		if r == row {
			return true
		}
	}
	return false
}

// CellString formats any column at row as a string for display and detail panel.
func (w *Window) CellString(col, row int) (string, bool) {
	if col < 0 || col >= len(w.cols) || row < 0 || row >= w.Rows {
		return "", false
	}
	cd := &w.cols[col]
	if isNullRow(cd.nulls, row) {
		return "", false
	}
	switch cd.kind {
	case colInt:
		return strconv.FormatInt(cd.ints[row], 10), true
	case colFloat:
		return strconv.FormatFloat(cd.floats[row], 'g', -1, 64), true
	case colBool:
		if cd.bools[row] {
			return "true", true
		}
		return "false", true
	case colText:
		return w.CellText(col, row)
	default:
		return "", false
	}
}

// CellText reads a text arena column at row.
func (w *Window) CellText(col, row int) (string, bool) {
	if col < 0 || col >= len(w.cols) || row < 0 || row >= w.Rows {
		return "", false
	}
	cd := &w.cols[col]
	if cd.kind != colText {
		return "", false
	}
	if isNullRow(cd.nulls, row) {
		return "", false
	}
	start, end, ok := textCellBounds(w.textOffsets, cd.offsetOff, cd.offsetLen, row)
	if !ok || start > end || end > len(w.arena) {
		return "", false
	}
	return string(w.arena[start:end]), true
}

func textCellBounds(textOffsets []int, offsetOff, offsetLen, row int) (start, end int, ok bool) {
	if offsetLen < 2 || row < 0 {
		return 0, 0, false
	}
	half := offsetLen / 2
	if row >= half {
		return 0, 0, false
	}
	start = textOffsets[offsetOff+row]
	end = textOffsets[offsetOff+half+row]
	return start, end, true
}

func isNullRow(nulls []int, row int) bool {
	for _, n := range nulls {
		if n == row {
			return true
		}
	}
	return false
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

func (b *ColumnBuilder) textRowOffsets() (starts, ends []int) {
	half := b.data.offsetLen / 2
	base := b.data.offsetOff
	return b.w.textOffsets[base : base+half], b.w.textOffsets[base+half : base+b.data.offsetLen]
}

func (b *ColumnBuilder) appendText(row int, s string) {
	truncated := false
	if len(s) > maxCellBytes {
		s = s[:maxCellBytes]
		truncated = true
	}
	start := len(b.w.arena)
	b.w.arena = append(b.w.arena, s...)
	starts, ends := b.textRowOffsets()
	if row < 0 || row >= len(starts) {
		return
	}
	starts[row] = start
	ends[row] = len(b.w.arena)
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
		if rows == 0 || col.offsetLen == 0 {
			break
		}
		for r := 0; r < rows; r++ {
			if r > 0 {
				buf.WriteByte(',')
			}
			start, end, ok := textCellBounds(textOffsets, col.offsetOff, col.offsetLen, r)
			if !ok || start > end || end > len(arena) {
				writeJSONString(buf, nil)
				continue
			}
			writeJSONString(buf, arena[start:end])
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
