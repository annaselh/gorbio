import { useState } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { ChevronDown, ChevronLeft } from "lucide-react";
import type { MenuDef, MenuSection } from "./types";
import { resolveIcon } from "@/shared/icons";
import { cn } from "@/shared/cn";

const SECTION_LABEL: Record<Exclude<MenuSection, "main">, string> = {
  modules: "Modules",
  settings: "Settings",
};

export function Sidebar({
  menu,
  collapsed,
  onToggleCollapse,
}: {
  menu: MenuDef[];
  collapsed: boolean;
  onToggleCollapse: () => void;
}) {
  const { pathname } = useLocation();
  const [closed, setClosed] = useState<Record<string, boolean>>({});

  const bySection = (s: MenuSection) => menu.filter((m) => (m.section ?? "main") === s);
  const active =
    [...menu]
      .filter((m) => pathname === m.path || pathname.startsWith(`${m.path}/`))
      .sort((a, b) => b.path.length - a.path.length)[0] ?? null;

  return (
    <aside
      className={cn(
        "flex h-full shrink-0 flex-col border-r border-hairline bg-surface transition-[width] duration-200",
        collapsed ? "w-[76px]" : "w-[244px]",
      )}
    >
      {/* Brand */}
      <div className="flex h-[68px] items-center gap-2.5 px-4">
        <span className="grid size-9 shrink-0 place-items-center rounded-[11px] bg-gradient-to-br from-[#FB923C] to-[#F97316] shadow-[0_2px_6px_rgba(249,115,22,0.35)]">
          <span className="block size-3.5 rounded-full border-[3px] border-white" />
        </span>
        {!collapsed && (
          <span className="text-[19px] font-bold tracking-tight text-ink">
            Orbio <span className="text-brand">ERP</span>
          </span>
        )}
        <button
          type="button"
          onClick={onToggleCollapse}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          aria-expanded={!collapsed}
          className={cn(
            "ml-auto hidden cursor-pointer rounded-lg p-1.5 text-ink-muted transition-colors hover:bg-hairline-soft hover:text-ink lg:block",
            collapsed && "ml-0",
          )}
        >
          <ChevronLeft
            className={cn("size-4 transition-transform", collapsed && "rotate-180")}
          />
        </button>
      </div>

      {/* Pinned current page */}
      {active && (
        <div
          className={cn(
            "flex items-center gap-3 bg-brand-wash px-5 py-3",
            collapsed && "justify-center px-0",
          )}
        >
          {(() => {
            const Icon = resolveIcon(active.icon);
            return <Icon className="size-[18px] shrink-0 text-brand-strong" />;
          })()}
          {!collapsed && (
            <span className="truncate text-sm font-semibold text-brand-strong">
              {active.label}
            </span>
          )}
        </div>
      )}

      <nav className="scrollbar-slim flex-1 overflow-y-auto px-3 py-4">
        <MenuGroup items={bySection("main")} collapsed={collapsed} heading="Main" />

        {(["modules", "settings"] as const).map((section) => {
          const items = bySection(section);
          if (items.length === 0) return null;
          const isClosed = closed[section] ?? false;
          return (
            <div key={section} className="mt-5">
              {!collapsed && (
                <button
                  type="button"
                  onClick={() =>
                    setClosed((c) => ({ ...c, [section]: !isClosed }))
                  }
                  aria-expanded={!isClosed}
                  className="mb-1 flex w-full cursor-pointer items-center justify-between px-2.5 py-1 text-[11px] font-semibold tracking-[0.08em] text-ink-muted uppercase transition-colors hover:text-ink-secondary"
                >
                  {SECTION_LABEL[section]}
                  <ChevronDown
                    className={cn(
                      "size-3.5 transition-transform",
                      isClosed && "-rotate-90",
                    )}
                  />
                </button>
              )}
              {!isClosed && <MenuGroup items={items} collapsed={collapsed} />}
            </div>
          );
        })}
      </nav>

      <CompanySwitcher collapsed={collapsed} />
    </aside>
  );
}

function MenuGroup({
  items,
  collapsed,
  heading,
}: {
  items: MenuDef[];
  collapsed: boolean;
  heading?: string;
}) {
  if (items.length === 0) return null;
  return (
    <>
      {heading && !collapsed && (
        <p className="mb-1 px-2.5 py-1 text-[11px] font-semibold tracking-[0.08em] text-ink-muted uppercase">
          {heading}
        </p>
      )}
      <ul className="space-y-0.5">
        {items.map((item) => (
          <li key={item.path}>
            <MenuLink item={item} collapsed={collapsed} />
          </li>
        ))}
      </ul>
    </>
  );
}

function MenuLink({ item, collapsed }: { item: MenuDef; collapsed: boolean }) {
  const Icon = resolveIcon(item.icon);
  return (
    <NavLink
      to={item.path}
      title={collapsed ? item.label : undefined}
      className={({ isActive }) =>
        cn(
          "group flex items-center gap-3 rounded-[10px] px-2.5 py-2 text-sm transition-colors",
          collapsed && "justify-center px-0",
          isActive
            ? "bg-brand-wash font-semibold text-brand-strong"
            : "font-medium text-ink-secondary hover:bg-hairline-soft hover:text-ink",
        )
      }
    >
      {({ isActive }) => (
        <>
          <Icon
            className="size-[18px] shrink-0"
            style={
              !isActive && item.tint ? { color: item.tint } : undefined
            }
          />
          {!collapsed && (
            <>
              <span className="truncate">{item.label}</span>
              {item.badge != null && (
                <span className="ml-auto rounded-md bg-brand-wash px-1.5 py-0.5 text-[11px] font-semibold text-brand-strong tnum">
                  {item.badge}
                </span>
              )}
            </>
          )}
        </>
      )}
    </NavLink>
  );
}

function CompanySwitcher({ collapsed }: { collapsed: boolean }) {
  return (
    <div className="border-t border-hairline p-3">
      <button
        type="button"
        className={cn(
          "flex w-full cursor-pointer items-center gap-2.5 rounded-xl border border-hairline px-2.5 py-2.5 text-left transition-colors hover:bg-hairline-soft",
          collapsed && "justify-center border-0 px-0",
        )}
      >
        <span className="grid size-8 shrink-0 place-items-center rounded-full bg-gradient-to-br from-[#FB923C] to-[#F97316] text-[11px] font-bold text-white">
          PT
        </span>
        {!collapsed && (
          <>
            <span className="min-w-0 flex-1">
              <span className="block truncate text-[13px] font-semibold text-ink">
                PT. Orbio Solusi
              </span>
              <span className="block text-xs text-ink-muted">Indonesia</span>
            </span>
            <ChevronDown className="size-4 shrink-0 text-ink-muted" />
          </>
        )}
      </button>
    </div>
  );
}
