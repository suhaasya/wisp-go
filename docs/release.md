# Release and install

Wisp ships as a single static Go binary per platform with an embedded web UI.

## One-line install

```sh
curl -fsSL https://raw.githubusercontent.com/suhaasya/wisp-go/main/scripts/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/suhaasya/wisp-go/main/scripts/install.ps1 | iex
```

Requires `curl`, `sha256sum` (Linux/macOS), and [minisign](https://github.com/jedisct1/minisign) on `PATH`.

Install location:

- Unix: `~/.local/bin/wisp` (override with `WISP_INSTALL_DIR`)
- Windows: `%LOCALAPPDATA%\Programs\wisp\wisp.exe`

## Self-update

Updates run **only** when you ask:

```sh
wisp --update
# or
wisp update
```

The running binary is replaced after checksum and minisign signature verification using the public key embedded at build time (`release/minisign.pub`).

## Package managers

GoReleaser publishes:

- Homebrew tap: `suhaasya/homebrew-wisp` (formula `wisp`)
- Scoop bucket: `suhaasya/scoop-wisp`

## GitHub Actions and secrets

Workflows live in `.github/workflows/`:

| Workflow | Trigger | Secrets needed |
|----------|---------|----------------|
| **ci** | PR + push to `main` | None (uses built-in token) |
| **gate** | PR + push to `main` | None |
| **release** | Push tag `v*` | See below |

### You do **not** create `GITHUB_TOKEN`

GitHub injects **`secrets.GITHUB_TOKEN`** automatically on every workflow run. It lets the job checkout code, upload artifacts, and create GitHub Releases on **`wisp-go`**.

In **Settings → Actions → General**, use **“Read and write permissions”** for workflows (or grant `contents: write` in the release job — already set in `release.yml`) so releases can publish assets.

### Release signing: `MINISIGN_SECRET_KEY` (you create this)

1. Install [minisign](https://github.com/jedisct1/minisign) locally.
2. Generate a keypair (once). **CI requires an unencrypted key** (no passphrase):
   ```sh
   minisign -G -W -p release/minisign.pub -s minisign.key
   ```
   The secret file must start with `untrusted comment: minisign secret key` (not `encrypted secret key`).
   **Commit** `release/minisign.pub` and the same content in `internal/update/trusted.pub` (embedded in the binary). **Never commit** `minisign.key`.
3. Put the **entire secret key file** into GitHub:
   - Repo → **Settings → Secrets and variables → Actions → New repository secret**
   - Name: `MINISIGN_SECRET_KEY`
   - Value: paste the full contents of `minisign.key` (including the `untrusted comment:` line and base64 block), as GoReleaser expects for `minisign -S -m`.

If this secret is **missing**, release still builds archives but **skips** signing (see `.goreleaser.yaml`); install/`--update` will fail signature checks until signing is enabled.

### Homebrew / Scoop (optional, separate repos)

GoReleaser tries to push formula/manifest updates to **`suhaasya/homebrew-wisp`** and **`suhaasya/scoop-wisp`**. Create those empty GitHub repos first.

The default `GITHUB_TOKEN` often **cannot** push to other repositories. Create a **Personal Access Token (classic)** or fine-grained PAT with **contents: write** on those repos, then add:

- Secret name: `HOMEBREW_TAP_GITHUB_TOKEN` (GoReleaser default for brew taps)

and wire it in `.goreleaser.yaml` under `brews[].repository.token` / scoop bucket if pushes fail. You can also set `skip_upload: true` on brews/scoops until taps exist.

### Apple notarization (optional)

Only if you macOS-sign releases. App Store Connect API key + team ID — add the secrets named in `.github/workflows/release.yml` (`AC_API_KEY_*`, `APPLE_*`). Hooks are not fully wired yet; safe to leave unset.

### After enabling CI

Push to GitHub (workflows do not run only on your laptop):

```sh
git push origin main
```

Open a PR or watch **Actions** on github.com. No PAT needed for normal **ci** / **gate**.

## Maintainer release flow

1. Replace `release/minisign.pub` / CI `MINISIGN_SECRET_KEY` with your production minisign keypair (`minisign -G`).
2. Tag `vX.Y.Z` and push; `.github/workflows/release.yml` runs GoReleaser.
3. Artifacts: five archives, `checksums.txt`, `checksums.txt.minisig`.
4. Optional macOS codesign/notarization: set Apple secrets listed in `release.yml` (skipped when unset).

Local snapshot (no publish):

```sh
make release-snapshot
```

## Verification

- **Download size:** each release archive must stay within the **10 MB** hard cap (7 MB target); CI checks tarball/zip sizes on tag builds.
- **Tampering:** install scripts and `wisp --update` reject bad checksums or signatures.

## Clean VM smoke test

On a fresh macOS, Linux, or Windows VM:

1. Install minisign.
2. Run the install script for your OS.
3. Run `wisp version`.
4. Run `wisp --update` when a newer tag exists (optional).

Report results in your release checklist.
