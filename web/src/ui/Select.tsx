import { cx } from "./cx";

export interface SelectOption {
  value: string;
  label: string;
}

export interface SelectProps {
  id?: string;
  label?: string;
  value?: string;
  options: SelectOption[];
  disabled?: boolean;
  onChange?: (value: string) => void;
}

const selectClass =
  "w-full rounded-control border border-line bg-panel py-[6px] px-[9px] text-13 text-ink disabled:opacity-50";

export function Select({ id, label, value, options, disabled, onChange }: SelectProps) {
  const selectId = id ?? (label ? `select-${label.replace(/\s+/g, "-").toLowerCase()}` : undefined);
  return (
    <div>
      {label ? (
        <label for={selectId} class="mb-1 block text-12-5 font-medium text-ink-2">
          {label}
        </label>
      ) : null}
      <select
        id={selectId}
        class={selectClass}
        value={value}
        disabled={disabled}
        onChange={(e) => onChange?.((e.target as HTMLSelectElement).value)}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}
