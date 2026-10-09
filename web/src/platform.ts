/** True when the UI runs on Windows (named pipe transport). */
export function isWindowsPlatform(): boolean {
  if (typeof navigator === "undefined") {
    return false;
  }
  return /Win/i.test(navigator.platform) || /Windows/i.test(navigator.userAgent);
}
