import { useState } from "react";
import type { FormEvent } from "react";
import { ApiError } from "@/core/apiClient";
import { Button } from "@/shared/ui/Button";
import { Field, FIELD_CLASS, FormError, Modal } from "@/shared/ui/Modal";
import {
  LineItemsEditor,
  emptyLine,
  validateLines,
  type DraftLine,
} from "@/shared/ui/LineItemsEditor";
import { useCustomers } from "@/modules/crm/data";
import {
  useCreateSalesOrder,
  useSalesOrder,
  useUpdateSalesOrder,
  type SalesOrder,
} from "../data";

function today() {
  return new Date().toISOString().slice(0, 10);
}

const WALK_IN = "__walk_in__";

/**
 * Creating and editing an order use one form, because they are the same form:
 * the server's update endpoint takes the order as it should now read, not a
 * patch, so the fields it accepts are exactly the fields create accepts. Two
 * dialogs would be two places for the validation to drift.
 *
 * Passing an orderID switches it to editing and loads the lines, which the list
 * response does not carry.
 */
export function SalesOrderFormDialog({
  orderID,
  onClose,
}: {
  orderID?: string;
  onClose: () => void;
}) {
  const detail = useSalesOrder(orderID ?? "");

  if (!orderID) return <SalesOrderForm onClose={onClose} />;

  if (detail.isPending || detail.isError || !detail.data) {
    return (
      <Modal title="Edit sales order" onClose={onClose} size="lg">
        <p className="text-sm text-ink-secondary">
          {detail.isPending ? "Loading the order…" : "Could not load this order."}
        </p>
      </Modal>
    );
  }

  return <SalesOrderForm order={detail.data} onClose={onClose} />;
}

function SalesOrderForm({
  order,
  onClose,
}: {
  order?: SalesOrder;
  onClose: () => void;
}) {
  const editing = Boolean(order);
  const create = useCreateSalesOrder();
  const update = useUpdateSalesOrder();
  const mutation = editing ? update : create;

  // Only active customers: the service refuses an order against an inactive
  // one, so offering them would be a dead end.
  const { data: customerList } = useCustomers({ limit: 200, offset: 0 });
  const customers = (customerList?.data ?? []).filter(
    (c) => c.status === "Active",
  );

  const [customerID, setCustomerID] = useState(order?.customer_id ?? WALK_IN);
  const [customer, setCustomer] = useState(
    order?.customer_id ? "" : (order?.customer_name ?? ""),
  );
  const [orderDate, setOrderDate] = useState(
    order ? order.order_date.slice(0, 10) : today(),
  );
  const [notes, setNotes] = useState(order?.notes ?? "");
  const [lines, setLines] = useState<DraftLine[]>(
    order?.lines?.length
      ? order.lines.map((line) => ({
          sku: line.sku,
          description: line.description,
          quantity: line.quantity,
          unitPrice: line.unit_price,
        }))
      : [emptyLine()],
  );
  const [localError, setLocalError] = useState<string | null>(null);

  // An order linked to a customer who has since been deactivated still has to
  // show that link, or reopening the order would silently turn it into a
  // walk-in sale.
  const linkedMissing =
    order?.customer_id && !customers.some((c) => c.id === order.customer_id);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    // Checked here only to save a round trip; the service validates the same
    // rules and remains the authority.
    const linked = customerID !== WALK_IN;
    if (!linked && !customer.trim()) {
      setLocalError("Choose a customer or enter a name.");
      return;
    }
    const lineError = validateLines(lines);
    if (lineError) {
      setLocalError(lineError);
      return;
    }
    setLocalError(null);

    const payload = {
      // When a CRM record is chosen the server resolves the name from it, so
      // sending a name too would only invite the two to disagree.
      customer_id: linked ? customerID : undefined,
      customer_name: linked ? undefined : customer,
      // Send an instant, not a bare date: the column is a timestamp and a
      // date-only string would be read as midnight UTC and could land in the
      // previous day for anyone east of Greenwich.
      order_date: new Date(`${orderDate}T00:00:00`).toISOString(),
      notes: notes.trim() || undefined,
      lines,
    };

    if (order) {
      update.mutate({ id: order.id, ...payload }, { onSuccess: onClose });
      return;
    }
    create.mutate(payload, { onSuccess: onClose });
  }

  const error =
    localError ??
    (mutation.error instanceof ApiError
      ? mutation.error.message
      : mutation.error
        ? `Could not ${editing ? "save" : "create"} the sales order.`
        : null);

  return (
    <Modal
      title={editing ? `Edit ${order?.number}` : "New sales order"}
      description={
        editing
          ? "Only draft orders can be changed. Confirming one makes it a record."
          : "Created as a draft. Confirm it once the customer agrees."
      }
      onClose={onClose}
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Customer">
            <select
              value={customerID}
              onChange={(e) => setCustomerID(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
            >
              {/* Walk-in stays available so a one-off sale does not force a
                  CRM record to be created first. */}
              <option value={WALK_IN}>Walk-in (type a name)</option>
              {linkedMissing && order?.customer_id && (
                <option value={order.customer_id}>
                  {order.customer_name} (inactive)
                </option>
              )}
              {customers.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.code} — {c.name}
                </option>
              ))}
            </select>
          </Field>
          {customerID === WALK_IN && (
            <Field label="Customer name">
              <input
                type="text"
                required
                value={customer}
                onChange={(e) => setCustomer(e.target.value)}
                disabled={mutation.isPending}
                className={FIELD_CLASS}
                placeholder="PT. Maju Bersama"
              />
            </Field>
          )}
          <Field label="Order date">
            <input
              type="date"
              required
              value={orderDate}
              onChange={(e) => setOrderDate(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
            />
          </Field>
        </div>

        <LineItemsEditor
          lines={lines}
          onChange={setLines}
          disabled={mutation.isPending}
        />

        <Field label="Notes" hint="(optional)">
          <textarea
            rows={2}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            disabled={mutation.isPending}
            className={FIELD_CLASS}
          />
        </Field>

        {error ? <FormError>{error}</FormError> : null}

        <div className="flex justify-end gap-2">
          <Button
            variant="outline"
            onClick={onClose}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={mutation.isPending}>
            {mutation.isPending
              ? editing
                ? "Saving…"
                : "Creating…"
              : editing
                ? "Save changes"
                : "Create draft"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
