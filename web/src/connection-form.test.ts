import { describe, expect, it } from "vitest";
import {
  buildTestRequest,
  defaultFormState,
  formFromParseResult,
  validateConnectionForm,
} from "./connection-form";

describe("formFromParseResult", () => {
  it("maps parse-url profile and password into form fields", () => {
    const form = formFromParseResult(
      {
        engine: "postgres",
        database: "app",
        user: "u",
        transport: { kind: "tcp", host: "db.example.com", port: 5432 },
        password_mode: "store",
      },
      "secret",
    );
    expect(form.name).toBe("app");
    expect(form.host).toBe("db.example.com");
    expect(form.password).toBe("secret");
    expect(form.passwordDirty).toBe(true);
  });
});

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
