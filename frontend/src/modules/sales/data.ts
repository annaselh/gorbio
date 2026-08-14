import {
  useMutation,
  useQuery,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { api } from "@/core/apiClient";
import type { StatusTone } from "@/shared/ui/Badge";
import type { DraftLine } from "@/shared/ui/LineItemsEditor";

/** Mirrors OrderStatus in app/modules/sales/models.go. */
export type OrderStatus = "Confirmed" | "Draft" | "Cancelled";

/** Mirrors OrderLine in app/modules/sales/models.go. */
export interface SalesOrderLine {
  id: string;
  sku: string;
  description: string;
  quantity: number;
  unit_price: number;
  line_total: number;
}

/**
 * Mirrors Order in app/modules/sales/models.go.
 *
 * Money arrives as an integer in minor units, matching how the backend stores
 * it, so no rounding happens in transit.
 */
export interface SalesOrder {
  id: string;
  number: string;
  customer_name: string;
  status: OrderStatus;
  order_date: string;
  currency: string;
  subtotal: number;
  discount_total: number;
  total: number;
  notes?: string;
  lines?: SalesOrderLine[];
}

interface ListResponse {
  data: SalesOrder[];
  total: number;
}

export const ORDER_STATUS_TONE: Record<OrderStatus, StatusTone> = {
  Confirmed: "good",
  Draft: "info",
  Cancelled: "critical",
};

export const salesKeys = {
  orders: (params: { limit: number; offset: number; status?: OrderStatus }) =>
    ["sales", "orders", params] as const,
  order: (id: string) => ["sales", "order", id] as const,
};

/**
 * Server-side pagination: the backend returns one page plus the unpaginated
 * count, so the table never downloads the whole ledger to render five rows.
 * keepPreviousData holds the current page on screen while the next one loads,
 * which stops the table collapsing to its empty state on every page change.
 */
export function useSalesOrders(params: {
  limit: number;
  offset: number;
  status?: OrderStatus;
}) {
  return useQuery({
    queryKey: salesKeys.orders(params),
    queryFn: () => {
      const query = new URLSearchParams({
        limit: String(params.limit),
        offset: String(params.offset),
      });
      if (params.status) query.set("status", params.status);
      return api.get<ListResponse>(`/sales/orders?${query}`);
    },
    placeholderData: keepPreviousData,
  });
}

export function useSalesOrder(id: string) {
  return useQuery({
    queryKey: salesKeys.order(id),
    queryFn: () => api.get<{ data: SalesOrder }>(`/sales/orders/${id}`),
    select: (response) => response.data,
    enabled: Boolean(id),
  });
}

/** Invalidating the whole "sales" tree also refreshes the dashboard figures
 *  that derive from these orders. */
function invalidateSales(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: ["sales"] });
  queryClient.invalidateQueries({ queryKey: ["dashboard"] });
}

export interface CreateSalesOrderInput {
  customer_name: string;
  order_date?: string;
  notes?: string;
  lines: DraftLine[];
}

export function useCreateSalesOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateSalesOrderInput) =>
      api.post<{ data: SalesOrder }>("/sales/orders", {
        customer_name: input.customer_name,
        order_date: input.order_date,
        notes: input.notes,
        lines: input.lines.map((line) => ({
          sku: line.sku,
          description: line.description,
          quantity: line.quantity,
          unit_price: line.unitPrice,
        })),
      }),
    onSuccess: () => invalidateSales(queryClient),
  });
}

export function useUpdateSalesOrderStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: OrderStatus }) =>
      api.put<{ data: SalesOrder }>(`/sales/orders/${id}/status`, { status }),
    onSuccess: () => invalidateSales(queryClient),
  });
}
