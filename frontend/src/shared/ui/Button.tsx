import type { ButtonHTMLAttributes } from "react";
import { cn } from "../cn";

type Variant = "primary" | "outline" | "ghost" | "link";

const VARIANTS: Record<Variant, string> = {
  primary:
    "bg-brand text-white hover:bg-brand-strong shadow-[0_1px_2px_rgba(16,24,40,0.06)]",
  outline:
    "border border-hairline bg-surface text-ink hover:bg-hairline-soft shadow-[0_1px_2px_rgba(16,24,40,0.04)]",
  ghost: "text-ink-secondary hover:bg-hairline-soft hover:text-ink",
  link: "text-ink-secondary hover:text-ink",
};

export function Button({
  variant = "outline",
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant }) {
  return (
    <button
      type="button"
      className={cn(
        "inline-flex cursor-pointer items-center justify-center gap-2 rounded-lg text-sm font-medium transition-colors",
        "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand",
        "disabled:cursor-not-allowed disabled:opacity-50",
        variant === "link" ? "px-0 py-0" : "px-3.5 py-2",
        VARIANTS[variant],
        className,
      )}
      {...props}
    />
  );
}

/** The small "View All" affordance repeated across dashboard cards. */
export function ViewAllButton(props: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <Button
      variant="outline"
      className="rounded-lg px-3 py-1.5 text-xs text-ink-secondary"
      {...props}
    >
      View All
    </Button>
  );
}
