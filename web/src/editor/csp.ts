export function readCspNonce(): string {
  if (typeof document === "undefined") {
    return "";
  }
  return document.querySelector('meta[name="csp-nonce"]')?.getAttribute("content") ?? "";
}
