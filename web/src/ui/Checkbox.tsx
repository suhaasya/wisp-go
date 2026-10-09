import { cx } from "./cx";

export interface CheckboxProps {
  id?: string;
  label: string;
  checked?: boolean;
  disabled?: boolean;
  onChange?: (checked: boolean) => void;
}

export function Checkbox({ id, label, checked, disabled, onChange }: CheckboxProps) {
  const boxId = id ?? `check-${label.replace(/\s+/g, "-").toLowerCase()}`;
  return (
    <label
      for={boxId}
      class={cx(
        "flex cursor-pointer items-center gap-2 text-12-5 text-ink",
        disabled && "cursor-not-allowed opacity-50",
      )}
    >
      <input
        id={boxId}
        type="checkbox"
        class="accent-accent"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange?.((e.target as HTMLInputElement).checked)}
      />
      {label}
    </label>
  );
}
