import { lazy } from "react";
import type { ModuleManifest } from "@/core/types";
import { PlaceholderPage } from "@/shared/ui/PlaceholderPage";

const Dashboard = lazy(() => import("./pages/Dashboard"));

/**
 * The root module. Owns the dashboard and the shell-level "Main" and "Settings"
 * menus, and declares the two dashboard slots other modules fill.
 */
export const baseModule: ModuleManifest = {
  name: "base",
  routes: [
    { path: "/dashboard", element: <Dashboard /> },
    { path: "/activities", element: <PlaceholderPage title="Activities" icon="activity" /> },
    { path: "/calendar", element: <PlaceholderPage title="Calendar" icon="calendar" /> },
    { path: "/discuss", element: <PlaceholderPage title="Discuss" icon="discuss" /> },
    { path: "/tasks", element: <PlaceholderPage title="Tasks" icon="tasks" /> },
    { path: "/settings/users", element: <PlaceholderPage title="Users & Roles" icon="users" /> },
    { path: "/settings/company", element: <PlaceholderPage title="Company" icon="company" /> },
    { path: "/settings/preferences", element: <PlaceholderPage title="Preferences" icon="preferences" /> },
    { path: "/settings/integrations", element: <PlaceholderPage title="Integrations" icon="integrations" /> },
    { path: "/settings/audit-log", element: <PlaceholderPage title="Audit Log" icon="audit" /> },
  ],
  menu: [
    { label: "Dashboard", icon: "dashboard", path: "/dashboard", sequence: 10, section: "main" },
    { label: "Activities", icon: "activity", path: "/activities", sequence: 20, section: "main" },
    { label: "Calendar", icon: "calendar", path: "/calendar", sequence: 30, section: "main" },
    { label: "Discuss", icon: "discuss", path: "/discuss", sequence: 40, section: "main" },
    { label: "Tasks", icon: "tasks", path: "/tasks", sequence: 50, section: "main", badge: 12 },

    { label: "Users & Roles", icon: "users", path: "/settings/users", sequence: 910, section: "settings" },
    { label: "Company", icon: "company", path: "/settings/company", sequence: 920, section: "settings" },
    { label: "Preferences", icon: "preferences", path: "/settings/preferences", sequence: 930, section: "settings" },
    { label: "Integrations", icon: "integrations", path: "/settings/integrations", sequence: 940, section: "settings" },
    { label: "Audit Log", icon: "audit", path: "/settings/audit-log", sequence: 950, section: "settings" },
  ],
};
