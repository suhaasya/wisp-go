import { cx } from "./cx";

export interface InputProps {
  id?: string;
  label?: string;
  type?: "text" | "password";
  mono?: boolean;
  value?: string;
  placeholder?: string;
  disabled?: boolean;
  ariaLabel?: string;
  onInput?: (value: string) => void;
}

const fieldClass =
  "w-full rounded-control border border-line bg-panel py-[6px] px-[9px] text-13 text-ink placeholder:text-ink-3 disabled:opacity-50";

export function Input({
  id,
  label,
  type = "text",
  mono,
  value,
  placeholder,
  disabled,
  ariaLabel,
  onInput,
}: InputProps) {
  const inputId = id ?? (label ? `input-${label.replace(/\s+/g, "-").toLowerCase()}` : undefined);
  return (
    <div>
      {label ? (
        <label for={inputId} class="mb-1 block text-12-5 font-medium text-ink-2">
          {label}
        </label>
      ) : null}
      {type === "password" ? (
        <input
          id={inputId}
          type="password"
          aria-label={ariaLabel}
          class={cx(fieldClass, mono && "font-mono text-12-5")}
          value={value}
          placeholder={placeholder}
          disabled={disabled}
          onInput={(e) => onInput?.((e.target as HTMLInputElement).value)}
        />
      ) : (
        <input
          id={inputId}
          aria-label={ariaLabel}
          class={cx(fieldClass, mono && "font-mono text-12-5")}
          value={value}
          placeholder={placeholder}
          disabled={disabled}
          onInput={(e) => onInput?.((e.target as HTMLInputElement).value)}
        />
      )}
    </div>
  );
}
