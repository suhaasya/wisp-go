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

## CI

Every pull request and every push to `main` runs two required workflows:

| Workflow | Purpose |
| --- | --- |
| **ci** | Lint (`make lint`, `govulncheck`, `npm audit`), tests on Ubuntu/macOS/Windows (`go test`, `-race` on Linux), cross-compiled binaries for all five release targets |
| **gate** | Budget gate (`make size`, `make bench`, `make gate`) on Linux; size/bench on macOS and Windows |

Download built binaries from the **ci** workflow run → Artifacts → **wisp-binaries** (darwin arm64/amd64, linux amd64/arm64, windows amd64).

Bench output is under the **gate** workflow artifact **bench-results**.

### Branch protection (repository settings)

Configure GitHub branch protection on `main` so that:

1. Required status checks **ci** and **gate** are green before merge.
2. At least **one approving review** is required.

With that setup, a PR that fails lint, vet, tests, vuln scan, or budget caps cannot be merged.

## Workflow

- Follow test-driven development (`.claude/skills/tdd/SKILL.md`).
- Ticket ids in commit messages (e.g. `LW-001: ...`).
- Run `make lint` and `make test` before pushing; run `make size bench gate` when changing binaries or bundles.
