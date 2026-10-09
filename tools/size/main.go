// Command size records binary and bundle measurements for the budget gate.
package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/suhaasya/wisp-go/internal/bench"
)

func main() {
	bin := "bin/wisp"
	if len(os.Args) > 1 {
		bin = os.Args[1]
	}
	info, err := os.Stat(bin) //nolint:gosec // G703: bin path from Makefile default
	if err != nil {
		fmt.Fprintf(os.Stderr, "size: %v\n", err)
		os.Exit(1)
	}
	binaryMB := float64(info.Size()) / (1024 * 1024)

	gzBytes, err := gzipFileSize(bin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "size: gzip: %v\n", err)
		os.Exit(1)
	}
	downloadMB := float64(gzBytes) / (1024 * 1024)

	metrics := map[string]bench.Measurement{
		"binary_mb":   {Value: binaryMB},
		"download_mb": {Value: downloadMB},
	}

	sizesPath := filepath.Join("web", "dist", "sizes.json")
	//nolint:gosec // G304: fixed web/dist/sizes.json path
	if data, err := os.ReadFile(sizesPath); err == nil {
		var sizes struct {
			InitialJSGzip  float64            `json:"initialJSGzip"`
			InitialCSSGzip float64            `json:"initialCSSGzip"`
			LazyChunksGzip map[string]float64 `json:"lazyChunksGzip"`
		}
		if json.Unmarshal(data, &sizes) == nil {
			metrics["js_initial_kb_gz"] = bench.Measurement{Value: sizes.InitialJSGzip / 1024}
			metrics["css_kb_gz"] = bench.Measurement{Value: sizes.InitialCSSGzip / 1024}
			var editor float64
			for name, sz := range sizes.LazyChunksGzip {
				if stringsContains(name, "editor") {
					editor += sz
				}
			}
			if editor == 0 {
				metrics["js_editor_kb_gz"] = bench.Measurement{Skipped: true, Reason: "editor chunk not built (LW-031)"}
			} else {
				metrics["js_editor_kb_gz"] = bench.Measurement{Value: editor / 1024}
			}
		}
	}

	if err := bench.WriteResults("bench-results/size.json", metrics); err != nil {
		fmt.Fprintf(os.Stderr, "size: %v\n", err)
		os.Exit(1)
	}
}

func gzipFileSize(path string) (int64, error) {
	//nolint:gosec // G304: gzip of release binary from Makefile
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	var w io.Writer
	buf := &countWriter{}
	gw := gzip.NewWriter(buf)
	if _, err := io.Copy(gw, f); err != nil {
		return 0, err
	}
	if err := gw.Close(); err != nil {
		return 0, err
	}
	_ = w
	return buf.n, nil
}

type countWriter struct{ n int64 }

func (c *countWriter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

func stringsContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
