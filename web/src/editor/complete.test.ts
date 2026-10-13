import { PostgreSQL, schemaCompletionSource } from "@codemirror/lang-sql";
import { describe, expect, it } from "vitest";
import {
  mergeCompletionSchema,
  mergeSnippetsIntoNamespace,
  referencedTableNames,
  tablesNeedingColumnHydration,
  toSQLNamespace,
} from "./complete";

describe("toSQLNamespace", () => {
  it("maps tables to column completion arrays for CodeMirror", () => {
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
    const orders = (ns as Record<string, { label: string; detail?: string }[]>).orders;
    expect(orders?.map((c) => c.label).sort()).toEqual(["id", "status"]);
    expect(orders?.find((c) => c.label === "id")?.detail).toBe("int4");
  });

  it("builds a schema completion source without blowing the stack", () => {
    const tables = Array.from({ length: 80 }, (_, i) => ({
      name: `table_${i}`,
      columns: [
        { name: "id", type: "int4" },
        { name: "note", type: "text" },
      ],
    }));
    const ns = toSQLNamespace({ schema: "public", tables });
    expect(() =>
      schemaCompletionSource({ dialect: PostgreSQL, schema: ns, defaultSchema: "public" }),
    ).not.toThrow();
  });
});

describe("mergeSnippetsIntoNamespace", () => {
  it("adds snippet names with tag detail", () => {
    const ns = mergeSnippetsIntoNamespace({}, [{ name: "daily_orders", tags: ["report"] }]);
    const row = (ns as Record<string, { self: { detail?: string } }>).daily_orders;
    expect(row?.self?.detail).toBe("report");
  });
});

describe("mergeCompletionSchema", () => {
  it("merges when patch tables is null (API empty slice)", () => {
    const current = {
      schema: "public",
      tables: [{ name: "orders", columns: [{ name: "id", type: "int4" }] }],
    };
    const merged = mergeCompletionSchema(current, {
      schema: "public",
      tables: null as unknown as [],
    });
    expect(merged.tables.map((t) => t.name)).toEqual(["orders"]);
  });

  it("adds tables from patch", () => {
    const merged = mergeCompletionSchema(
      { schema: "public", tables: [{ name: "a", columns: [] }] },
      { schema: "public", tables: [{ name: "b", columns: [{ name: "x", type: "text" }] }] },
    );
    expect(merged.tables.map((t) => t.name)).toEqual(["a", "b"]);
  });
});

describe("tablesNeedingColumnHydration", () => {
  it("requests columns for referenced tables without column metadata", () => {
    const need = tablesNeedingColumnHydration(
      [{ name: "user", columns: [] }, { name: "orders", columns: [{ name: "id", type: "int4" }] }],
      ["user", "orders"],
    );
    expect(need).toEqual(["user"]);
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
