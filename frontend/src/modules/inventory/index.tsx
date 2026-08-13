import { lazy } from "react";
import type { ModuleManifest } from "@/core/types";
import { StockAlert } from "./components/StockAlert";

const Stock = lazy(() => import("./pages/Stock"));

export const inventoryModule: ModuleManifest = {
  name: "inventory",
  dependencies: ["base"],
  routes: [{ path: "/inventory/stock", element: <Stock /> }],
  menu: [
    {
      label: "Inventory",
      icon: "inventory",
      path: "/inventory/stock",
      sequence: 140,
      section: "modules",
      tint: "#F59E0B",
    },
  ],
  slots: {
    "dashboard.aside": () => <StockAlert />,
  },
};
