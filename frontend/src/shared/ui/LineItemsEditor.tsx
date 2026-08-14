import { Plus, Trash2 } from "lucide-react";
import { Button } from "./Button";
import { FIELD_CLASS } from "./Modal";
import { formatIDR } from "../format";

/**
 * A draft order line. `unitPrice` is whole rupiah — the same integer the API
 * stores. IDR has no subdivision in practice, so there is deliberately no
 * hundredths conversion here; adding one would inflate every total by 100.
 */
export interface DraftLine {
  sku: string;
  description: string;
  quantity: number;
  unitPrice: number;
}

export function emptyLine(): DraftLine {
  return { sku: "", description: "", quantity: 1, unitPrice: 0 };
}

export function lineTotal(line: DraftLine): number {
  return line.quantity * line.unitPrice;
}

export function linesSubtotal(lines: DraftLine[]): number {
  return lines.reduce((sum, line) => sum + lineTotal(line), 0);
}

/** Mirrors the server-side validation in the sales and procurement services. */
export function validateLines(lines: DraftLine[]): string | null {
  if (lines.length === 0) return "Add at least one line.";
  for (const [index, line] of lines.entries()) {
    if (!line.sku.trim() || !line.description.trim()) {
      return `Line ${index + 1} needs a SKU and a description.`;
    }
    if (line.quantity <= 0) {
      return `Line ${index + 1} quantity must be greater than zero.`;
    }
    if (line.unitPrice < 0) {
      return `Line ${index + 1} unit price must not be negative.`;
    }
  }
  return null;
}

export function LineItemsEditor({
  lines,
  onChange,
  disabled,
}: {
  lines: DraftLine[];
  onChange: (lines: DraftLine[]) => void;
  disabled?: boolean;
}) {
  function update(index: number, patch: Partial<DraftLine>) {
    onChange(lines.map((line, i) => (i === index ? { ...line, ...patch } : line)));
  }

  function remove(index: number) {
    onChange(lines.filter((_, i) => i !== index));
  }

  const th =
    "px-2 py-2 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <div className="space-y-3">
      <div className="scrollbar-slim overflow-x-auto rounded-lg border border-hairline">
        <table className="w-full min-w-[560px] border-collapse">
          <thead>
            <tr className="border-b border-hairline-soft bg-hairline-soft/40">
              <th scope="col" className={th}>SKU</th>
              <th scope="col" className={th}>Description</th>
              <th scope="col" className={`${th} w-20`}>Qty</th>
              <th scope="col" className={`${th} w-36`}>Unit price</th>
              <th scope="col" className={`${th} w-32 text-right`}>Total</th>
              <th scope="col" className="w-10">
                <span className="sr-only">Remove</span>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-hairline-soft">
            {lines.map((line, index) => (
              <tr key={index}>
                <td className="p-1.5">
                  <input
                    type="text"
                    value={line.sku}
                    onChange={(e) => update(index, { sku: e.target.value })}
                    disabled={disabled}
                    aria-label={`Line ${index + 1} SKU`}
                    className={FIELD_CLASS}
                    placeholder="SKU-001"
                  />
                </td>
                <td className="p-1.5">
                  <input
                    type="text"
                    value={line.description}
                    onChange={(e) => update(index, { description: e.target.value })}
                    disabled={disabled}
                    aria-label={`Line ${index + 1} description`}
                    className={FIELD_CLASS}
                    placeholder="Item description"
                  />
                </td>
                <td className="p-1.5">
                  <input
                    type="number"
                    min={1}
                    value={line.quantity}
                    onChange={(e) =>
                      update(index, { quantity: Number(e.target.value) || 0 })
                    }
                    disabled={disabled}
                    aria-label={`Line ${index + 1} quantity`}
                    className={`${FIELD_CLASS} tnum`}
                  />
                </td>
                <td className="p-1.5">
                  <input
                    type="number"
                    min={0}
                    step={1000}
                    value={line.unitPrice}
                    onChange={(e) =>
                      update(index, { unitPrice: Number(e.target.value) || 0 })
                    }
                    disabled={disabled}
                    aria-label={`Line ${index + 1} unit price in rupiah`}
                    className={`${FIELD_CLASS} tnum`}
                  />
                </td>
                <td className="px-2 py-1.5 text-right text-xs font-medium whitespace-nowrap text-ink tnum">
                  {formatIDR(lineTotal(line))}
                </td>
                <td className="px-1 py-1.5">
                  <button
                    type="button"
                    onClick={() => remove(index)}
                    // The services reject an order with no lines, so the last
                    // row cannot be removed - only edited.
                    disabled={disabled || lines.length === 1}
                    aria-label={`Remove line ${index + 1}`}
                    className="grid size-8 cursor-pointer place-items-center rounded-lg text-ink-secondary transition-colors hover:bg-hairline-soft hover:text-status-critical disabled:cursor-not-allowed disabled:opacity-30"
                  >
                    <Trash2 className="size-4" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between gap-3">
        <Button
          variant="outline"
          onClick={() => onChange([...lines, emptyLine()])}
          disabled={disabled}
          className="px-3 py-1.5 text-xs"
        >
          <Plus aria-hidden className="size-3.5" />
          Add line
        </Button>

        <p className="text-sm">
          <span className="text-ink-secondary">Subtotal </span>
          <span className="font-semibold text-ink tnum">
            {formatIDR(linesSubtotal(lines))}
          </span>
        </p>
      </div>
    </div>
  );
}
