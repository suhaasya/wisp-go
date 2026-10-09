import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { VirtualList } from "./VirtualList";

describe("VirtualList", () => {
  it("renders a window of rows, not every item", () => {
    const items = Array.from({ length: 500 }, (_, i) => `row-${i}`);
    const { container } = render(
      <VirtualList
        items={items}
        rowHeight={26}
        class="h-[200px]"
        render={(item) => <div data-row={item}>{item}</div>}
      />,
    );
    const rows = container.querySelectorAll("[data-row]");
    expect(rows.length).toBeLessThan(50);
    expect(rows.length).toBeGreaterThan(0);
    expect(screen.getByText("row-0")).toBeTruthy();
  });
});
