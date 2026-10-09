import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { Tooltip } from "./Tooltip";

describe("Tooltip", () => {
  it("shows tip on focus", () => {
    render(<Tooltip label="Host" tip="Database hostname or IP" />);
    const trigger = screen.getByRole("button", { name: "Host" });
    fireEvent.mouseEnter(trigger.parentElement!);
    expect(screen.getByRole("tooltip").textContent).toBe("Database hostname or IP");
  });
});
