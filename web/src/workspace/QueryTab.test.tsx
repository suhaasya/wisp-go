import { render, waitFor } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { QueryTab } from "./QueryTab";

const { destroy, mountSqlEditor, setCompletionSchema, fetchSnippets } = vi.hoisted(() => {
  const destroy = vi.fn();
  const setCompletionSchema = vi.fn();
  const setRunErrorAt = vi.fn();
  const clearRunError = vi.fn();
  const fetchSnippets = vi.fn(() => Promise.resolve({ version: 1, snippets: [] }));
  const mountSqlEditor = vi.fn(() => ({
    view: {},
    destroy,
    setCompletionSchema,
    setRunErrorAt,
    clearRunError,
    getSelectionRange: () => ({ start: 0, end: 0 }),
    getCursor: () => 0,
    insertText: vi.fn(),
  }));
  return { destroy, mountSqlEditor, setCompletionSchema, fetchSnippets };
});

vi.mock("../editor/index.ts", () => ({
  mountSqlEditor,
}));

vi.mock("./completion-client.ts", () => ({
  fetchCompletionSchema: vi.fn(() =>
    Promise.resolve({ schema: "public", tables: [] }),
  ),
}));

vi.mock("./snippets-client.ts", () => ({
  fetchSnippets,
  saveSnippet: vi.fn(() => Promise.resolve({ id: "s1", name: "n", sql: "x" })),
}));

describe("QueryTab", () => {
  it("lazy-loads the editor and destroys on unmount", async () => {
    mountSqlEditor.mockClear();
    destroy.mockClear();
    const { unmount } = render(
      <QueryTab
        sessionId="s1"
        tab={{ id: "q1", kind: "query", title: "Query", queryText: "SELECT 1" }}
      />,
    );
    await waitFor(() => {
      expect(mountSqlEditor).toHaveBeenCalledTimes(1);
    });
    const firstCall = mountSqlEditor.mock.calls[0] as [{ value: string }] | undefined;
    expect(firstCall?.[0]?.value).toBe("SELECT 1");
    unmount();
    await waitFor(() => {
      expect(destroy).toHaveBeenCalledTimes(1);
    });
  });

  it("mounts the editor when snippets is null in the response", async () => {
    mountSqlEditor.mockClear();
    fetchSnippets.mockResolvedValueOnce({ version: 1, snippets: null as unknown as [] });
    render(
      <QueryTab
        sessionId="s1"
        tab={{ id: "q3", kind: "query", title: "Query", queryText: "" }}
      />,
    );
    await waitFor(() => {
      expect(mountSqlEditor).toHaveBeenCalledTimes(1);
    });
    const call = mountSqlEditor.mock.calls[0] as [{ snippets: unknown[] }] | undefined;
    expect(call?.[0]?.snippets).toEqual([]);
  });

  it("mounts the editor when snippets fetch fails", async () => {
    mountSqlEditor.mockClear();
    fetchSnippets.mockRejectedValueOnce(new Error("offline"));
    render(
      <QueryTab
        sessionId="s1"
        tab={{ id: "q2", kind: "query", title: "Query", queryText: "" }}
      />,
    );
    await waitFor(() => {
      expect(mountSqlEditor).toHaveBeenCalledTimes(1);
    });
    fetchSnippets.mockResolvedValue({ version: 1, snippets: [] });
  });
});
