import type { ModuleManifest } from "@/core/types";
import { PlaceholderPage } from "@/shared/ui/PlaceholderPage";

/**
 * Roadmap modules. Each registers a menu entry and one route that renders an
 * explicit "not implemented" page, so the sidebar matches the design without
 * pretending these features exist. Replace an entry here with a real
 * `modules/<name>/index.tsx` as it gets built.
 */
const PLANNED = [
  { name: "crm", label: "Customers (CRM)", icon: "customers", path: "/crm/customers", sequence: 110, tint: "#F97316" },
  { name: "purchase", label: "Purchase", icon: "purchase", path: "/purchase/orders", sequence: 130, tint: "#0D9488" },
  { name: "accounting", label: "Accounting", icon: "accounting", path: "/accounting", sequence: 150, tint: "#F59E0B" },
  { name: "hr", label: "HR", icon: "hr", path: "/hr", sequence: 160, tint: "#8B5CF6" },
  { name: "projects", label: "Projects", icon: "projects", path: "/projects", sequence: 170, tint: "#64748B" },
  { name: "manufacturing", label: "Manufacturing", icon: "manufacturing", path: "/manufacturing", sequence: 180, tint: "#14B8A6" },
  { name: "helpdesk", label: "Helpdesk", icon: "helpdesk", path: "/helpdesk", sequence: 190, tint: "#64748B" },
] as const;

export const plannedModules: ModuleManifest[] = PLANNED.map((m) => ({
  name: m.name,
  dependencies: ["base"],
  routes: [
    { path: m.path, element: <PlaceholderPage title={m.label} icon={m.icon} /> },
  ],
  menu: [
    {
      label: m.label,
      icon: m.icon,
      path: m.path,
      sequence: m.sequence,
      section: "modules" as const,
      tint: m.tint,
    },
  ],
}));
