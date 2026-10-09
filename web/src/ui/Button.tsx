import { cx } from "./cx";
import { Kbd } from "./Kbd";

export interface ButtonProps {
  label: string;
  variant?: "default" | "primary" | "warning" | "ghost" | "danger";
  size?: "sm" | "md";
  kbd?: string;
  disabled?: boolean;
  type?: "button" | "submit" | "reset";
  onClick?: () => void;
}

const variantClass: Record<NonNullable<ButtonProps["variant"]>, string> = {
  default: "border border-line bg-panel hover:bg-raise",
  primary: "border border-accent bg-accent text-on-accent hover:opacity-95",
  warning: "border border-modified-line bg-modified text-ink hover:bg-modified",
  ghost: "border border-transparent bg-transparent hover:bg-raise text-ink-2",
  danger: "border border-env-prod bg-panel text-env-prod hover:bg-deleted",
};

const sizeClass: Record<NonNullable<ButtonProps["size"]>, string> = {
  md: "py-[5px] px-[11px] text-13",
  sm: "py-[3px] px-2 text-12-5",
};

export function Button({
  label,
  variant = "default",
  size = "md",
  kbd,
  disabled,
  type = "button",
  onClick,
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      class={cx(
        "inline-flex items-center gap-2 rounded-control font-medium",
        variantClass[variant],
        sizeClass[size],
        disabled && "cursor-not-allowed opacity-50",
      )}
    >
      <span>{label}</span>
      {kbd ? <Kbd>{kbd}</Kbd> : null}
    </button>
  );
}
