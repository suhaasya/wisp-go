# PRD — Wisp: Local Database Client (Go + Browser UI)

Oct 9, 2026

## Overview

Wisp is a local database client for MySQL, MariaDB and PostgreSQL: one Go binary that runs a private server on 127.0.0.1 and shows its interface in the user's browser, using under 30 MB for the Go process and under 70 MB for its browser tab.

**Problem.** Developers who manage MySQL and PostgreSQL databases choose between heavy Electron/JVM tools that idle at 300 MB to 1 GB and lighter commercial apps that are closed-source and often macOS-only. A browser is already open on every developer machine; a small local server plus a lean page can deliver a TablePlus-class experience at a fraction of the cost.

**Vision.** Run `wisp`, the browser opens, and within 5 seconds you are browsing any database (local, cloud or behind a bastion) with smooth scrolling over millions of rows. The Go process and the tab stay small enough that nobody notices them.

**Product principles**

1. Memory and size budgets are features. Any change that breaks a budget does not ship, on the Go side or in the tab.
2. Data lives in Go, the browser only paints. The tab holds what is on screen plus a small margin, never whole result sets.
3. Local only. The server listens on 127.0.0.1, requires a per-launch secret, and refuses everything else.
4. Safe by default: edits are staged and reviewed before commit; production connections are colour-tagged and can be read-only.
5. Stream, never load: results, imports and exports move in bounded chunks.

**What changed from the Rust/GPUI plan.** The product scope, screens, connection support and safety model are unchanged. The desktop UI framework is replaced by a browser page built with Preact and TypeScript (bundled by Vite, styled with Tailwind CSS v4), the engine moves from Rust to Go, the memory promise is split into two budgets (the Go process and our tab), and every coding ticket is built test-first.

## Goals, non-goals and resource budgets

Both budget tables are release gates: CI measures every row on every pull request, and a build that exceeds any deal-breaker value cannot merge.

**Go process budgets**

| Metric | Target | Deal breaker | How measured |
| --- | --- | --- | --- |
| Binary size (stripped, per platform, assets embedded) | ≤ 15 MB | 25 MB | CI artifact size |
| Download size (compressed) | ≤ 7 MB | 10 MB | CI archive size |
| RSS at idle, no connection, one tab attached | ≤ 20 MB | 30 MB | RSS after 30 s idle |
| RSS with 1 connection + 500-table schema | ≤ 30 MB | 45 MB | Bench scenario |
| RSS per extra open connection | ≤ 2 MB | 4 MB | Bench scenario |
| RSS with a 1M-row table open and scrolled end to end | ≤ 50 MB | 80 MB | Bench scenario |
| Start to page served | ≤ 100 ms | 300 ms | Launch trace |
| CPU at idle | ~0% | > 1% sustained | 60 s sample |
| Config + saved connections on disk | ≤ 1 MB | 5 MB | Directory size |

**Browser tab budgets** (measured in Chromium via the DevTools protocol; Firefox and Safari checked manually each release)

| Metric | Target | Deal breaker |
| --- | --- | --- |
| JavaScript heap, idle on Connections screen | ≤ 5 MB | 10 MB |
| JavaScript heap, 1M-row table open and scrolled | ≤ 10 MB | 20 MB |
| DOM nodes at any time | ≤ 2,000 | 4,000 |
| Initial JavaScript (gzip) | ≤ 60 KB | 100 KB |
| SQL editor chunk, loaded on demand (gzip) | ≤ 200 KB | 280 KB |
| CSS (gzip) | ≤ 15 KB | 25 KB |
| Whole tab, Chrome Task Manager memory footprint | ≤ 40 MB | 70 MB |
| Long tasks while scrolling the grid | 0 over 50 ms | any over 100 ms |
| First screen usable after page load | ≤ 300 ms | 600 ms |

The whole-tab row includes the browser's fixed cost per page, so it is validated with a prototype in week 1 and adjusted once if needed; the JavaScript, DOM and bundle rows are fully under our control and stay strict.

**Goals (v1.0)**

- Connect to MySQL 5.7+, MariaDB 10.4+ and PostgreSQL 12+ through TCP, TLS, SSH tunnels, sockets and named pipes.
- Browse, filter, sort and inline-edit table data with staged commits.
- Write and run SQL with highlighting, autocomplete and multiple result tabs.
- View and alter table structure; import and export CSV, JSON and SQL in streaming mode.
- Ship one binary each for macOS (arm64, x86_64), Linux (x86_64, arm64) and Windows (x86_64).
- Work in current Chrome, Edge, Firefox and Safari.

