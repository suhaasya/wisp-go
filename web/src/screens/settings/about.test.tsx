import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { WISP_FEEDBACK_URL } from "../../feedback";
import { AboutPane } from "./about";

describe("AboutPane feedback", () => {
  it("links to GitHub Issues for beta feedback", () => {
    render(<AboutPane />);
    const link = screen.getByRole("link", { name: "Send feedback" });
    expect(link.getAttribute("href")).toBe(WISP_FEEDBACK_URL);
    expect(link.getAttribute("target")).toBe("_blank");
    expect(link.getAttribute("rel")).toContain("noopener");
  });
});
