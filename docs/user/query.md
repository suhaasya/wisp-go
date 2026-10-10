# Query editor

## Running SQL

Open the **Query** workspace for the active connection. Type SQL in the editor (lazy-loaded CodeMirror chunk). Run the current statement or selection with the run control or keyboard shortcut (see [shortcuts.md](shortcuts.md)).

Every statement from the UI passes through Wisp’s **write guard**: destructive or DDL operations may require confirmation based on **Data safety** settings.

## Parameters

Use bound parameters where the editor supports them; values are never interpolated into SQL as raw strings.

## Results

- **SELECT** results appear in a result grid with the same windowed scrolling model as table browse.
- Multiple result sets (if the driver returns them) appear as separate tabs or sections.
- Errors show code, message, and detail from the server without leaking secrets.

## History and snippets

Recent queries and saved snippets are stored locally under your Wisp config directory.

## Explain and metadata

Where supported, use structure tools or catalog queries from the sidebar; heavy DDL is covered in [structure.md](structure.md).
