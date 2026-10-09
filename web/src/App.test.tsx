import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { App } from "./App";

describe("App", () => {
  it("renders connections screen", async () => {
    render(<App />);
    expect(await screen.findByLabelText("Search connections")).toBeTruthy();
  });
});
