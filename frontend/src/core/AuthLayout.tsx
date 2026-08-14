import type { ReactNode } from "react";
import { cn } from "@/shared/cn";

/** Shared input styling for every unauthenticated form. */
export const FIELD_CLASS = cn(
  "w-full rounded-lg border border-hairline bg-surface px-3 py-2 text-sm text-ink",
  "placeholder:text-ink-secondary",
  "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand",
  "disabled:cursor-not-allowed disabled:opacity-50",
);

/** Centred card used by sign-in, password reset and email verification. */
export function AuthLayout({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  /** Optional: a purely informational state such as "Verifying…" has no body. */
  children?: ReactNode;
}) {
  return (
    <main className="grid min-h-dvh place-items-center bg-canvas px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <h1 className="text-xl font-semibold text-ink">{title}</h1>
          {subtitle ? (
            <p className="mt-1 text-sm text-ink-secondary">{subtitle}</p>
          ) : null}
        </div>
        <div className="space-y-4">{children}</div>
      </div>
    </main>
  );
}

export function AuthMessage({
  children,
  tone = "info",
}: {
  children: ReactNode;
  tone?: "info" | "alert";
}) {
  return (
    <p
      role={tone === "alert" ? "alert" : undefined}
      className={cn(
        "rounded-lg border px-3 py-2 text-sm",
        tone === "alert"
          ? "border-hairline bg-hairline-soft text-ink"
          : "border-hairline bg-hairline-soft text-ink-secondary",
      )}
    >
      {children}
    </p>
  );
}
