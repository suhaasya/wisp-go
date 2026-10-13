package driver_test

import (
	"encoding/json"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestWindowMarshalMultipleTextColumnsOneRow(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int4"},
		{Name: "first_name", Type: "text"},
		{Name: "last_name", Type: "text"},
		{Name: "official_email", Type: "text"},
	}, 1)
	w.Column(0).SetInt(0, 1)
	w.Column(1).SetText(0, "Akshay")
	w.Column(2).SetText(0, "Smith")
	w.Column(3).SetText(0, "a@b.c")
	w.Rows = 1

	got, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Data [][]any `json:"data"`
	}
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Data) != 4 {
		t.Fatalf("columns = %d", len(doc.Data))
	}
	if doc.Data[2][0] != "Smith" || doc.Data[3][0] != "a@b.c" {
		t.Fatalf("data = %v", doc.Data)
	}
}
