# Wisp

Local database client for PostgreSQL, MySQL, and MariaDB. One Go binary serves the UI on `127.0.0.1`; the browser runs Preact + TypeScript (Vite, Tailwind CSS v4).

## Development

```bash
cd web && npm ci && cd ..
make lint
make test
make build
./bin/wisp --version
make dev   # Vite dev server + go run ./cmd/wisp -dev
```

See [AGENTS.md](AGENTS.md) and [docs/PRD.md](docs/PRD.md) for architecture and rules.

End users: [docs/user/README.md](docs/user/README.md) · Landing page: [website/index.html](website/index.html)

Licensed under the [Apache License 2.0](LICENSE).
