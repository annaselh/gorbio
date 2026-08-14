import { useEffect, useRef, useState } from "react";
import {
  Bell,
  ChevronDown,
  CircleHelp,
  KeyRound,
  LogOut,
  Menu,
  MessageSquare,
  Plus,
  Search,
} from "lucide-react";
import { cn } from "@/shared/cn";
import { useAuth } from "./auth";
import { ChangePasswordDialog } from "./ChangePasswordDialog";

const iconButton =
  "grid size-9 cursor-pointer place-items-center rounded-lg text-ink-secondary transition-colors hover:bg-hairline-soft hover:text-ink";

export function Topbar({ onToggleNav }: { onToggleNav: () => void }) {
  const searchRef = useRef<HTMLInputElement>(null);

  // ⌘K / Ctrl+K focuses search — the shortcut the placeholder advertises.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key.toLowerCase() === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <header className="flex h-[68px] shrink-0 items-center gap-3 border-b border-hairline bg-surface px-4 sm:px-5">
      <button
        type="button"
        onClick={onToggleNav}
        className={cn(iconButton, "shrink-0")}
        aria-label="Toggle navigation"
      >
        <Menu className="size-5" />
      </button>

      {/* min-w-0 lets the search shrink instead of forcing the page to scroll
          horizontally on narrow viewports. */}
      <div className="relative min-w-0 flex-1 sm:max-w-[380px]">
        <Search
          aria-hidden
          className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-ink-muted"
        />
        <input
          ref={searchRef}
          type="search"
          placeholder="Search anything..."
          aria-label="Search"
          className="w-full rounded-xl border border-hairline bg-surface py-2.5 pr-16 pl-10 text-sm text-ink placeholder:text-ink-muted focus:outline-2 focus:outline-offset-0 focus:outline-brand"
        />
        <kbd className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 rounded-md border border-hairline bg-hairline-soft px-1.5 py-0.5 text-[11px] font-medium text-ink-muted">
          ⌘ K
        </kbd>
      </div>

      <div className="ml-auto flex shrink-0 items-center gap-1">
        <button
          type="button"
          aria-label="Create new"
          className={cn(iconButton, "hidden border border-hairline sm:grid")}
        >
          <Plus className="size-[18px]" />
        </button>

        <button
          type="button"
          aria-label="Notifications, 3 unread"
          className={cn(iconButton, "relative")}
        >
          <Bell className="size-[18px]" />
          <span className="absolute top-1 right-1 grid min-w-4 place-items-center rounded-full bg-status-critical px-1 text-[10px] font-semibold text-white tnum">
            3
          </span>
        </button>

        {/* Secondary actions fold away before the layout can overflow. */}
        <button type="button" aria-label="Messages" className={cn(iconButton, "hidden md:grid")}>
          <MessageSquare className="size-[18px]" />
        </button>
        <button type="button" aria-label="Help" className={cn(iconButton, "hidden md:grid")}>
          <CircleHelp className="size-[18px]" />
        </button>

        <span aria-hidden className="mx-2 hidden h-7 w-px bg-hairline sm:block" />

        <UserMenu />
      </div>
    </header>
  );
}

/** Derives up to two initials from a display name, falling back to the email. */
function initialsOf(name: string, email: string) {
  const source = name.trim() || email.trim();
  if (!source) return "?";
  const parts = source.split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
  return source.slice(0, 2).toUpperCase();
}

function UserMenu() {
  const { session, logout } = useAuth();
  const [open, setOpen] = useState(false);
  const [changingPassword, setChangingPassword] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  if (!session) return null;

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
        className="flex cursor-pointer items-center gap-2.5 rounded-lg px-1 py-1 transition-colors hover:bg-hairline-soft"
      >
        <span className="relative">
          <span className="grid size-9 place-items-center rounded-full bg-gradient-to-br from-[#94A3B8] to-[#64748B] text-xs font-semibold text-white">
            {initialsOf(session.display_name, session.email)}
          </span>
          <span
            aria-hidden
            className="absolute right-0 bottom-0 size-2.5 rounded-full border-2 border-surface bg-status-good"
          />
        </span>
        <span className="hidden text-left sm:block">
          <span className="block text-[13px] font-semibold text-ink">
            {session.display_name || session.email}
          </span>
          <span className="block text-xs text-ink-muted">
            {session.tenant_name}
          </span>
        </span>
        <ChevronDown className="hidden size-4 text-ink-muted sm:block" />
      </button>

      {open && (
        <div
          role="menu"
          className="absolute right-0 z-50 mt-2 w-56 rounded-xl border border-hairline bg-surface p-1.5 shadow-[0_8px_24px_rgba(16,24,40,0.12)]"
        >
          <div className="border-b border-hairline px-2.5 py-2">
            <p className="truncate text-[13px] font-semibold text-ink">
              {session.display_name || session.email}
            </p>
            <p className="truncate text-xs text-ink-muted">{session.email}</p>
          </div>
          <button
            type="button"
            role="menuitem"
            onClick={() => {
              setOpen(false);
              setChangingPassword(true);
            }}
            className="mt-1 flex w-full cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-ink transition-colors hover:bg-hairline-soft"
          >
            <KeyRound className="size-4 text-ink-secondary" />
            Change password
          </button>
          <button
            type="button"
            role="menuitem"
            onClick={() => {
              setOpen(false);
              void logout();
            }}
            className="flex w-full cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-ink transition-colors hover:bg-hairline-soft"
          >
            <LogOut className="size-4 text-ink-secondary" />
            Sign out
          </button>
        </div>
      )}

      {changingPassword && (
        <ChangePasswordDialog onClose={() => setChangingPassword(false)} />
      )}
    </div>
  );
}
