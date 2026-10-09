package store_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
)

func TestSnippetsUpsertRoundTrip(t *testing.T) {
	paths := testPaths(t)
	st := store.NewSnippetsStore(paths)
	if err := st.Upsert(store.Snippet{Name: "Orders", SQL: "SELECT * FROM orders", Tags: []string{"report"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Snippets) != 1 || doc.Snippets[0].Name != "Orders" {
		t.Fatalf("doc=%+v", doc)
	}
}
