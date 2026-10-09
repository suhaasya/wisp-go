package api

import "testing"

func TestParseResultWindowPath(t *testing.T) {
	rest := "results/rid-1/window"
	path := rest
	path = stringsTrimPrefix(path, "results/")
	path = stringsTrimSuffix(path, "/window")
	if path != "rid-1" {
		t.Fatalf("rid = %q", path)
	}
}

func stringsTrimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func stringsTrimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