**Non-goals (v1.0)**

- Listening on any interface other than loopback, multi-user access, logins or team sharing.
- Other engines (SQLite, SQL Server, Redis, MongoDB); the driver interface allows them later.
- A bundled browser, Electron, a system webview wrapper or a tray icon (would need CGO).
- React, Vue, Angular, React-compatibility layers (preact/compat) and UI component libraries.
- ER diagram designer, query builder GUI, dashboards, AI assistant, cloud sync.

## Target users

The primary user is a backend developer with 5 to 30 saved connections across local, staging and production who keeps the tool open all day next to an IDE and Docker.

| Persona | Context | What they need most |
| --- | --- | --- |
| Backend developer | Docker MySQL/Postgres locally, staging on RDS, production behind a bastion | Fast switching, SSH tunnels, safe production edits |
| Full-stack / indie developer | Supabase, Neon, PlanetScale, Railway | Paste a connection URL and go; TLS that just works |
| DBA / SRE | Many servers, large tables, incident work | Process list, read-only mode, low overhead |
| Data analyst | Read-heavy, exports | Fast filtering, CSV/JSON export, saved queries |
| Developer on a low-spec machine | 8 GB laptop, Linux | A tool that stays small next to Docker and the IDE |

## Connection support

v1.0 supports every engine, transport and authentication method below; a connection is created from a form or by pasting a URL.

| Engine | Versions | Default port | Go driver |
| --- | --- | --- | --- |
| PostgreSQL | 12 – 17 (plus Supabase, Neon, Aurora PG, Timescale) | 5432 | `jackc/pgx/v5` (native protocol, no `database/sql`) |
| MySQL | 5.7, 8.0, 8.4+ (plus PlanetScale, Aurora MySQL, Cloud SQL) | 3306 | `go-sql-driver/mysql` |
| MariaDB | 10.4+ | 3306 | `go-sql-driver/mysql` with MariaDB dialect |

| Transport | Requirement |
| --- | --- |
| TCP | Host, port, database; IPv4 and IPv6 |
| Unix socket | Auto-detect common PostgreSQL and MySQL socket paths (Homebrew, Postgres.app, Linux packages, Docker) |
| Windows named pipe | MySQL on Windows |
| TLS | Modes disable, prefer, require, verify-ca, verify-full; custom CA; client certificate and key; system trust store via Go `crypto/tls` |
| SSH tunnel | Password, private key (Ed25519, ECDSA, RSA, with passphrase), ssh-agent, `~/.ssh/config` aliases, `known_hosts` verification, via `golang.org/x/crypto/ssh` |
| SSH jump hosts | Up to 3 hops, including `ProxyJump` from `~/.ssh/config` |

**Authentication.** PostgreSQL: SCRAM-SHA-256, MD5, trust, client certificate, `.pgpass`. MySQL: `caching_sha2_password`, `mysql_native_password`, client certificate. Option to ask for the password on every connect and never store it. Cloud IAM tokens are v1.1.

**Connection URL import.** Pasting any of these fills the form: `postgres://` and `postgresql://` URLs, libpq `key=value` strings, `mysql://` and `mariadb://` URLs, and `.env` lines such as `DATABASE_URL=...`. Provider hosts (Supabase, Neon, RDS, PlanetScale, Cloud SQL, Azure) set a sensible TLS default.

**Connection behaviour.** Test button shows latency, server version and TLS details. Keep-alive every 60 s, auto-reconnect with backoff. Per-connection colour tag (local, development, staging, production), read-only mode, confirm-every-write mode, default schema, session time zone, statement timeout and startup SQL. Folders and search across saved connections.

## Functional requirements

Priorities: P0 = required for v1.0, P1 = v1.0 if time allows.

