import { useState } from "react";
import type { FormEvent } from "react";
import { ApiError } from "@/core/apiClient";
import { Button } from "@/shared/ui/Button";
import { Field, FIELD_CLASS, FormError, Modal } from "@/shared/ui/Modal";
import { useCreateCustomer, useUpdateCustomer, type Customer } from "../data";

/** Doubles as create and edit: `customer` present means edit. */
export function CustomerDialog({
  customer,
  onClose,
}: {
  customer?: Customer;
  onClose: () => void;
}) {
  const create = useCreateCustomer();
  const update = useUpdateCustomer();
  const mutation = customer ? update : create;

  const [name, setName] = useState(customer?.name ?? "");
  const [code, setCode] = useState(customer?.code ?? "");
  const [email, setEmail] = useState(customer?.email ?? "");
  const [phone, setPhone] = useState(customer?.phone ?? "");
  const [address, setAddress] = useState(customer?.address ?? "");
  const [taxID, setTaxID] = useState(customer?.tax_id ?? "");
  const [creditTerm, setCreditTerm] = useState(customer?.credit_term_days ?? 30);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const input = {
      name,
      email: email.trim() || undefined,
      phone: phone.trim() || undefined,
      address: address.trim() || undefined,
      tax_id: taxID.trim() || undefined,
      credit_term_days: creditTerm,
    };

    if (customer) {
      update.mutate({ id: customer.id, input }, { onSuccess: onClose });
    } else {
      create.mutate(
        // A blank code lets the service allocate C-0001; sending "" would be
        // read as an explicit empty code.
        { ...input, code: code.trim() || undefined },
        { onSuccess: onClose },
      );
    }
  }

  const error =
    mutation.error instanceof ApiError
      ? mutation.error.message
      : mutation.error
        ? "Could not save the customer."
        : null;

  return (
    <Modal
      title={customer ? "Edit customer" : "New customer"}
      description={
        customer
          ? "Existing sales orders keep the name they were raised with."
          : "Leave the code blank to have one allocated."
      }
      onClose={onClose}
    >
      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        <Field label="Name">
          <input
            type="text"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={mutation.isPending}
            className={FIELD_CLASS}
            placeholder="PT. Maju Bersama"
          />
        </Field>

        {!customer && (
          <Field label="Code" hint="(optional)">
            <input
              type="text"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
              placeholder="C-0001"
            />
          </Field>
        )}

        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Email" hint="(optional)">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
            />
          </Field>
          <Field label="Phone" hint="(optional)">
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
            />
          </Field>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Tax ID" hint="(optional)">
            <input
              type="text"
              value={taxID}
              onChange={(e) => setTaxID(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
            />
          </Field>
          <Field label="Credit term" hint="(days)">
            <input
              type="number"
              min={0}
              value={creditTerm}
              onChange={(e) => setCreditTerm(Number(e.target.value) || 0)}
              disabled={mutation.isPending}
              className={`${FIELD_CLASS} tnum`}
            />
          </Field>
        </div>

        <Field label="Address" hint="(optional)">
          <textarea
            rows={2}
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            disabled={mutation.isPending}
            className={FIELD_CLASS}
          />
        </Field>

        {error ? <FormError>{error}</FormError> : null}

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose} disabled={mutation.isPending}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={mutation.isPending}>
            {mutation.isPending
              ? "Saving…"
              : customer
                ? "Save changes"
                : "Create customer"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
