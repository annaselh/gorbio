import { useState } from "react";
import { Plus } from "lucide-react";
import { PageHeader } from "@/core/Shell";
import { useAuth } from "@/core/auth";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Badge, Dot } from "@/shared/ui/Badge";
import { Button } from "@/shared/ui/Button";
import { Pagination } from "@/shared/ui/Pagination";
import { formatDate, formatIDR } from "@/shared/format";
import {
  PURCHASE_STATUS_TONE,
  usePurchaseOrders,
  useUpdatePurchaseStatus,
  type PurchaseOrder,
} from "../data";
import { NewPurchaseOrderDialog } from "../components/NewPurchaseOrderDialog";

const PAGE_SIZE = 10;

export default function PurchaseOrders() {
  const [page, setPage] = useState(1);
  const [creating, setCreating] = useState(false);
  const { hasPermission } = useAuth();
  const canManage = hasPermission("procurement.manage");
  const updateStatus = useUpdatePurchaseStatus();

  const { data, isPending, isError, isPlaceholderData } = usePurchaseOrders({
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
  });

  const rows = data?.data ?? [];
  const total = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const th =
    "px-2.5 py-2.5 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <>
      <PageHeader
        title="Purchase Orders"
        subtitle="What you have ordered from suppliers."
      />

      <Card className="flex flex-col">
        <CardHeader
          title="All Purchase Orders"
          action={
            canManage ? (
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
          <Message>Loading purchase orders…</Message>
        ) : isError ? (
          <Message>Purchase orders are unavailable right now.</Message>
        ) : rows.length === 0 ? (
          <Message>No purchase orders yet.</Message>
        ) : (
          <>
            <div
              className="scrollbar-slim overflow-x-auto"
              aria-busy={isPlaceholderData}
              style={{ opacity: isPlaceholderData ? 0.6 : 1 }}
            >
              <table className="w-full min-w-[680px] border-collapse">
                <thead>
                  <tr className="border-y border-hairline-soft bg-hairline-soft/40">
                    <th scope="col" className={th}>PO No.</th>
                    <th scope="col" className={th}>Vendor</th>
                    <th scope="col" className={th}>Order Date</th>
                    <th scope="col" className={th}>Expected</th>
                    <th scope="col" className={th}>Total</th>
                    <th scope="col" className={th}>Status</th>
                    {canManage && <th scope="col" className={th}>Actions</th>}
                  </tr>
                </thead>
                <tbody className="divide-y divide-hairline-soft">
                  {rows.map((order) => {
                    const tone = PURCHASE_STATUS_TONE[order.status];
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
                        <td className="max-w-[180px] truncate px-2.5 py-3 text-xs text-ink-secondary">
                          {order.vendor_name}
                        </td>
                        <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
                          {formatDate(order.order_date)}
                        </td>
                        <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
                          {order.expected_date ? formatDate(order.expected_date) : "—"}
                        </td>
                        <td className="px-2.5 py-3 text-xs font-medium whitespace-nowrap text-ink tnum">
                          {formatIDR(order.total)}
                        </td>
                        <td className="px-2.5 py-3">
                          <Badge tone={tone}>{order.status}</Badge>
                        </td>
                        {canManage && (
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
                pageSize={PAGE_SIZE}
                onPageChange={(p) => setPage(Math.min(Math.max(p, 1), pageCount))}
              />
            </div>
          </>
        )}
      </Card>

      {creating && (
        <NewPurchaseOrderDialog onClose={() => setCreating(false)} />
      )}
    </>
  );
}

/**
 * Offers only the transitions the service accepts: Draft -> Confirmed ->
 * Received, with cancellation allowed until receipt. A received order is
 * terminal because stock has already moved.
 */
function StatusActions({
  order,
  pending,
  onChange,
}: {
  order: PurchaseOrder;
  pending: boolean;
  onChange: (status: PurchaseOrder["status"]) => void;
}) {
  if (order.status === "Received" || order.status === "Cancelled") {
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
      {order.status === "Confirmed" && (
        <Button
          variant="outline"
          className="px-2.5 py-1 text-xs"
          disabled={pending}
          onClick={() => onChange("Received")}
        >
          Receive
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

function Message({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid place-items-center px-4 py-12">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
