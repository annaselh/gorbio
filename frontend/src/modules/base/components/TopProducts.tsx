import { ArrowUp } from "lucide-react";
import { Card, CardHeader } from "@/shared/ui/Card";
import { ViewAllButton } from "@/shared/ui/Button";
import { formatIDR } from "@/shared/format";
import { topProducts } from "../data";

export function TopProducts() {
  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Top Products" action={<ViewAllButton />} />
      <ul className="flex-1 divide-y divide-hairline-soft px-4 pb-2">
        {topProducts.map((p) => (
          <li key={p.id} className="flex items-center gap-2 py-3">
            <span
              aria-hidden
              className="grid size-7 shrink-0 place-items-center rounded-lg bg-hairline-soft text-xs"
            >
              {p.emoji}
            </span>
            <span className="min-w-0 flex-1 truncate text-xs font-medium text-ink">
              {p.name}
            </span>
            {/* shrink-0 so the figures hold their width and the name is the
                only thing that ever truncates. */}
            <span className="shrink-0 text-[11px] whitespace-nowrap text-ink-secondary tnum">
              {formatIDR(p.revenue)}
            </span>
            <span className="flex shrink-0 items-center justify-end gap-0.5 text-[11px] font-semibold whitespace-nowrap text-status-good tnum">
              <ArrowUp aria-hidden className="size-3" />
              {p.delta.toFixed(1)}%
            </span>
          </li>
        ))}
      </ul>
    </Card>
  );
}
