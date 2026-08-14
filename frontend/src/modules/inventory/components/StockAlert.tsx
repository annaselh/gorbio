import { Card, CardHeader } from "@/shared/ui/Card";
import { ViewAllButton } from "@/shared/ui/Button";
import { cn } from "@/shared/cn";
import { isOutOfStock, useStockItems } from "../data";

export function StockAlert() {
  const { data: items, isPending, isError } = useStockItems({
    lowStockOnly: true,
  });

  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Stock Alert" action={<ViewAllButton />} />

      {isPending ? (
        <StockAlertMessage>Loading stock levels…</StockAlertMessage>
      ) : isError ? (
        <StockAlertMessage>
          Stock levels are unavailable right now.
        </StockAlertMessage>
      ) : items.length === 0 ? (
        <StockAlertMessage>
          Every item is above its reorder level.
        </StockAlertMessage>
      ) : (
        <ul className="flex-1 divide-y divide-hairline-soft px-4 pb-2">
          {items.map((item) => {
            const out = isOutOfStock(item);
            return (
              <li key={item.id} className="flex items-center gap-2.5 py-3">
                <span
                  aria-hidden
                  className="grid size-8 shrink-0 place-items-center rounded-lg bg-hairline-soft text-[11px] font-semibold text-ink-secondary"
                >
                  {item.unit}
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-[13px] font-medium text-ink">
                    {item.name}
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
                  {item.quantity}
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function StockAlertMessage({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid flex-1 place-items-center px-4 py-8">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