| ID | Area | Requirement | Priority |
| --- | --- | --- | --- |
| F-00 | Launcher | `wisp` starts the server on a random loopback port and opens the default browser; `--port`, `--no-browser`, `--app` (Chrome/Edge app window if installed), `--update`, `stop`; second launch reuses the running instance | P0 |
| F-01 | Connections screen | Cards with colour tag, engine badge, host, tags, last used; folders; search; duplicate; delete | P0 |
| F-02 | Connection form | Tabs General, SSH, SSL/TLS, Advanced; URL paste; Test; local file picker served by Go for keys and certificates | P0 |
| F-03 | Workspace tabs | Several connections and tables open at once; each browser tab can host one connection workspace | P0 |
| F-04 | Schema sidebar | Databases/schemas → tables, views, materialized views, functions, procedures, triggers, sequences, enums; lazy-loaded; filter | P0 |
| F-05 | Data grid | Virtualised rows (fixed DOM row pool) and columns; data fetched in windows from Go; resize, reorder, hide columns; NULL distinct; JSON/BLOB preview | P0 |
| F-06 | Filters and sort | Filter bar (=, !=, <, >, LIKE, ILIKE, IN, IS NULL, BETWEEN, raw WHERE); multi-column sort, all executed by the database | P0 |
| F-07 | Inline editing | Edit cells, add/duplicate/delete rows; change set held in Go; yellow/green/red markers; Review SQL then commit in one transaction | P0 |
| F-08 | Row detail panel | Selected row as a form, full values loaded on demand, foreign-key jump links | P0 |
| F-09 | SQL editor | CodeMirror 6 loaded on demand; highlighting; autocomplete from schema; run statement/selection/all; history; snippets | P0 |
| F-10 | Results | Result tabs, row count, timing, cancel, EXPLAIN / EXPLAIN ANALYZE tree | P0 |
| F-11 | Structure | View and edit columns, indexes, foreign keys; generated ALTER SQL reviewed before running | P0 |
| F-12 | Create/drop | New table wizard, truncate, drop with typed confirmation on production | P0 |
| F-13 | Import | CSV, JSON lines, SQL file chosen from a local path; streamed by Go; column mapping; progress, cancel | P1 |
| F-14 | Export | CSV, JSON, SQL INSERT, Markdown; streamed by Go straight to a browser download or a local path | P0 |
| F-15 | Command palette | Cmd/Ctrl+K: tables, actions, connections, snippets | P0 |
| F-16 | Server tools | Process list with cancel/kill, variables, users and grants (read-only) | P1 |
| F-17 | Theming | Light, dark, system; compact and comfortable density | P0 |
| F-18 | Settings | Settings screen + JSON file; keybinding overrides | P0 |
| F-19 | Session restore | Open tabs, SQL text and staged edits survive a crash or restart of the Go process | P0 |
| F-20 | Self-update | `wisp --update` downloads, verifies and replaces the binary | P1 |

## Local-only security model

Only the person who launched `wisp` can use it: the server binds to loopback, every request must carry a per-launch secret, and requests from other websites are rejected even though they run in the same browser.

| Threat | Control |
| --- | --- |
| Other machines on the network | Listen on `127.0.0.1` (and `::1`) only. There is no flag to bind elsewhere in v1.0. |
| Other local programs or users | A 256-bit random token is created at each launch and placed in the URL the launcher opens. The page exchanges it once for an `HttpOnly`, `SameSite=Strict` cookie and removes it from the address bar. Every API call without a valid cookie gets 401. |
| Malicious websites calling `localhost` (CSRF) | All state-changing calls require a custom header (`X-Wisp-CSRF`) holding a second per-session token; `Origin` must equal the server's own origin; no CORS headers are ever sent. |
| DNS rebinding | The `Host` header must be exactly `127.0.0.1:<port>` or `localhost:<port>`; anything else gets 421. |
| Script injection in the page | Strict Content-Security-Policy (`default-src 'self'`, no inline scripts, no `eval`), all text rendered with `textContent`, never `innerHTML` with data. |
| Secrets at rest | Passwords and passphrases in the OS keychain via `zalando/go-keyring` (pure Go); fallback encrypted file (Argon2id + XChaCha20-Poly1305) when no keychain exists. Config files hold references only. |
| Secrets in transit to the page | The API never returns stored passwords; forms show a placeholder and send a new value only when changed. |
| Leaks through logs | Passwords, URLs with credentials and parameter values are never logged. |
| Accidental writes to production | Server-side write guard on every statement; read-only sessions enforced in the database session too; typed confirmation for drops on production. |
| Stale processes | The server exits 10 minutes after the last tab disconnects (configurable), and `wisp stop` ends it immediately. |

## System architecture

Wisp is one Go binary with the built frontend embedded: Go owns connections, data, secrets and safety rules, and the browser tab only renders what Go sends.

