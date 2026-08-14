import type { ReactNode } from "react";
import { useAuth } from "./auth";
import { LoginPage } from "./LoginPage";

/**
 * Decides between the login screen and the authenticated shell.
 *
 * Gating here rather than per route means a new module cannot accidentally
 * publish an unguarded page: every registry route renders inside this subtree.
 * The server enforces the same rule independently - this only shapes the UI.
 */
export function AuthGate({ children }: { children: ReactNode }) {
  const { session, isReady } = useAuth();

  if (!isReady) {
    return (
      <div className="grid min-h-dvh place-items-center bg-canvas">
        <p className="text-sm text-ink-secondary">Loading…</p>
      </div>
    );
  }

  if (!session) {
    return <LoginPage />;
  }

  return <>{children}</>;
}
