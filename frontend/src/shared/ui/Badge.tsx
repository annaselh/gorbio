import { cn } from "../cn";

export type StatusTone = "good" | "info" | "critical" | "warning" | "neutral";

const TONES: Record<StatusTone, string> = {
  good: "bg-status-good-wash text-status-good",
  info: "bg-status-info-wash text-status-info",
  critical: "bg-status-critical-wash text-status-critical",
  warning: "bg-status-warning-wash text-status-warning",
  neutral: "bg-hairline-soft text-ink-secondary",
};

/**
 * Status is never colour-alone — the label is always rendered, so the tone is
 * redundant reinforcement rather than the only signal.
 */
export function Badge({
  tone = "neutral",
  children,
  className,
}: {
  tone?: StatusTone;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md px-2.5 py-1 text-xs font-medium",
        TONES[tone],
        className,
      )}
    >
      {children}
    </span>
  );
}

export function Dot({ tone = "neutral" }: { tone?: StatusTone }) {
  const colors: Record<StatusTone, string> = {
    good: "bg-status-good",
    info: "bg-status-info",
    critical: "bg-status-critical",
    warning: "bg-brand",
    neutral: "bg-ink-muted",
  };
  return (
    <span
      aria-hidden
      className={cn("inline-block size-2 shrink-0 rounded-full", colors[tone])}
    />
  );
}
