import {
  useMutation,
  useQuery,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { api } from "@/core/apiClient";
import type { StatusTone } from "@/shared/ui/Badge";

/** Mirrors CustomerStatus in app/modules/crm/models.go. */
export type CustomerStatus = "Active" | "Inactive";

export interface Customer {
  id: string;
  code: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  tax_id?: string;
  credit_term_days: number;
  status: CustomerStatus;
  notes?: string;
}

export const CUSTOMER_STATUS_TONE: Record<CustomerStatus, StatusTone> = {
  Active: "good",
  Inactive: "neutral",
};

interface ListResponse {
  data: Customer[];
  total: number;
}

export const crmKeys = {
  customers: (params: { limit: number; offset: number; search?: string }) =>
    ["crm", "customers", params] as const,
};

export function useCustomers(params: {
  limit: number;
  offset: number;
  search?: string;
}) {
  return useQuery({
    queryKey: crmKeys.customers(params),
    queryFn: () => {
      const query = new URLSearchParams({
        limit: String(params.limit),
        offset: String(params.offset),
      });
      if (params.search) query.set("search", params.search);
      return api.get<ListResponse>(`/crm/customers?${query}`);
    },
    placeholderData: keepPreviousData,
  });
}

/** Sales orders reference customers, so their lists go stale on a change too. */
function invalidateCustomers(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: ["crm"] });
  queryClient.invalidateQueries({ queryKey: ["sales"] });
}

export interface CustomerInput {
  code?: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  tax_id?: string;
  credit_term_days?: number;
  notes?: string;
}

export function useCreateCustomer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CustomerInput) =>
      api.post<{ data: Customer }>("/crm/customers", input),
    onSuccess: () => invalidateCustomers(queryClient),
  });
}

export function useUpdateCustomer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: CustomerInput }) =>
      api.put<{ data: Customer }>(`/crm/customers/${id}`, input),
    onSuccess: () => invalidateCustomers(queryClient),
  });
}

export function useSetCustomerStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: CustomerStatus }) =>
      api.put<{ data: Customer }>(`/crm/customers/${id}/status`, { status }),
    onSuccess: () => invalidateCustomers(queryClient),
  });
}
