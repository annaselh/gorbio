import { lazy } from "react";
import type { ModuleManifest } from "@/core/types";
import { SalesOrdersTable } from "./components/SalesOrdersTable";

const SalesOrders = lazy(() => import("./pages/SalesOrders"));

/**
 * Sales. Fills the dashboard's wide slot with its own orders table — base
 * renders it without importing this module.
 */
export const salesModule: ModuleManifest = {
  name: "sales",
  dependencies: ["base", "inventory"],
  routes: [{ path: "/sales/orders", element: <SalesOrders /> }],
  menu: [
    {
      label: "Sales",
      icon: "sales",
      path: "/sales/orders",
      sequence: 120,
      section: "modules",
      tint: "#F97316",
    },
  ],
  slots: {
    "dashboard.wide": () => <SalesOrdersTable />,
  },
};
