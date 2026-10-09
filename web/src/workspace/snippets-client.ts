import { request } from "../api";

export type Snippet = {
  id: string;
  name: string;
  tags?: string[];
  sql: string;
};

export type SnippetsDocument = {
  version: number;
  snippets: Snippet[];
};

export async function fetchSnippets(): Promise<SnippetsDocument> {
  const doc = await request<SnippetsDocument>("GET", "/api/v1/snippets");
  return { ...doc, snippets: doc.snippets ?? [] };
}

export async function saveSnippet(snippet: Omit<Snippet, "id"> & { id?: string }): Promise<Snippet> {
  return request<Snippet>("POST", "/api/v1/snippets", snippet);
}
