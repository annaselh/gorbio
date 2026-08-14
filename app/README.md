# Gorbio App

Gorbio is a Go-based modular ERP backend foundation designed to support extensible business modules, dependency-aware bootstrapping, and a clean separation between the application core and domain modules.

This project is currently structured as a framework skeleton for a modular ERP platform, with the core runtime and module registration system already in place.

## Overview

The application follows a modular architecture where:

- the core provides shared runtime behavior,
- modules register themselves with the application,
- dependencies between modules are resolved before boot,
- module lifecycle hooks and migrations are executed in order,
- the app exposes an HTTP server and internal event bus.

## Architecture

### Core runtime

The main runtime lives in the `core` package:

- `core/app.go` initializes the application and boots registered modules.
- `core/registry.go` manages module registration and topological dependency resolution.
- `core/eventbus.go` provides an internal event bus for module communication.
- `core/router.go` exposes the app HTTP routing layer.

### Module system

Modules are defined and registered through the application core. The dependency order is validated and resolved automatically before installation and route registration.

The initial module layout is under `modules/`, with a base module scaffold in `modules/base`.

### Server entrypoint

The application bootstrap is defined in:

- `cmd/server/main.go`

This is the place where the server is started and the app is assembled.

## Project structure

```text
app/
├── cmd/
│   └── server/
│       └── main.go
├── core/
│   ├── acl/
│   ├── app.go
│   ├── boot.go
│   ├── context.go
│   ├── db.go
│   ├── eventbus.go
│   ├── module.go
│   ├── registry.go
│   ├── registry_test.go
│   └── router.go
├── internal/
├── modules/
│   └── base/
│       ├── base.go
│       ├── handlers.go
│       ├── migrations/
│       ├── models.go
│       ├── module.go
│       └── service.go
├── plugins/
├── proto/
├── test/
├── web/
├── go.mod
├── go.sum
└── README.md
```

## Current capabilities

The current implementation includes the following foundational pieces:

- module registry and dependency validation,
- lifecycle boot flow,
- migration execution hooks,
- route registration hooks,
- event subscriber registration,
- HTTP server startup via the app runtime,
- cookie session authentication with Argon2id password hashing,
- role and permission checks (RBAC) applied per route,
- password reset and email verification over SMTP,
- credentialed CORS with an explicit origin allowlist.

## HTTP endpoints

Registered by the `base` module. Routes marked *auth* require a valid session
cookie; the rest are reachable without one by design, since their caller is
someone who cannot sign in.

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| POST | `/api/auth/login` | – | Sign in; sets the session cookie |
| POST | `/api/auth/logout` | auth | Revoke the current session |
| GET | `/api/auth/me` | auth | Profile, tenant and permission codes |
| POST | `/api/auth/password/forgot` | – | Mail a reset link; always answers 202 |
| POST | `/api/auth/password/reset` | – | Consume a reset token, set a new password |
| POST | `/api/auth/password/change` | auth | Change password by proving the current one |
| POST | `/api/auth/email/verify` | – | Consume a verification token |
| POST | `/api/auth/email/resend` | auth | Re-send the verification email |

Business modules register their own routes behind `RequireAuth` and an explicit
permission; see `modules/sales/routes.go` for the pattern.

