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
import { useCreatePurchaseOrder, useVendors } from "../data";

function today() {
  return new Date().toISOString().slice(0, 10);
}

export function NewPurchaseOrderDialog({ onClose }: { onClose: () => void }) {
  const create = useCreatePurchaseOrder();
  // Only active vendors: the service refuses an order against an inactive one,
  // so offering them would be a dead end.
  const { data: vendorList, isPending: vendorsLoading } = useVendors({
    limit: 200,
    offset: 0,
  });
  const vendors = (vendorList?.data ?? []).filter((v) => v.status === "Active");

  const [vendorID, setVendorID] = useState("");
  const [orderDate, setOrderDate] = useState(today());
  const [expectedDate, setExpectedDate] = useState("");
  const [notes, setNotes] = useState("");
  const [lines, setLines] = useState<DraftLine[]>([emptyLine()]);
  const [localError, setLocalError] = useState<string | null>(null);

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

    create.mutate(
      {
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
      },
      { onSuccess: onClose },
    );
  }

  const error =
    localError ??
    (create.error instanceof ApiError
      ? create.error.message
      : create.error
        ? "Could not create the purchase order."
        : null);

  return (
    <Modal
      title="New purchase order"
      description="Created as a draft. Confirm it once the supplier accepts."
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
              disabled={create.isPending || vendorsLoading}
              className={FIELD_CLASS}
            >
              <option value="">
                {vendorsLoading ? "Loading…" : "Select a vendor"}
              </option>
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
              disabled={create.isPending}
              className={FIELD_CLASS}
            />
          </Field>
          <Field label="Expected" hint="(optional)">
            <input
              type="date"
              value={expectedDate}
              onChange={(e) => setExpectedDate(e.target.value)}
              disabled={create.isPending}
              className={FIELD_CLASS}
            />
          </Field>
        </div>

        {!vendorsLoading && vendors.length === 0 && (
          <FormError>
            No active vendors yet. Create one before raising a purchase order.
          </FormError>
        )}

        <LineItemsEditor
          lines={lines}
          onChange={setLines}
          disabled={create.isPending}
        />

        <Field label="Notes" hint="(optional)">
          <textarea
            rows={2}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            disabled={create.isPending}
            className={FIELD_CLASS}
          />
        </Field>

        {error ? <FormError>{error}</FormError> : null}

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose} disabled={create.isPending}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={create.isPending || vendors.length === 0}
          >
            {create.isPending ? "Creating…" : "Create draft"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
