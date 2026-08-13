import { Card, CardHeader } from "@/shared/ui/Card";
import { ViewAllButton } from "@/shared/ui/Button";
import { cn } from "@/shared/cn";
import { stockAlerts } from "../data";

export function StockAlert() {
  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Stock Alert" action={<ViewAllButton />} />
      <ul className="flex-1 divide-y divide-hairline-soft px-4 pb-2">
        {stockAlerts.map((s) => {
          const out = s.qty === 0;
          return (
            <li key={s.id} className="flex items-center gap-2.5 py-3">
              <span
                aria-hidden
                className="grid size-8 shrink-0 place-items-center rounded-lg bg-hairline-soft text-sm"
              >
                {s.emoji}
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate text-[13px] font-medium text-ink">
                  {s.name}
                </span>
                {/* State is stated in words, not carried by the colour alone. */}
                <span
                  className={cn(
                    "block text-xs font-medium",
                    out ? "text-status-critical" : "text-brand-strong",
                  )}
                >
                  {out ? "Out of Stock" : "Low Stock"}
                </span>
              </span>
              <span
                className={cn(
                  "grid min-w-9 place-items-center rounded-lg px-2 py-1 text-[13px] font-semibold tnum",
                  out
                    ? "bg-status-critical-wash text-status-critical"
                    : "bg-brand-wash text-brand-strong",
                )}
              >
                {s.qty}
              </span>
            </li>
          );
        })}
      </ul>
    </Card>
  );
}
