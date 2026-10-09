import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import type { TableDetailsResponse } from "../schema-client";
import { StructureColumnsSection } from "./columns";

const fixture: TableDetailsResponse = {
  schema: "wisp_meta",
  table: "parent",
  columns: [
    { name: "id", type: "bigint", nullable: false },
    { name: "status", type: "text", nullable: false, default: "'new'::text" },
  ],
  primary_key: ["id"],
  foreign_keys: [],
};

describe("StructureColumnsSection", () => {
  it("renders column rows and primary key label", () => {
    render(<StructureColumnsSection details={fixture} />);
    expect(screen.getByTestId("structure-columns")).toBeTruthy();
    expect(screen.getByText("id")).toBeTruthy();
    expect(screen.getByText("Primary")).toBeTruthy();
    expect(screen.getByText("'new'::text")).toBeTruthy();
  });
});
