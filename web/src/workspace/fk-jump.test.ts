import { describe, expect, it } from "vitest";
import { buildForeignKeyJump } from "./fk-jump";

describe("buildForeignKeyJump", () => {
  it("builds composite AND filter and link label", () => {
    const jump = buildForeignKeyJump(
      {
        name: "orders_customer_fk",
        columns: ["tenant_id", "customer_id"],
        ref_schema: "public",
        ref_table: "customers",
        ref_columns: ["tenant_id", "id"],
      },
      { tenant_id: "1", customer_id: "40000" },
    );
    if (!jump) {
      throw new Error("expected jump");
    }
    expect(jump.refSchema).toBe("public");
    expect(jump.refTable).toBe("customers");
    expect(jump.linkLabel).toBe("Open customers 40000");
    expect(jump.browseQuery.filter.logic).toBe("and");
    expect(jump.browseQuery.filter.conditions).toEqual([
      { column: "tenant_id", op: "=", value: "1" },
      { column: "id", op: "=", value: "40000" },
    ]);
  });
});
