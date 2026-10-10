# First run

1. Start Wisp (`wisp`). The browser opens to the **Connections** screen.
2. Click **New connection** and pick PostgreSQL, MySQL, or MariaDB.
3. Fill host, port, user, and database. Use **Test** before saving.
4. **Connect**. The sidebar shows schemas and tables for that connection.
5. Open a table to browse rows, or open the **Query** tab to run SQL.

## Settings

Open **Settings** from the app chrome. Useful defaults:

- **Appearance** — theme and density
- **Data safety** — confirm destructive SQL, production colour tags
- **Privacy** — idle exit and Go memory limit (optional)
- **About & feedback** — version and beta feedback link

## Data location

Profiles, history, and snippets live under Wisp’s config directory (platform-specific; see [security.md](security.md)). Passwords are stored via the OS secret service when available.

## Single instance

Only one Wisp server runs per user session. A second `wisp` command attaches to the existing server and opens another browser tab.
