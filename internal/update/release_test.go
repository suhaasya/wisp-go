package update

import (
	"testing"
)

func TestSelectAsset_linux_amd64(t *testing.T) {
	t.Parallel()
	rel := GitHubRelease{
		TagName: "v1.2.3",
		Assets: []GitHubAsset{
			{Name: "wisp_1.2.3_linux_amd64.tar.gz", BrowserDownloadURL: "https://example/a.tgz"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example/checksums.txt"},
			{Name: "checksums.txt.minisig", BrowserDownloadURL: "https://example/checksums.txt.minisig"},
		},
	}
	set, err := SelectAsset(rel, "linux", "amd64")
	if err != nil {
		t.Fatalf("SelectAsset: %v", err)
	}
	if set.ArchiveName != "wisp_1.2.3_linux_amd64.tar.gz" {
		t.Fatalf("archive = %q", set.ArchiveName)
	}
	if set.ArchiveURL == "" || set.ChecksumsURL == "" || set.SignatureURL == "" {
		t.Fatalf("missing urls: %+v", set)
	}
}

func TestSelectAsset_missing(t *testing.T) {
	t.Parallel()
	rel := GitHubRelease{TagName: "v0.0.1", Assets: []GitHubAsset{{Name: "checksums.txt"}}}
	if _, err := SelectAsset(rel, "linux", "amd64"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()
	if got := NormalizeVersion("v1.0.0"); got != "1.0.0" {
		t.Fatalf("got %q", got)
	}
}
