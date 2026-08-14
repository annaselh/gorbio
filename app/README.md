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
| POST | `/api/auth/email/verify` | – | Consume a verification token |
| POST | `/api/auth/email/resend` | auth | Re-send the verification email |

Business modules register their own routes behind `RequireAuth` and an explicit
permission; see `modules/inventory/routes.go` for the pattern.

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