| Method | Path | Permission | Purpose |
| --- | --- | --- | --- |
| GET | `/api/sales/orders` | `sales.read` | Paged list with `limit`, `offset`, `status`, `customer` |
| GET | `/api/sales/orders/:id` | `sales.read` | One order with its lines |
| POST | `/api/sales/orders` | `sales.manage` | Create a draft order |
| PUT | `/api/sales/orders/:id/status` | `sales.manage` | Confirm or cancel |
| POST | `/api/sales/orders/:id/discount` | `sales.manage` | Applied by the sales-discount extension |
| GET | `/api/inventory/items` | `inventory.read` | Stock list; `low_stock=true` filters to alerts |
| GET | `/api/inventory/items/:id` | `inventory.read` | One stock item |
| POST | `/api/inventory/items` | `inventory.manage` | Create a stock item |
| POST | `/api/inventory/items/:id/adjust` | `inventory.manage` | Signed quantity delta |
| GET | `/api/procurement/vendors` | `procurement.read` | Vendor list; `search`, `status` |
| GET | `/api/procurement/vendors/:id` | `procurement.read` | One vendor |
| POST | `/api/procurement/vendors` | `procurement.manage` | Create a vendor |
| PUT | `/api/procurement/vendors/:id` | `procurement.manage` | Update vendor details |
| PUT | `/api/procurement/vendors/:id/status` | `procurement.manage` | Activate or deactivate |
| GET | `/api/procurement/orders` | `procurement.read` | Purchase orders; `status`, `vendor_id` |
| GET | `/api/procurement/orders/:id` | `procurement.read` | One purchase order with lines |
| POST | `/api/procurement/orders` | `procurement.manage` | Raise a draft purchase order |
| PUT | `/api/procurement/orders/:id/status` | `procurement.manage` | Confirm, receive or cancel |
| GET | `/api/crm/customers` | `crm.read` | Customer list; `search`, `status` |
| GET | `/api/crm/customers/:id` | `crm.read` | One customer |
| POST | `/api/crm/customers` | `crm.manage` | Create a customer |
| PUT | `/api/crm/customers/:id` | `crm.manage` | Update customer details |
| PUT | `/api/crm/customers/:id/status` | `crm.manage` | Activate or deactivate |
| GET | `/api/members` | `membership.read` | People with access to the tenant |
| GET | `/api/roles` | `membership.read` | Assignable roles |
| POST | `/api/members` | `membership.manage` | Invite; mails a set-password link |
| PUT | `/api/members/:id/roles` | `membership.manage` | Replace a member's roles |
| PUT | `/api/members/:id/status` | `membership.manage` | Suspend or reactivate |
| GET | `/api/dashboard/summary` | `tenant.read` | KPI figures with month-over-month deltas |
| GET | `/api/dashboard/sales-series` | `tenant.read` | Daily revenue, orders and purchases |
| GET | `/api/dashboard/top-products` | `tenant.read` | SKUs ranked by confirmed revenue |
| GET | `/api/dashboard/cash-flow` | `tenant.read` | Confirmed income against purchase spend |
| GET | `/api/dashboard/activities` | `tenant.read` | Recent business events from the audit trail |

An invited user is created with **no password hash**: sign-in is impossible
until they complete the emailed reset, so a credential is never transmitted.
Suspending a membership revokes that member's sessions immediately, and the
service refuses to remove or demote a tenant's last active owner.

A sales order may carry a `customer_id` linking to a CRM record, or just a
`customer_name` for a walk-in. When the link is present the server resolves the
name from the CRM record and ignores any name the client sent, so the two cannot
diverge. The dependency points one way on purpose: **CRM depends on sales and
hands it a customer lookup at registration; sales knows nothing about CRM** and
still works with the module uninstalled.

The dashboard module depends on base, sales, inventory, procurement and CRM
rather than the reverse - putting these queries in base would invert the dependency
every other module relies on. Only *Confirmed* sales count as revenue and only
*Confirmed* or *Received* purchases count as spend: a draft is a proposal and a
cancelled order never happened. Gross margin is revenue minus purchase spend,
which is the closest honest figure without cost-of-goods tracking. Cash flow
reports two slices, not three; the mockup's "Other" category has no source in
the data model and inventing one would put a fabricated figure on a finance
chart. The activity feed reads the shared audit trail rather than a table of its
own, so it cannot drift from what actually happened.

Sales money is stored in minor currency units as `int64`. Binary floating point
cannot represent `0.1` exactly, so summing float lines drifts; an order total a
customer disputes is worse than one that is inconvenient to write. `Order.
Recalculate` is the single place that decides what an order costs, which is why
the discount extension goes through the module's service rather than its tables.

Recovery tokens are stored as SHA-256 hashes, are single-use, expire (1 hour for
reset, 24 hours for verification), and requesting a new one supersedes any
outstanding token. A completed reset revokes every existing session.

## Getting started

### Prerequisites

- Go 1.25 or newer
- Optional: PostgreSQL for future persistence and module integration

### Install dependencies

```bash
cd app
go mod tidy
```

### Run the application

```bash
cd app
go run ./cmd/server
```

### Run tests

```bash
cd app
go test ./...
```

## Notes

This repository is still in a foundational stage. The current code establishes the application skeleton and modular runtime architecture, while the business-specific ERP modules and integrations are meant to be added on top of this base.

## Roadmap direction

The project is intended to evolve toward a complete ERP platform with:

- module registration and lifecycle management,
- database migrations per module,
- authentication and authorization,
- event-driven communication across modules,
- domain modules such as base, inventory, and sales,
- API and frontend integration layers.

## License

This project is licensed under the GNU Lesser General Public License v3.0 (LGPL-3.0).

LGPL is a good fit for a modular ERP foundation because it allows the core framework to remain open and reusable while still giving you more flexibility for integrating the software into proprietary or custom deployments. It is especially suitable when you want to preserve openness for the platform itself, while permitting closed-source extensions or customer-specific add-ons that interact with the library in a controlled way.

See the full license text in [LICENSE](LICENSE).
