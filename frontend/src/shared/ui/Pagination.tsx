import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "../cn";

/** Compact page list with ellipsis: 1 2 3 … 25 */
function pageList(page: number, pageCount: number): (number | "gap")[] {
  if (pageCount <= 6)
    return Array.from({ length: pageCount }, (_, i) => i + 1);

  const around = [page - 1, page, page + 1].filter(
    (p) => p > 1 && p < pageCount,
  );
  const core = [...new Set([1, ...around])].sort((a, b) => a - b);

  const out: (number | "gap")[] = [];
  for (const p of core) {
    if (out.length) {
      const prev = out[out.length - 1];
      if (typeof prev === "number" && p - prev > 1) out.push("gap");
    }
    out.push(p);
  }
  const last = out[out.length - 1];
  if (typeof last === "number" && pageCount - last > 1) out.push("gap");
  if (last !== pageCount) out.push(pageCount);
  return out;
}

export function Pagination({
  page,
  pageCount,
  total,
  pageSize,
  onPageChange,
}: {
  page: number;
  pageCount: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}) {
  const from = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, total);

  const arrow =
    "grid size-8 place-items-center rounded-lg border border-hairline bg-surface text-ink-secondary transition-colors hover:bg-hairline-soft disabled:cursor-not-allowed disabled:opacity-40";

  return (
    <nav
      aria-label="Pagination"
      className="flex flex-wrap items-center justify-between gap-3 px-5 py-3.5"
    >
      <p className="text-xs text-ink-secondary tnum">
        {from}–{to} of {total}
      </p>

      <div className="flex items-center gap-1.5">
        <button
          type="button"
          className={arrow}
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
          aria-label="Previous page"
        >
          <ChevronLeft className="size-4" />
        </button>

        {pageList(page, pageCount).map((p, i) =>
          p === "gap" ? (
            <span
              key={`gap-${i}`}
              aria-hidden
              className="px-1 text-xs text-ink-muted"
            >
              …
            </span>
          ) : (
            <button
              key={p}
              type="button"
              onClick={() => onPageChange(p)}
              aria-label={`Page ${p}`}
              aria-current={p === page ? "page" : undefined}
              className={cn(
                "grid size-8 cursor-pointer place-items-center rounded-lg text-xs font-medium transition-colors tnum",
                p === page
                  ? "bg-brand-wash text-brand-strong"
                  : "text-ink-secondary hover:bg-hairline-soft",
              )}
            >
              {p}
            </button>
          ),
        )}

        <button
          type="button"
          className={arrow}
          onClick={() => onPageChange(page + 1)}
          disabled={page >= pageCount}
          aria-label="Next page"
        >
          <ChevronRight className="size-4" />
        </button>
      </div>
    </nav>
  );
}
