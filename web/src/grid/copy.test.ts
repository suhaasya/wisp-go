import { describe, expect, it, vi } from "vitest";
import { buildColumns } from "./columns";
import { formatSelectionSync } from "./copy";
import { selectCell } from "./selection";

describe("formatSelectionSync", () => {
  it("formats a cell range as TSV", () => {
    const cols = buildColumns([
      { name: "a", type: "text" },
      { name: "b", type: "int8" },
    ]);
    const anchor = selectCell(1, 0);
    const extended = {
      kind: "cell" as const,
      anchor: anchor.kind === "cell" ? anchor.anchor : { row: 1, col: 0 },
      focus: { row: 2, col: 1 },
    };
    const text = formatSelectionSync(
      extended,
      cols,
      (row, col) => `r${row}c${col}`,
      "tsv",
      "t",
    );
    expect(text).toBe("r1c0\tr1c1\nr2c0\tr2c1");
  });
});

describe("large copy", () => {
  it("yields between chunks for large row selections", async () => {
    vi.useFakeTimers();
    const { copySelection } = await import("./copy");
    const { createWindowBuffer } = await import("./buffer");
    const cols = buildColumns([{ name: "id", type: "int8" }]);
    const write = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", { clipboard: { writeText: write } });

    const buffer = createWindowBuffer({
      windowRows: 150,
      maxWindows: 3,
      fetch: async (offset) => ({
        rid: "r",
        offset,
        total: { kind: "exact", value: 20_000 },
        columns: [{ name: "id", type: "int8" }],
        data: [Array.from({ length: 150 }, (_, i) => offset + i)],
      }),
    });

    const state = {
      kind: "cell" as const,
      anchor: { row: 0, col: 0 },
      focus: { row: 9999, col: 0 },
    };

    const p = copySelection(state, cols, buffer, "tsv", "t", (row) => String(row));
    await vi.runAllTimersAsync();
    await p;
    expect(write).toHaveBeenCalled();
    const arg = write.mock.calls[0]![0] as string;
    expect(arg.split("\n").length).toBe(10_000);
    buffer.destroy();
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });
});
