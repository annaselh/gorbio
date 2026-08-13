/**
 * Dashboard fixtures. The Go backend exposes no endpoints yet (see
 * ../../../app), so these stand in for `GET /api/v1/...` responses and are
 * shaped the way those responses will be. Swap each `useX()` hook for a
 * TanStack Query call against `api` when the backend lands.
 */

const M = 1_000_000;

export interface SeriesPoint {
  date: string;
  value: number;
}

function series(month: string, values: number[]): SeriesPoint[] {
  return values.map((v, i) => ({
    date: `${month}-${String(i + 1).padStart(2, "0")}`,
    value: Math.round(v * M),
  }));
}

export const salesThisMonth = series("2024-05", [
  78, 92, 104, 111, 116, 120, 119, 118, 124, 131, 137, 135, 122, 110, 103, 112,
  128, 121, 101, 100, 118, 140, 133, 120, 134, 152, 143, 132, 148, 163, 178,
]);

export const salesLastMonth = series("2024-04", [
  48, 60, 71, 78, 82, 85, 88, 92, 95, 90, 84, 80, 88, 98, 105, 101, 96, 92, 99,
  108, 115, 112, 108, 105, 113, 120, 125, 122, 118, 120, 128,
]);

export interface Kpi {
  id: string;
  label: string;
  value: string;
  delta: number;
  icon: string;
  tint: string;
  wash: string;
  spark: SeriesPoint[];
}

const spark = (values: number[]): SeriesPoint[] =>
  values.map((v, i) => ({
    date: `2024-05-${String(i * 2 + 1).padStart(2, "0")}`,
    value: v,
  }));

export const kpis: Kpi[] = [
  {
    id: "sales",
    label: "Total Sales",
    value: "Rp 1.250.000.000",
    delta: 12.5,
    icon: "wallet",
    tint: "var(--color-spark-sales)",
    wash: "#FFF2E7",
    spark: spark([42, 46, 44, 52, 49, 58, 55, 63, 61, 70, 68, 79, 76, 88, 95]),
  },
  {
    id: "orders",
    label: "Total Orders",
    value: "1.240",
    delta: 8.1,
    icon: "sales",
    tint: "var(--color-spark-orders)",
    wash: "#E8F1FE",
    spark: spark([30, 34, 32, 38, 36, 42, 40, 45, 43, 50, 48, 55, 53, 60, 66]),
  },
  {
    id: "customers",
    label: "Total Customers",
    value: "320",
    delta: 5.4,
    icon: "customers",
    tint: "var(--color-spark-customers)",
    wash: "#E6F6EF",
    spark: spark([22, 25, 24, 28, 27, 31, 30, 33, 32, 36, 35, 39, 38, 43, 47]),
  },
  {
    id: "profit",
    label: "Total Profit",
    value: "Rp 320.000.000",
    delta: 15.3,
    icon: "accounting",
    tint: "var(--color-spark-profit)",
    wash: "#F1EBFE",
    spark: spark([18, 21, 20, 25, 23, 29, 27, 33, 31, 38, 36, 44, 42, 51, 58]),
  },
];

export interface TopProduct {
  id: string;
  name: string;
  revenue: number;
  delta: number;
  emoji: string;
}

export const topProducts: TopProduct[] = [
  { id: "p1", name: "iPhone 15 Pro Max", revenue: 320 * M, delta: 18.5, emoji: "📱" },
  { id: "p2", name: "Samsung Galaxy S24", revenue: 280 * M, delta: 12.2, emoji: "📱" },
  { id: "p3", name: "MacBook Air M3", revenue: 180 * M, delta: 9.4, emoji: "💻" },
  { id: "p4", name: "iPad Air 6", revenue: 140 * M, delta: 7.6, emoji: "📲" },
  { id: "p5", name: "Apple Watch Series 9", revenue: 120 * M, delta: 6.1, emoji: "⌚" },
];

export interface ActivityItem {
  id: string;
  title: string;
  detail: string;
  at: string;
  icon: string;
  tint: string;
  wash: string;
}

/** Timestamps are relative to a fixed "now" so the demo renders the mockup's ages. */
export const DEMO_NOW = new Date("2024-05-31T17:00:00Z").getTime();
const ago = (mins: number) => new Date(DEMO_NOW - mins * 60_000).toISOString();

export const activities: ActivityItem[] = [
  {
    id: "a1",
    title: "Sales Order SO-000123",
    detail: "confirmed by John Doe",
    at: ago(2),
    icon: "wallet",
    tint: "#EA580C",
    wash: "#FFF2E7",
  },
  {
    id: "a2",
    title: "Purchase Order PO-000321",
    detail: "confirmed by Jane Smith",
    at: ago(15),
    icon: "purchase",
    tint: "#2563EB",
    wash: "#E8F1FE",
  },
  {
    id: "a3",
    title: "New Customer",
    detail: "PT. Maju Bersama created",
    at: ago(60),
    icon: "customers",
    tint: "#059669",
    wash: "#E6F6EF",
  },
  {
    id: "a4",
    title: "Payment Received",
    detail: "Rp 25.000.000 from BCA",
    at: ago(120),
    icon: "accounting",
    tint: "#D97706",
    wash: "#FEF3C7",
  },
  {
    id: "a5",
    title: "Stock Adjustment",
    detail: "Product: iPhone 15 Pro Max",
    at: ago(180),
    icon: "inventory",
    tint: "#7C3AED",
    wash: "#F1EBFE",
  },
];

export interface CashFlowSlice {
  name: string;
  value: number;
  color: string;
}

/**
 * Validated categorical palette (mirrors --color-chart-1..3 in index.css).
 * Literal hex rather than var() because recharts reads these into SVG
 * attributes and the tooltip swatch. Order is fixed, never cycled.
 */
export const cashFlow: CashFlowSlice[] = [
  { name: "Income", value: 850 * M, color: "#059669" },
  { name: "Expense", value: 250 * M, color: "#EA580C" },
  { name: "Other", value: 150 * M, color: "#2563EB" },
];
