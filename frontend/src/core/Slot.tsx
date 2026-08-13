import { registry } from "./modules";

/**
 * Extension point — the UI mirror of the backend's `_inherit`. A module owning
 * a form declares `<Slot name="partner.form.extra" />`; other modules fill it
 * without either side importing the other.
 */
export function Slot({ name, ctx }: { name: string; ctx?: unknown }) {
  const comps = registry.slots[name] ?? [];
  if (comps.length === 0) return null;
  return (
    <>
      {comps.map((C, i) => (
        <C key={`${name}-${i}`} ctx={ctx} />
      ))}
    </>
  );
}
