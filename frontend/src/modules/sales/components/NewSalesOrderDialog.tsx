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
import { useCreateSalesOrder } from "../data";

function today() {
  return new Date().toISOString().slice(0, 10);
}

export function NewSalesOrderDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateSalesOrder();
  const [customer, setCustomer] = useState("");
  const [orderDate, setOrderDate] = useState(today());
  const [notes, setNotes] = useState("");
  const [lines, setLines] = useState<DraftLine[]>([emptyLine()]);
  const [localError, setLocalError] = useState<string | null>(null);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    // Checked here only to save a round trip; the service validates the same
    // rules and remains the authority.
    const lineError = validateLines(lines);
    if (lineError) {
      setLocalError(lineError);
      return;
    }
    setLocalError(null);

    create.mutate(
      {
        customer_name: customer,
        // Send an instant, not a bare date: the column is a timestamp and a
        // date-only string would be read as midnight UTC and could land in the
        // previous day for anyone east of Greenwich.
        order_date: new Date(`${orderDate}T00:00:00`).toISOString(),
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
        ? "Could not create the sales order."
        : null);

  return (
    <Modal
      title="New sales order"
      description="Created as a draft. Confirm it once the customer agrees."
      onClose={onClose}
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Customer">
            <input
              type="text"
              required
              value={customer}
              onChange={(e) => setCustomer(e.target.value)}
              disabled={create.isPending}
              className={FIELD_CLASS}
              placeholder="PT. Maju Bersama"
            />
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
        </div>

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
          <Button type="submit" variant="primary" disabled={create.isPending}>
            {create.isPending ? "Creating…" : "Create draft"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
