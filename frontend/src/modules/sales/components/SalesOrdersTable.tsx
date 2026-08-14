import { useState } from "react";
import { Plus } from "lucide-react";
import { useAuth } from "@/core/auth";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Button } from "@/shared/ui/Button";
import { Badge, Dot } from "@/shared/ui/Badge";
import { Pagination } from "@/shared/ui/Pagination";
import { formatDate, formatIDR } from "@/shared/format";
import {
  ORDER_STATUS_TONE,
  useSalesOrders,
  useUpdateSalesOrderStatus,
  type SalesOrder,
} from "../data";
import { NewSalesOrderDialog } from "./NewSalesOrderDialog";

const PAGE_SIZE = 5;

export function SalesOrdersTable({
  title = "Sales Orders",
  pageSize = PAGE_SIZE,
  /** The dashboard widget is read-only; the full page enables the actions. */
  showActions = false,
}: {
  title?: string;
  pageSize?: number;
  showActions?: boolean;
}) {
  const [page, setPage] = useState(1);
  const [creating, setCreating] = useState(false);
  const { hasPermission } = useAuth();
  const canManage = hasPermission("sales.manage");
  const updateStatus = useUpdateSalesOrderStatus();

  const { data, isPending, isError, isPlaceholderData } = useSalesOrders({
    limit: pageSize,
    offset: (page - 1) * pageSize,
  });

  const rows = data?.data ?? [];
  const total = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  const th =
    "px-2.5 py-2.5 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  const showActionColumn = showActions && canManage;

  return (
    <Card className="flex h-full flex-col">
      <CardHeader
        title={title}
        action={
          showActions && canManage ? (
            <Button
              variant="primary"
              className="px-3 py-1.5 text-xs"
              onClick={() => setCreating(true)}
            >
              <Plus aria-hidden className="size-3.5" />
              New order
            </Button>
          ) : undefined
        }
      />

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
                  {showActionColumn && (
                    <th scope="col" className={th}>Actions</th>
                  )}
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
                      {showActionColumn && (
                        <td className="px-2.5 py-3">
                          <StatusActions
                            order={order}
                            pending={updateStatus.isPending}
                            onChange={(status) =>
                              updateStatus.mutate({ id: order.id, status })
                            }
                          />
                        </td>
                      )}
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

      {creating && <NewSalesOrderDialog onClose={() => setCreating(false)} />}
    </Card>
  );
}

/**
 * Offers only the transitions the service accepts: a draft may be confirmed or
 * cancelled, a confirmed order may still be cancelled, and a cancelled order is
 * terminal. Rendering a button the server would reject just teaches users to
 * ignore errors.
 */
function StatusActions({
  order,
  pending,
  onChange,
}: {
  order: SalesOrder;
  pending: boolean;
  onChange: (status: SalesOrder["status"]) => void;
}) {
  if (order.status === "Cancelled") {
    return <span className="text-xs text-ink-muted">—</span>;
  }

  return (
    <span className="flex gap-1.5">
      {order.status === "Draft" && (
        <Button
          variant="outline"
          className="px-2.5 py-1 text-xs"
          disabled={pending}
          onClick={() => onChange("Confirmed")}
        >
          Confirm
        </Button>
      )}
      <Button
        variant="ghost"
        className="px-2.5 py-1 text-xs"
        disabled={pending}
        onClick={() => onChange("Cancelled")}
      >
        Cancel
      </Button>
    </span>
  );
}

function TableMessage({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid flex-1 place-items-center px-4 py-10">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
