/** Workspace registers runtime handlers for palette item runs. */
export type PaletteHandlers = {
  openTable?: (schema: string, table: string) => void;
  openSnippet?: (snippetId: string) => void;
};

export const paletteHandlers: PaletteHandlers = {};
