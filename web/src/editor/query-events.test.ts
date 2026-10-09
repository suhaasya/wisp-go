import { describe, expect, it, vi } from "vitest";
import { emitQueryEvent, onQueryEvent } from "./query-events";

describe("query-events", () => {
  it("notifies listeners for query SSE payloads", () => {
    const fn = vi.fn();
    const off = onQueryEvent(fn);
    emitQueryEvent({ sid: "s1", qid: "q1", done: true });
    expect(fn).toHaveBeenCalledWith({ sid: "s1", qid: "q1", done: true });
    off();
    emitQueryEvent({ sid: "s2" });
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
