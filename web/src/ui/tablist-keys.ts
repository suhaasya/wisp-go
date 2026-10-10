/** WAI-ARIA tabs pattern: arrows move focus; Enter/Space activate the focused tab. */
export function handleHorizontalTablistKeydown(
  e: KeyboardEvent,
  onActivateIndex: (index: number) => void,
): void {
  const list = e.currentTarget as HTMLElement;
  const items = [...list.querySelectorAll<HTMLElement>('[role="tab"]')];
  if (items.length === 0) {
    return;
  }

  let i = items.indexOf(document.activeElement as HTMLElement);
  if (i < 0) {
    i = items.findIndex((el) => el.getAttribute("aria-selected") === "true");
  }
  if (i < 0) {
    i = 0;
  }

  if (e.key === "ArrowRight") {
    e.preventDefault();
    items[(i + 1) % items.length]?.focus();
    return;
  }
  if (e.key === "ArrowLeft") {
    e.preventDefault();
    items[(i - 1 + items.length) % items.length]?.focus();
    return;
  }
  if (e.key === "Home") {
    e.preventDefault();
    items[0]?.focus();
    return;
  }
  if (e.key === "End") {
    e.preventDefault();
    items[items.length - 1]?.focus();
    return;
  }
  if (e.key === "Enter" || e.key === " ") {
    e.preventDefault();
    const focused = items.indexOf(document.activeElement as HTMLElement);
    if (focused >= 0) {
      onActivateIndex(focused);
    }
  }
}
