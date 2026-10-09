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
