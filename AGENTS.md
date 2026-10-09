# AGENTS.md — rules for AI agents working on Wisp

Wisp is a LOCAL database client for PostgreSQL, MySQL and MariaDB.
One Go binary runs a private HTTP server on 127.0.0.1 and the UI runs in the
user's browser (Preact + TypeScript, built by Vite, styled with Tailwind CSS v4).
Read this file
before every task. It overrides your defaults.

## 1. Deal breakers (CI enforces these; a PR that breaks one is rejected)
Go process                                   Target    Hard cap
  Binary size (assets embedded)              15 MB     25 MB
  Download size (compressed)                  7 MB     10 MB
  RSS idle, one tab attached                 20 MB     30 MB
  RSS, 1 connection + 500-table schema       30 MB     45 MB
  RSS per extra connection                    2 MB      4 MB
  RSS, 1M-row table scrolled end to end      50 MB     80 MB
  Start to page served                      100 ms    300 ms
Browser tab                                  Target    Hard cap
  JS heap idle                                5 MB     10 MB
  JS heap, 1M-row table scrolled             10 MB     20 MB
  DOM nodes at any time                      2,000     4,000
  Initial JS (gzip)                          60 KB    100 KB
  SQL editor chunk (gzip, lazy)             200 KB    280 KB
  CSS (gzip)                                 15 KB     25 KB
  Whole tab (Chrome Task Manager)            40 MB     70 MB
  Long tasks while scrolling                 <50 ms   none >100 ms
Run `make size bench gate` and report the numbers.

## 2. Repository layout
cmd/wisp/            CLI, launcher, single instance, stop, --update, bench mode
internal/server/      HTTP server, security middleware, SSE hub, static files
internal/api/         handlers and request/response types (one file per area)
internal/session/     sessions, keep-alive, write guard, change sets, journal
internal/driver/      Driver interface, Window type, dialects; postgres/, mysql/
internal/pager/       result windows, keyset/cursor/offset, counts, overlays
internal/sqlgen/      browse queries, DML, DDL, ALTER, statement splitter
internal/transport/   TLS, SSH tunnels and jump hosts, sockets and pipes
internal/store/       paths, settings, profiles, secrets, history, snippets
internal/mover/       streaming export and import
internal/bench/       budgets and RSS measurement used by CI
web/                  index.html, vite.config.ts, src/ (TS/TSX), src/styles/app.css
tools/                tabbench (Playwright), gate, check scripts
.claude/skills/tdd/   the TDD skill every coding task follows (section 6)
GLOSSARY.md           domain words; use them in test and function names
docs/PRD.md           product requirements (source of truth for behaviour)
design/               HTML mockup + screenshots of every screen (light/dark)

Import rules: api -> session -> pager/sqlgen -> driver -> transport;
store is used by api and session; server uses api. No cycles. CGO stays OFF.

## 3. Go rules
- Allowed dependencies: jackc/pgx/v5, go-sql-driver/mysql, golang.org/x/crypto,
  golang.org/x/sys, zalando/go-keyring (+ go-winio on Windows for named pipes).
  Anything else needs a reason and its size impact in your report.
- Every function doing I/O takes context.Context and honours cancellation.
- Results are Windows (column-wise vectors + one byte arena). Never build
  []map[string]any or one struct per cell.
- Errors returned to the page as {"error":{"code","message","detail"}}.
- Never log passwords, tokens, URLs with credentials or parameter values.
- Every SQL statement from the page goes through session.Guard().
- Values are always bound parameters; identifiers go through Dialect.QuoteIdent.
- gofmt, go vet, golangci-lint clean. Table-driven tests.

## 4. Frontend rules
- Preact + TypeScript (.ts / .tsx), built by Vite. `npm run check` (tsc) clean.
- Runtime packages allowed: preact, @preact/signals, CodeMirror 6 modules
  (lazy loaded). Nothing else without approval.
- NEVER install react, react-dom or preact/compat, and no UI/component/icon
  libraries, state libraries, CSS-in-JS or date libraries.
- Function components with typed props. Shared state = signals in
  web/src/state.ts. Every effect that adds listeners, timers or fetches
  returns a cleanup (AbortController for fetches).
- Render data as JSX text (escaped). Never use dangerouslySetInnerHTML.
  No inline <script> or style="" attributes (strict CSP).
- The data grid's row pool (web/src/grid/pool.ts) is plain TypeScript that
  writes the DOM directly; do not convert it to JSX.
- Data lives in Go. The page keeps only what is on screen (grid: max 3 windows).
- Styling: Tailwind v4 utilities using theme tokens only (bg-panel, text-ink-2,
  border-line, bg-env-prod ...). No arbitrary colour values like bg-[#123456].
- Use the existing components in web/src/ui/ before writing new markup.
- Browser-owned shortcuts (Ctrl/Cmd+W, T, N, L, R, Tab) are never rebound.

## 5. UI work and the mockup
- design/wisp-design-mockup.html is the visual reference. Use the screenshot
  named in the task for layout, the mockup CSS for exact sizes, and the
  Tailwind tokens (web/src/styles/app.css) for colours.
- The mockup's macOS window frame (traffic lights) is decoration; the real app
  is a web page and has no window controls.
- The mockup's browser prompt() and fake data are stand-ins; build real flows.
- Check light and dark themes.

## 6. Tests: test-driven development (follow .claude/skills/tdd/SKILL.md)
- Red before green: write one failing test, then only the code to pass it.
- Vertical slices: one seam -> one test -> minimal code -> repeat. Never
  write all tests first.
- Seams first: before writing ANY test, list the seams (public interfaces)
  you will test, each with one line on what it catches and what it misses,
  and wait for the human to confirm them. The task's "Tests" section is the
  starting proposal.
- Test behaviour through public interfaces (HTTP handlers, exported Go
  functions, component props and rendered output). Do not test private
  functions or mock internal collaborators; fakes only at system boundaries
  (database, network, clock, filesystem).
- Expected values come from an independent source (spec, literal, worked
  example), never recomputed the way the code computes them.
- No refactoring inside the loop; list refactor ideas in the report.
- Go: tests next to code; DB tests use build tag `integration` and
  ci/docker-compose.yml. Web: Vitest + @testing-library/preact; Playwright
  for flows and memory gates (web/ and e2e/).
- Tickets with no code under test (CI config, docs, packaging) say so in the
  plan and skip the seams step.

## 7. Definition of done
1. All acceptance criteria pass.
2. make lint and make test pass; make gate passes (from LW-005 on).
3. No work outside the task's scope.
4. Final report written (format below).

## 8. Final report format
### Summary
### Files changed
### Seams and tests       (confirmed seams; tests written per seam, in order)
### Acceptance criteria   (each PASS/FAIL + evidence)
### Measurements          (binary MB, Go RSS, JS heap, DOM nodes, bundle KB)
### Dependencies added    (name, version, why, size impact)
### Refactor ideas       (left for review, not done)
### Open questions / follow-ups

## 9. When something is unclear
Stop and ask, or put the question in the report. Do not invent product
behaviour that is not in the task or docs/PRD.md.
