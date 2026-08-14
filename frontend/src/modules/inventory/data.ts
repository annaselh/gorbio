import { useQuery } from "@tanstack/react-query";
import { api } from "@/core/apiClient";

/** Mirrors StockItem in app/modules/inventory/models.go. */
export interface StockItem {
  id: string;
  tenant_id: string;
  sku: string;
  name: string;
  unit: string;
  quantity: number;
  reorder_level: number;
}

interface ListResponse {
  data: StockItem[];
}

export const inventoryKeys = {
  items: (lowStockOnly: boolean) =>
    ["inventory", "items", { lowStockOnly }] as const,
};

/**
 * Reads real stock from the backend. `low_stock=true` asks the server to apply
 * the quantity <= reorder_level rule so the client never has to fetch the whole
 * catalogue just to render the alert widget.
 */
export function useStockItems(options: { lowStockOnly?: boolean } = {}) {
  const lowStockOnly = options.lowStockOnly ?? false;

  return useQuery({
    queryKey: inventoryKeys.items(lowStockOnly),
    queryFn: () =>
      api.get<ListResponse>(
        `/inventory/items${lowStockOnly ? "?low_stock=true" : ""}`,
      ),
    select: (response) => response.data ?? [],
  });
}

export function isOutOfStock(item: StockItem) {
  return item.quantity === 0;
}
