import { useState } from "react";
import { Card, CardHeader } from "@/shared/ui/Card";
import { ViewAllButton } from "@/shared/ui/Button";
import { Badge, Dot } from "@/shared/ui/Badge";
import { Pagination } from "@/shared/ui/Pagination";
import { formatDate, formatIDR } from "@/shared/format";
import { ORDER_STATUS_TONE, useSalesOrders } from "../data";

const PAGE_SIZE = 5;

export function SalesOrdersTable({
  title = "Sales Orders",
  pageSize = PAGE_SIZE,
}: {
  title?: string;
  pageSize?: number;
}) {
  const [page, setPage] = useState(1);

  const { data, isPending, isError, isPlaceholderData } = useSalesOrders({
    limit: pageSize,
    offset: (page - 1) * pageSize,
  });

  const rows = data?.data ?? [];
  const total = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  const th =
    "px-2.5 py-2.5 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <Card className="flex h-full flex-col">
      <CardHeader title={title} action={<ViewAllButton />} />

      {isPending ? (
        <TableMessage>Loading sales orders…</TableMessage>
      ) : isError ? (
        <TableMessage>Sales orders are unavailable right now.</TableMessage>
      ) : rows.length === 0 ? (
        <TableMessage>No sales orders yet.</TableMessage>
      ) : (
        <>
          <div
            className="scrollbar-slim flex-1 overflow-x-auto"
            // Dim while the next page loads so the stale rows do not read as
            // current, without collapsing the layout.
            aria-busy={isPlaceholderData}
            style={{ opacity: isPlaceholderData ? 0.6 : 1 }}
          >
            <table className="w-full min-w-[520px] border-collapse">
              <thead>
                <tr className="border-y border-hairline-soft bg-hairline-soft/40">
                  <th scope="col" className={th}>Order No.</th>
                  <th scope="col" className={th}>Customer</th>
                  <th scope="col" className={th}>Order Date</th>
                  <th scope="col" className={th}>Total</th>
                  <th scope="col" className={th}>Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-hairline-soft">
                {rows.map((order) => {
                  const tone = ORDER_STATUS_TONE[order.status];
                  return (
                    <tr
                      key={order.id}
                      className="transition-colors hover:bg-hairline-soft/50"
                    >
                      <th
                        scope="row"
                        className="px-2.5 py-3 text-left text-xs font-medium whitespace-nowrap text-ink"
                      >
                        <span className="inline-flex items-center gap-2.5">
                          <Dot tone={tone} />
                          {order.number}
                        </span>
                      </th>
                      {/* Customer is the flexible column — it truncates so the
                          figures and the status badge never get clipped. */}
                      <td className="max-w-[130px] truncate px-2.5 py-3 text-xs text-ink-secondary">
                        {order.customer_name}
                      </td>
                      <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
                        {formatDate(order.order_date)}
                      </td>
                      <td className="px-2.5 py-3 text-xs font-medium whitespace-nowrap text-ink tnum">
                        {formatIDR(order.total)}
                      </td>
                      <td className="px-2.5 py-3">
                        <Badge tone={tone}>{order.status}</Badge>
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
              total={total}
              pageSize={pageSize}
              onPageChange={(p) => setPage(Math.min(Math.max(p, 1), pageCount))}
            />
          </div>
        </>
      )}
    </Card>
  );
}

function TableMessage({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid flex-1 place-items-center px-4 py-10">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
