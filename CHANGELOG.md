# Changelog

All notable changes to Wisp are documented here. Version numbers follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- User guide under `docs/user/`
- Static landing page under `website/`
- Settings **About & feedback** with GitHub Issues link
- E2E compatibility matrix (LW-046), browser/a11y pass (LW-044), packaging and self-update (LW-045)

## [1.0.0] — TBD

### Highlights

- Local PostgreSQL, MySQL, and MariaDB client: loopback server + browser UI
- Keyset browsing, staged change sets, SQL editor, import/export
- SSH tunnels, TLS, sockets/pipes where supported
- Signed releases and `wisp --update`

### Quality gates

- Go and browser tab memory/size budgets enforced in CI (`make gate`)
- Cross-browser E2E smoke (Chromium, Firefox, WebKit)

[Unreleased]: https://github.com/suhaasya/wisp-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/suhaasya/wisp-go/releases/tag/v1.0.0