```
Browser tab: Preact, TypeScript, Tailwind
  Connections | Schema sidebar | Data grid (40-row DOM pool) | SQL editor (CodeMirror, on demand)
        ^  JSON over HTTP, SSE events
        v
Go server on 127.0.0.1
  Security layer (token, CSRF, Host) | Sessions (write guard, retries) | Result pager (3 windows in RAM) | Data mover
        ^  Driver interface calls
        v
Drivers: one Driver interface
  PostgreSQL (jackc/pgx v5) | MySQL, MariaDB (go-sql-driver) | Introspection | Dialects
        ^  encrypted byte streams
        v
Transport and local storage
  TCP, socket, pipe | TLS (Go crypto/tls) | SSH tunnel (x/crypto/ssh) | Local store (keychain, JSON)
```

The result pager is the memory control for both sides: Go keeps at most three windows of rows per open result, and the tab keeps only what is on screen.

**Go packages**

| Package | Owns |
| --- | --- |
| `cmd/wisp` | CLI flags, launcher (opens browser), single-instance lock, `stop`, `--update` |
| `internal/server` | HTTP server, routing, security middleware, SSE hub, embedded static files |
| `internal/api` | Request/response types and handlers for each area |
| `internal/session` | Session lifecycle, keep-alive, reconnect, write guard, change sets, journal |
| `internal/driver` | `Driver` interface, `postgres`, `mysql` subpackages, introspection, dialects |
| `internal/pager` | Result windows, keyset/cursor/offset strategies, row counts |
| `internal/sqlgen` | Query builder, DML and DDL generation, statement splitter |
| `internal/transport` | TLS config, SSH tunnels and jump hosts, sockets and pipes |
| `internal/store` | Paths, settings, connection profiles, secrets, history, snippets |
| `internal/mover` | Streaming export and import |
| `internal/bench` | Memory and size scenarios used by CI gates |
| `web/` | HTML, JS modules, Tailwind CSS source, build output in `web/dist` |

**Dependencies (allow-list).** Go: `jackc/pgx/v5`, `go-sql-driver/mysql`, `golang.org/x/crypto` (ssh, argon2, chacha20poly1305), `golang.org/x/sys`, `zalando/go-keyring`, standard library for HTTP, TLS, JSON and `log/slog`. Page runtime: `preact`, `@preact/signals`, CodeMirror 6 modules (lazy). Build and test only: `vite`, `@preact/preset-vite`, `tailwindcss`, `@tailwindcss/vite`, `typescript`, `vitest`, `@testing-library/preact`, `happy-dom`, Playwright. Never `react`, `react-dom` or `preact/compat`. Anything else needs a written reason and its size impact in the pull request. CGO stays disabled so cross-compiling is one command.

**Memory techniques (Go).** `GOMEMLIMIT` set from settings (default 64 MiB soft limit), `GOGC=50` while results are open, `debug.FreeOSMemory()` after 15 s idle, row windows stored column-wise in typed slices with one byte arena per window, pooled buffers for JSON encoding, at most two database connections per session.

## Frontend architecture

The page is built with Preact (about 4 KB) and `@preact/signals` for state, written in strict TypeScript, bundled by Vite and styled with Tailwind CSS v4; the output is plain static files embedded in the Go binary, and the page holds only what is on screen.

| Concern | Choice | Why |
| --- | --- | --- |
| Framework | Preact 10 + `@preact/signals` (~6 KB gzip together); no `preact/compat` | React syntax agents know well, at a seventh of React's size; no React libraries can slip in |
| Language | TypeScript (`.ts`, `.tsx`), `strict` + `noUncheckedIndexedAccess`, `tsc --noEmit` in CI | Catches agent mistakes before runtime |
| Build | Vite with `@preact/preset-vite` and `@tailwindcss/vite`; hashed assets, code splitting; output embedded with `go:embed` | Instant hot reload in development; small static output |
| Styling | Tailwind CSS v4, CSS-first `@theme` holding the mockup's tokens, dark mode via `data-theme` | Tiny purged CSS; tokens in one place; utility classes are easy for agents |
| Components | Own small kit of Preact components: button, input, select, segmented, tabs, modal (native dialog), toast, menu, virtual list | A few KB instead of a component library |
| State | Signals in `web/src/state.ts`; local component state with signals or hooks | No state library; fine-grained updates |
| Grid | Preact wrapper + plain TypeScript row pool that writes cells directly (visible rows + 10 reused while scrolling); columns virtualised | No virtual-DOM diff per scroll frame; DOM size independent of table size |
| Data | Fetched from Go in windows of ~150 rows in a columnar JSON format; at most 3 windows kept | 1M-row tables cost the tab the same as 50 rows |
| SQL editor | CodeMirror 6 (`@codemirror/lang-sql`), dynamically imported on first query tab, destroyed on unmount | Only paid for when used |
| Tests | Vitest + `@testing-library/preact` for components; Playwright for flows and memory gates | Tests through public behaviour, matching the TDD rules |
| Fonts and icons | System font stacks; inline SVG icons | No font or icon-font downloads |

