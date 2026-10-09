//go:build integration

package postgres_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver/postgres"
)

func TestListDropDependentsIncludesView(t *testing.T) {
	conn := openConn(t, 0)
	defer func() { _ = conn.Close() }()
	execFixture(t, conn, "ci/fixtures/pg_objects.sql")

	deps, err := postgres.ListDropDependents(t.Context(), conn, "wisp_meta", "parent")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range deps {
		if d.Schema == "wisp_meta" && d.Name == "parent_v" && d.Kind == "view" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected parent_v in dependents, got %v", deps)
	}
}
