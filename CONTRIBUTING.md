# Contributing to Wisp

Read [AGENTS.md](AGENTS.md) before opening a pull request. It defines memory budgets, package layout, and review expectations.

## Prerequisites

- Go 1.24 or later (`CGO_ENABLED=0` for all builds)
- Node.js 22 and npm
- [golangci-lint](https://golangci-lint.run/welcome/install/) v2
- Docker (optional, for local `make gate` with database fixtures)

## Setup

```bash
cd web && npm ci && cd ..
make lint
make test
make build
```

## CI (GitHub Actions)

**`ci`** and **`gate`** run on pull requests and pushes to `main`. **`release`** runs on version tags (`v*`). Run the same checks locally before you push:

| Check | Command |
| --- | --- |
| Lint + unit tests | `make lint` and `make test` |
| Size / memory budgets | `make size bench gate` (Docker for full gate) |

Secrets and tokens: see [docs/release.md](docs/release.md) and the “GitHub secrets” section in the repo wiki or README (release signing uses `MINISIGN_SECRET_KEY`).

If branch protection on `main` requires **ci** / **gate**, ensure those workflow names match after you push.

## Workflow

- Follow test-driven development (`.claude/skills/tdd/SKILL.md`).
- Ticket ids in commit messages (e.g. `LW-001: ...`).
- Run `make lint` and `make test` before pushing; run `make size bench gate` when changing binaries or bundles.
