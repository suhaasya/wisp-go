import { describe, expect, it, vi } from "vitest";
import { handleHorizontalTablistKeydown } from "./tablist-keys";

function tablistWithTabs(count: number): HTMLDivElement {
  const list = document.createElement("div");
  list.setAttribute("role", "tablist");
  for (let i = 0; i < count; i++) {
    const tab = document.createElement("button");
    tab.type = "button";
    tab.setAttribute("role", "tab");
    tab.setAttribute("aria-selected", i === 0 ? "true" : "false");
    tab.textContent = `tab ${i}`;
    list.appendChild(tab);
  }
  document.body.appendChild(list);
  list.querySelector<HTMLElement>('[role="tab"]')?.focus();
  return list;
}

describe("handleHorizontalTablistKeydown", () => {
  it("moves focus with ArrowRight and activates on Enter", () => {
    const list = tablistWithTabs(3);
    const activate = vi.fn();
    list.addEventListener("keydown", (e) => handleHorizontalTablistKeydown(e, activate));

    list.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }));
    const tabs = [...list.querySelectorAll<HTMLElement>('[role="tab"]')];
    expect(document.activeElement).toBe(tabs[1]);

    list.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    expect(activate).toHaveBeenCalledWith(1);

    list.remove();
  });
});
