import type { ReactNode } from "react";

/**
 * Shared tooltip surface. Values wear text tokens, never the series colour —
 * the swatch beside them carries identity.
 */
export function TooltipShell({
  label,
  rows,
}: {
  label: ReactNode;
  rows: { key: string; color: string; name: string; value: string }[];
}) {
  return (
    <div className="pointer-events-none rounded-xl border border-hairline bg-surface px-3 py-2 shadow-[0_8px_24px_rgba(16,24,40,0.12)]">
      <p className="mb-1.5 text-xs font-medium text-ink-secondary">{label}</p>
      <ul className="space-y-1">
        {rows.map((r) => (
          <li key={r.key} className="flex items-center gap-2 text-xs">
            <span
              aria-hidden
              className="size-2 shrink-0 rounded-full"
              style={{ backgroundColor: r.color }}
            />
            <span className="text-ink-secondary">{r.name}</span>
            <span className="ml-auto pl-3 font-semibold text-ink tnum">
              {r.value}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
