import { describe, expect, it } from "vitest";
import { defaultSSLForHost, isLocalHost, sslDisableWarning } from "./ssl-helpers";

describe("defaultSSLForHost", () => {
  it("uses verify-full for detected providers", () => {
    expect(defaultSSLForHost("db.example.com", "neon").mode).toBe("verify-full");
  });

  it("uses require for remote hosts", () => {
    expect(defaultSSLForHost("db.example.com").mode).toBe("require");
  });

  it("uses disable for localhost", () => {
    expect(defaultSSLForHost("127.0.0.1").mode).toBe("disable");
  });
});

describe("sslDisableWarning", () => {
  it("warns when TLS is disabled for a remote host", () => {
    expect(sslDisableWarning("db.internal", "disable")).toBe(true);
    expect(sslDisableWarning("127.0.0.1", "disable")).toBe(false);
  });
});

describe("isLocalHost", () => {
  it("recognizes loopback names", () => {
    expect(isLocalHost("localhost")).toBe(true);
    expect(isLocalHost("10.0.0.1")).toBe(false);
  });
});
