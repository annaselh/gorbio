import { Modal } from "@/shared/ui/Modal";
import { Badge } from "@/shared/ui/Badge";
import { OrderLinesTable } from "@/shared/ui/OrderLinesTable";
import { formatDate } from "@/shared/format";
import { ORDER_STATUS_TONE, useSalesOrder } from "../data";

export function SalesOrderDialog({
  orderID,
  onClose,
}: {
  orderID: string;
  onClose: () => void;
}) {
  // The list endpoint omits lines, so the detail view fetches the order again.
  const { data: order, isPending, isError } = useSalesOrder(orderID);

  return (
    <Modal
      title={order ? `Sales order ${order.number}` : "Sales order"}
      description={order?.customer_name}
      onClose={onClose}
      size="lg"
    >
      {isPending ? (
        <p className="py-8 text-center text-sm text-ink-secondary">Loading…</p>
      ) : isError || !order ? (
        <p className="py-8 text-center text-sm text-ink-secondary">
          This order could not be loaded.
        </p>
      ) : (
        <div className="space-y-4">
          <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm sm:grid-cols-3">
            <Item label="Status">
              <Badge tone={ORDER_STATUS_TONE[order.status]}>{order.status}</Badge>
            </Item>
            <Item label="Order date">{formatDate(order.order_date)}</Item>
            <Item label="Currency">{order.currency}</Item>
          </dl>

          <OrderLinesTable lines={order.lines ?? []} totals={order} />

          {order.notes ? (
            <div>
              <p className="text-xs font-medium text-ink-secondary">Notes</p>
              <p className="mt-1 text-sm whitespace-pre-wrap text-ink">
                {order.notes}
              </p>
            </div>
          ) : null}
        </div>
      )}
    </Modal>
  );
}

function Item({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <dt className="text-xs font-medium text-ink-secondary">{label}</dt>
      <dd className="mt-0.5 text-sm text-ink">{children}</dd>
    </div>
  );
}
