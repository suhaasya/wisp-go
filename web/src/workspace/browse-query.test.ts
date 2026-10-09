import { describe, expect, it } from "vitest";
import { workspaceTabs } from "../state";
import {
  browseQueryKey,
  emptyBrowseQuery,
  updateTableTabBrowseQuery,
} from "./browse-query";

describe("browse query per tab", () => {
  it("keeps independent browseQuery on each tab", () => {
    workspaceTabs.value = [
      { id: "a", kind: "table_data", schema: "public", table: "orders", pendingResultLoad: false },
      { id: "b", kind: "table_data", schema: "public", table: "items", pendingResultLoad: false },
    ];
    const qA = { ...emptyBrowseQuery(), rawWhere: "status = 'paid'" };
    const qB = { ...emptyBrowseQuery(), rawWhere: "qty > 0" };
    updateTableTabBrowseQuery("a", qA);
    updateTableTabBrowseQuery("b", qB);
    const tabA = workspaceTabs.value.find((t) => t.id === "a");
    const tabB = workspaceTabs.value.find((t) => t.id === "b");
    expect(tabA?.kind === "table_data" ? tabA.browseQuery?.rawWhere : "").toBe("status = 'paid'");
    expect(tabB?.kind === "table_data" ? tabB.browseQuery?.rawWhere : "").toBe("qty > 0");
    const keyA = tabA?.kind === "table_data" ? browseQueryKey(tabA.browseQuery) : "";
    const keyB = tabB?.kind === "table_data" ? browseQueryKey(tabB.browseQuery) : "";
    expect(keyA).not.toBe(keyB);
  });
});
