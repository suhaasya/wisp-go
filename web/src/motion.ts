/** User prefers reduced motion (Safari/Firefox/Chrome). */
export function prefersReducedMotion(): boolean {
  if (typeof matchMedia === "undefined") {
    return false;
  }
  return matchMedia("(prefers-reduced-motion: reduce)").matches;
}
