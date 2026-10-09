# Contributing to Wisp

Read [AGENTS.md](AGENTS.md) before opening a pull request. It defines memory budgets, package layout, and review expectations.

## Prerequisites

- Go 1.24 or later (`CGO_ENABLED=0` for all builds)
- Node.js LTS and npm
- [golangci-lint](https://golangci-lint.run/welcome/install/)

## Setup

```bash
cd web && npm ci && cd ..
make lint
make test
make build
```

## Workflow

- Follow test-driven development (`.claude/skills/tdd/SKILL.md`).
- Ticket ids in commit messages (e.g. `LW-001: ...`).
- Run `make lint` and `make test` before pushing.
