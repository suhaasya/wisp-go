# Changelog

All notable changes to Wisp are documented here. Version numbers follow [Semantic Versioning](https://semver.org/).

## [1.0.0] — 2026-10-10

First public release. Wisp is a local database client for PostgreSQL, MySQL, and MariaDB: one Go binary on `127.0.0.1`, UI in your browser.

### Added

- Connections: TCP, URL paste, TLS, SSH tunnels and jump hosts, Unix sockets and Windows named pipes where supported
- Browse and edit with staged change sets, SQL editor (lazy-loaded), query runner, structure tools, import/export
- Command palette, keyboard shortcuts, settings (appearance, safety, privacy, shortcuts)
- Install scripts (Unix/Windows), Homebrew/Scoop via GoReleaser, **`wisp --update`** with minisign verification
- User guide: [`docs/user/`](docs/user/README.md)
- Landing page: [`website/index.html`](website/index.html)
- Settings → **About & feedback** (GitHub Issues)
- Apache License 2.0 ([`LICENSE`](LICENSE))
- CI: budget gate (`make gate`), cross-browser E2E, database version matrix (PR + nightly), compatibility doc generation

### Security

- Server binds to loopback with per-launch token; credentials via OS keychain when available
- Release artifacts signed with minisign; public key embedded in the binary

### Notes

- Memory and size budgets are enforced in CI; see [PRD](docs/PRD.md) and the landing page for targets
- Licence: Apache 2.0 — free for personal and commercial use

[1.0.0]: https://github.com/suhaasya/wisp-go/releases/tag/v1.0.0
