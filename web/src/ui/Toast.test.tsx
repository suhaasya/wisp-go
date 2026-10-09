import { render, screen } from "@testing-library/preact";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { dismissToast, pushToast, ToastViewport, toasts } from "./Toast";

describe("Toast", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    toasts.value = [];
  });

  afterEach(() => {
    vi.useRealTimers();
    toasts.value = [];
  });

  it("shows message and auto-dismisses", () => {
    pushToast("Connected", "success", 1000);
    render(<ToastViewport />);
    expect(screen.getByText("Connected")).toBeTruthy();
    vi.advanceTimersByTime(1000);
    expect(toasts.value).toHaveLength(0);
  });

  it("dismissToast removes one entry", () => {
    pushToast("One");
    pushToast("Two");
    dismissToast(toasts.value[0]!.id);
    expect(toasts.value).toHaveLength(1);
    expect(toasts.value[0]!.message).toBe("Two");
  });
});
