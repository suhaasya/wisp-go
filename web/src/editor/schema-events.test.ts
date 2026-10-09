import { describe, expect, it, vi } from "vitest";
import { emitSchemaInvalidated, onSchemaInvalidated } from "./schema-events";

describe("schema-events", () => {
  it("notifies listeners when schema is invalidated", () => {
    const fn = vi.fn();
    const off = onSchemaInvalidated(fn);
    emitSchemaInvalidated("sid-1");
    expect(fn).toHaveBeenCalledWith("sid-1");
    off();
    emitSchemaInvalidated("sid-2");
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
