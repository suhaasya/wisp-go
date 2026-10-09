package update

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

// GitHubRelease is a subset of GET /repos/{owner}/{repo}/releases/latest.
type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

// GitHubAsset is one release download.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// AssetSet names the three files needed for an update.
type AssetSet struct {
	ArchiveName  string
	ArchiveURL   string
	ChecksumsURL string
	SignatureURL string
}

// SelectAsset picks archive, checksums, and signature URLs for this platform.
func SelectAsset(rel GitHubRelease, goos, goarch string) (AssetSet, error) {
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	osToken := goos
	if osToken == "darwin" {
		osToken = "darwin"
	}
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	wantArchive := fmt.Sprintf("_%s_%s%s", osToken, goarch, ext)
	var set AssetSet
	for _, a := range rel.Assets {
		name := a.Name
		switch {
		case strings.HasSuffix(name, wantArchive) && strings.HasPrefix(name, "wisp_"):
			set.ArchiveName = name
			set.ArchiveURL = a.BrowserDownloadURL
		case name == "checksums.txt":
			set.ChecksumsURL = a.BrowserDownloadURL
		case name == "checksums.txt.minisig":
			set.SignatureURL = a.BrowserDownloadURL
		}
	}
	if set.ArchiveURL == "" || set.ChecksumsURL == "" || set.SignatureURL == "" {
		return AssetSet{}, fmt.Errorf("update: release %q missing assets for %s/%s", rel.TagName, goos, goarch)
	}
	return set, nil
}

// ParseReleaseJSON decodes GitHub release metadata.
func ParseReleaseJSON(data []byte) (GitHubRelease, error) {
	var rel GitHubRelease
	if err := json.Unmarshal(data, &rel); err != nil {
		return GitHubRelease{}, fmt.Errorf("update: release json: %w", err)
	}
	if rel.TagName == "" {
		return GitHubRelease{}, fmt.Errorf("update: release missing tag_name")
	}
	return rel, nil
}

// NormalizeVersion strips a leading "v" from tag names.
func NormalizeVersion(tag string) string {
	return strings.TrimPrefix(strings.TrimSpace(tag), "v")
}
