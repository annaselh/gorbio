import { formatIDR } from "../format";

/** The line shape both sales and procurement return. */
export interface OrderLineView {
  id: string;
  sku: string;
  description: string;
  quantity: number;
  unit_price: number;
  line_total: number;
}

export interface OrderTotals {
  subtotal: number;
  discount_total: number;
  total: number;
}

/**
 * Read-only line table with the money summary. Shared so a sales order and a
 * purchase order present their arithmetic identically - the two use the same
 * Recalculate rule on the server, and disagreeing here would suggest they do
 * not.
 */
export function OrderLinesTable({
  lines,
  totals,
}: {
  lines: OrderLineView[];
  totals: OrderTotals;
}) {
  const th =
    "px-2.5 py-2 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <div className="space-y-3">
      <div className="scrollbar-slim overflow-x-auto rounded-lg border border-hairline">
        <table className="w-full min-w-[520px] border-collapse">
          <thead>
            <tr className="border-b border-hairline-soft bg-hairline-soft/40">
              <th scope="col" className={th}>SKU</th>
              <th scope="col" className={th}>Description</th>
              <th scope="col" className={`${th} text-right`}>Qty</th>
              <th scope="col" className={`${th} text-right`}>Unit price</th>
              <th scope="col" className={`${th} text-right`}>Total</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-hairline-soft">
            {lines.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-2.5 py-6 text-center text-sm text-ink-secondary">
                  This order has no lines.
                </td>
              </tr>
            ) : (
              lines.map((line) => (
                <tr key={line.id}>
                  <th
                    scope="row"
                    className="px-2.5 py-2.5 text-left text-xs font-medium whitespace-nowrap text-ink"
                  >
                    {line.sku}
                  </th>
                  <td className="max-w-[240px] truncate px-2.5 py-2.5 text-xs text-ink-secondary">
                    {line.description}
                  </td>
                  <td className="px-2.5 py-2.5 text-right text-xs whitespace-nowrap text-ink-secondary tnum">
                    {line.quantity}
                  </td>
                  <td className="px-2.5 py-2.5 text-right text-xs whitespace-nowrap text-ink-secondary tnum">
                    {formatIDR(line.unit_price)}
                  </td>
                  <td className="px-2.5 py-2.5 text-right text-xs font-medium whitespace-nowrap text-ink tnum">
                    {formatIDR(line.line_total)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <dl className="ml-auto w-full max-w-[260px] space-y-1.5 text-sm">
        <Row label="Subtotal" value={formatIDR(totals.subtotal)} />
        {/* Only shown when non-zero: a permanent "Discount Rp 0" line is noise
            on the many orders that never had one. */}
        {totals.discount_total > 0 && (
          <Row label="Discount" value={`− ${formatIDR(totals.discount_total)}`} />
        )}
        <div className="flex items-baseline justify-between gap-4 border-t border-hairline-soft pt-1.5">
          <dt className="font-medium text-ink">Total</dt>
          <dd className="font-semibold text-ink tnum">{formatIDR(totals.total)}</dd>
        </div>
      </dl>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4">
      <dt className="text-ink-secondary">{label}</dt>
      <dd className="text-ink tnum">{value}</dd>
    </div>
  );
}
