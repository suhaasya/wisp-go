package driver_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestClassifyProcessHighlight(t *testing.T) {
	t.Parallel()
	if got := driver.ClassifyProcessHighlight(5, true); got != driver.HighlightBlocked {
		t.Fatalf("blocked = %q", got)
	}
	if got := driver.ClassifyProcessHighlight(31, false); got != driver.HighlightLongRunning {
		t.Fatalf("long = %q", got)
	}
	if got := driver.ClassifyProcessHighlight(10, false); got != "" {
		t.Fatalf("ok = %q", got)
	}
}
