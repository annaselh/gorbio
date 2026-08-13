package core

import "context"

type Extension interface {
	Name() string

	Module() string

	Depends() []string

	Register(app *App) error

	Boot(app *App) error
}

// ExtensionShutdownable is the optional shutdown contract for extensions.
type ExtensionShutdownable interface {
	Shutdown(context.Context, *App) error
}
