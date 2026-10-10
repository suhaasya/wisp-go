# Import and export

Wisp moves data in **streams** so large tables do not load entirely into memory.

## Export

From a table or query result, use export actions to write CSV (or supported formats) to a path you choose. Export runs on the Go side reading chunks from the database.

## Import

Import maps file columns to table columns and inserts in batches. Review target table and options before starting; long imports honour cancellation.

## Permissions

Import and export require appropriate database privileges (`SELECT`, `INSERT`, file access on your machine). Wisp does not bypass server-side restrictions.

## Tips

- Prefer keyset-friendly tables for huge exports (indexed sort columns).
- For one-off bulk loads, consider native server tools (`COPY`, `LOAD DATA`) alongside Wisp for maximum speed; Wisp focuses on safe, bounded transfers.
