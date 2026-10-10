# Install and update

## Requirements

- macOS, Linux, or Windows
- A supported browser (see [browser support](../browser-support.md))
- For install script verification: `curl`, `sha256sum` (Unix), and [minisign](https://github.com/jedisct1/minisign)

## One-line install

```sh
curl -fsSL https://raw.githubusercontent.com/suhaasya/wisp-go/main/scripts/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/suhaasya/wisp-go/main/scripts/install.ps1 | iex
```

Default locations:

- Unix: `~/.local/bin/wisp` (override with `WISP_INSTALL_DIR`)
- Windows: `%LOCALAPPDATA%\Programs\wisp\wisp.exe`

## Package managers

- **Homebrew:** tap `suhaasya/homebrew-wisp`, formula `wisp`
- **Scoop:** bucket `suhaasya/scoop-wisp`

See [release.md](../release.md) for maintainer details.

## Self-update

Updates run only when you ask:

```sh
wisp --update
```

The binary is replaced after checksum and minisign signature verification.

## Run

```sh
wisp
```

Your default browser opens to the local UI. To stop the server:

```sh
wisp stop
```

Or quit from the system tray / menu where your platform exposes it.
