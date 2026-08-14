import { useEffect, useId, useRef } from "react";
import type { ReactNode } from "react";
import { X } from "lucide-react";
import { cn } from "../cn";

/**
 * Centred dialog. Closes on Escape and on a click that starts outside the
 * panel — starting the check on mousedown rather than click stops a drag that
 * began inside the panel and ended outside from dismissing the form.
 */
export function Modal({
  title,
  description,
  onClose,
  children,
  size = "md",
}: {
  title: string;
  description?: string;
  onClose: () => void;
  children: ReactNode;
  size?: "md" | "lg";
}) {
  const panelRef = useRef<HTMLDivElement>(null);
  const titleId = useId();

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    // Stop the page behind from scrolling while the dialog is open.
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = previousOverflow;
    };
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center overflow-y-auto bg-ink/30 p-4"
      onMouseDown={(event) => {
        if (!panelRef.current?.contains(event.target as Node)) onClose();
      }}
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className={cn(
          "w-full rounded-2xl border border-hairline bg-surface shadow-[0_16px_48px_rgba(16,24,40,0.18)]",
          size === "lg" ? "max-w-3xl" : "max-w-md",
        )}
      >
        <header className="flex items-start gap-4 border-b border-hairline-soft px-5 py-4">
          <div className="min-w-0 flex-1">
            <h2 id={titleId} className="text-lg font-semibold text-ink">
              {title}
            </h2>
            {description ? (
              <p className="mt-0.5 text-sm text-ink-secondary">{description}</p>
            ) : null}
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="grid size-8 shrink-0 cursor-pointer place-items-center rounded-lg text-ink-secondary transition-colors hover:bg-hairline-soft hover:text-ink"
          >
            <X className="size-4" />
          </button>
        </header>

        <div className="px-5 py-4">{children}</div>
      </div>
    </div>
  );
}

/** Shared field styling for every form input in the app. */
export const FIELD_CLASS = cn(
  "w-full rounded-lg border border-hairline bg-surface px-3 py-2 text-sm text-ink",
  "placeholder:text-ink-muted",
  "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand",
  "disabled:cursor-not-allowed disabled:opacity-50",
);

export function Field({
  label,
  hint,
  error,
  children,
}: {
  label: string;
  hint?: string;
  error?: string;
  children: ReactNode;
}) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-sm font-medium text-ink">
        {label}
        {hint ? (
          <span className="ml-1 font-normal text-ink-secondary">{hint}</span>
        ) : null}
      </span>
      {children}
      {error ? <span className="text-xs text-status-critical">{error}</span> : null}
    </label>
  );
}

export function FormError({ children }: { children: ReactNode }) {
  return (
    <p
      role="alert"
      className="rounded-lg border border-hairline bg-hairline-soft px-3 py-2 text-sm text-status-critical"
    >
      {children}
    </p>
  );
}
