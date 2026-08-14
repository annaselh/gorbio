import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { api } from "@/core/apiClient";
import type { StatusTone } from "@/shared/ui/Badge";

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
};

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
