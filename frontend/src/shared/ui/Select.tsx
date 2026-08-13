import { ChevronDown } from "lucide-react";
import { cn } from "../cn";

/**
 * Native select under a styled shell — keyboard and screen-reader behaviour
 * comes free, which a div-based menu would have to re-implement.
 */
export function Select({
  value,
  onChange,
  options,
  className,
  "aria-label": ariaLabel,
}: {
  value: string;
  onChange: (value: string) => void;
  options: readonly string[];
  className?: string;
  "aria-label": string;
}) {
  return (
    <div
      className={cn(
        "relative inline-flex items-center rounded-lg border border-hairline bg-surface",
        "shadow-[0_1px_2px_rgba(16,24,40,0.04)] focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-brand",
        className,
      )}
    >
      <select
        aria-label={ariaLabel}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="cursor-pointer appearance-none bg-transparent py-2 pr-8 pl-3.5 text-sm font-medium text-ink focus:outline-none"
      >
        {options.map((o) => (
          <option key={o} value={o}>
            {o}
          </option>
        ))}
      </select>
      <ChevronDown
        aria-hidden
        className="pointer-events-none absolute right-2.5 size-4 text-ink-muted"
      />
    </div>
  );
}
