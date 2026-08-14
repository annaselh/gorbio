import { lazy } from "react";
import type { ModuleManifest } from "@/core/types";

const PurchaseOrders = lazy(() => import("./pages/PurchaseOrders"));
const Vendors = lazy(() => import("./pages/Vendors"));

export const procurementModule: ModuleManifest = {
  name: "procurement",
  dependencies: ["base"],
  routes: [
    { path: "/procurement/orders", element: <PurchaseOrders /> },
    { path: "/procurement/vendors", element: <Vendors /> },
  ],
  menu: [
    {
      label: "Purchase",
      icon: "purchase",
      path: "/procurement/orders",
      sequence: 130,
      section: "modules",
      tint: "#2563EB",
    },
    {
      label: "Vendors",
      icon: "company",
      path: "/procurement/vendors",
      sequence: 135,
      section: "modules",
      tint: "#2563EB",
    },
  ],
};
