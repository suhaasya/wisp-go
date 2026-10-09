package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	extminisign "github.com/jedisct1/go-minisign"
)

const testSignSecretKey = `untrusted comment: minisign encrypted secret key
RWQAAEIyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOItWpGuGQbG4C9WXaxEYLgZ2xxuqfbuZmDgAhQ8Unot8t7SyxZ0nVh0gESesJ6Ay57fGFJ9T1ajVmanT7MFMCCDbPZ8uqDcSAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
`

const testSignPublicKey = `untrusted comment: minisign public key B141866BA4568B38
RWQ4i1aka4ZBsR0gESesJ6Ay57fGFJ9T1ajVmanT7MFMCCDbPZ8uqDcS
`

func TestRun_fakeRelease(t *testing.T) {
	t.Parallel()
	wispBin := []byte("#!/bin/sh\necho wisp-test\n")
	archiveName := "wisp_9.9.9_linux_amd64.tar.gz"
	archive := buildTarGz(t, map[string][]byte{"wisp": wispBin})
	sum := sha256.Sum256(archive)
	checksums := []byte(hex.EncodeToString(sum[:]) + "  " + archiveName + "\n")
	sigBytes := signChecksums(t, checksums)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			body := `{"tag_name":"v9.9.9","assets":[` +
				`{"name":"` + archiveName + `","browser_download_url":"` + base + `/dl/archive"},` +
				`{"name":"checksums.txt","browser_download_url":"` + base + `/dl/checksums"},` +
				`{"name":"checksums.txt.minisig","browser_download_url":"` + base + `/dl/sig"}` +
				`]}`
			_, _ = w.Write([]byte(body))
		case r.URL.Path == "/dl/archive":
			_, _ = w.Write(archive)
		case r.URL.Path == "/dl/checksums":
			_, _ = w.Write(checksums)
		case r.URL.Path == "/dl/sig":
			_, _ = w.Write(sigBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	exe := filepath.Join(dir, "wisp")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	res, err := Run(context.Background(), Options{
		CurrentVersion: "1.0.0",
		ExecPath:       exe,
		PublicKeyPEM:   testSignPublicKey,
		ReleaseURL:     srv.URL + "/releases/latest",
		GOOS:           "linux",
		GOARCH:         "amd64",
		HTTPClient:     srv.Client(),
		Out:            &out,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.To != "9.9.9" {
		t.Fatalf("To = %q", res.To)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(wispBin) {
		t.Fatalf("installed binary mismatch")
	}
	if !strings.Contains(out.String(), "1.0.0") || !strings.Contains(out.String(), "9.9.9") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRun_rejectsTamperedArchive(t *testing.T) {
	t.Parallel()
	wispBin := []byte("good")
	archiveName := "wisp_9.9.9_linux_amd64.tar.gz"
	archive := buildTarGz(t, map[string][]byte{"wisp": wispBin})
	sum := sha256.Sum256(archive)
	checksums := []byte(hex.EncodeToString(sum[:]) + "  " + archiveName + "\n")
	sigBytes := signChecksums(t, checksums)
	tampered := append(append([]byte{}, archive...), 'x')

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			body := `{"tag_name":"v9.9.9","assets":[` +
				`{"name":"` + archiveName + `","browser_download_url":"` + base + `/dl/archive"},` +
				`{"name":"checksums.txt","browser_download_url":"` + base + `/dl/checksums"},` +
				`{"name":"checksums.txt.minisig","browser_download_url":"` + base + `/dl/sig"}` +
				`]}`
			_, _ = w.Write([]byte(body))
		case r.URL.Path == "/dl/archive":
			_, _ = w.Write(tampered)
		case r.URL.Path == "/dl/checksums":
			_, _ = w.Write(checksums)
		case r.URL.Path == "/dl/sig":
			_, _ = w.Write(sigBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	exe := filepath.Join(dir, "wisp")
	_ = os.WriteFile(exe, []byte("old"), 0o755)

	_, err := Run(context.Background(), Options{
		CurrentVersion: "1.0.0",
		ExecPath:       exe,
		PublicKeyPEM:   testSignPublicKey,
		ReleaseURL:     srv.URL + "/releases/latest",
		GOOS:           "linux",
		GOARCH:         "amd64",
		HTTPClient:     srv.Client(),
	})
	if err == nil {
		t.Fatal("expected tampered archive to fail")
	}
}

func buildTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func signChecksums(t *testing.T, checksums []byte) []byte {
	t.Helper()
	sk, err := extminisign.DecodePrivateKey(testSignSecretKey)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := sk.Sign(checksums, extminisign.SignOptions{
		TrustedComment: "timestamp:1\tfile:checksums.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	return []byte(sig.Encode())
}
