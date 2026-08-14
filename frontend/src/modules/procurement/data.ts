import {
  useMutation,
  useQuery,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { api } from "@/core/apiClient";
import type { StatusTone } from "@/shared/ui/Badge";
import type { DraftLine } from "@/shared/ui/LineItemsEditor";

/** Mirrors VendorStatus in app/modules/procurement/models.go. */
export type VendorStatus = "Active" | "Inactive";

/** Mirrors PurchaseStatus in app/modules/procurement/models.go. */
export type PurchaseStatus = "Draft" | "Confirmed" | "Received" | "Cancelled";

export interface Vendor {
  id: string;
  code: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  tax_id?: string;
  payment_term_days: number;
  status: VendorStatus;
  notes?: string;
}

export interface PurchaseOrderLine {
  id: string;
  sku: string;
  description: string;
  quantity: number;
  unit_price: number;
  line_total: number;
}

/** Money arrives as an integer in minor units, as the backend stores it. */
export interface PurchaseOrder {
  id: string;
  number: string;
  vendor_id: string;
  vendor_name: string;
  status: PurchaseStatus;
  order_date: string;
  expected_date?: string;
  received_at?: string;
  currency: string;
  subtotal: number;
  discount_total: number;
  total: number;
  notes?: string;
  lines?: PurchaseOrderLine[];
}

export const VENDOR_STATUS_TONE: Record<VendorStatus, StatusTone> = {
  Active: "good",
  Inactive: "neutral",
};

export const PURCHASE_STATUS_TONE: Record<PurchaseStatus, StatusTone> = {
  Draft: "info",
  Confirmed: "good",
  Received: "good",
  Cancelled: "critical",
};

interface ListResponse<T> {
  data: T[];
  total: number;
}

export const procurementKeys = {
  vendors: (params: { limit: number; offset: number; search?: string }) =>
    ["procurement", "vendors", params] as const,
  orders: (params: { limit: number; offset: number; status?: PurchaseStatus }) =>
    ["procurement", "orders", params] as const,
  order: (id: string) => ["procurement", "order", id] as const,
};

/** Fetches one order with its lines; the list endpoint omits them. */
export function usePurchaseOrder(id: string) {
  return useQuery({
    queryKey: procurementKeys.order(id),
    queryFn: () => api.get<{ data: PurchaseOrder }>(`/procurement/orders/${id}`),
    select: (response) => response.data,
    enabled: Boolean(id),
  });
}

export function useVendors(params: {
  limit: number;
  offset: number;
  search?: string;
}) {
  return useQuery({
    queryKey: procurementKeys.vendors(params),
    queryFn: () => {
      const query = new URLSearchParams({
        limit: String(params.limit),
        offset: String(params.offset),
      });
      if (params.search) query.set("search", params.search);
      return api.get<ListResponse<Vendor>>(`/procurement/vendors?${query}`);
    },
    placeholderData: keepPreviousData,
  });
}

export function usePurchaseOrders(params: {
  limit: number;
  offset: number;
  status?: PurchaseStatus;
}) {
  return useQuery({
    queryKey: procurementKeys.orders(params),
    queryFn: () => {
      const query = new URLSearchParams({
        limit: String(params.limit),
        offset: String(params.offset),
      });
      if (params.status) query.set("status", params.status);
      return api.get<ListResponse<PurchaseOrder>>(`/procurement/orders?${query}`);
    },
    placeholderData: keepPreviousData,
  });
}

/** Invalidating the whole tree also refreshes the dashboard spend figures. */
function invalidateProcurement(
  queryClient: ReturnType<typeof useQueryClient>,
) {
  queryClient.invalidateQueries({ queryKey: ["procurement"] });
  queryClient.invalidateQueries({ queryKey: ["dashboard"] });
  // Receiving an order raises stock on the server, so a stock list already on
  // screen is stale the moment a receipt succeeds.
  queryClient.invalidateQueries({ queryKey: ["inventory"] });
}

export interface VendorInput {
  code?: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  tax_id?: string;
  payment_term_days?: number;
  notes?: string;
}

export function useCreateVendor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: VendorInput) =>
      api.post<{ data: Vendor }>("/procurement/vendors", input),
    onSuccess: () => invalidateProcurement(queryClient),
  });
}

export function useUpdateVendor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: VendorInput }) =>
      api.put<{ data: Vendor }>(`/procurement/vendors/${id}`, input),
    onSuccess: () => invalidateProcurement(queryClient),
  });
}

export function useSetVendorStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: VendorStatus }) =>
      api.put<{ data: Vendor }>(`/procurement/vendors/${id}/status`, { status }),
    onSuccess: () => invalidateProcurement(queryClient),
  });
}

export interface CreatePurchaseOrderInput {
  vendor_id: string;
  order_date?: string;
  expected_date?: string;
  notes?: string;
  lines: DraftLine[];
}

export function useCreatePurchaseOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreatePurchaseOrderInput) =>
      api.post<{ data: PurchaseOrder }>("/procurement/orders", {
        vendor_id: input.vendor_id,
        order_date: input.order_date,
        expected_date: input.expected_date,
        notes: input.notes,
        lines: input.lines.map((line) => ({
          sku: line.sku,
          description: line.description,
          quantity: line.quantity,
          unit_price: line.unitPrice,
        })),
      }),
    onSuccess: () => invalidateProcurement(queryClient),
  });
}

/**
 * The order as it should read after the edit. It replaces rather than patches -
 * the lines sent become the order's lines - which is what the PUT endpoint
 * expects. The number is absent because it is the order's identity, not a
 * field, and the server will not change it.
 */
export type UpdatePurchaseOrderInput = CreatePurchaseOrderInput & { id: string };

export function useUpdatePurchaseOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdatePurchaseOrderInput) =>
      api.put<{ data: PurchaseOrder }>(`/procurement/orders/${input.id}`, {
        vendor_id: input.vendor_id,
        order_date: input.order_date,
        expected_date: input.expected_date,
        notes: input.notes,
        lines: input.lines.map((line) => ({
          sku: line.sku,
          description: line.description,
          quantity: line.quantity,
          unit_price: line.unitPrice,
        })),
      }),
    onSuccess: () => invalidateProcurement(queryClient),
  });
}

export function useUpdatePurchaseStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: PurchaseStatus }) =>
      api.put<{ data: PurchaseOrder }>(`/procurement/orders/${id}/status`, {
        status,
      }),
    onSuccess: () => invalidateProcurement(queryClient),
  });
}
