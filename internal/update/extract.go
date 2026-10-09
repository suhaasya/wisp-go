package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

// ExtractWispBinary returns the wisp executable from a release archive.
func ExtractWispBinary(archiveName string, data []byte) ([]byte, error) {
	if strings.HasSuffix(strings.ToLower(archiveName), ".zip") {
		return extractZip(data)
	}
	return extractTarGz(data)
}

func extractTarGz(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("update: gzip: %w", err)
	}
	defer func() { _ = gr.Close() }()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("update: tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := hdr.Name
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		if base != "wisp" && base != "wisp.exe" {
			continue
		}
		return io.ReadAll(io.LimitReader(tr, 64<<20))
	}
	return nil, fmt.Errorf("update: wisp binary not found in archive")
}

func extractZip(data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("update: zip: %w", err)
	}
	for _, f := range zr.File {
		base := f.Name
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		if base != "wisp" && base != "wisp.exe" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, 64<<20))
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, fmt.Errorf("update: wisp binary not found in zip")
}