**Memory rules for every screen**

- Every effect that adds listeners, timers or fetches returns a cleanup; fetches use `AbortController`.
- Background workspace tabs keep only their state (scroll position, filters); their components unmount and remount on return.
- No data caches beyond the 3-window grid buffer and small metadata lists; everything else is asked from Go again.
- Large values (over 4 KB) are fetched only when the detail panel opens and released when selection changes.
- A dev-only overlay shows JS heap and DOM node count live, so regressions are seen while building.

**Tailwind setup.** `web/src/styles/app.css` imports Tailwind and declares an `@theme` block with every colour, font size, radius and spacing value from the mockup as tokens (for example `--color-env-prod`, `--color-ink-2`, `--radius-card`). Dark values are set under `[data-theme=dark]` and a custom `dark` variant. Components use token utilities only (`bg-panel`, `text-ink-2`, `border-line`); arbitrary hex values are rejected by a CI check.

## API between browser and Go

The page talks to Go through a small JSON-over-HTTP API under `/api/v1`, plus one Server-Sent Events stream for push updates; there is no WebSocket and no GraphQL.

| Area | Endpoints (examples) |
| --- | --- |
| Bootstrap | `GET /auth/exchange?t=` (one-time token → cookies, then redirect), `GET /api/v1/hello` (version, settings, CSRF token) |
| Events | `GET /api/v1/events` (SSE: session state, query progress, export progress, server shutdown, heartbeat every 15 s) |
| Settings | `GET/PUT /api/v1/settings` |
| Connections | `GET/POST /api/v1/connections`, `PUT/DELETE /api/v1/connections/{id}`, `POST /api/v1/connections/test`, `POST /api/v1/connections/parse-url` |
| Sessions | `POST /api/v1/sessions` (open), `DELETE /api/v1/sessions/{sid}`, `GET /api/v1/sessions/{sid}/info` |
| Schema | `GET /api/v1/sessions/{sid}/schemas`, `.../objects?schema=&kind=`, `.../tables/{schema}/{table}` (columns, keys, indexes) |
| Results | `POST .../results` (open a browse or query result → `rid`), `GET .../results/{rid}/window?offset=&limit=`, `GET .../results/{rid}/value?row=&col=`, `DELETE .../results/{rid}` |
| Edits | `POST .../results/{rid}/changes` (stage), `GET .../changes/sql` (review), `POST .../changes/commit`, `DELETE .../changes` |
| Queries | `POST .../run` (script → statement results), `POST .../cancel/{qid}`, `POST .../explain` |
| Files | `GET /api/v1/fs/list?path=` (local file picker, home-relative), export `GET .../export?...` (streamed download), import `POST .../import` (local path) |

**Window wire format.** Rows travel column-wise to avoid one object per cell:

```json
{"rid":"r7","offset":300,"total":{"kind":"estimate","value":1248920},
 "columns":[{"name":"id","type":"int8"},{"name":"status","type":"order_status"}],
 "data":[[1248920,1248919],["paid","shipped"]],
 "nulls":[[],[1]], "truncated":[[],[]], "overlay":{"mod":[[0,1]],"ins":[],"del":[]}}
```

**Rules.** Every response is JSON with `{"error":{"code","message","detail"}}` on failure. Long operations return an id immediately and report progress over SSE. Requests are cancellable from the page with `AbortController`, which Go maps to context cancellation all the way to the database.

## UX and browser concerns

The screens follow the existing HTML mockup exactly (connections, connection form, table data with sidebar and detail panel, SQL editor, structure, command palette, review SQL); the browser adds four constraints the design must handle.

| Concern | Decision |
| --- | --- |
| Shortcuts the browser owns (Cmd/Ctrl+W, T, N, Tab) | App uses Alt-based alternatives: Alt+W close tab, Alt+T new query, Alt+1..9 switch tab. Cmd/Ctrl+K, Cmd/Ctrl+S, Cmd/Ctrl+Enter, Esc work as in the mockup. `--app` mode opens a Chrome/Edge app window without browser tabs. |
| File paths (SSH keys, certificates, import files) | Browsers hide real paths, so Go serves a small file picker listing folders under the user's home; the chosen path is used by Go directly. |
| Downloads | Exports stream from Go as a normal browser download (`Content-Disposition`), or to a chosen local path for very large files. |
| Losing the page | Closing or reloading the tab loses nothing: staged edits, open tabs and SQL text live in Go and are restored when the page reconnects. A banner shows if the Go process stopped. |

