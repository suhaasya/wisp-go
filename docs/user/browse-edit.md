# Browse and edit data

## Opening a table

From the sidebar, select a table. Wisp loads a **window** of rows (keyset pagination), not the full table. Scroll to fetch more; only a few windows stay in memory at once.

## Grid

- Sort and filter using column headers where supported.
- Click a cell to edit (when the connection is writable and the column is editable).
- Primary keys (or unique keys) identify rows for updates.

## Change sets (safe edits)

Edits are **staged** in a change set until you **Commit** or **Discard**:

1. Change cells or add/delete rows in the grid.
2. Review the change set panel.
3. **Commit** runs DML inside a transaction (dialect-specific).
4. **Discard** drops staged changes.

This avoids accidental writes, especially on production-tagged connections.

## Read-only

Read-only profiles or session mode block commits; the UI shows edits as disabled or prompts you to switch mode where allowed.

## Large tables

Scrolling through millions of rows stays bounded: the Go process streams pages and the grid reuses DOM nodes. If counts are expensive, Wisp may show approximate counts depending on dialect and settings.
