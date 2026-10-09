import { cx } from "./cx";

export interface SegmentedOption {
  value: string;
  label: string;
}

export interface SegmentedProps {
  options: SegmentedOption[];
  value: string;
  "aria-label"?: string;
  onChange: (value: string) => void;
}

export function Segmented({ options, value, "aria-label": ariaLabel, onChange }: SegmentedProps) {
  return (
    <div
      class="inline-flex overflow-hidden rounded-control border border-line bg-panel"
      role="group"
      aria-label={ariaLabel}
    >
      {options.map((opt, i) => (
        <button
          key={opt.value}
          type="button"
          class={cx(
            "border-0 px-[11px] py-1 text-13",
            i > 0 && "border-l border-line",
            value === opt.value
              ? "bg-accent-soft font-medium text-ink ring-1 ring-inset ring-accent"
              : "bg-transparent text-ink-2",
          )}
          aria-pressed={value === opt.value}
          onClick={() => onChange(opt.value)}
        >
          {opt.label}
        </button>
      ))}
    </div>
  );
}
