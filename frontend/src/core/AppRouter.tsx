import { Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import type { RouteDef } from "./types";

function RouteFallback() {
  return (
    <div className="grid min-h-64 place-items-center">
      <p className="text-sm text-ink-secondary">Loading…</p>
    </div>
  );
}

function NotFound() {
  return (
    <div className="grid min-h-64 place-items-center text-center">
      <div>
        <p className="text-lg font-semibold text-ink">Page not found</p>
        <p className="mt-1 text-sm text-ink-secondary">
          No installed module registered this route.
        </p>
      </div>
    </div>
  );
}

/** Router assembled entirely from registry routes — the shell hardcodes none. */
export function AppRouter({ routes }: { routes: RouteDef[] }) {
  return (
    <Suspense fallback={<RouteFallback />}>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        {routes.map((r) => (
          <Route key={r.path} path={r.path} element={r.element} />
        ))}
        <Route path="*" element={<NotFound />} />
      </Routes>
    </Suspense>
  );
}
