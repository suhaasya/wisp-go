# Security model

## Local-only server

- Wisp binds to **127.0.0.1** (loopback), not your LAN.
- Each launch uses a secret token; the browser must present it on API requests.
- There is no cloud account or telemetry channel in 1.0 (see **Privacy** settings).

## Credentials

- Saved passwords use the OS keyring (macOS Keychain, Windows Credential Manager, Secret Service on Linux) when available.
- Logs and API errors do not include passwords, tokens, or connection URLs with credentials.

## SQL safety

- Identifiers are quoted through the dialect; values use bound parameters.
- User SQL runs through **session guard** rules and your **Data safety** preferences.
- Production-tagged connections can be read-only or require extra confirmation.

## Updates

Binary updates verify checksums and minisign signatures before replace ([install.md](install.md)).

## Your responsibility

- Anyone who can use your logged-in session can use Wisp while it runs.
- SSH keys and TLS material are read from paths you configure; protect those files as you would for any DB client.

## Config paths (typical)

| OS | Config |
| --- | --- |
| macOS | `~/Library/Application Support/Wisp/` |
| Linux | `~/.config/wisp/` |
| Windows | `%APPDATA%\Wisp\` |

Exact paths may vary slightly by platform build; use in-app help or `wisp` diagnostics if exposed.
