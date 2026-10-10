# Structure (schema tools)

## Sidebar

The connection sidebar lists schemas, tables, views, and related objects depending on dialect. Expand nodes to inspect columns, indexes, and constraints where implemented.

## DDL and ALTER

Structure changes go through generated DDL with quoted identifiers. Review SQL before apply; destructive operations respect **Data safety** confirmations.

## Differences by engine

- **PostgreSQL:** schemas, roles, extensions as exposed in the catalog browser.
- **MySQL/MariaDB:** databases and tables; engine-specific metadata follows `information_schema` / `SHOW` patterns.

If an action is not yet available in the UI, use the query editor with carefully reviewed DDL.
