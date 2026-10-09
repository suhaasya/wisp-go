import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { ObjectContextMenu } from "./menu";

describe("ObjectContextMenu", () => {
  it("hides destructive items when read-only", () => {
    render(
      <ObjectContextMenu
        open
        x={0}
        y={0}
        readOnly
        tablesOnly
        onClose={() => {}}
        onAction={() => {}}
      />,
    );
    expect(screen.queryByText("Drop…")).toBeNull();
  });
});
