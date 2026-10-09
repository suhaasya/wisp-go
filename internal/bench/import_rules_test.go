package bench_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/bench"
)

func TestInternalImportRules(t *testing.T) {
	t.Parallel()
	violations, err := bench.CheckInternalImportRules()
	if err != nil {
		t.Fatalf("CheckInternalImportRules: %v", err)
	}
	if len(violations) > 0 {
		for _, v := range violations {
			t.Errorf("%s imports %s (not allowed)", v.Importer, v.Imported)
		}
	}
}
