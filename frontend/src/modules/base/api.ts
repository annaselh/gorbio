import { useQuery } from "@tanstack/react-query";
import { api } from "@/core/apiClient";

/**
 * Dashboard endpoints. The backend returns figures only - no icons, colours or
 * emoji. Presentation stays in the components, so a change of palette never
 * requires a server deploy.
 *
 * Money is an integer in minor units, matching how the backend stores it.
 */

export interface Metric {
  value: number;
  previous: number;
  delta: number;
}

export interface DashboardSummary {
  revenue: Metric;
  orders: Metric;
  customers: Metric;
  purchases: Metric;
  gross_margin: Metric;
  period: { start: string; end: string };
}

/** One day across every headline figure; one request backs all sparklines. */
export interface SeriesPoint {
  date: string;
  revenue: number;
  orders: number;
  purchases: number;
}

export interface TopProduct {
  sku: string;
  description: string;
  revenue: number;
  quantity: number;
  delta: number;
}

export interface CashFlowSlice {
  name: string;
  value: number;
}

export interface ActivityEntry {
  id: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  actor_name?: string;
  created_at: string;
}

export const dashboardKeys = {
  summary: ["dashboard", "summary"] as const,
  salesSeries: ["dashboard", "sales-series"] as const,
  topProducts: (limit: number) => ["dashboard", "top-products", limit] as const,
  cashFlow: ["dashboard", "cash-flow"] as const,
  activities: (limit: number) => ["dashboard", "activities", limit] as const,
};

export function useDashboardSummary() {
  return useQuery({
    queryKey: dashboardKeys.summary,
    queryFn: () => api.get<{ data: DashboardSummary }>("/dashboard/summary"),
    select: (response) => response.data,
  });
}

export function useSalesSeries() {
  return useQuery({
    queryKey: dashboardKeys.salesSeries,
    queryFn: () =>
      api.get<{ data: { current: SeriesPoint[]; previous: SeriesPoint[] } }>(
        "/dashboard/sales-series",
      ),
    select: (response) => response.data,
  });
}

export function useTopProducts(limit = 5) {
  return useQuery({
    queryKey: dashboardKeys.topProducts(limit),
    queryFn: () =>
      api.get<{ data: TopProduct[] }>(`/dashboard/top-products?limit=${limit}`),
    select: (response) => response.data ?? [],
  });
}

export function useCashFlow() {
  return useQuery({
    queryKey: dashboardKeys.cashFlow,
    queryFn: () => api.get<{ data: CashFlowSlice[] }>("/dashboard/cash-flow"),
    select: (response) => response.data ?? [],
  });
}

export function useActivities(limit = 5) {
  return useQuery({
    queryKey: dashboardKeys.activities(limit),
    queryFn: () =>
      api.get<{ data: ActivityEntry[] }>(`/dashboard/activities?limit=${limit}`),
    select: (response) => response.data ?? [],
  });
}

/**
 * Audit actions are dotted verb phrases such as "sales.order_confirmed".
 * Turning one into a sentence here keeps the backend free of display copy.
 */
export function describeAction(action: string): string {
  const [, verb = action] = action.split(".");
  return verb.replace(/_/g, " ").replace(/^./, (c) => c.toUpperCase());
}

/** Maps an audit action to the module whose icon should represent it. */
export function moduleOfAction(action: string): string {
  return action.split(".")[0] ?? "base";
}
