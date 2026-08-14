import { ArrowDown, ArrowUp } from "lucide-react";
import { Card, CardHeader } from "@/shared/ui/Card";
import { formatIDR } from "@/shared/format";
import { cn } from "@/shared/cn";
import { useTopProducts } from "../api";

export function TopProducts() {
  const { data: products, isPending, isError } = useTopProducts(5);

  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Top Products" />

      {isPending ? (
        <Message>Loading top products…</Message>
      ) : isError ? (
        <Message>Top products are unavailable right now.</Message>
      ) : products.length === 0 ? (
        <Message>No confirmed sales this month.</Message>
      ) : (
        <ul className="flex-1 divide-y divide-hairline-soft px-4 pb-2">
          {products.map((product) => {
            const up = product.delta >= 0;
            const Arrow = up ? ArrowUp : ArrowDown;
            return (
              <li key={product.sku} className="flex items-center gap-2 py-3">
                <span
                  aria-hidden
                  className="grid size-7 shrink-0 place-items-center rounded-lg bg-hairline-soft text-[10px] font-semibold text-ink-secondary"
                >
                  {product.quantity}
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-xs font-medium text-ink">
                    {product.description}
                  </span>
                  <span className="block truncate text-[11px] text-ink-muted">
                    {product.sku}
                  </span>
                </span>
                {/* shrink-0 so the figures hold their width and the name is the
                    only thing that ever truncates. */}
                <span className="shrink-0 text-[11px] whitespace-nowrap text-ink-secondary tnum">
                  {formatIDR(product.revenue)}
                </span>
                <span
                  className={cn(
                    "flex shrink-0 items-center justify-end gap-0.5 text-[11px] font-semibold whitespace-nowrap tnum",
                    up ? "text-status-good" : "text-status-critical",
                  )}
                >
                  <Arrow aria-hidden className="size-3" />
                  {Math.abs(product.delta).toFixed(1)}%
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function Message({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid flex-1 place-items-center px-4 py-10">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
