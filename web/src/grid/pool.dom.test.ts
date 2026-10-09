import { describe, expect, it } from "vitest";
import { createWindowBuffer } from "./buffer";
import { buildColumns } from "./columns";
import { createRowPool } from "./pool";
import { emptySelection } from "./selection";
import type { WindowPayload } from "./window";

describe("createRowPool", () => {
  it("reuses a bounded number of row elements while scrolling", () => {
    const scroll = document.createElement("div");
    Object.defineProperty(scroll, "clientHeight", { value: 300, configurable: true });
    Object.defineProperty(scroll, "clientWidth", { value: 800, configurable: true });
    scroll.style.height = "300px";
    scroll.style.overflow = "auto";
    document.body.appendChild(scroll);

    const header = document.createElement("div");
    header.style.height = "40px";
    const inner = document.createElement("div");
    const spacer = document.createElement("div");
    const body = document.createElement("div");
    inner.appendChild(header);
    inner.appendChild(spacer);
    inner.appendChild(body);
    scroll.appendChild(inner);

    const buffer = createWindowBuffer({
      windowRows: 150,
      maxWindows: 3,
      fetch: async (offset) =>
        ({
          rid: "r",
          offset,
          total: { kind: "exact", value: 100_000 },
          columns: [{ name: "id", type: "int8" }],
          data: [[offset, offset + 1]],
        }) satisfies WindowPayload,
    });

    const pool = createRowPool({
      scrollEl: scroll,
      headerEl: header,
      bodyEl: body,
      spacerEl: spacer,
      rowHeight: 26,
      totalRows: 100_000,
      total: { kind: "exact", value: 100_000 },
      buffer,
      columns: buildColumns([{ name: "id", type: "int8" }]),
      getSelection: () => emptySelection(),
      onCellClick: () => {},
      onRowClick: () => {},
    });

    pool.layoutNow();
    const initial = pool.rowElementCount();
    expect(initial).toBeGreaterThan(0);
    expect(initial).toBeLessThan(30);
    scroll.scrollTop = 50_000;
    pool.layoutNow();
    expect(pool.rowElementCount()).toBeLessThan(30);
    expect(pool.rowElementCount()).toBe(initial);

    pool.destroy();
    buffer.destroy();
    scroll.remove();
  });
});
