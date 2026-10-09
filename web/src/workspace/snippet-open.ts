import { fetchSnippets } from "./snippets-client";
import { openQueryTabWithText } from "./tabs";

export async function openWorkspaceSnippet(snippetId: string): Promise<void> {
  const doc = await fetchSnippets();
  const snippet = doc.snippets.find((s) => s.id === snippetId);
  if (snippet) {
    openQueryTabWithText(snippet.sql);
  }
}
