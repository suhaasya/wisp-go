package pager

import (
	"bytes"
	"encoding/json"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func buildWindowJSON(rid string, offset int, total Total, w *driver.Window, ov Overlay) ([]byte, error) {
	inner, err := w.MarshalJSON()
	if err != nil {
		return nil, err
	}
	ovBytes, err := json.Marshal(ov.wire())
	if err != nil {
		return nil, err
	}
	totalBytes, err := json.Marshal(total)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Grow(len(inner) + 128)
	buf.WriteString(`{"rid":`)
	buf.WriteString(strJSON(rid))
	buf.WriteString(`,"offset":`)
	buf.WriteString(intJSON(offset))
	buf.WriteString(`,"total":`)
	buf.Write(totalBytes)
	buf.WriteByte(',')
	if len(inner) == 0 || inner[0] != '{' {
		return nil, json.Unmarshal(inner, &struct{}{})
	}
	buf.Write(inner[1:])
	if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] == '}' {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteString(`,"overlay":`)
	buf.Write(ovBytes)
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func strJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func intJSON(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
