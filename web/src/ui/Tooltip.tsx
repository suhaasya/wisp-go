import { useId, useState } from "preact/hooks";
import { cx } from "./cx";

export interface TooltipProps {
  label: string;
  tip: string;
  class?: string;
}

export function Tooltip({ label, tip, class: className }: TooltipProps) {
  const [open, setOpen] = useState(false);
  const tipId = useId();

  return (
    <span
      class={cx("relative inline-flex", className)}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocusIn={() => setOpen(true)}
      onFocusOut={() => setOpen(false)}
    >
      <button
        type="button"
        class="border-0 bg-transparent p-0 text-13 text-accent underline decoration-dotted"
        aria-describedby={open ? tipId : undefined}
      >
        {label}
      </button>
      {open ? (
        <span
          id={tipId}
          role="tooltip"
          class="absolute bottom-full left-1/2 z-50 mb-1 max-w-xs -translate-x-1/2 rounded-control border border-line bg-panel px-2 py-1 text-12-5 text-ink shadow-lg"
        >
          {tip}
        </span>
      ) : null}
    </span>
  );
}
