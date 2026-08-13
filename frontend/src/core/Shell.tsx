import { useState, type ReactNode } from "react";
import type { MenuDef } from "./types";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import { cn } from "@/shared/cn";

/**
 * Layout host. Knows nothing about business modules — it only renders the menu
 * the registry handed it and whatever the router put in `children`.
 */
export function Shell({
  menu,
  children,
}: {
  menu: MenuDef[];
  children: ReactNode;
}) {
  const [collapsed, setCollapsed] = useState(false);
  const [navOpen, setNavOpen] = useState(false);

  return (
    <div className="flex h-dvh overflow-hidden bg-canvas">
      {/* Off-canvas drawer below lg, static column at lg and up */}
      <div
        className={cn(
          "fixed inset-y-0 left-0 z-40 transition-transform duration-200 lg:static lg:translate-x-0",
          navOpen ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <Sidebar
          menu={menu}
          collapsed={collapsed}
          onToggleCollapse={() => setCollapsed((c) => !c)}
        />
      </div>

      {navOpen && (
        <button
          type="button"
          aria-label="Close navigation"
          onClick={() => setNavOpen(false)}
          className="fixed inset-0 z-30 bg-ink/20 lg:hidden"
        />
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar onToggleNav={() => setNavOpen((o) => !o)} />
        <main className="scrollbar-slim flex-1 overflow-y-auto">
          <div className="mx-auto max-w-[1600px] px-4 py-6 sm:px-7">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}

export function PageHeader({
  title,
  subtitle,
  actions,
}: {
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}) {
  return (
    <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 className="text-[26px] font-bold tracking-tight text-ink">
          {title}
        </h1>
        {subtitle && (
          <p className="mt-1 text-sm text-ink-secondary">{subtitle}</p>
        )}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2.5">{actions}</div>}
    </div>
  );
}
