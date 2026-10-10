# Wisp user guide

Wisp is a **local** database client: one small Go binary listens on `127.0.0.1`, and your browser shows the UI. Nothing is sent to Wisp’s servers; credentials stay in your OS keychain or local config.

## Contents

| Topic | Guide |
| --- | --- |
| Install and update | [install.md](install.md) |
| First launch | [first-run.md](first-run.md) |
| Connections (TCP, URL, TLS, SSH, sockets) | [connections.md](connections.md) |
| Browse and edit data safely | [browse-edit.md](browse-edit.md) |
| SQL editor and results | [query.md](query.md) |
| Import and export | [import-export.md](import-export.md) |
| Schema and structure tools | [structure.md](structure.md) |
| Keyboard shortcuts | [shortcuts.md](shortcuts.md) |
| Security model | [security.md](security.md) |
| Problems and fixes | [troubleshooting.md](troubleshooting.md) |

## Related docs

- [Release and install](../release.md) — install scripts, self-update, signing
- [Browser support](../browser-support.md) — supported browsers
- [Accessibility](../accessibility.md) — keyboard and screen reader notes
- [Compatibility matrix](../compatibility.md) — tested database versions (CI)

## Feedback (beta)

In the app: **Settings → About & feedback → Send feedback** (GitHub Issues), or open an issue on the project repository. Include OS, browser, and database engine version.
