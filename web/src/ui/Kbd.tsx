import { cx } from "./cx";

export interface KbdProps {
  children: string;
  class?: string;
}

export function Kbd({ children, class: className }: KbdProps) {
  return (
    <kbd
      class={cx(
        "inline-block rounded border border-line border-b-2 bg-raise px-1 font-mono text-11-5 font-medium text-ink-2",
        className,
      )}
    >
      {children}
    </kbd>
  );
}
