package web_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/web"
)

func TestEmbed_distContainsEntryAssets(t *testing.T) {
	t.Parallel()
	root := web.DistFS()
	hasIndex := false
	hasJS := false
	hasCSS := false
	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if path == "index.html" {
			hasIndex = true
		}
		if strings.HasPrefix(path, "assets/") && strings.HasSuffix(path, ".js") {
			hasJS = true
		}
		if strings.HasPrefix(path, "assets/") && strings.HasSuffix(path, ".css") {
			hasCSS = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasIndex {
		t.Fatal("embedded dist missing index.html (run make web first)")
	}
	if !hasJS || !hasCSS {
		t.Fatal("embedded dist missing hashed assets under assets/")
	}
}
