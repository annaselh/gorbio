import type { ReactNode } from "react";
import { cn } from "../cn";

export function Card({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <section
      className={cn(
        "rounded-2xl border border-hairline bg-surface shadow-[0_1px_2px_rgba(16,24,40,0.04)]",
        className,
      )}
    >
      {children}
    </section>
  );
}

export function CardHeader({
  title,
  action,
  className,
}: {
  title: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <header
      className={cn(
        "flex items-center justify-between gap-3 px-5 pt-4 pb-3",
        className,
      )}
    >
      <h2 className="text-[15px] font-semibold text-ink">{title}</h2>
      {action}
    </header>
  );
}
