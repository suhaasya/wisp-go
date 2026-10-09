import { describe, expect, it } from "vitest";
import { referencedTableNames, toSQLNamespace } from "./complete";

describe("toSQLNamespace", () => {
  it("maps tables and column detail from completion payload", () => {
    const ns = toSQLNamespace({
      schema: "public",
      tables: [
        {
          name: "orders",
          columns: [
            { name: "id", type: "int4" },
            { name: "status", type: "text" },
          ],
        },
      ],
    });
    const orders = (ns as Record<string, Record<string, { detail?: string }>>).orders;
    expect(orders?.id?.detail).toBe("int4");
    expect(orders?.status?.detail).toBe("text");
  });
});

describe("referencedTableNames", () => {
  it("collects names after FROM and JOIN", () => {
    const names = referencedTableNames(
      "SELECT * FROM users u JOIN orders o ON o.user_id = u.id",
    );
    expect(names.sort()).toEqual(["orders", "users"]);
  });
});
