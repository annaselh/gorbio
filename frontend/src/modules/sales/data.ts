import type { StatusTone } from "@/shared/ui/Badge";

export type OrderStatus = "Confirmed" | "Draft" | "Cancelled";

export interface SalesOrder {
  id: string;
  number: string;
  customer: string;
  date: string;
  total: number;
  status: OrderStatus;
}

export const ORDER_STATUS_TONE: Record<OrderStatus, StatusTone> = {
  Confirmed: "good",
  Draft: "info",
  Cancelled: "critical",
};

const CUSTOMERS = [
  "PT. Maju Bersama",
  "CV. Sejahtera Abadi",
  "PT. Sukses Makmur",
  "PT. Cahaya Sentosa",
  "UD. Berkah Abadi",
  "PT. Nusantara Jaya",
  "CV. Mitra Sejati",
  "PT. Bumi Persada",
];

const STATUSES: OrderStatus[] = [
  "Confirmed",
  "Confirmed",
  "Draft",
  "Confirmed",
  "Cancelled",
];

const TOTALS = [25_000_000, 18_500_000, 12_750_000, 9_800_000, 6_500_000];

/** 125 rows so the mockup's "1–5 of 125 · … · 25" pagination is real. */
export const salesOrders: SalesOrder[] = Array.from({ length: 125 }, (_, i) => {
  const n = 123 - i;
  const d = new Date(Date.UTC(2024, 4, 31 - Math.floor(i / 2)));
  return {
    id: `so-${n}`,
    number: `SO-${String(Math.abs(n)).padStart(6, "0")}`,
    customer: CUSTOMERS[i % CUSTOMERS.length],
    date: d.toISOString(),
    total: TOTALS[i % TOTALS.length],
    status: STATUSES[i % STATUSES.length],
  };
});
