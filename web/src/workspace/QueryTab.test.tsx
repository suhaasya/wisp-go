import { render, waitFor } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { QueryTab } from "./QueryTab";

const { destroy, mountSqlEditor, setCompletionSchema } = vi.hoisted(() => {
  const destroy = vi.fn();
  const setCompletionSchema = vi.fn();
  const mountSqlEditor = vi.fn(() => ({ view: {}, destroy, setCompletionSchema }));
  return { destroy, mountSqlEditor, setCompletionSchema };
});

vi.mock("../editor/index.ts", () => ({
  mountSqlEditor,
}));

vi.mock("./completion-client.ts", () => ({
  fetchCompletionSchema: vi.fn(() =>
    Promise.resolve({ schema: "public", tables: [] }),
  ),
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
});
