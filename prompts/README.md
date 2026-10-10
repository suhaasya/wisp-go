# Ticket order

| Step | Ticket | Title | Depends on |
|---|---|---|---|
| 1 | [LW-001](LW-001.md) | Repository, Go module and tooling | - |
| 2 | [LW-002](LW-002.md) | Launcher: CLI, single instance, browser open, shutdown | LW-001 |
| 3 | [LW-003](LW-003.md) | Local HTTP server and security layer | LW-002 |
| 4 | [LW-004](LW-004.md) | Frontend build: Vite, Preact, TypeScript, Tailwind v4, embedding | LW-001 |
| 5 | [LW-005](LW-005.md) | Budget gates: Go size and RAM, tab memory, bundle sizes | LW-003, LW-004 |
| 6 | [LW-006](LW-006.md) | CI pipeline and release matrix | LW-001 |
| 7 | [LW-007](LW-007.md) | Design tokens in Tailwind theme, light and dark | LW-004 |
| 8 | [LW-008](LW-008.md) | UI component kit (Preact + TypeScript) and gallery | LW-007 |
| 9 | [LW-009](LW-009.md) | App shell: router, API client, SSE events, state signals, title and status bars | LW-003, LW-008 |
| 10 | [LW-010](LW-010.md) | Settings store and API | LW-009 |
| 11 | [LW-011](LW-011.md) | Secret storage: OS keychain and encrypted fallback | LW-001 |
| 12 | [LW-012](LW-012.md) | Driver interface, compact window format and mock driver | LW-001 |
| 13 | [LW-013](LW-013.md) | PostgreSQL driver (pgx) | LW-012 |
| 14 | [LW-014](LW-014.md) | MySQL and MariaDB driver | LW-012 |
| 15 | [LW-015](LW-015.md) | Connection profiles store and API | LW-010, LW-011 |
| 16 | [LW-016](LW-016.md) | Connections screen | LW-015, LW-009 |
| 17 | [LW-017](LW-017.md) | Connection form, Test button and local file picker | LW-016, LW-013, LW-014 |
| 18 | [LW-018](LW-018.md) | Connection URL parser and paste-to-create | LW-017 |
| 19 | [LW-019](LW-019.md) | TLS support | LW-013, LW-014 |
| 20 | [LW-020](LW-020.md) | SSH tunnels and jump hosts | LW-019, LW-017 |
| 21 | [LW-021](LW-021.md) | Session manager: lifecycle, keep-alive, write guard | LW-020 |
| 22 | [LW-022](LW-022.md) | Schema introspection and metadata cache | LW-021 |
| 23 | [LW-023](LW-023.md) | Schema sidebar | LW-022, LW-008 |
| 24 | [LW-024](LW-024.md) | Result pager in Go and window API | LW-021, LW-012 |
| 25 | [LW-025](LW-025.md) | Virtualised data grid (row pool) | LW-024, LW-008 |
| 26 | [LW-026](LW-026.md) | Filter bar and server-side sorting | LW-025 |
| 27 | [LW-027](LW-027.md) | Workspace tabs and state persistence | LW-025 |
| 28 | [LW-028](LW-028.md) | Row detail panel with foreign-key navigation | LW-025, LW-022 |
| 29 | [LW-029](LW-029.md) | Inline editing with change sets held in Go | LW-028, LW-021 |
| 30 | [LW-030](LW-030.md) | Review SQL and transactional commit | LW-029 |
| 31 | [LW-031](LW-031.md) | SQL editor with CodeMirror 6 (lazy loaded) | LW-027 |
| 32 | [LW-032](LW-032.md) | Autocomplete from schema metadata | LW-031, LW-022 |
| 33 | [LW-033](LW-033.md) | Query runner, result tabs, cancel and EXPLAIN | LW-031, LW-024 |
| 34 | [LW-034](LW-034.md) | History, snippets and session restore | LW-033, LW-029 |
| 35 | [LW-035](LW-035.md) | Structure viewer | LW-022, LW-027 |
| 36 | [LW-036](LW-036.md) | Structure editor with ALTER generation | LW-035, LW-030 |
| 37 | [LW-037](LW-037.md) | Create, drop and truncate safeguards | LW-036 |
| 38 | [LW-038](LW-038.md) | Streaming export | LW-024 |
| 39 | [LW-039](LW-039.md) | Streaming import | LW-038, LW-017 |
| 40 | [LW-040](LW-040.md) | Command palette | LW-023, LW-008 |
| 41 | [LW-041](LW-041.md) | Shortcuts and settings screen | LW-040, LW-010 |
| 42 | [LW-042](LW-042.md) | Sockets, named pipes and local database detection | LW-013, LW-014, LW-017 |
| 43 | [LW-043](LW-043.md) | Server tools: processes, variables, users | LW-025 |
| 44 | [LW-044](LW-044.md) | Browser compatibility and accessibility pass | LW-041 |
| 45 | [LW-045](LW-045.md) | Packaging, install script and self-update | LW-006 |
| 46 | [LW-046](LW-046.md) | End-to-end test matrix | LW-044, LW-020 |
| 47 | [LW-047](LW-047.md) | Performance and memory audit | LW-046 |
| 48 | [LW-048](LW-048.md) | Documentation, beta and 1.0 release | LW-045, LW-047 |
| — | [LW-049](LW-049.md) | Release finish checklist (manual backlog) | LW-048 |
