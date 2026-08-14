import { useState } from "react";
import type { FormEvent } from "react";
import { ApiError } from "@/core/apiClient";
import { Button } from "@/shared/ui/Button";
import { Field, FIELD_CLASS, FormError, Modal } from "@/shared/ui/Modal";
import { useCreateVendor, useUpdateVendor, type Vendor } from "../data";

/** Doubles as create and edit: `vendor` present means edit. */
export function VendorDialog({
  vendor,
  onClose,
}: {
  vendor?: Vendor;
  onClose: () => void;
}) {
  const create = useCreateVendor();
  const update = useUpdateVendor();
  const mutation = vendor ? update : create;

  const [name, setName] = useState(vendor?.name ?? "");
  const [code, setCode] = useState(vendor?.code ?? "");
  const [email, setEmail] = useState(vendor?.email ?? "");
  const [phone, setPhone] = useState(vendor?.phone ?? "");
  const [address, setAddress] = useState(vendor?.address ?? "");
  const [taxID, setTaxID] = useState(vendor?.tax_id ?? "");
  const [paymentTerm, setPaymentTerm] = useState(
    vendor?.payment_term_days ?? 30,
  );

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const input = {
      name,
      email: email.trim() || undefined,
      phone: phone.trim() || undefined,
      address: address.trim() || undefined,
      tax_id: taxID.trim() || undefined,
      payment_term_days: paymentTerm,
    };

    if (vendor) {
      update.mutate({ id: vendor.id, input }, { onSuccess: onClose });
    } else {
      create.mutate(
        // A blank code lets the service allocate V-0001; sending "" would be
        // rejected as an explicit empty code.
        { ...input, code: code.trim() || undefined },
        { onSuccess: onClose },
      );
    }
  }

  const error =
    mutation.error instanceof ApiError
      ? mutation.error.message
      : mutation.error
        ? "Could not save the vendor."
        : null;

  return (
    <Modal
      title={vendor ? "Edit vendor" : "New vendor"}
      description={
        vendor
          ? "Existing purchase orders keep the name they were raised with."
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
            placeholder="PT. Sumber Makmur"
          />
        </Field>

        {!vendor && (
          <Field label="Code" hint="(optional)">
            <input
              type="text"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={mutation.isPending}
              className={FIELD_CLASS}
              placeholder="V-0001"
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
          <Field label="Payment term" hint="(days)">
            <input
              type="number"
              min={0}
              value={paymentTerm}
              onChange={(e) => setPaymentTerm(Number(e.target.value) || 0)}
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
            {mutation.isPending ? "Saving…" : vendor ? "Save changes" : "Create vendor"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
