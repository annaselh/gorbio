import { lazy } from "react";
import type { ModuleManifest } from "@/core/types";

const Customers = lazy(() => import("./pages/Customers"));

export const crmModule: ModuleManifest = {
  name: "crm",
  dependencies: ["base"],
  routes: [{ path: "/crm/customers", element: <Customers /> }],
  menu: [
    {
      label: "Customers",
      icon: "customers",
      path: "/crm/customers",
      sequence: 110,
      section: "modules",
      tint: "#F97316",
    },
  ],
};
