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
import {
  useCreatePurchaseOrder,
  usePurchaseOrder,
  useUpdatePurchaseOrder,
  useVendors,
  type PurchaseOrder,
} from "../data";

function today() {
  return new Date().toISOString().slice(0, 10);
}

/** A timestamp from the API, as the value a date input expects. */
function asDateInput(timestamp?: string) {
  return timestamp ? timestamp.slice(0, 10) : "";
}

/**
 * Raising and editing an order use one form, mirroring the sales side: the
 * update endpoint takes the order as it should now read rather than a patch, so
 * it accepts exactly the fields create accepts.
 *
 * Passing an orderID switches it to editing and loads the lines, which the list
 * response does not carry.
 */
export function PurchaseOrderFormDialog({
  orderID,
  onClose,
}: {
  orderID?: string;
  onClose: () => void;
}) {
  const detail = usePurchaseOrder(orderID ?? "");

  if (!orderID) return <PurchaseOrderForm onClose={onClose} />;

  if (detail.isPending || detail.isError || !detail.data) {
    return (
      <Modal title="Edit purchase order" onClose={onClose} size="lg">
        <p className="text-sm text-ink-secondary">
          {detail.isPending ? "Loading the order…" : "Could not load this order."}
        </p>
      </Modal>
    );
  }

  return <PurchaseOrderForm order={detail.data} onClose={onClose} />;
}

function PurchaseOrderForm({
  order,
  onClose,
}: {
  order?: PurchaseOrder;
  onClose: () => void;
}) {
  const editing = Boolean(order);
  const create = useCreatePurchaseOrder();
  const update = useUpdatePurchaseOrder();
  const mutation = editing ? update : create;

  // Only active vendors: the service refuses an order against an inactive one,
  // so offering them would be a dead end.
  const { data: vendorList, isPending: vendorsLoading } = useVendors({
    limit: 200,
    offset: 0,
  });
  const vendors = (vendorList?.data ?? []).filter((v) => v.status === "Active");

  const [vendorID, setVendorID] = useState(order?.vendor_id ?? "");
  const [orderDate, setOrderDate] = useState(
    order ? asDateInput(order.order_date) : today(),
  );
  const [expectedDate, setExpectedDate] = useState(
    asDateInput(order?.expected_date),
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

  // A draft raised against a vendor who has since been deactivated still has to
  // show that vendor. Leaving the select empty would look like a missing choice
  // rather than a vendor that has been switched off, and the service refuses
  // the save either way - better to say so than to look broken.
  const vendorDeactivated =
    order?.vendor_id && !vendors.some((v) => v.id === order.vendor_id);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!vendorID) {
      setLocalError("Choose a vendor.");
      return;
    }
    const lineError = validateLines(lines);
    if (lineError) {
      setLocalError(lineError);
      return;
    }
    setLocalError(null);

    const payload = {
      vendor_id: vendorID,
      // Send an instant, not a bare date: the column is a timestamp, and a
      // date-only string parsed as midnight UTC can land a day early for
      // anyone east of Greenwich.
      order_date: new Date(`${orderDate}T00:00:00`).toISOString(),
      expected_date: expectedDate
        ? new Date(`${expectedDate}T00:00:00`).toISOString()
        : undefined,
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
        ? `Could not ${editing ? "save" : "create"} the purchase order.`
        : null);

  const noVendorToOrderFrom = !vendorsLoading && vendors.length === 0;

  return (
    <Modal
      title={editing ? `Edit ${order?.number}` : "New purchase order"}
      description={
        editing
          ? "Only draft orders can be changed. Confirming one commits it to the supplier."
          : "Created as a draft. Confirm it once the supplier accepts."
      }
      onClose={onClose}
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-3">
          <Field label="Vendor">
            <select
              required
              value={vendorID}
              onChange={(e) => setVendorID(e.target.value)}
              disabled={mutation.isPending || vendorsLoading}
              className={FIELD_CLASS}
            >
              <option value="">
                {vendorsLoading ? "Loading…" : "Select a vendor"}
              </option>
              {vendorDeactivated && order?.vendor_id && (
                <option value={order.vendor_id}>
                  {order.vendor_name} (inactive)
                </option>
              )}
              {vendors.map((vendor) => (
                <option key={vendor.id} value={vendor.id}>
                  {vendor.code} — {vendor.name}
                </option>
              ))}
            </select>
          </Field>
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
          <Field label="Expected" hint="(optional)">
            <input
              type="date"
              value={expectedDate}
              onChange={(e) => setExpectedDate(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
            />
          </Field>
        </div>

        {vendorDeactivated && (
          <FormError>
            This order's vendor is no longer active. Pick another vendor, or
            reactivate that one, before saving.
          </FormError>
        )}

        {noVendorToOrderFrom && !editing && (
          <FormError>
            No active vendors yet. Create one before raising a purchase order.
          </FormError>
        )}

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
          <Button
            type="submit"
            variant="primary"
            disabled={mutation.isPending || (noVendorToOrderFrom && !editing)}
          >
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
