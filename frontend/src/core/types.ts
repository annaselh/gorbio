import type { ComponentType, ReactNode } from "react";

export interface RouteDef {
  path: string;
  element: ReactNode;
}

/**
 * Menu groups rendered as labelled sections in the sidebar.
 * `main` is pinned above the group headings.
 */
export type MenuSection = "main" | "modules" | "settings";

export interface MenuDef {
  label: string;
  /** Icon name resolved through shared/icons — kept a string so a manifest
   *  stays serialisable for the federated / server-driven cases (FE-5). */
  icon?: string;
  path: string;
  sequence: number;
  section?: MenuSection;
  parent?: string;
  badge?: number | string;
  /** Module identity colour for the sidebar icon. Optional presentation hint —
   *  a module that omits it renders in neutral ink. */
  tint?: string;
}

export interface ModuleManifest {
  name: string;
  dependencies?: string[];
  routes?: RouteDef[];
  menu?: MenuDef[];
  slots?: Record<string, ComponentType<SlotProps>>;
}

export interface SlotProps {
  ctx?: unknown;
}

export interface Registry {
  ordered: ModuleManifest[];
  routes: RouteDef[];
  menuItems: MenuDef[];
  slots: Record<string, ComponentType<SlotProps>[]>;
}
