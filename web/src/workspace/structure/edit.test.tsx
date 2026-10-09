import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { StructureDDLPanel } from "./ddl";
import { StructureColumnsSection } from "./columns";
import type { TableDetailsResponse } from "../schema-client";

describe("Structure editor UI", () => {
  it("shows Pending changes title when pending SQL is set", () => {
    render(
      <StructureDDLPanel title="Pending changes" ddl={"ALTER TABLE t ADD COLUMN c int;"} loading={false} />,
    );
    expect(screen.getByTestId("structure-ddl-title").textContent).toContain("Pending changes");
  });

  it("marks modified rows with bg-modified", () => {
    const details: TableDetailsResponse = {
      schema: "public",
      table: "t",
      columns: [{ name: "id", type: "int", nullable: false }],
    };
    render(<StructureColumnsSection details={details} rowMarks={{ id: "mod" }} />);
    const row = document.querySelector('[data-row-mark="mod"]');
    expect(row?.className.includes("bg-modified")).toBe(true);
  });
});
