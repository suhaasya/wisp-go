package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Options configures a self-update run.
type Options struct {
	CurrentVersion string
	ExecPath       string
	PublicKeyPEM   string
	// Repository is "owner/name" on GitHub. Ignored when ReleaseURL is set.
	Repository string
	// ReleaseURL overrides the GitHub latest-release URL (for tests).
	ReleaseURL string
	// GOOS and GOARCH override the platform for asset selection (tests).
	GOOS       string
	GOARCH     string
	HTTPClient *http.Client
	// Replace installs the new binary (defaults to ReplaceExecutable).
	Replace func(exePath string, newBytes []byte) error
	Out     io.Writer
}

// Result holds the version transition.
type Result struct {
	From string
	To   string
}

// Run downloads, verifies, and installs the latest release.
func Run(ctx context.Context, opts Options) (Result, error) {
	if opts.PublicKeyPEM == "" {
		return Result{}, fmt.Errorf("update: missing public key")
	}
	if opts.ExecPath == "" {
		return Result{}, fmt.Errorf("update: missing executable path")
	}
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	replace := opts.Replace
	if replace == nil {
		replace = ReplaceExecutable
	}
	out := opts.Out
	if out == nil {
		out = io.Discard
	}

	releaseURL := opts.ReleaseURL
	if releaseURL == "" {
		repo := opts.Repository
		if repo == "" {
			repo = "suhaasya/wisp-go"
		}
		releaseURL = "https://api.github.com/repos/" + repo + "/releases/latest"
	}

	relJSON, err := httpGet(ctx, client, releaseURL)
	if err != nil {
		return Result{}, err
	}
	rel, err := ParseReleaseJSON(relJSON)
	if err != nil {
		return Result{}, err
	}
	newVer := NormalizeVersion(rel.TagName)
	if newVer == opts.CurrentVersion {
		_, _ = fmt.Fprintf(out, "wisp: already at %s\n", opts.CurrentVersion)
		return Result{From: opts.CurrentVersion, To: newVer}, nil
	}

	assets, err := SelectAsset(rel, opts.GOOS, opts.GOARCH)
	if err != nil {
		return Result{}, err
	}

	checksums, err := httpGet(ctx, client, assets.ChecksumsURL)
	if err != nil {
		return Result{}, err
	}
	checkSig, err := httpGet(ctx, client, assets.SignatureURL)
	if err != nil {
		return Result{}, err
	}
	archive, err := httpGet(ctx, client, assets.ArchiveURL)
	if err != nil {
		return Result{}, err
	}

	if err := VerifyReleaseArchive(archive, assets.ArchiveName, checksums, checkSig, opts.PublicKeyPEM); err != nil {
		return Result{}, err
	}

	bin, err := ExtractWispBinary(assets.ArchiveName, archive)
	if err != nil {
		return Result{}, err
	}
	if err := replace(opts.ExecPath, bin); err != nil {
		return Result{}, err
	}
	_, _ = fmt.Fprintf(out, "wisp: updated %s → %s\n", opts.CurrentVersion, newVer)
	return Result{From: opts.CurrentVersion, To: newVer}, nil
}

func httpGet(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	if strings.Contains(url, "api.github.com") {
		req.Header.Set("Accept", "application/vnd.github+json")
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update: download: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update: download %s: HTTP %d", url, res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 128<<20))
	if err != nil {
		return nil, fmt.Errorf("update: read body: %w", err)
	}
	return body, nil
}