**Key screens**

| Screen | Key interactions |
| --- | --- |
| Connections | Double-click or Enter to connect; paste a URL anywhere to create; Alt+N new connection |
| Connection form | Paste URL to fill; Test; tabs General / SSH / SSL / Advanced |
| Workspace – data | Inline edit, Cmd/Ctrl+S review and commit, Cmd/Ctrl+F filter, Space row detail |
| Workspace – SQL | Cmd/Ctrl+Enter run statement, Shift+Cmd/Ctrl+Enter run all, Esc cancel |
| Structure | Edit then Review SQL |
| Command palette | Cmd/Ctrl+K anywhere |

**Design rules.** 26 px rows in compact mode, monospaced cells, 13 px UI text. Colour only for meaning: environment, staged edit state, errors. Status bar always shows environment, server version, latency, pending changes, Go memory and tab memory. Every destructive action shows its SQL first.

## Release plan

About 20 weeks for 2 engineers working with AI agents, in five milestones; a milestone ends only when both budget tables still pass. The companion agent ticket pack holds one ticket and one ready-to-paste prompt per step. Every coding ticket is built test-first with the TDD skill: the agent proposes the test seams, a human confirms them, then the agent works in red → green slices.

1. **M0 – Foundations (weeks 1–3).** Repo and toolchain, launcher and local server with security layer, frontend build with Tailwind and embedding, CI with budget gates for Go and the tab, design tokens, component kit, app shell, settings, secrets, driver interface, PostgreSQL and MySQL drivers.
2. **M1 – Connect and browse (weeks 4–8).** Connection profiles, connections screen, form, URL import, TLS, SSH and jump hosts, sessions with write guard, introspection, sidebar, result pager, virtualised grid, filters and sort, workspace tabs.
3. **M2 – Edit and query (weeks 9–13).** Row detail panel, staged edits held in Go, review and commit, CodeMirror SQL editor, autocomplete, query runner with results and EXPLAIN, history, snippets and session restore.
4. **M3 – Structure and data movement (weeks 14–17).** Structure viewer and editor, create/drop safeguards, streaming export and import, command palette, shortcuts and settings screen.
5. **M4 – Release (weeks 18–20).** Sockets and named pipes, server tools, browser compatibility and accessibility pass, packaging with self-update, end-to-end test matrix, performance audit, docs and 1.0.

v1.1 candidates: cloud IAM auth, connection import from other tools, ER diagram viewer, SQLite driver.

## Success metrics, risks and open questions

**Success metrics (90 days after 1.0):** 100% of releases within every deal-breaker budget; median Go idle RSS under 25 MB on user machines (opt-in report); launch to first query result under 5 s (p50); crash-free sessions above 99.5%; 5,000 weekly active users.

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Whole-tab memory higher than estimated in some browsers | Tab budget breach | Week-1 prototype measured in Chrome, Firefox, Safari; adjust only the whole-tab row, keep JS/DOM rows strict |
| Agents add libraries or frameworks casually | Bundle and heap growth | Bundle-size and heap gates in CI; AGENTS.md bans React, preact/compat and UI libraries; dependency review per PR |
| Local server attacked by websites | Data exposure | Token cookie, CSRF header, Origin and Host checks, CSP; dedicated security tests in CI |
| Go GC lets RSS drift up | Go budget breach | `GOMEMLIMIT`, tuned `GOGC`, `debug.FreeOSMemory` after idle, 8-hour soak test |
| Browser shortcut conflicts annoy users | Poor feel | Alt-based shortcuts, `--app` window mode, shortcut cheat sheet |
| MySQL auth plugin and TLS variety | Connection failures | Docker matrix of every supported version in CI |

**Open questions**

- Licence and business model: open source core with paid extras, or fully commercial?
- Trademark, domain and package-name check for Wisp (GitHub, Homebrew, Scoop).
- Should v1.0 offer an opt-in `--app` default when Chrome or Edge is installed?
- Is a minimal SQLite driver worth adding to v1.0, given how cheap it is in Go?
