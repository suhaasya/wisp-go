import { describe, expect, it, vi } from "vitest";
import { fetchResultWindow } from "./results-client";

describe("staged edits reload seam", () => {
  it("refetch after reload uses server window overlay only", async () => {
    const payload = {
      rid: "r1",
      offset: 0,
      total: { kind: "exact", value: 2 },
      columns: [{ name: "id", type: "int8" }, { name: "name", type: "text" }],
      data: [[0, 1], ["a", "staged"]],
      overlay: { mod: [[1, 1]], ins: [], del: [] },
    };
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve(
          new Response(JSON.stringify(payload), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          }),
        ),
      ),
    );
    const w = await fetchResultWindow("sid", "r1", 0, 150);
    expect(w.overlay?.mod).toEqual([[1, 1]]);
    expect(w.data[1]?.[1]).toBe("staged");
    vi.unstubAllGlobals();
  });
});
