# Glossary — Wisp

Use these words in code, test names and messages. Do not invent synonyms.

| Term | Meaning |
|---|---|
| Profile | A saved connection definition (host, engine, SSH, TLS, environment). No secrets inside. |
| Secret | A password, SSH passphrase or token. Lives only in the keychain or encrypted file. |
| Environment | The tag on a profile: local, development, staging or production. Drives colours and safety. |
| Session | A live connection opened from a profile: main connection + metadata connection + tunnel. |
| Metadata connection | The second database connection a session uses for schema queries and keep-alive. |
| Write guard | The check every statement passes before reaching the database (read-only, confirm-every-write). |
| Read-only session | Session where writes are blocked by the write guard and by the database session setting. |
| Safe mode | "Confirm every write": writes need explicit confirmation from the user. |
| Result | An open query or table browse identified by a rid; owned by the pager. |
| Window | A block of rows (default 150) stored column-wise with one byte arena; the unit sent to the page. |
| Pager | Go component that keeps at most three windows per result and fetches others on request. |
| Strategy | How a pager reads rows: keyset, cursor or offset. |
| Overlay | Staged-edit markers (modified cells, inserted rows, deleted rows) added to a window. |
| Change set | Staged edits for one result, held in Go until commit or discard. |
| Review SQL | The generated SQL shown to the user before a commit. |
| Commit | Running a change set's SQL in one transaction. |
| Journal | On-disk copy of workspace state and change sets used for session restore. |
| Workspace | The screen for one session: sidebar, tabs, status bar. |
| Tab | A workspace tab: table data, structure, query or server tools. |
| Row pool | The fixed set of grid row elements reused while scrolling. |
| Launch token | One-time secret in the URL the launcher opens; exchanged for the session cookie. |
| Seam | A public interface where tests observe behaviour (see .claude/skills/tdd). |
| Gate | The CI check that fails a change exceeding a budget's hard cap. |
