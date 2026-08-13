package core

import "context"

type Module interface {
	Name() string
	Depends() []string
	Register(app *App) error
	Boot(app *App) error
}

// Shutdownable is optional while existing modules are migrated. Implement it
// for modules that own resources such as workers, subscriptions, or clients.
type Shutdownable interface {
	Shutdown(context.Context, *App) error
}

// Migrator is implemented by modules that own persistent schema.
type Migrator interface {
	Migrate(*App) error
}
