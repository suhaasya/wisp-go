import { describe, expect, it } from "vitest";
import {
  buildTestRequest,
  defaultFormState,
  validateConnectionForm,
} from "./connection-form";

describe("validateConnectionForm", () => {
  it("requires production safe mode", () => {
    const f = { ...defaultFormState(), name: "x", user: "u", database: "d", env: "production" as const };
    const err = validateConnectionForm(f);
    expect(err.productionSafe).toBeTruthy();
  });
});

describe("buildTestRequest", () => {
  it("omits password when unchanged on edit", () => {
    const f = {
      ...defaultFormState(),
      id: "p1",
      name: "n",
      user: "u",
      database: "d",
      password: "",
      passwordDirty: false,
    };
    const body = buildTestRequest(f, true);
    expect(body.password).toBeUndefined();
    expect(body.profile_id).toBe("p1");
  });

  it("includes password when dirty", () => {
    const f = {
      ...defaultFormState(),
      id: "p1",
      name: "n",
      user: "u",
      database: "d",
      password: "new",
      passwordDirty: true,
    };
    const body = buildTestRequest(f, true);
    expect(body.password).toBe("new");
  });
});
