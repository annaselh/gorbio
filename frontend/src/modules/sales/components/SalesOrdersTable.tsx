import { useState } from "react";
import { ArrowUpDown } from "lucide-react";
import { Card, CardHeader } from "@/shared/ui/Card";
import { ViewAllButton } from "@/shared/ui/Button";
import { Badge, Dot } from "@/shared/ui/Badge";
import { Pagination } from "@/shared/ui/Pagination";
import { formatDate, formatIDR } from "@/shared/format";
import { ORDER_STATUS_TONE, salesOrders } from "../data";

const PAGE_SIZE = 5;

export function SalesOrdersTable({
  title = "Sales Orders",
  pageSize = PAGE_SIZE,
}: {
  title?: string;
  pageSize?: number;
}) {
  const [page, setPage] = useState(1);
  const [desc, setDesc] = useState(true);

  const sorted = [...salesOrders].sort((a, b) =>
    desc ? b.number.localeCompare(a.number) : a.number.localeCompare(b.number),
  );
  const pageCount = Math.ceil(sorted.length / pageSize);
  const rows = sorted.slice((page - 1) * pageSize, page * pageSize);

  const th =
    "px-2.5 py-2.5 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <Card className="flex h-full flex-col">
      <CardHeader title={title} action={<ViewAllButton />} />

      <div className="scrollbar-slim flex-1 overflow-x-auto">
        <table className="w-full min-w-[520px] border-collapse">
          <thead>
            <tr className="border-y border-hairline-soft bg-hairline-soft/40">
              <th scope="col" className={th}>
                <button
                  type="button"
                  onClick={() => {
                    setDesc((d) => !d);
                    setPage(1);
                  }}
                  aria-label={`Sort by order number, currently ${desc ? "descending" : "ascending"}`}
                  className="inline-flex cursor-pointer items-center gap-1.5 hover:text-ink"
                >
                  Order No.
                  <ArrowUpDown aria-hidden className="size-3" />
                </button>
              </th>
              <th scope="col" className={th}>Customer</th>
              <th scope="col" className={th}>Order Date</th>
              <th scope="col" className={th}>Total</th>
              <th scope="col" className={th}>Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-hairline-soft">
            {rows.map((o) => {
              const tone = ORDER_STATUS_TONE[o.status];
              return (
                <tr key={o.id} className="transition-colors hover:bg-hairline-soft/50">
                  <th scope="row" className="px-2.5 py-3 text-left text-xs font-medium whitespace-nowrap text-ink">
                    <span className="inline-flex items-center gap-2.5">
                      <Dot tone={tone} />
                      {o.number}
                    </span>
                  </th>
                  {/* Customer is the flexible column — it truncates so the
                      figures and the status badge never get clipped. */}
                  <td className="max-w-[130px] truncate px-2.5 py-3 text-xs text-ink-secondary">
                    {o.customer}
                  </td>
                  <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
                    {formatDate(o.date)}
                  </td>
                  <td className="px-2.5 py-3 text-xs font-medium whitespace-nowrap text-ink tnum">
                    {formatIDR(o.total)}
                  </td>
                  <td className="px-2.5 py-3">
                    <Badge tone={tone}>{o.status}</Badge>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className="border-t border-hairline-soft">
        <Pagination
          page={page}
          pageCount={pageCount}
          total={sorted.length}
          pageSize={pageSize}
          onPageChange={(p) => setPage(Math.min(Math.max(p, 1), pageCount))}
        />
      </div>
    </Card>
  );
}
