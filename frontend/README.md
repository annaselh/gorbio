# Orbio ERP — Frontend

React shell for the modular ERP platform. Implements **FE-1** (shell + registry),
**FE-2** (core modules, Strategy 1), **FE-3** (server-driven `DynamicForm`) and
**FE-4** (slots) from `../docs/frontend-architect.md`. FE-5 (Module Federation)
is deliberately not started.

## Run

```bash
npm install
npm run dev      # http://localhost:5173
npm run build    # tsc -b && vite build
```

`/api` proxies to `http://localhost:8080` (the Go backend in `../app`). Override
the API base with `VITE_API_BASE_URL`.

## Stack

React 19 · Vite 8 · TypeScript · Tailwind v4 · TanStack Query · React Router 7 ·
Zustand · Recharts 3 · lucide-react.

The architecture doc lists Ant Design / Mantine / shadcn as options. This uses
Tailwind with hand-built components: the mockup is a bespoke design system
(custom sidebar, card, and badge treatments), and theming Ant Design to match it
would cost more than building the dozen primitives in `shared/ui`.

## Layout

```
src/
├── core/              SHELL — knows no business logic
│   ├── types.ts       ModuleManifest, RouteDef, MenuDef
│   ├── registry.ts    resolveDependencyOrder (topo-sort) + buildRegistry
│   ├── modules.ts     the only file that knows which modules are installed
│   ├── Shell.tsx      layout host + PageHeader
│   ├── Sidebar.tsx    menu built from the registry, grouped by section
│   ├── Topbar.tsx     search (⌘K), notifications, account
│   ├── Slot.tsx       cross-module UI injection (the `_inherit` mirror)
│   ├── DynamicForm.tsx server-driven form for runtime custom fields
│   ├── apiClient.ts   fetch wrapper: cookie credentials, ApiError
│   ├── auth.tsx       session context: login, logout, permission checks
│   ├── AuthGate.tsx   unauthenticated screens vs. the shell; token links
│   ├── AuthLayout.tsx shared card + field styling for the auth screens
│   ├── LoginPage.tsx  email / password / optional company form
│   ├── ForgotPasswordPage.tsx  requests a reset link
│   ├── ResetPasswordPage.tsx   consumes ?token= from the reset email
│   ├── VerifyEmailPage.tsx     consumes ?token= from the verification email
│   └── AppRouter.tsx  router assembled from registry.routes
│
├── modules/           BUSINESS — may use core, never each other
│   ├── base/          dashboard, main + settings menus, dashboard slots
│   ├── sales/         orders page; fills `dashboard.wide`
│   ├── inventory/     stock page; fills `dashboard.aside`
│   └── planned.tsx    roadmap modules — menu entry + "not implemented" page
│
└── shared/            cross-module utilities
    ├── ui/            Card, Badge, Button, Select, Pagination, Placeholder
    ├── charts/        Sparkline, tooltip shell, screen-reader table
    ├── format.ts      id-ID currency/number/date formatting
    └── icons.ts       name → lucide component (keeps manifests serialisable)
```

Dependency rule: `App.tsx → modules → core`. Modules never import each other —
the dashboard's Sales Orders and Stock Alert cards arrive through slots, so
`base` has no reference to `sales` or `inventory`.

## Adding a module

```tsx
// src/modules/purchase/index.tsx
export const purchaseModule: ModuleManifest = {
  name: "purchase",
  dependencies: ["base"],
  routes: [{ path: "/purchase/orders", element: <PurchaseOrders /> }],
  menu: [{ label: "Purchase", icon: "purchase", path: "/purchase/orders",
           sequence: 130, section: "modules" }],
  slots: { "dashboard.aside": () => <PurchaseWidget /> },
};
```

Register it in `core/modules.ts` and delete its row from `modules/planned.tsx`.
`buildRegistry` throws at boot on a cycle, a missing dependency, or a duplicate
name.

## Data

Every figure is fixture data in each module's `data.ts`, shaped like the REST
responses it will replace — the Go backend exposes no endpoints yet. Swap each
import for a TanStack Query call against `core/apiClient` as endpoints land.

## Charts

Colours were validated with the dataviz palette checker, not chosen by eye. The
Cash Flow categorical palette is `#059669 / #EA580C / #2563EB`; the mockup's
brighter `#22C55E / #F97316` pair fails colourblind separation (deutan ΔE 6.2)
and sub-3:1 surface contrast. Every chart ships a hover layer, a legend when it
has two or more series, and a screen-reader `<SrTable>` data view.

## Known gaps

- **Light theme only.** The mockup specifies no dark palette; adding one means
  re-stepping the chart ramps against a dark surface and re-running the
  validator, not flipping the tokens.
- **Auth is wired.** `core/auth.tsx` holds the session, `AuthGate` renders the
  login screen until `GET /api/auth/me` succeeds, and the topbar shows the real
  user and company with a sign-out action. The session is an HttpOnly cookie set
  by the backend, so `apiClient` sends `credentials: "include"` and never
  handles a token itself.
- **Sales and inventory read live data**; the `base` dashboard widgets (KPIs,
  cash flow, sales overview, top products, recent activities) are still
  fixture-backed until their backend endpoints exist. `SalesOrdersTable`
  paginates server-side and keeps the previous page on screen while the next
  one loads.
- Five `oxlint` fast-refresh warnings: three from manifest files exporting a
  const beside lazy component refs, two from `auth.tsx` and `AuthLayout.tsx`
  exporting a hook and a shared class string beside their components. Both are
  inherent to the pattern and affect HMR granularity only.
