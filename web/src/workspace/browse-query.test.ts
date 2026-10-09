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
      { id: "a", schema: "public", table: "orders", mode: "data" },
      { id: "b", schema: "public", table: "items", mode: "data" },
    ];
    const qA = { ...emptyBrowseQuery(), rawWhere: "status = 'paid'" };
    const qB = { ...emptyBrowseQuery(), rawWhere: "qty > 0" };
    updateTableTabBrowseQuery("a", qA);
    updateTableTabBrowseQuery("b", qB);
    expect(workspaceTabs.value.find((t) => t.id === "a")?.browseQuery?.rawWhere).toBe(
      "status = 'paid'",
    );
    expect(workspaceTabs.value.find((t) => t.id === "b")?.browseQuery?.rawWhere).toBe("qty > 0");
    expect(browseQueryKey(workspaceTabs.value[0]?.browseQuery)).not.toBe(
      browseQueryKey(workspaceTabs.value[1]?.browseQuery),
    );
  });
});
