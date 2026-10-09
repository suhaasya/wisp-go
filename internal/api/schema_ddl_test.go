package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestGetTableDDLFromCachedMetadata(t *testing.T) {
	s := session.NewTestSession(store.Profile{Engine: "postgres"})
	s.SeedTableMetadataForTest("wisp_meta", "parent", &driver.TableDetails{
		Schema: "wisp_meta",
		Table:  "parent",
		Columns: []driver.ColumnDetail{
			{Name: "id", Type: "bigint", Nullable: false},
		},
		PrimaryKey: []string{"id"},
	})

	mgr := session.NewManager(store.NewProfileStore(t.TempDir()+"/c.json"), nil, nil)
	mgr.AdoptSession(s)
	h := api.NewSessionsHandler(mgr, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"/api/v1/sessions/test/tables/wisp_meta/parent/ddl",
		nil,
	)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		DDL string `json:"ddl"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.DDL, "CREATE TABLE") || !strings.Contains(out.DDL, "parent") {
		t.Fatalf("ddl=%q", out.DDL)
	}
}
