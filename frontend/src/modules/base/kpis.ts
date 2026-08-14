import { formatIDR, formatNumber } from "@/shared/format";
import type { DashboardSummary, Metric, SeriesPoint } from "./api";

/**
 * Presentation for the KPI row. The backend returns figures only, so icon,
 * colour and formatting are decided here - a palette change never needs a
 * server deploy.
 */
export interface KpiCard {
  id: string;
  label: string;
  value: string;
  delta: number;
  icon: string;
  tint: string;
  wash: string;
  spark: { date: string; value: number }[];
}

type SparkSelector = (point: SeriesPoint) => number;

interface KpiSpec {
  id: string;
  label: string;
  icon: string;
  tint: string;
  wash: string;
  metric: (summary: DashboardSummary) => Metric;
  format: (value: number) => string;
  /** Omitted when no daily figure backs this card, so it renders flat rather
   *  than borrowing another card's shape. */
  spark?: SparkSelector;
}

const SPECS: KpiSpec[] = [
  {
    id: "revenue",
    label: "Revenue",
    icon: "wallet",
    tint: "var(--color-spark-sales)",
    wash: "#FFF2E7",
    metric: (s) => s.revenue,
    format: formatIDR,
    spark: (p) => p.revenue,
  },
  {
    id: "orders",
    label: "Sales Orders",
    icon: "sales",
    tint: "var(--color-spark-orders)",
    wash: "#E8F1FE",
    metric: (s) => s.orders,
    format: formatNumber,
    spark: (p) => p.orders,
  },
  {
    id: "customers",
    label: "Customers",
    icon: "customers",
    tint: "var(--color-spark-customers)",
    wash: "#E6F6EF",
    metric: (s) => s.customers,
    format: formatNumber,
    // A distinct-customer count per day does not accumulate into a meaningful
    // line, so this card carries no sparkline rather than a misleading one.
  },
  {
    id: "gross_margin",
    label: "Gross Margin",
    icon: "accounting",
    tint: "var(--color-spark-profit)",
    wash: "#F1EBFE",
    metric: (s) => s.gross_margin,
    format: formatIDR,
    spark: (p) => p.revenue - p.purchases,
  },
];

export function buildKpiCards(
  summary: DashboardSummary,
  series: SeriesPoint[],
): KpiCard[] {
  return SPECS.map((spec) => {
    const metric = spec.metric(summary);
    return {
      id: spec.id,
      label: spec.label,
      value: spec.format(metric.value),
      delta: metric.delta,
      icon: spec.icon,
      tint: spec.tint,
      wash: spec.wash,
      spark: spec.spark
        ? series.map((point) => ({
            date: point.date,
            value: spec.spark!(point),
          }))
        : [],
    };
  });
}
